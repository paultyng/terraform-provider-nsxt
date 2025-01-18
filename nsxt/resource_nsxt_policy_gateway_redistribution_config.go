// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	tier0s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_0s"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		BgpEnabled: &bgpEnabled,
	}

	if util.NsxVersionHigherOrEqual("3.1.0") {
		redistributionStruct.OspfEnabled = &ospfEnabled
	}

	setLocaleServiceRedistributionRulesConfig(rulesConfig, &redistributionStruct)

	lsType := "LocaleServices"
	serviceStruct := model.LocaleServices{
		Id:                        &localeServiceID,
		ResourceType:              &lsType,
		RouteRedistributionConfig: &redistributionStruct,
	}

	doPatch := func() error {
		client := tier0s.NewLocaleServicesClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		return client.Patch(gwID, localeServiceID, serviceStruct)
	}
	// since redistribution config is not a separate API endpoint, but sub-clause of Tier0,
	// concurrency issues may arise that require retry from client side.
	return doPatch()
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
	context := getSessionContext(d, m)
	if isPolicyGlobalManager(m) {
		if sitePath == "" {
			return attributeRequiredGlobalManagerError("site_path", "nsxt_policy_gateway_redistribution_config")
		}
		localeServices, err := listPolicyTier0GatewayLocaleServices(context, connector, gwID)
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
		localeService, err := getPolicyTier0GatewayLocaleServiceWithEdgeCluster(context, gwID, connector)
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

	client := tier0s.NewLocaleServicesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	obj, err := client.Get(gwID, localeServiceID)
	if err != nil {
		return handleReadError(d, "Tier0 Redistribution Config", id, err)
	}

	config := obj.RouteRedistributionConfig
	if config != nil {
		d.Set("bgp_enabled", config.BgpEnabled)
		d.Set("ospf_enabled", config.OspfEnabled)
		d.Set("rule", getLocaleServiceRedistributionRuleConfig(config))
	}
	if isPolicyGlobalManager(m) && obj.EdgeClusterPath != nil {
		d.Set("site_path", getSitePathFromEdgePath(*obj.EdgeClusterPath))
	} else {
		d.Set("site_path", "")
	}
	d.Set("gateway_path", getGatewayPathFromLocaleServicesPath(*obj.Path))

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

	// Update the locale service with empty Redistribution config using get/post
	doUpdate := func() error {
		client := tier0s.NewLocaleServicesClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		obj, err := client.Get(gwID, localeServiceID)
		if err != nil {
			return err
		}
		obj.RouteRedistributionConfig = nil
		_, err = client.Update(gwID, localeServiceID, obj)
		return err

	}

	commonProviderConfig := getCommonProviderConfig(m)
	err := retryUponPreconditionFailed(doUpdate, commonProviderConfig.MaxRetries)
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
	client := tier0s.NewLocaleServicesClient(getSessionContext(d, m), connector)
	if client == nil {
		return nil, policyResourceNotSupportedError()
	}
	obj, err := client.Get(gwID, localeServiceID)
	if err != nil || obj.RouteRedistributionConfig == nil {
		return nil, fmt.Errorf("Failed to retrieve redistribution config for locale service %s on gateway %s", localeServiceID, gwID)
	}

	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)

	d.SetId(newUUID())

	return []*schema.ResourceData{d}, nil
}
