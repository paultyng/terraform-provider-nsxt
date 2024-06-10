/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gm_locale_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/locale_services"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyBgpConfig() *schema.Resource {
	bgpSchema := getPolicyBGPConfigSchema()
	bgpSchema["gateway_path"] = getPolicyPathSchema(true, true, "Gateway for this BGP config")
	bgpSchema["site_path"] = getPolicyPathSchema(false, true, "Site Path for this BGP config")
	bgpSchema["gateway_id"] = getComputedGatewayIDSchema()
	bgpSchema["locale_service_id"] = getComputedLocaleServiceIDSchema()

	return &schema.Resource{
		Create: resourceNsxtPolicyBgpConfigCreate,
		Read:   resourceNsxtPolicyBgpConfigRead,
		Update: resourceNsxtPolicyBgpConfigUpdate,
		Delete: resourceNsxtPolicyBgpConfigDelete,

		Schema: bgpSchema,
	}
}

func resourceNsxtPolicyBgpConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}
	serviceID := d.Get("locale_service_id").(string)
	var lmRoutingConfig model.BgpRoutingConfig
	if isPolicyGlobalManager(m) {

		client := gm_locale_services.NewBgpClient(connector)
		gmObj, err := client.Get(gwID, serviceID)
		if err != nil {
			return handleReadError(d, "BGP Config", serviceID, err)
		}
		lmObj, convErr := convertModelBindingType(gmObj, gm_model.BgpRoutingConfigBindingType(), model.BgpRoutingConfigBindingType())
		if convErr != nil {
			return convErr
		}
		lmRoutingConfig = lmObj.(model.BgpRoutingConfig)
	} else {
		var err error
		client := locale_services.NewBgpClient(connector)
		lmRoutingConfig, err = client.Get(gwID, serviceID)
		if err != nil {
			return handleReadError(d, "BGP Config", serviceID, err)
		}
	}

	data := initPolicyTier0BGPConfigMap(&lmRoutingConfig)

	for key, value := range data {
		d.Set(key, value)
	}

	return nil
}

func resourceNsxtPolicyBgpConfigToStruct(d *schema.ResourceData, isVRF bool) (*model.BgpRoutingConfig, error) {
	ecmp := d.Get("ecmp").(bool)
	enabled := d.Get("enabled").(bool)
	interSrIbgp := d.Get("inter_sr_ibgp").(bool)
	localAsNum := d.Get("local_as_num").(string)
	multipathRelax := d.Get("multipath_relax").(bool)
	restartMode := d.Get("graceful_restart_mode").(string)
	restartTimer := int64(d.Get("graceful_restart_timer").(int))
	staleTimer := int64(d.Get("graceful_restart_stale_route_timer").(int))
	tags := getPolicyTagsFromSchema(d)

	var aggregationStructs []model.RouteAggregationEntry
	routeAggregations := d.Get("route_aggregation").([]interface{})
	if len(routeAggregations) > 0 {
		for _, agg := range routeAggregations {
			data := agg.(map[string]interface{})
			prefix := data["prefix"].(string)
			summary := data["summary_only"].(bool)
			elem := model.RouteAggregationEntry{
				Prefix:      &prefix,
				SummaryOnly: &summary,
			}

			aggregationStructs = append(aggregationStructs, elem)
		}
	}

	routeStruct := model.BgpRoutingConfig{
		Ecmp:              &ecmp,
		Enabled:           &enabled,
		RouteAggregations: aggregationStructs,
		Tags:              tags,
	}

	if len(localAsNum) > 0 {
		routeStruct.LocalAsNum = &localAsNum
	}

	// For BGP on VRF, only limited attributes can be set
	if isVRF {
		vrfError := "%s can not be specified on VRF Gateway"
		if restartTimer != int64(policyBGPGracefulRestartTimerDefault) {
			return &routeStruct, fmt.Errorf(vrfError, "graceful_restart_timer")
		}
		if staleTimer != int64(policyBGPGracefulRestartStaleRouteTimerDefault) {
			return &routeStruct, fmt.Errorf(vrfError, "graceful_restart_stale_route_timer")
		}
		if restartMode != model.BgpGracefulRestartConfig_MODE_HELPER_ONLY {
			return &routeStruct, fmt.Errorf(vrfError, "graceful_restart_mode")
		}
	} else {
		restartTimerStruct := model.BgpGracefulRestartTimer{
			RestartTimer:    &restartTimer,
			StaleRouteTimer: &staleTimer,
		}

		restartConfigStruct := model.BgpGracefulRestartConfig{
			Mode:  &restartMode,
			Timer: &restartTimerStruct,
		}

		routeStruct.InterSrIbgp = &interSrIbgp
		routeStruct.MultipathRelax = &multipathRelax
		routeStruct.GracefulRestartConfig = &restartConfigStruct
	}

	return &routeStruct, nil
}

