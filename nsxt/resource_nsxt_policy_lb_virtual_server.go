/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var lbClientSSLModeValues = []string{
	model.LBClientSslProfileBinding_CLIENT_AUTH_REQUIRED,
	model.LBClientSslProfileBinding_CLIENT_AUTH_IGNORE,
}

var lbServerSSLModeValues = []string{
	model.LBServerSslProfileBinding_SERVER_AUTH_REQUIRED,
	model.LBServerSslProfileBinding_SERVER_AUTH_IGNORE,
	model.LBServerSslProfileBinding_SERVER_AUTH_AUTO_APPLY,
}

var lbAccessListControlActionValues = []string{
	model.LBAccessListControl_ACTION_ALLOW,
	model.LBAccessListControl_ACTION_DROP,
}

var lbRuleMatchStrategyValues = []string{
	model.LBRule_MATCH_STRATEGY_ALL,
	model.LBRule_MATCH_STRATEGY_ANY,
}

var lbRulePhaseValues = []string{
	model.LBRule_PHASE_HTTP_REQUEST_REWRITE,
	model.LBRule_PHASE_HTTP_FORWARDING,
	model.LBRule_PHASE_HTTP_RESPONSE_REWRITE,
	model.LBRule_PHASE_HTTP_ACCESS,
	model.LBRule_PHASE_TRANSPORT,
}

var lbHttpSslConditionUsedSslCipherValues = []string{
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_3DES_EDE_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_256_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_RSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_ECDSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_USED_SSL_CIPHER_ECDH_RSA_WITH_AES_256_GCM_SHA384,
}

var lbHttpSslConditionClientSupportedSslCiphersValues = []string{
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_256_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_3DES_EDE_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_256_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_RSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_ECDSA_WITH_AES_256_GCM_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_128_CBC_SHA,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_128_CBC_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_128_GCM_SHA256,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_256_CBC_SHA384,
	model.LBHttpSslCondition_CLIENT_SUPPORTED_SSL_CIPHERS_ECDH_RSA_WITH_AES_256_GCM_SHA384,
}

var lbHttpSslConditionSessionReusedValues = []string{
	model.LBHttpSslCondition_SESSION_REUSED_IGNORE,
	model.LBHttpSslCondition_SESSION_REUSED_REUSED,
	model.LBHttpSslCondition_SESSION_REUSED_NEW,
}

var lbHttpSslConditionUsedProtocolValues = []string{
	model.LBHttpSslCondition_USED_PROTOCOL_SSL_V2,
	model.LBHttpSslCondition_USED_PROTOCOL_SSL_V3,
	model.LBHttpSslCondition_USED_PROTOCOL_TLS_V1,
	model.LBHttpSslCondition_USED_PROTOCOL_TLS_V1_1,
	model.LBHttpSslCondition_USED_PROTOCOL_TLS_V1_2,
}

var lbSslModeSelectionActionValues = []string{
	model.LBSslModeSelectionAction_SSL_MODE_PASSTHROUGH,
	model.LBSslModeSelectionAction_SSL_MODE_END_TO_END,
	model.LBSslModeSelectionAction_SSL_MODE_OFFLOAD,
}

func resourceNsxtPolicyLBVirtualServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBVirtualServerCreate,
		Read:   resourceNsxtPolicyLBVirtualServerRead,
		Update: resourceNsxtPolicyLBVirtualServerUpdate,
		Delete: resourceNsxtPolicyLBVirtualServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":                   getNsxIDSchema(),
			"path":                     getPathSchema(),
			"display_name":             getDisplayNameSchema(),
			"description":              getDescriptionSchema(),
			"revision":                 getRevisionSchema(),
			"tag":                      getTagsSchema(),
			"application_profile_path": getPolicyPathSchema(true, false, "Application profile for this virtual server"),
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Flag to enable Virtual Server",
				Optional:    true,
				Default:     true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "Virtual Server IP address",
				Required:     true,
				ValidateFunc: validateSingleIP(),
			},
			"ports": {
				Type:        schema.TypeList,
				Description: "Virtual Server ports",
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
			},
			"persistence_profile_path": getPolicyPathSchema(false, false, "Path to persistence profile allowing related client connections to be sent to the same backend server."),
			"service_path":             getPolicyPathSchema(false, false, "Virtual Server can be associated with Load Balancer Service"),
			"default_pool_member_ports": {
				Type:        schema.TypeList,
				Description: "Default pool member ports when member port is not defined",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
				MaxItems: 14,
			},
			"access_log_enabled": {
				Type:        schema.TypeBool,
				Description: "If enabled, all connections/requests sent to virtual server are logged to the access log file",
				Optional:    true,
				Default:     false,
			},
			"max_concurrent_connections": {
				Type:         schema.TypeInt,
				Description:  "To ensure one virtual server does not over consume resources, connections to a virtual server can be capped.",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_new_connection_rate": {
				Type:         schema.TypeInt,
				Description:  "To ensure one virtual server does not over consume resources, connections to a member can be rate limited.",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"pool_path":       getPolicyPathSchema(false, false, "Path for Load Balancer Pool"),
			"sorry_pool_path": getPolicyPathSchema(false, false, "When load balancer can not select server in default pool or pool in rules, the request would be served by sorry pool"),
			"client_ssl": {
				Type:        schema.TypeList,
				Description: "This setting is used when load balancer terminates client SSL connection",
				Elem:        getPolicyLbClientSSLBindingSchema(),
				Optional:    true,
				MaxItems:    1,
			},
			"server_ssl": {
				Type:        schema.TypeList,
				Description: "This setting is used when load balancer establishes connection to the backend server",
				Elem:        getPolicyLbServerSSLBindingSchema(),
				Optional:    true,
				MaxItems:    1,
			},
			"log_significant_event_only": {
				Type:        schema.TypeBool,
				Description: "Flag to log significant events in access log, if access log is enabed",
				Optional:    true,
				Default:     false,
			},
			"access_list_control": {
				Type:        schema.TypeList,
				Description: "IP access list control for filtering the connections from clients",
				Elem:        getPolicyLbAccessListControlSchema(),
				Optional:    true,
				MaxItems:    1,
			},
			"rule": {
				Type:        schema.TypeList,
				Description: "List of load balancer rules for layer 7 virtual servers with LBHttpProfile",
				Optional:    true,
				MaxItems:    4000,
				Elem:        getPolicyLbRuleBindingSchema(),
			},
		},
	}
}

func getPolicyLbClientSSLBindingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_auth": {
				Type:         schema.TypeString,
				Description:  "Client authentication mode",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(lbClientSSLModeValues, false),
				Default:      model.LBClientSslProfileBinding_CLIENT_AUTH_IGNORE,
			},
			"certificate_chain_depth": {
				Type:         schema.TypeInt,
				Description:  "Certificate chain depth",
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ca_paths": {
				Type:        schema.TypeList,
				Description: "If client auth type is REQUIRED, client certificate must be signed by one Certificate Authorities",
				Optional:    true,
				Elem:        getElemPolicyPathSchema(),
			},
			"crl_paths": {
				Type:        schema.TypeList,
				Description: "Certificate Revocation Lists can be specified to disallow compromised certificates",
				Optional:    true,
				Elem:        getElemPolicyPathSchema(),
			},
			"default_certificate_path": getPolicyPathSchema(true, false, "Default Certificate Path"),
			"sni_paths": {
				Type:        schema.TypeList,
				Description: "This setting allows multiple certificates, for different hostnames, to be bound to the same virtual server",
				Optional:    true,
				Elem:        getElemPolicyPathSchema(),
			},
			"ssl_profile_path": getPolicyPathSchema(false, false, "Client SSL Profile Path"),
		},
	}
}

func getPolicyLbServerSSLBindingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server_auth": {
				Type:         schema.TypeString,
				Description:  "Server authentication mode",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(lbServerSSLModeValues, false),
				Default:      model.LBServerSslProfileBinding_SERVER_AUTH_AUTO_APPLY,
			},
			"client_certificate_path": getPolicyPathSchema(false, false, "Client certificat path for client authentication"),
			"certificate_chain_depth": {
				Type:         schema.TypeInt,
				Description:  "Certificate chain depth",
				Optional:     true,
				Default:      3,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"ca_paths": {
				Type:        schema.TypeList,
				Description: "If server auth type is REQUIRED, server certificate must be signed by one Certificate Authorities",
				Optional:    true,
				Elem:        getElemPolicyPathSchema(),
			},
			"crl_paths": {
				Type:        schema.TypeList,
				Description: "Certificate Revocation Lists can be specified disallow compromised certificates",
				Optional:    true,
				Elem:        getElemPolicyPathSchema(),
			},
			"ssl_profile_path": getPolicyPathSchema(false, false, "Server SSL Profile Path"),
		},
	}
}

func getPolicyLbAccessListControlSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Description:  "Action to apply to connections matching the group",
				Required:     true,
				ValidateFunc: validation.StringInSlice(lbAccessListControlActionValues, false),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Flag to enable access list control",
				Optional:    true,
				Default:     true,
			},
			"group_path": getPolicyPathSchema(true, false, "The path of grouping object which defines the IP addresses or ranges to match the client IP"),
		},
	}
}

func getPolicyLbRuleBindingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"display_name": getDisplayNameSchema(),
			"match_strategy": {
				Description:  "Match strategy for determining match of multiple conditions (ALL or ANY, default: ANY).",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(lbRuleMatchStrategyValues, false),
				Optional:     true,
				Default:      "ANY",
			},
			"phase": {
				Description:  "Load balancer processing phase, one of HTTP_REQUEST_REWRITE, HTTP_FORWARDING (Default), HTTP_RESPONSE_REWRITE, HTTP_ACCESS, TRANSPORT.",
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(lbRulePhaseValues, false),
				Optional:     true,
				Default:      "HTTP_FORWARDING",
			},
			"action": {
				Description: "A list of actions to be executed at specified phase when load balancer rule matches.",
				Type:        schema.TypeList,
				Optional:    false,
				Required:    true,
				Elem:        getPolicyLbRuleActionsSchema(),
			},
			"condition": {
				Description: "A list of match conditions used to match application traffic.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        getPolicyLbRuleConditionsSchema(),
			},
		},
	}
}

func getPolicyLbRuleActionsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"connection_drop":              getPolicyLbRuleConnectionDropActionSchema(),
			"http_redirect":                getPolicyLbRuleHttpRedirectActionSchema(),
			"http_reject":                  getPolicyLbRuleHttpRejectActionSchema(),
			"http_request_header_delete":   getPolicyLbRuleHttpMessageHeaderDeleteActionSchema("Action to delete header fields of HTTP request messages at HTTP_REQUEST_REWRITE phase."),
			"http_request_header_rewrite":  getPolicyLbRuleHttpMessageHeaderRewriteActionSchema("Action to rewrite header fields of HTTP request messages to specified new values at HTTP_REQUEST_REWRITE phase."),
			"http_request_uri_rewrite":     getPolicyLbRuleHttpRequestUriRewriteActionSchema(),
			"http_response_header_delete":  getPolicyLbRuleHttpMessageHeaderDeleteActionSchema("Action to delete header fields of HTTP response messages at HTTP_RESPONSE_REWRITE phase."),
			"http_response_header_rewrite": getPolicyLbRuleHttpMessageHeaderRewriteActionSchema("Action to rewrite header fields of HTTP response messages to specified new values at HTTP_RESPONSE_REWRITE phase."),
			"jwt_auth":                     getPolicyLbRuleJwtAuthActionSchema(),
			"select_pool":                  getPolicyLbRuleSelectPoolActionSchema(),
			"ssl_mode_selection":           getPolicyLbRuleSslModeSelectionSchema(),
			"variable_assignment":          getPolicyLbRuleVariableAssignmentActionSchema(),
			"variable_persistence_learn":   getPolicyLbRuleVariablePersistenceLearnActionSchema(),
			"variable_persistence_on":      getPolicyLbRuleVariablePersistenceLearnActionSchema(),
		},
	}
}

func getPolicyLbRuleConditionsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"http_request_body":          getPolicyLbRuleHTTPRequestBodyConditionSchema(),
			"http_request_cookie":        getPolicyLbRuleNameValueConditionSchema("cookie", "Rule condition based on HTTP cookie"),
			"http_request_header":        getPolicyLbRuleNameValueConditionSchema("header", "Rule condition based on HTTP request header"),
			"http_request_method":        getPolicyLbRuleHTTPRequestMethodConditionSchema(),
			"http_request_uri_arguments": getPolicyLbRuleHTTPRequestURIArgumentsConditionSchema(),
			"http_request_uri":           getPolicyLbRuleHTTPRequestURIConditionSchema(),
			"http_request_version":       getPolicyLbRuleHTTPVersionConditionSchema(),
			"http_response_header":       getPolicyLbRuleNameValueConditionSchema("header", "Rule condition based on HTTP response header"),
			"http_ssl":                   getPolicyLbRuleHTTPSslConditionSchema(),
			"ip_header":                  getPolicyLbRuleIPConditionSchema(),
			"ssl_sni":                    getPolicyLbRuleSslSniConditionSchema(),
			"tcp_header":                 getPolicyLbRuleTCPConditionSchema(),
			"variable":                   getPolicyLbRuleNameValueConditionSchema("variable", "Rule condition based on IP header"),
		},
	}
}

func getPolicyLbRuleTCPConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on TCP settings of the message",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"source_port": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateSinglePort(),
				},
			},
		},
	}
}

func getPolicyLbRuleHTTPVersionConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on http request version",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"version": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"HTTP_VERSION_1_0", "HTTP_VERSION_1_1"}, false),
				},
			},
		},
	}
}

func getPolicyLbRuleHTTPRequestURIConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on http request URI",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"uri": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getPolicyLbRuleHTTPRequestURIArgumentsConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on http request URI arguments (query string)",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"uri_arguments": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getPolicyLbRuleHTTPRequestMethodConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on http request method",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"method": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice([]string{"GET", "OPTIONS", "POST", "HEAD", "PUT"}, false),
				},
			},
		},
	}
}

