// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	"github.com/vmware/terraform-provider-nsxt/api/infra/segments"
	tier1s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_segments "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/segments"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var connectivityValues = []string{
	model.SegmentAdvancedConfig_CONNECTIVITY_ON,
	model.SegmentAdvancedConfig_CONNECTIVITY_OFF,
}

var urpfModeValues = []string{
	model.SegmentAdvancedConfig_URPF_MODE_NONE,
	model.SegmentAdvancedConfig_URPF_MODE_STRICT,
}

var replicationModeValues = []string{
	model.Segment_REPLICATION_MODE_MTEP,
	model.Segment_REPLICATION_MODE_SOURCE,
}

func getPolicySegmentDhcpV4ConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server_address": {
				Type:         schema.TypeString,
				Description:  "IP address of the DHCP server in CIDR format",
				Optional:     true,
				ValidateFunc: validateIPCidr(),
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Description: "IP addresses of DNS servers for subnet",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
				Optional: true,
			},
			"lease_time":          getDhcpLeaseTimeSchema(),
			"dhcp_option_121":     getDhcpOptions121Schema(),
			"dhcp_generic_option": getDhcpGenericOptionsSchema(),
		},
	}
}

func getPolicySegmentDhcpV6ConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server_address": {
				Type:        schema.TypeString,
				Description: "IP address of the DHCP server in CIDR format",
				Optional:    true,
				// TODO: validate IPv6 only
				ValidateFunc: validateIPCidr(),
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Description: "IP addresses of DNS servers for subnet",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
				Optional: true,
			},
			"lease_time":     getDhcpLeaseTimeSchema(),
			"domain_names":   getDomainNamesSchema(),
			"excluded_range": getAllocationRangeListSchema(false, "Excluded addresses to define dynamic ip allocation ranges"),
			"preferred_time": getDhcpPreferredTimeSchema(),
			"sntp_servers": {
				Type:        schema.TypeList,
				Description: "IPv6 address of SNTP servers for subnet",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
				Optional: true,
			},
			// TODO: add options
		},
	}
}

func getPolicySegmentSubnetSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dhcp_v4_config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     getPolicySegmentDhcpV4ConfigSchema(),
				MaxItems: 1,
			},
			"dhcp_v6_config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     getPolicySegmentDhcpV6ConfigSchema(),
				MaxItems: 1,
			},
			"dhcp_ranges": {
				Type:        schema.TypeList,
				Description: "DHCP address ranges for dynamic IP allocation",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrOrIPOrRange(),
				},
				Optional: true,
			},
			"cidr": {
				Type:         schema.TypeString,
				Description:  "Gateway IP address in CIDR format",
				Optional:     true,
				ValidateFunc: validateIPCidr(),
			},
			"network": {
				Type:        schema.TypeString,
				Description: "Network CIDR for subnet",
				Computed:    true,
			},
		},
	}
}

func getPolicySegmentBridgeConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"profile_path": getPolicyPathSchema(true, false, "profile path"),
			"uplink_teaming_policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vlan_ids": {
				Type:        schema.TypeList,
				Description: "VLAN specification for bridge endpoint",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateVLANIdOrRange,
				},
				Required: true,
			},
			"transport_zone_path": getPolicyPathSchema(true, false, "vlan transport zone path"),
		},
	}
}

func getPolicySegmentL2ExtensionConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"l2vpn_paths": {
				Type:        schema.TypeList,
				Description: "Policy paths of associated L2 VPN sessions",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePolicyPath(),
				},
				Optional: true,
			},
			"tunnel_id": {
				Type:         schema.TypeInt,
				Description:  "Tunnel ID",
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 4093),
			},
		},
	}
}

func getPolicySegmentAdvancedConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address_pool_path": {
				Type:         schema.TypeString,
				Description:  "Policy path to IP address pool",
				Optional:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"connectivity": {
				Type:         schema.TypeString,
				Description:  "Connectivity configuration to manually connect (ON) or disconnect (OFF)",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(connectivityValues, false),
			},
			"hybrid": {
				Type:        schema.TypeBool,
				Description: "Flag to identify a hybrid logical switch",
				Optional:    true,
				Default:     false,
			},
			"local_egress": {
				Type:        schema.TypeBool,
				Description: "Flag to enable local egress",
				Optional:    true,
				Default:     false,
			},
			"uplink_teaming_policy": {
				Type:        schema.TypeString,
				Description: "The name of the switching uplink teaming policy for the bridge endpoint",
				Optional:    true,
			},
			"urpf_mode": {
				Type:         schema.TypeString,
				Description:  "This URPF mode is applied to the downlink logical router port created while attaching this segment to gateway",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(urpfModeValues, false),
				Default:      model.SegmentAdvancedConfig_URPF_MODE_STRICT,
			},
		},
	}
}

func getPolicySegmentDiscoveryProfilesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ip_discovery_profile_path":  getPolicyPathSchema(false, false, "Policy path of associated IP Discovery Profile"),
			"mac_discovery_profile_path": getPolicyPathSchema(false, false, "Policy path of associated Mac Discovery Profile"),
			"binding_map_path":           getComputedPolicyPathSchema("Policy path of profile binding map"),
			"revision":                   getRevisionSchema(),
		},
	}
}

func getPolicySegmentQosProfilesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"qos_profile_path": getPolicyPathSchema(true, false, "Policy path of associated QoS Profile"),
			"binding_map_path": getComputedPolicyPathSchema("Policy path of profile binding map"),
			"revision":         getRevisionSchema(),
		},
	}
}

func getPolicySegmentSecurityProfilesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"spoofguard_profile_path": getPolicyPathSchema(false, false, "Policy path of associated Spoofguard Profile"),
			"security_profile_path":   getPolicyPathSchema(false, false, "Policy path of associated Segment Security Profile"),
			"binding_map_path":        getComputedPolicyPathSchema("Policy path of profile binding map"),
			"revision":                getRevisionSchema(),
		},
	}
}

