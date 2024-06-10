/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyHostTransportNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyHostTransportNodeRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"unique_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A unique identifier assigned by the system",
			},
		},
	}
}

func dataSourceNsxtPolicyHostTransportNodeRead(d *schema.ResourceData, m interface{}) error {
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	obj, err := policyDataSourceResourceRead(d, getPolicyConnector(m), context, "HostTransportNode", nil)
	if err != nil {
		return err
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(obj, model.HostTransportNodeBindingType())
	if len(errors) > 0 {
		return errors[0]
	}
	htn := dataValue.(model.HostTransportNode)
	d.Set("unique_id", htn.UniqueId)

	return nil
}
