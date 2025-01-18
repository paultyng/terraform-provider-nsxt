// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyGatewayConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayConnectionRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"tier0_path":   getPolicyPathSchema(false, false, "Tier0 Gateway path"),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtPolicyGatewayConnectionRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("tier0_path").(string)
	query := make(map[string]string)
	if len(gwPath) > 0 {
		query["tier0_path"] = gwPath
	}
	_, err := policyDataSourceResourceReadWithValidation(d, connector, getSessionContext(d, m), "GatewayConnection", query, false)
	if err != nil {
		return err
	}

	return nil
}