func getPolicyLbRuleHTTPRequestBodyConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on HTTP request body",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"body_value": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getPolicyLbRuleNameValueConditionSchema(key string, desc string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: desc,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse":        getLbRuleInverseSchema(),
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
				key + "_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				key + "_value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleHTTPSslConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on HTTP SSL handshake and connection",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"client_certificate_issuer_dn": {
					Type: schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"issuer_dn": {
								Type:     schema.TypeString,
								Required: true,
							},
							"case_sensitive": getLbRuleCaseSensitiveSchema(),
							"match_type":     getLbRuleMatchTypeSchema(),
						},
					},
					Optional: true,
				},
				"client_certificate_subject_dn": {
					Type: schema.TypeSet,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"subject_dn": {
								Type:     schema.TypeString,
								Required: true,
							},
							"case_sensitive": getLbRuleCaseSensitiveSchema(),
							"match_type":     getLbRuleMatchTypeSchema(),
						},
					},
					Optional: true,
				},
				"client_supported_ssl_ciphers": {
					Type:        schema.TypeList,
					Description: "Supported SSL ciphers",
					Optional:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringInSlice(lbHttpSslConditionClientSupportedSslCiphersValues, false),
					},
				},
				"session_reused": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(lbHttpSslConditionSessionReusedValues, false),
				},
				"used_protocol": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(lbHttpSslConditionUsedProtocolValues, false),
				},
				"used_ssl_cipher": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(lbHttpSslConditionUsedSslCipherValues, false),
				},
			},
		},
	}
}

func getPolicyLbRuleIPConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on IP settings of the message",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"source_address": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateSingleIP(),
				},
				"group_path": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func getPolicyLbRuleSslSniConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Rule condition based on SSL SNI in client hello",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"sni": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getPolicyLbRuleConnectionDropActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to drop the connection.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"_dummy": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "dummy_value_to_indicate_presence_of_section",
				},
			},
		},
	}
}

func getPolicyLbRuleHttpRedirectActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to redirect HTTP request messages to a new URL.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"redirect_status": {
					Type:     schema.TypeString,
					Required: true,
				},
				"redirect_url": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleHttpRejectActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to reject HTTP request messages",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"reply_message": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"reply_status": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleHttpRequestUriRewriteActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to rewrite URIs in matched HTTP request messages.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uri": {
					Type:     schema.TypeString,
					Required: true,
				},
				"uri_arguments": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func getPolicyLbRuleHttpMessageHeaderDeleteActionSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: description,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"header_name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleHttpMessageHeaderRewriteActionSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: description,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"header_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"header_value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleJwtAuthActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action  to control access to backend server resources using JSON Web Token(JWT) authentication.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 1,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"certificate_path": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"public_key_content": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"symmetric_key": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"pass_jwt_to_pool": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"realm": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"tokens": {
					Type:        schema.TypeList,
					Description: "JWT tokens",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func getPolicyLbRuleSelectPoolActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to select a pool for matched HTTP request messages.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"pool_id": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleSslModeSelectionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to select SSL mode.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ssl_mode": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringInSlice(lbSslModeSelectionActionValues, false),
				},
			},
		},
	}
}

func getPolicyLbRuleVariableAssignmentActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to create a new variable and assign value to it.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"variable_name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"variable_value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyLbRuleVariablePersistenceLearnActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Action to create a new variable and assign value to it.",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"persistence_profile_path": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"variable_hash_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"variable_name": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getPolicyClientSSLBindingFromSchema(d *schema.ResourceData) *model.LBClientSslProfileBinding {
	bindings := d.Get("client_ssl").([]interface{})
	for _, binding := range bindings {
		// maximum count is one
		data := binding.(map[string]interface{})
		chainDepth := int64(data["certificate_chain_depth"].(int))
		clientAuth := data["client_auth"].(string)
		caList := interface2StringList(data["ca_paths"].([]interface{}))
		crlList := interface2StringList(data["crl_paths"].([]interface{}))
		certPath := data["default_certificate_path"].(string)
		profilePath := data["ssl_profile_path"].(string)
		profileBinding := model.LBClientSslProfileBinding{
			CertificateChainDepth:  &chainDepth,
			ClientAuth:             &clientAuth,
			ClientAuthCaPaths:      caList,
			ClientAuthCrlPaths:     crlList,
			DefaultCertificatePath: &certPath,
			SslProfilePath:         &profilePath,
		}

		return &profileBinding
	}

	return nil
}

func setPolicyClientSSLBindingInSchema(d *schema.ResourceData, binding *model.LBClientSslProfileBinding) {
	var bindingList []map[string]interface{}
	if binding != nil {
		elem := make(map[string]interface{})
		if binding.CertificateChainDepth != nil {
			elem["certificate_chain_depth"] = *binding.CertificateChainDepth
		}
		if binding.ClientAuth != nil {
			elem["client_auth"] = *binding.ClientAuth
		}
		elem["ca_paths"] = binding.ClientAuthCaPaths
		elem["crl_paths"] = binding.ClientAuthCrlPaths
		elem["default_certificate_path"] = binding.DefaultCertificatePath
		if binding.SslProfilePath != nil {
			elem["ssl_profile_path"] = *binding.SslProfilePath
		}

		bindingList = append(bindingList, elem)
	}

	err := d.Set("client_ssl", bindingList)
	if err != nil {
		log.Printf("[WARNING] Failed to set client ssl in schema: %v", err)
	}
}

func getPolicyServerSSLBindingFromSchema(d *schema.ResourceData) *model.LBServerSslProfileBinding {
	bindings := d.Get("server_ssl").([]interface{})
	for _, binding := range bindings {
		// maximum count is one
		data := binding.(map[string]interface{})
		chainDepth := int64(data["certificate_chain_depth"].(int))
		serverAuth := data["server_auth"].(string)
		caList := interface2StringList(data["ca_paths"].([]interface{}))
		crlList := interface2StringList(data["crl_paths"].([]interface{}))
		certPath := data["client_certificate_path"].(string)
		profilePath := data["ssl_profile_path"].(string)
		profileBinding := model.LBServerSslProfileBinding{
			CertificateChainDepth: &chainDepth,
			ServerAuth:            &serverAuth,
			ServerAuthCaPaths:     caList,
			ServerAuthCrlPaths:    crlList,
			ClientCertificatePath: &certPath,
			SslProfilePath:        &profilePath,
		}

		return &profileBinding
	}

	return nil
}

func setPolicyServerSSLBindingInSchema(d *schema.ResourceData, binding *model.LBServerSslProfileBinding) {
	var bindingList []map[string]interface{}
	if binding != nil {
		elem := make(map[string]interface{})
		if binding.CertificateChainDepth != nil {
			elem["certificate_chain_depth"] = *binding.CertificateChainDepth
		}
		if binding.ServerAuth != nil {
			elem["server_auth"] = *binding.ServerAuth
		}
		elem["ca_paths"] = binding.ServerAuthCaPaths
		elem["crl_paths"] = binding.ServerAuthCrlPaths
		if binding.ClientCertificatePath != nil {
			elem["client_certificate_path"] = *binding.ClientCertificatePath
		}
		if binding.SslProfilePath != nil {
			elem["ssl_profile_path"] = *binding.SslProfilePath
		}

		bindingList = append(bindingList, elem)
	}

	err := d.Set("server_ssl", bindingList)
	if err != nil {
		log.Printf("[WARNING] Failed to set server ssl in schema: %v", err)
	}
}

func getPolicyAccessListControlFromSchema(d *schema.ResourceData) *model.LBAccessListControl {
	controls := d.Get("access_list_control").([]interface{})
	for _, control := range controls {
		// maximum count is one
		data := control.(map[string]interface{})
		action := data["action"].(string)
		enabled := data["enabled"].(bool)
		groupPath := data["group_path"].(string)
		result := model.LBAccessListControl{
			Action:    &action,
			Enabled:   &enabled,
			GroupPath: &groupPath,
		}

		return &result
	}

	return nil
}

