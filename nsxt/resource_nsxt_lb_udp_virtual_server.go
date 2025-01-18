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

func resourceNsxtLbUDPVirtualServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbUDPVirtualServerCreate,
		Read:   resourceNsxtLbUDPVirtualServerRead,
		Update: resourceNsxtLbUDPVirtualServerUpdate,
		Delete: resourceNsxtLbUDPVirtualServerDelete,
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
			"access_log_enabled": {
				Type:        schema.TypeBool,
				Description: "Whether access log is enabled",
				Optional:    true,
				Default:     false,
			},
			"application_profile_id": {
				Type:        schema.TypeString,
				Description: "The tcp application profile defines the application protocol characteristics",
				Required:    true,
				// TODO: add validation for fast UDP profile only
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "whether the virtual server is enabled",
				Optional:    true,
				Default:     true,
			},
			"ip_address": {
				Type:        schema.TypeString,
				Description: "virtual server IP address",
				Required:    true,
			},
			"ports": {
				Type:        schema.TypeList,
				Description: "Single port, multiple ports or port ranges",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
				Required: true,
			},
			"default_pool_member_ports": {
				Type:        schema.TypeList,
				Description: "Default pool member ports or port range",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
				Optional: true,
			},
			"max_concurrent_connections": {
				Type:        schema.TypeInt,
				Description: "If not specified, connections are unlimited",
				Optional:    true,
			},
			"max_new_connection_rate": {
				Type:        schema.TypeInt,
				Description: "If not specified, connection rate is unlimited",
				Optional:    true,
			},
			"persistence_profile_id": {
				Type:        schema.TypeString,
				Description: "Persistence profile is used to allow related client connections to be sent to the same backend server. Source ip persistence is supported.",
				Optional:    true,
				// TODO: add validation for source IP persistence only
			},
			"pool_id": {
				Type:        schema.TypeString,
				Description: "Server pool for backend connections",
				Optional:    true,
			},
			"sorry_pool_id": {
				Type:        schema.TypeString,
				Description: "When load balancer can not select a backend server to serve the request in default pool, the request would be served by sorry server pool",
				Optional:    true,
			},
		},
	}
}

func resourceNsxtLbUDPVirtualServerCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	accessLogEnabled := d.Get("access_log_enabled").(bool)
	applicationProfileID := d.Get("application_profile_id").(string)
	defaultPoolMemberPorts := interface2StringList(d.Get("default_pool_member_ports").([]interface{}))
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	persistenceProfileID := d.Get("persistence_profile_id").(string)
	poolID := d.Get("pool_id").(string)
	ports := interface2StringList(d.Get("ports").([]interface{}))
	sorryPoolID := d.Get("sorry_pool_id").(string)
	lbVirtualServer := loadbalancer.LbVirtualServer{
		Description:              description,
		DisplayName:              displayName,
		Tags:                     tags,
		AccessLogEnabled:         accessLogEnabled,
		ApplicationProfileId:     applicationProfileID,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  enabled,
		IpAddress:                ipAddress,
		IpProtocol:               "UDP",
		MaxConcurrentConnections: maxConcurrentConnections,
		MaxNewConnectionRate:     maxNewConnectionRate,
		PersistenceProfileId:     persistenceProfileID,
		PoolId:                   poolID,
		Ports:                    ports,
		SorryPoolId:              sorryPoolID,
	}

	lbVirtualServer, resp, err := nsxClient.ServicesApi.CreateLoadBalancerVirtualServer(nsxClient.Context, lbVirtualServer)

	if err != nil {
		return fmt.Errorf("Error during LbVirtualServer create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbVirtualServer create: %v", resp.StatusCode)
	}
	d.SetId(lbVirtualServer.Id)

	return resourceNsxtLbUDPVirtualServerRead(d, m)
}

func resourceNsxtLbUDPVirtualServerRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbVirtualServer, resp, err := nsxClient.ServicesApi.ReadLoadBalancerVirtualServer(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbVirtualServer %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbVirtualServer read: %v", err)
	}

	d.Set("revision", lbVirtualServer.Revision)
	d.Set("description", lbVirtualServer.Description)
	d.Set("display_name", lbVirtualServer.DisplayName)
	setTagsInSchema(d, lbVirtualServer.Tags)
	d.Set("access_log_enabled", lbVirtualServer.AccessLogEnabled)
	d.Set("application_profile_id", lbVirtualServer.ApplicationProfileId)
	d.Set("default_pool_member_ports", lbVirtualServer.DefaultPoolMemberPorts)
	d.Set("enabled", lbVirtualServer.Enabled)
	d.Set("ip_address", lbVirtualServer.IpAddress)
	d.Set("max_concurrent_connections", lbVirtualServer.MaxConcurrentConnections)
	d.Set("max_new_connection_rate", lbVirtualServer.MaxNewConnectionRate)
	d.Set("persistence_profile_id", lbVirtualServer.PersistenceProfileId)
	d.Set("pool_id", lbVirtualServer.PoolId)
	d.Set("ports", lbVirtualServer.Ports)
	d.Set("sorry_pool_id", lbVirtualServer.SorryPoolId)

	return nil
}

func resourceNsxtLbUDPVirtualServerUpdate(d *schema.ResourceData, m interface{}) error {
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
	accessLogEnabled := d.Get("access_log_enabled").(bool)
	applicationProfileID := d.Get("application_profile_id").(string)
	defaultPoolMemberPorts := interface2StringList(d.Get("default_pool_member_ports").([]interface{}))
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	persistenceProfileID := d.Get("persistence_profile_id").(string)
	poolID := d.Get("pool_id").(string)
	ports := interface2StringList(d.Get("ports").([]interface{}))
	sorryPoolID := d.Get("sorry_pool_id").(string)
	lbVirtualServer := loadbalancer.LbVirtualServer{
		Revision:                 revision,
		Description:              description,
		DisplayName:              displayName,
		Tags:                     tags,
		AccessLogEnabled:         accessLogEnabled,
		ApplicationProfileId:     applicationProfileID,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  enabled,
		IpAddress:                ipAddress,
		IpProtocol:               "UDP",
		MaxConcurrentConnections: maxConcurrentConnections,
		MaxNewConnectionRate:     maxNewConnectionRate,
		PersistenceProfileId:     persistenceProfileID,
		PoolId:                   poolID,
		Ports:                    ports,
		SorryPoolId:              sorryPoolID,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerVirtualServer(nsxClient.Context, id, lbVirtualServer)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbVirtualServer update: %v", err)
	}

	return resourceNsxtLbUDPVirtualServerRead(d, m)
}

// TODO: move this and other common VS code to utils
func resourceNsxtLbUDPVirtualServerDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerVirtualServer(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LbVirtualServer delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbVirtualServer %s not found", id)
		d.SetId("")
	}
	return nil
}
