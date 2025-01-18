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

func resourceNsxtPolicyLBHttpMonitorProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBHttpMonitorProfileCreate,
		Read:   resourceNsxtPolicyLBHttpMonitorProfileRead,
		Update: resourceNsxtPolicyLBHttpMonitorProfileUpdate,
		Delete: resourceNsxtPolicyLBHttpMonitorProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":                getNsxIDSchema(),
			"path":                  getPathSchema(),
			"display_name":          getDisplayNameSchema(),
			"description":           getDescriptionSchema(),
			"revision":              getRevisionSchema(),
			"tag":                   getTagsSchema(),
			"request_body":          getLbMonitorRequestBodySchema(),
			"request_header":        getLbHTTPHeaderSchema("Array of HTTP request headers"),
			"request_method":        getLbMonitorRequestMethodSchema(),
			"request_url":           getLbMonitorRequestURLSchema(),
			"request_version":       getLbMonitorRequestVersionSchema(),
			"response_body":         getLbMonitorResponseBodySchema(),
			"response_status_codes": getLbMonitorResponseStatusCodesSchema(),
			"fall_count":            getLbMonitorFallCountSchema(),
			"interval":              getLbMonitorIntervalSchema(),
			"rise_count":            getLbMonitorRiseCountSchema(),
			"timeout":               getLbMonitorTimeoutSchema(),
			"monitor_port":          getPolicyLbMonitorPortSchema(),
		},
	}
}

func resourceNsxtPolicyLBHttpMonitorProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	requestBody := d.Get("request_body").(string)
	requestHeaders := getPolicyLbHTTPHeaderFromSchema(d, "request_header")
	requestMethod := d.Get("request_method").(string)
	requestURL := d.Get("request_url").(string)
	requestVersion := d.Get("request_version").(string)
	responseBody := d.Get("response_body").(string)
	responseStatusCodes := interface2Int64List(d.Get("response_status_codes").([]interface{}))
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	monitorPort := int64(d.Get("monitor_port").(int))

	resourceType := model.LBMonitorProfile_RESOURCE_TYPE_LBHTTPMONITORPROFILE

	obj := model.LBHttpMonitorProfile{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		RequestBody:         &requestBody,
		RequestHeaders:      requestHeaders,
		RequestMethod:       &requestMethod,
		RequestUrl:          &requestURL,
		RequestVersion:      &requestVersion,
		ResponseBody:        &responseBody,
		ResponseStatusCodes: responseStatusCodes,
		FallCount:           &fallCount,
		Interval:            &interval,
		MonitorPort:         &monitorPort,
		RiseCount:           &riseCount,
		Timeout:             &timeout,
		ResourceType:        resourceType,
	}

	log.Printf("[INFO] Patching LBHttpMonitorProfile with ID %s", id)
	dataValue, errs := converter.ConvertToVapi(obj, model.LBHttpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("LBMonitorProfile %s is not of type LBHttpMonitorProfile %s", id, errs[0])
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	return client.Patch(id, dataValue.(*data.StructValue))
}

func resourceNsxtPolicyLBHttpMonitorProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBMonitorProfileExistsWrapper)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBHttpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBHttpMonitorProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBHttpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBHttpMonitorProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBHttpMonitorProfile ID")
	}

	client := infra.NewLbMonitorProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBHttpMonitorProfile", id, err)
	}

	baseObj, errs := converter.ConvertToGolang(obj, model.LBHttpMonitorProfileBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting LBHttpMonitorProfile %s", errs[0])
	}

	lbHTTPMonitor := baseObj.(model.LBHttpMonitorProfile)

	d.Set("revision", lbHTTPMonitor.Revision)
	d.Set("description", lbHTTPMonitor.Description)
	d.Set("display_name", lbHTTPMonitor.DisplayName)
	setPolicyTagsInSchema(d, lbHTTPMonitor.Tags)
	d.Set("fall_count", lbHTTPMonitor.FallCount)
	d.Set("interval", lbHTTPMonitor.Interval)
	d.Set("monitor_port", lbHTTPMonitor.MonitorPort)
	d.Set("rise_count", lbHTTPMonitor.RiseCount)
	d.Set("timeout", lbHTTPMonitor.Timeout)
	d.Set("request_body", lbHTTPMonitor.RequestBody)
	d.Set("path", lbHTTPMonitor.Path)
	d.Set("nsx_id", id)
	setPolicyLbHTTPHeaderInSchema(d, "request_header", lbHTTPMonitor.RequestHeaders)
	d.Set("request_method", lbHTTPMonitor.RequestMethod)
	d.Set("request_url", lbHTTPMonitor.RequestUrl)
	d.Set("request_version", lbHTTPMonitor.RequestVersion)
	d.Set("response_body", lbHTTPMonitor.ResponseBody)
	d.Set("response_status_codes", int64List2Interface(lbHTTPMonitor.ResponseStatusCodes))

	return nil
}

func resourceNsxtPolicyLBHttpMonitorProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBHttpMonitorProfile ID")
	}

	err := resourceNsxtPolicyLBHttpMonitorProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBHttpMonitorProfile", id, err)
	}

	return resourceNsxtPolicyLBHttpMonitorProfileRead(d, m)
}

func resourceNsxtPolicyLBHttpMonitorProfileDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyLBMonitorProfileDelete(d, m)
}
