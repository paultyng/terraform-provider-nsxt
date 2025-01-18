// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtComputeCollection_basic(t *testing.T) {
	ComputeCollectionName := getComputeCollectionName()
	testResourceName := "data.nsxt_compute_collection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_COMPUTE_COLLECTION")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXComputeCollectionReadTemplate(ComputeCollectionName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", ComputeCollectionName),
					resource.TestCheckResourceAttr(testResourceName, "origin_type", "VC_Cluster"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "origin_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "cm_local_id"),
				),
			},
		},
	})
}

func testAccNSXComputeCollectionReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_compute_collection" "test" {
  display_name = "%s"
}`, name)
}
