/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

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
		Schema: getPolicySecurityPolicySchema(false, true, false),
	}
}

func parentSecurityPolicySchemaToModel(d *schema.ResourceData, id string) model.SecurityPolicy {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	category := d.Get("category").(string)
	comments := d.Get("comments").(string)
	locked := d.Get("locked").(bool)

	scope := getStringListFromSchemaSet(d, "scope")
	sequenceNumber := int64(d.Get("sequence_number").(int))
	stateful := d.Get("stateful").(bool)
	tcpStrict := d.Get("tcp_strict").(bool)
	objType := "SecurityPolicy"

	return model.SecurityPolicy{
		Id:             &id,
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		Category:       &category,
		Comments:       &comments,
		Locked:         &locked,
		Scope:          scope,
		SequenceNumber: &sequenceNumber,
		Stateful:       &stateful,
		TcpStrict:      &tcpStrict,
		ResourceType:   &objType,
	}
}

func parentSecurityPolicyModelToSchema(d *schema.ResourceData, m interface{}) (*model.SecurityPolicy, error) {
	connector := getPolicyConnector(m)
	id := d.Id()
	domainName := d.Get("domain").(string)
	if id == "" {
		return nil, fmt.Errorf("Error obtaining Security Policy id")
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return nil, err
	}
	client := domains.NewSecurityPoliciesClient(context, connector)
	obj, err := client.Get(domainName, id)
	if err != nil {
		return nil, err
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
	return resourceNsxtPolicySecurityPolicyGeneralCreate(d, m, false)
}

func resourceNsxtPolicyParentSecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, false)
}

func resourceNsxtPolicyParentSecurityPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralUpdate(d, m, false)
}

func resourceNsxtPolicyParentSecurityPolicyDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyDelete(d, m)
}
