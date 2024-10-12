// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyGatewayRouteMap() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayRouteMapRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"gateway_path": getPolicyPathSchema(false, false, "Gateway path"),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtPolicyGatewayRouteMapRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	query := make(map[string]string)
	if len(gwPath) > 0 {
		query["parent_path"] = fmt.Sprintf("%s*", gwPath)
	}
	_, err := policyDataSourceResourceReadWithValidation(d, connector, getSessionContext(d, m), "Tier0RouteMap", query, false)
	if err != nil {
		return err
	}

	return nil
}
