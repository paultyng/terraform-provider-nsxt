// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtEdgeCluster_basic(t *testing.T) {
	edgeClusterName := getEdgeClusterName()
	testResourceName := "data.nsxt_edge_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeClusterReadTemplate(edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", edgeClusterName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "member_node_type"),
					resource.TestCheckResourceAttrSet(testResourceName, "deployment_type"),
				),
			},
		},
	})
}

func testAccNSXEdgeClusterReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "test" {
  display_name = "%s"
}`, name)
}
