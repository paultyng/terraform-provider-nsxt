// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_bgp "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/locale_services/bgp"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services/bgp"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var bgpNeighborConfigGracefulRestartModeValues = []string{
	model.BgpNeighborConfig_GRACEFUL_RESTART_MODE_HELPER_ONLY,
	model.BgpNeighborConfig_GRACEFUL_RESTART_MODE_GR_AND_HELPER,
	model.BgpNeighborConfig_GRACEFUL_RESTART_MODE_DISABLE,
}
var bgpNeighborConfigRouteFilteringAddressFamilyValues = []string{
	model.BgpRouteFiltering_ADDRESS_FAMILY_IPV4,
	model.BgpRouteFiltering_ADDRESS_FAMILY_IPV6,
	model.BgpRouteFiltering_ADDRESS_FAMILY_L2VPN_EVPN,
}

func resourceNsxtPolicyBgpNeighbor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyBgpNeighborCreate,
		Read:   resourceNsxtPolicyBgpNeighborRead,
		Update: resourceNsxtPolicyBgpNeighborUpdate,
		Delete: resourceNsxtPolicyBgpNeighborDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyBgpNeighborImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"bgp_path":     getPolicyPathSchema(true, true, "Policy path to the BGP for this neighbor"),
			"allow_as_in": {
				Description: "Flag to enable allowas_in option for BGP neighbor",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"bfd_config": {
				Type:        schema.TypeList,
				Description: "BFD configuration for failure detection",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Flag to enable/disable BFD configuration",
						},
						"interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      500,
							Description:  "Time interval between heartbeat packets in milliseconds",
							ValidateFunc: validation.IntBetween(50, 60000),
						},
						"multiple": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							Description:  "Number of times heartbeat packet is missed before BFD declares the neighbor is down",
							ValidateFunc: validation.IntBetween(2, 16),
						},
					},
				},
			},
			"graceful_restart_mode": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(bgpNeighborConfigGracefulRestartModeValues, false),
				Optional:     true,
				Description:  "BGP Graceful Restart Configuration Mode",
				Default:      model.BgpNeighborConfig_GRACEFUL_RESTART_MODE_HELPER_ONLY,
			},
			"hold_down_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      180,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "Wait time in seconds before declaring peer dead",
			},
			"keep_alive_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(1, 65535),
				Description:  "Interval between keep alive messages sent to peer",
			},
			"maximum_hop_limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 255),
				Description:  "Maximum number of hops allowed to reach BGP neighbor",
			},
			"neighbor_address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Neighbor IP Address",
				ValidateFunc: validateSingleIP(),
			},
			"password": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Password for BGP neighbor authentication",
				ValidateFunc: validation.StringLenBetween(0, 20),
				Sensitive:    true,
			},
			"remote_as_num": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ASN of the neighbor in ASPLAIN or ASDOT Format",
				ValidateFunc: validateASPlainOrDot,
			},
			"source_addresses": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Source IP Addresses for BGP peering",
				MaxItems:    8,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
			},
			"route_filtering": {
				Type:        schema.TypeList,
				Description: "Enable address families and route filtering in each direction",
				Optional:    true,
				Computed:    true,
				MaxItems:    2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_family": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Address family type",
							ValidateFunc: validation.StringInSlice(bgpNeighborConfigRouteFilteringAddressFamilyValues, false),
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Flag to enable/disable address family",
						},
						"in_route_filter": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Prefix-list or route map path for IN direction",
							ValidateFunc: validatePolicyPath(),
						},
						"maximum_routes": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(1, 1000000),
							Description:  "Maximum number of routes for the address family",
						},
						"out_route_filter": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Prefix-list or route map path for OUT direction",
							ValidateFunc: validatePolicyPath(),
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyBgpNeighborParseIDs(bgpPath string) (string, string) {
	t0ID := getResourceIDFromResourcePath(bgpPath, "tier-0s")
	lsID := getResourceIDFromResourcePath(bgpPath, "locale-services")
	return t0ID, lsID
}

