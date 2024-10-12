// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtLogicalTier0Router_basic(t *testing.T) {
	routerName := getTier0RouterName()
	testResourceName := "data.nsxt_logical_tier0_router.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalTier0RouterReadTemplate(routerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", routerName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "high_availability_mode"),
				),
			},
		},
	})
}

func testAccNSXLogicalTier0RouterReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_logical_tier0_router" "test" {
  display_name = "%s"
}`, name)
}
