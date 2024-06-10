/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var conditionKeyValues = []string{
	model.Condition_KEY_TAG,
	model.Condition_KEY_NAME,
	model.Condition_KEY_OSNAME,
	model.Condition_KEY_COMPUTERNAME,
	model.Condition_KEY_NODETYPE,
	model.Condition_KEY_GROUPTYPE,
	model.Condition_KEY_ALL,
	model.Condition_KEY_IPADDRESS,
	model.Condition_KEY_PODCIDR,
}

var conditionMemberTypeValues = []string{
	model.Condition_MEMBER_TYPE_IPSET,
	model.Condition_MEMBER_TYPE_VIRTUALMACHINE,
	model.Condition_MEMBER_TYPE_LOGICALPORT,
	model.Condition_MEMBER_TYPE_LOGICALSWITCH,
	model.Condition_MEMBER_TYPE_SEGMENT,
	model.Condition_MEMBER_TYPE_SEGMENTPORT,
	model.Condition_MEMBER_TYPE_POD,
	model.Condition_MEMBER_TYPE_SERVICE,
	model.Condition_MEMBER_TYPE_NAMESPACE,
	model.Condition_MEMBER_TYPE_TRANSPORTNODE,
	model.Condition_MEMBER_TYPE_GROUP,
	model.Condition_MEMBER_TYPE_DVPG,
	model.Condition_MEMBER_TYPE_DVPORT,
	model.Condition_MEMBER_TYPE_IPADDRESS,
	model.Condition_MEMBER_TYPE_KUBERNETESCLUSTER,
	model.Condition_MEMBER_TYPE_KUBERNETESNAMESPACE,
	model.Condition_MEMBER_TYPE_ANTREAEGRESS,
	model.Condition_MEMBER_TYPE_ANTREAIPPOOL,
	model.Condition_MEMBER_TYPE_KUBERNETESINGRESS,
	model.Condition_MEMBER_TYPE_KUBERNETESGATEWAY,
	model.Condition_MEMBER_TYPE_KUBERNETESSERVICE,
	model.Condition_MEMBER_TYPE_KUBERNETESNODE,
}

var conditionOperatorValues = []string{
	model.Condition_OPERATOR_EQUALS,
	model.Condition_OPERATOR_CONTAINS,
	model.Condition_OPERATOR_STARTSWITH,
	model.Condition_OPERATOR_ENDSWITH,
	model.Condition_OPERATOR_NOTEQUALS,
	model.Condition_OPERATOR_NOTIN,
	model.Condition_OPERATOR_MATCHES,
	model.Condition_OPERATOR_IN,
}
var conjunctionOperatorValues = []string{
	model.ConjunctionOperator_CONJUNCTION_OPERATOR_OR,
	model.ConjunctionOperator_CONJUNCTION_OPERATOR_AND,
}

var externalMemberTypeValues = []string{
	model.ExternalIDExpression_MEMBER_TYPE_VIRTUALMACHINE,
	model.ExternalIDExpression_MEMBER_TYPE_VIRTUALNETWORKINTERFACE,
	model.ExternalIDExpression_MEMBER_TYPE_CLOUDNATIVESERVICEINSTANCE,
	model.ExternalIDExpression_MEMBER_TYPE_PHYSICALSERVER,
}

var groupTypeValues = []string{
	model.Group_GROUP_TYPE_IPADDRESS,
	model.Group_GROUP_TYPE_ANTREA,
}

func resourceNsxtPolicyGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGroupCreate,
		Read:   resourceNsxtPolicyGroupRead,
		Update: resourceNsxtPolicyGroupUpdate,
		Delete: resourceNsxtPolicyGroupDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDomainResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"domain":       getDomainNameSchema(),
			"group_type": {
				Type:         schema.TypeString,
				Description:  "Indicates the group type",
				ValidateFunc: validation.StringInSlice(groupTypeValues, false),
				Optional:     true,
			},
			"criteria": {
				Type:        schema.TypeList,
				Description: "Criteria to determine Group membership",
				Elem:        getCriteriaSetSchema(),
				Optional:    true,
			},
			"conjunction": {
				Type:        schema.TypeList,
				Description: "A conjunction applied to 2 sets of criteria.",
				Elem:        getConjunctionSchema(),
				Optional:    true,
			},
			"extended_criteria": {
				Type:        schema.TypeList,
				Description: "Extended criteria to determine group membership. extended_criteria is implicitly \"AND\" with criteria",
				Elem:        getExtendedCriteriaSetSchema(),
				Optional:    true,
				MaxItems:    1,
			},
		},
	}
}

func getIPAddressExpressionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ip_addresses": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of single IP addresses, IP address ranges or Subnets. Cannot mix IPv4 and IPv6 in a single list",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateCidrOrIPOrRange(),
				},
			},
		},
	}
}

func getMACAddressExpressionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"mac_addresses": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of Mac Addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsMACAddress,
				},
			},
		},
	}
}

func getExternalIDExpressionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"external_ids": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of external IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"member_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "External ID member type, default to virtual machine if not specified",
				ValidateFunc: validation.StringInSlice(externalMemberTypeValues, false),
				Default:      model.ExternalIDExpression_MEMBER_TYPE_VIRTUALMACHINE,
			},
		},
	}
}

func getPathExpressionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"member_paths": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "List of policy paths of direct group members",
				Elem:        getElemPolicyPathSchema(),
			},
		},
	}
}

func getConditionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The resource key attribute to apply the condition to.",
				ValidateFunc: validation.StringInSlice(conditionKeyValues, false),
			},
			"member_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The NSX member to apply the condition to. Can be one of; IPSet, LogicalPort, LogicalSwitch, Segment, SegmentPort or VirtualMachine",
				ValidateFunc: validation.StringInSlice(conditionMemberTypeValues, false),
			},
			"operator": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The operator to use for the condition. Can be one of; CONTAINS, ENDSWITH, EQUALS, NOTEQUALS or STARTSWITH",
				ValidateFunc: validation.StringInSlice(conditionOperatorValues, false),
			},
			"value": {
				Type:        schema.TypeString,
				Description: "The value to check for in the condition",
				Required:    true,
			},
		},
	}
}

func getConjunctionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"operator": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The conjunction operator; either OR or AND",
				ValidateFunc: validation.StringInSlice(conjunctionOperatorValues, false),
			},
		},
	}
}

func getIdentityGroupSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"distinguished_name": {
				Type:        schema.TypeString,
				Description: "LDAP distinguished name",
				Optional:    true,
			},
			"domain_base_distinguished_name": {
				Type:        schema.TypeString,
				Description: "Identity (Directory) domain base distinguished name",
				Optional:    true,
			},
			"sid": {
				Type:        schema.TypeString,
				Description: "Identity (Directory) Group SID (security identifier)",
				Optional:    true,
			},
		},
	}
}

func getCriteriaSetSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ipaddress_expression": {
				Type:        schema.TypeList,
				Description: "An IP Address expression specifying IP Address members in the Group",
				Elem:        getIPAddressExpressionSchema(),
				Optional:    true,
				MaxItems:    1,
			},
			"condition": {
				Type:        schema.TypeList,
				Description: "A Condition querying resources for membership in the Group",
				Elem:        getConditionSchema(),
				Optional:    true,
				MaxItems:    5,
			},
			"path_expression": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of object paths for members in the group",
				Elem:        getPathExpressionSchema(),
				MaxItems:    1,
			},
			"macaddress_expression": {
				Type:        schema.TypeList,
				Description: "MAC address expression specifying MAC Address members in the Group",
				Elem:        getMACAddressExpressionSchema(),
				Optional:    true,
				MaxItems:    1,
			},
			"external_id_expression": {
				Type:        schema.TypeList,
				Description: "External ID expression specifying additional members in the Group",
				Elem:        getExternalIDExpressionSchema(),
				Optional:    true,
			},
		},
	}
}

func getExtendedCriteriaSetSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"identity_group": {
				Type:        schema.TypeSet,
				Description: "Identity Group expression",
				Elem:        getIdentityGroupSchema(),
				Optional:    true,
			},
		},
	}
}

func resourceNsxtPolicyGroupExistsInDomain(sessionContext utl.SessionContext, id string, domain string, connector client.Connector) (bool, error) {
	client := domains.NewGroupsClient(sessionContext, connector)
	_, err := client.Get(domain, id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving Group", err)

}

func resourceNsxtPolicyGroupExistsInDomainPartial(domain string) func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicyGroupExistsInDomain(sessionContext, id, domain, connector)
	}
}

func validateNestedGroupConditions(conditions []interface{}) (string, error) {
	memberType := ""
	for _, cond := range conditions {
		condMap := cond.(map[string]interface{})
		condMemberType := condMap["member_type"].(string)
		if memberType != "" && condMemberType != memberType {
			return "", fmt.Errorf("Nested conditions must all use the same member_type, but found '%v' with '%v'", condMemberType, memberType)
		}
		memberType = condMemberType
	}
	return memberType, nil
}

type criteriaMeta struct {
	ExpressionType string
	MemberType     string
	IsNested       bool
	criteriaBlocks []interface{}
}

func validateGroupCriteriaSets(criteriaSets []interface{}) ([]criteriaMeta, error) {
	var criteria []criteriaMeta
	for _, criteriaBlock := range criteriaSets {
		seenExp := ""
		criteriaMap := criteriaBlock.(map[string]interface{})
		for expName, expVal := range criteriaMap {
			memberType := ""
			expValList := expVal.([]interface{})
			if len(expValList) > 0 {
				if seenExp != "" {
					return nil, fmt.Errorf("Criteria blocks are homogeneous, but found '%v' with '%v'", expName, seenExp)
				}
				if expName == "condition" {
					mType, err := validateNestedGroupConditions(expValList)
					if err != nil {
						return nil, err
					}
					memberType = mType
				} else if strings.HasSuffix(expName, "_expression") {
					memberType = ""
				} else {
					return nil, fmt.Errorf("Unknown criteria: %v", expName)
				}
				criteriaType := criteriaMeta{
					MemberType:     memberType,
					ExpressionType: expName,
					IsNested:       len(expValList) > 1,
					criteriaBlocks: expValList}
				criteria = append(criteria, criteriaType)
				seenExp = expName
			}
		}
	}

	return criteria, nil
}

func validateGroupConjunctions(conjunctions []interface{}, criteriaMeta []criteriaMeta) error {
	for index, conjunctionIFace := range conjunctions {
		conjunction := conjunctionIFace.(map[string]interface{})
		if conjunction["operator"] == model.ConjunctionOperator_CONJUNCTION_OPERATOR_AND {
			metaA := criteriaMeta[index]
			metaB := criteriaMeta[index+1]
			if metaA.ExpressionType != metaB.ExpressionType {
				return fmt.Errorf("AND conjunctions must use the same types of criteria expressions, but got %v and %v",
					metaA.ExpressionType, metaB.ExpressionType)
			}
			if metaA.MemberType != metaB.MemberType {
				return fmt.Errorf("AND conjunctions with conditions must have the same member types, but got %v and %v",
					metaA.MemberType, metaB.MemberType)
			}
		}
	}
	return nil
}

