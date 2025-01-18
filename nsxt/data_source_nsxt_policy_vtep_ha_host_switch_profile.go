// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
)

func dataSourceNsxtVtepHAHostSwitchProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtVtepHAHostSwitchProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtVtepHAHostSwitchProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	_, err := policyDataSourceResourceReadWithValidation(d, connector, getSessionContext(d, m), infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYVTEPHAHOSTSWITCHPROFILE, nil, false)
	if err == nil {
		return nil
	}

	return fmt.Errorf("PolicyVtepHAHostSwitchProfile with name '%s' was not found", d.Get("display_name").(string))
}
