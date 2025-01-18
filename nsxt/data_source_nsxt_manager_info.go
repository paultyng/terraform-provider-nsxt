// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const nsxVersionID = "nsxversion"

func dataSourceNsxtManagerInfo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtManagerInfoRead,
		Schema: map[string]*schema.Schema{
			"version": {
				Type:        schema.TypeString,
				Description: "Version of NSXT manager",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtManagerInfoRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	version, err := getNSXVersion(connector)
	if err != nil {
		return fmt.Errorf("failed to retrieve NSX version: %v", err)
	}
	d.Set("version", version)
	d.SetId(nsxVersionID)

	return nil
}
