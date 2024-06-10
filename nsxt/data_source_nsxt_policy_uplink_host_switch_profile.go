/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
)

func dataSourceNsxtUplinkHostSwitchProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtUplinkHostSwitchProfileRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtUplinkHostSwitchProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	_, err = policyDataSourceResourceReadWithValidation(d, connector, context, infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE, nil, false)
	if err == nil {
		return nil
	}

	return fmt.Errorf("PolicyUplinkHostSwitchProfile with name '%s' was not found", d.Get("display_name").(string))
}
