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

func resourceNsxtLbClientSslProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbClientSslProfileCreate,
		Read:   resourceNsxtLbClientSslProfileRead,
		Update: resourceNsxtLbClientSslProfileUpdate,
		Delete: resourceNsxtLbClientSslProfileDelete,
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
			"prefer_server_ciphers": {
				Type:        schema.TypeBool,
				Description: "Allow server to override the client's preference",
				Optional:    true,
				Default:     false,
			},
			"protocols": getSSLProtocolsSchema(),
			"session_cache_enabled": {
				Type:        schema.TypeBool,
				Description: "Reuse previously negotiated security parameters during handshake",
				Optional:    true,
				Default:     true,
			},
			"session_cache_timeout": {
				Type:        schema.TypeInt,
				Description: "For how long the SSL session parameters can be reused",
				Optional:    true,
				Default:     300,
			},
		},
	}
}

func resourceNsxtLbClientSslProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	preferServerCiphers := d.Get("prefer_server_ciphers").(bool)
	protocols := getStringListFromSchemaSet(d, "protocols")
	sessionCacheEnabled := d.Get("session_cache_enabled").(bool)
	sessionCacheTimeout := int64(d.Get("session_cache_timeout").(int))
	lbClientSslProfile := loadbalancer.LbClientSslProfile{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		Ciphers:             ciphers,
		PreferServerCiphers: preferServerCiphers,
		Protocols:           protocols,
		SessionCacheEnabled: sessionCacheEnabled,
		SessionCacheTimeout: sessionCacheTimeout,
	}

	lbClientSslProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerClientSslProfile(nsxClient.Context, lbClientSslProfile)

	if err != nil {
		return fmt.Errorf("Error during LbClientSslProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbClientSslProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbClientSslProfile.Id)

	return resourceNsxtLbClientSslProfileRead(d, m)
}

func resourceNsxtLbClientSslProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbClientSslProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerClientSslProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbClientSslProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbClientSslProfile read: %v", err)
	}

	d.Set("revision", lbClientSslProfile.Revision)
	d.Set("description", lbClientSslProfile.Description)
	d.Set("display_name", lbClientSslProfile.DisplayName)
	setTagsInSchema(d, lbClientSslProfile.Tags)
	d.Set("ciphers", lbClientSslProfile.Ciphers)
	d.Set("is_secure", lbClientSslProfile.IsSecure)
	d.Set("prefer_server_ciphers", lbClientSslProfile.PreferServerCiphers)
	d.Set("protocols", lbClientSslProfile.Protocols)
	d.Set("session_cache_enabled", lbClientSslProfile.SessionCacheEnabled)
	d.Set("session_cache_timeout", lbClientSslProfile.SessionCacheTimeout)

	return nil
}

func resourceNsxtLbClientSslProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	preferServerCiphers := d.Get("prefer_server_ciphers").(bool)
	protocols := getStringListFromSchemaSet(d, "protocols")
	sessionCacheEnabled := d.Get("session_cache_enabled").(bool)
	sessionCacheTimeout := int64(d.Get("session_cache_timeout").(int))
	lbClientSslProfile := loadbalancer.LbClientSslProfile{
		Revision:            revision,
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		Ciphers:             ciphers,
		PreferServerCiphers: preferServerCiphers,
		Protocols:           protocols,
		SessionCacheEnabled: sessionCacheEnabled,
		SessionCacheTimeout: sessionCacheTimeout,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerClientSslProfile(nsxClient.Context, id, lbClientSslProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbClientSslProfile update: %v", err)
	}

	return resourceNsxtLbClientSslProfileRead(d, m)
}

func resourceNsxtLbClientSslProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerClientSslProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbClientSslProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbClientSslProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