func resourceNsxtPolicyBgpNeighborExists(t0ID string, localeServiceID string, neighborID string, isGlobalManager bool, connector client.Connector) (bool, error) {

	var err error
	if isGlobalManager {
		client := gm_bgp.NewNeighborsClient(connector)
		_, err = client.Get(t0ID, localeServiceID, neighborID)
	} else {
		client := bgp.NewNeighborsClient(connector)
		_, err = client.Get(t0ID, localeServiceID, neighborID)
	}
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyBgpNeighborResourceDataToStruct(d *schema.ResourceData, id string) (model.BgpNeighborConfig, error) {
	var neighborStruct model.BgpNeighborConfig

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	allowAsIn := d.Get("allow_as_in").(bool)
	gracefulRestartMode := d.Get("graceful_restart_mode").(string)
	holdDownTime := int64(d.Get("hold_down_time").(int))
	keepAliveTime := int64(d.Get("keep_alive_time").(int))
	maximumHopLimit := int64(d.Get("maximum_hop_limit").(int))
	neighborAddress := d.Get("neighbor_address").(string)
	password := d.Get("password").(string)
	remoteAsNum := d.Get("remote_as_num").(string)
	sourceAddresses := interface2StringList(d.Get("source_addresses").([]interface{}))

	var bfdConfig *model.BgpBfdConfig
	for _, bfd := range d.Get("bfd_config").([]interface{}) {
		data := bfd.(map[string]interface{})
		enabled := data["enabled"].(bool)
		interval := int64(data["interval"].(int))
		multiple := int64(data["multiple"].(int))
		bfdConfig = &model.BgpBfdConfig{
			Enabled:  &enabled,
			Interval: &interval,
			Multiple: &multiple,
		}
		break
	}

	var rFilters []model.BgpRouteFiltering
	routeFiltering := d.Get("route_filtering").([]interface{})
	if len(routeFiltering) > 1 && util.NsxVersionLower("3.0.0") {
		return neighborStruct, fmt.Errorf("Only 1 element for 'route_filtering' is supported with NSX-T versions up to 3.0.0")
	}
	for _, filter := range routeFiltering {
		data := filter.(map[string]interface{})
		addrFamily := data["address_family"].(string)
		if addrFamily == model.BgpRouteFiltering_ADDRESS_FAMILY_L2VPN_EVPN && util.NsxVersionLower("3.0.0") {
			return neighborStruct, fmt.Errorf("'%s' is not supported for 'address_family' with NSX-T versions less than 3.0.0", model.BgpRouteFiltering_ADDRESS_FAMILY_L2VPN_EVPN)
		}
		enabled := data["enabled"].(bool)

		filterStruct := model.BgpRouteFiltering{
			AddressFamily: &addrFamily,
			Enabled:       &enabled,
		}

		var inFilters, outFilters []string
		inFilter := data["in_route_filter"].(string)
		outFilter := data["out_route_filter"].(string)
		if len(inFilter) > 0 {
			inFilters = append(inFilters, inFilter)
			filterStruct.InRouteFilters = inFilters
		}
		if len(outFilter) > 0 {
			outFilters = append(outFilters, outFilter)
			filterStruct.OutRouteFilters = outFilters
		}

		if util.NsxVersionHigherOrEqual("3.0.0") && data["maximum_routes"] != 0 {
			maxRoutes := int64(data["maximum_routes"].(int))
			filterStruct.MaximumRoutes = &maxRoutes
		}

		rFilters = append(rFilters, filterStruct)
	}

	neighborStruct = model.BgpNeighborConfig{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		AllowAsIn:           &allowAsIn,
		Bfd:                 bfdConfig,
		GracefulRestartMode: &gracefulRestartMode,
		HoldDownTime:        &holdDownTime,
		KeepAliveTime:       &keepAliveTime,
		MaximumHopLimit:     &maximumHopLimit,
		NeighborAddress:     &neighborAddress,
		RemoteAsNum:         &remoteAsNum,
		RouteFiltering:      rFilters,
		SourceAddresses:     sourceAddresses,
		Id:                  &id,
	}

	if d.HasChange("password") {
		neighborStruct.Password = &password
	}

	return neighborStruct, nil
}

func resourceNsxtPolicyBgpNeighborConvertAndPatch(id string, d *schema.ResourceData, m interface{}) error {
	bgpPath := d.Get("bgp_path").(string)
	t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
	if t0ID == "" || serviceID == "" {
		return fmt.Errorf("Invalid bgp_path %s", bgpPath)
	}

	obj, err := resourceNsxtPolicyBgpNeighborResourceDataToStruct(d, id)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	// Create the resource using PATCH
	log.Printf("[INFO] Creating BgpNeighbor with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.BgpNeighborConfigBindingType(), gm_model.BgpNeighborConfigBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_bgp.NewNeighborsClient(connector)
		err = client.Patch(t0ID, serviceID, id, gmObj.(gm_model.BgpNeighborConfig), nil)

	} else {
		client := bgp.NewNeighborsClient(connector)
		err = client.Patch(t0ID, serviceID, id, obj, nil)
	}
	if err != nil {
		return handleCreateError("BgpNeighbor", id, err)
	}

	return nil
}

func resourceNsxtPolicyBgpNeighborCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	isGlobalManager := isPolicyGlobalManager(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}
	bgpPath := d.Get("bgp_path").(string)
	t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
	if t0ID == "" || serviceID == "" {
		return fmt.Errorf("Invalid bgp_path %s", bgpPath)
	}
	exists, err := resourceNsxtPolicyBgpNeighborExists(t0ID, serviceID, id, isGlobalManager, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("BGP Neighbor with ID %s already exists for Tier-O %s and Locale Service %s", id, t0ID, serviceID)
	}

	err = resourceNsxtPolicyBgpNeighborConvertAndPatch(id, d, m)
	if err != nil {
		return err
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyBgpNeighborRead(d, m)
}

func resourceNsxtPolicyBgpNeighborRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining BgpNeighbor ID")
	}

	bgpPath := d.Get("bgp_path").(string)
	t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
	if t0ID == "" || serviceID == "" {
		return fmt.Errorf("Invalid bgp_path %s", bgpPath)
	}

	var obj model.BgpNeighborConfig
	if isPolicyGlobalManager(m) {
		client := gm_bgp.NewNeighborsClient(connector)
		gmObj, err := client.Get(t0ID, serviceID, id)
		if err != nil {
			return handleReadError(d, "BgpNeighbor", id, err)
		}
		lmObj, err := convertModelBindingType(gmObj, gm_model.BgpNeighborConfigBindingType(), model.BgpNeighborConfigBindingType())
		if err != nil {
			return err
		}
		obj = lmObj.(model.BgpNeighborConfig)

	} else {
		var err error
		client := bgp.NewNeighborsClient(connector)
		obj, err = client.Get(t0ID, serviceID, id)
		if err != nil {
			return handleReadError(d, "BgpNeighbor", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	// NOTE: password is not returned on API responses
	d.Set("allow_as_in", obj.AllowAsIn)
	d.Set("graceful_restart_mode", obj.GracefulRestartMode)
	d.Set("hold_down_time", int(*obj.HoldDownTime))
	d.Set("keep_alive_time", int(*obj.KeepAliveTime))
	d.Set("maximum_hop_limit", int(*obj.MaximumHopLimit))
	d.Set("neighbor_address", obj.NeighborAddress)
	d.Set("remote_as_num", obj.RemoteAsNum)
	d.Set("source_addresses", obj.SourceAddresses)

	var bfdConfigs []interface{}
	if obj.Bfd != nil {
		bfd := make(map[string]interface{})
		bfd["enabled"] = obj.Bfd.Enabled
		bfd["interval"] = int(*obj.Bfd.Interval)
		bfd["multiple"] = int(*obj.Bfd.Multiple)
		bfdConfigs = append(bfdConfigs, bfd)
	}
	d.Set("bfd_config", bfdConfigs)

	var rFilters []interface{}
	for _, filter := range obj.RouteFiltering {
		rf := make(map[string]interface{})
		rf["address_family"] = filter.AddressFamily
		rf["enabled"] = filter.Enabled

		inFilter := ""
		outFilter := ""
		if len(filter.InRouteFilters) > 0 {
			inFilter = filter.InRouteFilters[0]
		}
		if len(filter.OutRouteFilters) > 0 {
			outFilter = filter.OutRouteFilters[0]
		}
		rf["in_route_filter"] = inFilter
		rf["out_route_filter"] = outFilter
		if util.NsxVersionHigherOrEqual("3.0.0") && filter.MaximumRoutes != nil {
			rf["maximum_routes"] = int(*filter.MaximumRoutes)
		}
		rFilters = append(rFilters, rf)
	}
	d.Set("route_filtering", rFilters)

	return nil
}

func resourceNsxtPolicyBgpNeighborUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining BgpNeighbor ID")
	}

	err := resourceNsxtPolicyBgpNeighborConvertAndPatch(id, d, m)
	if err != nil {
		return err
	}

	return resourceNsxtPolicyBgpNeighborRead(d, m)
}

func resourceNsxtPolicyBgpNeighborDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining BgpNeighbor ID")
	}

	connector := getPolicyConnector(m)

	bgpPath := d.Get("bgp_path").(string)
	t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
	if t0ID == "" || serviceID == "" {
		return fmt.Errorf("Invalid bgp_path %s", bgpPath)
	}

	var err error
	if isPolicyGlobalManager(m) {
		client := gm_bgp.NewNeighborsClient(connector)
		err = client.Delete(t0ID, serviceID, id, nil)
	} else {
		client := bgp.NewNeighborsClient(connector)
		err = client.Delete(t0ID, serviceID, id, nil)
	}
	if err != nil {
		return handleDeleteError("BgpNeighbor", id, err)
	}

	return nil
}

func resourceNsxtPolicyBgpNeighborImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 3 {
		return nil, fmt.Errorf("Please provide <tier0-id>/<locale-service-id>/<neighbor-id> as an input")
	}

	tier0ID := s[0]
	serviceID := s[1]
	neighborID := s[2]
	var parentPath string

	connector := getPolicyConnector(m)
	if isPolicyGlobalManager(m) {
		client := gm_bgp.NewNeighborsClient(connector)
		neighbor, err := client.Get(tier0ID, serviceID, neighborID)
		if err != nil {
			return nil, err
		}
		parentPath = *neighbor.ParentPath
	} else {
		client := bgp.NewNeighborsClient(connector)
		neighbor, err := client.Get(tier0ID, serviceID, neighborID)
		if err != nil {
			return nil, err
		}
		parentPath = *neighbor.ParentPath
	}
	d.Set("bgp_path", parentPath)

	d.SetId(neighborID)

	return []*schema.ResourceData{d}, nil
}
