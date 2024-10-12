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

func resourceNsxtIPBlock() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIPBlockCreate,
		Read:   resourceNsxtIPBlockRead,
		Update: resourceNsxtIPBlockUpdate,
		Delete: resourceNsxtIPBlockDelete,
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
			"cidr": {
				Type:        schema.TypeString,
				Description: "Represents network address and the prefix length which will be associated with a layer-2 broadcast domain",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceNsxtIPBlockCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	cidr := d.Get("cidr").(string)
	ipBlock := manager.IpBlock{
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		Cidr:        cidr,
	}
	// Create the IP Block
	ipBlock, resp, err := nsxClient.PoolManagementApi.CreateIpBlock(nsxClient.Context, ipBlock)

	if err != nil {
		return fmt.Errorf("Error during IpBlock create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during IpBlock create: %v", resp.StatusCode)
	}
	d.SetId(ipBlock.Id)

	return resourceNsxtIPBlockRead(d, m)
}

func resourceNsxtIPBlockRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	ipBlock, resp, err := nsxClient.PoolManagementApi.ReadIpBlock(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpBlock %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during IpBlock read: %v", err)
	}

	d.Set("revision", ipBlock.Revision)
	d.Set("description", ipBlock.Description)
	d.Set("display_name", ipBlock.DisplayName)
	setTagsInSchema(d, ipBlock.Tags)
	d.Set("cidr", ipBlock.Cidr)

	return nil
}

func resourceNsxtIPBlockUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	cidr := d.Get("cidr").(string)
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	ipBlock := manager.IpBlock{
		DisplayName: displayName,
		Description: description,
		Cidr:        cidr,
		Tags:        tags,
		Revision:    revision,
	}
	// Update the IP block
	_, resp, err := nsxClient.PoolManagementApi.UpdateIpBlock(nsxClient.Context, id, ipBlock)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during IpBlock update: %v", err)
	}

	return resourceNsxtIPBlockRead(d, m)
}

func resourceNsxtIPBlockDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}
	resp, err := nsxClient.PoolManagementApi.DeleteIpBlock(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during IpBlock delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpBlock %s not found", id)
		d.SetId("")
	}
	return nil
}
