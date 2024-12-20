/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicySegmentSecurityProfile_basic(t *testing.T) {
	// Use existing system defined profile
	name := "default-segment-security-profile"
	testResourceName := "data.nsxt_policy_segment_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicySegmentSecurityProfile_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_segment_security_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileMultitenancyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicySegmentSecurityProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_segment_security_profile" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicySegmentSecurityProfileMultitenancyTemplate(name string) string {
	context := testAccNsxtProjectContext()
	return fmt.Sprintf(`
resource "nsxt_policy_segment_security_profile" "test" {
%s
  display_name = "%s"
}
data "nsxt_policy_segment_security_profile" "test" {
%s
  display_name = nsxt_policy_segment_security_profile.test.display_name
}`, context, name, context)
}
