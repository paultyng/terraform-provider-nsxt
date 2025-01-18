// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var defaultDomain = "default"
var defaultSite = "default"
var securityPolicyCategoryValues = []string{"Ethernet", "Emergency", "Infrastructure", "Environment", "Application"}
var securityPolicyDirectionValues = []string{model.Rule_DIRECTION_IN, model.Rule_DIRECTION_OUT, model.Rule_DIRECTION_IN_OUT}
var securityPolicyIPProtocolValues = []string{"NONE", model.Rule_IP_PROTOCOL_IPV4, model.Rule_IP_PROTOCOL_IPV6, model.Rule_IP_PROTOCOL_IPV4_IPV6}

// TODO: change last string to sdk constant when available
var securityPolicyActionValues = []string{model.Rule_ACTION_ALLOW, model.Rule_ACTION_DROP, model.Rule_ACTION_REJECT, "JUMP_TO_APPLICATION"}
var gatewayPolicyCategoryWritableValues = []string{"Emergency", "SharedPreRules", "LocalGatewayRules", "Default"}
var policyFailOverModeValues = []string{model.Tier1_FAILOVER_MODE_PREEMPTIVE, model.Tier1_FAILOVER_MODE_NON_PREEMPTIVE}
var failOverModeDefaultPolicyT0Value = model.Tier0_FAILOVER_MODE_NON_PREEMPTIVE
var defaultPolicyLocaleServiceID = "default"

var mpObjectResourceDeprecationMessage = "Please use corresponding policy resource instead"
var mpObjectDataSourceDeprecationMessage = "Please use corresponding policy data source instead"

func getNsxIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "NSX ID for this resource",
		Optional:     true,
		Computed:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringLenBetween(1, 1024),
	}
}

func getFlexNsxIDSchema(readOnly bool) *schema.Schema {
	s := schema.Schema{
		Type:        schema.TypeString,
		Description: "NSX ID for this resource",
		Optional:    !readOnly,
		Computed:    true,
	}
	if !readOnly {
		s.ValidateFunc = validation.StringLenBetween(1, 1024)
	}
	return &s
}

func getComputedNsxIDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "NSX ID for this resource",
		Computed:    true,
	}
}

func getPathSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "Policy path for this resource",
		Computed:    true,
	}
}

func getDisplayNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Display name for this resource",
		Required:     true,
		ValidateFunc: validation.StringLenBetween(1, 255),
	}
}

func getOptionalDisplayNameSchema(isComputed bool) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Display name for this resource",
		Optional:     true,
		Computed:     isComputed,
		ValidateFunc: validation.StringLenBetween(0, 255),
	}
}

func getDescriptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Description for this resource",
		Optional:     true,
		ValidateFunc: validation.StringLenBetween(0, 1024),
	}
}

func getComputedDescriptionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "Description for this resource",
		Optional:    true,
		Computed:    true,
	}
}

func getComputedDisplayNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "Display name for this resource",
		Computed:    true,
	}
}

func getDataSourceStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: description,
		Optional:    true,
		Computed:    true,
	}
}

func getRequiredStringSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: description,
		Required:    true,
	}
}

func getDomainNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The domain name to use for resources. If not specified 'default' is used",
		Optional:    true,
		Default:     defaultDomain,
		ForceNew:    true,
	}
}

func getDataSourceDomainNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeString,
		Description: "The domain name. If not specified 'default' is used",
		Optional:    true,
		Default:     defaultDomain,
	}
}

func getPolicyGatewayPathSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "The NSX-T Policy path to the Tier0 or Tier1 Gateway for this resource",
		Required:     true,
		ValidateFunc: validatePolicyPath(),
		ForceNew:     true,
	}
}

func getPolicyRuleActionSchema(isIds bool) *schema.Schema {
	validationSlice := securityPolicyActionValues
	defaultValue := model.Rule_ACTION_ALLOW
	if isIds {
		validationSlice = policyIntrusionServiceRuleActionValues
		defaultValue = model.IdsRule_ACTION_DETECT
	}
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  "Action",
		Optional:     true,
		ValidateFunc: validation.StringInSlice(validationSlice, false),
		Default:      defaultValue,
	}
}

