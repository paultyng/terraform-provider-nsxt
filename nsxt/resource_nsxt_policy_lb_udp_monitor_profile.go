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

func resourceNsxtPolicyLBUdpMonitorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBUdpMonitorProfileCreate,
		Read:   resourceNsxtPolicyLBUdpMonitorProfileRead,
		Update: resourceNsxtPolicyLBUdpMonitorProfileUpdate,
		Delete: resourceNsxtPolicyLBUdpMonitorProfileDelete,
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
			"receive": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The expected data string to be received from the response, can be anywhere in the response",
			},
			"send": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The data to be sent to the monitored server.",
			},
			"fall_count":   getLbMonitorFallCountSchema(),
			"interval":     getLbMonitorIntervalSchema(),
			"monitor_port": getPolicyLbMonitorPortSchema(),
			"rise_count":   getLbMonitorRiseCountSchema(),
			"timeout":      getLbMonitorTimeoutSchema(),
		},
	}
}

func resourceNsxtPolicyLBUdpMonitorProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	receive := d.Get("receive").(string)
	send := d.Get("send").(string)
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	monitorPort := int64(d.Get("monitor_port").(int))
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	resourceType := model.LBMonitorProfile_RESOURCE_TYPE_LBUDPMONITORPROFILE

	obj := model.LBUdpMonitorProfile{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		Receive:      &receive,
		Send:         &send,
		FallCount:    &fallCount,
		Interval:     &interval,
		MonitorPort:  &monitorPort,
		RiseCount:    &riseCount,
		Timeout:      &timeout,
		ResourceType: resourceType,
	}

	log.Printf("[INFO] Patching LBUdpMonitorProfile with ID %s", id)

	dataValue, errs := converter.ConvertToVapi(obj, model.LBUdpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("Error converting LBUdpMonitorProfile %s", errs[0])
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	return client.Patch(id, dataValue.(*data.StructValue))
}

func resourceNsxtPolicyLBUdpMonitorProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBMonitorProfileExistsWrapper)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBUdpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBUdpMonitorProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBUdpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBUdpMonitorProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBUdpMonitorProfile ID")
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBUdpMonitorProfile", id, err)
	}

	baseObj, errs := converter.ConvertToGolang(obj, model.LBUdpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("LBMonitorProfile %s is not of type LBUdpMonitorProfile %s", id, errs[0])
	}
	lbUDPMonitor := baseObj.(model.LBUdpMonitorProfile)

	d.Set("display_name", lbUDPMonitor.DisplayName)
	d.Set("description", lbUDPMonitor.Description)
	setPolicyTagsInSchema(d, lbUDPMonitor.Tags)
	d.Set("nsx_id", id)
	d.Set("path", lbUDPMonitor.Path)
	d.Set("revision", lbUDPMonitor.Revision)

	d.Set("receive", lbUDPMonitor.Receive)
	d.Set("send", lbUDPMonitor.Send)
	d.Set("fall_count", lbUDPMonitor.FallCount)
	d.Set("interval", lbUDPMonitor.Interval)
	d.Set("monitor_port", lbUDPMonitor.MonitorPort)
	d.Set("rise_count", lbUDPMonitor.RiseCount)
	d.Set("timeout", lbUDPMonitor.Timeout)

	return nil
}

func resourceNsxtPolicyLBUdpMonitorProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBUdpMonitorProfile ID")
	}

	err := resourceNsxtPolicyLBUdpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBUdpMonitorProfile", id, err)
	}

	return resourceNsxtPolicyLBUdpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBUdpMonitorProfileDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyLBMonitorProfileDelete(d, m)
}
