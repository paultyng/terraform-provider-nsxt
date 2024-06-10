/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicySecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySecurityPolicyCreate,
		Read:   resourceNsxtPolicySecurityPolicyRead,
		Update: resourceNsxtPolicySecurityPolicyUpdate,
		Delete: resourceNsxtPolicySecurityPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDomainResourceImporter,
		},
		Schema: getPolicySecurityPolicySchema(false, true, true),
	}
}

func getSecurityPolicyInDomain(sessionContext utl.SessionContext, id string, domainName string, connector client.Connector) (model.SecurityPolicy, error) {
	client := domains.NewSecurityPoliciesClient(sessionContext, connector)
	return client.Get(domainName, id)

}

func resourceNsxtPolicySecurityPolicyExistsInDomain(sessionContext utl.SessionContext, id string, domainName string, connector client.Connector) (bool, error) {
	client := domains.NewSecurityPoliciesClient(sessionContext, connector)
	_, err := client.Get(domainName, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving Security Policy", err)
}

func resourceNsxtPolicySecurityPolicyExistsPartial(domainName string) func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicySecurityPolicyExistsInDomain(sessionContext, id, domainName, connector)
	}
}

func policySecurityPolicyBuildAndPatch(d *schema.ResourceData, m interface{}, id string, createFlow, withRule bool) error {
	obj := parentSecurityPolicySchemaToModel(d, id)
	domain := d.Get("domain").(string)
	revision := int64(d.Get("revision").(int))
	log.Printf("[INFO] Creating Security Policy with ID %s", id)

	if createFlow && withRule {
		if err := validatePolicyRuleSequence(d); err != nil {
			return err
		}
	}
	if !createFlow {
		// This is update flow
		obj.Revision = &revision
	}

	if withRule {
		policyChildren, err := getUpdatedRuleChildren(d)
		if err != nil {
			return err
		}
		if len(policyChildren) > 0 {
			obj.Children = policyChildren
		}
	}

	log.Printf("[INFO] Using selective H-API for policy with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	return securityPolicyInfraPatch(context, obj, domain, m)
}

func resourceNsxtPolicySecurityPolicyCreate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralCreate(d, m, true)
}

func resourceNsxtPolicySecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, true)
}

func resourceNsxtPolicySecurityPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralUpdate(d, m, true)
}

func resourceNsxtPolicySecurityPolicyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy id")
	}

	connector := getPolicyConnector(m)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := domains.NewSecurityPoliciesClient(context, connector)
	err = client.Delete(d.Get("domain").(string), id)

	if err != nil {
		return handleDeleteError("Security Policy", id, err)
	}

	return nil
}

func resourceNsxtPolicySecurityPolicyGeneralCreate(d *schema.ResourceData, m interface{}, withRule bool) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicySecurityPolicyExistsPartial(d.Get("domain").(string)))
	if err != nil {
		return err
	}

	err = policySecurityPolicyBuildAndPatch(d, m, id, true, withRule)

	if err != nil {
		return handleCreateError("Security Policy", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, withRule)
}

func resourceNsxtPolicySecurityPolicyGeneralRead(d *schema.ResourceData, m interface{}, withRule bool) error {
	obj, err := parentSecurityPolicyModelToSchema(d, m)
	if err != nil {
		return handleReadError(d, "SecurityPolicy", d.Id(), err)
	}
	if withRule {
		return setPolicyRulesInSchema(d, obj.Rules)
	}
	return nil
}

func resourceNsxtPolicySecurityPolicyGeneralUpdate(d *schema.ResourceData, m interface{}, withRule bool) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Security Policy id")
	}
	err := policySecurityPolicyBuildAndPatch(d, m, id, false, withRule)
	if err != nil {
		return handleUpdateError("Security Policy", id, err)
	}

	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, withRule)
}
