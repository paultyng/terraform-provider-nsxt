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

func resourceNsxtLbPassiveMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbPassiveMonitorCreate,
		Read:   resourceNsxtLbPassiveMonitorRead,
		Update: resourceNsxtLbPassiveMonitorUpdate,
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
			"tag": getTagsSchema(),
			"max_fails": {
				Type:        schema.TypeInt,
				Description: "When the consecutive failures reach this value, then the member is considered temporarily unavailable for a configurable period",
				Optional:    true,
				Default:     5,
			},
			"timeout": {
				Type:        schema.TypeInt,
				Description: "After this timeout period, the member is tried again for a new connection to see if it is available",
				Optional:    true,
				Default:     5,
			},
		},
	}
}

func resourceNsxtLbPassiveMonitorCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	maxFails := int64(d.Get("max_fails").(int))
	timeout := int64(d.Get("timeout").(int))
	lbPassiveMonitor := loadbalancer.LbPassiveMonitor{
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		MaxFails:    maxFails,
		Timeout:     timeout,
	}

	lbPassiveMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerPassiveMonitor(nsxClient.Context, lbPassiveMonitor)

	if err != nil {
		return fmt.Errorf("Error during LbMonitor create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbMonitor create: %v", resp.StatusCode)
	}
	d.SetId(lbPassiveMonitor.Id)

	return resourceNsxtLbPassiveMonitorRead(d, m)
}

func resourceNsxtLbPassiveMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbPassiveMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerPassiveMonitor(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbMonitor %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbMonitor read: %v", err)
	}

	d.Set("revision", lbPassiveMonitor.Revision)
	d.Set("description", lbPassiveMonitor.Description)
	d.Set("display_name", lbPassiveMonitor.DisplayName)
	setTagsInSchema(d, lbPassiveMonitor.Tags)
	d.Set("max_fails", lbPassiveMonitor.MaxFails)
	d.Set("timeout", lbPassiveMonitor.Timeout)

	return nil
}

func resourceNsxtLbPassiveMonitorUpdate(d *schema.ResourceData, m interface{}) error {
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
	maxFails := int64(d.Get("max_fails").(int))
	timeout := int64(d.Get("timeout").(int))
	lbPassiveMonitor := loadbalancer.LbPassiveMonitor{
		Revision:    revision,
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		MaxFails:    maxFails,
		Timeout:     timeout,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerPassiveMonitor(nsxClient.Context, id, lbPassiveMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbMonitor update: %v", err)
	}

	return resourceNsxtLbPassiveMonitorRead(d, m)
}
