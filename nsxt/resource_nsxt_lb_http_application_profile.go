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

var xForwardedValues = []string{"INSERT", "REPLACE"}

func resourceNsxtLbHTTPApplicationProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbHTTPApplicationProfileCreate,
		Read:   resourceNsxtLbHTTPApplicationProfileRead,
		Update: resourceNsxtLbHTTPApplicationProfileUpdate,
		Delete: resourceNsxtLbHTTPApplicationProfileDelete,
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
			"http_redirect_to": {
				Type:        schema.TypeString,
				Description: "A URL that incoming requests for that virtual server can be temporarily redirected to, If a website is temporarily down or has moved",
				Optional:    true,
			},
			"http_redirect_to_https": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether the client will automatically be redirected to use SSL",
				Optional:    true,
				Default:     false,
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Description:  "Timeout in seconds to specify how long an HTTP application can remain idle",
				Optional:     true,
				Default:      15,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ntlm": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether NTLM challenge/response methodology will be used over HTTP",
				Optional:    true,
				Default:     false,
			},
			"request_body_size": {
				Type:        schema.TypeInt,
				Description: "Maximum request body size in bytes (Unlimited if not specified)",
				Optional:    true,
			},
			"request_header_size": {
				Type:         schema.TypeInt,
				Description:  "Maximum request header size in bytes. Requests with larger header size will be processed as best effort whereas a request with header below this specified size is guaranteed to be processed",
				Optional:     true,
				Default:      1024,
				ValidateFunc: validation.IntBetween(1, 65536),
			},
			"response_timeout": {
				Type:         schema.TypeInt,
				Description:  "Number of seconds waiting for the server response before the connection is closed",
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"x_forwarded_for": {
				Type:         schema.TypeString,
				Description:  "When this value is set, the x_forwarded_for header in the incoming request will be inserted or replaced",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(xForwardedValues, false),
			},
		},
	}
}

func resourceNsxtLbHTTPApplicationProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	httpRedirectTo := d.Get("http_redirect_to").(string)
	httpRedirectToHTTPS := d.Get("http_redirect_to_https").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	ntlm := d.Get("ntlm").(bool)
	requestBodySize := int64(d.Get("request_body_size").(int))
	requestHeaderSize := int64(d.Get("request_header_size").(int))
	responseTimeout := int64(d.Get("response_timeout").(int))
	xForwardedFor := d.Get("x_forwarded_for").(string)
	lbHTTPApplicationProfile := loadbalancer.LbHttpProfile{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		HttpRedirectTo:      httpRedirectTo,
		HttpRedirectToHttps: httpRedirectToHTTPS,
		IdleTimeout:         idleTimeout,
		Ntlm:                ntlm,
		RequestBodySize:     requestBodySize,
		RequestHeaderSize:   requestHeaderSize,
		ResponseTimeout:     responseTimeout,
		XForwardedFor:       xForwardedFor,
	}

	lbHTTPApplicationProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerHttpProfile(nsxClient.Context, lbHTTPApplicationProfile)

	if err != nil {
		return fmt.Errorf("Error during LbHTTPApplicationProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbHTTPApplicationProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbHTTPApplicationProfile.Id)

	return resourceNsxtLbHTTPApplicationProfileRead(d, m)
}

func resourceNsxtLbHTTPApplicationProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbHTTPApplicationProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerHttpProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbHTTPApplicationProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbHTTPApplicationProfile read: %v", err)
	}

	d.Set("revision", lbHTTPApplicationProfile.Revision)
	d.Set("description", lbHTTPApplicationProfile.Description)
	d.Set("display_name", lbHTTPApplicationProfile.DisplayName)
	setTagsInSchema(d, lbHTTPApplicationProfile.Tags)
	d.Set("http_redirect_to", lbHTTPApplicationProfile.HttpRedirectTo)
	d.Set("http_redirect_to_https", lbHTTPApplicationProfile.HttpRedirectToHttps)
	d.Set("idle_timeout", lbHTTPApplicationProfile.IdleTimeout)
	d.Set("ntlm", lbHTTPApplicationProfile.Ntlm)
	d.Set("request_body_size", lbHTTPApplicationProfile.RequestBodySize)
	d.Set("request_header_size", lbHTTPApplicationProfile.RequestHeaderSize)
	d.Set("response_timeout", lbHTTPApplicationProfile.ResponseTimeout)
	d.Set("x_forwarded_for", lbHTTPApplicationProfile.XForwardedFor)

	return nil
}

func resourceNsxtLbHTTPApplicationProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	httpRedirectTo := d.Get("http_redirect_to").(string)
	httpRedirectToHTTPS := d.Get("http_redirect_to_https").(bool)
	idleTimeout := int64(d.Get("idle_timeout").(int))
	ntlm := d.Get("ntlm").(bool)
	requestBodySize := int64(d.Get("request_body_size").(int))
	requestHeaderSize := int64(d.Get("request_header_size").(int))
	responseTimeout := int64(d.Get("response_timeout").(int))
	xForwardedFor := d.Get("x_forwarded_for").(string)
	lbHTTPApplicationProfile := loadbalancer.LbHttpProfile{
		Revision:            revision,
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		HttpRedirectTo:      httpRedirectTo,
		HttpRedirectToHttps: httpRedirectToHTTPS,
		IdleTimeout:         idleTimeout,
		Ntlm:                ntlm,
		RequestBodySize:     requestBodySize,
		RequestHeaderSize:   requestHeaderSize,
		ResponseTimeout:     responseTimeout,
		XForwardedFor:       xForwardedFor,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerHttpProfile(nsxClient.Context, id, lbHTTPApplicationProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbHTTPApplicationProfile update: %v", err)
	}

	return resourceNsxtLbHTTPApplicationProfileRead(d, m)
}

func resourceNsxtLbHTTPApplicationProfileDelete(d *schema.ResourceData, m interface{}) error {
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
		return fmt.Errorf("Error during LbHTTPApplicationProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbHTTPApplicationProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
