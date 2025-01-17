/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var segmentPathExample = getMultitenancyPathExample("/infra/segments/[segment]")

func resourceNsxtPolicySegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySegmentCreate,
		Read:   resourceNsxtPolicySegmentRead,
		Update: resourceNsxtPolicySegmentUpdate,
		Delete: resourceNsxtPolicySegmentDelete,
		Importer: &schema.ResourceImporter{
			State: getPolicyPathOrIDResourceImporter(segmentPathExample),
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
