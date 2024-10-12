// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyHostTransportNodeProfile_basic(t *testing.T) {
	htnpName := getHostTransportNodeProfileName()
	testResourceName := "data.nsxt_policy_host_transport_node_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_HOST_TRANSPORT_NODE_PROFILE")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXHostTransportNodeProfileReadTemplate(htnpName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", htnpName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
				),
			},
		},
	})
}

func testAccNSXHostTransportNodeProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_host_transport_node_profile" "test" {
  display_name = "%s"
}`, name)
}
