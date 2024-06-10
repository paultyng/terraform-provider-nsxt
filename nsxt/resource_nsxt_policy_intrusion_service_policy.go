/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

// TODO: revisit with new SDK if constant is available
var policyIntrusionServiceRuleActionValues = []string{model.IdsRule_ACTION_DETECT, "DETECT_PREVENT"}

func resourceNsxtPolicyIntrusionServicePolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIntrusionServicePolicyCreate,
		Read:   resourceNsxtPolicyIntrusionServicePolicyRead,
		Update: resourceNsxtPolicyIntrusionServicePolicyUpdate,
		Delete: resourceNsxtPolicyIntrusionServicePolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDomainResourceImporter,
		},
		Schema: getPolicySecurityPolicySchema(true, true, true),
	}
}

func getIdsProfilesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "List of policy Paths for IDS Profiles",
		Required:    true,
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validatePolicyPath(),
		},
	}
}

func resourceNsxtPolicyIntrusionServicePolicyExistsInDomain(sessionContext utl.SessionContext, id string, domainName string, connector client.Connector) (bool, error) {
	client := domains.NewIntrusionServicePoliciesClient(sessionContext, connector)
	_, err := client.Get(domainName, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving Intrusion Service Policy", err)
}

func setPolicyIdsRulesInSchema(d *schema.ResourceData, rules []model.IdsRule) error {
	var rulesList []map[string]interface{}
	for _, rule := range rules {
		elem := make(map[string]interface{})
		elem["display_name"] = rule.DisplayName
		elem["description"] = rule.Description
		elem["notes"] = rule.Notes
		elem["logged"] = rule.Logged
		elem["log_label"] = rule.Tag
		elem["action"] = rule.Action
		elem["destinations_excluded"] = rule.DestinationsExcluded
		elem["sources_excluded"] = rule.SourcesExcluded
		elem["ip_version"] = rule.IpProtocol
		elem["direction"] = rule.Direction
		elem["disabled"] = rule.Disabled
		elem["revision"] = rule.Revision
		setPathListInMap(elem, "source_groups", rule.SourceGroups)
		setPathListInMap(elem, "destination_groups", rule.DestinationGroups)
		setPathListInMap(elem, "services", rule.Services)
		setPathListInMap(elem, "scope", rule.Scope)
		elem["sequence_number"] = rule.SequenceNumber
		elem["nsx_id"] = rule.Id
		setPathListInMap(elem, "ids_profiles", rule.IdsProfiles)

		var tagList []map[string]string
		for _, tag := range rule.Tags {
			tags := make(map[string]string)
			tags["scope"] = *tag.Scope
			tags["tag"] = *tag.Tag
			tagList = append(tagList, tags)
		}
		elem["tag"] = tagList

		rulesList = append(rulesList, elem)
	}

	return d.Set("rule", rulesList)
}

func getPolicyIdsRulesFromSchema(d *schema.ResourceData) []model.IdsRule {
	rules := d.Get("rule").([]interface{})
	var ruleList []model.IdsRule
	seq := 0
	for _, rule := range rules {
		data := rule.(map[string]interface{})
		displayName := data["display_name"].(string)
		description := data["description"].(string)
		action := data["action"].(string)
		logged := data["logged"].(bool)
		tag := data["log_label"].(string)
		disabled := data["disabled"].(bool)
		sourcesExcluded := data["sources_excluded"].(bool)
		destinationsExcluded := data["destinations_excluded"].(bool)
		ipProtocol := data["ip_version"].(string)
		direction := data["direction"].(string)
		notes := data["notes"].(string)
		sequenceNumber := int64(seq)
		tagStructs := getPolicyTagsFromSet(data["tag"].(*schema.Set))

		// Use a different random Id each time, otherwise Update requires revision
		// to be set for existing rules, and NOT be set for new rules
		id := newUUID()

		resourceType := "IdsRule"
		elem := model.IdsRule{
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
			IpProtocol:           &ipProtocol,
			Direction:            &direction,
			SourceGroups:         getPathListFromMap(data, "source_groups"),
			DestinationGroups:    getPathListFromMap(data, "destination_groups"),
			Services:             getPathListFromMap(data, "services"),
			Scope:                getPathListFromMap(data, "scope"),
			SequenceNumber:       &sequenceNumber,
			IdsProfiles:          getPathListFromMap(data, "ids_profiles"),
		}

		ruleList = append(ruleList, elem)
		seq = seq + 1
	}

	return ruleList
}

func resourceNsxtPolicyIntrusionServicePolicyExistsPartial(sessionContext utl.SessionContext, domainName string) func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicyIntrusionServicePolicyExistsInDomain(sessionContext, id, domainName, connector)
	}
}

