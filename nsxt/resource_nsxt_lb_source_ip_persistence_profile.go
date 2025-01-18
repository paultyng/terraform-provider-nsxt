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

func resourceNsxtLbSourceIPPersistenceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbSourceIPPersistenceProfileCreate,
		Read:   resourceNsxtLbSourceIPPersistenceProfileRead,
		Update: resourceNsxtLbSourceIPPersistenceProfileUpdate,
		Delete: resourceNsxtLbSourceIPPersistenceProfileDelete,
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
			"persistence_shared": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether the cookie persistence is private or shared",
				Optional:    true,
			},
			"ha_persistence_mirroring": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether persistence entries will be synchronized to the HA peer",
				Optional:    true,
				Default:     false,
			},
			"purge_when_full": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether entries will be purged when the persistence table is full",
				Optional:    true,
				Default:     true,
			},
			"timeout": {
				Type:        schema.TypeInt,
				Description: "Persistence expiration time in seconds, counted from the time all the connections are completed",
				Optional:    true,
				Default:     300,
			},
		},
	}
}

func resourceNsxtLbSourceIPPersistenceProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	persistenceShared := d.Get("persistence_shared").(bool)
	haPersistenceMirroring := d.Get("ha_persistence_mirroring").(bool)
	purgeFlag := d.Get("purge_when_full").(bool)
	purge := "FULL"
	if !purgeFlag {
		purge = "NO_PURGE"
	}
	timeout := int64(d.Get("timeout").(int))
	lbSourceIPPersistenceProfile := loadbalancer.LbSourceIpPersistenceProfile{
		Description:                   description,
		DisplayName:                   displayName,
		Tags:                          tags,
		PersistenceShared:             persistenceShared,
		HaPersistenceMirroringEnabled: haPersistenceMirroring,
		Purge:                         purge,
		Timeout:                       timeout,
	}

	lbSourceIPPersistenceProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerSourceIpPersistenceProfile(nsxClient.Context, lbSourceIPPersistenceProfile)

	if err != nil {
		return fmt.Errorf("Error during LbSourceIPPersistenceProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbSourceIPPersistenceProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbSourceIPPersistenceProfile.Id)

	return resourceNsxtLbSourceIPPersistenceProfileRead(d, m)
}

func resourceNsxtLbSourceIPPersistenceProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbSourceIPPersistenceProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerSourceIpPersistenceProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbSourceIPPersistenceProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbSourceIPPersistenceProfile read: %v", err)
	}

	d.Set("revision", lbSourceIPPersistenceProfile.Revision)
	d.Set("description", lbSourceIPPersistenceProfile.Description)
	d.Set("display_name", lbSourceIPPersistenceProfile.DisplayName)
	setTagsInSchema(d, lbSourceIPPersistenceProfile.Tags)
	d.Set("persistence_shared", lbSourceIPPersistenceProfile.PersistenceShared)
	d.Set("ha_persistence_mirroring", lbSourceIPPersistenceProfile.HaPersistenceMirroringEnabled)
	if lbSourceIPPersistenceProfile.Purge == "FULL" {
		d.Set("purge_when_full", true)
	} else {
		d.Set("purge_when_full", false)
	}
	d.Set("timeout", lbSourceIPPersistenceProfile.Timeout)

	return nil
}

func resourceNsxtLbSourceIPPersistenceProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	persistenceShared := d.Get("persistence_shared").(bool)
	haPersistenceMirroring := d.Get("ha_persistence_mirroring").(bool)
	purgeFlag := d.Get("purge_when_full").(bool)
	purge := "FULL"
	if !purgeFlag {
		purge = "NO_PURGE"
	}
	timeout := int64(d.Get("timeout").(int))
	lbSourceIPPersistenceProfile := loadbalancer.LbSourceIpPersistenceProfile{
		Revision:                      revision,
		Description:                   description,
		DisplayName:                   displayName,
		Tags:                          tags,
		PersistenceShared:             persistenceShared,
		HaPersistenceMirroringEnabled: haPersistenceMirroring,
		Purge:                         purge,
		Timeout:                       timeout,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerSourceIpPersistenceProfile(nsxClient.Context, id, lbSourceIPPersistenceProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbSourceIPPersistenceProfile update: %v", err)
	}

	return resourceNsxtLbSourceIPPersistenceProfileRead(d, m)
}

func resourceNsxtLbSourceIPPersistenceProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerPersistenceProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbSourceIPPersistenceProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbSourceIPPersistenceProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
