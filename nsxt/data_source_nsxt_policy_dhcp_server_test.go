/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyDhcpServer_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyDhcpServerBasic(t, false, func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") })
}

func TestAccDataSourceNsxtPolicyDhcpServer_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyDhcpServerBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyDhcpServerBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_dhcp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpServerReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyDhcpServerReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtProjectContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_dhcp_server" "test" {
%s
  display_name = "%s"
  description  = "test"
}

data "nsxt_policy_dhcp_server" "test" {
%s
  id = nsxt_policy_dhcp_server.test.nsx_id
}`, context, name, context)
}
