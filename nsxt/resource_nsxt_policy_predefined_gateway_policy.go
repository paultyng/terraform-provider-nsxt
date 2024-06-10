/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyPredefinedGatewayPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyPredefinedGatewayPolicyCreate,
		Read:   resourceNsxtPolicyPredefinedGatewayPolicyRead,
		Update: resourceNsxtPolicyPredefinedGatewayPolicyUpdate,
		Delete: resourceNsxtPolicyPredefinedGatewayPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPredefinedPolicyImporter,
		},

		Schema: getPolicyPredefinedGatewayPolicySchema(),
	}
}

func getPolicyPredefinedGatewayPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"path":         getPolicyPathSchema(true, true, "Path for this Gateway Policy"),
		"description":  getComputedDescriptionSchema(),
		"tag":          getTagsSchema(),
		"rule":         getSecurityPolicyAndGatewayRulesSchema(true, false, false),
		"default_rule": getGatewayPolicyDefaultRulesSchema(),
		"revision":     getRevisionSchema(),
		"context":      getContextSchema(false, false, false),
	}
}

func getGatewayPolicyDefaultRulesSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeList,
		Description:   "List of default rules",
		Optional:      true,
		Computed:      true,
		ConflictsWith: []string{"rule"},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"nsx_id":      getComputedNsxIDSchema(),
				"scope":       getPolicyPathSchema(true, false, "Scope for this rule"),
				"description": getComputedDescriptionSchema(),
				"path":        getPathSchema(),
				"revision":    getRevisionSchema(),
				"logged": {
					Type:        schema.TypeBool,
					Description: "Flag to enable packet logging",
					Optional:    true,
					Default:     false,
				},
				"tag": getTagsSchema(),
				"log_label": {
					Type:        schema.TypeString,
					Description: "Additional information (string) which will be propagated to the rule syslog",
					Optional:    true,
				},
				"action": {
					Type:         schema.TypeString,
					Description:  "Action",
					Optional:     true,
					ValidateFunc: validation.StringInSlice(securityPolicyActionValues, false),
					Default:      model.Rule_ACTION_ALLOW,
				},
				"sequence_number": {
					Type:        schema.TypeInt,
					Description: "Sequence number of the this rule",
					Computed:    true,
				},
			},
		},
	}
}

func updateGatewayPolicyDefaultRuleByScope(rule model.Rule, d *schema.ResourceData, connector client.Connector, isGlobalManager bool) *model.Rule {
	defaultRules := d.Get("default_rule").([]interface{})

	for _, obj := range defaultRules {
		defaultRule := obj.(map[string]interface{})
		scope := defaultRule["scope"].(string)

		if len(rule.Scope) == 1 && scope == rule.Scope[0] {
			description := defaultRule["description"].(string)
			rule.Description = &description
			action := defaultRule["action"].(string)
			rule.Action = &action
			logLabel := defaultRule["log_label"].(string)
			rule.Tag = &logLabel
			logged := defaultRule["logged"].(bool)
			rule.Logged = &logged
			tags := getPolicyTagsFromSet(defaultRule["tag"].(*schema.Set))
			if len(tags) > 0 || len(rule.Tags) > 0 {
				rule.Tags = tags
			}

			log.Printf("[DEBUG] Updating Default Rule with ID %s", *rule.Id)
			return &rule
		}
	}

	// This rule is not present in new config - check if was just deleted
	// If so, the rule needs to be reverted
	_, oldRules := d.GetChange("default_rule")
	for _, oldRule := range oldRules.([]interface{}) {
		oldRuleMap := oldRule.(map[string]interface{})
		if oldID, ok := oldRuleMap["nsx_id"]; ok {
			if (rule.Id != nil) && (*rule.Id == oldID.(string)) {
				rule := revertGatewayPolicyDefaultRule(rule)

				log.Printf("[DEBUG] Reverting Default Rule with ID %s", *rule.Id)
				return &rule
			}
		}
	}

	return nil
}

func setPolicyDefaultRulesInSchema(d *schema.ResourceData, rules []model.Rule) error {
	var rulesList []map[string]interface{}
	for _, rule := range rules {
		elem := make(map[string]interface{})
		elem["description"] = rule.Description
		elem["log_label"] = rule.Tag
		elem["logged"] = rule.Logged
		elem["action"] = rule.Action
		elem["revision"] = rule.Revision
		if len(rule.Scope) > 0 {
			elem["scope"] = rule.Scope[0]
		}
		elem["sequence_number"] = rule.SequenceNumber
		elem["tag"] = initPolicyTagsSet(rule.Tags)
		elem["path"] = rule.Path
		elem["nsx_id"] = rule.Id

		rulesList = append(rulesList, elem)
	}

	return d.Set("default_rule", rulesList)
}