func buildGroupConditionData(condition interface{}) (*data.StructValue, error) {
	conditionMap := condition.(map[string]interface{})
	key := conditionMap["key"].(string)
	memberType := conditionMap["member_type"].(string)
	operator := conditionMap["operator"].(string)
	value := conditionMap["value"].(string)
	conditionModel := model.Condition{
		Key:          &key,
		MemberType:   &memberType,
		Operator:     &operator,
		Value:        &value,
		ResourceType: model.Condition__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(conditionModel, model.ConditionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupConjunctionData(conjunction string) (*data.StructValue, error) {
	conjunctionStruct := model.ConjunctionOperator{
		ConjunctionOperator: &conjunction,
		ResourceType:        model.ConjunctionOperator__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(conjunctionStruct, model.ConjunctionOperatorBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupIPAddressData(ipaddr interface{}) (*data.StructValue, error) {
	ipaddrMap := ipaddr.(map[string]interface{})
	var ipList []string
	for _, ip := range ipaddrMap["ip_addresses"].(*schema.Set).List() {
		ipList = append(ipList, ip.(string))
	}
	ipaddrStruct := model.IPAddressExpression{
		IpAddresses:  ipList,
		ResourceType: model.IPAddressExpression__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(ipaddrStruct, model.IPAddressExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupMacAddressData(ipaddr interface{}) (*data.StructValue, error) {
	addrMap := ipaddr.(map[string]interface{})
	var macList []string
	for _, mac := range addrMap["mac_addresses"].(*schema.Set).List() {
		macList = append(macList, mac.(string))
	}
	addrStruct := model.MACAddressExpression{
		MacAddresses: macList,
		ResourceType: model.MACAddressExpression__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(addrStruct, model.MACAddressExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupExternalIDExpressionData(externalID interface{}) (*data.StructValue, error) {
	idMap := externalID.(map[string]interface{})
	memberType := idMap["member_type"].(string)
	var extIDs []string

	for _, extID := range idMap["external_ids"].(*schema.Set).List() {
		extIDs = append(extIDs, extID.(string))
	}
	extIDStruct := model.ExternalIDExpression{
		MemberType:   &memberType,
		ExternalIds:  extIDs,
		ResourceType: model.ExternalIDExpression__TYPE_IDENTIFIER,
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(extIDStruct, model.ExternalIDExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupMemberPathData(paths interface{}) (*data.StructValue, error) {
	pathMap := paths.(map[string]interface{})
	var pathList []string
	for _, path := range pathMap["member_paths"].(*schema.Set).List() {
		pathList = append(pathList, path.(string))
	}
	ipaddrStruct := model.PathExpression{
		Paths:        pathList,
		ResourceType: model.PathExpression__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(ipaddrStruct, model.PathExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildNestedGroupExpressionData(expressions []*data.StructValue) (*data.StructValue, error) {
	var completeExpressions []*data.StructValue
	for i := 0; i < len(expressions)-1; i++ {
		completeExpressions = append(completeExpressions, expressions[i])
		// nested conditions have implicit AND
		conjData, err := buildGroupConjunctionData(model.ConjunctionOperator_CONJUNCTION_OPERATOR_AND)
		if err != nil {
			return nil, err
		}
		completeExpressions = append(completeExpressions, conjData)
	}
	completeExpressions = append(completeExpressions, expressions[len(expressions)-1])

	nestedStruct := model.NestedExpression{
		Expressions:  completeExpressions,
		ResourceType: model.NestedExpression__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(nestedStruct, model.NestedExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupExpressionDataFromType(expressionType string, datum interface{}) (*data.StructValue, error) {
	if datum == nil {
		return nil, fmt.Errorf("Empty set is not supported for expression type: %v", expressionType)
	}
	if expressionType == "condition" {
		data, err := buildGroupConditionData(datum)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else if expressionType == "ipaddress_expression" {
		data, err := buildGroupIPAddressData(datum)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else if expressionType == "path_expression" {
		data, err := buildGroupMemberPathData(datum)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else if expressionType == "macaddress_expression" {
		data, err := buildGroupMacAddressData(datum)
		if err != nil {
			return nil, err
		}
		return data, nil
	} else if expressionType == "external_id_expression" {
		data, err := buildGroupExternalIDExpressionData(datum)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return nil, fmt.Errorf("Unknown expression type: %v", expressionType)
}

func buildIdentityGroupExpressionListData(identityGroups []interface{}) (*data.StructValue, error) {
	var identityGroupExpressionList model.IdentityGroupExpression
	var identityGroupsList []model.IdentityGroupInfo
	for _, value := range identityGroups {
		identityGroupMap := value.(map[string]interface{})
		distinguishedName := identityGroupMap["distinguished_name"].(string)
		domainBaseDistinguishedName := identityGroupMap["domain_base_distinguished_name"].(string)
		sid := identityGroupMap["sid"].(string)
		identityGroupStruct := model.IdentityGroupInfo{
			DistinguishedName:           &distinguishedName,
			DomainBaseDistinguishedName: &domainBaseDistinguishedName,
			Sid:                         &sid,
		}
		identityGroupsList = append(identityGroupsList, identityGroupStruct)
	}
	identityGroupExpressionList.IdentityGroups = identityGroupsList
	identityGroupExpressionList.ResourceType = model.Expression_RESOURCE_TYPE_IDENTITYGROUPEXPRESSION
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(identityGroupExpressionList, model.IdentityGroupExpressionBindingType())
	if errors != nil {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildGroupExpressionData(criteriaMeta []criteriaMeta, conjunctions []interface{}) ([]*data.StructValue, error) {
	var expressionData []*data.StructValue
	for index, meta := range criteriaMeta {
		if meta.IsNested {
			var nestedExpressionData []*data.StructValue
			for _, nestedMeta := range meta.criteriaBlocks {
				nestedData, err := buildGroupExpressionDataFromType(meta.ExpressionType, nestedMeta)
				if err != nil {
					return nil, err
				}
				nestedExpressionData = append(nestedExpressionData, nestedData)
			}
			nested, err := buildNestedGroupExpressionData(nestedExpressionData)
			if err != nil {
				return nil, err
			}
			expressionData = append(expressionData, nested)
		} else {
			data, err := buildGroupExpressionDataFromType(meta.ExpressionType, meta.criteriaBlocks[0])
			if err != nil {
				return nil, err
			}
			expressionData = append(expressionData, data)
		}
		if index < len(conjunctions) {
			conjMap := conjunctions[index].(map[string]interface{})
			data, err := buildGroupConjunctionData(conjMap["operator"].(string))
			if err != nil {
				return nil, err
			}
			expressionData = append(expressionData, data)
		}
	}
	return expressionData, nil
}

func groupConditionDataToMap(expData *data.StructValue) (map[string]interface{}, error) {
	converter := bindings.NewTypeConverter()
	condData, errors := converter.ConvertToGolang(expData, model.ConditionBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}
	condStruct := condData.(model.Condition)
	var condMap = make(map[string]interface{})
	condMap["key"] = condStruct.Key
	condMap["member_type"] = condStruct.MemberType
	condMap["operator"] = condStruct.Operator
	condMap["value"] = condStruct.Value
	return condMap, nil
}

func fromGroupExpressionData(expressions []*data.StructValue) ([]map[string]interface{}, []map[string]interface{}, error) {
	var parsedConjunctions []map[string]interface{}
	var parsedCriteria []map[string]interface{}
	converter := bindings.NewTypeConverter()

	for _, expression := range expressions {
		expData, errs := converter.ConvertToGolang(expression, model.ExpressionBindingType())
		if len(errs) > 0 {
			return nil, nil, errs[0]
		}
		expStruct := expData.(model.Expression)

		if expStruct.ResourceType == model.Expression_RESOURCE_TYPE_CONJUNCTIONOPERATOR {
			log.Printf("[DEBUG] Parsing conjunction operator")
			conjData, errors := converter.ConvertToGolang(expression, model.ConjunctionOperatorBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			conjStruct := conjData.(model.ConjunctionOperator)
			var conjMap = make(map[string]interface{})
			conjMap["operator"] = conjStruct.ConjunctionOperator
			parsedConjunctions = append(parsedConjunctions, conjMap)
		} else if expStruct.ResourceType == model.IPAddressExpression__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing ipaddress expression")
			ipData, errors := converter.ConvertToGolang(expression, model.IPAddressExpressionBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			ipStruct := ipData.(model.IPAddressExpression)
			var addrList []map[string]interface{}
			var addrMap = make(map[string]interface{})
			addrMap["ip_addresses"] = ipStruct.IpAddresses
			var ipMap = make(map[string]interface{})
			addrList = append(addrList, addrMap)
			ipMap["ipaddress_expression"] = addrList
			parsedCriteria = append(parsedCriteria, ipMap)
		} else if expStruct.ResourceType == model.PathExpression__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing path expression")
			pathData, errors := converter.ConvertToGolang(expression, model.PathExpressionBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			paths := pathData.(model.PathExpression)
			var pathList []map[string]interface{}
			var pathMap = make(map[string]interface{})
			pathMap["member_paths"] = paths.Paths
			var exprMap = make(map[string]interface{})
			pathList = append(pathList, pathMap)
			exprMap["path_expression"] = pathList
			parsedCriteria = append(parsedCriteria, exprMap)
		} else if expStruct.ResourceType == model.MACAddressExpression__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing mac address expression")
			macData, errors := converter.ConvertToGolang(expression, model.MACAddressExpressionBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			macStruct := macData.(model.MACAddressExpression)
			var addrList []map[string]interface{}
			var addrMap = make(map[string]interface{})
			addrMap["mac_addresses"] = macStruct.MacAddresses
			var macMap = make(map[string]interface{})
			addrList = append(addrList, addrMap)
			macMap["macaddress_expression"] = addrList
			parsedCriteria = append(parsedCriteria, macMap)
		} else if expStruct.ResourceType == model.ExternalIDExpression__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing external id expression")
			extIDData, errors := converter.ConvertToGolang(expression, model.ExternalIDExpressionBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			extIDStruct := extIDData.(model.ExternalIDExpression)
			var idList []map[string]interface{}
			var extIDMap = make(map[string]interface{})
			extIDMap["member_type"] = extIDStruct.MemberType
			extIDMap["external_ids"] = extIDStruct.ExternalIds
			var idMap = make(map[string]interface{})
			idList = append(idList, extIDMap)
			idMap["external_id_expression"] = idList
			parsedCriteria = append(parsedCriteria, idMap)
		} else if expStruct.ResourceType == model.Condition__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing condition")
			condMap, err := groupConditionDataToMap(expression)
			if err != nil {
				return nil, nil, err
			}
			var condList []map[string]interface{}
			condList = append(condList, condMap)
			criteriaMap := make(map[string]interface{})
			criteriaMap["condition"] = condList
			parsedCriteria = append(parsedCriteria, criteriaMap)
		} else if expStruct.ResourceType == model.NestedExpression__TYPE_IDENTIFIER {
			log.Printf("[DEBUG] Parsing nested expression")
			nestedData, errors := converter.ConvertToGolang(expression, model.NestedExpressionBindingType())
			if len(errors) > 0 {
				return nil, nil, errors[0]
			}
			var criteriaList []interface{}
			nestedExp := nestedData.(model.NestedExpression)
			for _, nestedExpression := range nestedExp.Expressions {
				expData, errors := converter.ConvertToGolang(nestedExpression, model.ExpressionBindingType())
				if len(errors) > 0 {
					return nil, nil, errors[0]
				}
				eStruct := expData.(model.Expression)
				if eStruct.ResourceType == model.Condition__TYPE_IDENTIFIER {
					condMap, err := groupConditionDataToMap(nestedExpression)
					if err != nil {
						return nil, nil, err
					}
					criteriaList = append(criteriaList, condMap)
				}
			}
			criteriaMap := make(map[string]interface{})
			criteriaMap["condition"] = criteriaList
			parsedCriteria = append(parsedCriteria, criteriaMap)
		} else {
			return nil, nil, fmt.Errorf("Unsupported criteria type: %v", expStruct.ResourceType)
		}
	}
	return parsedCriteria, parsedConjunctions, nil
}

func getIdentityGroupsData(expressions []*data.StructValue) ([]map[string]interface{}, error) {
	var parsedIdentityGroups []map[string]interface{}
	converter := bindings.NewTypeConverter()
	for _, expr := range expressions {
		exprData, errs := converter.ConvertToGolang(expr, model.IdentityGroupExpressionBindingType())
		if len(errs) > 0 {
			return nil, errs[0]
		}
		exprStruct := exprData.(model.IdentityGroupExpression)
		log.Printf("[DEBUG] Parsing identity group")
		for _, identityGroup := range exprStruct.IdentityGroups {
			identityGroupMap := make(map[string]interface{})
			identityGroupMap["distinguished_name"] = identityGroup.DistinguishedName
			identityGroupMap["sid"] = identityGroup.Sid
			identityGroupMap["domain_base_distinguished_name"] = identityGroup.DomainBaseDistinguishedName
			parsedIdentityGroups = append(parsedIdentityGroups, identityGroupMap)
		}
	}
	return parsedIdentityGroups, nil
}

func validateGroupCriteriaAndConjunctions(criteriaSets []interface{}, conjunctions []interface{}) ([]criteriaMeta, error) {
	if len(criteriaSets)+len(conjunctions) == 0 {
		return nil, nil
	}
	if (len(criteriaSets)+len(conjunctions))%2 == 0 {
		if len(conjunctions) < len(criteriaSets)-1 {
			return nil, fmt.Errorf("Missing conjunction for criteria")
		}
		return nil, fmt.Errorf("Missing criteria for last conjunction")
	}
	criteriaMeta, err := validateGroupCriteriaSets(criteriaSets)
	if err != nil {
		return nil, err
	}

	err = validateGroupConjunctions(conjunctions, criteriaMeta)
	if err != nil {
		return nil, err
	}
	return criteriaMeta, nil
}

func resourceNsxtPolicyGroupCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyGroupExistsInDomainPartial(d.Get("domain").(string)))
	if err != nil {
		return err
	}

	criteriaSets := d.Get("criteria").([]interface{})
	conjunctions := d.Get("conjunction").([]interface{})

	criteriaMeta, err := validateGroupCriteriaAndConjunctions(criteriaSets, conjunctions)
	if err != nil {
		return err
	}

	expressionData, err := buildGroupExpressionData(criteriaMeta, conjunctions)
	if err != nil {
		return err
	}

	extendedCriteriaSets := d.Get("extended_criteria").([]interface{})
	err = validateExtendedCriteriaLocalManager(extendedCriteriaSets, m)
	if err != nil {
		return err
	}
	extendedExpressionList, err := buildGroupExtendedExpressionListData(extendedCriteriaSets)
	if err != nil {
		return err
	}
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	var groupTypes []string
	groupType := d.Get("group_type").(string)
	groupTypes = append(groupTypes, groupType)
	obj := model.Group{
		DisplayName:        &displayName,
		Description:        &description,
		Tags:               tags,
		Expression:         expressionData,
		ExtendedExpression: extendedExpressionList,
	}

	if groupType != "" && util.NsxVersionHigherOrEqual("3.2.0") {
		obj.GroupType = groupTypes
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewGroupsClient(context, connector)
	err = client.Patch(d.Get("domain").(string), id, obj)

	// Create the resource using PATCH
	log.Printf("[INFO] Creating Group with ID %s", id)
	if err != nil {
		return handleCreateError("Group", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGroupRead(d, m)
}

func resourceNsxtPolicyGroupRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	domainName := d.Get("domain").(string)
	if id == "" {
		return fmt.Errorf("Error obtaining Group ID")
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewGroupsClient(context, connector)
	obj, err := client.Get(domainName, id)
	if err != nil {
		return handleReadError(d, "Group", id, err)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("domain", getDomainFromResourcePath(*obj.Path))
	d.Set("revision", obj.Revision)
	groupType := ""
	if len(obj.GroupType) > 0 && util.NsxVersionHigherOrEqual("3.2.0") {
		groupType = obj.GroupType[0]
		d.Set("group_type", groupType)
	}
	criteria, conditions, err := fromGroupExpressionData(obj.Expression)
	log.Printf("[INFO] Found %d criteria, %d conjunctions for group %s", len(criteria), len(conditions), id)
	if err != nil {
		return err
	}
	d.Set("criteria", criteria)
	d.Set("conjunction", conditions)
	identityGroups, err := getIdentityGroupsData(obj.ExtendedExpression)
	log.Printf("[INFO] Found %d identity groups for group %s", len(identityGroups), id)
	if err != nil {
		return err
	}
	var extendedCriteria []map[string]interface{}
	if len(identityGroups) > 0 {
		identityGroupsMap := make(map[string]interface{})
		identityGroupsMap["identity_group"] = identityGroups
		extendedCriteria = append(extendedCriteria, identityGroupsMap)
	}
	d.Set("extended_criteria", extendedCriteria)

	return nil
}

func resourceNsxtPolicyGroupUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Group ID")
	}

	criteriaSets := d.Get("criteria").([]interface{})
	conjunctions := d.Get("conjunction").([]interface{})

	criteriaMeta, err := validateGroupCriteriaAndConjunctions(criteriaSets, conjunctions)
	if err != nil {
		return err
	}

	expressionData, err := buildGroupExpressionData(criteriaMeta, conjunctions)
	if err != nil {
		return err
	}

	extendedCriteriaSets := d.Get("extended_criteria").([]interface{})
	err = validateExtendedCriteriaLocalManager(extendedCriteriaSets, m)
	if err != nil {
		return err
	}
	extendedExpressionList, err := buildGroupExtendedExpressionListData(extendedCriteriaSets)
	if err != nil {
		return err
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	var groupTypes []string
	groupType := d.Get("group_type").(string)
	groupTypes = append(groupTypes, groupType)
	obj := model.Group{
		DisplayName:        &displayName,
		Description:        &description,
		Tags:               tags,
		Expression:         expressionData,
		ExtendedExpression: extendedExpressionList,
	}

	if groupType != "" && util.NsxVersionHigherOrEqual("3.2.0") {
		obj.GroupType = groupTypes
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewGroupsClient(context, connector)

	// Update the resource using PATCH
	err = client.Patch(d.Get("domain").(string), id, obj)
	if err != nil {
		return handleUpdateError("Group", id, err)
	}

	return resourceNsxtPolicyGroupRead(d, m)
}

func resourceNsxtPolicyGroupDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Group ID")
	}

	connector := getPolicyConnector(m)
	forceDelete := true
	failIfSubtreeExists := false

	doDelete := func() error {
		context, err := getSessionContext(d, m)
		if err != nil {
			return err
		}
		client := domains.NewGroupsClient(context, connector)
		return client.Delete(d.Get("domain").(string), id, &failIfSubtreeExists, &forceDelete)
	}

	err := doDelete()
	if err != nil {
		return handleDeleteError("Group", id, err)
	}

	return nil
}

func buildGroupExtendedExpressionListData(extendedCriteriaSets []interface{}) ([]*data.StructValue, error) {
	// Currently no nested criteria is supported in extended_expression, so extendedCriteriaSets has at most one element
	// Currently only identity groups are supported in extended_expression
	var identityGroups []interface{}
	for _, extendedCriteria := range extendedCriteriaSets {
		extendedCriteriaMap := extendedCriteria.(map[string]interface{})
		identityGroups = append(identityGroups, extendedCriteriaMap["identity_group"].(*schema.Set).List()...)
	}

	var extendedExpressionList []*data.StructValue
	if len(identityGroups) > 0 {
		identityGroupExpressionListData, err := buildIdentityGroupExpressionListData(identityGroups)
		if err != nil {
			return nil, err
		}
		extendedExpressionList = append(extendedExpressionList, identityGroupExpressionListData)
	}
	return extendedExpressionList, nil
}

func validateExtendedCriteriaLocalManager(extendedCriteriaSets []interface{}, clients interface{}) error {
	if len(extendedCriteriaSets) > 0 && isPolicyGlobalManager(clients) {
		err := fmt.Errorf("%s is not supported for Global Manager", "extended_criteria")
		return err
	}
	return nil
}