func getPolicyCommonSegmentSchema(vlanRequired bool, isFixed bool) map[string]*schema.Schema {
	schema := map[string]*schema.Schema{
		"nsx_id":       getNsxIDSchema(),
		"path":         getPathSchema(),
		"display_name": getDisplayNameSchema(),
		"description":  getDescriptionSchema(),
		"revision":     getRevisionSchema(),
		"tag":          getTagsSchema(),
		"ignore_tags":  getIgnoreTagsSchema(),
		"context":      getContextSchema(false, false, false),
		"advanced_config": {
			Type:        schema.TypeList,
			Description: "Advanced segment configuration",
			Elem:        getPolicySegmentAdvancedConfigurationSchema(),
			Optional:    true,
			MaxItems:    1,
		},
		"connectivity_path": {
			Type:         schema.TypeString,
			Description:  "Policy path to the connecting Tier-0 or Tier-1",
			Required:     isFixed,
			Optional:     !isFixed,
			ForceNew:     isFixed,
			ValidateFunc: validatePolicyPath(),
		},
		"domain_name": {
			Type:        schema.TypeString,
			Description: "DNS domain names",
			Optional:    true,
		},
		"l2_extension": {
			Type:        schema.TypeList,
			Description: "Configuration for extending Segment through L2 VPN",
			Elem:        getPolicySegmentL2ExtensionConfigurationSchema(),
			Optional:    true,
			MaxItems:    1,
		},
		"overlay_id": {
			Type:        schema.TypeInt,
			Description: "Overlay connectivity ID for this Segment",
			Optional:    true,
			Computed:    true,
		},
		"subnet": {
			Type:        schema.TypeList,
			Description: "Subnet configuration with at most 1 IPv4 CIDR and multiple IPv6 CIDRs",
			Elem:        getPolicySegmentSubnetSchema(),
			Optional:    true,
		},
		"dhcp_config_path": getPolicyPathSchema(false, false, "Policy path to DHCP server or relay configuration to use for subnets configured on this segment"),
		"transport_zone_path": {
			Type:         schema.TypeString,
			Description:  "Policy path to the transport zone",
			Optional:     true,
			ForceNew:     true,
			Computed:     true,
			ValidateFunc: validatePolicyPath(),
		},
		"vlan_ids": {
			Type:        schema.TypeList,
			Description: "VLAN IDs for VLAN backed Segment",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validateVLANIdOrRange,
			},
			Required: vlanRequired,
			Optional: !vlanRequired,
		},
		"discovery_profile": {
			Type:        schema.TypeList,
			Description: "IP and MAC discovery profiles for this segment",
			Elem:        getPolicySegmentDiscoveryProfilesSchema(),
			Optional:    true,
			MaxItems:    1,
		},
		"qos_profile": {
			Type:        schema.TypeList,
			Description: "QoS profiles for this segment",
			Elem:        getPolicySegmentQosProfilesSchema(),
			Optional:    true,
			MaxItems:    1,
		},
		"security_profile": {
			Type:        schema.TypeList,
			Description: "Security profiles for this segment",
			Elem:        getPolicySegmentSecurityProfilesSchema(),
			Optional:    true,
			MaxItems:    1,
		},
		"replication_mode": {
			Type:         schema.TypeString,
			Description:  "Replication mode - MTEP or SOURCE",
			Optional:     true,
			Default:      model.Segment_REPLICATION_MODE_MTEP,
			ValidateFunc: validation.StringInSlice(replicationModeValues, false),
		},
		"bridge_config": {
			Type:        schema.TypeList,
			Description: "Bridge configuration",
			Elem:        getPolicySegmentBridgeConfigSchema(),
			Optional:    true,
		},
		"metadata_proxy_paths": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Metadata Proxy Configuration Paths",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicyPath(),
			},
		},
	}

	if isFixed {
		// Profile assignment is not yet supported in the SDK
		delete(schema, "discovery_profile")
		delete(schema, "qos_profile")
		delete(schema, "security_profile")
	}

	return schema
}

func getPolicyDhcpOptions121(opts []interface{}) model.DhcpOption121 {
	var opt121Struct model.DhcpOption121
	var routes []model.ClasslessStaticRoute
	for _, opt121 := range opts {
		data := opt121.(map[string]interface{})
		network := data["network"].(string)
		nextHop := data["next_hop"].(string)
		elem := model.ClasslessStaticRoute{
			Network: &network,
			NextHop: &nextHop,
		}
		routes = append(routes, elem)
	}
	if len(routes) > 0 {
		opt121Struct = model.DhcpOption121{
			StaticRoutes: routes,
		}
	}
	return opt121Struct
}

func getDhcpLeaseTimeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Description:  "DHCP lease time in seconds",
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(60),
		Default:      86400,
	}
}

func getDhcpPreferredTimeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Description:  "The time interval in seconds, in which the prefix is advertised as preferred",
		Optional:     true,
		ValidateFunc: validation.IntAtLeast(48),
	}
}

func getDomainNamesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Domain names",
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func getPolicyDhcpOptions121FromStruct(opt *model.DhcpOption121) []map[string]interface{} {
	var dhcpOpt121 []map[string]interface{}
	for _, route := range opt.StaticRoutes {
		elem := make(map[string]interface{})
		elem["network"] = route.Network
		elem["next_hop"] = route.NextHop
		dhcpOpt121 = append(dhcpOpt121, elem)
	}
	return dhcpOpt121
}

func getPolicyDhcpGenericOptions(opts []interface{}) []model.GenericDhcpOption {
	var options []model.GenericDhcpOption
	for _, opt := range opts {
		data := opt.(map[string]interface{})
		code := int64(data["code"].(int))
		elem := model.GenericDhcpOption{
			Code:   &code,
			Values: interface2StringList(data["values"].([]interface{})),
		}
		options = append(options, elem)
	}
	return options
}

func getPolicyDhcpGenericOptionsFromStruct(opts []model.GenericDhcpOption) []map[string]interface{} {
	var dhcpOptions []map[string]interface{}
	for _, opt := range opts {
		elem := make(map[string]interface{})
		elem["code"] = opt.Code
		elem["values"] = opt.Values
		dhcpOptions = append(dhcpOptions, elem)
	}
	return dhcpOptions
}

func getDhcpOptsFromMap(dhcpConfig map[string]interface{}) *model.DhcpV4Options {
	dhcpOpts := model.DhcpV4Options{}

	dhcp121Opts := dhcpConfig["dhcp_option_121"].([]interface{})
	if len(dhcp121Opts) > 0 {
		dhcp121OptStruct := getPolicyDhcpOptions121(dhcp121Opts)
		dhcpOpts.Option121 = &dhcp121OptStruct
	}

	otherDhcpOpts := dhcpConfig["dhcp_generic_option"].([]interface{})
	if len(otherDhcpOpts) > 0 {
		otherOptStructs := getPolicyDhcpGenericOptions(otherDhcpOpts)
		dhcpOpts.Others = otherOptStructs
	}

	if len(dhcp121Opts)+len(otherDhcpOpts) > 0 {
		return &dhcpOpts
	}

	return nil

}

