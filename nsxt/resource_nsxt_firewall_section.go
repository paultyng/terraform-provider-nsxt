// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var firewallRuleIPProtocolValues = []string{"IPV4", "IPV6", "IPV4_IPV6"}
var firewallRuleActionValues = []string{"ALLOW", "DROP", "REJECT"}
var firewallRuleDirectionValues = []string{"IN", "OUT", "IN_OUT"}
var firewallSectionTypeValues = []string{"LAYER2", "LAYER3"}

func resourceNsxtFirewallSection() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtFirewallSectionCreate,
		Read:   resourceNsxtFirewallSectionRead,
		Update: resourceNsxtFirewallSectionUpdate,
		Delete: resourceNsxtFirewallSectionDelete,
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
			"is_default": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether a firewall section is default section or not",
				Computed:    true,
			},
			"section_type": {
				Type:         schema.TypeString,
				Description:  "Type of the rules which a section can contain. Only homogeneous sections are supported",
				Required:     true,
				ValidateFunc: validation.StringInSlice(firewallSectionTypeValues, false),
			},
			"stateful": {
				Type:        schema.TypeBool,
				Description: "Stateful or Stateless nature of firewall section is enforced on all rules inside the section",
				Required:    true,
				ForceNew:    true,
			},
			"applied_to": getResourceReferencesSetSchema(false, false, []string{"LogicalPort", "LogicalSwitch", "NSGroup", "LogicalRouter"}, "List of objects where the rules in this section will be enforced. This will take precedence over rule level appliedTo"),
			"insert_before": {
				Type:        schema.TypeString,
				Description: "Id of section that should come after this one",
				Optional:    true,
				ForceNew:    true,
			},
			"rule": getRulesSchema(),
		},
	}
}

func getRulesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of firewall rules in the section. Only homogeneous rules are supported",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Description: "ID of this resource",
					Computed:    true,
				},
				"revision": getRevisionSchema(),
				"description": {
					Type:        schema.TypeString,
					Description: "Description of this resource",
					Optional:    true,
				},
				"display_name": {
					Type:        schema.TypeString,
					Description: "Defaults to ID if not set",
					Optional:    true,
				},
				"action": {
					Type:         schema.TypeString,
					Description:  "Action enforced on the packets which matches the firewall rule",
					Required:     true,
					ValidateFunc: validation.StringInSlice(firewallRuleActionValues, false),
				},
				"applied_to":  getResourceReferencesSetSchema(false, false, []string{"LogicalPort", "LogicalSwitch", "NSGroup", "LogicalRouterPort"}, "List of objects where rule will be enforced. The section level field overrides this one. Null will be treated as any"),
				"destination": getResourceReferencesSetSchema(false, false, []string{"IPSet", "LogicalPort", "LogicalSwitch", "NSGroup", "MACSet"}, "List of the destinations. Null will be treated as any"),
				"destinations_excluded": {
					Type:        schema.TypeBool,
					Description: "When this boolean flag is set to true, the rule destinations will be negated",
					Optional:    true,
				},
				"direction": {
					Type:         schema.TypeString,
					Description:  "Rule direction in case of stateless firewall rules. This will only be considered if section level parameter is set to stateless. Default to IN_OUT if not specified",
					Optional:     true,
					ValidateFunc: validation.StringInSlice(firewallRuleDirectionValues, false),
				},
				"disabled": {
					Type:        schema.TypeBool,
					Description: "Flag to disable rule. Disabled will only be persisted but never provisioned/realized",
					Optional:    true,
				},
				"ip_protocol": {
					Type:         schema.TypeString,
					Description:  "Type of IP packet that should be matched while enforcing the rule (IPV4, IPV6, IPV4_IPV6)",
					Optional:     true,
					Default:      "IPV4_IPV6",
					ValidateFunc: validation.StringInSlice(firewallRuleIPProtocolValues, false),
				},
				"logged": {
					Type:        schema.TypeBool,
					Description: "Flag to enable packet logging. Default is disabled",
					Optional:    true,
				},
				"notes": {
					Type:        schema.TypeString,
					Description: "User notes specific to the rule",
					Optional:    true,
				},
				"rule_tag": {
					Type:        schema.TypeString,
					Description: "User level field which will be printed in CLI and packet logs",
					Optional:    true,
				},
				"source": getResourceReferencesSetSchema(false, false, []string{"IPSet", "LogicalPort", "LogicalSwitch", "NSGroup", "MACSet"}, "List of sources. Null will be treated as any"),
				"sources_excluded": {
					Type:        schema.TypeBool,
					Description: "When this boolean flag is set to true, the rule sources will be negated",
					Optional:    true,
				},
				"service": getResourceReferencesSetSchema(false, false, []string{"NSService", "NSServiceGroup"}, "List of the services. Null will be treated as any"),
			},
		},
	}
}

