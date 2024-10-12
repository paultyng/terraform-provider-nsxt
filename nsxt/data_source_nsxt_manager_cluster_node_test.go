// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtManagerClusterNode_basic(t *testing.T) {
	testResourceName := "data.nsxt_manager_cluster_node.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
			testAccEnvDefined(t, "NSXT_TEST_MANAGER_CLUSTER_NODE")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyManagerClusterNodeReadTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "appliance_mgmt_listen_address"),
				),
			},
		},
	})
}

func testAccNsxtPolicyManagerClusterNodeReadTemplate() string {
	return fmt.Sprintf(`
data "nsxt_manager_cluster_node" "test" {
	display_name = "%s"
}`, getTestManagerClusterNode())
}
