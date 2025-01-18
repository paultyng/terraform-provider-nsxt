// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtHostUpgradeGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtHostUpgradeGroupRead,

		Schema: map[string]*schema.Schema{
			"upgrade_prepare_id": {
				Type:        schema.TypeString,
				Description: "ID of corresponding nsxt_upgrade_prepare resource",
				Required:    true,
			},
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
		},
	}
}

func dataSourceNsxtHostUpgradeGroupRead(d *schema.ResourceData, m interface{}) error {
	return upgradeGroupRead(d, m, hostUpgradeGroup)
}