func revertPolicyPredefinedGatewayPolicy(predefinedPolicy model.GatewayPolicy, m interface{}) (model.GatewayPolicy, error) {

	// Default values for Name and Description are ID
	empty := ""
	predefinedPolicy.Description = &empty

	var childRules []*data.StructValue

	for _, rule := range predefinedPolicy.Rules {
		if rule.IsDefault != nil && *rule.IsDefault {
			log.Printf("[DEBUG]: Reverting default rule %s", *rule.Id)
			revertedRule := revertGatewayPolicyDefaultRule(rule)
			childRule, err := createPolicyChildRule(*revertedRule.Id, revertedRule, false)
			if err != nil {
				return model.GatewayPolicy{}, err
			}
			childRules = append(childRules, childRule)
		} else {
			// Mark for delete all non-default rules
			log.Printf("[DEBUG]: Deleting rule %s", *rule.Id)
			childRule, err := createPolicyChildRule(*rule.Id, rule, true)
			if err != nil {
				return model.GatewayPolicy{}, err
			}
			childRules = append(childRules, childRule)
		}
	}

	predefinedPolicy.Rules = nil
	if len(childRules) > 0 {
		predefinedPolicy.Children = childRules
	}

	if len(predefinedPolicy.Tags) > 0 {
		tags := make([]model.Tag, 0)
		predefinedPolicy.Tags = tags
	}

	return predefinedPolicy, nil
}

func revertGatewayPolicyDefaultRule(rule model.Rule) model.Rule {

	// NOTE: ability to control default action from gateway config is deprecated and discouraged
	// Deleting default rule in here will not respect force_whitelisting setting on the gateway
	if len(rule.Tags) > 0 {
		tags := make([]model.Tag, 0)
		rule.Tags = tags
	}

	empty := ""
	rule.Description = &empty
	defaultAction := model.Rule_ACTION_ALLOW
	rule.Action = &defaultAction
	return rule
}

func createPolicyChildRule(ruleID string, rule model.Rule, shouldDelete bool) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	childRule := model.ChildRule{
		ResourceType:    "ChildRule",
		Id:              &ruleID,
		Rule:            &rule,
		MarkedForDelete: &shouldDelete,
	}

	dataValue, errors := converter.ConvertToVapi(childRule, model.ChildRuleBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return dataValue.(*data.StructValue), nil
}

func createChildDomainWithGatewayPolicy(domain string, policyID string, policy model.GatewayPolicy) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	childPolicy := model.ChildGatewayPolicy{
		Id:            &policyID,
		ResourceType:  "ChildGatewayPolicy",
		GatewayPolicy: &policy,
	}

	dataValue, errors := converter.ConvertToVapi(childPolicy, model.ChildGatewayPolicyBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}

	var domainChildren []*data.StructValue
	domainChildren = append(domainChildren, dataValue.(*data.StructValue))

	targetType := "Domain"
	childDomain := model.ChildResourceReference{
		Id:           &domain,
		ResourceType: "ChildResourceReference",
		TargetType:   &targetType,
		Children:     domainChildren,
	}

	dataValue, errors = converter.ConvertToVapi(childDomain, model.ChildResourceReferenceBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func gatewayPolicyInfraPatch(context utl.SessionContext, policy model.GatewayPolicy, domain string, m interface{}) error {
	childDomain, err := createChildDomainWithGatewayPolicy(domain, *policy.Id, policy)
	if err != nil {
		return fmt.Errorf("Failed to create H-API for Predefined Gateway Policy: %s", err)
	}

	var infraChildren []*data.StructValue
	infraChildren = append(infraChildren, childDomain)

	infraType := "Infra"
	infraObj := model.Infra{
		Children:     infraChildren,
		ResourceType: &infraType,
	}

	return policyInfraPatch(context, infraObj, getPolicyConnector(m), false)

}

func updatePolicyPredefinedGatewayPolicy(id string, d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)
	isGlobalManager := isPolicyGlobalManager(m)
	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)

	if domain == "" {
		return fmt.Errorf("Failed to extract domain from Gateway Policy path %s", path)
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	predefinedPolicy, err := getGatewayPolicyInDomain(context, id, domain, connector)
	if err != nil {
		return err
	}

	if predefinedPolicy.Category != nil && *predefinedPolicy.Category == "SystemRules" {
		return fmt.Errorf("System policy can not be modified")
	}

	if d.HasChange("description") {
		description := d.Get("description").(string)
		predefinedPolicy.Description = &description
	}

	if d.HasChange("tag") {
		predefinedPolicy.Tags = getPolicyTagsFromSchema(d)
	}

	var childRules []*data.StructValue
	if d.HasChange("rule") {
		oldRules, _ := d.GetChange("rule")
		rules := getPolicyRulesFromSchema(d)

		existingRules := make(map[string]bool)
		for _, rule := range rules {
			ruleID := newUUID()
			if rule.Id != nil {
				ruleID = *rule.Id
				existingRules[ruleID] = true
			} else {
				rule.Id = &ruleID
			}

			childRule, err := createPolicyChildRule(ruleID, rule, false)
			if err != nil {
				return err
			}
			log.Printf("[DEBUG]: Adding child rule with id %s", ruleID)
			childRules = append(childRules, childRule)
		}

		// We need to delete old rules that are not present in config anymore
		for _, oldRule := range oldRules.([]interface{}) {
			oldRuleMap := oldRule.(map[string]interface{})
			oldRuleID := oldRuleMap["nsx_id"].(string)
			if _, exists := existingRules[oldRuleID]; !exists {
				resourceType := "Rule"
				rule := model.Rule{
					Id:           &oldRuleID,
					ResourceType: &resourceType,
				}

				childRule, err := createPolicyChildRule(oldRuleID, rule, true)
				if err != nil {
					return err
				}
				log.Printf("[DEBUG]: Deleting child rule with id %s", oldRuleID)
				childRules = append(childRules, childRule)

			}
		}
	} else if d.HasChange("default_rule") {
		log.Printf("[DEBUG]: Default rule configuration has changed")
		for _, existingDefaultRule := range predefinedPolicy.Rules {
			if existingDefaultRule.IsDefault != nil && *existingDefaultRule.IsDefault {
				updatedDefaultRule := updateGatewayPolicyDefaultRuleByScope(existingDefaultRule, d, connector, isGlobalManager)
				if updatedDefaultRule != nil {
					childRule, err := createPolicyChildRule(*updatedDefaultRule.Id, *updatedDefaultRule, false)
					if err != nil {
						return err
					}
					childRules = append(childRules, childRule)
				}
			}
		}
	}

	log.Printf("[DEBUG]: Updating default policy %s with %d child rules", id, len(childRules))
	predefinedPolicy.Rules = nil
	if len(childRules) > 0 {
		predefinedPolicy.Children = childRules
	}

	if err != nil {
		return err
	}
	err = gatewayPolicyInfraPatch(context, predefinedPolicy, domain, m)
	if err != nil {
		return handleUpdateError("Predefined Gateway Policy", id, err)
	}

	return nil
}

