// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNsxtPolicySegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySegmentCreate,
		Read:   resourceNsxtPolicySegmentRead,
		Update: resourceNsxtPolicySegmentUpdate,
		Delete: resourceNsxtPolicySegmentDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: getPolicyCommonSegmentSchema(false, false),
	}
}

func resourceNsxtPolicySegmentCreate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentCreate(d, m, false, false)
}

func resourceNsxtPolicySegmentRead(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentRead(d, m, false, false)
}

func resourceNsxtPolicySegmentUpdate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentUpdate(d, m, false, false)
}

func resourceNsxtPolicySegmentDelete(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentDelete(d, m, false)
}
