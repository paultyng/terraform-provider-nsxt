/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyCertificateRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func dataSourceNsxtPolicyCertificateRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceRead(d, connector, context, "TlsCertificate", nil)
	if err != nil {
		return err
	}

	return nil
}