func getSegmentSubnetDhcpConfigFromSchema(schemaConfig map[string]interface{}) (*data.StructValue, error) {
	if util.NsxVersionLower("3.0.0") {
		return nil, nil
	}

	dhcpV4Config := schemaConfig["dhcp_v4_config"].([]interface{})
	dhcpV6Config := schemaConfig["dhcp_v6_config"].([]interface{})

	if (len(dhcpV4Config) > 0) && (len(dhcpV6Config) > 0) {
		return nil, fmt.Errorf("Only one of ['dhcp_v4_config','dhcp_v6_config'] should be specified in single subnet")
	}

	converter := bindings.NewTypeConverter()

	if len(dhcpV4Config) > 0 {
		dhcpConfig := dhcpV4Config[0].(map[string]interface{})
		serverAddress := dhcpConfig["server_address"].(string)
		dnsServers := dhcpConfig["dns_servers"].([]interface{})
		leaseTime := int64(dhcpConfig["lease_time"].(int))

		config := model.SegmentDhcpV4Config{
			ResourceType: model.SegmentDhcpConfig_RESOURCE_TYPE_SEGMENTDHCPV4CONFIG,
			DnsServers:   interface2StringList(dnsServers),
		}

		if len(serverAddress) > 0 {
			config.ServerAddress = &serverAddress
		}

		config.Options = getDhcpOptsFromMap(dhcpConfig)

		if leaseTime > 0 {
			config.LeaseTime = &leaseTime
		}

		dataValue, errs := converter.ConvertToVapi(config, model.SegmentDhcpV4ConfigBindingType())
		if errs != nil {
			return nil, errs[0]
		}

		return dataValue.(*data.StructValue), nil
	}

	if len(dhcpV6Config) > 0 {
		dhcpConfig := dhcpV6Config[0].(map[string]interface{})
		serverAddress := dhcpConfig["server_address"].(string)
		dnsServers := dhcpConfig["dns_servers"].([]interface{})
		sntpServers := dhcpConfig["sntp_servers"].([]interface{})
		domainNames := dhcpConfig["domain_names"].([]interface{})
		excludedRanges := dhcpConfig["excluded_range"].([]interface{})
		leaseTime := int64(dhcpConfig["lease_time"].(int))
		preferredTime := int64(dhcpConfig["preferred_time"].(int))

		config := model.SegmentDhcpV6Config{
			ResourceType: model.SegmentDhcpConfig_RESOURCE_TYPE_SEGMENTDHCPV6CONFIG,
			DnsServers:   interface2StringList(dnsServers),
		}

		if len(serverAddress) > 0 {
			config.ServerAddress = &serverAddress
		}

		if len(domainNames) > 0 {
			config.DomainNames = interface2StringList(domainNames)
		}

		if len(sntpServers) > 0 {
			config.SntpServers = interface2StringList(sntpServers)
		}

		if len(excludedRanges) > 0 {
			var rangeList []string
			for _, excludedRange := range excludedRanges {
				rangeMap := excludedRange.(map[string]interface{})
				rangeList = append(rangeList, fmt.Sprintf("%s-%s", rangeMap["start"].(string), rangeMap["end"].(string)))
			}
			config.ExcludedRanges = rangeList
		}

		if leaseTime > 0 {
			config.LeaseTime = &leaseTime
		}

		if preferredTime > 0 {
			config.PreferredTime = &preferredTime
		}

		dataValue, errs := converter.ConvertToVapi(config, model.SegmentDhcpV6ConfigBindingType())
		if errs != nil {
			return nil, errs[0]
		}

		return dataValue.(*data.StructValue), nil
	}

	return nil, nil

}

func setSegmentSubnetDhcpConfigInSchema(schemaConfig map[string]interface{}, subnetConfig model.SegmentSubnet) error {
	converter := bindings.NewTypeConverter()

	var resultConfigs []map[string]interface{}
	resultConfig := make(map[string]interface{})

	if subnetConfig.DhcpConfig == nil {
		return nil
	}

	obj, errs := converter.ConvertToGolang(subnetConfig.DhcpConfig, model.SegmentDhcpConfigBindingType())
	if errs != nil {
		return errs[0]
	}

	resourceType := obj.(model.SegmentDhcpConfig).ResourceType
	if resourceType == model.SegmentDhcpConfig_RESOURCE_TYPE_SEGMENTDHCPV4CONFIG {
		obj, errs := converter.ConvertToGolang(subnetConfig.DhcpConfig, model.SegmentDhcpV4ConfigBindingType())
		if errs != nil {
			return errs[0]
		}
		dhcpV4Config := obj.(model.SegmentDhcpV4Config)
		resultConfig["server_address"] = dhcpV4Config.ServerAddress
		resultConfig["lease_time"] = dhcpV4Config.LeaseTime
		resultConfig["dns_servers"] = dhcpV4Config.DnsServers

		if dhcpV4Config.Options != nil {
			if dhcpV4Config.Options.Option121 != nil {
				opts := getPolicyDhcpOptions121FromStruct(dhcpV4Config.Options.Option121)
				resultConfig["dhcp_option_121"] = opts
			}
			if len(dhcpV4Config.Options.Others) > 0 {
				opts := getPolicyDhcpGenericOptionsFromStruct(dhcpV4Config.Options.Others)
				resultConfig["dhcp_generic_option"] = opts
			}
		}
		resultConfigs = append(resultConfigs, resultConfig)
		schemaConfig["dhcp_v4_config"] = resultConfigs
		return nil
	}

	if resourceType == model.SegmentDhcpConfig_RESOURCE_TYPE_SEGMENTDHCPV6CONFIG {
		obj, errs := converter.ConvertToGolang(subnetConfig.DhcpConfig, model.SegmentDhcpV6ConfigBindingType())
		if errs != nil {
			return errs[0]
		}

		dhcpV6Config := obj.(model.SegmentDhcpV6Config)
		resultConfig["server_address"] = dhcpV6Config.ServerAddress
		resultConfig["lease_time"] = dhcpV6Config.LeaseTime
		resultConfig["preferred_time"] = dhcpV6Config.PreferredTime
		resultConfig["dns_servers"] = dhcpV6Config.DnsServers
		resultConfig["sntp_servers"] = dhcpV6Config.SntpServers
		resultConfig["domain_names"] = dhcpV6Config.DomainNames

		var excludedRanges []map[string]interface{}
		for _, excludedRange := range dhcpV6Config.ExcludedRanges {
			addresses := strings.Split(excludedRange, "-")
			if len(addresses) == 2 {
				rangeMap := make(map[string]interface{})
				rangeMap["start"] = addresses[0]
				rangeMap["end"] = addresses[1]
				excludedRanges = append(excludedRanges, rangeMap)
			}
		}
		resultConfig["excluded_range"] = excludedRanges

		resultConfigs = append(resultConfigs, resultConfig)
		schemaConfig["dhcp_v6_config"] = resultConfigs
		return nil
	}

	return fmt.Errorf("Unrecognized DHCP Config Resource Type %s", resourceType)

}

