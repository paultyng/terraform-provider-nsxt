// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceNsxtPolicyGroups_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyGroupsBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyGroups_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyGroupsBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyGroupsBasic(t *testing.T, withContext bool, preCheck func()) {
	domain := "default"
	groupName := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_groups.test"
	checkResourceName := "data.nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyGroupDeleteByName(domain, groupName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyGroupCreate(domain, groupName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNSXPolicyGroupsReadTemplate(groupName, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(checkResourceName, "display_name", groupName),
				),
			},
		},
	})
}

func testAccNSXPolicyGroupsReadTemplate(groupName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_groups" "test" {
%s
}

locals {
// Get id from path
  path_split = split("/", data.nsxt_policy_groups.test.items["%s"])
  group_id = element(local.path_split, length(local.path_split) - 1)
}

data "nsxt_policy_group" "test" {
%s
  id = local.group_id
}
`, context, groupName, context)
}
