/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyIPPool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyIPPoolRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(false, false, false),
			"realized_id": {
				Type:        schema.TypeString,
				Description: "The ID of the realized resource",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyIPPoolRead(d *schema.ResourceData, m interface{}) error {
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	obj, err := policyDataSourceResourceRead(d, getPolicyConnector(m), context, "IpAddressPool", nil)
	if err != nil {
		return err
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(obj, model.IpAddressPoolBindingType())
	if len(errors) > 0 {
		return errors[0]
	}
	pool := dataValue.(model.IpAddressPool)
	d.Set("realized_id", pool.RealizationId)
	return nil
}
