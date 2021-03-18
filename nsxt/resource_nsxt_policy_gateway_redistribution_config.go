/* Copyright © 2021 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyGatewayRedistributionConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayRedistributionConfigCreate,
		Read:   resourceNsxtPolicyGatewayRedistributionConfigRead,
		Update: resourceNsxtPolicyGatewayRedistributionConfigUpdate,
		Delete: resourceNsxtPolicyGatewayRedistributionConfigDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyGatewayRedistributionConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path of the site the Tier0 redistribution",
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"gateway_path": getPolicyPathSchema(true, true, "Policy path for Tier0 gateway"),
			"bgp_enabled": {
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
			"locale_service_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Gateway Locale Service on NSX",
				Computed:    true,
			},
			"gateway_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Tier0 Gateway on NSX",
				Computed:    true,
			},
		},
	}
}

func policyGatewayRedistributionConfigPatch(d *schema.ResourceData, m interface{}, gwID string, localeServiceID string) error {

	connector := getPolicyConnector(m)

	bgpEnabled := d.Get("bgp_enabled").(bool)
	ospfEnabled := d.Get("ospf_enabled").(bool)
	rulesConfig := d.Get("rule").([]interface{})

	redistributionStruct := model.Tier0RouteRedistributionConfig{
		BgpEnabled:  &bgpEnabled,
		OspfEnabled: &ospfEnabled,
	}

	setLocaleServiceRedistributionRulesConfig(rulesConfig, &redistributionStruct)

	lsType := "LocaleServices"
	serviceStruct := model.LocaleServices{
		Id:                        &localeServiceID,
		ResourceType:              &lsType,
		RouteRedistributionConfig: &redistributionStruct,
	}

	if isPolicyGlobalManager(m) {
		// Use patch to only update the relevant fields
		rawObj, err := convertModelBindingType(serviceStruct, model.LocaleServicesBindingType(), gm_model.LocaleServicesBindingType())
		if err != nil {
			return err
		}
		client := gm_tier0s.NewDefaultLocaleServicesClient(connector)
		return client.Patch(gwID, localeServiceID, rawObj.(gm_model.LocaleServices))

	}
	client := tier_0s.NewDefaultLocaleServicesClient(connector)
	return client.Patch(gwID, localeServiceID, serviceStruct)
}

func resourceNsxtPolicyGatewayRedistributionConfigCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	sitePath := d.Get("site_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	localeServiceID := ""
	if isPolicyGlobalManager(m) {
		if sitePath == "" {
			return attributeRequiredGlobalManagerError("site_path", "nsxt_policy_gateway_redistribution_config")
		}
		localeServices, err := listPolicyTier0GatewayLocaleServices(connector, gwID, true)
		if err != nil {
			return err
		}
		localeServiceID, err = getGlobalPolicyGatewayLocaleServiceIDWithSite(localeServices, sitePath, gwID)
		if err != nil {
			return err
		}
	} else {
		if sitePath != "" {
			return globalManagerOnlyError()
		}
		localeService, err := getPolicyTier0GatewayLocaleServiceWithEdgeCluster(gwID, connector)
		if err != nil {
			return err
		}
		if localeService == nil {
			return fmt.Errorf("Edge cluster is mandatory on gateway %s in order to create interfaces", gwID)
		}
		localeServiceID = *localeService.Id
	}

	id := newUUID()
	err := policyGatewayRedistributionConfigPatch(d, m, gwID, localeServiceID)
	if err != nil {
		return handleCreateError("Tier0 Redistribution Config", id, err)
	}

	d.SetId(id)
	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)

	return resourceNsxtPolicyGatewayRedistributionConfigRead(d, m)
}

func resourceNsxtPolicyGatewayRedistributionConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || gwID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 Gateway id or Locale Service id")
	}

	var obj model.LocaleServices
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewDefaultLocaleServicesClient(connector)
		gmObj, err1 := client.Get(gwID, localeServiceID)
		if err1 != nil {
			return handleReadError(d, "Tier0 Redistribution Config", id, err1)
		}
		lmObj, err2 := convertModelBindingType(gmObj, model.LocaleServicesBindingType(), model.LocaleServicesBindingType())
		if err2 != nil {
			return err2
		}
		obj = lmObj.(model.LocaleServices)
	} else {
		var err error
		client := tier_0s.NewDefaultLocaleServicesClient(connector)
		obj, err = client.Get(gwID, defaultPolicyLocaleServiceID)
		if err != nil {
			return handleReadError(d, "Tier0 Redistribution Config", id, err)
		}
	}

	config := obj.RouteRedistributionConfig
	d.Set("bgp_enabled", config.BgpEnabled)
	d.Set("ospf_enabled", config.OspfEnabled)
	d.Set("rule", getLocaleServiceRedistributionRuleConfig(config))

	return nil
}

func resourceNsxtPolicyGatewayRedistributionConfigUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || gwID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 Gateway id or Locale Service id")
	}

	err := policyGatewayRedistributionConfigPatch(d, m, gwID, localeServiceID)
	if err != nil {
		return handleUpdateError("Tier0 Redistribution Config", id, err)
	}

	return resourceNsxtPolicyGatewayRedistributionConfigRead(d, m)
}

func resourceNsxtPolicyGatewayRedistributionConfigDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || gwID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 Gateway id or Locale Service id")
	}

	// Update the locale service with empty HaVipConfigs using get/post
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewDefaultLocaleServicesClient(connector)
		gmObj, err1 := client.Get(gwID, localeServiceID)
		if err1 != nil {
			return handleDeleteError("Tier0 Redistribution config", id, err)
		}
		gmObj.RouteRedistributionConfig = nil
		_, err = client.Update(gwID, localeServiceID, gmObj)
	} else {
		client := tier_0s.NewDefaultLocaleServicesClient(connector)
		obj, err1 := client.Get(gwID, localeServiceID)
		if err1 != nil {
			return handleDeleteError("Tier0 Redistribution config", id, err)
		}
		obj.RouteRedistributionConfig = nil
		_, err = client.Update(gwID, localeServiceID, obj)
	}
	if err != nil {
		return handleDeleteError("Tier0 RedistributionConfig config", id, err)
	}

	return nil
}

func resourceNsxtPolicyGatewayRedistributionConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <tier0-gateway-id>/<locale-service-id> as an input")
	}

	gwID := s[0]
	localeServiceID := s[1]
	connector := getPolicyConnector(m)
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewDefaultLocaleServicesClient(connector)
		obj, err := client.Get(gwID, localeServiceID)
		if err != nil || obj.RouteRedistributionConfig == nil {
			return nil, fmt.Errorf("Failed to retrieve redistribution config for locale service %s on gateway %s", localeServiceID, gwID)
		}
	} else {
		client := tier_0s.NewDefaultLocaleServicesClient(connector)
		obj, err := client.Get(gwID, localeServiceID)
		if err != nil || obj.RouteRedistributionConfig == nil {
			return nil, fmt.Errorf("Failed to retrieve redistribution config for locale service %s on gateway %s", localeServiceID, gwID)
		}
	}

	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)

	d.SetId(newUUID())

	return []*schema.ResourceData{d}, nil
}
