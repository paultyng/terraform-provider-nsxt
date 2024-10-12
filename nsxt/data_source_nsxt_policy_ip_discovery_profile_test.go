// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyIPDiscoveryProfile_basic(t *testing.T) {
	// Use existing system defined profile
	name := "default-ip-discovery-profile"
	testResourceName := "data.nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyIPDiscoveryProfile_multitenancy(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileMultitenancyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyIPDiscoveryProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_ip_discovery_profile" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyIPDiscoveryProfileMultitenancyTemplate(name string) string {
	context := testAccNsxtPolicyMultitenancyContext()
	return fmt.Sprintf(`
resource "nsxt_policy_ip_discovery_profile" "test" {
%s
  display_name = "%s"
}

data "nsxt_policy_ip_discovery_profile" "test" {
%s
  display_name = nsxt_policy_ip_discovery_profile.test.display_name
}`, context, name, context)
}
