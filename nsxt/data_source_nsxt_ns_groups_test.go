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

func TestAccDataSourceNsxtNsGroups_basic(t *testing.T) {
	groupName := getAccTestDataSourceName()
	testResourceName := "data.nsxt_ns_groups.test"
	// in order to verify correct functionality, we compare fetching by name with regular data source
	// and fetching using the map data source with same name
	// in order to verify map fetching with pair comparison, we use an extra resource
	checkDataSourceName := "data.nsxt_ns_group.check"
	checkResourceName := "nsxt_ns_group.check"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtNsGroupDeleteByName(groupName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtNsGroupCreate(groupName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNSXNsGroupsReadTemplate(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrPair(checkResourceName, "display_name", checkDataSourceName, "id"),
				),
			},
		},
	})
}

func testAccNSXNsGroupsReadTemplate(groupName string) string {
	return fmt.Sprintf(`
data "nsxt_ns_groups" "test" {
}

data "nsxt_ns_group" "check" {
  display_name = "%s"
}

resource "nsxt_ns_group" "check" {
  display_name = data.nsxt_ns_groups.test.items["%s"]
}`, groupName, groupName)
}
