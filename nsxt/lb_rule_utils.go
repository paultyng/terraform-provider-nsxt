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

func getLbRuleInverseSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Whether to reverse match result of this condition",
		Optional:    true,
		Default:     false,
	}
}

func getLbRuleCaseSensitiveSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "If true, case is significant in condition matching",
		Optional:    true,
		Default:     true,
	}
}

func getLbRuleMatchTypeSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Match type (STARTS_WITH, ENDS_WITH, EQUALS, CONTAINS, REGEX)",
		ValidateFunc: validation.StringInSlice([]string{"STARTS_WITH", "ENDS_WITH", "EQUALS", "CONTAINS", "REGEX"}, false),
		Required:     true,
	}
}

func getLbRuleHTTPRequestBodyConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Rule condition based on http request body",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getLbRuleHTTPHeaderConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Rule condition based on http header",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
				"case_sensitive": getLbRuleCaseSensitiveSchema(),
				"match_type":     getLbRuleMatchTypeSchema(),
			},
		},
	}
}

func getLbRuleHTTPRequestMethodConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
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

func getLbRuleHTTPVersionConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
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

func getLbRuleHTTPRequestURIConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
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

func getLbRuleHTTPRequestURIArgumentsConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
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

func getLbRuleIPConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Rule condition based on IP settings of the message",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"inverse": getLbRuleInverseSchema(),
				"source_address": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validateSingleIP(),
				},
			},
		},
	}
}

func getLbRuleTCPConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
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

func getLbRuleURIRewriteActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Uri to replace original URI in outgoing request",
		Optional:    true,
		MaxItems:    1,
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

func getLbRuleHeaderRewriteActionSchema(optional bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Header to replace original header in outgoing message",
		Optional:    optional,
		Required:    !optional,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func initLbHTTPRuleMatchConditionFromSchema(data map[string]interface{}, conditionType string, isHeader bool, isCookie bool) loadbalancer.LbRuleCondition {
	condition := loadbalancer.LbRuleCondition{
		Inverse:   data["inverse"].(bool),
		Type_:     conditionType,
		MatchType: data["match_type"].(string),
	}

	if isHeader {
		condition.HeaderName = data["name"].(string)
		condition.HeaderValue = data["value"].(string)
	}

	if isCookie {
		condition.CookieName = data["name"].(string)
		condition.CookieValue = data["value"].(string)
	}

	condition.CaseSensitive = new(bool)
	*condition.CaseSensitive = data["case_sensitive"].(bool)
	return condition
}

func fillLbHTTPRuleHeaderConditionInSchema(data map[string]interface{}, condition loadbalancer.LbRuleCondition, isCookie bool) {
	if isCookie {
		data["name"] = condition.CookieName
		data["value"] = condition.CookieValue
	} else {
		data["name"] = condition.HeaderName
		data["value"] = condition.HeaderValue
	}
	data["inverse"] = condition.Inverse
	data["match_type"] = condition.MatchType
	data["case_sensitive"] = *condition.CaseSensitive
}

func resourceNsxtLbHTTPRuleDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerRule(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LoadBalancerRule delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LoadBalancerRule %s not found", id)
		d.SetId("")
	}
	return nil
}
