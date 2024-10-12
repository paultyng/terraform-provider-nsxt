// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	securitypolicies "github.com/vmware/terraform-provider-nsxt/api/infra/domains/security_policies"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicySecurityPolicyRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySecurityPolicyRuleCreate,
		Read:   resourceNsxtPolicySecurityPolicyRuleRead,
		Update: resourceNsxtPolicySecurityPolicyRuleUpdate,
		Delete: resourceNsxtPolicySecurityPolicyRuleDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtSecurityPolicyRuleImporter,
		},
		Schema: getSecurityPolicyAndGatewayRuleSchema(false, false, false, true),
	}
}

func resourceNsxtPolicySecurityPolicyRuleCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	policyPath := d.Get("policy_path").(string)
	projectID := getProjectIDFromResourcePath(policyPath)
	domain := getDomainFromResourcePath(policyPath)
	policyID := getPolicyIDFromPath(policyPath)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicySecurityPolicyRuleExistsPartial(d, m, policyPath))
	if err != nil {
		return err
	}

	if err := setSecurityPolicyRuleContext(d, projectID); err != nil {
		return handleCreateError("SecurityPolicyRule", fmt.Sprintf("%s/%s", policyPath, id), err)
	}

	log.Printf("[INFO] Creating Security Policy Rule with ID %s under policy %s", id, policyPath)
	client := securitypolicies.NewRulesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	rule := securityPolicyRuleSchemaToModel(d, id)
	err = client.Patch(domain, policyID, id, rule)
	if err != nil {
		return handleCreateError("SecurityPolicyRule", fmt.Sprintf("%s/%s", policyPath, id), err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySecurityPolicyRuleRead(d, m)
}

func setSecurityPolicyRuleContext(d *schema.ResourceData, projectID string) error {
	providedProjectID, _ := getContextDataFromSchema(d)
	if providedProjectID == "" {
		contexts := make([]interface{}, 1)
		ctxMap := make(map[string]interface{})
		ctxMap["project_id"] = projectID
		contexts[0] = ctxMap
		return d.Set("context", contexts)
	} else if providedProjectID != projectID {
		return fmt.Errorf("provided project_id in context is inconsist with the project_id in policy_path")
	}
	return nil
}

func securityPolicyRuleSchemaToModel(d *schema.ResourceData, id string) model.Rule {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	action := d.Get("action").(string)
	logged := d.Get("logged").(bool)
	tag := d.Get("log_label").(string)
	disabled := d.Get("disabled").(bool)
	sourcesExcluded := d.Get("sources_excluded").(bool)
	destinationsExcluded := d.Get("destinations_excluded").(bool)

	var ipProtocol *string
	ipp := d.Get("ip_version").(string)
	if ipp != "NONE" {
		ipProtocol = &ipp
	}
	direction := d.Get("direction").(string)
	notes := d.Get("notes").(string)
	seq := d.Get("sequence_number").(int)
	sequenceNumber := int64(seq)
	tagStructs := getPolicyTagsFromSet(d.Get("tag").(*schema.Set))

	resourceType := "Rule"
	return model.Rule{
		ResourceType:         &resourceType,
		Id:                   &id,
		DisplayName:          &displayName,
		Notes:                &notes,
		Description:          &description,
		Action:               &action,
		Logged:               &logged,
		Tag:                  &tag,
		Tags:                 tagStructs,
		Disabled:             &disabled,
		SourcesExcluded:      &sourcesExcluded,
		DestinationsExcluded: &destinationsExcluded,
		IpProtocol:           ipProtocol,
		Direction:            &direction,
		SourceGroups:         getPathListFromSchema(d, "source_groups"),
		DestinationGroups:    getPathListFromSchema(d, "destination_groups"),
		Services:             getPathListFromSchema(d, "services"),
		Scope:                getPathListFromSchema(d, "scope"),
		Profiles:             getPathListFromSchema(d, "profiles"),
		SequenceNumber:       &sequenceNumber,
	}
}

func resourceNsxtPolicySecurityPolicyRuleExistsPartial(d *schema.ResourceData, m interface{}, policyPath string) func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	// we need to take context from the parent rather than from resource context clause,
	// which does not exist for policy rule resource
	parentContext := getParentContext(d, m, policyPath)
	return func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicySecurityPolicyRuleExists(parentContext, id, policyPath, connector)
	}
}

