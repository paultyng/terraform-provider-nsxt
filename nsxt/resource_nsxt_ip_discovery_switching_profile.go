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
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtIPDiscoverySwitchingProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIPDiscoverySwitchingProfileCreate,
		Read:   resourceNsxtIPDiscoverySwitchingProfileRead,
		Update: resourceNsxtIPDiscoverySwitchingProfileUpdate,
		Delete: resourceNsxtIPDiscoverySwitchingProfileDelete,
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
			"vm_tools_enabled": {
				Type:        schema.TypeBool,
				Description: "Indicating whether VM tools will be enabled. This option is only supported on ESX where vm-tools is installed",
				Optional:    true,
				Default:     false,
			},
			"arp_snooping_enabled": {
				Type:        schema.TypeBool,
				Description: "Indicates whether ARP snooping is enabled",
				Optional:    true,
				Default:     false,
			},
			"dhcp_snooping_enabled": {
				Type:        schema.TypeBool,
				Description: "Indicates whether DHCP snooping is enabled",
				Optional:    true,
				Default:     false,
			},
			"arp_bindings_limit": {
				Type:         schema.TypeInt,
				Description:  "Limit for the amount of ARP bindings",
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func resourceNsxtIPDiscoverySwitchingProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	dhcpSnoopingEnabled := d.Get("dhcp_snooping_enabled").(bool)
	arpSnoopingEnabled := d.Get("arp_snooping_enabled").(bool)
	arpBindingsLimit := d.Get("arp_bindings_limit").(int)
	vmToolsEnabled := d.Get("vm_tools_enabled").(bool)

	switchingProfile := manager.IpDiscoverySwitchingProfile{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		DhcpSnoopingEnabled: dhcpSnoopingEnabled,
		ArpSnoopingEnabled:  arpSnoopingEnabled,
		ArpBindingsLimit:    arpBindingsLimit,
		VmToolsEnabled:      vmToolsEnabled,
	}

	switchingProfile, resp, err := nsxClient.LogicalSwitchingApi.CreateIpDiscoverySwitchingProfile(nsxClient.Context, switchingProfile)

	if err != nil {
		return fmt.Errorf("Error during IPDiscoverySwitchingProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during IPDiscoverySwitchingProfile create: %v", resp.StatusCode)
	}
	d.SetId(switchingProfile.Id)

	return resourceNsxtIPDiscoverySwitchingProfileRead(d, m)
}

func resourceNsxtIPDiscoverySwitchingProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	switchingProfile, resp, err := nsxClient.LogicalSwitchingApi.GetIpDiscoverySwitchingProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IPDiscoverySwitchingProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during IPDiscoverySwitchingProfile read: %v", err)
	}

	d.Set("revision", switchingProfile.Revision)
	d.Set("description", switchingProfile.Description)
	d.Set("display_name", switchingProfile.DisplayName)
	d.Set("dhcp_snooping_enabled", switchingProfile.DhcpSnoopingEnabled)
	d.Set("arp_snooping_enabled", switchingProfile.ArpSnoopingEnabled)
	d.Set("arp_bindings_limit", switchingProfile.ArpBindingsLimit)
	d.Set("vm_tools_enabled", switchingProfile.VmToolsEnabled)
	setTagsInSchema(d, switchingProfile.Tags)

	return nil
}

func resourceNsxtIPDiscoverySwitchingProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	dhcpSnoopingEnabled := d.Get("dhcp_snooping_enabled").(bool)
	arpSnoopingEnabled := d.Get("arp_snooping_enabled").(bool)
	arpBindingsLimit := d.Get("arp_bindings_limit").(int)
	vmToolsEnabled := d.Get("vm_tools_enabled").(bool)

	switchingProfile := manager.IpDiscoverySwitchingProfile{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		DhcpSnoopingEnabled: dhcpSnoopingEnabled,
		ArpSnoopingEnabled:  arpSnoopingEnabled,
		ArpBindingsLimit:    arpBindingsLimit,
		VmToolsEnabled:      vmToolsEnabled,
		Revision:            revision,
	}

	_, resp, err := nsxClient.LogicalSwitchingApi.UpdateIpDiscoverySwitchingProfile(nsxClient.Context, id, switchingProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during IPDiscoverySwitchingProfile update: %v", err)
	}

	return resourceNsxtIPDiscoverySwitchingProfileRead(d, m)
}

func resourceNsxtIPDiscoverySwitchingProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.LogicalSwitchingApi.DeleteSwitchingProfile(nsxClient.Context, id, nil)
	if err != nil {
		return fmt.Errorf("Error during IPDiscoverySwitchingProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IPDiscoverySwitchingProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