func nsxtPolicySegmentAddGatewayToInfraStruct(d *schema.ResourceData, dataValue *data.StructValue) (*data.StructValue, error) {
	var gwChildren []*data.StructValue
	converter := bindings.NewTypeConverter()
	gwChildren = append(gwChildren, dataValue)
	targetType := "Tier1"
	gwPath := d.Get("connectivity_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if gwID == "" {
		return nil, fmt.Errorf("connectivity_path is not a valid gateway path")
	}
	if isT0 {
		return nil, fmt.Errorf("Tier0 fixed segments are not supported")
	}
	childGW := model.ChildResourceReference{
		Id:           &gwID,
		ResourceType: "ChildResourceReference",
		TargetType:   &targetType,
		Children:     gwChildren,
	}
	dataValue1, errors := converter.ConvertToVapi(childGW, model.ChildResourceReferenceBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting Gateway Child: %v", errors[0])
	}

	return dataValue1.(*data.StructValue), nil
}

func policySegmentResourceToInfraStruct(context utl.SessionContext, id string, d *schema.ResourceData, isVlan bool, isFixed bool) (model.Infra, error) {
	// Read the rest of the configured parameters
	var infraChildren []*data.StructValue

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	domainName := d.Get("domain_name").(string)
	tzPath := d.Get("transport_zone_path").(string)
	replicationMode := d.Get("replication_mode").(string)
	dhcpConfigPath := d.Get("dhcp_config_path").(string)
	revision := int64(d.Get("revision").(int))
	resourceType := "Segment"

	if (tzPath == "") && context.ClientType == utl.Local && !isFixed {
		return model.Infra{}, fmt.Errorf("transport_zone_path needs to be specified for infra segment on local manager")
	}

	obj := model.Segment{
		Id:           &id,
		DisplayName:  &displayName,
		Tags:         tags,
		Revision:     &revision,
		ResourceType: &resourceType,
	}

	if description != "" {
		obj.Description = &description
	}
	if domainName != "" {
		obj.DomainName = &domainName
	}
	if tzPath != "" {
		obj.TransportZonePath = &tzPath
	}
	if util.NsxVersionHigherOrEqual("3.0.0") {
		obj.ReplicationMode = &replicationMode
		if dhcpConfigPath != "" {
			obj.DhcpConfigPath = &dhcpConfigPath
		}
	}

	var vlanIds []string
	var subnets []interface{}
	var subnetStructs []model.SegmentSubnet
	// VLAN specific fields
	for _, vlanID := range d.Get("vlan_ids").([]interface{}) {
		vlanIds = append(vlanIds, vlanID.(string))
	}
	obj.VlanIds = vlanIds
	if !isVlan {
		// overlay specific fields
		connectivityPath := d.Get("connectivity_path").(string)
		overlayID, exists := d.GetOk("overlay_id")
		if exists {
			overlayID64 := int64(overlayID.(int))
			obj.OverlayId = &overlayID64
		}
		if connectivityPath != "" && !isFixed {
			obj.ConnectivityPath = &connectivityPath
		}
	}
	subnets = d.Get("subnet").([]interface{})
	if len(subnets) > 0 {
		for _, subnet := range subnets {
			subnetMap := subnet.(map[string]interface{})
			dhcpRanges := subnetMap["dhcp_ranges"].([]interface{})
			var dhcpRangeList []string
			if len(dhcpRanges) > 0 {
				for _, dhcpRange := range dhcpRanges {
					dhcpRangeList = append(dhcpRangeList, dhcpRange.(string))
				}
			}
			gwAddr := subnetMap["cidr"].(string)
			network := subnetMap["network"].(string)
			subnetStruct := model.SegmentSubnet{
				DhcpRanges: dhcpRangeList,
				Network:    &network,
			}
			if len(gwAddr) > 0 {
				subnetStruct.GatewayAddress = &gwAddr
			}

			config, err := getSegmentSubnetDhcpConfigFromSchema(subnetMap)
			if err != nil {
				return model.Infra{}, err
			}

			subnetStruct.DhcpConfig = config

			subnetStructs = append(subnetStructs, subnetStruct)
		}
	}
	obj.Subnets = subnetStructs

	advConfig := d.Get("advanced_config").([]interface{})
	if len(advConfig) > 0 {
		advConfigMap := advConfig[0].(map[string]interface{})
		connectivity := advConfigMap["connectivity"].(string)
		hybrid := advConfigMap["hybrid"].(bool)
		egress := advConfigMap["local_egress"].(bool)
		var poolPaths []string
		if advConfigMap["cidr"] != nil {
			poolPaths = append(poolPaths, advConfigMap["cidr"].(string))
		}
		advConfigStruct := model.SegmentAdvancedConfig{
			AddressPoolPaths: poolPaths,
			Hybrid:           &hybrid,
			LocalEgress:      &egress,
		}

		if connectivity != "" {
			advConfigStruct.Connectivity = &connectivity
		}

		if util.NsxVersionHigherOrEqual("3.0.0") {
			teamingPolicy := advConfigMap["uplink_teaming_policy"].(string)
			if teamingPolicy != "" {
				advConfigStruct.UplinkTeamingPolicyName = &teamingPolicy
			}

			poolPath := advConfigMap["address_pool_path"].(string)
			if poolPath != "" {
				advConfigStruct.AddressPoolPaths = append(advConfigStruct.AddressPoolPaths, poolPath)
			}

			if util.NsxVersionHigherOrEqual("3.1.0") {
				urpfMode := advConfigMap["urpf_mode"].(string)
				advConfigStruct.UrpfMode = &urpfMode
			}
		}
		obj.AdvancedConfig = &advConfigStruct
	}

	l2Ext := d.Get("l2_extension").([]interface{})
	if len(l2Ext) > 0 {
		l2ExtMap := l2Ext[0].(map[string]interface{})
		vpnPaths := interfaceListToStringList(l2ExtMap["l2vpn_paths"].([]interface{}))
		tunnelID := int64(l2ExtMap["tunnel_id"].(int))
		l2Struct := model.L2Extension{
			L2vpnPaths: vpnPaths,
			TunnelId:   &tunnelID,
		}
		obj.L2Extension = &l2Struct
	}

	if !isFixed {
		err := nsxtPolicySegmentProfilesSetInStruct(d, &obj)
		if err != nil {
			return model.Infra{}, err
		}
	}

	if context.ClientType != utl.Global {
		setBridgeConfigInStruct(d, &obj)
	}

	childSegment := model.ChildSegment{
		Segment:      &obj,
		ResourceType: "ChildSegment",
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childSegment, model.ChildSegmentBindingType())
	if errors != nil {
		return model.Infra{}, fmt.Errorf("Error converting Segment Child: %v", errors[0])
	}

	if isFixed {
		dataValue, err := nsxtPolicySegmentAddGatewayToInfraStruct(d, dataValue.(*data.StructValue))
		if err != nil {
			return model.Infra{}, err
		}
		infraChildren = append(infraChildren, dataValue)

	} else {
		infraChildren = append(infraChildren, dataValue.(*data.StructValue))

	}

	infraType := "Infra"
	infraStruct := model.Infra{
		Children:     infraChildren,
		ResourceType: &infraType,
	}

	return infraStruct, nil
}

