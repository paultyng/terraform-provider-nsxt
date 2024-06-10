//* Copyright © 2024 VMware, Inc. All Rights Reserved.
//   SPDX-License-Identifier: MPL-2.0 */

// This test file tests both distributed_flood_protection_profile and distributed_flood_protection_profile_binding
package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyT0GatewayFloodProtectionProfileBindingCreateAttributes = map[string]string{
	"description":      "terraform created",
	"profile_res_name": "test1",
	"parent_path":      "nsxt_policy_tier0_gateway.test.path",
}

var accTestPolicyT0GatewayFloodProtectionProfileBindingUpdateAttributes = map[string]string{
	"description":      "terraform updated",
	"profile_res_name": "test2",
	"parent_path":      "nsxt_policy_tier0_gateway.test.path",
}

var accTestPolicyT0LSFloodProtectionProfileBindingCreateAttributes = map[string]string{
	"description":      "terraform created",
	"profile_res_name": "test1",
	"parent_path":      "data.nsxt_policy_gateway_locale_service.test.path",
}

var accTestPolicyT0LSFloodProtectionProfileBindingUpdateAttributes = map[string]string{
	"description":      "terraform updated",
	"profile_res_name": "test2",
	"parent_path":      "data.nsxt_policy_gateway_locale_service.test.path",
}

var accTestPolicyT1GatewayFloodProtectionProfileBindingCreateAttributes = map[string]string{
	"description":      "terraform created",
	"profile_res_name": "test1",
	"parent_path":      "nsxt_policy_tier1_gateway.test.path",
}

var accTestPolicyT1GatewayFloodProtectionProfileBindingUpdateAttributes = map[string]string{
	"description":      "terraform updated",
	"profile_res_name": "test2",
	"parent_path":      "nsxt_policy_tier1_gateway.test.path",
}

func TestAccResourceNsxtPolicyT0GatewayFloodProtectionProfileBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t, false, func() {
		testAccPreCheck(t)
	}, "tier0")
}

func TestAccResourceNsxtPolicyT0LSFloodProtectionProfileBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyGlobalManager(t)
	}, "ls")
}

func TestAccResourceNsxtPolicyT1GatewayFloodProtectionProfileBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t, false, func() {
		testAccPreCheck(t)
	}, "tier1")
}

func TestAccResourceNsxtPolicyT1GatewayFloodProtectionProfileBinding_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	}, "tier1")
}

func TestAccResourceNsxtPolicyT1GatewayFloodProtectionProfileBinding_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	}, "tier1")
}

func testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingBasic(t *testing.T, withContext bool, preCheck func(), parent string) {
	testResourceName := "nsxt_policy_gateway_flood_protection_profile_binding.test"
	if withContext {
		testResourceName = "nsxt_policy_gateway_flood_protection_profile_binding.mttest"
	}
	name := getAccTestResourceName()
	updatedName := fmt.Sprintf("%s-updated", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayFloodProtectionProfileBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileBindingTemplate(true, withContext, name, parent),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayFloodProtectionProfileBindingExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "terraform created"),
					resource.TestCheckResourceAttrSet(testResourceName, "profile_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileBindingTemplate(false, withContext, updatedName, parent),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayFloodProtectionProfileBindingExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "terraform updated"),
					resource.TestCheckResourceAttrSet(testResourceName, "profile_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfileBinding_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingImportBasic(t, false, func() {
		testAccPreCheck(t)
	}, "tier1")
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfileBinding_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	}, "tier1")
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfileBinding_importBasic_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingImportBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	}, "tier1")
}

