// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtVpcIpAddressAllocation() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtVpcIpAddressAllocationRead,

		Schema: map[string]*schema.Schema{
			"id": getDataSourceIDSchema(),
			"allocation_ips": {
				Type:        schema.TypeString,
				Description: "Allocation IPs - single IP or cidr",
				Required:    true,
			},
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(true, false, true),
		},
	}
}

func dataSourceNsxtVpcIpAddressAllocationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	ips := d.Get("allocation_ips").(string)
	query := make(map[string]string)
	query["allocation_ips"] = ips

	_, err := policyDataSourceResourceReadWithValidation(d, connector, getSessionContext(d, m), "VpcIpAddressAllocation", query, false)
	if err != nil {
		return err
	}
	// validate that IP address indeed matches

	return nil
}
