// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var dataSourceNsxtPolicyTestAddress = "20.20.0.2"

func TestAccDataSourceNsxtPolicyIPSecVpnLocalEndpoint(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_ipsec_vpn_local_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointDataSourceTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "local_address", dataSourceNsxtPolicyTestAddress),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyIPSecVpnLocalEndpoint_withService(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_ipsec_vpn_local_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointDataSourceTemplateWithService(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "local_address", dataSourceNsxtPolicyTestAddress),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnLocalEndpointDataSourceTemplate(name string) string {
	return testAccNsxtPolicyDataSourceIPSecVpnLocalEndpointPreconfig(name) + `
data "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
  display_name = nsxt_policy_ipsec_vpn_local_endpoint.test.display_name
}`
}

func testAccNsxtPolicyIPSecVpnLocalEndpointDataSourceTemplateWithService(name string) string {
	return testAccNsxtPolicyDataSourceIPSecVpnLocalEndpointPreconfig(name) + `
data "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
  display_name = nsxt_policy_ipsec_vpn_local_endpoint.test.display_name
  service_path = nsxt_policy_ipsec_vpn_service.test.path
}`
}

func testAccNsxtPolicyDataSourceIPSecVpnLocalEndpointPreconfig(name string) string {
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_service" "test" {
  display_name        = "%s"
  locale_service_path =  one(nsxt_policy_tier0_gateway.test.locale_service).path
}

resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
  service_path  = nsxt_policy_ipsec_vpn_service.test.path
  display_name  = "%s"
  local_address = "%s"
}`, name, name, dataSourceNsxtPolicyTestAddress)
}
