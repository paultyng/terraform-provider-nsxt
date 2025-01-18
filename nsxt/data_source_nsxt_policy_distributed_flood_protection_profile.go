// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyDistributedFloodProtectionProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyDistributedFloodProtectionProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(false, false, false),
		},
	}
}

func dataSourceNsxtPolicyDistributedFloodProtectionProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	_, err := policyDataSourceResourceRead(d, connector, getSessionContext(d, m), "DistributedFloodProtectionProfile", nil)
	if err != nil {
		return err
	}

	return nil
}
