// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyPolicyLbPersistenceProfile_basic(t *testing.T) {
	testResourceName := "data.nsxt_policy_lb_persistence_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadTemplate("default-cookie-lb-persistence-profile"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-cookie-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "COOKIE"),
				),
			},
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadTemplate("default-generic-lb-persistence-profile"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-generic-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "GENERIC"),
				),
			},
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadTemplate("default-source-ip-lb-persistence-profile"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-source-ip-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "SOURCE_IP"),
				),
			},
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadWithTypeTemplate("default", "COOKIE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-cookie-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "COOKIE"),
				),
			},
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadWithTypeTemplate("default", "SOURCE_IP"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-source-ip-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "SOURCE_IP"),
				),
			},
			{
				Config: testAccNsxtPolicyPolicyLbPersistenceProfileReadWithTypeTemplate("default", "GENERIC"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default-generic-lb-persistence-profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(testResourceName, "type", "GENERIC"),
				),
			},
		},
	})
}

func testAccNsxtPolicyPolicyLbPersistenceProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_persistence_profile" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyPolicyLbPersistenceProfileReadWithTypeTemplate(name string, profileType string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_persistence_profile" "test" {
  display_name = "%s"
  type         = "%s"
}`, name, profileType)
}
