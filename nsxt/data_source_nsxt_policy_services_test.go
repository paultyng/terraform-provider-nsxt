// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyServices_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyServicesBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyServices_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyServicesBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyServicesBasic(t *testing.T, withContext bool, preCheck func()) {
	serviceName := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_services.test"
	checkResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyServicesReadTemplate(serviceName, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttr(checkResourceName, "display_name", serviceName),
				),
			},
		},
	})
}

func testAccNSXPolicyServicesReadTemplate(serviceName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyIcmpTypeServiceCreateTypeCodeTemplate(serviceName, "3", "1", "ICMPv4", withContext) + fmt.Sprintf(`
data "nsxt_policy_services" "test" {
  depends_on = [nsxt_policy_service.test]
%s
}

locals {
// Get id from path
  path_split = split("/", data.nsxt_policy_services.test.items["%s"])
  service_id = element(local.path_split, length(local.path_split) - 1)
}

data "nsxt_policy_service" "test" {
%s
  id = local.service_id
}
`, context, serviceName, context)
}
