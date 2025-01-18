// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
)

func resourceNsxtPolicyParentSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyParentSecurityPolicyCreate,
		Read:   resourceNsxtPolicyParentSecurityPolicyRead,
		Update: resourceNsxtPolicyParentSecurityPolicyUpdate,
		Delete: resourceNsxtPolicyParentSecurityPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDomainResourceImporter,
		},
		Schema: getPolicySecurityPolicySchema(false, true, false, false),
	}
}

func parentSecurityPolicySchemaToModel(d *schema.ResourceData, id string) (model.SecurityPolicy, error) {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags, tagErr := getValidatedTagsFromSchema(d)
	if tagErr != nil {
		return model.SecurityPolicy{}, tagErr
	}
	cat, ok := d.GetOk("category")
	category := ""
	if ok {
		category = cat.(string)
	}
	comments := d.Get("comments").(string)
	locked := d.Get("locked").(bool)

	scope := getStringListFromSchemaSet(d, "scope")
	sequenceNumber := int64(d.Get("sequence_number").(int))
	stateful := d.Get("stateful").(bool)
	tcpStrict := d.Get("tcp_strict").(bool)
	objType := "SecurityPolicy"

	obj := model.SecurityPolicy{
		Id:             &id,
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		Comments:       &comments,
		Locked:         &locked,
		Scope:          scope,
		SequenceNumber: &sequenceNumber,
		Stateful:       &stateful,
		TcpStrict:      &tcpStrict,
		ResourceType:   &objType,
	}
	if category != "" {
		obj.Category = &category
	}
	return obj, nil
}

func parentSecurityPolicyModelToSchema(d *schema.ResourceData, m interface{}, isVPC bool) (*model.SecurityPolicy, error) {
	connector := getPolicyConnector(m)
	id := d.Id()
	domainName := ""
	if !isVPC {
		domainName = d.Get("domain").(string)
	}
	if id == "" {
		return nil, fmt.Errorf("Error obtaining Security Policy id")
	}
	client := domains.NewSecurityPoliciesClient(getSessionContext(d, m), connector)
	if client == nil {
		return nil, policyResourceNotSupportedError()
	}
	obj, err := client.Get(domainName, id)
	if err != nil {
		return nil, err
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	if !isVPC {
		d.Set("domain", getDomainFromResourcePath(*obj.Path))
		d.Set("category", obj.Category)
	}
	d.Set("comments", obj.Comments)
	d.Set("locked", obj.Locked)
	if len(obj.Scope) == 1 && obj.Scope[0] == "ANY" {
		d.Set("scope", nil)
	} else {
		d.Set("scope", obj.Scope)
	}
	d.Set("sequence_number", obj.SequenceNumber)
	d.Set("stateful", obj.Stateful)
	d.Set("tcp_strict", obj.TcpStrict)
	d.Set("revision", obj.Revision)
	return &obj, nil
}

func resourceNsxtPolicyParentSecurityPolicyCreate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralCreate(d, m, false, false)
}

func resourceNsxtPolicyParentSecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, false, false)
}

func resourceNsxtPolicyParentSecurityPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralUpdate(d, m, false, false)
}

func resourceNsxtPolicyParentSecurityPolicyDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyDelete(d, m)
}
