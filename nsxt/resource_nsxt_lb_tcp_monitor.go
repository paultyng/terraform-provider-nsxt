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

func resourceNsxtLbTCPMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbTCPMonitorCreate,
		Read:   resourceNsxtLbTCPMonitorRead,
		Update: resourceNsxtLbTCPMonitorUpdate,
		Delete: resourceNsxtLbMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema:             getLbL4MonitorSchema("tcp"),
	}
}

func resourceNsxtLbTCPMonitorCreate(d *schema.ResourceData, m interface{}) error {
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
	receive := d.Get("receive").(string)
	send := d.Get("send").(string)
	lbTCPMonitor := loadbalancer.LbTcpMonitor{
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		FallCount:   fallCount,
		Interval:    interval,
		MonitorPort: monitorPort,
		RiseCount:   riseCount,
		Timeout:     timeout,
		Receive:     receive,
		Send:        send,
	}

	lbTCPMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerTcpMonitor(nsxClient.Context, lbTCPMonitor)

	if err != nil {
		return fmt.Errorf("Error during LbMonitor create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbMonitor create: %v", resp.StatusCode)
	}
	d.SetId(lbTCPMonitor.Id)

	return resourceNsxtLbTCPMonitorRead(d, m)
}

func resourceNsxtLbTCPMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbTCPMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerTcpMonitor(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbMonitor %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbMonitor read: %v", err)
	}

	d.Set("revision", lbTCPMonitor.Revision)
	d.Set("description", lbTCPMonitor.Description)
	d.Set("display_name", lbTCPMonitor.DisplayName)
	setTagsInSchema(d, lbTCPMonitor.Tags)
	d.Set("fall_count", lbTCPMonitor.FallCount)
	d.Set("interval", lbTCPMonitor.Interval)
	d.Set("monitor_port", lbTCPMonitor.MonitorPort)
	d.Set("rise_count", lbTCPMonitor.RiseCount)
	d.Set("timeout", lbTCPMonitor.Timeout)
	d.Set("receive", lbTCPMonitor.Receive)
	d.Set("send", lbTCPMonitor.Send)

	return nil
}

func resourceNsxtLbTCPMonitorUpdate(d *schema.ResourceData, m interface{}) error {
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
	receive := d.Get("receive").(string)
	send := d.Get("send").(string)
	lbTCPMonitor := loadbalancer.LbTcpMonitor{
		Revision:    revision,
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		FallCount:   fallCount,
		Interval:    interval,
		MonitorPort: monitorPort,
		RiseCount:   riseCount,
		Timeout:     timeout,
		Receive:     receive,
		Send:        send,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerTcpMonitor(nsxClient.Context, id, lbTCPMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbMonitor update: %v", err)
	}

	return resourceNsxtLbTCPMonitorRead(d, m)
}
