// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyOspfConfig() *schema.Resource {

	return &schema.Resource{
		Create: resourceNsxtPolicyOspfConfigCreate,
		Read:   resourceNsxtPolicyOspfConfigRead,
		Update: resourceNsxtPolicyOspfConfigUpdate,
		Delete: resourceNsxtPolicyOspfConfigDelete,

		Schema: getPolicyOspfConfigSchema(),
	}
}

var nsxtPolicyTier0GatewayOspfGracefulRestartModes = []string{
	model.OspfRoutingConfig_GRACEFUL_RESTART_MODE_DISABLE,
	model.OspfRoutingConfig_GRACEFUL_RESTART_MODE_HELPER_ONLY,
}

func getPolicyOspfConfigSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"display_name": getDisplayNameSchema(),
		"description":  getDescriptionSchema(),
		"tag":          getTagsSchema(),
		"revision":     getRevisionSchema(),
		"path":         getPathSchema(),
		"gateway_path": getPolicyPathSchema(true, true, "Policy path for the Tier0 Gateway"),
		"ecmp": {
			Type:        schema.TypeBool,
			Description: "Flag to enable ECMP",
			Optional:    true,
			Default:     true,
		},
		"default_originate": {
			Type:        schema.TypeBool,
			Description: "Flag to enable/disable advertisement of default route into OSPF domain",
			Optional:    true,
			Default:     false,
		},
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Flag to enable OSPF configuration",
			Optional:    true,
			Default:     true,
		},
		"graceful_restart_mode": {
			Type:         schema.TypeString,
			Description:  "Graceful Restart Mode",
			ValidateFunc: validation.StringInSlice(nsxtPolicyTier0GatewayOspfGracefulRestartModes, false),
			Optional:     true,
			Default:      model.OspfRoutingConfig_GRACEFUL_RESTART_MODE_HELPER_ONLY,
		},
		"summary_address": {
			Type:        schema.TypeList,
			Description: "List of addresses to summarize or filter external routes",
			Optional:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"prefix": {
						Type:         schema.TypeString,
						Description:  "OSPF Summary address in CIDR format",
						Optional:     true,
						ValidateFunc: validateCidr(),
					},
					"advertise": {
						Type:        schema.TypeBool,
						Description: "Used to filter the advertisement of external routes into the OSPF domain",
						Optional:    true,
						Default:     true,
					},
				},
			},
		},
		"locale_service_id": getComputedLocaleServiceIDSchema(),
		"gateway_id":        getComputedGatewayIDSchema(),
	}
}

func policyOspfConfigPatch(d *schema.ResourceData, m interface{}, gwID string, localeServiceID string) error {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	ecmp := d.Get("ecmp").(bool)
	enabled := d.Get("enabled").(bool)
	defaultOriginate := d.Get("default_originate").(bool)
	gracefulRestartMode := d.Get("graceful_restart_mode").(string)
	summaryAddresses := d.Get("summary_address").([]interface{})
	var addresses []model.OspfSummaryAddressConfig

	for _, address := range summaryAddresses {
		data := address.(map[string]interface{})
		prefix := data["prefix"].(string)
		advertise := data["advertise"].(bool)

		item := model.OspfSummaryAddressConfig{
			Prefix:    &prefix,
			Advertise: &advertise,
		}

		addresses = append(addresses, item)
	}

	obj := model.OspfRoutingConfig{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		Enabled:             &enabled,
		Ecmp:                &ecmp,
		DefaultOriginate:    &defaultOriginate,
		GracefulRestartMode: &gracefulRestartMode,
		SummaryAddresses:    addresses,
	}

	connector := getPolicyConnector(m)
	client := locale_services.NewOspfClient(connector)
	_, err := client.Patch(gwID, localeServiceID, obj)
	return err
}

func resourceNsxtPolicyOspfConfigCreate(d *schema.ResourceData, m interface{}) error {

	if isPolicyGlobalManager(m) {
		return localManagerOnlyError()
	}

	connector := getPolicyConnector(m)
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	localeService, err := getPolicyTier0GatewayLocaleServiceWithEdgeCluster(getSessionContext(d, m), gwID, connector)
	if err != nil {
		return fmt.Errorf("Tier0 Gateway path with configured edge cluster expected, got %s", gwPath)
	}

	err = policyOspfConfigPatch(d, m, gwID, *localeService.Id)
	if err != nil {
		return handleCreateError("Ospf Config", gwID, err)
	}

	d.SetId(newUUID())
	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeService.Id)

	return resourceNsxtPolicyOspfConfigRead(d, m)
}

func resourceNsxtPolicyOspfConfigRead(d *schema.ResourceData, m interface{}) error {

	if isPolicyGlobalManager(m) {
		return localManagerOnlyError()
	}

	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	localeServiceID := d.Get("locale_service_id").(string)

	client := locale_services.NewOspfClient(connector)
	obj, err := client.Get(gwID, localeServiceID)
	if err != nil {
		return handleReadError(d, "Ospf Config", gwID, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("enabled", obj.Enabled)
	d.Set("ecmp", obj.Ecmp)
	d.Set("default_originate", obj.DefaultOriginate)
	d.Set("graceful_restart_mode", obj.GracefulRestartMode)

	var summaryAddresses []map[string]interface{}
	for _, address := range obj.SummaryAddresses {
		data := make(map[string]interface{})
		data["prefix"] = address.Prefix
		data["advertise"] = address.Advertise
		summaryAddresses = append(summaryAddresses, data)
	}

	d.Set("summary_address", summaryAddresses)

	return nil
}

func resourceNsxtPolicyOspfConfigUpdate(d *schema.ResourceData, m interface{}) error {

	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)

	err := policyOspfConfigPatch(d, m, gwID, localeServiceID)
	if err != nil {
		return handleCreateError("Ospf Config", gwID, err)
	}

	return resourceNsxtPolicyOspfConfigRead(d, m)
}

func resourceNsxtPolicyOspfConfigDelete(d *schema.ResourceData, m interface{}) error {
	// OSPF object can not be deleted as long as locale service exist
	// Delete call on NSX should revert settings to default, but this is
	// not supported on platform as of today
	// In order to revert manually, NSX requires us to disable OSPF first
	// It makes more sense to avoid any action here (which mimics the computed
	// ospf behavior on T0 for local manager)

	return nil
}
