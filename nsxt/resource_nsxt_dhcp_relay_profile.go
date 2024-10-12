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

func resourceNsxtDhcpRelayProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtDhcpRelayProfileCreate,
		Read:   resourceNsxtDhcpRelayProfileRead,
		Update: resourceNsxtDhcpRelayProfileUpdate,
		Delete: resourceNsxtDhcpRelayProfileDelete,
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
			"server_addresses": {
				Type:        schema.TypeSet,
				Description: "Set of dhcp relay server addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
				Required: true,
			},
		},
	}
}

func resourceNsxtDhcpRelayProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	serverAddresses := getStringListFromSchemaSet(d, "server_addresses")
	dhcpRelayProfile := manager.DhcpRelayProfile{
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		ServerAddresses: serverAddresses,
	}

	dhcpRelayProfile, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateDhcpRelayProfile(nsxClient.Context, dhcpRelayProfile)

	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during DhcpRelayProfile create: %v", resp.StatusCode)
	}
	d.SetId(dhcpRelayProfile.Id)

	return resourceNsxtDhcpRelayProfileRead(d, m)
}

func resourceNsxtDhcpRelayProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	dhcpRelayProfile, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelayProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile read: %v", err)
	}

	d.Set("revision", dhcpRelayProfile.Revision)
	d.Set("description", dhcpRelayProfile.Description)
	d.Set("display_name", dhcpRelayProfile.DisplayName)
	setTagsInSchema(d, dhcpRelayProfile.Tags)
	d.Set("server_addresses", dhcpRelayProfile.ServerAddresses)

	return nil
}

func resourceNsxtDhcpRelayProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	serverAddresses := interface2StringList(d.Get("server_addresses").(*schema.Set).List())
	dhcpRelayProfile := manager.DhcpRelayProfile{
		Revision:        revision,
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		ServerAddresses: serverAddresses,
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateDhcpRelayProfile(nsxClient.Context, id, dhcpRelayProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during DhcpRelayProfile update: %v", err)
	}

	return resourceNsxtDhcpRelayProfileRead(d, m)
}

func resourceNsxtDhcpRelayProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining dhcp relay profile id")
	}

	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteDhcpRelayProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpRelayProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpRelayProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
