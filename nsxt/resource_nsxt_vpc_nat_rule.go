// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/vpcs/nat"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var policyVpcNatRuleActionValues = []string{
	model.PolicyVpcNatRule_ACTION_SNAT,
	model.PolicyVpcNatRule_ACTION_DNAT,
	model.PolicyVpcNatRule_ACTION_REFLEXIVE,
}

var policyVpcNatRuleFirewallMatchValues = []string{
	model.PolicyVpcNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS,
	model.PolicyVpcNatRule_FIREWALL_MATCH_MATCH_INTERNAL_ADDRESS,
	model.PolicyVpcNatRule_FIREWALL_MATCH_BYPASS,
}

var policyVpcNatRuleSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"parent_path":  metadata.GetExtendedSchema(getPolicyPathSchema(true, true, "Policy path of the parent")),
	"translated_network": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateCidrOrIPOrRangeList(),
			Optional:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "TranslatedNetwork",
		},
	},
	"logging": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "Logging",
		},
	},
	"destination_network": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateCidrOrIPOrRangeList(),
			Optional:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "DestinationNetwork",
			OmitIfEmpty:  true,
		},
	},
	"action": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(policyVpcNatRuleActionValues, false),
			Required:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "Action",
		},
	},
	"firewall_match": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(policyVpcNatRuleFirewallMatchValues, false),
			Optional:     true,
			Default:      model.PolicyVpcNatRule_FIREWALL_MATCH_MATCH_INTERNAL_ADDRESS,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "FirewallMatch",
		},
	},
	"source_network": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateCidrOrIPOrRangeList(),
			Optional:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "SourceNetwork",
			OmitIfEmpty:  true,
		},
	},
	"enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "Enabled",
		},
	},
	"sequence_number": {
		Schema: schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "int",
			SdkFieldName: "SequenceNumber",
		},
	},
}

func resourceNsxtPolicyVpcNatRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyVpcNatRuleCreate,
		Read:   resourceNsxtPolicyVpcNatRuleRead,
		Update: resourceNsxtPolicyVpcNatRuleUpdate,
		Delete: resourceNsxtPolicyVpcNatRuleDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtParentPathResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(policyVpcNatRuleSchema),
	}
}

func resourceNsxtPolicyVpcNatRuleExists(sessionContext utl.SessionContext, parentPath string, id string, connector client.Connector) (bool, error) {
	var err error
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 4)
	if pathErr != nil {
		return false, pathErr
	}
	client := clientLayer.NewNatRulesClient(connector)
	_, err = client.Get(parents[0], parents[1], parents[2], parents[3], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyVpcNatRuleCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateIDWithParent(d, m, resourceNsxtPolicyVpcNatRuleExists)
	if err != nil {
		return err
	}

	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 4)
	if pathErr != nil {
		return pathErr
	}
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.PolicyVpcNatRule{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, policyVpcNatRuleSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating PolicyVpcNatRule with ID %s", id)

	client := clientLayer.NewNatRulesClient(connector)
	err = client.Patch(parents[0], parents[1], parents[2], parents[3], id, obj)
	if err != nil {
		return handleCreateError("PolicyVpcNatRule", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyVpcNatRuleRead(d, m)
}

func resourceNsxtPolicyVpcNatRuleRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining PolicyVpcNatRule ID")
	}

	client := clientLayer.NewNatRulesClient(connector)
	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 4)
	if pathErr != nil {
		return pathErr
	}
	obj, err := client.Get(parents[0], parents[1], parents[2], parents[3], id)
	if err != nil {
		return handleReadError(d, "PolicyVpcNatRule", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, policyVpcNatRuleSchema, "", nil)
}

func resourceNsxtPolicyVpcNatRuleUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining PolicyVpcNatRule ID")
	}

	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 4)
	if pathErr != nil {
		return pathErr
	}
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.PolicyVpcNatRule{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, policyVpcNatRuleSchema, "", nil); err != nil {
		return err
	}
	client := clientLayer.NewNatRulesClient(connector)
	_, err := client.Update(parents[0], parents[1], parents[2], parents[3], id, obj)
	if err != nil {
		// Trigger partial update to avoid terraform updating state based on failed intent
		// TODO - move this into handleUpdateError
		d.Partial(true)
		return handleUpdateError("PolicyVpcNatRule", id, err)
	}

	return resourceNsxtPolicyVpcNatRuleRead(d, m)
}

func resourceNsxtPolicyVpcNatRuleDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining PolicyVpcNatRule ID")
	}

	connector := getPolicyConnector(m)
	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 4)
	if pathErr != nil {
		return pathErr
	}

	client := clientLayer.NewNatRulesClient(connector)
	err := client.Delete(parents[0], parents[1], parents[2], parents[3], id)

	if err != nil {
		return handleDeleteError("PolicyVpcNatRule", id, err)
	}

	return nil
}