func getSecurityPolicyAndGatewayRulesSchema(scopeRequired bool, isIds bool, nsxIDReadOnly bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of rules in the section",
		Optional:    true,
		MaxItems:    1000,
		Elem: &schema.Resource{
			Schema: getSecurityPolicyAndGatewayRuleSchema(scopeRequired, isIds, nsxIDReadOnly, false),
		},
	}
}

func getSecurityPolicyAndGatewayRuleSchema(scopeRequired bool, isIds bool, nsxIDReadOnly bool, separated bool) map[string]*schema.Schema {
	ruleSchema := map[string]*schema.Schema{
		"nsx_id":       getFlexNsxIDSchema(nsxIDReadOnly),
		"display_name": getDisplayNameSchema(),
		"description":  getDescriptionSchema(),
		"path":         getPathSchema(),
		"revision":     getRevisionSchema(),
		"destination_groups": {
			Type:        schema.TypeSet,
			Description: "List of destination groups",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicySourceDestinationGroups(),
			},
			Optional: true,
		},
		"destinations_excluded": {
			Type:        schema.TypeBool,
			Description: "Negation of destination groups",
			Optional:    true,
			Default:     false,
		},
		"direction": {
			Type:         schema.TypeString,
			Description:  "Traffic direction",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(securityPolicyDirectionValues, false),
			Default:      model.Rule_DIRECTION_IN_OUT,
		},
		"disabled": {
			Type:        schema.TypeBool,
			Description: "Flag to disable the rule",
			Optional:    true,
			Default:     false,
		},
		"ip_version": {
			Type:         schema.TypeString,
			Description:  "IP version",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(securityPolicyIPProtocolValues, false),
			Default:      model.Rule_IP_PROTOCOL_IPV4_IPV6,
		},
		"logged": {
			Type:        schema.TypeBool,
			Description: "Flag to enable packet logging",
			Optional:    true,
			Default:     false,
		},
		"notes": {
			Type:        schema.TypeString,
			Description: "Text for additional notes on changes",
			Optional:    true,
		},
		"profiles": {
			Type:        schema.TypeSet,
			Description: "List of profiles",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicyPath(),
			},
			Optional: true,
		},
		"rule_id": {
			Type:        schema.TypeInt,
			Description: "Unique positive number that is assigned by the system and is useful for debugging",
			Computed:    true,
		},
		"scope": {
			Type:        schema.TypeSet,
			Description: "List of policy paths where the rule is applied",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicyPath(),
			},
			Optional: !scopeRequired,
			Required: scopeRequired,
		},
		"services": {
			Type:        schema.TypeSet,
			Description: "List of services to match",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicyPath(),
			},
			Optional: true,
		},
		"source_groups": {
			Type:        schema.TypeSet,
			Description: "List of source groups",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicySourceDestinationGroups(),
			},
			Optional: true,
		},
		"sources_excluded": {
			Type:        schema.TypeBool,
			Description: "Negation of source groups",
			Optional:    true,
			Default:     false,
		},
		"tag": getTagsSchema(),
		"log_label": {
			Type:        schema.TypeString,
			Description: "Additional information (string) which will be propagated to the rule syslog",
			Optional:    true,
		},
		"action": getPolicyRuleActionSchema(isIds),
	}
	if isIds {
		ruleSchema["ids_profiles"] = getIdsProfilesSchema()
	}
	if separated {
		ruleSchema["policy_path"] = getPolicyPathSchema(true, true, "Security Policy path")
		ruleSchema["sequence_number"] = &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Sequence number of the this rule",
			Required:    true,
		}
		// Using computed context here, because context is required for consistency and
		// if it's not provided it can be derived from policy_path.
		ruleSchema["context"] = getContextSchema(false, true, false)
	} else {
		ruleSchema["sequence_number"] = &schema.Schema{
			Type:        schema.TypeInt,
			Description: "Sequence number of the this rule",
			Optional:    true,
			Computed:    true,
		}
	}
	return ruleSchema
}