func createPolicyChildIdsRule(ruleID string, rule model.IdsRule, shouldDelete bool) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	childRule := model.ChildIdsRule{
		ResourceType:    "ChildIdsRule",
		Id:              &ruleID,
		IdsRule:         &rule,
		MarkedForDelete: &shouldDelete,
	}

	dataValue, errors := converter.ConvertToVapi(childRule, model.ChildIdsRuleBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return dataValue.(*data.StructValue), nil
}

func createChildDomainWithIdsSecurityPolicy(domain string, policyID string, policy model.IdsSecurityPolicy) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	childPolicy := model.ChildIdsSecurityPolicy{
		Id:                &policyID,
		ResourceType:      "ChildIdsSecurityPolicy",
		IdsSecurityPolicy: &policy,
	}

	dataValue, errors := converter.ConvertToVapi(childPolicy, model.ChildIdsSecurityPolicyBindingType())
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

func updateIdsSecurityPolicy(id string, d *schema.ResourceData, m interface{}) error {

	domain := d.Get("domain").(string)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	comments := d.Get("comments").(string)
	locked := d.Get("locked").(bool)
	sequenceNumber := int64(d.Get("sequence_number").(int))
	stateful := d.Get("stateful").(bool)
	resourceType := "IdsSecurityPolicy"

	obj := model.IdsSecurityPolicy{
		Id:             &id,
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		Comments:       &comments,
		Locked:         &locked,
		SequenceNumber: &sequenceNumber,
		Stateful:       &stateful,
		ResourceType:   &resourceType,
	}

	var childRules []*data.StructValue
	if d.HasChange("rule") {
		oldRules, _ := d.GetChange("rule")
		rules := getPolicyIdsRulesFromSchema(d)

		existingRules := make(map[string]bool)
		for _, rule := range rules {
			ruleID := newUUID()
			if rule.Id != nil {
				ruleID = *rule.Id
				existingRules[ruleID] = true
			} else {
				rule.Id = &ruleID
			}

			childRule, err := createPolicyChildIdsRule(ruleID, rule, false)
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
				resourceType := "IdsRule"
				rule := model.IdsRule{
					Id:           &oldRuleID,
					ResourceType: &resourceType,
				}

				childRule, err := createPolicyChildIdsRule(oldRuleID, rule, true)
				if err != nil {
					return err
				}
				log.Printf("[DEBUG]: Deleting child rule with id %s", oldRuleID)
				childRules = append(childRules, childRule)

			}
		}
	}

	log.Printf("[DEBUG]: Updating IDS policy %s with %d child rules", id, len(childRules))
	if len(childRules) > 0 {
		obj.Children = childRules
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	return idsPolicyInfraPatch(context, obj, domain, m)
}

func idsPolicyInfraPatch(context utl.SessionContext, policy model.IdsSecurityPolicy, domain string, m interface{}) error {
	childDomain, err := createChildDomainWithIdsSecurityPolicy(domain, *policy.Id, policy)
	if err != nil {
		return fmt.Errorf("Failed to create H-API for Ids Policy: %s", err)
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

func resourceNsxtPolicyIntrusionServicePolicyCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyIntrusionServicePolicyExistsPartial(context, d.Get("domain").(string)))
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating Intrusion Service Policy with ID %s", id)
	err = updateIdsSecurityPolicy(id, d, m)

	if err != nil {
		return handleCreateError("Intrusion Service Policy", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIntrusionServicePolicyRead(d, m)
}

func resourceNsxtPolicyIntrusionServicePolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	domainName := d.Get("domain").(string)
	if id == "" {
		return fmt.Errorf("Error obtaining Intrusion Service Policy id")
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewIntrusionServicePoliciesClient(context, connector)
	obj, err := client.Get(domainName, id)
	if err != nil {
		return handleReadError(d, "Intrusion Service Policy", id, err)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("domain", getDomainFromResourcePath(*obj.Path))
	d.Set("comments", obj.Comments)
	d.Set("locked", obj.Locked)
	d.Set("sequence_number", obj.SequenceNumber)
	d.Set("stateful", obj.Stateful)
	d.Set("revision", obj.Revision)
	return setPolicyIdsRulesInSchema(d, obj.Rules)
}

func resourceNsxtPolicyIntrusionServicePolicyUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Intrusion Service Policy id")
	}

	log.Printf("[INFO] Updating Intrusion Service Policy with ID %s", id)
	err := updateIdsSecurityPolicy(id, d, m)

	if err != nil {
		return handleUpdateError("Intrusion Service Policy", id, err)
	}

	return resourceNsxtPolicyIntrusionServicePolicyRead(d, m)
}

func resourceNsxtPolicyIntrusionServicePolicyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Intrusion Service Policy id")
	}

	connector := getPolicyConnector(m)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewIntrusionServicePoliciesClient(context, connector)
	err = client.Delete(d.Get("domain").(string), id)

	if err != nil {
		return handleDeleteError("Intrusion Service Policy", id, err)
	}

	return nil
}
