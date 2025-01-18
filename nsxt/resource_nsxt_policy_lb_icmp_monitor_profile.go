// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyLBIcmpMonitorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBIcmpMonitorProfileCreate,
		Read:   resourceNsxtPolicyLBIcmpMonitorProfileRead,
		Update: resourceNsxtPolicyLBIcmpMonitorProfileUpdate,
		Delete: resourceNsxtPolicyLBIcmpMonitorProfileDelete,
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
			"data_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The data size (in byte) of the ICMP healthcheck packet",
			},
			"fall_count": getLbMonitorFallCountSchema(),
			"interval":   getLbMonitorIntervalSchema(),
			"rise_count": getLbMonitorRiseCountSchema(),
			"timeout":    getLbMonitorTimeoutSchema(),
		},
	}
}

func resourceNsxtPolicyLBIcmpMonitorProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	dataLength := int64(d.Get("data_length").(int))
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	resourceType := model.LBMonitorProfile_RESOURCE_TYPE_LBICMPMONITORPROFILE

	obj := model.LBIcmpMonitorProfile{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		DataLength:   &dataLength,
		FallCount:    &fallCount,
		Interval:     &interval,
		RiseCount:    &riseCount,
		Timeout:      &timeout,
		ResourceType: resourceType,
	}

	log.Printf("[INFO] Patching LBIcmpMonitorProfile with ID %s", id)
	dataValue, errs := converter.ConvertToVapi(obj, model.LBIcmpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("LBMonitorProfile %s is not of type LBIcmpMonitorProfile %s", id, errs[0])
	}
	client := infra.NewLbMonitorProfilesClient(connector)
	return client.Patch(id, dataValue.(*data.StructValue))
}

func resourceNsxtPolicyLBIcmpMonitorProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBMonitorProfileExistsWrapper)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBIcmpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBIcmpMonitorProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBIcmpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBIcmpMonitorProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBIcmpMonitorProfile ID")
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBHttpsMonitorProfile", id, err)
	}

	baseObj, errs := converter.ConvertToGolang(obj, model.LBIcmpMonitorProfileBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting LBIcmpMonitorProfile %s", errs[0])
	}
	lbICMPMonitor := baseObj.(model.LBIcmpMonitorProfile)

	d.Set("display_name", lbICMPMonitor.DisplayName)
	d.Set("description", lbICMPMonitor.Description)
	setPolicyTagsInSchema(d, lbICMPMonitor.Tags)
	d.Set("nsx_id", id)
	d.Set("path", lbICMPMonitor.Path)
	d.Set("revision", lbICMPMonitor.Revision)

	d.Set("data_length", lbICMPMonitor.DataLength)
	d.Set("fall_count", lbICMPMonitor.FallCount)
	d.Set("interval", lbICMPMonitor.Interval)
	d.Set("rise_count", lbICMPMonitor.RiseCount)
	d.Set("timeout", lbICMPMonitor.Timeout)

	return nil
}

func resourceNsxtPolicyLBIcmpMonitorProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBIcmpMonitorProfile ID")
	}

	err := resourceNsxtPolicyLBIcmpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBIcmpMonitorProfile", id, err)
	}

	return resourceNsxtPolicyLBIcmpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBIcmpMonitorProfileDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyLBMonitorProfileDelete(d, m)
}