func getPolicyGatewayPolicySchema(isVPC bool) map[string]*schema.Schema {
	secPolicy := getPolicySecurityPolicySchema(false, true, true, isVPC)
	// GW Policies don't support scope
	delete(secPolicy, "scope")
	if !isVPC {
		secPolicy["category"].ValidateFunc = validation.StringInSlice(gatewayPolicyCategoryWritableValues, false)
	}
	// GW Policy rules require scope to be set
	secPolicy["rule"] = getSecurityPolicyAndGatewayRulesSchema(!isVPC, false, true)
	return secPolicy
}

func getPolicySecurityPolicySchema(isIds, withContext, withRule, isVPC bool) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		"nsx_id":       getNsxIDSchema(),
		"path":         getPathSchema(),
		"display_name": getDisplayNameSchema(),
		"description":  getDescriptionSchema(),
		"revision":     getRevisionSchema(),
		"tag":          getTagsSchema(),
		"context":      getContextSchema(isVPC, false, isVPC),
		"domain":       getDomainNameSchema(),
		"category": {
			Type:         schema.TypeString,
			Description:  "Category",
			ValidateFunc: validation.StringInSlice(securityPolicyCategoryValues, false),
			Required:     true,
			ForceNew:     true,
		},
		"comments": {
			Type:        schema.TypeString,
			Description: "Comments for security policy lock/unlock",
			Optional:    true,
		},
		"locked": {
			Type:        schema.TypeBool,
			Description: "Indicates whether a security policy should be locked. If locked by a user, no other user would be able to modify this policy",
			Optional:    true,
			Default:     false,
		},
		"scope": {
			Type:        schema.TypeSet,
			Description: "The list of group paths where the rules in this policy will get applied",
			Optional:    true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validatePolicyPath(),
			},
		},
		// TODO - verify this is relevant not only across domains
		"sequence_number": {
			Type:        schema.TypeInt,
			Description: "This field is used to resolve conflicts between security policies across domains",
			Optional:    true,
			Default:     0,
		},
		"stateful": {
			Type:        schema.TypeBool,
			Description: "When it is stateful, the state of the network connects are tracked and a stateful packet inspection is performed",
			Optional:    true,
			Default:     true,
		},
		"tcp_strict": {
			Type:        schema.TypeBool,
			Description: "Ensures that a 3 way TCP handshake is done before the data packets are sent",
			Optional:    true,
			Computed:    true,
		},
		"rule": getSecurityPolicyAndGatewayRulesSchema(false, isIds, true),
	}

	if isIds {
		delete(result, "category")
		delete(result, "scope")
		delete(result, "tcp_strict")
	}

	if !withContext {
		delete(result, "context")
	}

	if !withRule {
		delete(result, "rule")
	}
	if isVPC {
		delete(result, "domain")
		delete(result, "category")
	}
	return result
}

func getIgnoreTagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"scopes": {
					Type:        schema.TypeList,
					Description: "List of scopes to ignore",
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"detected": {
					Type:        schema.TypeSet,
					Description: "Tags matching scopes to ignore",
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"scope": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"tag": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
			},
		},
	}
}

