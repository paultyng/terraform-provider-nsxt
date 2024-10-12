// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
)

func resourceNsxtVPCSecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVPCSecurityPolicyCreate,
		Read:   resourceNsxtVPCSecurityPolicyRead,
		Update: resourceNsxtVPCSecurityPolicyUpdate,
		Delete: resourceNsxtVPCSecurityPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPCPathResourceImporter,
		},
		Schema: getPolicySecurityPolicySchema(false, true, true, true),
	}
}

func resourceNsxtVPCSecurityPolicyCreate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralCreate(d, m, true, true)
}

func resourceNsxtVPCSecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralRead(d, m, true, true)
}

func resourceNsxtVPCSecurityPolicyUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicySecurityPolicyGeneralUpdate(d, m, true, true)
}

func resourceNsxtVPCSecurityPolicyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VPC Security Policy id")
	}

	connector := getPolicyConnector(m)

	client := domains.NewSecurityPoliciesClient(getSessionContext(d, m), connector)
	err := client.Delete("", id)

	if err != nil {
		return handleDeleteError("Security Policy", id, err)
	}

	return nil
}
