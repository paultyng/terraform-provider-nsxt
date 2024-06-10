/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_locale_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/locale_services"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var gatewayInterfaceTypeValues = []string{
	model.Tier0Interface_TYPE_SERVICE,
	model.Tier0Interface_TYPE_EXTERNAL,
	model.Tier0Interface_TYPE_LOOPBACK,
}

var gatewayInterfaceOspfNetworkTypeValues = []string{
	model.PolicyInterfaceOspfConfig_NETWORK_TYPE_BROADCAST,
	model.PolicyInterfaceOspfConfig_NETWORK_TYPE_P2P,
}

func resourceNsxtPolicyTier0GatewayInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyTier0GatewayInterfaceCreate,
		Read:   resourceNsxtPolicyTier0GatewayInterfaceRead,
		Update: resourceNsxtPolicyTier0GatewayInterfaceUpdate,
		Delete: resourceNsxtPolicyTier0GatewayInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0GatewayInterfaceImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":                 getNsxIDSchema(),
			"path":                   getPathSchema(),
			"display_name":           getDisplayNameSchema(),
			"description":            getDescriptionSchema(),
			"revision":               getRevisionSchema(),
			"tag":                    getTagsSchema(),
			"gateway_path":           getPolicyPathSchema(true, true, "Policy path for Tier0 gateway"),
			"segment_path":           getPolicyPathSchema(false, true, "Policy path for connected segment"),
			"subnets":                getGatewayInterfaceSubnetsSchema(),
			"mtu":                    getMtuSchema(),
			"ipv6_ndra_profile_path": getIPv6NDRAPathSchema(),
			"type": {
				Type:         schema.TypeString,
				Description:  "Interface Type",
				ValidateFunc: validation.StringInSlice(gatewayInterfaceTypeValues, false),
				Optional:     true,
				ForceNew:     true,
				Default:      model.Tier0Interface_TYPE_EXTERNAL,
			},
			"edge_node_path": getPolicyPathSchema(false, false, "Policy path for edge node"),
			"enable_pim": {
				Type:        schema.TypeBool,
				Description: "Enable Protocol Independent Multicast on Interface, applicable only when interface type is EXTERNAL",
				Optional:    true,
				Default:     false,
			},
			"access_vlan_id": {
				Type:         schema.TypeInt,
				Description:  "Vlan ID",
				Optional:     true,
				ValidateFunc: validateVLANId,
			},
			"locale_service_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Gateway Locale Service on NSX",
				Computed:    true,
			},
			"urpf_mode": getGatewayInterfaceUrpfModeSchema(),
			"ip_addresses": {
				Type:        schema.TypeList,
				Description: "Ip addresses",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path of the site the Tier0 edge cluster belongs to",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"ospf": getGatewayInterfaceOspfSchema(),
		},
	}
}

func getGatewayInterfaceOspfSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "OSPF configuration for the interface",
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"area_path": getPolicyPathSchema(true, false, "OSPF Area Path"),
				"enable_bfd": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"bfd_profile_path": getPolicyPathSchema(false, false, "BFD profile path to be applied to all OSPF peers in this interface"),
				"hello_interval": {
					Type:        schema.TypeInt,
					Description: "Interval in seconds between hello packets that OSPF sends on this interface",
					Optional:    true,
					Default:     10,
				},
				"dead_interval": {
					Type:        schema.TypeInt,
					Description: "Number of seconds that router must wait before it declares OSPF neighbor router as down",
					Optional:    true,
					Default:     40,
				},
				"network_type": {
					Type:         schema.TypeString,
					Description:  "OSPF network type",
					ValidateFunc: validation.StringInSlice(gatewayInterfaceOspfNetworkTypeValues, false),
					Optional:     true,
					Default:      model.PolicyInterfaceOspfConfig_NETWORK_TYPE_BROADCAST,
				},
			},
		},
	}
}

func gatewayInterfaceVersionDepenantSet(d *schema.ResourceData, m interface{}, obj *model.Tier0Interface) error {
	if util.NsxVersionLower("3.0.0") {
		return nil
	}
	interfaceType := d.Get("type").(string)
	// PIM config can only be configured on external interface and local manager
	if interfaceType == model.Tier0Interface_TYPE_EXTERNAL && !isPolicyGlobalManager(m) {
		enablePIM := d.Get("enable_pim").(bool)
		pimConfig := model.Tier0InterfacePimConfig{
			Enabled: &enablePIM,
		}
		obj.Multicast = &pimConfig
	}

	vlanID := int64(d.Get("access_vlan_id").(int))
	if vlanID > 0 {
		obj.AccessVlanId = &vlanID
	}

	urpfMode := d.Get("urpf_mode").(string)
	obj.UrpfMode = &urpfMode

	return policyTier0GatewayInterfaceOspfSet(d, m, obj)
}

