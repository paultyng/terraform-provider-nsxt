// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	nsx_policy "github.com/vmware/terraform-provider-nsxt/api"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	global_policy "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var nsxtPolicyTier0GatewayRedistributionRuleTypes = []string{
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_STATIC,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_CONNECTED,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_EXTERNAL_INTERFACE,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_SEGMENT,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_ROUTER_LINK,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_SERVICE_INTERFACE,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_LOOPBACK_INTERFACE,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_DNS_FORWARDER_IP,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_IPSEC_LOCAL_IP,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_NAT,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER0_EVPN_TEP_IP,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_NAT,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_STATIC,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_LB_VIP,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_LB_SNAT,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_DNS_FORWARDER_IP,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_CONNECTED,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_SERVICE_INTERFACE,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_SEGMENT,
	model.Tier0RouteRedistributionRule_ROUTE_REDISTRIBUTION_TYPES_TIER1_IPSEC_LOCAL_ENDPOINT,
}

func getFailoverModeSchema(defaultValue string) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Failover mode",
		Default:      defaultValue,
		Optional:     true,
		ValidateFunc: validation.StringInSlice(policyFailOverModeValues, false),
	}
}

func getIPv6NDRAPathSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The path of an IPv6 NDRA profile",
		Optional:    true,
		Computed:    true,
	}
}

func getIPv6DadPathSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The path of an IPv6 DAD profile",
		Optional:    true,
		Computed:    true,
	}
}

func getPolicyEdgeClusterPathSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The path of the edge cluster connected to this gateway",
		Optional:     true,
		ValidateFunc: validatePolicyPath(),
		Computed:     true,
	}
}

func getPolicyLocaleServiceSchema(isTier1 bool) *schema.Schema {
	nodeConficts := []string{}
	if isTier1 {
		// for Tier1, enable_standby_relocation can not be enabled if
		// preferred nodes are specified
		nodeConficts = append(nodeConficts, "enable_standby_relocation")
	}

	elemSchema := map[string]*schema.Schema{
		"nsx_id": getNsxIDSchema(),
		"edge_cluster_path": {
			Type:         schema.TypeString,
			Description:  "The path of the edge cluster connected to this gateway",
			Required:     true,
			ValidateFunc: validatePolicyPath(),
		},
		"preferred_edge_paths": {
			Type:          schema.TypeList,
			Description:   "Paths of specific edge nodes",
			Optional:      true,
			Elem:          getElemPolicyPathSchema(),
			ConflictsWith: nodeConficts,
		},
		"redistribution_config": getRedistributionConfigSchema(),
		"path":                  getPathSchema(),
		"revision":              getRevisionSchema(),
		"display_name":          getComputedDisplayNameSchema(),
	}
	if isTier1 {
		delete(elemSchema, "redistribution_config")
	}

	result := &schema.Schema{
		Type:          schema.TypeSet,
		Optional:      true,
		Description:   "Locale Service for the gateway",
		ConflictsWith: []string{"edge_cluster_path"},
		Elem: &schema.Resource{
			Schema: elemSchema,
		},
	}

	return result
}

func getIpv6ProfilePathsFromSchema(d *schema.ResourceData) []string {
	var profiles []string
	if d.Get("ipv6_ndra_profile_path") != "" {
		profiles = append(profiles, d.Get("ipv6_ndra_profile_path").(string))
	}
	if d.Get("ipv6_dad_profile_path") != "" {
		profiles = append(profiles, d.Get("ipv6_dad_profile_path").(string))
	}
	return profiles
}

func setIpv6ProfilePathsInSchema(d *schema.ResourceData, paths []string) error {
	for _, path := range paths {
		tokens := strings.Split(path, "/")
		// First token is empty, second is infra, global-infra or orgs, third is profile type
		if len(tokens) < 4 {
			return fmt.Errorf("Unexpected ipv6 profile path: %s", path)
		}
		tokenOffset := 2
		if tokens[1] == "orgs" && tokens[3] == "projects" {
			// It's a multitenancy environment
			tokenOffset = 6
		}
		if tokens[tokenOffset] == "ipv6-ndra-profiles" {
			d.Set("ipv6_ndra_profile_path", path)
		}
		if tokens[tokenOffset] == "ipv6-dad-profiles" {
			d.Set("ipv6_dad_profile_path", path)
		}
	}
	return nil
}

func getGatewayInterfaceSubnetsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of IP addresses and network prefixes for this interface",
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateIPCidr(),
		},
		Required: true,
	}
}

func getMtuSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeInt,
		Optional:     true,
		Description:  "Maximum transmission unit specifies the size of the largest packet that a network protocol can transmit",
		ValidateFunc: validation.IntAtLeast(64),
	}
}

var gatewayInterfaceUrpfModeValues = []string{
	model.Tier0Interface_URPF_MODE_NONE,
	model.Tier0Interface_URPF_MODE_STRICT,
}

func getGatewayInterfaceUrpfModeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Unicast Reverse Path Forwarding mode",
		ValidateFunc: validation.StringInSlice(gatewayInterfaceUrpfModeValues, false),
		Default:      model.Tier0Interface_URPF_MODE_STRICT,
	}
}

func getGatewayIntersiteConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    true,
		Description: "Locale Service for the gateway",
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"transit_subnet": {
					Type:         schema.TypeString,
					Description:  "IPv4 subnet for inter-site transit segment connecting service routers across sites for stretched gateway. For IPv6 link local subnet is auto configured",
					Computed:     true,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validateCidr(),
				},
				"primary_site_path": {
					Type:         schema.TypeString,
					Description:  "Primary egress site for gateway",
					Optional:     true,
					ValidateFunc: validatePolicyPath(),
				},
				"fallback_site_paths": {
					Type:        schema.TypeSet,
					Description: "Fallback sites to be used as new primary site on current primary site failure",
					Optional:    true,
					Elem:        getElemPolicyPathSchema(),
				},
			},
		},
	}
}

func listPolicyGatewayLocaleServices(context utl.SessionContext, connector client.Connector, gwID string, listLocaleServicesFunc func(utl.SessionContext, client.Connector, string, *string) (model.LocaleServicesListResult, error)) ([]model.LocaleServices, error) {
	var results []model.LocaleServices
	var cursor *string
	var count int64
	total := int64(0)

	for {
		listResponse, err := listLocaleServicesFunc(context, connector, gwID, cursor)
		if err != nil {
			return results, err
		}
		cursor = listResponse.Cursor
		count = *listResponse.ResultCount
		results = append(results, listResponse.Results...)
		if total == 0 {
			// first response
			total = count
		}
		if int64(len(results)) >= total {
			return results, nil
		}
	}
}

func getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices []model.LocaleServices, sitePath string, gatewayID string) (string, error) {
	if len(localeServices) == 0 {
		return "", fmt.Errorf("Edge cluster is mandatory on gateway %s in order to create interfaces", gatewayID)
	}
	for _, localeService := range localeServices {
		if localeService.EdgeClusterPath != nil {
			if strings.HasPrefix(*localeService.EdgeClusterPath, sitePath) {
				localeServiceID := *localeService.Id
				return localeServiceID, nil
			}
		}
	}
	return "", fmt.Errorf("Edge cluster is mandatory on gateway %s in order to create interfaces", gatewayID)
}

func initChildLocaleService(serviceStruct *model.LocaleServices, markForDelete bool) (*data.StructValue, error) {
	childService := model.ChildLocaleServices{
		ResourceType:    "ChildLocaleServices",
		LocaleServices:  serviceStruct,
		MarkedForDelete: &markForDelete,
	}

	converter := bindings.NewTypeConverter()
	dataValue, err := converter.ConvertToVapi(childService, model.ChildLocaleServicesBindingType())

	if err != nil {
		return nil, err[0]
	}

	return dataValue.(*data.StructValue), nil
}