func resourceNsxtPolicySecurityPolicyRuleExists(sessionContext utl.SessionContext, id string, policyPath string, connector client.Connector) (bool, error) {
	client := securitypolicies.NewRulesClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}

	domain := getDomainFromResourcePath(policyPath)
	policyID := getPolicyIDFromPath(policyPath)
	_, err := client.Get(domain, policyID, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving Security Policy Rule", err)
}

func resourceNsxtPolicySecurityPolicyRuleRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy Rule ID")
	}

	policyPath := d.Get("policy_path").(string)
	projectID := getProjectIDFromResourcePath(policyPath)
	domain := getDomainFromResourcePath(policyPath)
	policyID := getPolicyIDFromPath(policyPath)

	if err := setSecurityPolicyRuleContext(d, projectID); err != nil {
		return handleReadError(d, "SecurityPolicyRule", fmt.Sprintf("%s/%s", policyPath, id), err)
	}

	client := securitypolicies.NewRulesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	rule, err := client.Get(domain, policyID, id)
	if err != nil {
		return handleReadError(d, "SecurityPolicyRule", fmt.Sprintf("%s/%s", policyPath, id), err)
	}

	securityPolicyRuleModelToSchema(d, rule)
	return nil
}

func securityPolicyRuleModelToSchema(d *schema.ResourceData, rule model.Rule) {
	d.Set("display_name", rule.DisplayName)
	d.Set("description", rule.Description)
	d.Set("path", rule.Path)
	d.Set("notes", rule.Notes)
	d.Set("logged", rule.Logged)
	d.Set("log_label", rule.Tag)
	d.Set("action", rule.Action)
	d.Set("destinations_excluded", rule.DestinationsExcluded)
	d.Set("sources_excluded", rule.SourcesExcluded)
	if rule.IpProtocol == nil {
		d.Set("ip_version", "NONE")
	} else {
		d.Set("ip_version", rule.IpProtocol)
	}
	d.Set("direction", rule.Direction)
	d.Set("disabled", rule.Disabled)
	d.Set("revision", rule.Revision)
	setPathListInSchema(d, "source_groups", rule.SourceGroups)
	setPathListInSchema(d, "destination_groups", rule.DestinationGroups)
	setPathListInSchema(d, "profiles", rule.Profiles)
	setPathListInSchema(d, "services", rule.Services)
	setPathListInSchema(d, "scope", rule.Scope)
	d.Set("sequence_number", rule.SequenceNumber)
	d.Set("nsx_id", rule.Id)
	d.Set("rule_id", rule.RuleId)

	setPolicyTagsInSchema(d, rule.Tags)
}

func resourceNsxtPolicySecurityPolicyRuleUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy Rule ID")
	}

	policyPath := d.Get("policy_path").(string)
	log.Printf("[INFO] Updating Security Policy Rule with ID %s under policy %s", id, policyPath)
	domain := getDomainFromResourcePath(policyPath)
	policyID := getPolicyIDFromPath(policyPath)

	client := securitypolicies.NewRulesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	rule := securityPolicyRuleSchemaToModel(d, id)
	err := client.Patch(domain, policyID, id, rule)
	if err != nil {
		return handleUpdateError("SecurityPolicyRule", fmt.Sprintf("%s/%s", policyPath, id), err)
	}

	return resourceNsxtPolicySecurityPolicyRuleRead(d, m)
}

func resourceNsxtPolicySecurityPolicyRuleDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Get("nsx_id").(string)
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy Rule ID")
	}

	connector := getPolicyConnector(m)

	policyPath := d.Get("policy_path").(string)
	log.Printf("[INFO] Deleting Security Policy Rule with ID %s under policy %s", id, policyPath)
	domain := getDomainFromResourcePath(policyPath)
	policyID := getPolicyIDFromPath(policyPath)

	client := securitypolicies.NewRulesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	return client.Delete(domain, policyID, id)
}

func nsxtSecurityPolicyRuleImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}
	ruleIdx := strings.Index(importID, "rule")
	if ruleIdx <= 0 {
		return nil, fmt.Errorf("invalid path of Security Policy Rule to import")
	}
	d.Set("policy_path", importID[:ruleIdx-1])
	return []*schema.ResourceData{d}, nil
}