func policyTier0GatewayInterfaceOspfSet(d *schema.ResourceData, m interface{}, obj *model.Tier0Interface) error {
	ospfConfigs := d.Get("ospf").([]interface{})
	if len(ospfConfigs) == 0 {
		return nil
	}

	if isPolicyGlobalManager(m) {
		return fmt.Errorf("Ospf configuration is not supported on Global Manager")
	}

	interfaceType := d.Get("type").(string)
	if interfaceType != model.Tier0Interface_TYPE_EXTERNAL {
		// Ospf config can not be set for non-external interface
		return fmt.Errorf("Ospf configuration can only be set for EXTERNAL interface")
	}

	ospfConfig := ospfConfigs[0].(map[string]interface{})
	enabled := ospfConfig["enabled"].(bool)
	enableBfd := ospfConfig["enable_bfd"].(bool)
	bfdProfilePath := ospfConfig["bfd_profile_path"].(string)
	areaPath := ospfConfig["area_path"].(string)
	networkType := ospfConfig["network_type"].(string)
	helloInterval := int64(ospfConfig["hello_interval"].(int))
	deadInterval := int64(ospfConfig["dead_interval"].(int))

	ospf := model.PolicyInterfaceOspfConfig{
		Enabled:       &enabled,
		EnableBfd:     &enableBfd,
		OspfArea:      &areaPath,
		NetworkType:   &networkType,
		HelloInterval: &helloInterval,
		DeadInterval:  &deadInterval,
	}

	if bfdProfilePath != "" {
		ospf.BfdPath = &bfdProfilePath
	}

	obj.Ospf = &ospf
	return nil
}

func policyTier0GatewayInterfaceOspfSetInSchema(d *schema.ResourceData, obj *model.Tier0Interface) error {

	if obj.Ospf == nil {
		return nil
	}

	var ospfConfigs []map[string]interface{}
	ospfConfig := make(map[string]interface{})
	ospfConfig["enabled"] = obj.Ospf.Enabled
	ospfConfig["enable_bfd"] = obj.Ospf.EnableBfd
	ospfConfig["bfd_profile_path"] = obj.Ospf.BfdPath
	ospfConfig["area_path"] = obj.Ospf.OspfArea
	ospfConfig["network_type"] = obj.Ospf.NetworkType
	ospfConfig["hello_interval"] = obj.Ospf.HelloInterval
	ospfConfig["dead_interval"] = obj.Ospf.DeadInterval

	ospfConfigs = append(ospfConfigs, ospfConfig)
	return d.Set("ospf", ospfConfigs)
}

