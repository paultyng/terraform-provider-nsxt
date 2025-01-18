// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyGatewayPolicy_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyGatewayPolicyBasic(t, false, func() { testAccPreCheck(t) })
}

func TestAccDataSourceNsxtPolicyGatewayPolicy_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyGatewayPolicyBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyGatewayPolicyBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	category := "LocalGatewayRules"
	testResourceName := "data.nsxt_policy_gateway_policy.test"
	withCategory := fmt.Sprintf(`category = "%s"`, category)
	withDomain := `domain = "default"`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyTemplate(name, category, "", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "category", category),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyTemplate(name, category, withCategory, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "category", category),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyTemplate(name, category, withDomain, withContext),
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

func testAccNsxtPolicyGatewayPolicyTemplate(name string, category string, extra string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	return fmt.Sprintf(`
resource "nsxt_policy_gateway_policy" "test" {
%s
  display_name = "%s"
  description  = "%s"
  category     = "%s"
}

data "nsxt_policy_gateway_policy" "test" {
%s
  display_name = nsxt_policy_gateway_policy.test.display_name
  %s
}`, context, name, name, category, context, extra)
}
