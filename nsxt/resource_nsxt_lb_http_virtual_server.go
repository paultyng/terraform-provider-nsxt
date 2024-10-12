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

func resourceNsxtLbHTTPVirtualServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbHTTPVirtualServerCreate,
		Read:   resourceNsxtLbHTTPVirtualServerRead,
		Update: resourceNsxtLbHTTPVirtualServerUpdate,
		Delete: resourceNsxtLbHTTPVirtualServerDelete,
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
				Description: "The http application profile defines the application protocol characteristics",
				Required:    true,
			},
			"client_ssl": getLbClientSSLBindingSchema(),
			"default_pool_member_port": {
				Type:         schema.TypeString,
				Description:  "Default pool member port",
				ValidateFunc: validateSinglePort(),
				Optional:     true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the virtual server is enabled",
				Optional:    true,
				Default:     true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "Virtual server IP address",
				ValidateFunc: validateSingleIP(),
				Required:     true,
			},
			"port": {
				Type:         schema.TypeString,
				Description:  "Virtual server port",
				ValidateFunc: validateSinglePort(),
				Required:     true,
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
				Description: "Persistence profile is used to allow related client connections to be sent to the same backend server",
				Optional:    true,
			},
			"pool_id": {
				Type:        schema.TypeString,
				Description: "Server pool for backend connections",
				Optional:    true,
			},
			"rule_ids": {
				Type:        schema.TypeList,
				Description: "Customization of load balancing behavior using match/action rules",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"sorry_pool_id": {
				Type:        schema.TypeString,
				Description: "When load balancer can not select a backend server to serve the request in default pool or pool in rules, the request would be served by sorry server pool",
				Optional:    true,
			},
			"server_ssl": getLbServerSSLBindingSchema(),
		},
	}
}

func getLbClientSSLBindingSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Client SSL settings for Virtual Server",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"client_ssl_profile_id": {
					Type:        schema.TypeString,
					Description: "Id of client SSL profile that defines reusable properties",
					Required:    true,
				},
				"default_certificate_id": {
					Type:        schema.TypeString,
					Description: "Id of certificate that will be used if the server does not host multiple hostnames on the same IP address or if the client does not support SNI extension",
					Required:    true,
				},
				"certificate_chain_depth": getCertificateChainDepthSchema(),
				"client_auth": {
					Type:        schema.TypeBool,
					Description: "Whether client certificate authentication is mandatory",
					Optional:    true,
					Default:     false,
				},
				"ca_ids":              getIDSetSchema("List of CA ids for client authentication"),
				"crl_ids":             getIDSetSchema("List of CRL ids for client authentication"),
				"sni_certificate_ids": getIDSetSchema("List of certificates to serve different hostnames"),
			},
		},
	}
}

func getLbServerSSLBindingSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Server SSL settings for Virtual Server",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"server_ssl_profile_id": {
					Type:        schema.TypeString,
					Description: "Id of server SSL profile that defines reusable properties",
					Required:    true,
				},
				"client_certificate_id": {
					Type:        schema.TypeString,
					Description: "Id of certificate that will be used if the server does not host multiple hostnames on the same IP address or if the client does not support SNI extension",
					Optional:    true,
				},
				"certificate_chain_depth": getCertificateChainDepthSchema(),
				"server_auth": {
					Type:        schema.TypeBool,
					Description: "Server authentication mode",
					Optional:    true,
					Default:     false,
				},
				"ca_ids":  getIDSetSchema("List of CA ids for server authentication"),
				"crl_ids": getIDSetSchema("List of CRL ids for server authentication"),
			},
		},
	}
}

func getAuthFromBool(value bool) string {
	if value {
		return "REQUIRED"
	}
	return "IGNORE"
}

func getBoolFromAuth(value string) bool {
	return value != "IGNORE"
}

