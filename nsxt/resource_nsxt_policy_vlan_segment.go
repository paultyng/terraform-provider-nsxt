// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNsxtPolicyVlanSegment() *schema.Resource {
	segSchema := getPolicyCommonSegmentSchema(true, false)
	delete(segSchema, "overlay_id")
	delete(segSchema, "connectivity_path")

	return &schema.Resource{
		Create: resourceNsxtPolicyVlanSegmentCreate,
		Read:   resourceNsxtPolicyVlanSegmentRead,
		Update: resourceNsxtPolicyVlanSegmentUpdate,
		Delete: resourceNsxtPolicyVlanSegmentDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: segSchema,
	}
}

func resourceNsxtPolicyVlanSegmentCreate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentCreate(d, m, true, false)
}

func resourceNsxtPolicyVlanSegmentRead(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentRead(d, m, true, false)
}

func resourceNsxtPolicyVlanSegmentUpdate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentUpdate(d, m, true, false)
}

func resourceNsxtPolicyVlanSegmentDelete(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentDelete(d, m, false)
}
