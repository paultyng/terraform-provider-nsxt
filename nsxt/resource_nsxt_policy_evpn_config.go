// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var policyEvpnConfigTypeValues = []string{
	model.EvpnConfig_MODE_INLINE,
	model.EvpnConfig_MODE_ROUTE_SERVER,
}

func resourceNsxtPolicyEvpnConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyEvpnConfigCreate,
		Read:   resourceNsxtPolicyEvpnConfigRead,
		Update: resourceNsxtPolicyEvpnConfigUpdate,
		Delete: resourceNsxtPolicyEvpnConfigDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyEvpnConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"path":          getPathSchema(),
			"display_name":  getDisplayNameSchema(),
			"description":   getDescriptionSchema(),
			"revision":      getRevisionSchema(),
			"tag":           getTagsSchema(),
			"gateway_path":  getPolicyPathSchema(true, true, "Policy path for the Gateway"),
			"vni_pool_path": getPolicyPathSchema(false, false, "Policy path for VNI Pool"),
			"evpn_tenant_path": {
				Type:          schema.TypeString,
				Description:   "Policy path for EVPN Tenant",
				Optional:      true,
				ValidateFunc:  validatePolicyPath(),
				ConflictsWith: []string{"vni_pool_path"},
			},
			"mode": {
				Type:         schema.TypeString,
				Description:  "EVPN Mode",
				Required:     true,
				ValidateFunc: validation.StringInSlice(policyEvpnConfigTypeValues, false),
			},
		},
	}
}

func policyEvpnConfigGet(connector client.Connector, gwID string) (model.EvpnConfig, error) {
	client := tier_0s.NewEvpnClient(connector)
	return client.Get(gwID)
}

func resourceNsxtPolicyEvpnConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("Tier0 gateway path expected, got %s", gwPolicyPath)
	}

	obj, err := policyEvpnConfigGet(connector, gwID)

	if err != nil {
		return handleReadError(d, "Evpn Config", gwID, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("mode", obj.Mode)
	if obj.EncapsulationMethod != nil {
		d.Set("vni_pool_path", obj.EncapsulationMethod.VniPoolPath)
		d.Set("evpn_tenant_path", obj.EncapsulationMethod.EvpnTenantConfigPath)
	}

	return nil
}

func patchNsxtPolicyEvpnConfig(connector client.Connector, d *schema.ResourceData, gwID string, isGlobalManager bool) error {

	var obj model.EvpnConfig
	if d != nil {
		displayName := d.Get("display_name").(string)
		description := d.Get("description").(string)
		tags := getPolicyTagsFromSchema(d)
		vniPoolPath := d.Get("vni_pool_path").(string)
		evpnTenantPath := d.Get("evpn_tenant_path").(string)
		mode := d.Get("mode").(string)
		encapConfig := model.EvpnEncapConfig{}

		if len(vniPoolPath) > 0 {
			encapConfig.VniPoolPath = &vniPoolPath
		}

		if len(evpnTenantPath) > 0 {
			encapConfig.EvpnTenantConfigPath = &evpnTenantPath
		}

		obj.DisplayName = &displayName
		obj.Description = &description
		obj.Mode = &mode
		obj.Tags = tags
		obj.EncapsulationMethod = &encapConfig
	} else {
		// DELETE use case (actual DELETE is not supported)
		mode := model.EvpnConfig_MODE_DISABLE
		obj.Mode = &mode
	}
	client := tier_0s.NewEvpnClient(connector)
	return client.Patch(gwID, obj)
}

func resourceNsxtPolicyEvpnConfigCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("Tier0 gateway path expected, got %s", gwPolicyPath)
	}

	isGlobalManager := isPolicyGlobalManager(m)

	log.Printf("[INFO] Creating EVPN Config for Gateway %s", gwID)

	err := patchNsxtPolicyEvpnConfig(connector, d, gwID, isGlobalManager)
	if err != nil {
		return handleCreateError("Evpn Config", gwID, err)
	}

	d.SetId(gwID)

	return resourceNsxtPolicyEvpnConfigRead(d, m)
}

func resourceNsxtPolicyEvpnConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	log.Printf("[INFO] Updating Evpn Config with ID %s", gwID)
	err := patchNsxtPolicyEvpnConfig(connector, d, gwID, isPolicyGlobalManager(m))
	if err != nil {
		return handleUpdateError("Evpn Config", gwID, err)
	}

	return resourceNsxtPolicyEvpnConfigRead(d, m)
}

func resourceNsxtPolicyEvpnConfigDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	// There is no DELETE API for this object - we need to just disable it
	err := patchNsxtPolicyEvpnConfig(connector, nil, gwID, isPolicyGlobalManager(m))
	if err != nil {
		return handleDeleteError("Evpn Config", gwID, err)
	}

	return nil
}

func resourceNsxtPolicyEvpnConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	gwPath := d.Id()

	d.Set("gateway_path", gwPath)
	_, gwID := parseGatewayPolicyPath(gwPath)
	d.SetId(gwID)

	return []*schema.ResourceData{d}, nil
}
