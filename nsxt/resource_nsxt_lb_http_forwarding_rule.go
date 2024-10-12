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

func resourceNsxtLbHTTPForwardingRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbHTTPForwardingRuleCreate,
		Read:   resourceNsxtLbHTTPForwardingRuleRead,
		Update: resourceNsxtLbHTTPForwardingRuleUpdate,
		Delete: resourceNsxtLbHTTPRuleDelete,
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
			"match_strategy": {
				Type:         schema.TypeString,
				Description:  "Strategy when multiple match conditions are specified in one rule (ANY vs ALL)",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"ALL", "ANY"}, false),
				Default:      "ALL",
			},
			"header_condition":  getLbRuleHTTPHeaderConditionSchema(),
			"body_condition":    getLbRuleHTTPRequestBodyConditionSchema(),
			"method_condition":  getLbRuleHTTPRequestMethodConditionSchema(),
			"cookie_condition":  getLbRuleHTTPHeaderConditionSchema(),
			"version_condition": getLbRuleHTTPVersionConditionSchema(),
			"uri_condition":     getLbRuleHTTPRequestURIConditionSchema(),
			"ip_condition":      getLbRuleIPConditionSchema(),
			"tcp_condition":     getLbRuleTCPConditionSchema(),

			"http_reject_action": {
				Type:        schema.TypeSet,
				Description: "Reject the request with a defined status and message",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"reply_status": {
							Type:     schema.TypeString,
							Required: true,
						},
						"reply_message": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"http_redirect_action": {
				Type:        schema.TypeSet,
				Description: "Redirect the request with a defined status and url",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"redirect_status": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"301", "302", "303", "307"}, false),
							Required:     true,
						},
						"redirect_url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"select_pool_action": {
				Type:        schema.TypeSet,
				Description: "Forward the request to the a defined pool",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pool_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func getLbRuleHTTPForwardingConditionsFromSchema(d *schema.ResourceData) []loadbalancer.LbRuleCondition {
	var conditionList []loadbalancer.LbRuleCondition
	conditions := d.Get("header_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := initLbHTTPRuleMatchConditionFromSchema(data, "LbHttpRequestHeaderCondition", true, false)

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("cookie_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := initLbHTTPRuleMatchConditionFromSchema(data, "LbHttpRequestCookieCondition", false, true)

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("body_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := initLbHTTPRuleMatchConditionFromSchema(data, "LbHttpRequestBodyCondition", false, false)
		elem.BodyValue = data["value"].(string)

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("method_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := loadbalancer.LbRuleCondition{
			Inverse: data["inverse"].(bool),
			Type_:   "LbHttpRequestMethodCondition",
			Method:  data["method"].(string),
		}

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("version_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := loadbalancer.LbRuleCondition{
			Inverse: data["inverse"].(bool),
			Type_:   "LbHttpRequestVersionCondition",
			Version: data["version"].(string),
		}

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("uri_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := initLbHTTPRuleMatchConditionFromSchema(data, "LbHttpRequestUriCondition", false, false)
		elem.Uri = data["uri"].(string)

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("ip_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := loadbalancer.LbRuleCondition{
			Inverse:       data["inverse"].(bool),
			Type_:         "LbIpHeaderCondition",
			SourceAddress: data["source_address"].(string),
		}

		conditionList = append(conditionList, elem)
	}

	conditions = d.Get("tcp_condition").(*schema.Set).List()
	for _, condition := range conditions {
		data := condition.(map[string]interface{})
		elem := loadbalancer.LbRuleCondition{
			Inverse:    data["inverse"].(bool),
			Type_:      "LbTcpHeaderCondition",
			SourcePort: data["source_port"].(string),
		}

		conditionList = append(conditionList, elem)
	}

	return conditionList
}

func setLbRuleHTTPForwardingConditionsInSchema(d *schema.ResourceData, conditions []loadbalancer.LbRuleCondition) {
	var headerConditionList []map[string]interface{}
	var cookieConditionList []map[string]interface{}
	var bodyConditionList []map[string]interface{}
	var methodConditionList []map[string]interface{}
	var versionConditionList []map[string]interface{}
	var uriConditionList []map[string]interface{}
	var ipConditionList []map[string]interface{}
	var tcpConditionList []map[string]interface{}

	for _, condition := range conditions {
		elem := make(map[string]interface{})

		if condition.Type_ == "LbHttpRequestHeaderCondition" {
			elem["name"] = condition.HeaderName
			elem["value"] = condition.HeaderValue
			elem["inverse"] = condition.Inverse
			elem["match_type"] = condition.MatchType
			elem["case_sensitive"] = *condition.CaseSensitive
			headerConditionList = append(headerConditionList, elem)
		}

		if condition.Type_ == "LbHttpRequestCookieCondition" {
			elem["name"] = condition.CookieName
			elem["value"] = condition.CookieValue
			elem["inverse"] = condition.Inverse
			elem["match_type"] = condition.MatchType
			elem["case_sensitive"] = *condition.CaseSensitive
			cookieConditionList = append(cookieConditionList, elem)
		}

		if condition.Type_ == "LbHttpRequestBodyCondition" {
			elem["value"] = condition.BodyValue
			elem["inverse"] = condition.Inverse
			elem["match_type"] = condition.MatchType
			elem["case_sensitive"] = *condition.CaseSensitive
			bodyConditionList = append(bodyConditionList, elem)
		}

		if condition.Type_ == "LbHttpRequestMethodCondition" {
			elem["method"] = condition.Method
			elem["inverse"] = condition.Inverse
			methodConditionList = append(methodConditionList, elem)
		}

		if condition.Type_ == "LbHttpRequestVersionCondition" {
			elem["version"] = condition.Version
			elem["inverse"] = condition.Inverse
			versionConditionList = append(versionConditionList, elem)
		}

		if condition.Type_ == "LbHttpRequestUriCondition" {
			elem["uri"] = condition.Uri
			elem["inverse"] = condition.Inverse
			elem["match_type"] = condition.MatchType
			elem["case_sensitive"] = *condition.CaseSensitive
			uriConditionList = append(uriConditionList, elem)
		}

		if condition.Type_ == "LbIpHeaderCondition" {
			elem["source_address"] = condition.SourceAddress
			elem["inverse"] = condition.Inverse
			ipConditionList = append(ipConditionList, elem)
		}

		if condition.Type_ == "LbTcpHeaderCondition" {
			elem["source_port"] = condition.SourcePort
			elem["inverse"] = condition.Inverse
			tcpConditionList = append(tcpConditionList, elem)
		}

		// TODO: optimize this code with map of conditions and a loop
		warningString := "[WARNING]: Failed to set %s in schema: %v"
		err := d.Set("header_condition", headerConditionList)
		if err != nil {
			log.Printf(warningString, "header_condition", err)
		}

		err = d.Set("cookie_condition", cookieConditionList)
		if err != nil {
			log.Printf(warningString, "cookie_condition", err)
		}

		err = d.Set("body_condition", bodyConditionList)
		if err != nil {
			log.Printf(warningString, "body_condition", err)
		}

		err = d.Set("method_condition", methodConditionList)
		if err != nil {
			log.Printf(warningString, "method_condition", err)
		}

		err = d.Set("version_condition", versionConditionList)
		if err != nil {
			log.Printf(warningString, "version_condition", err)
		}

		err = d.Set("uri_condition", uriConditionList)
		if err != nil {
			log.Printf(warningString, "uri_condition", err)
		}

		err = d.Set("ip_condition", ipConditionList)
		if err != nil {
			log.Printf(warningString, "ip_condition", err)
		}

		err = d.Set("tcp_condition", tcpConditionList)
		if err != nil {
			log.Printf(warningString, "tcp_condition", err)
		}

	}
}

func getLbRuleForwardingActionsFromSchema(d *schema.ResourceData) []loadbalancer.LbRuleAction {
	var actionList []loadbalancer.LbRuleAction
	actions := d.Get("http_reject_action").(*schema.Set).List()
	for _, action := range actions {
		data := action.(map[string]interface{})
		elem := loadbalancer.LbRuleAction{
			Type_:        "LbHttpRejectAction",
			ReplyStatus:  data["reply_status"].(string),
			ReplyMessage: data["reply_message"].(string),
		}

		actionList = append(actionList, elem)
	}

	actions = d.Get("http_redirect_action").(*schema.Set).List()
	for _, action := range actions {
		data := action.(map[string]interface{})
		elem := loadbalancer.LbRuleAction{
			Type_:          "LbHttpRedirectAction",
			RedirectStatus: data["redirect_status"].(string),
			RedirectUrl:    data["redirect_url"].(string),
		}

		actionList = append(actionList, elem)
	}

	actions = d.Get("select_pool_action").(*schema.Set).List()
	for _, action := range actions {
		data := action.(map[string]interface{})
		elem := loadbalancer.LbRuleAction{
			Type_:  "LbSelectPoolAction",
			PoolId: data["pool_id"].(string),
		}

		actionList = append(actionList, elem)
	}

	return actionList
}

func setLbRuleForwardingActionsInSchema(d *schema.ResourceData, actions []loadbalancer.LbRuleAction) error {
	var rejectActionList []map[string]string
	var redirectActionList []map[string]string
	var selectActionList []map[string]string

	for _, action := range actions {
		elem := make(map[string]string)
		if action.Type_ == "LbHttpRejectAction" {
			elem["reply_status"] = action.ReplyStatus
			elem["reply_message"] = action.ReplyMessage
			rejectActionList = append(rejectActionList, elem)
		}

		if action.Type_ == "LbHttpRedirectAction" {
			elem["redirect_status"] = action.RedirectStatus
			elem["redirect_url"] = action.RedirectUrl
			redirectActionList = append(redirectActionList, elem)
		}

		if action.Type_ == "LbSelectPoolAction" {
			elem["pool_id"] = action.PoolId
			selectActionList = append(selectActionList, elem)
		}

		err := d.Set("http_reject_action", rejectActionList)
		if err != nil {
			return err
		}

		err = d.Set("http_redirect_action", redirectActionList)
		if err != nil {
			return err
		}

		err = d.Set("select_pool_action", selectActionList)
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceNsxtLbHTTPForwardingRuleCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	matchConditions := getLbRuleHTTPForwardingConditionsFromSchema(d)
	actions := getLbRuleForwardingActionsFromSchema(d)
	matchStrategy := d.Get("match_strategy").(string)
	phase := "HTTP_FORWARDING"

	lbRule := loadbalancer.LbRule{
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		Actions:         actions,
		MatchConditions: matchConditions,
		MatchStrategy:   matchStrategy,
		Phase:           phase,
	}

	lbRule, resp, err := nsxClient.ServicesApi.CreateLoadBalancerRule(nsxClient.Context, lbRule)

	if err != nil {
		return fmt.Errorf("Error during LoadBalancerRule create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LoadBalancerRule create: %v", resp.StatusCode)
	}
	d.SetId(lbRule.Id)

	return resourceNsxtLbHTTPForwardingRuleRead(d, m)
}

func resourceNsxtLbHTTPForwardingRuleRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbRule, resp, err := nsxClient.ServicesApi.ReadLoadBalancerRule(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LoadBalancerRule %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LoadBalancerRule read: %v", err)
	}

	d.Set("revision", lbRule.Revision)
	d.Set("description", lbRule.Description)
	d.Set("display_name", lbRule.DisplayName)
	setTagsInSchema(d, lbRule.Tags)
	setLbRuleHTTPForwardingConditionsInSchema(d, lbRule.MatchConditions)
	d.Set("match_strategy", lbRule.MatchStrategy)
	err = setLbRuleForwardingActionsInSchema(d, lbRule.Actions)

	if err != nil {
		log.Printf("[DEBUG] Failed to set action in LoadBalancerRule %v: %v", id, err)
		d.SetId("")
		return nil
	}

	return nil
}

func resourceNsxtLbHTTPForwardingRuleUpdate(d *schema.ResourceData, m interface{}) error {
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
	matchConditions := getLbRuleHTTPForwardingConditionsFromSchema(d)
	actions := getLbRuleForwardingActionsFromSchema(d)
	matchStrategy := d.Get("match_strategy").(string)
	phase := "HTTP_FORWARDING"

	lbRule := loadbalancer.LbRule{
		Revision:        revision,
		Description:     description,
		DisplayName:     displayName,
		MatchStrategy:   matchStrategy,
		Phase:           phase,
		Actions:         actions,
		MatchConditions: matchConditions,
		Tags:            tags,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerRule(nsxClient.Context, id, lbRule)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LoadBalancerRule update: %v", err)
	}

	return resourceNsxtLbHTTPForwardingRuleRead(d, m)
}
