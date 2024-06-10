/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyVPC() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyVPCRead,
		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(true, false, false),
			"short_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceNsxtPolicyVPCRead(d *schema.ResourceData, m interface{}) error {
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	obj, err := policyDataSourceResourceRead(d, getPolicyConnector(m), context, "Vpc", nil)
	if err != nil {
		return err
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(obj, model.VpcBindingType())
	if len(errors) > 0 {
		return errors[0]
	}
	vpc := dataValue.(model.Vpc)

	d.Set("short_id", vpc.ShortId)

	return nil
}
