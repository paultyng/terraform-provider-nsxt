// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyGatewayLocaleService_default(t *testing.T) {
	testAccDataSourceNsxtPolicyGatewayLocaleServiceDefault(t, false, func() { testAccPreCheck(t); testAccOnlyLocalManager(t) })
}

func testAccDataSourceNsxtPolicyGatewayLocaleServiceDefault(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_gateway_locale_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayLocaleServiceTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "bgp_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyGatewayLocaleService_single(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_gateway_locale_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayLocaleServiceSingleTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", "default"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "bgp_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyGatewayLocaleService_globalManager(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_gateway_locale_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayLocaleServiceGMTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "bgp_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyGatewayLocaleService_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyGatewayLocaleServiceDefault(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccNsxtPolicyGatewayLocaleServiceTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
}

data "nsxt_policy_gateway_locale_service" "test" {
%s
  gateway_path = nsxt_policy_tier1_gateway.test.path
  id           = "default"
}`, getEdgeClusterName(), context, name, context)
}

func testAccNsxtPolicyGatewayLocaleServiceSingleTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
}

data "nsxt_policy_gateway_locale_service" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
}`, getEdgeClusterName(), name)
}

func testAccNsxtPolicyGatewayLocaleServiceGMTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_site" "test" {
  display_name = "%s"
}

data "nsxt_policy_edge_cluster" "test" {
  site_path = data.nsxt_policy_site.test.path
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name = "%s"

  locale_service {
    edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  }
}

data "nsxt_policy_gateway_locale_service" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = one(nsxt_policy_tier0_gateway.test.locale_service).display_name
}`, getTestSiteName(), name)
}
