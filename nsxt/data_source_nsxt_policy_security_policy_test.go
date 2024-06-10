/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicySecurityPolicy_basic(t *testing.T) {
	testAccDataSourceNsxtPolicySecurityPolicyBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicySecurityPolicy_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicySecurityPolicyBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicySecurityPolicy_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicySecurityPolicyBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccDataSourceNsxtPolicySecurityPolicyBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	category := "Application"
	testResourceName := "data.nsxt_policy_security_policy.test"
	withCategory := fmt.Sprintf(`category = "%s"`, category)
	withDomain := `domain = "default"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyTemplate(name, category, "", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "category", category),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyTemplate(name, category, withCategory, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "category", category),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyTemplate(name, category, withDomain, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "category", category),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicySecurityPolicy_default(t *testing.T) {
	testResourceName := "data.nsxt_policy_security_policy.test"
	category1 := "Ethernet"
	category2 := "Application"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDefaultSecurityPolicyTemplate(category1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "category", category1),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyDefaultSecurityPolicyTemplate(category2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "category", category2),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicySecurityPolicyTemplate(name string, category string, extra string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_security_policy" "test" {
%s
  display_name = "%s"
  description  = "%s"
  category     = "%s"
}

data "nsxt_policy_security_policy" "test" {
%s
  display_name = nsxt_policy_security_policy.test.display_name
  %s
}`, context, name, name, category, context, extra)
}

func testAccNsxtPolicyDefaultSecurityPolicyTemplate(category string) string {
	return fmt.Sprintf(`
data "nsxt_policy_security_policy" "test" {
  is_default = true
  category   = "%s"
}`, category)
}
