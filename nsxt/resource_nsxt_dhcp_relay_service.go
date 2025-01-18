// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtDhcpRelayService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtDhcpRelayServiceCreate,
		Read:   resourceNsxtDhcpRelayServiceRead,
		Update: resourceNsxtDhcpRelayServiceUpdate,
		Delete: resourceNsxtDhcpRelayServiceDelete,
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
			"dhcp_relay_profile_id": {
				Type:        schema.TypeString,
				Description: "DHCP relay profile referenced by the dhcp relay service",
				Required:    true,
			},
		},
	}
}

func resourceNsxtDhcpRelayServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	dhcpRelayProfileID := d.Get("dhcp_relay_profile_id").(string)
	dhcpRelayService := manager.DhcpRelayService{
		Description:        description,
		DisplayName:        displayName,
		Tags:               tags,
		DhcpRelayProfileId: dhcpRelayProfileID,
	}

	dhcpRelayService, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateDhcpRelay(nsxClient.Context, dhcpRelayService)

	if err != nil {
		return fmt.Errorf("Error during DhcpRelayService create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during DhcpRelayService create: %v", resp.StatusCode)
	}
	d.SetId(dhcpRelayService.Id)

	return resourceNsxtDhcpRelayServiceRead(d, m)
}

func resourceNsxtDhcpRelayServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay service id")
	}

	dhcpRelayService, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelay(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayService %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayService read: %v", err)
	}

	d.Set("revision", dhcpRelayService.Revision)
	d.Set("description", dhcpRelayService.Description)
	d.Set("display_name", dhcpRelayService.DisplayName)
	setTagsInSchema(d, dhcpRelayService.Tags)
	d.Set("dhcp_relay_profile_id", dhcpRelayService.DhcpRelayProfileId)

	return nil
}

func resourceNsxtDhcpRelayServiceUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay service id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	dhcpRelayProfileID := d.Get("dhcp_relay_profile_id").(string)
	dhcpRelayService := manager.DhcpRelayService{
		Revision:           revision,
		Description:        description,
		DisplayName:        displayName,
		Tags:               tags,
		DhcpRelayProfileId: dhcpRelayProfileID,
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateDhcpRelay(nsxClient.Context, id, dhcpRelayService)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during DhcpRelayService update: %v", err)
	}

	return resourceNsxtDhcpRelayServiceRead(d, m)
}

func resourceNsxtDhcpRelayServiceDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay service id")
	}

	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteDhcpRelay(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayService delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayService %s not found", id)
		d.SetId("")
	}
	return nil
}