func getClientSSLBindingFromSchema(d *schema.ResourceData) *loadbalancer.ClientSslProfileBinding {
	bindings := d.Get("client_ssl").([]interface{})
	for _, binding := range bindings {
		// maximum count is one
		data := binding.(map[string]interface{})
		profileBinding := loadbalancer.ClientSslProfileBinding{
			CertificateChainDepth: int64(data["certificate_chain_depth"].(int)),
			ClientAuth:            getAuthFromBool(data["client_auth"].(bool)),
			ClientAuthCaIds:       interface2StringList(data["ca_ids"].(*schema.Set).List()),
			ClientAuthCrlIds:      interface2StringList(data["crl_ids"].(*schema.Set).List()),
			DefaultCertificateId:  data["default_certificate_id"].(string),
			SniCertificateIds:     interface2StringList(data["sni_certificate_ids"].(*schema.Set).List()),
			SslProfileId:          data["client_ssl_profile_id"].(string),
		}

		return &profileBinding
	}

	return nil
}

func setClientSSLBindingInSchema(d *schema.ResourceData, binding *loadbalancer.ClientSslProfileBinding) {
	var bindingList []map[string]interface{}
	if binding != nil {
		elem := make(map[string]interface{})
		elem["certificate_chain_depth"] = binding.CertificateChainDepth
		elem["client_auth"] = getBoolFromAuth(binding.ClientAuth)
		elem["ca_ids"] = binding.ClientAuthCaIds
		elem["crl_ids"] = binding.ClientAuthCrlIds
		elem["default_certificate_id"] = binding.DefaultCertificateId
		elem["sni_certificate_ids"] = binding.SniCertificateIds
		elem["client_ssl_profile_id"] = binding.SslProfileId

		bindingList = append(bindingList, elem)
	}

	err := d.Set("client_ssl", bindingList)
	if err != nil {
		log.Printf("[WARNING]: Failed to set client SSL in schema: %v", err)
	}
}

func getServerSSLBindingFromSchema(d *schema.ResourceData) *loadbalancer.ServerSslProfileBinding {
	bindings := d.Get("server_ssl").([]interface{})
	for _, binding := range bindings {
		// maximum count is one
		data := binding.(map[string]interface{})
		profileBinding := loadbalancer.ServerSslProfileBinding{
			CertificateChainDepth: int64(data["certificate_chain_depth"].(int)),
			ServerAuth:            getAuthFromBool(data["server_auth"].(bool)),
			ServerAuthCaIds:       interface2StringList(data["ca_ids"].(*schema.Set).List()),
			ServerAuthCrlIds:      interface2StringList(data["crl_ids"].(*schema.Set).List()),
			ClientCertificateId:   data["client_certificate_id"].(string),
			SslProfileId:          data["server_ssl_profile_id"].(string),
		}

		return &profileBinding
	}

	return nil
}

func setServerSSLBindingInSchema(d *schema.ResourceData, binding *loadbalancer.ServerSslProfileBinding) {
	var bindingList []map[string]interface{}
	if binding != nil {
		elem := make(map[string]interface{})
		elem["certificate_chain_depth"] = binding.CertificateChainDepth
		elem["server_auth"] = getBoolFromAuth(binding.ServerAuth)
		elem["ca_ids"] = binding.ServerAuthCaIds
		elem["crl_ids"] = binding.ServerAuthCrlIds
		elem["client_certificate_id"] = binding.ClientCertificateId
		elem["server_ssl_profile_id"] = binding.SslProfileId

		bindingList = append(bindingList, elem)
	}

	err := d.Set("server_ssl", bindingList)
	if err != nil {
		log.Printf("[WARNING]: Failed to set server SSL in schema: %v", err)
	}
}