func initGatewayLocaleServices(context utl.SessionContext, d *schema.ResourceData, connector client.Connector, listLocaleServicesFunc func(utl.SessionContext, client.Connector, string) ([]model.LocaleServices, error)) ([]*data.StructValue, error) {
	var localeServices []*data.StructValue

	services := d.Get("locale_service").(*schema.Set).List()

	existingServices := make(map[string]model.LocaleServices)
	if len(d.Id()) > 0 {
		// This is an update - we might need to delete locale services
		existingServiceObjects, errList := listLocaleServicesFunc(context, connector, d.Id())
		if errList != nil {
			return nil, errList
		}

		for _, obj := range existingServiceObjects {
			existingServices[*obj.Id] = obj
		}
	}
	lsType := "LocaleServices"
	idMap := make(map[string]bool)
	for _, service := range services {
		cfg := service.(map[string]interface{})
		edgeClusterPath := cfg["edge_cluster_path"].(string)
		edgeNodes := interfaceListToStringList(cfg["preferred_edge_paths"].([]interface{}))
		path := cfg["path"].(string)
		nsxID := cfg["nsx_id"].(string)
		// validate unique ids are provided for services
		if len(nsxID) > 0 {
			if _, ok := idMap[nsxID]; ok {
				return nil, fmt.Errorf("Duplicate nsx_id for locale_service %s - please specify unique id", nsxID)
			}
			idMap[nsxID] = true
		}

		var serviceID string
		if nsxID != "" {
			log.Printf("[DEBUG] Updating locale service %s", path)
			serviceID = nsxID
		} else {
			serviceID = newUUID()
			log.Printf("[DEBUG] Preparing to create locale service %s for gateway %s", serviceID, d.Id())
		}
		serviceStruct := model.LocaleServices{
			Id:                 &serviceID,
			ResourceType:       &lsType,
			EdgeClusterPath:    &edgeClusterPath,
			PreferredEdgePaths: edgeNodes,
		}

		redistribution := cfg["redistribution_config"]
		if redistribution != nil {
			redistributionConfigs := redistribution.([]interface{})
			if len(redistributionConfigs) > 0 {
				setLocaleServiceRedistributionConfig(redistributionConfigs, &serviceStruct)
				d.Set("redistribution_set", true)
			} else {
				d.Set("redistribution_set", false)
			}
		}

		if obj, ok := existingServices[serviceID]; ok {
			// if this is an update for existing locale service,
			// we need revision, and keep the HA vip config
			serviceStruct.Revision = obj.Revision
			serviceStruct.HaVipConfigs = obj.HaVipConfigs
		}
		dataValue, err := initChildLocaleService(&serviceStruct, false)
		if err != nil {
			return localeServices, err
		}

		localeServices = append(localeServices, dataValue)
		delete(existingServices, serviceID)
	}
	// Add instruction to delete services that are no longer present in intent
	for id := range existingServices {
		serviceStruct := model.LocaleServices{
			Id:           &id,
			ResourceType: &lsType,
		}
		dataValue, err := initChildLocaleService(&serviceStruct, true)
		if err != nil {
			return localeServices, err
		}
		localeServices = append(localeServices, dataValue)
		log.Printf("[DEBUG] Preparing to delete locale service %s for gateway %s", id, d.Id())
	}

	return localeServices, nil
}

func getPolicyGatewayIntersiteConfigFromSchema(d *schema.ResourceData) *model.IntersiteGatewayConfig {
	cfg, isSet := d.GetOk("intersite_config")
	if !isSet {
		return nil
	}

	configs := cfg.([]interface{})
	for _, elem := range configs {
		data := elem.(map[string]interface{})

		subnet := data["transit_subnet"].(string)
		primarySitePath := data["primary_site_path"].(string)
		fallbackSites := interface2StringList(data["fallback_site_paths"].(*schema.Set).List())
		intersiteConfig := model.IntersiteGatewayConfig{
			FallbackSites: fallbackSites,
		}
		if len(primarySitePath) > 0 {
			intersiteConfig.PrimarySitePath = &primarySitePath
		}
		if len(subnet) > 0 {
			intersiteConfig.IntersiteTransitSubnet = &subnet
		}

		return &intersiteConfig
	}

	return nil
}

func setPolicyGatewayIntersiteConfigInSchema(d *schema.ResourceData, config *model.IntersiteGatewayConfig) error {
	if config == nil {
		return nil
	}

	var result []map[string]interface{}
	elem := make(map[string]interface{})

	if config.IntersiteTransitSubnet != nil {
		elem["transit_subnet"] = config.IntersiteTransitSubnet
	}

	if config.PrimarySitePath != nil {
		elem["primary_site_path"] = config.PrimarySitePath
	}

	elem["fallback_site_paths"] = config.FallbackSites
	result = append(result, elem)

	return d.Set("intersite_config", result)
}