func setPolicyRulesInSchema(d *schema.ResourceData, rules []model.Rule) error {
	var rulesList []map[string]interface{}
	for _, rule := range rules {
		elem := make(map[string]interface{})
		elem["display_name"] = rule.DisplayName
		elem["description"] = rule.Description
		elem["path"] = rule.Path
		elem["notes"] = rule.Notes
		elem["logged"] = rule.Logged
		elem["log_label"] = rule.Tag
		elem["action"] = rule.Action
		elem["destinations_excluded"] = rule.DestinationsExcluded
		elem["sources_excluded"] = rule.SourcesExcluded
		if rule.IpProtocol == nil {
			elem["ip_version"] = "NONE"
		} else {
			elem["ip_version"] = rule.IpProtocol
		}
		elem["direction"] = rule.Direction
		elem["disabled"] = rule.Disabled
		elem["revision"] = rule.Revision
		setPathListInMap(elem, "source_groups", rule.SourceGroups)
		setPathListInMap(elem, "destination_groups", rule.DestinationGroups)
		setPathListInMap(elem, "profiles", rule.Profiles)
		setPathListInMap(elem, "services", rule.Services)
		setPathListInMap(elem, "scope", rule.Scope)
		elem["sequence_number"] = rule.SequenceNumber
		elem["nsx_id"] = rule.Id
		elem["rule_id"] = rule.RuleId

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

func validatePolicyRuleSequence(d *schema.ResourceData) error {
	rules := d.Get("rule").([]interface{})
	latestNum := int64(0)
	for _, rule := range rules {
		data := rule.(map[string]interface{})
		sequenceNumber := int64(data["sequence_number"].(int))
		displayName := data["display_name"].(string)
		if sequenceNumber > 0 && sequenceNumber <= latestNum {
			return fmt.Errorf("when sequence_number is specified in a rule, it must be consistent with rule order. To avoid confusion, it is recommended to either specify sequence numbers in all rules, or none. Error detected with rule %s: %v <= %v", displayName, sequenceNumber, latestNum)
		}

		if sequenceNumber == 0 {
			// Sequence number is unspecified, leave space for this rule to validate potential next rules
			latestNum++
		} else {
			latestNum = sequenceNumber
		}
	}
	return nil
}

func getPolicyRulesFromSchema(d *schema.ResourceData) []model.Rule {
	rules := d.Get("rule").([]interface{})
	var ruleList []model.Rule
	lastSequence := int64(0)
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

		var ipProtocol *string
		ipp := data["ip_version"].(string)
		if ipp != "NONE" {
			ipProtocol = &ipp
		}
		direction := data["direction"].(string)
		notes := data["notes"].(string)
		seq := data["sequence_number"].(int)
		sequenceNumber := int64(seq)
		tagStructs := getPolicyTagsFromSet(data["tag"].(*schema.Set))

		id := newUUID()
		nsxID := data["nsx_id"].(string)
		if nsxID != "" {
			id = nsxID
		}

		resourceType := "Rule"
		if sequenceNumber == 0 || sequenceNumber <= lastSequence {
			// We overwrite sequence number in case its not specified,
			// or out of order, which might be due to provider upgrade
			// or bad user configuration
			if sequenceNumber <= lastSequence {
				log.Printf("[WARNING] Sequence_number %v for rule %s is out of order - overriding with sequence number %v", sequenceNumber, displayName, lastSequence+1)
			}
			sequenceNumber = lastSequence + 1
		}
		lastSequence = sequenceNumber

		elem := model.Rule{
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
			SourceGroups:         getPathListFromMap(data, "source_groups"),
			DestinationGroups:    getPathListFromMap(data, "destination_groups"),
			Services:             getPathListFromMap(data, "services"),
			Scope:                getPathListFromMap(data, "scope"),
			Profiles:             getPathListFromMap(data, "profiles"),
			SequenceNumber:       &sequenceNumber,
		}

		ruleList = append(ruleList, elem)
	}

	return ruleList
}

func getDataSourceDisplayNameSchema() *schema.Schema {
	return getDataSourceStringSchema("Display name of this resource")
}

func getDataSourceExtendedDisplayNameSchema() *schema.Schema {
	return &schema.Schema{
		Type:          schema.TypeString,
		Description:   "Display name of this resource",
		ConflictsWith: []string{"id"},
		Optional:      true,
		Computed:      true,
	}
}

func getDataSourceDescriptionSchema() *schema.Schema {
	return getDataSourceStringSchema("Description for this resource")
}

func getDataSourceIDSchema() *schema.Schema {
	return getDataSourceStringSchema("Unique ID of this resource")
}

func parseGatewayPolicyPath(gwPath string) (bool, string) {
	// sample path looks like "/infra/tier-0s/mytier0gw"
	// Or "/global-infra/tier-0s/mytier0gw" in Global Manager
	isT0 := true
	segs := strings.Split(gwPath, "/")
	if len(segs) < 3 {
		return false, ""
	}
	if segs[len(segs)-2] != "tier-0s" {
		isT0 = false
	}
	return isT0, segs[len(segs)-1]
}

func parseLocaleServicePolicyPath(path string) (bool, string, string, error) {
	segs := strings.Split(path, "/")
	// Path should be like /infra/tier-0s/aaa/locale-services/default
	segCount := len(segs)
	if (segCount < 6) || (segs[segCount-2] != "locale-services") {
		// error - this is not a segment path
		return false, "", "", fmt.Errorf("Invalid Locale service path %s", path)
	}

	localeServiceID := segs[segCount-1]
	gwPath := strings.Join(segs[:4], "/")

	isT0, gwID := parseGatewayPolicyPath(gwPath)
	return isT0, gwID, localeServiceID, nil
}

func getPolicyPathSchemaSimple() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validatePolicyPath(),
	}
}