func testAccResourceNsxtPolicyGatewayFloodProtectionProfileBindingImportBasic(t *testing.T, withContext bool, preCheck func(), parent string) {
	testResourceName := "nsxt_policy_gateway_flood_protection_profile_binding.test"
	if withContext {
		testResourceName = "nsxt_policy_gateway_flood_protection_profile_binding.mttest"
	}
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayFloodProtectionProfileBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileBindingTemplate(true, withContext, name, parent),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyGatewayFloodProtectionProfileBindingExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy GatewayFloodProtectionProfileBinding resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy GatewayFloodProtectionProfileBinding resource ID not set in resources")
		}
		parentPath := rs.Primary.Attributes["parent_path"]
		exists, err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingExists(testAccGetSessionContext(), connector, parentPath, resourceID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy GatewayFloodProtectionProfileBinding %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyGatewayFloodProtectionProfileBindingCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_flood_protection_profile_binding" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		parentPath := rs.Primary.Attributes["parent_path"]
		exists, err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingExists(testAccGetSessionContext(), connector, parentPath, resourceID)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy GatewayFloodProtectionProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayFloodProtectionProfileBindingTemplate(createFlow, withContext bool, name, parent string) string {
	var attrMap map[string]string
	if createFlow {
		switch parent {
		case "tier0":
			attrMap = accTestPolicyT0GatewayFloodProtectionProfileBindingCreateAttributes
		case "tier1":
			attrMap = accTestPolicyT1GatewayFloodProtectionProfileBindingCreateAttributes
		case "ls":
			attrMap = accTestPolicyT0LSFloodProtectionProfileBindingCreateAttributes
		}
	} else {
		switch parent {
		case "tier0":
			attrMap = accTestPolicyT0GatewayFloodProtectionProfileBindingUpdateAttributes
		case "tier1":
			attrMap = accTestPolicyT1GatewayFloodProtectionProfileBindingUpdateAttributes
		case "ls":
			attrMap = accTestPolicyT0LSFloodProtectionProfileBindingUpdateAttributes
		}
	}
	context := ""
	resourceName := "test"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		resourceName = "mttest"
	}
	return testAccNsxtPolicyGatewayFloodProtectionProfileBindingDeps(withContext) + fmt.Sprintf(`
resource "nsxt_policy_gateway_flood_protection_profile_binding" "%s" {
%s
 display_name = "%s"
 description  = "%s"
 profile_path = nsxt_policy_gateway_flood_protection_profile.%s.path 
 parent_path = %s

 tag {
   scope = "scope1"
   tag   = "tag1"
 }
}
`, resourceName, context, name, attrMap["description"], attrMap["profile_res_name"], attrMap["parent_path"])
}

func testAccNsxtPolicyGatewayFloodProtectionProfileBindingDeps(withContext bool) string {
	context := ""
	parentDeps := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		parentDeps = fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name             = "test"
}
`, context)
	} else if testAccIsGlobalManager() {
		parentDeps = fmt.Sprintf(`
data "nsxt_policy_site" "site1" {
  display_name = "%s"
}

data "nsxt_policy_edge_cluster" "ec_site1" {
  site_path = data.nsxt_policy_site.site1.path
}

data "nsxt_policy_edge_node" "en_site1" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.ec_site1.path
  member_index      = 0
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "test"
  locale_service {
    edge_cluster_path    = data.nsxt_policy_edge_cluster.ec_site1.path
    preferred_edge_paths = [data.nsxt_policy_edge_node.en_site1.path]
  }
}

data "nsxt_policy_gateway_locale_service" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name = "test"
}`, getTestSiteName())
	} else {
		parentDeps = `
resource "nsxt_policy_tier0_gateway" "test" {
  display_name = "test"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name = "test"
}
`
	}
	return parentDeps + fmt.Sprintf(`
resource "nsxt_policy_gateway_flood_protection_profile" "test1" {
%s
  display_name = "gfpp1"
  description  = "Acceptance Test"
  icmp_active_flow_limit = 3
  other_active_conn_limit = 3
  tcp_half_open_conn_limit = 3
  udp_active_flow_limit = 3
  nat_active_conn_limit = 3
}

resource "nsxt_policy_gateway_flood_protection_profile" "test2" {
%s
  display_name = "gfpp2"
  description  = "Acceptance Test"
  icmp_active_flow_limit = 4
  other_active_conn_limit = 4
  tcp_half_open_conn_limit = 4
  udp_active_flow_limit = 4
  nat_active_conn_limit = 4
}
`, context, context)
}