func resourceNsxtPolicySegmentExists(context utl.SessionContext, gwPath string, isFixed bool) func(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
		_, err := nsxtPolicyGetSegment(context, connector, id, gwPath, isFixed)
		if err == nil {
			return true, nil
		}

		if isNotFoundError(err) {
			return false, nil
		}

		return false, logAPIError("Error retrieving Segment", err)

	}
}

func setBridgeConfigInStruct(d *schema.ResourceData, segment *model.Segment) {

	configs := d.Get("bridge_config").([]interface{})
	if len(configs) == 0 {
		return
	}

	for _, config := range configs {
		bridgeConfig := config.(map[string]interface{})
		profilePath := bridgeConfig["profile_path"].(string)
		policyName := bridgeConfig["uplink_teaming_policy"].(string)
		tzPath := bridgeConfig["transport_zone_path"].(string)
		vlanIds := interfaceListToStringList(bridgeConfig["vlan_ids"].([]interface{}))

		profile := model.BridgeProfileConfig{
			BridgeProfilePath: &profilePath,
		}

		if len(policyName) > 0 {
			profile.UplinkTeamingPolicyName = &policyName
		}
		if len(tzPath) > 0 {
			profile.VlanTransportZonePath = &tzPath
		}

		if len(vlanIds) > 0 {
			profile.VlanIds = vlanIds
		}

		segment.BridgeProfiles = append(segment.BridgeProfiles, profile)

	}
}

func nsxtPolicySegmentProfilesSetInStruct(d *schema.ResourceData, segment *model.Segment) error {
	var children []*data.StructValue

	child, err := nsxtPolicySegmentDiscoveryProfileSetInStruct(d)
	if err != nil {
		return err
	}

	if child != nil {
		children = append(children, child)
	}

	child, err = nsxtPolicySegmentQosProfileSetInStruct(d)
	if err != nil {
		return err
	}

	if child != nil {
		children = append(children, child)
	}

	child, err = nsxtPolicySegmentSecurityProfileSetInStruct(d)
	if err != nil {
		return err
	}

	if child != nil {
		children = append(children, child)
	}

	segment.Children = children
	return nil

}

func getOldProfileDataForRemoval(oldProfiles interface{}) (string, int64) {
	profileMap := oldProfiles.([]interface{})[0].(map[string]interface{})
	segmentProfileMapID := getPolicyIDFromPath(profileMap["binding_map_path"].(string))
	revision := int64(profileMap["revision"].(int))

	return segmentProfileMapID, revision
}

func nsxtPolicySegmentDiscoveryProfileSetInStruct(d *schema.ResourceData) (*data.StructValue, error) {
	segmentProfileMapID := "default"

	ipDiscoveryProfilePath := ""
	macDiscoveryProfilePath := ""
	revision := int64(0)
	shouldDelete := false
	oldProfiles, newProfiles := d.GetChange("discovery_profile")
	if len(newProfiles.([]interface{})) > 0 {
		profileMap := newProfiles.([]interface{})[0].(map[string]interface{})

		ipDiscoveryProfilePath = profileMap["ip_discovery_profile_path"].(string)
		macDiscoveryProfilePath = profileMap["mac_discovery_profile_path"].(string)
		if len(profileMap["binding_map_path"].(string)) > 0 {
			segmentProfileMapID = getPolicyIDFromPath(profileMap["binding_map_path"].(string))
		}

		revision = int64(profileMap["revision"].(int))
	} else {
		if len(oldProfiles.([]interface{})) == 0 {
			return nil, nil
		}
		segmentProfileMapID, revision = getOldProfileDataForRemoval(oldProfiles)
		shouldDelete = true
	}

	resourceType := "SegmentDiscoveryProfileBindingMap"
	discoveryMap := model.SegmentDiscoveryProfileBindingMap{
		ResourceType: &resourceType,
		Id:           &segmentProfileMapID,
	}

	if len(oldProfiles.([]interface{})) > 0 {
		// This is an update
		discoveryMap.Revision = &revision
	}

	if len(ipDiscoveryProfilePath) > 0 {
		discoveryMap.IpDiscoveryProfilePath = &ipDiscoveryProfilePath
	}

	if len(macDiscoveryProfilePath) > 0 {
		discoveryMap.MacDiscoveryProfilePath = &macDiscoveryProfilePath
	}

	childConfig := model.ChildSegmentDiscoveryProfileBindingMap{
		ResourceType:                      "ChildSegmentDiscoveryProfileBindingMap",
		SegmentDiscoveryProfileBindingMap: &discoveryMap,
		Id:                                &segmentProfileMapID,
		MarkedForDelete:                   &shouldDelete,
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childConfig, model.ChildSegmentDiscoveryProfileBindingMapBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting child segment discovery map: %v", errors[0])
	}

	return dataValue.(*data.StructValue), nil
}