func returnServicesResourceReferences(services []manager.FirewallService) *schema.Set {
	var servicesList []interface{}
	for _, srv := range services {
		elem := make(map[string]interface{})
		elem["is_valid"] = srv.IsValid
		elem["target_display_name"] = srv.TargetDisplayName
		elem["target_id"] = srv.TargetId
		elem["target_type"] = srv.TargetType
		servicesList = append(servicesList, elem)
	}
	s := schema.NewSet(resourceReferenceHash, servicesList)
	return s
}

func setRulesInSchema(d *schema.ResourceData, rules []manager.FirewallRule) error {
	var rulesList []map[string]interface{}
	for _, rule := range rules {
		elem := make(map[string]interface{})
		elem["id"] = rule.Id
		elem["display_name"] = rule.DisplayName
		elem["description"] = rule.Description
		elem["rule_tag"] = rule.RuleTag
		elem["notes"] = rule.Notes
		elem["logged"] = rule.Logged
		elem["action"] = rule.Action
		elem["destinations_excluded"] = rule.DestinationsExcluded
		elem["sources_excluded"] = rule.SourcesExcluded
		elem["ip_protocol"] = rule.IpProtocol
		elem["disabled"] = rule.Disabled
		elem["revision"] = rule.Revision
		elem["direction"] = rule.Direction
		elem["source"] = returnResourceReferencesSet(rule.Sources)
		elem["destination"] = returnResourceReferencesSet(rule.Destinations)
		elem["service"] = returnServicesResourceReferences(rule.Services)
		elem["applied_to"] = returnResourceReferencesSet(rule.AppliedTos)

		rulesList = append(rulesList, elem)
	}
	err := d.Set("rule", rulesList)
	return err
}

func getServicesResourceReferences(services []interface{}) []manager.FirewallService {
	var servicesList []manager.FirewallService
	for _, srv := range services {
		data := srv.(map[string]interface{})
		elem := manager.FirewallService{
			IsValid:           data["is_valid"].(bool),
			TargetDisplayName: data["target_display_name"].(string),
			TargetId:          data["target_id"].(string),
			TargetType:        data["target_type"].(string),
		}
		servicesList = append(servicesList, elem)
	}
	return servicesList
}

func getRulesFromSchema(d *schema.ResourceData) []manager.FirewallRule {
	rules := d.Get("rule").([]interface{})
	var ruleList []manager.FirewallRule
	for _, rule := range rules {
		data := rule.(map[string]interface{})
		elem := manager.FirewallRule{
			DisplayName:          data["display_name"].(string),
			Id:                   data["id"].(string),
			RuleTag:              data["rule_tag"].(string),
			Notes:                data["notes"].(string),
			Description:          data["description"].(string),
			Action:               data["action"].(string),
			Logged:               data["logged"].(bool),
			Disabled:             data["disabled"].(bool),
			Revision:             int64(data["revision"].(int)),
			SourcesExcluded:      data["sources_excluded"].(bool),
			DestinationsExcluded: data["destinations_excluded"].(bool),
			IpProtocol:           data["ip_protocol"].(string),
			Direction:            data["direction"].(string),
			Sources:              getResourceReferences(data["source"].(*schema.Set).List()),
			Destinations:         getResourceReferences(data["destination"].(*schema.Set).List()),
			Services:             getServicesResourceReferences(data["service"].(*schema.Set).List()),
			AppliedTos:           getResourceReferences(data["applied_to"].(*schema.Set).List()),
		}

		ruleList = append(ruleList, elem)
	}
	return ruleList
}

func resourceNsxtFirewallSectionCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	rules := getRulesFromSchema(d)
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	appliedTos := getResourceReferencesFromSchemaSet(d, "applied_to")
	sectionType := d.Get("section_type").(string)
	stateful := d.Get("stateful").(bool)
	insertBefore := d.Get("insert_before")
	firewallSection := manager.FirewallSectionRuleList{
		FirewallSection: manager.FirewallSection{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			AppliedTos:  appliedTos,
			SectionType: sectionType,
			Stateful:    stateful,
		},
		Rules: rules,
	}

	localVarOptionals := make(map[string]interface{})
	if insertBefore != "" {
		localVarOptionals["operation"] = "insert_before"
		localVarOptionals["id"] = insertBefore
	}

	var resp *http.Response
	var err error
	if len(rules) == 0 {
		section := *firewallSection.GetFirewallSection()
		section, resp, err = nsxClient.ServicesApi.AddSection(nsxClient.Context, section, localVarOptionals)
		d.SetId(section.Id)
	} else {
		firewallSection, resp, err = nsxClient.ServicesApi.AddSectionWithRulesCreateWithRules(nsxClient.Context, firewallSection, localVarOptionals)
		d.SetId(firewallSection.Id)
	}

	if err != nil {
		return fmt.Errorf("Error during FirewallSection create with rules: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during FirewallSection create with rules: %v", resp.StatusCode)
	}

	return resourceNsxtFirewallSectionRead(d, m)
}

func resourceNsxtFirewallSectionRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	firewallSection, resp, err := nsxClient.ServicesApi.GetSectionWithRulesListWithRules(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during FirewallSection %s read: %v", id, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] FirewallSection %s not found", id)
		d.SetId("")
		return nil
	}

	d.Set("revision", firewallSection.Revision)
	d.Set("description", firewallSection.Description)
	d.Set("display_name", firewallSection.DisplayName)
	d.Set("is_default", firewallSection.IsDefault)
	d.Set("section_type", firewallSection.SectionType)
	d.Set("stateful", firewallSection.Stateful)
	setTagsInSchema(d, firewallSection.Tags)
	err = setRulesInSchema(d, firewallSection.Rules)
	if err != nil {
		return fmt.Errorf("Error during FirewallSection rules set in schema: %v", err)
	}

	// Getting the applied tos will require another api call (for NSX 2.1 or less)
	firewallSection2, resp, err := nsxClient.ServicesApi.GetSection(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] FirewallSection %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during FirewallSection %s read: %v", id, err)
	}
	err = setResourceReferencesInSchema(d, firewallSection2.AppliedTos, "applied_to")
	if err != nil {
		return fmt.Errorf("Error during FirewallSection AppliedTos set in schema: %v", err)
	}

	return nil
}

func resourceNsxtFirewallSectionUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	rules := getRulesFromSchema(d)
	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	appliedTos := getResourceReferencesFromSchemaSet(d, "applied_to")
	sectionType := d.Get("section_type").(string)
	stateful := d.Get("stateful").(bool)
	firewallSection := manager.FirewallSectionRuleList{
		FirewallSection: manager.FirewallSection{
			Revision:    revision,
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			AppliedTos:  appliedTos,
			SectionType: sectionType,
			Stateful:    stateful,
			Id:          id,
		},
		Rules: rules,
	}

	var resp *http.Response
	var err error
	if len(rules) == 0 || util.NsxVersionLower("2.2.0") {
		// Due to an NSX bug, the empty update should also be called to update ToS & tags fields
		section := *firewallSection.GetFirewallSection()
		// Update the section ignoring the rules
		_, resp, err = nsxClient.ServicesApi.UpdateSection(nsxClient.Context, id, section)

		if len(rules) == 0 {
			// Read the section, and delete all current rules from it
			currSection, resp2, err2 := nsxClient.ServicesApi.GetSectionWithRulesListWithRules(nsxClient.Context, id)
			if resp2.StatusCode == http.StatusNotFound {
				return fmt.Errorf("FirewallSection %s not found during update empty action", id)
			}
			if err2 != nil {
				return fmt.Errorf("Error during FirewallSection %s update empty: cannot read the section: %v", id, err2)
			}
			for _, rule := range currSection.Rules {
				_, err3 := nsxClient.ServicesApi.DeleteRule(nsxClient.Context, id, rule.Id)
				if err3 != nil {
					return fmt.Errorf("Error during FirewallSection %s update: failed to delete rule %s due to %v", id, rule.Id, err3)
				}
			}
		}
	}
	if len(rules) > 0 {
		// If we have rules - update the section with the rules
		_, resp, err = nsxClient.ServicesApi.UpdateSectionWithRulesUpdateWithRules(nsxClient.Context, id, firewallSection)
	}

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during FirewallSection %s update: %v", id, err)
	}

	return resourceNsxtFirewallSectionRead(d, m)
}

func resourceNsxtFirewallSectionDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id to delete")
	}

	localVarOptionals := make(map[string]interface{})
	localVarOptionals["cascade"] = true
	resp, err := nsxClient.ServicesApi.DeleteSection(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during FirewallSection %s delete: %v", id, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] FirewallSection %s not found", id)
		d.SetId("")
	}
	return nil
}
