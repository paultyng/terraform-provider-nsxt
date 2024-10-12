// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyMacDiscoveryProfile_basic(t *testing.T) {
	// Use existing system defined profile
	name := "default-mac-discovery-profile"
	testResourceName := "data.nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyMacDiscoveryProfile_prefix(t *testing.T) {
	// Use existing system defined profile
	name := "default-mac-discovery-profile-for-ens"
	namePrefix := name[0 : len(name)-3]
	testResourceName := "data.nsxt_policy_mac_discovery_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileReadTemplate(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Default MacDiscovery Profile for ENS"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyMacDiscoveryProfile_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileMultitenancyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyMacDiscoveryProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_mac_discovery_profile" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyMacDiscoveryProfileMultitenancyTemplate(name string) string {
	context := testAccNsxtPolicyMultitenancyContext()
	return fmt.Sprintf(`
resource "nsxt_policy_mac_discovery_profile" "test" {
%s
  display_name = "%s"
}

data "nsxt_policy_mac_discovery_profile" "test" {
%s
  display_name = nsxt_policy_mac_discovery_profile.test.display_name
}`, context, name, context)
}
