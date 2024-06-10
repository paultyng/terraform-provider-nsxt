/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyContextProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyContextProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(false, false, false),
		},
	}
}

func dataSourceNsxtPolicyContextProfileRead(d *schema.ResourceData, m interface{}) error {
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceRead(d, getPolicyConnector(m), context, "PolicyContextProfile", nil)
	if err != nil {
		return err
	}
	return nil
}