func resourceNsxtPolicyBgpConfigCreate(d *schema.ResourceData, m interface{}) error {
	// This is not a create operation on NSX, since BGP config us auto created
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}
	sitePath := d.Get("site_path").(string)

	isVrf, err := resourceNsxtPolicyTier0GatewayIsVrf(gwID, connector, isPolicyGlobalManager(m))
	if err != nil {
		return handleCreateError("BgpRoutingConfig", gwID, err)
	}
	obj, err := resourceNsxtPolicyBgpConfigToStruct(d, isVrf)
	if err != nil {
		return handleCreateError("BgpRoutingConfig", gwID, err)
	}

	var localeServiceID string
	if isPolicyGlobalManager(m) {
		serviceID, err1 := findTier0LocaleServiceForSite(context, connector, gwID, sitePath)
		if err1 != nil {
			return handleCreateError("BgpRoutingConfig", gwID, err1)
		}

		localeServiceID = serviceID

		gmObj, convErr := convertModelBindingType(obj, model.BgpRoutingConfigBindingType(), gm_model.BgpRoutingConfigBindingType())
		if convErr != nil {
			return convErr
		}
		gmRoutingConfig := gmObj.(gm_model.BgpRoutingConfig)

		client := gm_locale_services.NewBgpClient(connector)
		err = client.Patch(gwID, serviceID, gmRoutingConfig, nil)
	} else {
		localeService, err1 := getPolicyTier0GatewayLocaleServiceWithEdgeCluster(context, gwID, connector)
		if err1 != nil {
			return fmt.Errorf("Tier0 Gateway path with configured edge cluster expected, got %s", gwPath)
		}
		localeServiceID = *localeService.Id

		client := locale_services.NewBgpClient(connector)
		err = client.Patch(gwID, localeServiceID, *obj, nil)
	}
	if err != nil {
		return handleCreateError("BgpRoutingConfig", gwID, err)
	}

	d.SetId(newUUID())
	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)

	return resourceNsxtPolicyBgpConfigRead(d, m)
}

func resourceNsxtPolicyBgpConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	revision := int64(d.Get("revision").(int))
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)
	serviceID := d.Get("locale_service_id").(string)

	isVrf, err := resourceNsxtPolicyTier0GatewayIsVrf(gwID, connector, isPolicyGlobalManager(m))
	if err != nil {
		return handleCreateError("BgpRoutingConfig", gwID, err)
	}

	obj, err := resourceNsxtPolicyBgpConfigToStruct(d, isVrf)
	if err != nil {
		return handleUpdateError("BgpRoutingConfig", gwID, err)
	}

	obj.Revision = &revision
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.BgpRoutingConfigBindingType(), gm_model.BgpRoutingConfigBindingType())
		if convErr != nil {
			return convErr
		}
		gmRoutingConfig := gmObj.(gm_model.BgpRoutingConfig)

		client := gm_locale_services.NewBgpClient(connector)
		_, err = client.Update(gwID, serviceID, gmRoutingConfig, nil)
	} else {
		client := locale_services.NewBgpClient(connector)
		_, err = client.Update(gwID, serviceID, *obj, nil)
	}

	if err != nil {
		return handleUpdateError("BgpRoutingConfig", gwID, err)
	}

	return resourceNsxtPolicyBgpConfigRead(d, m)
}

func resourceNsxtPolicyBgpConfigDelete(d *schema.ResourceData, m interface{}) error {
	// BGP object can not be deleted as long as locale service exist
	// Delete call on NSX should revert settings to default, but this is
	// not supported on platform as of today
	// In order to revert manually, NSX requires us to disable BGP first
	// It makes more sense to avoid any action here (which mimics the computed
	// bgp behavior on T0 for local manager)

	return nil
}
