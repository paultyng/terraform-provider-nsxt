// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtComputeManager_basic(t *testing.T) {
	computeManagerName := getComputeManagerName()
	testResourceName := "data.nsxt_compute_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_COMPUTE_MANAGER")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXComputeManagerReadTemplate(computeManagerName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", computeManagerName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "server"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtComputeManager_single(t *testing.T) {
	testResourceName := "data.nsxt_compute_manager.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXComputeManagerSingleReadTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "server"),
				),
			},
		},
	})
}

func testAccNSXComputeManagerReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_compute_manager" "test" {
  display_name = "%s"
}`, name)
}

func testAccNSXComputeManagerSingleReadTemplate() string {
	return `
data "nsxt_compute_manager" "test" {
}`
}