func policyInfraPatch(context utl.SessionContext, obj model.Infra, connector client.Connector, enforceRevision bool) error {
	if context.ClientType == utl.Global {
		infraClient := global_policy.NewGlobalInfraClient(connector)
		gmObj, err := convertModelBindingType(obj, model.InfraBindingType(), gm_model.InfraBindingType())
		if err != nil {
			return err
		}

		return infraClient.Patch(gmObj.(gm_model.Infra), &enforceRevision)
	} else if context.ClientType == utl.VPC {
		context = utl.SessionContext{
			ClientType: utl.Multitenancy,
			ProjectID:  context.ProjectID,
		}
	}

	infraClient := nsx_policy.NewInfraClient(context, connector)
	if infraClient == nil {
		return policyResourceNotSupportedError()
	}
	return infraClient.Patch(obj, &enforceRevision)
}

func getRedistributionConfigRuleSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of routes to be aggregated",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Description: "Rule name",
					Optional:    true,
				},
				"route_map_path": {
					Type:        schema.TypeString,
					Description: "Route map to be associated with the redistribution rule",
					Optional:    true,
				},
				"types": {
					Type:        schema.TypeSet,
					Description: "List of redistribution types",
					Optional:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice(nsxtPolicyTier0GatewayRedistributionRuleTypes, false),
					},
				},
				"bgp": {
					Type:        schema.TypeBool,
					Description: "BGP destination for this rule",
					Optional:    true,
					Default:     true,
				},
				"ospf": {
					Type:        schema.TypeBool,
					Description: "OSPF destination for this rule",
					Optional:    true,
					Default:     false,
				},
			},
		},
	}
}

func getRedistributionConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Route Redistribution configuration",
		Optional:    true,
		MaxItems:    1,
		Computed:    true,
		Deprecated:  "Use nsxt_policy_gateway_redistribution_config resource instead",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Flag to enable route redistribution for BGP",
					Optional:    true,
					Default:     true,
				},
				"ospf_enabled": {
					Type:        schema.TypeBool,
					Description: "Flag to enable route redistribution for OSPF",
					Optional:    true,
					Default:     false,
				},
				"rule": getRedistributionConfigRuleSchema(),
			},
		},
	}
}

func setLocaleServiceRedistributionRulesConfig(rulesConfig []interface{}, config *model.Tier0RouteRedistributionConfig) {
	var rules []model.Tier0RouteRedistributionRule
	for _, ruleConfig := range rulesConfig {
		data := ruleConfig.(map[string]interface{})
		name := data["name"].(string)
		routeMapPath := data["route_map_path"].(string)
		types := data["types"].(*schema.Set).List()
		bgp := data["bgp"].(bool)
		ospf := data["ospf"].(bool)

		rule := model.Tier0RouteRedistributionRule{
			RouteRedistributionTypes: interface2StringList(types),
		}

		if len(name) > 0 {
			rule.Name = &name
		}

		if len(routeMapPath) > 0 {
			rule.RouteMapPath = &routeMapPath
		}

		if util.NsxVersionHigherOrEqual("3.1.0") {
			if bgp {
				rule.Destinations = append(rule.Destinations, model.Tier0RouteRedistributionRule_DESTINATIONS_BGP)
			}
			if ospf {
				rule.Destinations = append(rule.Destinations, model.Tier0RouteRedistributionRule_DESTINATIONS_OSPF)
			}
		}

		rules = append(rules, rule)
	}

	if len(rules) > 0 {
		config.RedistributionRules = rules
	}
}

func setLocaleServiceRedistributionConfig(redistributionConfigs []interface{}, serviceStruct *model.LocaleServices) {
	if len(redistributionConfigs) == 0 {
		return
	}

	redistributionConfig := redistributionConfigs[0].(map[string]interface{})
	bgpEnabled := redistributionConfig["enabled"].(bool)
	ospfEnabled := redistributionConfig["ospf_enabled"].(bool)
	rulesConfig := redistributionConfig["rule"].([]interface{})

	redistributionStruct := model.Tier0RouteRedistributionConfig{
		BgpEnabled: &bgpEnabled,
	}

	if util.NsxVersionHigherOrEqual("3.1.0") {
		redistributionStruct.OspfEnabled = &ospfEnabled
	}

	setLocaleServiceRedistributionRulesConfig(rulesConfig, &redistributionStruct)
	serviceStruct.RouteRedistributionConfig = &redistributionStruct
}

