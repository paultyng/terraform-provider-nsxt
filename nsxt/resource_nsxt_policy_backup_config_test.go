/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtPolicyBackupConfig_basic(t *testing.T) {
	vmID := getTestVMID()
	testResourceName := "nsxt_policy_backup_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_BACKUP_CONFIG")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyBackupConfigCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyBackupConfigCreateTemplate(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyBackupConfigCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "backup_enabled", "true"),
				),
			},
			{
				Config: testAccNSXPolicyBackupConfigUpdateTemplate(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyBackupConfigCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "backup_enabled", "false"),
				),
			},
		},
	})
}
