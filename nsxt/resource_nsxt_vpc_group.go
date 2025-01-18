// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceNsxtVPCGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVPCGroupCreate,
		Read:   resourceNsxtVPCGroupRead,
		Update: resourceNsxtVPCGroupUpdate,
		Delete: resourceNsxtVPCGroupDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPCPathResourceImporter,
		},

		Schema: getPolicyGroupSchema(false),
	}
}

func resourceNsxtVPCGroupCreate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyGroupGeneralCreate(d, m, false)
}

func resourceNsxtVPCGroupRead(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyGroupGeneralRead(d, m, false)
}

func resourceNsxtVPCGroupUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyGroupGeneralUpdate(d, m, false)
}

func resourceNsxtVPCGroupDelete(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyGroupGeneralDelete(d, m, false)
}
