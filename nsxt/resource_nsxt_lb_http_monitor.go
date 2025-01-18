// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

func resourceNsxtLbHTTPMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbHTTPMonitorCreate,
		Read:   resourceNsxtLbHTTPMonitorRead,
		Update: resourceNsxtLbHTTPMonitorUpdate,
		Delete: resourceNsxtLbMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag":                   getTagsSchema(),
			"fall_count":            getLbMonitorFallCountSchema(),
			"interval":              getLbMonitorIntervalSchema(),
			"monitor_port":          getLbMonitorPortSchema(),
			"rise_count":            getLbMonitorRiseCountSchema(),
			"timeout":               getLbMonitorTimeoutSchema(),
			"request_body":          getLbMonitorRequestBodySchema(),
			"request_header":        getLbHTTPHeaderSchema("Array of HTTP request headers"),
			"request_method":        getLbMonitorRequestMethodSchema(),
			"request_url":           getLbMonitorRequestURLSchema(),
			"request_version":       getLbMonitorRequestVersionSchema(),
			"response_body":         getLbMonitorResponseBodySchema(),
			"response_status_codes": getLbMonitorResponseStatusCodesSchema(),
		},
	}
}

func resourceNsxtLbHTTPMonitorCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	monitorPort := d.Get("monitor_port").(string)
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	requestBody := d.Get("request_body").(string)
	requestHeaders := getLbHTTPHeaderFromSchema(d, "request_header")
	requestMethod := d.Get("request_method").(string)
	requestURL := d.Get("request_url").(string)
	requestVersion := d.Get("request_version").(string)
	responseBody := d.Get("response_body").(string)
	responseStatusCodes := interface2Int32List(d.Get("response_status_codes").([]interface{}))
	lbHTTPMonitor := loadbalancer.LbHttpMonitor{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		FallCount:           fallCount,
		Interval:            interval,
		MonitorPort:         monitorPort,
		RiseCount:           riseCount,
		Timeout:             timeout,
		RequestBody:         requestBody,
		RequestHeaders:      requestHeaders,
		RequestMethod:       requestMethod,
		RequestUrl:          requestURL,
		RequestVersion:      requestVersion,
		ResponseBody:        responseBody,
		ResponseStatusCodes: responseStatusCodes,
	}

	lbHTTPMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerHttpMonitor(nsxClient.Context, lbHTTPMonitor)

	if err != nil {
		return fmt.Errorf("Error during LbHttpMonitor create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbHttpMonitor create: %v", resp.StatusCode)
	}
	d.SetId(lbHTTPMonitor.Id)

	return resourceNsxtLbHTTPMonitorRead(d, m)
}

func resourceNsxtLbHTTPMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbHTTPMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerHttpMonitor(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbHttpMonitor %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbHttpMonitor read: %v", err)
	}

	d.Set("revision", lbHTTPMonitor.Revision)
	d.Set("description", lbHTTPMonitor.Description)
	d.Set("display_name", lbHTTPMonitor.DisplayName)
	setTagsInSchema(d, lbHTTPMonitor.Tags)
	d.Set("fall_count", lbHTTPMonitor.FallCount)
	d.Set("interval", lbHTTPMonitor.Interval)
	d.Set("monitor_port", lbHTTPMonitor.MonitorPort)
	d.Set("rise_count", lbHTTPMonitor.RiseCount)
	d.Set("timeout", lbHTTPMonitor.Timeout)
	d.Set("request_body", lbHTTPMonitor.RequestBody)
	setLbHTTPHeaderInSchema(d, "request_header", lbHTTPMonitor.RequestHeaders)
	d.Set("request_method", lbHTTPMonitor.RequestMethod)
	d.Set("request_url", lbHTTPMonitor.RequestUrl)
	d.Set("request_version", lbHTTPMonitor.RequestVersion)
	d.Set("response_body", lbHTTPMonitor.ResponseBody)
	d.Set("response_status_codes", int32List2Interface(lbHTTPMonitor.ResponseStatusCodes))

	return nil
}

func resourceNsxtLbHTTPMonitorUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	revision := int32(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	monitorPort := d.Get("monitor_port").(string)
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	requestBody := d.Get("request_body").(string)
	requestHeaders := getLbHTTPHeaderFromSchema(d, "request_header")
	requestMethod := d.Get("request_method").(string)
	requestURL := d.Get("request_url").(string)
	requestVersion := d.Get("request_version").(string)
	responseBody := d.Get("response_body").(string)
	responseStatusCodes := interface2Int32List(d.Get("response_status_codes").([]interface{}))
	lbHTTPMonitor := loadbalancer.LbHttpMonitor{
		Revision:            revision,
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		FallCount:           fallCount,
		Interval:            interval,
		MonitorPort:         monitorPort,
		RiseCount:           riseCount,
		Timeout:             timeout,
		RequestBody:         requestBody,
		RequestHeaders:      requestHeaders,
		RequestMethod:       requestMethod,
		RequestUrl:          requestURL,
		RequestVersion:      requestVersion,
		ResponseBody:        responseBody,
		ResponseStatusCodes: responseStatusCodes,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerHttpMonitor(nsxClient.Context, id, lbHTTPMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbHttpMonitor update: %v", err)
	}

	return resourceNsxtLbHTTPMonitorRead(d, m)
}