func resourceNsxtLbHTTPVirtualServerCreate(d *schema.ResourceData, m interface{}) error {
	var defaultPoolMemberPorts []string
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	accessLogEnabled := d.Get("access_log_enabled").(bool)
	applicationProfileID := d.Get("application_profile_id").(string)
	clientSslProfileBinding := getClientSSLBindingFromSchema(d)
	defaultPoolMemberPort := d.Get("default_pool_member_port").(string)
	if defaultPoolMemberPort != "" {
		defaultPoolMemberPorts = append(defaultPoolMemberPorts, defaultPoolMemberPort)
	}
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	persistenceProfileID := d.Get("persistence_profile_id").(string)
	poolID := d.Get("pool_id").(string)
	ports := []string{d.Get("port").(string)}
	ruleIds := interface2StringList(d.Get("rule_ids").([]interface{}))
	serverSslProfileBinding := getServerSSLBindingFromSchema(d)
	sorryPoolID := d.Get("sorry_pool_id").(string)
	lbVirtualServer := loadbalancer.LbVirtualServer{
		Description:              description,
		DisplayName:              displayName,
		Tags:                     tags,
		AccessLogEnabled:         accessLogEnabled,
		ApplicationProfileId:     applicationProfileID,
		ClientSslProfileBinding:  clientSslProfileBinding,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  enabled,
		IpAddress:                ipAddress,
		IpProtocol:               "TCP",
		MaxConcurrentConnections: maxConcurrentConnections,
		MaxNewConnectionRate:     maxNewConnectionRate,
		PersistenceProfileId:     persistenceProfileID,
		PoolId:                   poolID,
		Ports:                    ports,
		RuleIds:                  ruleIds,
		ServerSslProfileBinding:  serverSslProfileBinding,
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

	return resourceNsxtLbHTTPVirtualServerRead(d, m)
}

func resourceNsxtLbHTTPVirtualServerRead(d *schema.ResourceData, m interface{}) error {
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
	setClientSSLBindingInSchema(d, lbVirtualServer.ClientSslProfileBinding)
	if len(lbVirtualServer.DefaultPoolMemberPorts) > 0 {
		d.Set("default_pool_member_port", lbVirtualServer.DefaultPoolMemberPorts[0])
	}
	d.Set("enabled", lbVirtualServer.Enabled)
	d.Set("ip_address", lbVirtualServer.IpAddress)
	d.Set("max_concurrent_connections", lbVirtualServer.MaxConcurrentConnections)
	d.Set("max_new_connection_rate", lbVirtualServer.MaxNewConnectionRate)
	d.Set("persistence_profile_id", lbVirtualServer.PersistenceProfileId)
	d.Set("pool_id", lbVirtualServer.PoolId)
	d.Set("port", lbVirtualServer.Ports[0])
	d.Set("rule_ids", lbVirtualServer.RuleIds)
	setServerSSLBindingInSchema(d, lbVirtualServer.ServerSslProfileBinding)
	d.Set("sorry_pool_id", lbVirtualServer.SorryPoolId)

	return nil
}

func resourceNsxtLbHTTPVirtualServerUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	var defaultPoolMemberPorts []string
	revision := int32(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	accessLogEnabled := d.Get("access_log_enabled").(bool)
	applicationProfileID := d.Get("application_profile_id").(string)
	clientSslProfileBinding := getClientSSLBindingFromSchema(d)
	defaultPoolMemberPort := d.Get("default_pool_member_port").(string)
	if defaultPoolMemberPort != "" {
		defaultPoolMemberPorts = append(defaultPoolMemberPorts, defaultPoolMemberPort)
	}
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	persistenceProfileID := d.Get("persistence_profile_id").(string)
	poolID := d.Get("pool_id").(string)
	ports := []string{d.Get("port").(string)}
	ruleIds := interface2StringList(d.Get("rule_ids").([]interface{}))
	serverSslProfileBinding := getServerSSLBindingFromSchema(d)
	sorryPoolID := d.Get("sorry_pool_id").(string)
	lbVirtualServer := loadbalancer.LbVirtualServer{
		Revision:                 revision,
		Description:              description,
		DisplayName:              displayName,
		Tags:                     tags,
		AccessLogEnabled:         accessLogEnabled,
		ApplicationProfileId:     applicationProfileID,
		ClientSslProfileBinding:  clientSslProfileBinding,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  enabled,
		IpAddress:                ipAddress,
		IpProtocol:               "TCP",
		MaxConcurrentConnections: maxConcurrentConnections,
		MaxNewConnectionRate:     maxNewConnectionRate,
		PersistenceProfileId:     persistenceProfileID,
		PoolId:                   poolID,
		Ports:                    ports,
		RuleIds:                  ruleIds,
		ServerSslProfileBinding:  serverSslProfileBinding,
		SorryPoolId:              sorryPoolID,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerVirtualServer(nsxClient.Context, id, lbVirtualServer)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbVirtualServer update: %v", err)
	}

	return resourceNsxtLbHTTPVirtualServerRead(d, m)
}

func resourceNsxtLbHTTPVirtualServerDelete(d *schema.ResourceData, m interface{}) error {
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
