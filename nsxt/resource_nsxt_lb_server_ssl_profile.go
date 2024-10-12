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

func resourceNsxtLbServerSslProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbServerSslProfileCreate,
		Read:   resourceNsxtLbServerSslProfileRead,
		Update: resourceNsxtLbServerSslProfileUpdate,
		Delete: resourceNsxtLbServerSslProfileDelete,
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
			"tag":       getTagsSchema(),
			"ciphers":   getSSLCiphersSchema(),
			"is_secure": getIsSecureSchema(),
			"protocols": getSSLProtocolsSchema(),
			"session_cache_enabled": {
				Type:        schema.TypeBool,
				Description: "Reuse previously negotiated security parameters during handshake",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func resourceNsxtLbServerSslProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	protocols := getStringListFromSchemaSet(d, "protocols")
	sessionCacheEnabled := d.Get("session_cache_enabled").(bool)
	lbServerSslProfile := loadbalancer.LbServerSslProfile{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		Ciphers:             ciphers,
		Protocols:           protocols,
		SessionCacheEnabled: sessionCacheEnabled,
	}

	lbServerSslProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerServerSslProfile(nsxClient.Context, lbServerSslProfile)

	if err != nil {
		return fmt.Errorf("Error during LbServerSslProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbServerSslProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbServerSslProfile.Id)

	return resourceNsxtLbServerSslProfileRead(d, m)
}

func resourceNsxtLbServerSslProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbServerSslProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerServerSslProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbServerSslProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbServerSslProfile read: %v", err)
	}

	d.Set("revision", lbServerSslProfile.Revision)
	d.Set("description", lbServerSslProfile.Description)
	d.Set("display_name", lbServerSslProfile.DisplayName)
	setTagsInSchema(d, lbServerSslProfile.Tags)
	d.Set("ciphers", lbServerSslProfile.Ciphers)
	d.Set("is_secure", lbServerSslProfile.IsSecure)
	d.Set("protocols", lbServerSslProfile.Protocols)
	d.Set("session_cache_enabled", lbServerSslProfile.SessionCacheEnabled)

	return nil
}

func resourceNsxtLbServerSslProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	protocols := getStringListFromSchemaSet(d, "protocols")
	sessionCacheEnabled := d.Get("session_cache_enabled").(bool)
	lbServerSslProfile := loadbalancer.LbServerSslProfile{
		Revision:            revision,
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		Ciphers:             ciphers,
		Protocols:           protocols,
		SessionCacheEnabled: sessionCacheEnabled,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerServerSslProfile(nsxClient.Context, id, lbServerSslProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbServerSslProfile update: %v", err)
	}

	return resourceNsxtLbServerSslProfileRead(d, m)
}

func resourceNsxtLbServerSslProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerServerSslProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbServerSslProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbServerSslProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
