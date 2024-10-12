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

func resourceNsxtPolicyLBTcpMonitorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBTcpMonitorProfileCreate,
		Read:   resourceNsxtPolicyLBTcpMonitorProfileRead,
		Update: resourceNsxtPolicyLBTcpMonitorProfileUpdate,
		Delete: resourceNsxtPolicyLBTcpMonitorProfileDelete,
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

func resourceNsxtPolicyLBTcpMonitorProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
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
	resourceType := model.LBMonitorProfile_RESOURCE_TYPE_LBTCPMONITORPROFILE

	obj := model.LBTcpMonitorProfile{
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

	log.Printf("[INFO] Patching LBTcpMonitorProfile with ID %s", id)

	dataValue, errs := converter.ConvertToVapi(obj, model.LBTcpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("LBMonitorProfile %s is not of type LBTcpMonitorProfile %s", id, errs[0])
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	return client.Patch(id, dataValue.(*data.StructValue))
}

func resourceNsxtPolicyLBTcpMonitorProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBMonitorProfileExistsWrapper)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBTcpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBTcpMonitorProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBTcpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBTcpMonitorProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBTcpMonitorProfile ID")
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBHTcpMonitorProfile", id, err)
	}

	baseObj, errs := converter.ConvertToGolang(obj, model.LBTcpMonitorProfileBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting LBTcpMonitorProfile %s", errs[0])
	}
	lbTCPMonitor := baseObj.(model.LBTcpMonitorProfile)

	d.Set("display_name", lbTCPMonitor.DisplayName)
	d.Set("description", lbTCPMonitor.Description)
	setPolicyTagsInSchema(d, lbTCPMonitor.Tags)
	d.Set("nsx_id", id)
	d.Set("path", lbTCPMonitor.Path)
	d.Set("revision", lbTCPMonitor.Revision)

	d.Set("receive", lbTCPMonitor.Receive)
	d.Set("send", lbTCPMonitor.Send)
	d.Set("fall_count", lbTCPMonitor.FallCount)
	d.Set("interval", lbTCPMonitor.Interval)
	d.Set("monitor_port", lbTCPMonitor.MonitorPort)
	d.Set("rise_count", lbTCPMonitor.RiseCount)
	d.Set("timeout", lbTCPMonitor.Timeout)

	return nil
}

func resourceNsxtPolicyLBTcpMonitorProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBTcpMonitorProfile ID")
	}

	err := resourceNsxtPolicyLBTcpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBTcpMonitorProfile", id, err)
	}

	return resourceNsxtPolicyLBTcpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBTcpMonitorProfileDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyLBMonitorProfileDelete(d, m)
}