func setPolicyAccessListControlInSchema(d *schema.ResourceData, control *model.LBAccessListControl) {
	var controlList []map[string]interface{}
	if control != nil {
		elem := make(map[string]interface{})
		elem["action"] = control.Action
		elem["enabled"] = control.Enabled
		elem["group_path"] = control.GroupPath

		controlList = append(controlList, elem)
	}

	err := d.Set("access_list_control", controlList)
	if err != nil {
		log.Printf("[WARNING] Failed to set access list control in schema: %v", err)
	}
}

func setPolicyLbRulesInSchema(d *schema.ResourceData, rules []model.LBRule) {
	converter := bindings.NewTypeConverter()
	converter.SetMode(bindings.REST)

	var ruleList []interface{}
	for _, rule := range rules {
		ruleElem := make(map[string]interface{})
		if rule.DisplayName != nil {
			ruleElem["display_name"] = *rule.DisplayName
		}
		if rule.MatchStrategy != nil {
			ruleElem["match_strategy"] = *rule.MatchStrategy
		}
		if rule.Phase != nil {
			ruleElem["phase"] = *rule.Phase
		}

		// Actions
		var connectionDropActionList []interface{}
		var selectPoolActionList []interface{}
		var httpRedirectActionList []interface{}
		var httpRequestUriRewriteActionList []interface{}
		var httpRequestHeaderRewriteActionList []interface{}
		var httpRejectActionList []interface{}
		var httpResponseHeaderRewriteActionList []interface{}
		var httpRequestHeaderDeleteActionList []interface{}
		var httpResponseHeaderDeleteActionList []interface{}
		var variableAssignmentActionList []interface{}
		var variablePersistenceOnActionList []interface{}
		var variablePersistenceLearnActionList []interface{}
		var jwtAuthActionList []interface{}
		var sslModeSelectionActionList []interface{}

		for _, action := range rule.Actions {
			actionElem := make(map[string]interface{})

			basicType, _ := converter.ConvertToGolang(action, model.LBRuleActionBindingType())
			actionType := basicType.(model.LBRuleAction).Type_

			if actionType == model.LBRuleAction_TYPE_LBCONNECTIONDROPACTION {
				actionElem["_dummy"] = "dummy_value_to_indicate_presence_of_section"
				connectionDropActionList = append(connectionDropActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBSELECTPOOLACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBSelectPoolActionBindingType())
				actionElem["pool_id"] = specificType.(model.LBSelectPoolAction).PoolId
				selectPoolActionList = append(selectPoolActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPREDIRECTACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpRedirectActionBindingType())
				actionElem["redirect_status"] = specificType.(model.LBHttpRedirectAction).RedirectStatus
				actionElem["redirect_url"] = specificType.(model.LBHttpRedirectAction).RedirectUrl
				httpRedirectActionList = append(httpRedirectActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPREQUESTURIREWRITEACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpRequestUriRewriteActionBindingType())
				actionElem["uri"] = specificType.(model.LBHttpRequestUriRewriteAction).Uri
				actionElem["uri_arguments"] = specificType.(model.LBHttpRequestUriRewriteAction).UriArguments
				httpRequestUriRewriteActionList = append(httpRequestUriRewriteActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPREQUESTHEADERREWRITEACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpRequestHeaderRewriteActionBindingType())
				actionElem["header_name"] = specificType.(model.LBHttpRequestHeaderRewriteAction).HeaderName
				actionElem["header_value"] = specificType.(model.LBHttpRequestHeaderRewriteAction).HeaderValue
				httpRequestHeaderRewriteActionList = append(httpRequestHeaderRewriteActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPREJECTACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpRejectActionBindingType())
				actionElem["reply_message"] = specificType.(model.LBHttpRejectAction).ReplyMessage
				actionElem["reply_status"] = specificType.(model.LBHttpRejectAction).ReplyStatus
				httpRejectActionList = append(httpRejectActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPRESPONSEHEADERREWRITEACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpResponseHeaderRewriteActionBindingType())
				actionElem["header_name"] = specificType.(model.LBHttpResponseHeaderRewriteAction).HeaderName
				actionElem["header_value"] = specificType.(model.LBHttpResponseHeaderRewriteAction).HeaderValue
				httpResponseHeaderRewriteActionList = append(httpResponseHeaderRewriteActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPREQUESTHEADERDELETEACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpRequestHeaderDeleteActionBindingType())
				actionElem["header_name"] = specificType.(model.LBHttpRequestHeaderDeleteAction).HeaderName
				httpRequestHeaderDeleteActionList = append(httpRequestHeaderDeleteActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBHTTPRESPONSEHEADERDELETEACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBHttpResponseHeaderDeleteActionBindingType())
				actionElem["header_name"] = specificType.(model.LBHttpResponseHeaderDeleteAction).HeaderName
				httpResponseHeaderDeleteActionList = append(httpResponseHeaderDeleteActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBVARIABLEASSIGNMENTACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBVariableAssignmentActionBindingType())
				actionElem["variable_name"] = specificType.(model.LBVariableAssignmentAction).VariableName
				actionElem["variable_value"] = specificType.(model.LBVariableAssignmentAction).VariableValue
				variableAssignmentActionList = append(variableAssignmentActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBVARIABLEPERSISTENCEONACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBVariablePersistenceOnActionBindingType())
				actionElem["persistence_profile_path"] = specificType.(model.LBVariablePersistenceOnAction).PersistenceProfilePath
				actionElem["variable_hash_enabled"] = specificType.(model.LBVariablePersistenceOnAction).VariableHashEnabled
				actionElem["variable_name"] = specificType.(model.LBVariablePersistenceOnAction).VariableName
				variablePersistenceOnActionList = append(variablePersistenceOnActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBVARIABLEPERSISTENCELEARNACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBVariablePersistenceLearnActionBindingType())
				actionElem["persistence_profile_path"] = specificType.(model.LBVariablePersistenceLearnAction).PersistenceProfilePath
				actionElem["variable_hash_enabled"] = specificType.(model.LBVariablePersistenceLearnAction).VariableHashEnabled
				actionElem["variable_name"] = specificType.(model.LBVariablePersistenceLearnAction).VariableName
				variablePersistenceLearnActionList = append(variablePersistenceLearnActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBSSLMODESELECTIONACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBSslModeSelectionActionBindingType())
				actionElem["ssl_mode"] = specificType.(model.LBSslModeSelectionAction).SslMode
				sslModeSelectionActionList = append(sslModeSelectionActionList, actionElem)
			} else if actionType == model.LBRuleAction_TYPE_LBJWTAUTHACTION {
				specificType, _ := converter.ConvertToGolang(action, model.LBJwtAuthActionBindingType())
				actionElem["pass_jwt_to_pool"] = specificType.(model.LBJwtAuthAction).PassJwtToPool
				actionElem["realm"] = specificType.(model.LBJwtAuthAction).Realm

				var keyList []interface{}
				keyElem := make(map[string]interface{})

				key := specificType.(model.LBJwtAuthAction).Key
				basicKeyType, _ := converter.ConvertToGolang(key, model.LBJwtKeyBindingType())
				keyType := basicKeyType.(model.LBJwtKey).Type_

				if keyType == model.LBJwtKey_TYPE_LBJWTCERTIFICATEKEY {
					specificKeyType, _ := converter.ConvertToGolang(key, model.LBJwtCertificateKeyBindingType())
					keyElem["certificate_path"] = specificKeyType.(model.LBJwtCertificateKey).CertificatePath
				} else if keyType == model.LBJwtKey_TYPE_LBJWTSYMMETRICKEY {
					keyElem["symmetric_key"] = "dummy"
				} else if keyType == model.LBJwtKey_TYPE_LBJWTPUBLICKEY {
					specificKeyType, _ := converter.ConvertToGolang(key, model.LBJwtPublicKeyBindingType())
					keyElem["public_key_content"] = specificKeyType.(model.LBJwtPublicKey).PublicKeyContent
				}

				keyList = append(keyList, keyElem)
				actionElem["key"] = schema.NewSet(resourceKeyValueHash, keyList)

				var tokens []interface{}
				for _, v := range specificType.(model.LBJwtAuthAction).Tokens {
					tokens = append(tokens, v)
				}
				actionElem["tokens"] = tokens

				jwtAuthActionList = append(jwtAuthActionList, actionElem)
			}
		}

		actionElem := make(map[string]interface{})
		actionElem["connection_drop"] = connectionDropActionList
		actionElem["select_pool"] = selectPoolActionList
		actionElem["http_redirect"] = httpRedirectActionList
		actionElem["http_request_uri_rewrite"] = httpRequestUriRewriteActionList
		actionElem["http_request_header_rewrite"] = httpRequestHeaderRewriteActionList
		actionElem["http_reject"] = httpRejectActionList
		actionElem["http_response_header_rewrite"] = httpResponseHeaderRewriteActionList
		actionElem["http_request_header_delete"] = httpRequestHeaderDeleteActionList
		actionElem["http_response_header_delete"] = httpResponseHeaderDeleteActionList
		actionElem["variable_assignment"] = variableAssignmentActionList
		actionElem["variable_persistence_on"] = variablePersistenceOnActionList
		actionElem["variable_persistence_learn"] = variablePersistenceLearnActionList
		actionElem["jwt_auth"] = jwtAuthActionList
		actionElem["ssl_mode_selection"] = sslModeSelectionActionList

		var actionList []interface{}
		actionList = append(actionList, actionElem)
		ruleElem["action"] = actionList

		// MatchConditions
		var httpRequestBodyConditionList []interface{}
		var httpRequestUriConditionList []interface{}
		var httpRequestHeaderConditionList []interface{}
		var httpRequestMethodConditionList []interface{}
		var httpRequestUriArgumentsConditionList []interface{}
		var httpRequestVersionConditionList []interface{}
		var httpRequestCookieConditionList []interface{}
		var httpResponseHeaderConditionList []interface{}
		var tcpHeaderConditionList []interface{}
		var ipHeaderConditionList []interface{}
		var variableConditionList []interface{}
		var httpSslConditionList []interface{}
		var sslSniConditionList []interface{}

		var conditionCount int = 0
		for _, condition := range rule.MatchConditions {
			conditionCount = conditionCount + 1
			conditionElem := make(map[string]interface{})

			basicType, _ := converter.ConvertToGolang(condition, model.LBRuleConditionBindingType())
			conditionType := basicType.(model.LBRuleCondition).Type_

			if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTBODYCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestBodyConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpRequestBodyCondition).CaseSensitive
				conditionElem["inverse"] = specificType.(model.LBHttpRequestBodyCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpRequestBodyCondition).MatchType
				conditionElem["body_value"] = specificType.(model.LBHttpRequestBodyCondition).BodyValue
				httpRequestBodyConditionList = append(httpRequestBodyConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTURICONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestUriConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpRequestUriCondition).CaseSensitive
				conditionElem["inverse"] = specificType.(model.LBHttpRequestUriCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpRequestUriCondition).MatchType
				conditionElem["uri"] = specificType.(model.LBHttpRequestUriCondition).Uri
				httpRequestUriConditionList = append(httpRequestUriConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTHEADERCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestHeaderConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpRequestHeaderCondition).CaseSensitive
				conditionElem["header_name"] = specificType.(model.LBHttpRequestHeaderCondition).HeaderName
				conditionElem["header_value"] = specificType.(model.LBHttpRequestHeaderCondition).HeaderValue
				conditionElem["inverse"] = specificType.(model.LBHttpRequestHeaderCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpRequestHeaderCondition).MatchType
				httpRequestHeaderConditionList = append(httpRequestHeaderConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTMETHODCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestMethodConditionBindingType())
				conditionElem["inverse"] = specificType.(model.LBHttpRequestMethodCondition).Inverse
				conditionElem["method"] = specificType.(model.LBHttpRequestMethodCondition).Method
				httpRequestMethodConditionList = append(httpRequestMethodConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTURIARGUMENTSCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestUriArgumentsConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpRequestUriArgumentsCondition).CaseSensitive
				conditionElem["inverse"] = specificType.(model.LBHttpRequestUriArgumentsCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpRequestUriArgumentsCondition).MatchType
				conditionElem["uri_arguments"] = specificType.(model.LBHttpRequestUriArgumentsCondition).UriArguments
				httpRequestUriArgumentsConditionList = append(httpRequestUriArgumentsConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTVERSIONCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestVersionConditionBindingType())
				conditionElem["inverse"] = specificType.(model.LBHttpRequestVersionCondition).Inverse
				conditionElem["version"] = specificType.(model.LBHttpRequestVersionCondition).Version
				httpRequestVersionConditionList = append(httpRequestVersionConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPREQUESTCOOKIECONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpRequestCookieConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpRequestCookieCondition).CaseSensitive
				conditionElem["cookie_name"] = specificType.(model.LBHttpRequestCookieCondition).CookieName
				conditionElem["cookie_value"] = specificType.(model.LBHttpRequestCookieCondition).CookieValue
				conditionElem["inverse"] = specificType.(model.LBHttpRequestCookieCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpRequestCookieCondition).MatchType
				httpRequestCookieConditionList = append(httpRequestCookieConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPRESPONSEHEADERCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpResponseHeaderConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBHttpResponseHeaderCondition).CaseSensitive
				conditionElem["header_name"] = specificType.(model.LBHttpResponseHeaderCondition).HeaderName
				conditionElem["header_value"] = specificType.(model.LBHttpResponseHeaderCondition).HeaderValue
				conditionElem["inverse"] = specificType.(model.LBHttpResponseHeaderCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBHttpResponseHeaderCondition).MatchType
				httpResponseHeaderConditionList = append(httpResponseHeaderConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBTCPHEADERCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBTcpHeaderConditionBindingType())
				conditionElem["inverse"] = specificType.(model.LBTcpHeaderCondition).Inverse
				conditionElem["source_port"] = specificType.(model.LBTcpHeaderCondition).SourcePort
				tcpHeaderConditionList = append(tcpHeaderConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBIPHEADERCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBIpHeaderConditionBindingType())
				conditionElem["group_path"] = specificType.(model.LBIpHeaderCondition).GroupPath
				conditionElem["inverse"] = specificType.(model.LBIpHeaderCondition).Inverse
				conditionElem["source_address"] = specificType.(model.LBIpHeaderCondition).SourceAddress
				ipHeaderConditionList = append(ipHeaderConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBVARIABLECONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBVariableConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBVariableCondition).CaseSensitive
				conditionElem["inverse"] = specificType.(model.LBVariableCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBVariableCondition).MatchType
				conditionElem["variable_name"] = specificType.(model.LBVariableCondition).VariableName
				conditionElem["variable_value"] = specificType.(model.LBVariableCondition).VariableValue
				variableConditionList = append(variableConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBSSLSNICONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBSslSniConditionBindingType())
				conditionElem["case_sensitive"] = specificType.(model.LBSslSniCondition).CaseSensitive
				conditionElem["inverse"] = specificType.(model.LBSslSniCondition).Inverse
				conditionElem["match_type"] = specificType.(model.LBSslSniCondition).MatchType
				conditionElem["sni"] = specificType.(model.LBSslSniCondition).Sni
				sslSniConditionList = append(sslSniConditionList, conditionElem)
			} else if conditionType == model.LBRuleCondition_TYPE_LBHTTPSSLCONDITION {
				specificType, _ := converter.ConvertToGolang(condition, model.LBHttpSslConditionBindingType())
				conditionElem["inverse"] = specificType.(model.LBHttpSslCondition).Inverse
				conditionElem["session_reused"] = specificType.(model.LBHttpSslCondition).SessionReused
				conditionElem["used_protocol"] = specificType.(model.LBHttpSslCondition).UsedProtocol
				conditionElem["used_ssl_cipher"] = specificType.(model.LBHttpSslCondition).UsedSslCipher

				issuer_dn := specificType.(model.LBHttpSslCondition).ClientCertificateIssuerDn
				var issuerDnList []interface{}
				issuerElem := make(map[string]interface{})
				issuerElem["case_sensitive"] = issuer_dn.CaseSensitive
				issuerElem["issuer_dn"] = issuer_dn.IssuerDn
				issuerElem["match_type"] = issuer_dn.MatchType
				issuerDnList = append(issuerDnList, issuerElem)
				conditionElem["client_certificate_issuer_dn"] = schema.NewSet(resourceKeyValueHash, issuerDnList)

				subject_dn := specificType.(model.LBHttpSslCondition).ClientCertificateSubjectDn
				var subjectDnList []interface{}
				subjectElem := make(map[string]interface{})
				subjectElem["case_sensitive"] = subject_dn.CaseSensitive
				subjectElem["subject_dn"] = subject_dn.SubjectDn
				subjectElem["match_type"] = subject_dn.MatchType
				subjectDnList = append(subjectDnList, subjectElem)
				conditionElem["client_certificate_subject_dn"] = schema.NewSet(resourceKeyValueHash, subjectDnList)

				var ssl_ciphers []interface{}
				for _, v := range specificType.(model.LBHttpSslCondition).ClientSupportedSslCiphers {
					ssl_ciphers = append(ssl_ciphers, v)
				}
				conditionElem["client_supported_ssl_ciphers"] = ssl_ciphers

				httpSslConditionList = append(httpSslConditionList, conditionElem)
			}
		}

		// Optional argument, only set it if we get anything back
		if conditionCount > 0 {
			conditionElem := make((map[string]interface{}))
			conditionElem["http_request_body"] = httpRequestBodyConditionList
			conditionElem["http_request_uri"] = httpRequestUriConditionList
			conditionElem["http_request_header"] = httpRequestHeaderConditionList
			conditionElem["http_request_method"] = httpRequestMethodConditionList
			conditionElem["http_request_uri_arguments"] = httpRequestUriArgumentsConditionList
			conditionElem["http_request_version"] = httpRequestVersionConditionList
			conditionElem["http_request_cookie"] = httpRequestCookieConditionList
			conditionElem["http_response_header"] = httpResponseHeaderConditionList
			conditionElem["tcp_header"] = tcpHeaderConditionList
			conditionElem["ip_header"] = ipHeaderConditionList
			conditionElem["variable"] = variableConditionList
			conditionElem["http_ssl"] = httpSslConditionList
			conditionElem["ssl_sni"] = sslSniConditionList

			var conditionList []interface{}
			conditionList = append(conditionList, conditionElem)
			ruleElem["condition"] = conditionList
		}

		ruleList = append(ruleList, ruleElem)
	}

	err := d.Set("rule", ruleList)
	if err != nil {
		log.Printf("[WARNING] Failed to set rule list in schema: %v", err)
	}

}

