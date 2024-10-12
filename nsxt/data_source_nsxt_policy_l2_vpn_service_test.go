// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceNsxtPolicyL2VpnService_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyL2VpnService_withGateway(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceReadTemplateWithGateway(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyL2VpnServiceReadTemplate(name string) string {
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
   resource "nsxt_policy_l2_vpn_service" "test" {
	display_name          = "%s"
	locale_service_path   = one(nsxt_policy_tier0_gateway.test.locale_service).path
   }`, name) + `
   data "nsxt_policy_l2_vpn_service" "test" {
	 display_name = nsxt_policy_l2_vpn_service.test.display_name
   }`
}

func testAccNsxtPolicyL2VpnServiceReadTemplateWithGateway(name string) string {
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
   resource "nsxt_policy_l2_vpn_service" "test" {
	display_name          = "%s"
	locale_service_path   = one(nsxt_policy_tier0_gateway.test.locale_service).path
   }`, name) + `
   data "nsxt_policy_l2_vpn_service" "test" {
	 display_name = nsxt_policy_l2_vpn_service.test.display_name
	 gateway_path = nsxt_policy_tier0_gateway.test.path
   }`
}
