// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

func resourceNsxtLbIcmpMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbIcmpMonitorCreate,
		Read:   resourceNsxtLbIcmpMonitorRead,
		Update: resourceNsxtLbIcmpMonitorUpdate,
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
			"tag":          getTagsSchema(),
			"fall_count":   getLbMonitorFallCountSchema(),
			"interval":     getLbMonitorIntervalSchema(),
			"monitor_port": getLbMonitorPortSchema(),
			"rise_count":   getLbMonitorRiseCountSchema(),
			"timeout":      getLbMonitorTimeoutSchema(),
			"data_length": {
				Type:         schema.TypeInt,
				Description:  "The data size (in bytes) of the ICMP healthcheck packet",
				Optional:     true,
				Default:      56,
				ValidateFunc: validation.IntBetween(0, 65507),
			},
		},
	}
}

func resourceNsxtLbIcmpMonitorCreate(d *schema.ResourceData, m interface{}) error {
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
	dataLength := int64(d.Get("data_length").(int))
	lbIcmpMonitor := loadbalancer.LbIcmpMonitor{
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		FallCount:   fallCount,
		Interval:    interval,
		MonitorPort: monitorPort,
		RiseCount:   riseCount,
		Timeout:     timeout,
		DataLength:  dataLength,
	}

	lbIcmpMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerIcmpMonitor(nsxClient.Context, lbIcmpMonitor)
	if resp != nil && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbMonitor create: %v", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("Error during LbMonitor create: %v", err)
	}

	d.SetId(lbIcmpMonitor.Id)

	return resourceNsxtLbIcmpMonitorRead(d, m)
}

func resourceNsxtLbIcmpMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbIcmpMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerIcmpMonitor(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbMonitor read: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbMonitor %s not found", id)
		d.SetId("")
		return nil
	}

	d.Set("revision", lbIcmpMonitor.Revision)
	d.Set("description", lbIcmpMonitor.Description)
	d.Set("display_name", lbIcmpMonitor.DisplayName)
	setTagsInSchema(d, lbIcmpMonitor.Tags)
	d.Set("fall_count", lbIcmpMonitor.FallCount)
	d.Set("interval", lbIcmpMonitor.Interval)
	d.Set("monitor_port", lbIcmpMonitor.MonitorPort)
	d.Set("rise_count", lbIcmpMonitor.RiseCount)
	d.Set("timeout", lbIcmpMonitor.Timeout)
	d.Set("data_length", lbIcmpMonitor.DataLength)

	return nil
}

func resourceNsxtLbIcmpMonitorUpdate(d *schema.ResourceData, m interface{}) error {
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
	dataLength := int64(d.Get("data_length").(int))
	lbIcmpMonitor := loadbalancer.LbIcmpMonitor{
		Revision:    revision,
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		FallCount:   fallCount,
		Interval:    interval,
		MonitorPort: monitorPort,
		RiseCount:   riseCount,
		Timeout:     timeout,
		DataLength:  dataLength,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerIcmpMonitor(nsxClient.Context, id, lbIcmpMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbMonitor update: %v", err)
	}

	return resourceNsxtLbIcmpMonitorRead(d, m)
}
