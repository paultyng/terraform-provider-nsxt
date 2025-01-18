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

func resourceNsxtLbFastUDPApplicationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbFastUDPApplicationProfileCreate,
		Read:   resourceNsxtLbFastUDPApplicationProfileRead,
		Update: resourceNsxtLbFastUDPApplicationProfileUpdate,
		Delete: resourceNsxtLbFastUDPApplicationProfileDelete,
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
			"idle_timeout": {
				Type:         schema.TypeInt,
				Description:  "Timeout in seconds to specify how long an idle UDP connection in ESTABLISHED state should be kept for this application before cleaning up",
				Optional:     true,
				Default:      300,
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

func resourceNsxtLbFastUDPApplicationProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	haFlowMirroringEnabled := d.Get("ha_flow_mirroring").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	lbFastUDPProfile := loadbalancer.LbFastUdpProfile{
		Description:          description,
		DisplayName:          displayName,
		Tags:                 tags,
		IdleTimeout:          idleTimeout,
		FlowMirroringEnabled: haFlowMirroringEnabled,
	}

	lbFastUDPProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerFastUdpProfile(nsxClient.Context, lbFastUDPProfile)
	if resp != nil && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbFastUdpProfile create: %v", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("Error during LbFastUdpProfile create: %v", err)
	}

	d.SetId(lbFastUDPProfile.Id)

	return resourceNsxtLbFastUDPApplicationProfileRead(d, m)
}

func resourceNsxtLbFastUDPApplicationProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbFastUDPProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerFastUdpProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbFastUdpProfile read: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbFastUdpProfile %s not found", id)
		d.SetId("")
		return nil
	}
	d.Set("revision", lbFastUDPProfile.Revision)
	d.Set("description", lbFastUDPProfile.Description)
	d.Set("display_name", lbFastUDPProfile.DisplayName)
	setTagsInSchema(d, lbFastUDPProfile.Tags)
	d.Set("ha_flow_mirroring", lbFastUDPProfile.FlowMirroringEnabled)
	d.Set("idle_timeout", lbFastUDPProfile.IdleTimeout)

	return nil
}

func resourceNsxtLbFastUDPApplicationProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	haFlowMirroringEnabled := d.Get("ha_flow_mirroring").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	lbFastUDPProfile := loadbalancer.LbFastUdpProfile{
		Revision:             revision,
		Description:          description,
		DisplayName:          displayName,
		Tags:                 tags,
		IdleTimeout:          idleTimeout,
		FlowMirroringEnabled: haFlowMirroringEnabled,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerFastUdpProfile(nsxClient.Context, id, lbFastUDPProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbFastUdpProfile update: %v", err)
	}

	return resourceNsxtLbFastUDPApplicationProfileRead(d, m)
}

func resourceNsxtLbFastUDPApplicationProfileDelete(d *schema.ResourceData, m interface{}) error {
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
		return fmt.Errorf("Error during LbFastUdpProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbFastUdpProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
