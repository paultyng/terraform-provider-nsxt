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

func resourceNsxtLbFastTCPApplicationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbFastTCPApplicationProfileCreate,
		Read:   resourceNsxtLbFastTCPApplicationProfileRead,
		Update: resourceNsxtLbFastTCPApplicationProfileUpdate,
		Delete: resourceNsxtLbFastTCPApplicationProfileDelete,
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
			"close_timeout": {
				Type:         schema.TypeInt,
				Description:  "Timeout in seconds to specify how long a closed TCP connection should be kept for this application before cleaning up the connection",
				Optional:     true,
				Default:      8,
				ValidateFunc: validation.IntBetween(1, 60),
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Description:  "Timeout in seconds to specify how long an idle TCP connection in ESTABLISHED state should be kept for this application before cleaning up",
				Optional:     true,
				Default:      1800,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ha_flow_mirroring": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether flow mirroring is enabled, and all the flows to the bounded virtual server are mirrored to the standby node",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceNsxtLbFastTCPApplicationProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	closeTimeout := int64(d.Get("close_timeout").(int))
	haFlowMirroringEnabled := d.Get("ha_flow_mirroring").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	lbFastTCPProfile := loadbalancer.LbFastTcpProfile{
		Description:            description,
		DisplayName:            displayName,
		Tags:                   tags,
		CloseTimeout:           closeTimeout,
		HaFlowMirroringEnabled: haFlowMirroringEnabled,
		IdleTimeout:            idleTimeout,
	}

	lbFastTCPProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerFastTcpProfile(nsxClient.Context, lbFastTCPProfile)

	if err != nil {
		return fmt.Errorf("Error during LbFastTcpProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbFastTcpProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbFastTCPProfile.Id)

	return resourceNsxtLbFastTCPApplicationProfileRead(d, m)
}

func resourceNsxtLbFastTCPApplicationProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbFastTCPProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerFastTcpProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbFastTcpProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbFastTcpProfile read: %v", err)
	}

	d.Set("revision", lbFastTCPProfile.Revision)
	d.Set("description", lbFastTCPProfile.Description)
	d.Set("display_name", lbFastTCPProfile.DisplayName)
	setTagsInSchema(d, lbFastTCPProfile.Tags)
	d.Set("close_timeout", lbFastTCPProfile.CloseTimeout)
	d.Set("ha_flow_mirroring", lbFastTCPProfile.HaFlowMirroringEnabled)
	d.Set("idle_timeout", lbFastTCPProfile.IdleTimeout)

	return nil
}

func resourceNsxtLbFastTCPApplicationProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	closeTimeout := int64(d.Get("close_timeout").(int))
	haFlowMirroringEnabled := d.Get("ha_flow_mirroring").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	lbFastTCPProfile := loadbalancer.LbFastTcpProfile{
		Revision:               revision,
		Description:            description,
		DisplayName:            displayName,
		Tags:                   tags,
		CloseTimeout:           closeTimeout,
		HaFlowMirroringEnabled: haFlowMirroringEnabled,
		IdleTimeout:            idleTimeout,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerFastTcpProfile(nsxClient.Context, id, lbFastTCPProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbFastTcpProfile update: %v", err)
	}

	return resourceNsxtLbFastTCPApplicationProfileRead(d, m)
}

func resourceNsxtLbFastTCPApplicationProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerApplicationProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbFastTcpProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbFastTcpProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
