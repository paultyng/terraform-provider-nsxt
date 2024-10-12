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

func resourceNsxtLbHTTPSMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbHTTPSMonitorCreate,
		Read:   resourceNsxtLbHTTPSMonitorRead,
		Update: resourceNsxtLbHTTPSMonitorUpdate,
		Delete: resourceNsxtLbMonitorDelete,
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
			"tag":                     getTagsSchema(),
			"fall_count":              getLbMonitorFallCountSchema(),
			"interval":                getLbMonitorIntervalSchema(),
			"monitor_port":            getLbMonitorPortSchema(),
			"rise_count":              getLbMonitorRiseCountSchema(),
			"timeout":                 getLbMonitorTimeoutSchema(),
			"certificate_chain_depth": getCertificateChainDepthSchema(),
			"ciphers":                 getSSLCiphersSchema(),
			"client_certificate_id": {
				Type:        schema.TypeString,
				Description: "client certificate can be specified to support client authentication",
				Optional:    true,
			},
			"is_secure":             getIsSecureSchema(),
			"protocols":             getSSLProtocolsSchema(),
			"request_body":          getLbMonitorRequestBodySchema(),
			"request_header":        getLbHTTPHeaderSchema("Array of HTTP request headers"),
			"request_method":        getLbMonitorRequestMethodSchema(),
			"request_url":           getLbMonitorRequestURLSchema(),
			"request_version":       getLbMonitorRequestVersionSchema(),
			"response_body":         getLbMonitorResponseBodySchema(),
			"response_status_codes": getLbMonitorResponseStatusCodesSchema(),
			"server_auth": {
				Type:         schema.TypeString,
				Description:  "Server authentication mode",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"REQUIRED", "IGNORE"}, false),
				Default:      "IGNORE",
			},
			"server_auth_ca_ids":  getIDSetSchema("If server auth type is REQUIRED, server certificate must be signed by one of the CAs"),
			"server_auth_crl_ids": getIDSetSchema("Certificate Revocation List (CRL) to disallow compromised server certificates"),
		},
	}
}

func resourceNsxtLbHTTPSMonitorCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	monitorPort := d.Get("monitor_port").(string)
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	certificateChainDepth := int64(d.Get("certificate_chain_depth").(int))
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	clientCertificateID := d.Get("client_certificate_id").(string)
	isSecure := d.Get("is_secure").(bool)
	protocols := getStringListFromSchemaSet(d, "protocols")
	requestBody := d.Get("request_body").(string)
	requestHeaders := getLbHTTPHeaderFromSchema(d, "request_header")
	requestMethod := d.Get("request_method").(string)
	requestURL := d.Get("request_url").(string)
	requestVersion := d.Get("request_version").(string)
	responseBody := d.Get("response_body").(string)
	responseStatusCodes := interface2Int32List(d.Get("response_status_codes").([]interface{}))
	serverAuth := d.Get("server_auth").(string)
	serverAuthCaIds := getStringListFromSchemaSet(d, "server_auth_ca_ids")
	serverAuthCrlIds := getStringListFromSchemaSet(d, "server_auth_crl_ids")
	lbHTTPSMonitor := loadbalancer.LbHttpsMonitor{
		Description:           description,
		DisplayName:           displayName,
		Tags:                  tags,
		FallCount:             fallCount,
		Interval:              interval,
		MonitorPort:           monitorPort,
		RiseCount:             riseCount,
		Timeout:               timeout,
		CertificateChainDepth: certificateChainDepth,
		Ciphers:               ciphers,
		ClientCertificateId:   clientCertificateID,
		IsSecure:              isSecure,
		Protocols:             protocols,
		RequestBody:           requestBody,
		RequestHeaders:        requestHeaders,
		RequestMethod:         requestMethod,
		RequestUrl:            requestURL,
		RequestVersion:        requestVersion,
		ResponseBody:          responseBody,
		ResponseStatusCodes:   responseStatusCodes,
		ServerAuth:            serverAuth,
		ServerAuthCaIds:       serverAuthCaIds,
		ServerAuthCrlIds:      serverAuthCrlIds,
	}

	lbHTTPSMonitor, resp, err := nsxClient.ServicesApi.CreateLoadBalancerHttpsMonitor(nsxClient.Context, lbHTTPSMonitor)

	if err != nil {
		return fmt.Errorf("Error during LbHttpsMonitor create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbHttpsMonitor create: %v", resp.StatusCode)
	}
	d.SetId(lbHTTPSMonitor.Id)

	return resourceNsxtLbHTTPSMonitorRead(d, m)
}

func resourceNsxtLbHTTPSMonitorRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbHTTPSMonitor, resp, err := nsxClient.ServicesApi.ReadLoadBalancerHttpsMonitor(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbHttpsMonitor %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbHttpsMonitor read: %v", err)
	}

	d.Set("revision", lbHTTPSMonitor.Revision)
	d.Set("description", lbHTTPSMonitor.Description)
	d.Set("display_name", lbHTTPSMonitor.DisplayName)
	setTagsInSchema(d, lbHTTPSMonitor.Tags)
	d.Set("fall_count", lbHTTPSMonitor.FallCount)
	d.Set("interval", lbHTTPSMonitor.Interval)
	d.Set("monitor_port", lbHTTPSMonitor.MonitorPort)
	d.Set("rise_count", lbHTTPSMonitor.RiseCount)
	d.Set("timeout", lbHTTPSMonitor.Timeout)
	d.Set("certificate_chain_depth", lbHTTPSMonitor.CertificateChainDepth)
	d.Set("ciphers", lbHTTPSMonitor.Ciphers)
	d.Set("client_certificate_id", lbHTTPSMonitor.ClientCertificateId)
	d.Set("is_secure", lbHTTPSMonitor.IsSecure)
	d.Set("protocols", lbHTTPSMonitor.Protocols)
	d.Set("request_body", lbHTTPSMonitor.RequestBody)
	setLbHTTPHeaderInSchema(d, "request_header", lbHTTPSMonitor.RequestHeaders)
	d.Set("request_method", lbHTTPSMonitor.RequestMethod)
	d.Set("request_url", lbHTTPSMonitor.RequestUrl)
	d.Set("request_version", lbHTTPSMonitor.RequestVersion)
	d.Set("response_body", lbHTTPSMonitor.ResponseBody)
	d.Set("response_status_codes", int32List2Interface(lbHTTPSMonitor.ResponseStatusCodes))
	d.Set("server_auth", lbHTTPSMonitor.ServerAuth)
	d.Set("server_auth_ca_ids", lbHTTPSMonitor.ServerAuthCaIds)
	d.Set("server_auth_crl_ids", lbHTTPSMonitor.ServerAuthCrlIds)

	return nil
}

func resourceNsxtLbHTTPSMonitorUpdate(d *schema.ResourceData, m interface{}) error {
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
	fallCount := int64(d.Get("fall_count").(int))
	interval := int64(d.Get("interval").(int))
	monitorPort := d.Get("monitor_port").(string)
	riseCount := int64(d.Get("rise_count").(int))
	timeout := int64(d.Get("timeout").(int))
	certificateChainDepth := int64(d.Get("certificate_chain_depth").(int))
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	clientCertificateID := d.Get("client_certificate_id").(string)
	isSecure := d.Get("is_secure").(bool)
	protocols := getStringListFromSchemaSet(d, "protocols")
	requestBody := d.Get("request_body").(string)
	requestHeaders := getLbHTTPHeaderFromSchema(d, "request_header")
	requestMethod := d.Get("request_method").(string)
	requestURL := d.Get("request_url").(string)
	requestVersion := d.Get("request_version").(string)
	responseBody := d.Get("response_body").(string)
	responseStatusCodes := interface2Int32List(d.Get("response_status_codes").([]interface{}))
	serverAuth := d.Get("server_auth").(string)
	serverAuthCaIds := getStringListFromSchemaSet(d, "server_auth_ca_ids")
	serverAuthCrlIds := getStringListFromSchemaSet(d, "server_auth_crl_ids")
	lbHTTPSMonitor := loadbalancer.LbHttpsMonitor{
		Revision:              revision,
		Description:           description,
		DisplayName:           displayName,
		Tags:                  tags,
		FallCount:             fallCount,
		Interval:              interval,
		MonitorPort:           monitorPort,
		RiseCount:             riseCount,
		Timeout:               timeout,
		CertificateChainDepth: certificateChainDepth,
		Ciphers:               ciphers,
		ClientCertificateId:   clientCertificateID,
		IsSecure:              isSecure,
		Protocols:             protocols,
		RequestBody:           requestBody,
		RequestHeaders:        requestHeaders,
		RequestMethod:         requestMethod,
		RequestUrl:            requestURL,
		RequestVersion:        requestVersion,
		ResponseBody:          responseBody,
		ResponseStatusCodes:   responseStatusCodes,
		ServerAuth:            serverAuth,
		ServerAuthCaIds:       serverAuthCaIds,
		ServerAuthCrlIds:      serverAuthCrlIds,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerHttpsMonitor(nsxClient.Context, id, lbHTTPSMonitor)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbHttpsMonitor update: %v", err)
	}

	return resourceNsxtLbHTTPSMonitorRead(d, m)
}
