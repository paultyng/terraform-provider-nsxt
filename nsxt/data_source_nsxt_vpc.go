// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtVPC() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtVPCRead,
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

func dataSourceNsxtVPCRead(d *schema.ResourceData, m interface{}) error {
	obj, err := policyDataSourceResourceRead(d, getPolicyConnector(m), getSessionContext(d, m), "Vpc", nil)
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
