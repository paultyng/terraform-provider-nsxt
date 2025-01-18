// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyEdgeNode_basic(t *testing.T) {
	edgeClusterName := getEdgeClusterName()
	testResourceName := "data.nsxt_policy_edge_node.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNSXGlobalManagerSitePrecheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyEdgeNodeReadTemplate(edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "member_index", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyEdgeNodeReadTemplate(name string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(name) + `

data "nsxt_policy_edge_node" "test" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  member_index      = 0
}`
}