func getPolicyPathSchema(isRequired bool, forceNew bool, description string) *schema.Schema {
	attrSchema := getPolicyPathSchemaSimple()
	attrSchema.Description = description
	attrSchema.ForceNew = forceNew
	attrSchema.Required = isRequired
	attrSchema.Optional = !isRequired
	return attrSchema
}

func getPolicyPathSchemaExtended(isRequired bool, forceNew bool, description string, deprecation string, conflictsWith []string) *schema.Schema {
	attrSchema := getPolicyPathSchema(isRequired, forceNew, description)
	attrSchema.Deprecated = deprecation
	attrSchema.ConflictsWith = conflictsWith
	return attrSchema
}

func getComputedPolicyPathSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Description:  description,
		Optional:     true,
		Computed:     true,
		ValidateFunc: validatePolicyPath(),
	}
}

func getElemPolicyPathSchemaWithFlags(isOptional, isComputed, isRequired bool) *schema.Schema {
	s := schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validatePolicyPath(),
	}
	if isOptional {
		s.Optional = true
	}
	if isComputed {
		s.Computed = true
	}
	if isRequired {
		s.Required = true
	}
	return &s
}

func getElemPolicyPathSchema() *schema.Schema {
	return getElemPolicyPathSchemaWithFlags(false, false, false)
}

func getAllocationRangeListSchema(required bool, description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: description,
		Required:    required,
		Optional:    !required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"start": {
					Type:         schema.TypeString,
					Description:  "The start IP Address for the range",
					Required:     true,
					ValidateFunc: validateSingleIP(),
				},
				"end": {
					Type:         schema.TypeString,
					Description:  "The end IP Address for the range",
					Required:     true,
					ValidateFunc: validateSingleIP(),
				},
			},
		},
	}
}

func localManagerOnlyError() error {
	return fmt.Errorf("This configuration is not supported with NSX Global Manager")
}

func globalManagerOnlyError() error {
	return fmt.Errorf("This configuration is only supported with NSX Global Manager. To mark your endpoint as Global Manager, please set 'global_manager' flag to 'true' in the provider")
}

func attributeRequiredGlobalManagerError(attribute string, resource string) error {
	return fmt.Errorf("%s requires %s configuration for NSX Global Manager", resource, attribute)
}

func buildQueryStringFromMap(query map[string]string) string {
	if query == nil {
		return ""
	}
	keyValues := make([]string, 0, len(query))
	for key, value := range query {
		value = strings.Replace(value, "/", "\\/", -1)
		keyValue := strings.Join([]string{key, value}, ":")
		keyValues = append(keyValues, keyValue)
	}
	return strings.Join(keyValues, " AND ")
}

func getSitePathFromEdgePath(edgePath string) string {
	// Sample Edge cluster path looks like:
	//"/global-infra/sites/<site-id>/enforcement-points/default/edge-clusters/<edge-cluster-id>"
	pathList := strings.Split(edgePath, "/")[:4]
	return strings.Join(pathList, "/")
}

func getGatewayPathFromLocaleServicesPath(localeServicesPath string) string {
	// Sample Locale services path looks like:
	// "/infra/tier-0s/<tier0-id>/locale-services/<locale-services-id>"
	pathList := strings.Split(localeServicesPath, "/")[:4]
	return strings.Join(pathList, "/")
}
