// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyLBPassiveMonitorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBPassiveMonitorProfileCreate,
		Read:   resourceNsxtPolicyLBPassiveMonitorProfileRead,
		Update: resourceNsxtPolicyLBPassiveMonitorProfileUpdate,
		Delete: resourceNsxtPolicyLBPassiveMonitorProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"max_fails": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Number of consecutive failures before a member is considered temporarily unavailable",
				Default:      5,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"timeout": getLbMonitorTimeoutSchema(),
		},
	}
}

func resourceNsxtPolicyLBPassiveMonitorProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	maxFails := int64(d.Get("max_fails").(int))
	timeout := int64(d.Get("timeout").(int))
	resourceType := model.LBMonitorProfile_RESOURCE_TYPE_LBPASSIVEMONITORPROFILE

	obj := model.LBPassiveMonitorProfile{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		MaxFails:     &maxFails,
		Timeout:      &timeout,
		ResourceType: resourceType,
	}

	log.Printf("[INFO] Patching LBPassiveMonitorProfile with ID %s", id)

	dataValue, errs := converter.ConvertToVapi(obj, model.LBPassiveMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("LBMonitorProfile %s is not of type LBPassiveMonitorProfile %s", id, errs[0])
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	return client.Patch(id, dataValue.(*data.StructValue))
}

func resourceNsxtPolicyLBPassiveMonitorProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBMonitorProfileExistsWrapper)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBPassiveMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBPassiveMonitorProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBPassiveMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBPassiveMonitorProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBPassiveMonitorProfile ID")
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBPassiveMonitorProfile", id, err)
	}

	baseObj, errs := converter.ConvertToGolang(obj, model.LBPassiveMonitorProfileBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting LBPassiveMonitorProfile %s", errs[0])
	}
	lbPassiveMonitor := baseObj.(model.LBPassiveMonitorProfile)

	d.Set("display_name", lbPassiveMonitor.DisplayName)
	d.Set("description", lbPassiveMonitor.Description)
	setPolicyTagsInSchema(d, lbPassiveMonitor.Tags)
	d.Set("nsx_id", id)
	d.Set("path", lbPassiveMonitor.Path)
	d.Set("revision", lbPassiveMonitor.Revision)
	d.Set("max_fails", lbPassiveMonitor.MaxFails)
	d.Set("timeout", lbPassiveMonitor.Timeout)

	return nil
}

func resourceNsxtPolicyLBPassiveMonitorProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBPassiveMonitorProfile ID")
	}

	err := resourceNsxtPolicyLBPassiveMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBPassiveMonitorProfile", id, err)
	}

	return resourceNsxtPolicyLBPassiveMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBPassiveMonitorProfileDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyLBMonitorProfileDelete(d, m)
}