func nsxtPolicySegmentQosProfileSetInStruct(d *schema.ResourceData) (*data.StructValue, error) {
	segmentProfileMapID := "default"

	qosProfilePath := ""
	revision := int64(0)
	oldProfiles, newProfiles := d.GetChange("qos_profile")
	shouldDelete := false
	if len(newProfiles.([]interface{})) > 0 {
		profileMap := newProfiles.([]interface{})[0].(map[string]interface{})

		qosProfilePath = profileMap["qos_profile_path"].(string)
		if len(profileMap["binding_map_path"].(string)) > 0 {
			segmentProfileMapID = getPolicyIDFromPath(profileMap["binding_map_path"].(string))
		}

		revision = int64(profileMap["revision"].(int))
	} else {
		if len(oldProfiles.([]interface{})) == 0 {
			return nil, nil
		}
		// Profile should be deleted
		segmentProfileMapID, revision = getOldProfileDataForRemoval(oldProfiles)
		shouldDelete = true
	}

	resourceType := "SegmentQoSProfileBindingMap"
	qosMap := model.SegmentQosProfileBindingMap{
		ResourceType: &resourceType,
		Id:           &segmentProfileMapID,
	}

	if len(oldProfiles.([]interface{})) > 0 {
		// This is an update
		qosMap.Revision = &revision
	}

	if len(qosProfilePath) > 0 {
		qosMap.QosProfilePath = &qosProfilePath
	}

	childConfig := model.ChildSegmentQosProfileBindingMap{
		ResourceType:                "ChildSegmentQoSProfileBindingMap",
		SegmentQosProfileBindingMap: &qosMap,
		Id:                          &segmentProfileMapID,
		MarkedForDelete:             &shouldDelete,
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childConfig, model.ChildSegmentQosProfileBindingMapBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting child segment QoS map: %v", errors[0])
	}

	return dataValue.(*data.StructValue), nil
}

func nsxtPolicySegmentSecurityProfileSetInStruct(d *schema.ResourceData) (*data.StructValue, error) {
	segmentProfileMapID := "default"

	spoofguardProfilePath := ""
	securityProfilePath := ""
	revision := int64(0)
	oldProfiles, newProfiles := d.GetChange("security_profile")
	shouldDelete := false
	if len(newProfiles.([]interface{})) > 0 {
		profileMap := newProfiles.([]interface{})[0].(map[string]interface{})

		spoofguardProfilePath = profileMap["spoofguard_profile_path"].(string)
		securityProfilePath = profileMap["security_profile_path"].(string)
		if len(profileMap["binding_map_path"].(string)) > 0 {
			segmentProfileMapID = getPolicyIDFromPath(profileMap["binding_map_path"].(string))
		}

		revision = int64(profileMap["revision"].(int))
	} else {
		if len(oldProfiles.([]interface{})) == 0 {
			return nil, nil
		}
		// Profile should be deleted
		segmentProfileMapID, revision = getOldProfileDataForRemoval(oldProfiles)
		shouldDelete = true
	}

	resourceType := "SegmentSecurityProfileBindingMap"
	securityMap := model.SegmentSecurityProfileBindingMap{
		ResourceType: &resourceType,
		Id:           &segmentProfileMapID,
	}

	if len(oldProfiles.([]interface{})) > 0 {
		// This is an update
		securityMap.Revision = &revision
	}

	if len(spoofguardProfilePath) > 0 {
		securityMap.SpoofguardProfilePath = &spoofguardProfilePath
	}

	if len(securityProfilePath) > 0 {
		securityMap.SegmentSecurityProfilePath = &securityProfilePath
	}

	childConfig := model.ChildSegmentSecurityProfileBindingMap{
		ResourceType:                     "ChildSegmentSecurityProfileBindingMap",
		SegmentSecurityProfileBindingMap: &securityMap,
		Id:                               &segmentProfileMapID,
		MarkedForDelete:                  &shouldDelete,
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childConfig, model.ChildSegmentSecurityProfileBindingMapBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting child segment security map: %v", errors[0])
	}

	return dataValue.(*data.StructValue), nil
}

