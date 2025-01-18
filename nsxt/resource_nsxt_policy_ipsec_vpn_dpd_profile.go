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
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var iPSecVpnDpdProfileDpdProbeModeValues = []string{
	model.IPSecVpnDpdProfile_DPD_PROBE_MODE_ON_DEMAND,
	model.IPSecVpnDpdProfile_DPD_PROBE_MODE_PERIODIC,
}

func resourceNsxtPolicyIPSecVpnDpdProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPSecVpnDpdProfileCreate,
		Read:   resourceNsxtPolicyIPSecVpnDpdProfileRead,
		Update: resourceNsxtPolicyIPSecVpnDpdProfileUpdate,
		Delete: resourceNsxtPolicyIPSecVpnDpdProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"dpd_probe_interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  60,
			},
			"dpd_probe_mode": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(iPSecVpnDpdProfileDpdProbeModeValues, false),
				Optional:     true,
				Default:      model.IPSecVpnDpdProfile_DPD_PROBE_MODE_PERIODIC,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"retry_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
		},
	}
}

func resourceNsxtPolicyIPSecVpnDpdProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewIpsecVpnDpdProfilesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPSecVpnDpdProfileCreate(d *schema.ResourceData, m interface{}) error {

	if isPolicyGlobalManager(m) {
		return resourceNotSupportedError()
	}

	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyIPSecVpnDpdProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	dpdProbeInterval := int64(d.Get("dpd_probe_interval").(int))
	dpdProbeMode := d.Get("dpd_probe_mode").(string)
	enabled := d.Get("enabled").(bool)
	retryCount := int64(d.Get("retry_count").(int))

	obj := model.IPSecVpnDpdProfile{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		DpdProbeInterval: &dpdProbeInterval,
		DpdProbeMode:     &dpdProbeMode,
		Enabled:          &enabled,
		RetryCount:       &retryCount,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating IPSecVpnDpdProfile with ID %s", id)
	client := infra.NewIpsecVpnDpdProfilesClient(connector)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("IPSecVpnDpdProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPSecVpnDpdProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnDpdProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnDpdProfile ID")
	}

	client := infra.NewIpsecVpnDpdProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "IPSecVpnDpdProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("dpd_probe_interval", obj.DpdProbeInterval)
	d.Set("dpd_probe_mode", obj.DpdProbeMode)
	d.Set("enabled", obj.Enabled)
	d.Set("retry_count", obj.RetryCount)

	return nil
}

func resourceNsxtPolicyIPSecVpnDpdProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnDpdProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	dpdProbeInterval := int64(d.Get("dpd_probe_interval").(int))
	dpdProbeMode := d.Get("dpd_probe_mode").(string)
	enabled := d.Get("enabled").(bool)
	retryCount := int64(d.Get("retry_count").(int))

	obj := model.IPSecVpnDpdProfile{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		DpdProbeInterval: &dpdProbeInterval,
		DpdProbeMode:     &dpdProbeMode,
		Enabled:          &enabled,
		RetryCount:       &retryCount,
	}

	client := infra.NewIpsecVpnDpdProfilesClient(connector)
	err := client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("IPSecVpnDpdProfile", id, err)
	}

	return resourceNsxtPolicyIPSecVpnDpdProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnDpdProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnDpdProfile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewIpsecVpnDpdProfilesClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("IPSecVpnDpdProfile", id, err)
	}

	return nil
}
