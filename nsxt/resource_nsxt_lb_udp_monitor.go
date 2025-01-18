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

func resourceNsxtLbUDPMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbUDPMonitorCreate,
		Read:   resourceNsxtLbUDPMonitorRead,
		Update: resourceNsxtLbUDPMonitorUpdate,
		Delete: resourceNsxtLbMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema:             getLbL4MonitorSchema("udp"),
	}
}

func resourceNsxtLbUDPMonitorCreate(d *schema.ResourceData, m interface{}) error {
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
	lbUDPMonitor := loadbalancer.LbUdpMonitor{
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

	lbUDPMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerUdpMonitor(nsxClient.Context, lbUDPMonitor)

	if err != nil {
		return fmt.Errorf("Error during LbMonitor create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbMonitor create: %v", resp.StatusCode)
	}
	d.SetId(lbUDPMonitor.Id)

	return resourceNsxtLbUDPMonitorRead(d, m)
}

func resourceNsxtLbUDPMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbUDPMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerUdpMonitor(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbMonitor %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbMonitor read: %v", err)
	}

	d.Set("revision", lbUDPMonitor.Revision)
	d.Set("description", lbUDPMonitor.Description)
	d.Set("display_name", lbUDPMonitor.DisplayName)
	setTagsInSchema(d, lbUDPMonitor.Tags)
	d.Set("fall_count", lbUDPMonitor.FallCount)
	d.Set("interval", lbUDPMonitor.Interval)
	d.Set("monitor_port", lbUDPMonitor.MonitorPort)
	d.Set("rise_count", lbUDPMonitor.RiseCount)
	d.Set("timeout", lbUDPMonitor.Timeout)
	d.Set("receive", lbUDPMonitor.Receive)
	d.Set("send", lbUDPMonitor.Send)

	return nil
}

func resourceNsxtLbUDPMonitorUpdate(d *schema.ResourceData, m interface{}) error {
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
	lbUDPMonitor := loadbalancer.LbUdpMonitor{
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

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerUdpMonitor(nsxClient.Context, id, lbUDPMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbMonitor update: %v", err)
	}

	return resourceNsxtLbUDPMonitorRead(d, m)
}
