/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

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

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceReadWithValidation(d, connector, context, infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYVTEPHAHOSTSWITCHPROFILE, nil, false)
	if err == nil {
		return nil
	}

	return fmt.Errorf("PolicyVtepHAHostSwitchProfile with name '%s' was not found", d.Get("display_name").(string))
}