func nsxtPolicySegmentDiscoveryProfileRead(d *schema.ResourceData, m interface{}) error {
	errorMessage := "Failed to read Discovery Profile Map for segment %s: %s"
	connector := getPolicyConnector(m)
	segmentID := d.Id()
	var results model.SegmentDiscoveryProfileBindingMapListResult
	if isPolicyGlobalManager(m) {
		client := gm_segments.NewSegmentDiscoveryProfileBindingMapsClient(connector)
		gmResults, err := client.List(segmentID, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
		lmResults, err := convertModelBindingType(gmResults, gm_model.SegmentDiscoveryProfileBindingMapListResultBindingType(), model.SegmentDiscoveryProfileBindingMapListResultBindingType())
		if err != nil {
			return err
		}
		results = lmResults.(model.SegmentDiscoveryProfileBindingMapListResult)
	} else {
		client := segments.NewSegmentDiscoveryProfileBindingMapsClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		var err error
		results, err = client.List(segmentID, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
	}

	config := make(map[string]interface{})
	var configList []map[string]interface{}

	for _, obj := range results.Results {
		config["ip_discovery_profile_path"] = obj.IpDiscoveryProfilePath
		config["mac_discovery_profile_path"] = obj.MacDiscoveryProfilePath
		config["binding_map_path"] = obj.Path
		config["revision"] = obj.Revision
		configList = append(configList, config)
		d.Set("discovery_profile", configList)
		return nil
	}

	return nil
}

func nsxtPolicySegmentQosProfileRead(d *schema.ResourceData, m interface{}) error {
	errorMessage := "Failed to read QoS Profile Map for segment %s: %s"
	connector := getPolicyConnector(m)
	segmentID := d.Id()
	var results model.SegmentQosProfileBindingMapListResult
	if isPolicyGlobalManager(m) {
		client := gm_segments.NewSegmentQosProfileBindingMapsClient(connector)
		gmResults, err := client.List(segmentID, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
		lmResults, err := convertModelBindingType(gmResults, gm_model.SegmentQosProfileBindingMapListResultBindingType(), model.SegmentQosProfileBindingMapListResultBindingType())
		if err != nil {
			return err
		}
		results = lmResults.(model.SegmentQosProfileBindingMapListResult)
	} else {
		client := segments.NewSegmentQosProfileBindingMapsClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		var err error
		results, err = client.List(segmentID, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
	}

	config := make(map[string]interface{})
	var configList []map[string]interface{}

	for _, obj := range results.Results {
		if obj.QosProfilePath != nil && (len(*obj.QosProfilePath) > 0) {
			config["qos_profile_path"] = obj.QosProfilePath
			config["binding_map_path"] = obj.Path
			config["revision"] = obj.Revision
			configList = append(configList, config)
			d.Set("qos_profile", configList)
			return nil
		}
	}

	return nil
}

func nsxtPolicySegmentSecurityProfileRead(d *schema.ResourceData, m interface{}) error {
	errorMessage := "Failed to read Security Profile Map for segment %s: %s"
	connector := getPolicyConnector(m)
	segmentID := d.Id()
	var results model.SegmentSecurityProfileBindingMapListResult
	if isPolicyGlobalManager(m) {
		client := gm_segments.NewSegmentSecurityProfileBindingMapsClient(connector)
		gmResults, err := client.List(segmentID, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
		lmResults, err := convertModelBindingType(gmResults, gm_model.SegmentSecurityProfileBindingMapListResultBindingType(), model.SegmentSecurityProfileBindingMapListResultBindingType())
		if err != nil {
			return err
		}
		results = lmResults.(model.SegmentSecurityProfileBindingMapListResult)
	} else {
		client := segments.NewSegmentSecurityProfileBindingMapsClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		var err error
		results, err = client.List(segmentID, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf(errorMessage, segmentID, err)
		}
	}

	config := make(map[string]interface{})
	var configList []map[string]interface{}

	for _, obj := range results.Results {
		config["security_profile_path"] = obj.SegmentSecurityProfilePath
		config["spoofguard_profile_path"] = obj.SpoofguardProfilePath
		config["binding_map_path"] = obj.Path
		config["revision"] = obj.Revision
		configList = append(configList, config)
		d.Set("security_profile", configList)
		return nil
	}

	return nil
}

func nsxtPolicySegmentProfilesRead(d *schema.ResourceData, m interface{}) error {

	err := nsxtPolicySegmentDiscoveryProfileRead(d, m)
	if err != nil {
		return err
	}

	err = nsxtPolicySegmentQosProfileRead(d, m)
	if err != nil {
		return err
	}

	err = nsxtPolicySegmentSecurityProfileRead(d, m)
	if err != nil {
		return err
	}

	return nil
}

func setSegmentBridgeConfigInSchema(d *schema.ResourceData, obj *model.Segment) {
	var configs []map[string]interface{}

	for _, profile := range obj.BridgeProfiles {
		config := make(map[string]interface{})
		config["profile_path"] = profile.BridgeProfilePath
		config["uplink_teaming_policy"] = profile.UplinkTeamingPolicyName
		config["vlan_ids"] = profile.VlanIds
		config["transport_zone_path"] = profile.VlanTransportZonePath

		configs = append(configs, config)
	}

	d.Set("bridge_config", configs)
}

func nsxtPolicyGetSegment(context utl.SessionContext, connector client.Connector, id string, gwPath string, isFixed bool) (model.Segment, error) {
	if !isFixed {
		client := infra.NewSegmentsClient(context, connector)
		if client == nil {
			return model.Segment{}, policyResourceNotSupportedError()
		}
		return client.Get(id)
	}

	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if gwID == "" {
		return model.Segment{}, fmt.Errorf("connectivity_path is not a valid gateway path")
	}
	if isT0 {
		return model.Segment{}, fmt.Errorf("Tier-0 fixed segments are not supported")
	}

	client := tier1s.NewSegmentsClient(context, connector)
	if client == nil {
		return model.Segment{}, policyResourceNotSupportedError()
	}
	return client.Get(gwID, id)
}

func nsxtPolicySegmentRead(d *schema.ResourceData, m interface{}, isVlan bool, isFixed bool) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Segment ID")
	}

	gwPath := ""
	if !isVlan {
		gwPath = d.Get("connectivity_path").(string)
	}

	obj, err := nsxtPolicyGetSegment(getSessionContext(d, m), connector, id, gwPath, isFixed)

	if err != nil {
		return handleReadError(d, "Segment", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	if !isVlan {
		d.Set("connectivity_path", obj.ConnectivityPath)
	}
	d.Set("dhcp_config_path", obj.DhcpConfigPath)
	d.Set("domain_name", obj.DomainName)
	d.Set("transport_zone_path", obj.TransportZonePath)

	d.Set("vlan_ids", obj.VlanIds)
	if !isVlan {
		if obj.OverlayId != nil {
			d.Set("overlay_id", int(*obj.OverlayId))
		} else {
			d.Set("overlay_id", 0)
		}
	}

	if util.NsxVersionHigherOrEqual("3.0.0") {
		d.Set("replication_mode", obj.ReplicationMode)
	}

	if obj.AdvancedConfig != nil {
		advConfig := make(map[string]interface{})
		poolPaths := obj.AdvancedConfig.AddressPoolPaths
		if len(poolPaths) > 0 {
			advConfig["address_pool_path"] = poolPaths[0]
		}
		advConfig["connectivity"] = obj.AdvancedConfig.Connectivity
		advConfig["hybrid"] = obj.AdvancedConfig.Hybrid
		advConfig["local_egress"] = obj.AdvancedConfig.LocalEgress
		if obj.AdvancedConfig.UplinkTeamingPolicyName != nil {
			advConfig["uplink_teaming_policy"] = *obj.AdvancedConfig.UplinkTeamingPolicyName
		}
		if obj.AdvancedConfig.UrpfMode != nil {
			advConfig["urpf_mode"] = *obj.AdvancedConfig.UrpfMode
		} else {
			if util.NsxVersionLower("3.1.0") {
				// set to default in early versions
				advConfig["urpf_mode"] = model.SegmentAdvancedConfig_URPF_MODE_STRICT
			}
		}
		// This is a list with 1 element
		var advConfigList []map[string]interface{}
		advConfigList = append(advConfigList, advConfig)
		userConfig := d.Get("advanced_config").([]interface{})
		if len(userConfig) > 0 {
			d.Set("advanced_config", advConfigList)
		}
	}

	if obj.L2Extension != nil {
		l2Ext := make(map[string]interface{})
		l2Ext["l2vpn_paths"] = obj.L2Extension.L2vpnPaths
		l2Ext["tunnel_id"] = obj.L2Extension.TunnelId
		// This is a list with 1 element
		var l2ExtList []map[string]interface{}
		l2ExtList = append(l2ExtList, l2Ext)
		d.Set("l2_extension", l2ExtList)
	}

	var subnetSegments []interface{}
	for _, subnetSeg := range obj.Subnets {
		seg := make(map[string]interface{})
		seg["dhcp_ranges"] = subnetSeg.DhcpRanges
		seg["cidr"] = subnetSeg.GatewayAddress
		seg["network"] = subnetSeg.Network
		err := setSegmentSubnetDhcpConfigInSchema(seg, subnetSeg)
		if err != nil {
			return err
		}
		subnetSegments = append(subnetSegments, seg)
	}

	d.Set("subnet", subnetSegments)

	if !isFixed {
		err = nsxtPolicySegmentProfilesRead(d, m)
		if err != nil {
			return err
		}
	}

	if !isPolicyGlobalManager(m) {
		setSegmentBridgeConfigInSchema(d, &obj)
	}

	return nil
}

func nsxtPolicySegmentCreate(d *schema.ResourceData, m interface{}, isVlan bool, isFixed bool) error {

	// Initialize resource Id and verify this ID is not yet used
	gwPath := ""
	if !isVlan {
		gwPath = d.Get("connectivity_path").(string)
	}

	id, err := getOrGenerateID2(d, m, resourceNsxtPolicySegmentExists(getSessionContext(d, m), gwPath, isFixed))
	if err != nil {
		return err
	}

	obj, err := policySegmentResourceToInfraStruct(getSessionContext(d, m), id, d, isVlan, isFixed)
	if err != nil {
		return err
	}

	err = policyInfraPatch(getSessionContext(d, m), obj, getPolicyConnector(m), false)
	if err != nil {
		return handleCreateError("Segment", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return nsxtPolicySegmentRead(d, m, isVlan, isFixed)
}

func nsxtPolicySegmentUpdate(d *schema.ResourceData, m interface{}, isVlan bool, isFixed bool) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Segment ID")
	}

	obj, err := policySegmentResourceToInfraStruct(getSessionContext(d, m), id, d, isVlan, isFixed)
	if err != nil {
		return err
	}

	err = policyInfraPatch(getSessionContext(d, m), obj, getPolicyConnector(m), true)
	if err != nil {
		return handleCreateError("Segment", id, err)
	}

	return nsxtPolicySegmentRead(d, m, isVlan, isFixed)
}

func nsxtPolicySegmentDelete(d *schema.ResourceData, m interface{}, isFixed bool) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Segment ID")
	}

	connector := getPolicyConnector(m)

	// During bulk destroy, VMs might be destroyed before segments, but
	// VIF release is not yet propagated to NSX. NSX will reply with
	// InvalidRequest on attempted delete if ports are present on the
	// segment. The code below waits till possible ports are deleted.
	pendingStates := []string{"pending"}
	targetStates := []string{"ok", "error"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			var ports interface{}
			var numOfPorts int
			if isPolicyGlobalManager(m) {
				portsClient := gm_segments.NewPortsClient(connector)
				gmPorts, err := portsClient.List(id, nil, nil, nil, nil, nil, nil)
				if err != nil {
					return gmPorts, "error", logAPIError("Error listing segment ports", err)
				}
				numOfPorts = len(gmPorts.Results)
				ports = gmPorts
			} else {
				portsClient := segments.NewPortsClient(getSessionContext(d, m), connector)
				if portsClient == nil {
					return nil, "error", policyResourceNotSupportedError()
				}
				lmPorts, err := portsClient.List(id, nil, nil, nil, nil, nil, nil)
				if err != nil {
					return lmPorts, "error", logAPIError("Error listing segment ports", err)
				}
				numOfPorts = len(lmPorts.Results)
				ports = lmPorts
			}

			log.Printf("[DEBUG] Current number of ports on segment %s is %d", id, numOfPorts)

			if numOfPorts > 0 {
				return ports, "pending", nil
			}
			return ports, "ok", nil

		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	if !isFixed {
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("Failed to get port information for segment %s: %v", id, err)
		}
	}

	var infraChildren []*data.StructValue
	converter := bindings.NewTypeConverter()
	boolTrue := true

	objType := "Segment"
	obj := model.Segment{
		Id:           &id,
		ResourceType: &objType,
	}

	childObj := model.ChildSegment{
		MarkedForDelete: &boolTrue,
		Segment:         &obj,
		ResourceType:    "ChildSegment",
	}
	dataValue, errors := converter.ConvertToVapi(childObj, model.ChildSegmentBindingType())
	if errors != nil {
		return fmt.Errorf("Error converting Child Segment: %v", errors[0])
	}

	if isFixed {
		dataValue, err := nsxtPolicySegmentAddGatewayToInfraStruct(d, dataValue.(*data.StructValue))
		if err != nil {
			return err
		}
		infraChildren = append(infraChildren, dataValue)

	} else {
		infraChildren = append(infraChildren, dataValue.(*data.StructValue))

	}

	infraType := "Infra"
	infraObj := model.Infra{
		Children:     infraChildren,
		ResourceType: &infraType,
	}

	log.Printf("[DEBUG] Using H-API to delete segment with ID %s", id)
	err := policyInfraPatch(getSessionContext(d, m), infraObj, getPolicyConnector(m), false)
	if err != nil {
		return handleDeleteError("Segment", id, err)
	}
	log.Printf("[DEBUG] Success deleting Segment with ID %s", id)

	return nil
}

func parseSegmentPolicyPath(path string) (bool, string, string) {
	segs := strings.Split(path, "/")
	if (len(segs) < 3) || (segs[len(segs)-2] != "segments") {
		// error - this is not a segment path
		return false, "", ""
	}

	segmentID := segs[len(segs)-1]

	if len(segs) == 4 {
		// This is an infra segment
		return false, "", segmentID
	} else if segs[1] == "orgs" && segs[3] == "projects" && len(segs) == 8 {
		return false, "", segmentID
	}

	// This is a fixed segment
	gwPath := path[0:strings.Index(path, "/segments/")]

	isT0, gwID := parseGatewayPolicyPath(gwPath)
	return isT0, gwID, segmentID
}
