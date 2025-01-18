// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtVPCGatewayPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVPCGatewayPolicyCreate,
		Read:   resourceNsxtVPCGatewayPolicyRead,
		Update: resourceNsxtVPCGatewayPolicyUpdate,
		Delete: resourceNsxtVPCGatewayPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPCPathResourceImporter,
		},

		Schema: getPolicyGatewayPolicySchema(true),
	}
}

func resourceNsxtVPCGatewayPolicyCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyGatewayPolicyExistsPartial(""))
	if err != nil {
		return err
	}

	err = policyGatewayPolicyBuildAndPatch(d, m, connector, isPolicyGlobalManager(m), id, true)
	if err != nil {
		return handleCreateError("VPC Gateway Policy", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVPCGatewayPolicyRead(d, m)
}

func resourceNsxtVPCGatewayPolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining VPC Gateway Policy ID")
	}

	obj, err := getGatewayPolicy(getSessionContext(d, m), id, "", connector)
	if err != nil {
		return handleReadError(d, "VPC Gateway Policy", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("comments", obj.Comments)
	d.Set("locked", obj.Locked)
	d.Set("sequence_number", obj.SequenceNumber)
	d.Set("stateful", obj.Stateful)
	if obj.TcpStrict != nil {
		// tcp_strict is dependent on stateful and maybe nil
		d.Set("tcp_strict", *obj.TcpStrict)
	}
	d.Set("revision", obj.Revision)
	return setPolicyRulesInSchema(d, obj.Rules)
}

func resourceNsxtVPCGatewayPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining VPC Gateway Policy ID")
	}

	err := policyGatewayPolicyBuildAndPatch(d, m, connector, isPolicyGlobalManager(m), id, true)
	if err != nil {
		return handleUpdateError("VPC Gateway Policy", id, err)
	}

	return resourceNsxtVPCGatewayPolicyRead(d, m)
}

func resourceNsxtVPCGatewayPolicyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining VPC Gateway Policy ID")
	}

	connector := getPolicyConnector(m)
	client := domains.NewGatewayPoliciesClient(getSessionContext(d, m), connector)
	err := client.Delete("", id)
	if err != nil {
		return handleDeleteError("VPC Gateway Policy", id, err)
	}

	return nil
}

func createChildVPCWithGatewayPolicy(context utl.SessionContext, policyID string, policy model.GatewayPolicy) (*data.StructValue, error) {
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

	childVPC := model.ChildResourceReference{
		Id:           &context.VPCID,
		ResourceType: "ChildResourceReference",
		TargetType:   strPtr("Vpc"),
		Children:     []*data.StructValue{dataValue.(*data.StructValue)},
	}

	dataValue, errors = converter.ConvertToVapi(childVPC, model.ChildResourceReferenceBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}
	childProject := model.ChildResourceReference{
		Id:           &context.ProjectID,
		ResourceType: "ChildResourceReference",
		TargetType:   strPtr("Project"),
		Children:     []*data.StructValue{dataValue.(*data.StructValue)},
	}
	dataValue, errors = converter.ConvertToVapi(childProject, model.ChildResourceReferenceBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}

	childOrg := model.ChildResourceReference{
		Id:           strPtr(defaultOrgID),
		ResourceType: "ChildResourceReference",
		TargetType:   strPtr("Org"),
		Children:     []*data.StructValue{dataValue.(*data.StructValue)},
	}
	dataValue, errors = converter.ConvertToVapi(childOrg, model.ChildResourceReferenceBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}

	return dataValue.(*data.StructValue), nil
}