func resourceNsxtPolicyTier0GatewayInterfaceCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	tier0Path := d.Get("gateway_path").(string)
	tier0ID := getPolicyIDFromPath(tier0Path)

	segmentPath := d.Get("segment_path").(string)
	objSitePath := d.Get("site_path").(string)
	ifType := d.Get("type").(string)
	if len(segmentPath) == 0 && ifType != model.Tier0Interface_TYPE_LOOPBACK {
		// segment_path in required for all interfaces other than loopback
		return fmt.Errorf("segment_path is mandatory for interface of type %s", ifType)
	}

	localeServiceID := ""
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if isPolicyGlobalManager(m) {
		enablePIM := d.Get("enable_pim").(bool)
		if enablePIM {
			return fmt.Errorf("enable_pim configuration is only supported with NSX Local Manager")
		}
		if objSitePath == "" {
			return attributeRequiredGlobalManagerError("site_path", "nsxt_policy_tier0_gateway_interface")
		}
		localeServices, err := listPolicyTier0GatewayLocaleServices(context, connector, tier0ID)
		if err != nil {
			return err
		}
		localeServiceID, err = getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices, objSitePath, tier0ID)
		if err != nil {
			return err
		}
	} else {
		if objSitePath != "" {
			return globalManagerOnlyError()
		}
		localeService, err := getPolicyTier0GatewayLocaleServiceWithEdgeCluster(context, tier0ID, connector)
		if err != nil {
			return err
		}
		if localeService == nil {
			return fmt.Errorf("Edge cluster is mandatory on gateway %s in order to create interfaces", tier0ID)
		}
		localeServiceID = *localeService.Id
	}

	if id == "" {
		id = newUUID()
	} else {
		var err error
		if isPolicyGlobalManager(m) {
			client := gm_locale_services.NewInterfacesClient(connector)
			_, err = client.Get(tier0ID, localeServiceID, id)

		} else {
			client := locale_services.NewInterfacesClient(connector)
			_, err = client.Get(tier0ID, localeServiceID, id)
		}
		if err == nil {
			return fmt.Errorf("Interface with ID '%s' already exists on Tier0 Gateway %s", id, tier0ID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	interfaceSubnetList := getGatewayInterfaceSubnetList(d)
	var ipv6ProfilePaths []string
	if d.Get("ipv6_ndra_profile_path").(string) != "" {
		ipv6ProfilePaths = append(ipv6ProfilePaths, d.Get("ipv6_ndra_profile_path").(string))
	}
	mtu := int64(d.Get("mtu").(int))
	edgePath := d.Get("edge_node_path").(string)
	obj := model.Tier0Interface{
		Id:               &id,
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		Type_:            &ifType,
		Subnets:          interfaceSubnetList,
		Ipv6ProfilePaths: ipv6ProfilePaths,
	}

	if len(segmentPath) > 0 {
		obj.SegmentPath = &segmentPath
	}

	if mtu > 0 {
		obj.Mtu = &mtu
	}

	if edgePath != "" {
		obj.EdgePath = &edgePath
	}

	err = gatewayInterfaceVersionDepenantSet(d, m, &obj)
	if err != nil {
		return handleCreateError("Tier0 Interface", id, err)
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating Tier0 interface with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, err1 := convertModelBindingType(obj, model.Tier0InterfaceBindingType(), gm_model.Tier0InterfaceBindingType())
		if err1 != nil {
			return err1
		}
		client := gm_locale_services.NewInterfacesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, id, gmObj.(gm_model.Tier0Interface), nil)
	} else {
		client := locale_services.NewInterfacesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, id, obj, nil)
	}

	if err != nil {
		return handleCreateError("Tier0 Interface", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	d.Set("locale_service_id", localeServiceID)

	return resourceNsxtPolicyTier0GatewayInterfaceRead(d, m)
}

func resourceNsxtPolicyTier0GatewayInterfaceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0Path := d.Get("gateway_path").(string)
	tier0ID := getPolicyIDFromPath(tier0Path)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || tier0ID == "" {
		return fmt.Errorf("Error obtaining Tier0 Interface id")
	}

	var obj model.Tier0Interface
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_locale_services.NewInterfacesClient(connector)
		gmObj, err1 := client.Get(tier0ID, localeServiceID, id)
		if err1 != nil {
			return handleReadError(d, "Tier0 Interface", id, err1)
		}
		lmObj, err2 := convertModelBindingType(gmObj, model.Tier0InterfaceBindingType(), model.Tier0InterfaceBindingType())
		if err2 != nil {
			return err2
		}
		obj = lmObj.(model.Tier0Interface)
		tier0Client := gm_tier0s.NewLocaleServicesClient(connector)
		localeService, err := tier0Client.Get(tier0ID, localeServiceID)
		if err != nil {
			return err
		}
		sitePath := getSitePathFromEdgePath(*localeService.EdgeClusterPath)
		d.Set("site_path", sitePath)
	} else {
		client := locale_services.NewInterfacesClient(connector)
		obj, err = client.Get(tier0ID, localeServiceID, id)
	}
	if err != nil {
		return handleReadError(d, "Tier0 Interface", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("segment_path", obj.SegmentPath)
	d.Set("edge_node_path", obj.EdgePath)
	d.Set("type", obj.Type_)

	if obj.Ipv6ProfilePaths != nil {
		d.Set("ipv6_ndra_profile_path", obj.Ipv6ProfilePaths[0]) // only one supported for now
	}
	if obj.Mtu != nil {
		d.Set("mtu", *obj.Mtu)
	}

	if obj.Multicast != nil {
		d.Set("enable_pim", *obj.Multicast.Enabled)
	} else {
		d.Set("enable_pim", false)
	}

	if obj.AccessVlanId != nil {
		d.Set("access_vlan_id", *obj.AccessVlanId)
	}

	if obj.Subnets != nil {
		var subnetList []string
		var ipList []string
		for _, subnet := range obj.Subnets {
			cidr := fmt.Sprintf("%s/%d", subnet.IpAddresses[0], *subnet.PrefixLen)
			subnetList = append(subnetList, cidr)
			ipList = append(ipList, subnet.IpAddresses[0])
		}
		d.Set("subnets", subnetList)
		d.Set("ip_addresses", ipList)
	}

	if obj.UrpfMode != nil {
		d.Set("urpf_mode", obj.UrpfMode)
	} else {
		// assign default for version that is lower than 3.0.0
		d.Set("urpf_mode", model.Tier0Interface_URPF_MODE_STRICT)
	}

	return policyTier0GatewayInterfaceOspfSetInSchema(d, &obj)
}

func resourceNsxtPolicyTier0GatewayInterfaceUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0Path := d.Get("gateway_path").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	tier0ID := getPolicyIDFromPath(tier0Path)
	if id == "" || tier0ID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	if isPolicyGlobalManager(m) {
		enablePIM := d.Get("enable_pim").(bool)
		if enablePIM {
			return fmt.Errorf("enable_pim configuration is only supported with NSX Local Manager")
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	interfaceSubnetList := getGatewayInterfaceSubnetList(d)
	segmentPath := d.Get("segment_path").(string)
	var ipv6ProfilePaths []string
	if d.Get("ipv6_ndra_profile_path").(string) != "" {
		ipv6ProfilePaths = append(ipv6ProfilePaths, d.Get("ipv6_ndra_profile_path").(string))
	}
	mtu := int64(d.Get("mtu").(int))
	ifType := d.Get("type").(string)
	edgePath := d.Get("edge_node_path").(string)
	revision := int64(d.Get("revision").(int))
	obj := model.Tier0Interface{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		Type_:            &ifType,
		Subnets:          interfaceSubnetList,
		Ipv6ProfilePaths: ipv6ProfilePaths,
		Revision:         &revision,
	}

	if mtu > 0 {
		obj.Mtu = &mtu
	}

	if edgePath != "" {
		obj.EdgePath = &edgePath
	}

	if len(segmentPath) > 0 {
		obj.SegmentPath = &segmentPath
	}

	err := gatewayInterfaceVersionDepenantSet(d, m, &obj)
	if err != nil {
		return handleUpdateError("Tier0 Interface", id, err)
	}

	if isPolicyGlobalManager(m) {
		gmObj, err1 := convertModelBindingType(obj, model.Tier0InterfaceBindingType(), gm_model.Tier0InterfaceBindingType())
		if err1 != nil {
			return err1
		}
		client := gm_locale_services.NewInterfacesClient(connector)
		_, err = client.Update(tier0ID, localeServiceID, id, gmObj.(gm_model.Tier0Interface), nil)
	} else {
		client := locale_services.NewInterfacesClient(connector)
		_, err = client.Update(tier0ID, localeServiceID, id, obj, nil)
	}
	if err != nil {
		return handleUpdateError("Tier0 Interface", id, err)
	}

	return resourceNsxtPolicyTier0GatewayInterfaceRead(d, m)
}

func resourceNsxtPolicyTier0GatewayInterfaceDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0Path := d.Get("gateway_path").(string)
	tier0ID := getPolicyIDFromPath(tier0Path)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || tier0ID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_locale_services.NewInterfacesClient(connector)
		err = client.Delete(tier0ID, localeServiceID, id, nil)
	} else {
		client := locale_services.NewInterfacesClient(connector)
		err = client.Delete(tier0ID, localeServiceID, id, nil)
	}
	if err != nil {
		return handleDeleteError("Tier0 Interface", id, err)
	}

	return nil
}

func resourceNsxtPolicyTier0GatewayInterfaceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 3 {
		return nil, fmt.Errorf("Please provide <gateway-id>/<locale-service-id>/<interface-id> as an input")
	}

	gwID := s[0]
	connector := getPolicyConnector(m)
	var tier0GW model.Tier0
	if isPolicyGlobalManager(m) {
		client := gm_infra.NewTier0sClient(connector)
		gmObj, err1 := client.Get(gwID)
		if err1 != nil {
			return nil, err1
		}

		convertedObj, err2 := convertModelBindingType(gmObj, model.Tier0BindingType(), model.Tier0BindingType())
		if err2 != nil {
			return nil, err2
		}

		tier0GW = convertedObj.(model.Tier0)
	} else {
		client := infra.NewTier0sClient(connector)
		var err error
		tier0GW, err = client.Get(gwID)
		if err != nil {
			return nil, err
		}
	}

	d.Set("gateway_path", tier0GW.Path)
	d.Set("locale_service_id", s[1])

	d.SetId(s[2])

	return []*schema.ResourceData{d}, nil
}