func getRuleActionOrMethod(ruleData map[string]interface{}, key string, string_fields []string, bool_fields []string, internal_type string) []*data.StructValue {
	var result []*data.StructValue
	for _, object := range ruleData[key].([]interface{}) {
		object_data := object.(map[string]interface{})
		var fields = make(map[string]data.DataValue)
		fields["type"] = data.NewStringValue(internal_type)
		for _, field := range string_fields {
			if object_data[field] != nil && object_data[field].(string) != "" {
				fields[field] = data.NewStringValue(object_data[field].(string))
			}
		}
		for _, field := range bool_fields {
			if object_data[field] != nil {
				fields[field] = data.NewBooleanValue(object_data[field].(bool))
			}
		}
		elem := data.NewStructValue("", fields)
		result = append(result, elem)
	}
	return result
}

func getPolicyLbRulesFromSchema(d *schema.ResourceData) []model.LBRule {
	var ruleList []model.LBRule
	rules := d.Get("rule").([]interface{})

	for _, rule := range rules {
		rule_data := rule.(map[string]interface{})

		displayName := rule_data["display_name"].(string)
		matchStrategy := rule_data["match_strategy"].(string)
		phase := rule_data["phase"].(string)

		var actions []*data.StructValue

		rule_actions := rule_data["action"]
		for _, rule_action := range rule_actions.([]interface{}) {

			rule_action := rule_action.(map[string]interface{})

			// Just strings and booleans, we use a helper function for there
			actions = append(actions, getRuleActionOrMethod(rule_action, "connection_drop", []string{}, []string{}, model.LBRuleAction_TYPE_LBCONNECTIONDROPACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_redirect", []string{"redirect_status", "redirect_url"}, []string{}, model.LBRuleAction_TYPE_LBHTTPREDIRECTACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_reject", []string{"reply_message", "reply_status"}, []string{}, model.LBRuleAction_TYPE_LBHTTPREJECTACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_request_header_delete", []string{"header_name"}, []string{}, model.LBRuleAction_TYPE_LBHTTPREQUESTHEADERDELETEACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_request_header_rewrite", []string{"header_name", "header_value"}, []string{}, model.LBRuleAction_TYPE_LBHTTPREQUESTHEADERREWRITEACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_request_uri_rewrite", []string{"uri", "uri_arguments"}, []string{}, model.LBRuleAction_TYPE_LBHTTPREQUESTURIREWRITEACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_response_header_delete", []string{"header_name"}, []string{}, model.LBRuleAction_TYPE_LBHTTPRESPONSEHEADERDELETEACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "http_response_header_rewrite", []string{"header_name", "header_value"}, []string{}, model.LBRuleAction_TYPE_LBHTTPRESPONSEHEADERREWRITEACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "select_pool", []string{"pool_id"}, []string{}, model.LBRuleAction_TYPE_LBSELECTPOOLACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "ssl_mode_selection", []string{"ssl_mode"}, []string{}, model.LBRuleAction_TYPE_LBSSLMODESELECTIONACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "variable_assignment", []string{"variable_name", "variable_value"}, []string{}, model.LBRuleAction_TYPE_LBVARIABLEASSIGNMENTACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "variable_persistence_learn", []string{"persistence_profile_path", "variable_name"}, []string{"variable_hash_enabled"}, model.LBRuleAction_TYPE_LBVARIABLEPERSISTENCELEARNACTION)...)
			actions = append(actions, getRuleActionOrMethod(rule_action, "variable_persistence_on", []string{"persistence_profile_path", "variable_name"}, []string{"variable_hash_enabled"}, model.LBRuleAction_TYPE_LBVARIABLEPERSISTENCEONACTION)...)

			// more complicated actions
			for _, action := range rule_action["jwt_auth"].([]interface{}) {
				action_data := action.(map[string]interface{})
				var fields = make(map[string]data.DataValue)
				fields["type"] = data.NewStringValue(model.LBRuleAction_TYPE_LBJWTAUTHACTION)
				if action_data["realm"] != nil {
					fields["realm"] = data.NewStringValue(action_data["realm"].(string))
				}
				if action_data["pass_jwt_to_pool"] != nil {
					fields["pass_jwt_to_pool"] = data.NewBooleanValue(action_data["pass_jwt_to_pool"].(bool))
				}
				// I still haven't fully figured out why key comes as a (*schema.Set) but that's what we want,
				// a set where no key can be there more than once
				for _, key := range action_data["key"].(*schema.Set).List() {
					key_data := key.(map[string]interface{})
					var key_fields = make(map[string]data.DataValue)
					if key_data["certificate_path"] != nil && key_data["certificate_path"].(string) != "" {
						key_fields["certificate_path"] = data.NewStringValue(key_data["certificate_path"].(string))
						key_fields["type"] = data.NewStringValue(model.LBJwtKey_TYPE_LBJWTCERTIFICATEKEY)
					} else if key_data["public_key_content"] != nil && key_data["public_key_content"].(string) != "" {
						key_fields["public_key_content"] = data.NewStringValue(key_data["public_key_content"].(string))
						key_fields["type"] = data.NewStringValue(model.LBJwtKey_TYPE_LBJWTPUBLICKEY)
					} else if key_data["symmetric_key"] != nil && key_data["symmetric_key"].(string) != "" {
						// the API only wants the marker id, no actual content parameters
						key_fields["type"] = data.NewStringValue(model.LBJwtKey_TYPE_LBJWTSYMMETRICKEY)
					}
					fields["key"] = data.NewStructValue("", key_fields)
				}
				if action_data["tokens"] != nil {
					token_list := data.NewListValue()
					for _, token := range action_data["tokens"].([]interface{}) {
						token_list.Add(data.NewStringValue(token.(string)))
					}
					fields["tokens"] = token_list
				}
				elem := data.NewStructValue("", fields)
				actions = append(actions, elem)
			}
		}

		var matchConditions []*data.StructValue

		rule_conditions := rule_data["condition"]
		for _, rule_condition := range rule_conditions.([]interface{}) {

			rule_condition := rule_condition.(map[string]interface{})

			// Just strings and booleans, we use a helper function for there
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_body", []string{"body_value", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTBODYCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_cookie", []string{"cookie_name", "cookie_value", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTCOOKIECONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_header", []string{"header_name", "header_value", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTHEADERCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_method", []string{"method"}, []string{}, model.LBRuleCondition_TYPE_LBHTTPREQUESTMETHODCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_uri_arguments", []string{"uri_arguments", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTURIARGUMENTSCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_uri", []string{"uri", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTURICONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_request_version", []string{"version"}, []string{"inverse"}, model.LBRuleCondition_TYPE_LBHTTPREQUESTVERSIONCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "http_response_header", []string{"header_name", "header_value", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBHTTPRESPONSEHEADERCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "ip_header", []string{"group_path", "source_address"}, []string{"inverse"}, model.LBRuleCondition_TYPE_LBIPHEADERCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "ssl_sni", []string{"sni", "match_type"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBSSLSNICONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "tcp_header", []string{"source_port"}, []string{"inverse"}, model.LBRuleCondition_TYPE_LBTCPHEADERCONDITION)...)
			matchConditions = append(matchConditions, getRuleActionOrMethod(rule_condition, "variable", []string{"match_type", "variable_name", "variable_value"}, []string{"case_sensitive", "inverse"}, model.LBRuleCondition_TYPE_LBVARIABLECONDITION)...)

			// more complicated conditions
			for _, condition := range rule_condition["http_ssl"].([]interface{}) {
				condition_data := condition.(map[string]interface{})
				var fields = make(map[string]data.DataValue)
				fields["type"] = data.NewStringValue(model.LBRuleCondition_TYPE_LBHTTPSSLCONDITION)
				// Not actually a list but that's what the Schems says
				for _, issuer_dn := range condition_data["client_certificate_issuer_dn"].(*schema.Set).List() {
					issuer_dn_data := issuer_dn.(map[string]interface{})
					var issuer_dn_fields = make(map[string]data.DataValue)
					if issuer_dn_data["case_sensitive"] != nil {
						issuer_dn_fields["case_sensitive"] = data.NewBooleanValue(issuer_dn_data["case_sensitive"].(bool))
					}
					if issuer_dn_data["issuer_dn"] != nil {
						issuer_dn_fields["issuer_dn"] = data.NewStringValue(issuer_dn_data["issuer_dn"].(string))
					}
					if issuer_dn_data["match_type"] != nil {
						issuer_dn_fields["match_type"] = data.NewStringValue(issuer_dn_data["match_type"].(string))
					}
					fields["client_certificate_issuer_dn"] = data.NewStructValue("", issuer_dn_fields)
				}
				// Not actually a list but that's what the Schems says
				for _, subject_dn := range condition_data["client_certificate_subject_dn"].(*schema.Set).List() {
					subject_dn_data := subject_dn.(map[string]interface{})
					var subject_dn_fields = make(map[string]data.DataValue)
					if subject_dn_data["case_sensitive"] != nil {
						subject_dn_fields["case_sensitive"] = data.NewBooleanValue(subject_dn_data["case_sensitive"].(bool))
					}
					if subject_dn_data["subject_dn"] != nil {
						subject_dn_fields["subject_dn"] = data.NewStringValue(subject_dn_data["subject_dn"].(string))
					}
					if subject_dn_data["match_type"] != nil {
						subject_dn_fields["match_type"] = data.NewStringValue(subject_dn_data["match_type"].(string))
					}
					fields["client_certificate_subject_dn"] = data.NewStructValue("", subject_dn_fields)
				}
				if condition_data["client_supported_ssl_ciphers"] != nil {
					cipher_list := data.NewListValue()
					for _, cipher := range condition_data["client_supported_ssl_ciphers"].([]interface{}) {
						cipher_list.Add(data.NewStringValue(cipher.(string)))
					}
					fields["client_supported_ssl_ciphers"] = cipher_list
				}
				if condition_data["inverse"] != nil {
					fields["inverse"] = data.NewBooleanValue(condition_data["inverse"].(bool))
				}
				if condition_data["session_reused"] != nil {
					fields["session_reused"] = data.NewStringValue(condition_data["session_reused"].(string))
				}
				if condition_data["used_protocol"] != nil {
					fields["used_protocol"] = data.NewStringValue(condition_data["used_protocol"].(string))
				}
				if condition_data["used_ssl_cipher"] != nil {
					fields["used_ssl_cipher"] = data.NewStringValue(condition_data["used_ssl_cipher"].(string))
				}
				elem := data.NewStructValue("", fields)
				matchConditions = append(matchConditions, elem)
			}
		}

		elem := model.LBRule{
			DisplayName:     &displayName,
			MatchStrategy:   &matchStrategy,
			Phase:           &phase,
			Actions:         actions,
			MatchConditions: matchConditions,
		}
		ruleList = append(ruleList, elem)
	}
	return ruleList
}

func policyLBVirtualServerVersionDependantSet(d *schema.ResourceData, obj *model.LBVirtualServer) {
	if nsxVersionHigherOrEqual("3.0.0") {
		logSignificantOnly := d.Get("log_significant_event_only").(bool)
		obj.LogSignificantEventOnly = &logSignificantOnly
		obj.AccessListControl = getPolicyAccessListControlFromSchema(d)
	}
}

func resourceNsxtPolicyLBVirtualServerExists(id string, connector *client.RestConnector, isGlobalManager bool) (bool, error) {
	client := infra.NewDefaultLbVirtualServersClient(connector)

	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyLBVirtualServerCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDefaultLbVirtualServersClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBVirtualServerExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	accessLogEnabled := d.Get("access_log_enabled").(bool)
	applicationProfilePath := d.Get("application_profile_path").(string)
	clientSSLProfileBinding := getPolicyClientSSLBindingFromSchema(d)
	defaultPoolMemberPorts := getStringListFromSchemaList(d, "default_pool_member_ports")
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	lbPersistenceProfilePath := d.Get("persistence_profile_path").(string)
	lbServicePath := d.Get("service_path").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	poolPath := d.Get("pool_path").(string)
	ports := getStringListFromSchemaList(d, "ports")
	serverSSLProfileBinding := getPolicyServerSSLBindingFromSchema(d)
	sorryPoolPath := d.Get("sorry_pool_path").(string)
	rules := getPolicyLbRulesFromSchema(d)

	obj := model.LBVirtualServer{
		DisplayName:              &displayName,
		Description:              &description,
		Tags:                     tags,
		AccessLogEnabled:         &accessLogEnabled,
		ApplicationProfilePath:   &applicationProfilePath,
		ClientSslProfileBinding:  clientSSLProfileBinding,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  &enabled,
		IpAddress:                &ipAddress,
		LbPersistenceProfilePath: &lbPersistenceProfilePath,
		LbServicePath:            &lbServicePath,
		PoolPath:                 &poolPath,
		Ports:                    ports,
		ServerSslProfileBinding:  serverSSLProfileBinding,
		SorryPoolPath:            &sorryPoolPath,
		Rules:                    rules,
	}

	policyLBVirtualServerVersionDependantSet(d, &obj)

	if maxNewConnectionRate > 0 {
		obj.MaxNewConnectionRate = &maxNewConnectionRate
	}

	if maxConcurrentConnections > 0 {
		obj.MaxConcurrentConnections = &maxConcurrentConnections
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating LBVirtualServer with ID %s", id)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("LBVirtualServer", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBVirtualServerRead(d, m)
}

func resourceNsxtPolicyLBVirtualServerRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDefaultLbVirtualServersClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBVirtualServer ID")
	}

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBVirtualServer", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("access_log_enabled", obj.AccessLogEnabled)
	d.Set("application_profile_path", obj.ApplicationProfilePath)
	setPolicyClientSSLBindingInSchema(d, obj.ClientSslProfileBinding)
	d.Set("default_pool_member_ports", obj.DefaultPoolMemberPorts)
	d.Set("enabled", obj.Enabled)
	d.Set("ip_address", obj.IpAddress)
	d.Set("persistence_profile_path", obj.LbPersistenceProfilePath)
	d.Set("service_path", obj.LbServicePath)
	d.Set("max_concurrent_connections", obj.MaxConcurrentConnections)
	d.Set("max_new_connection_rate", obj.MaxNewConnectionRate)
	d.Set("pool_path", obj.PoolPath)
	d.Set("ports", obj.Ports)
	setPolicyServerSSLBindingInSchema(d, obj.ServerSslProfileBinding)
	d.Set("sorry_pool_path", obj.SorryPoolPath)
	d.Set("log_significant_event_only", obj.LogSignificantEventOnly)
	setPolicyAccessListControlInSchema(d, obj.AccessListControl)
	setPolicyLbRulesInSchema(d, obj.Rules)

	return nil
}

func resourceNsxtPolicyLBVirtualServerUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDefaultLbVirtualServersClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBVirtualServer ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	accessLogEnabled := d.Get("access_log_enabled").(bool)
	clientSSLProfileBinding := getPolicyClientSSLBindingFromSchema(d)
	applicationProfilePath := d.Get("application_profile_path").(string)
	defaultPoolMemberPorts := getStringListFromSchemaList(d, "default_pool_member_ports")
	enabled := d.Get("enabled").(bool)
	ipAddress := d.Get("ip_address").(string)
	lbPersistenceProfilePath := d.Get("persistence_profile_path").(string)
	lbServicePath := d.Get("service_path").(string)
	maxConcurrentConnections := int64(d.Get("max_concurrent_connections").(int))
	maxNewConnectionRate := int64(d.Get("max_new_connection_rate").(int))
	poolPath := d.Get("pool_path").(string)
	ports := getStringListFromSchemaList(d, "ports")
	serverSSLProfileBinding := getPolicyServerSSLBindingFromSchema(d)
	sorryPoolPath := d.Get("sorry_pool_path").(string)
	rules := getPolicyLbRulesFromSchema(d)

	obj := model.LBVirtualServer{
		DisplayName:              &displayName,
		Description:              &description,
		Tags:                     tags,
		AccessLogEnabled:         &accessLogEnabled,
		ApplicationProfilePath:   &applicationProfilePath,
		ClientSslProfileBinding:  clientSSLProfileBinding,
		DefaultPoolMemberPorts:   defaultPoolMemberPorts,
		Enabled:                  &enabled,
		IpAddress:                &ipAddress,
		LbPersistenceProfilePath: &lbPersistenceProfilePath,
		LbServicePath:            &lbServicePath,
		PoolPath:                 &poolPath,
		Ports:                    ports,
		ServerSslProfileBinding:  serverSSLProfileBinding,
		SorryPoolPath:            &sorryPoolPath,
		Rules:                    rules,
	}

	policyLBVirtualServerVersionDependantSet(d, &obj)

	if maxNewConnectionRate > 0 {
		obj.MaxNewConnectionRate = &maxNewConnectionRate
	}

	if maxConcurrentConnections > 0 {
		obj.MaxConcurrentConnections = &maxConcurrentConnections
	}

	// Update the resource using PATCH
	err := client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("LBVirtualServer", id, err)
	}

	return resourceNsxtPolicyLBVirtualServerRead(d, m)
}

func resourceNsxtPolicyLBVirtualServerDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBVirtualServer ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewDefaultLbVirtualServersClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	force := true
	err := client.Delete(id, &force)
	if err != nil {
		return handleDeleteError("LBVirtualServer", id, err)
	}

	return nil
}
