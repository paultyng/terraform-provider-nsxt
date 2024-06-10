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
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyPredefinedSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyPredefinedSecurityPolicyCreate,
		Read:   resourceNsxtPolicyPredefinedSecurityPolicyRead,
		Update: resourceNsxtPolicyPredefinedSecurityPolicyUpdate,
		Delete: resourceNsxtPolicyPredefinedSecurityPolicyDelete,

		Schema: getPolicyPredefinedSecurityPolicySchema(),
	}
}

func getSecurityPolicyDefaultRulesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of default rules",
		Optional:    true,
		Computed:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"nsx_id":  getComputedNsxIDSchema(),
				"context": getContextSchema(false, false, false),
				"scope": {
					Type:        schema.TypeString,
					Description: "Scope for this rule",
					Computed:    true,
				},
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

func getPolicyPredefinedSecurityPolicySchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"path":         getPolicyPathSchema(true, true, "Path for this Security Policy"),
		"description":  getComputedDescriptionSchema(),
		"tag":          getTagsSchema(),
		"rule":         getSecurityPolicyAndGatewayRulesSchema(false, false, false),
		"default_rule": getSecurityPolicyDefaultRulesSchema(),
		"revision":     getRevisionSchema(),
		"context":      getContextSchema(false, false, false),
	}
}

func updateSecurityPolicyDefaultRule(rule model.Rule, d *schema.ResourceData) *model.Rule {
	defaultRules := d.Get("default_rule").([]interface{})

	for _, obj := range defaultRules {
		// Only one is allowed here
		defaultRule := obj.(map[string]interface{})

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

	// This rule is not present in new config - check if was just deleted
	// If so, the rule needs to be reverted
	if d.HasChange("default_rule") {
		rule := revertSecurityPolicyDefaultRule(rule)
		return &rule
	}

	return nil
}

func revertPolicyPredefinedSecurityPolicy(predefinedPolicy model.SecurityPolicy, m interface{}) (model.SecurityPolicy, error) {

	var childRules []*data.StructValue

	for _, rule := range predefinedPolicy.Rules {
		if rule.IsDefault != nil && *rule.IsDefault {
			log.Printf("[DEBUG]: Reverting default rule %s", *rule.Id)
			revertedRule := revertSecurityPolicyDefaultRule(rule)
			childRule, err := createPolicyChildRule(*revertedRule.Id, revertedRule, false)
			if err != nil {
				return model.SecurityPolicy{}, err
			}
			childRules = append(childRules, childRule)
		} else {
			// Mark for delete all non-default rules
			log.Printf("[DEBUG]: Deleting rule %s", *rule.Id)
			childRule, err := createPolicyChildRule(*rule.Id, rule, true)
			if err != nil {
				return model.SecurityPolicy{}, err
			}
			childRules = append(childRules, childRule)
		}
	}

	empty := ""
	predefinedPolicy.Description = &empty
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

func revertSecurityPolicyDefaultRule(rule model.Rule) model.Rule {

	if len(rule.Tags) > 0 {
		tags := make([]model.Tag, 0)
		rule.Tags = tags
	}
	defaultAction := "ALLOW"
	rule.Action = &defaultAction
	logged := false
	rule.Logged = &logged
	logLabel := ""
	rule.Tag = &logLabel

	return rule
}

func createChildDomainWithSecurityPolicy(domain string, policyID string, policy model.SecurityPolicy) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	childPolicy := model.ChildSecurityPolicy{
		ResourceType:   "ChildSecurityPolicy",
		SecurityPolicy: &policy,
	}

	dataValue, errors := converter.ConvertToVapi(childPolicy, model.ChildSecurityPolicyBindingType())
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

func updatePolicyPredefinedSecurityPolicy(id string, d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)
	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)

	if domain == "" {
		return fmt.Errorf("Failed to extract domain from Security Policy path %s", path)
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	predefinedPolicy, err := getSecurityPolicyInDomain(context, id, domain, connector)
	if err != nil {
		return err
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
	}

	if d.HasChange("default_rule") {
		for _, existingDefaultRule := range predefinedPolicy.Rules {
			if existingDefaultRule.IsDefault != nil && *existingDefaultRule.IsDefault {
				updatedDefaultRule := updateSecurityPolicyDefaultRule(existingDefaultRule, d)
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

	predefinedPolicy.Rules = nil
	log.Printf("[DEBUG]: Updating default policy %s with %d child rules", id, len(childRules))
	if len(childRules) > 0 {
		predefinedPolicy.Children = childRules
	}

	if err != nil {
		return err
	}
	err = securityPolicyInfraPatch(context, predefinedPolicy, domain, m)
	if err != nil {
		return handleUpdateError("Predefined Security Policy", id, err)
	}

	return nil
}

func resourceNsxtPolicyPredefinedSecurityPolicyCreate(d *schema.ResourceData, m interface{}) error {
	path := d.Get("path").(string)
	id := getPolicyIDFromPath(path)

	if id == "" {
		return fmt.Errorf("Failed to extract ID from Security Policy path %s", path)
	}

	err := updatePolicyPredefinedSecurityPolicy(id, d, m)
	if err != nil {
		return err
	}

	d.SetId(id)

	return resourceNsxtPolicyPredefinedSecurityPolicyRead(d, m)
}

func resourceNsxtPolicyPredefinedSecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy ID")
	}

	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewSecurityPoliciesClient(context, connector)
	obj, err := client.Get(domain, id)
	if err != nil {
		return handleReadError(d, "Predefined Security Policy", id, err)
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

func resourceNsxtPolicyPredefinedSecurityPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Predefined Security Policy ID")
	}
	err := updatePolicyPredefinedSecurityPolicy(id, d, m)
	if err != nil {
		return err
	}

	return resourceNsxtPolicyPredefinedSecurityPolicyRead(d, m)
}

func resourceNsxtPolicyPredefinedSecurityPolicyDelete(d *schema.ResourceData, m interface{}) error {
	// Delete means revert back to default
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Predefined Security Policy ID")
	}

	path := d.Get("path").(string)
	domain := getDomainFromResourcePath(path)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	predefinedPolicy, err := getSecurityPolicyInDomain(context, id, domain, getPolicyConnector(m))
	if err != nil {
		return err
	}

	revertedPolicy, err := revertPolicyPredefinedSecurityPolicy(predefinedPolicy, m)
	if err != nil {
		return fmt.Errorf("Failed to revert Predefined Security Policy %s: %s", id, err)
	}

	err = securityPolicyInfraPatch(context, revertedPolicy, domain, m)
	if err != nil {
		return handleUpdateError("Predefined Security Policy", id, err)
	}
	return nil
}

func securityPolicyInfraPatch(context utl.SessionContext, policy model.SecurityPolicy, domain string, m interface{}) error {
	childDomain, err := createChildDomainWithSecurityPolicy(domain, *policy.Id, policy)
	if err != nil {
		return fmt.Errorf("Failed to create H-API for Predefined Security Policy: %s", err)
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
