/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyGatewayPrefixList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayPrefixListRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"gateway_path": getPolicyPathSchema(false, false, "Gateway path"),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtPolicyGatewayPrefixListRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	query := make(map[string]string)
	if len(gwPath) > 0 {
		query["parent_path"] = fmt.Sprintf("%s*", gwPath)
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceReadWithValidation(d, connector, context, "PrefixList", query, false)
	if err != nil {
		return err
	}

	return nil
}
