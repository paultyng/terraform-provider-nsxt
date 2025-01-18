// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyGatewayPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayPolicyCreate,
		Read:   resourceNsxtPolicyGatewayPolicyRead,
		Update: resourceNsxtPolicyGatewayPolicyUpdate,
		Delete: resourceNsxtPolicyGatewayPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDomainResourceImporter,
		},

		Schema: getPolicyGatewayPolicySchema(false),
	}
}

func getGatewayPolicy(sessionContext utl.SessionContext, id string, domainName string, connector client.Connector) (model.GatewayPolicy, error) {
	client := domains.NewGatewayPoliciesClient(sessionContext, connector)
	if client == nil {
		return model.GatewayPolicy{}, policyResourceNotSupportedError()
	}
	return client.Get(domainName, id)

}

func resourceNsxtPolicyGatewayPolicyExistsInDomain(sessionContext utl.SessionContext, id string, domainName string, connector client.Connector) (bool, error) {
	_, err := getGatewayPolicy(sessionContext, id, domainName, connector)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving Gateway Policy", err)
}

func resourceNsxtPolicyGatewayPolicyExistsPartial(domainName string) func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicyGatewayPolicyExistsInDomain(sessionContext, id, domainName, connector)
	}
}

func getUpdatedRuleChildren(d *schema.ResourceData) ([]*data.StructValue, error) {
	var policyChildren []*data.StructValue

	if !d.HasChange("rule") {
		return nil, nil
	}

	oldRules, newRules := d.GetChange("rule")
	rules := getPolicyRulesFromSchema(d)
	newRulesCount := len(newRules.([]interface{}))
	oldRulesCount := len(oldRules.([]interface{}))
	for ruleNo := 0; ruleNo < newRulesCount; ruleNo++ {
		ruleIndicator := fmt.Sprintf("rule.%d", ruleNo)
		autoAssignedSequence := false
		originalSequence := d.Get(fmt.Sprintf("%s.sequence_number", ruleIndicator)).(int)
		if originalSequence == 0 {
			autoAssignedSequence = true
		}
		if d.HasChange(ruleIndicator) || autoAssignedSequence {
			// If the provider assigned sequence number to this rule, we need to update it even
			// though terraform sees no diff
			rule := rules[ruleNo]
			// New or updated rule
			ruleID := newUUID()
			if rule.Id != nil {
				ruleID = *rule.Id
				log.Printf("[DEBUG]: Updating child rule with id %s", *rule.Id)
			} else {
				log.Printf("[DEBUG]: Adding child rule with id %s", ruleID)
				rule.Id = &ruleID
			}

			childRule, err := createPolicyChildRule(ruleID, rule, false)
			if err != nil {
				return policyChildren, err
			}
			policyChildren = append(policyChildren, childRule)
		}
	}

	resourceType := "Rule"
	for ruleNo := newRulesCount; ruleNo < oldRulesCount; ruleNo++ {
		// delete
		ruleIndicator := fmt.Sprintf("rule.%d", ruleNo)
		oldRule, _ := d.GetChange(ruleIndicator)
		oldRuleMap := oldRule.(map[string]interface{})
		oldRuleID := oldRuleMap["nsx_id"].(string)
		rule := model.Rule{
			Id:           &oldRuleID,
			ResourceType: &resourceType,
		}

		childRule, err := createPolicyChildRule(oldRuleID, rule, true)
		if err != nil {
			return policyChildren, err
		}
		log.Printf("[DEBUG]: Deleting child rule with id %s", oldRuleID)
		policyChildren = append(policyChildren, childRule)

	}

	return policyChildren, nil

}

func policyGatewayPolicyBuildAndPatch(d *schema.ResourceData, m interface{}, connector client.Connector, isGlobalManager bool, id string, isVPC bool) error {

	domain := ""
	if !isVPC {
		domain = d.Get("domain").(string)
	}
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags, tagErr := getValidatedTagsFromSchema(d)
	if tagErr != nil {
		return tagErr
	}

	comments := d.Get("comments").(string)
	locked := d.Get("locked").(bool)
	sequenceNumber := int64(d.Get("sequence_number").(int))
	stateful := d.Get("stateful").(bool)
	revision := int64(d.Get("revision").(int))
	objType := "GatewayPolicy"

	obj := model.GatewayPolicy{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		Comments:       &comments,
		Locked:         &locked,
		SequenceNumber: &sequenceNumber,
		Stateful:       &stateful,
		ResourceType:   &objType,
		Id:             &id,
	}

	if !isVPC {
		category := d.Get("category").(string)
		obj.Category = &category
	}
	_, isSet := d.GetOkExists("tcp_strict")
	if isSet {
		tcpStrict := d.Get("tcp_strict").(bool)
		obj.TcpStrict = &tcpStrict
	}

	if len(d.Id()) > 0 {
		// This is update flow
		obj.Revision = &revision
	}

	policyChildren, err := getUpdatedRuleChildren(d)
	if err != nil {
		return err
	}
	if len(policyChildren) > 0 {
		obj.Children = policyChildren
	}

	return gatewayPolicyInfraPatch(getSessionContext(d, m), obj, domain, m)
}

func resourceNsxtPolicyGatewayPolicyCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyGatewayPolicyExistsPartial(d.Get("domain").(string)))
	if err != nil {
		return err
	}

	err = policyGatewayPolicyBuildAndPatch(d, m, connector, isPolicyGlobalManager(m), id, false)
	if err != nil {
		return handleCreateError("Gateway Policy", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayPolicyRead(d, m)
}

func resourceNsxtPolicyGatewayPolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Policy ID")
	}

	obj, err := getGatewayPolicy(getSessionContext(d, m), id, d.Get("domain").(string), connector)
	if err != nil {
		return handleReadError(d, "Gateway Policy", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("domain", getDomainFromResourcePath(*obj.Path))
	d.Set("category", obj.Category)
	d.Set("comments", obj.Comments)
	d.Set("locked", obj.Locked)
	d.Set("sequence_number", obj.SequenceNumber)
	d.Set("stateful", obj.Stateful)
	if obj.TcpStrict != nil {
		// tcp_strict is dependant on stateful and maybe nil
		d.Set("tcp_strict", *obj.TcpStrict)
	}
	d.Set("revision", obj.Revision)
	return setPolicyRulesInSchema(d, obj.Rules)
}

func resourceNsxtPolicyGatewayPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Policy ID")
	}

	err := policyGatewayPolicyBuildAndPatch(d, m, connector, isPolicyGlobalManager(m), id, false)
	if err != nil {
		return handleUpdateError("Gateway Policy", id, err)
	}

	return resourceNsxtPolicyGatewayPolicyRead(d, m)
}

func resourceNsxtPolicyGatewayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Policy ID")
	}

	connector := getPolicyConnector(m)
	client := domains.NewGatewayPoliciesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err := client.Delete(d.Get("domain").(string), id)
	if err != nil {
		return handleDeleteError("Gateway Policy", id, err)
	}

	return nil
}
