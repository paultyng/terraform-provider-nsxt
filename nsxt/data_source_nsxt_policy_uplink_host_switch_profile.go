// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtUplinkHostSwitchProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtUplinkHostSwitchProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"realized_id": {
				Type:        schema.TypeString,
				Description: "The ID of the realized resource",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtUplinkHostSwitchProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	obj, err := policyDataSourceResourceReadWithValidation(d, connector, getSessionContext(d, m), infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE, nil, false)
	if err != nil {
		return fmt.Errorf("PolicyUplinkHostSwitchProfile with name '%s' was not found", d.Get("display_name").(string))
	}
	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(obj, model.PolicyUplinkHostSwitchProfileBindingType())
	if len(errors) > 0 {
		return errors[0]
	}
	pool := dataValue.(model.PolicyUplinkHostSwitchProfile)
	d.Set("realized_id", pool.RealizationId)

	return nil
}