func resourceNsxtPolicyPredefinedGatewayPolicyCreate(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)
	id := getPolicyIDFromPath(path)

	if id == "" {
		return fmt.Errorf("Failed to extract ID from Gateway Policy path %s", path)
	}

	err := updatePolicyPredefinedGatewayPolicy(id, d, m)
	if err != nil {
		return err
	}

	d.SetId(id)

	return resourceNsxtPolicyPredefinedGatewayPolicyRead(d, m)
}

func resourceNsxtPolicyPredefinedGatewayPolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Policy ID")
	}

	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewGatewayPoliciesClient(context, connector)
	obj, err := client.Get(domain, id)
	if err != nil {
		return handleReadError(d, "Predefined Gateway Policy", id, err)
	}

	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	var rules []model.Rule
	var defaultRules []model.Rule

	for _, rule := range obj.Rules {
		if rule.IsDefault != nil && *rule.IsDefault {
			defaultRules = append(defaultRules, rule)
		} else {
			rules = append(rules, rule)
		}
	}

	err = setPolicyRulesInSchema(d, rules)
	if err != nil {
		return err
	}
	return setPolicyDefaultRulesInSchema(d, defaultRules)
}

func resourceNsxtPolicyPredefinedGatewayPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Predefined Gateway Policy ID")
	}
	err := updatePolicyPredefinedGatewayPolicy(id, d, m)
	if err != nil {
		return err
	}

	return resourceNsxtPolicyPredefinedGatewayPolicyRead(d, m)
}

func resourceNsxtPolicyPredefinedGatewayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	// Delete means revert back to default
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Predefined Gateway Policy ID")
	}

	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	predefinedPolicy, err := getGatewayPolicyInDomain(context, id, domain, getPolicyConnector(m))
	if err != nil {
		return err
	}

	revertedPolicy, err := revertPolicyPredefinedGatewayPolicy(predefinedPolicy, m)
	if err != nil {
		return fmt.Errorf("Failed to revert Predefined Gateway Policy %s: %s", id, err)
	}

	err = gatewayPolicyInfraPatch(context, revertedPolicy, domain, m)
	if err != nil {
		return handleUpdateError("Predefined Gateway Policy", id, err)
	}
	return nil
}

func nsxtPredefinedPolicyImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importPath := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		d.Set("path", importPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	d.Set("path", importPath)
	id := getPolicyIDFromPath(importPath)
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}
