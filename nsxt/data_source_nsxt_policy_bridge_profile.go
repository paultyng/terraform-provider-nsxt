/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyBridgeProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyBridgeProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func dataSourceNsxtPolicyBridgeProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceRead(d, connector, context, "L2BridgeEndpointProfile", nil)
	if err != nil {
		return err
	}

	return nil
}