func getLocaleServiceRedistributionRuleConfig(config *model.Tier0RouteRedistributionConfig) []map[string]interface{} {
	var rules []map[string]interface{}
	for _, ruleConfig := range config.RedistributionRules {
		rule := make(map[string]interface{})
		rule["name"] = ruleConfig.Name
		rule["route_map_path"] = ruleConfig.RouteMapPath
		rule["types"] = ruleConfig.RouteRedistributionTypes
		if util.NsxVersionHigherOrEqual("3.1.0") {
			bgp := false
			ospf := false
			for _, destination := range ruleConfig.Destinations {
				if destination == model.Tier0RouteRedistributionRule_DESTINATIONS_BGP {
					bgp = true
				}
				if destination == model.Tier0RouteRedistributionRule_DESTINATIONS_OSPF {
					ospf = true
				}
			}
			rule["bgp"] = bgp
			rule["ospf"] = ospf
		}

		rules = append(rules, rule)
	}

	return rules
}

func getLocaleServiceRedistributionConfig(serviceStruct *model.LocaleServices) []map[string]interface{} {
	var redistributionConfigs []map[string]interface{}
	config := serviceStruct.RouteRedistributionConfig
	if config == nil {
		return redistributionConfigs
	}

	elem := make(map[string]interface{})
	elem["enabled"] = config.BgpEnabled
	elem["ospf_enabled"] = config.OspfEnabled
	elem["rule"] = getLocaleServiceRedistributionRuleConfig(config)
	redistributionConfigs = append(redistributionConfigs, elem)
	return redistributionConfigs
}

func findTier0LocaleServiceForSite(context utl.SessionContext, connector client.Connector, gwID string, sitePath string) (string, error) {
	localeServices, err := listPolicyTier0GatewayLocaleServices(context, connector, gwID)
	if err != nil {
		return "", err
	}
	return getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices, sitePath, gwID)
}

func parseGatewayInterfacePolicyPath(path string) (bool, string, string, string) {
	// interface path must be /*infra/tier-Xs/gw-id/locale-services/ls-id/interfaces/if-id
	segs := strings.Split(path, "/")
	if (len(segs) != 8) || (segs[len(segs)-2] != "interfaces") {
		// error - this is not an interface path
		return false, "", "", ""
	}

	isT0 := true
	if segs[2] != "tier-0s" {
		isT0 = false
	}

	gwID := segs[3]
	localeServiceID := segs[5]
	interfaceID := segs[7]

	return isT0, gwID, localeServiceID, interfaceID
}

func getComputedLocaleServiceIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "NSX ID of associated Gateway Locale Service",
		Computed:    true,
	}
}

func getComputedGatewayIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "NSX ID of associated Tier0 Gateway",
		Computed:    true,
	}
}

func policyTier0GetLocaleService(gwID string, localeServiceID string, connector client.Connector, isGlobalManager bool) *model.LocaleServices {
	if isGlobalManager {
		nsxClient := gm_tier0s.NewLocaleServicesClient(connector)
		gmObj, err := nsxClient.Get(gwID, localeServiceID)
		if err != nil {
			log.Printf("[DEBUG] Failed to get locale service %s for gateway %s: %s", gwID, localeServiceID, err)
			return nil
		}

		convObj, convErr := convertModelBindingType(gmObj, gm_model.LocaleServicesBindingType(), model.LocaleServicesBindingType())
		if convErr != nil {
			log.Printf("[DEBUG] Failed to convert locale service %s for gateway %s: %s", gwID, localeServiceID, convErr)
			return nil
		}
		obj := convObj.(model.LocaleServices)
		return &obj
	}
	nsxClient := tier_0s.NewLocaleServicesClient(connector)
	obj, err := nsxClient.Get(gwID, localeServiceID)
	if err != nil {
		log.Printf("[DEBUG] Failed to get locale service %s for gateway %s: %s", gwID, localeServiceID, err)
		return nil
	}
	return &obj
}
