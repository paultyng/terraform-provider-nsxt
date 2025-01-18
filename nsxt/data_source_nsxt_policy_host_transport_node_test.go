// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyHostTransportNode_basic(t *testing.T) {
	htnName := getHostTransportNodeName()
	testResourceName := "data.nsxt_policy_host_transport_node.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_HOST_TRANSPORT_NODE")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXHostTransportNodeReadTemplate(htnName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", htnName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "unique_id"),
				),
			},
		},
	})
}

func testAccNSXHostTransportNodeReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_host_transport_node" "test" {
  display_name = "%s"
}`, name)
}
