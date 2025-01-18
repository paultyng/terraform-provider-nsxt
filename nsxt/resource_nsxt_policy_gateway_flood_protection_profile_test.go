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

var accTestPolicyGatewayFloodProtectionProfileCreateAttributes = map[string]string{
	"description":              "terraform created",
	"icmp_active_flow_limit":   "2",
	"other_active_conn_limit":  "2",
	"tcp_half_open_conn_limit": "2",
	"udp_active_flow_limit":    "2",
	"nat_active_conn_limit":    "2",
}

var accTestPolicyGatewayFloodProtectionProfileUpdateAttributes = map[string]string{
	"description":              "terraform updated",
	"icmp_active_flow_limit":   "5",
	"other_active_conn_limit":  "5",
	"tcp_half_open_conn_limit": "5",
	"udp_active_flow_limit":    "5",
	"nat_active_conn_limit":    "5",
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfile_minimal(t *testing.T) {
	testResourceName := "nsxt_policy_gateway_flood_protection_profile.test"
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileMinimalistic(false, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayFloodProtectionProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyGatewayFloodProtectionProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_gateway_flood_protection_profile.test"
	if withContext {
		testResourceName = "nsxt_policy_gateway_flood_protection_profile.mttest"
	}
	name := getAccTestResourceName()
	updatedName := fmt.Sprintf("%s-updated", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileTemplate(true, withContext, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayFloodProtectionProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "icmp_active_flow_limit", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["icmp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "other_active_conn_limit", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["other_active_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_half_open_conn_limit", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["tcp_half_open_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "udp_active_flow_limit", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["udp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "nat_active_conn_limit", accTestPolicyGatewayFloodProtectionProfileCreateAttributes["nat_active_conn_limit"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileTemplate(false, withContext, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayFloodProtectionProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "icmp_active_flow_limit", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["icmp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "other_active_conn_limit", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["other_active_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_half_open_conn_limit", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["tcp_half_open_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "udp_active_flow_limit", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["udp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "nat_active_conn_limit", accTestPolicyGatewayFloodProtectionProfileUpdateAttributes["nat_active_conn_limit"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfile_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileImportBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyGatewayFloodProtectionProfile_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGatewayFloodProtectionProfileImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyGatewayFloodProtectionProfileImportBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_flood_protection_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayFloodProtectionProfileMinimalistic(withContext, name),
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

func testAccNsxtPolicyGatewayFloodProtectionProfileExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy FloodProtectionProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy GatewayFloodProtectionProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyFloodProtectionProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy GatewayFloodProtectionProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyGatewayFloodProtectionProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_flood_protection_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyFloodProtectionProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy GatewayFloodProtectionProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayFloodProtectionProfileTemplate(createFlow, withContext bool, name string) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyGatewayFloodProtectionProfileCreateAttributes
	} else {
		attrMap = accTestPolicyGatewayFloodProtectionProfileUpdateAttributes
	}
	context := ""
	resourceName := "test"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		resourceName = "mttest"
	}
	return fmt.Sprintf(`
resource "nsxt_policy_gateway_flood_protection_profile" "%s" {
%s
  display_name = "%s"
  description  = "%s"
  icmp_active_flow_limit = %s
  other_active_conn_limit = %s
  tcp_half_open_conn_limit = %s
  udp_active_flow_limit = %s
  nat_active_conn_limit = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_gateway_flood_protection_profile" "%s" {
%s
  display_name = "%s"
  depends_on = [nsxt_policy_gateway_flood_protection_profile.%s]
}`, resourceName, context, name, attrMap["description"], attrMap["icmp_active_flow_limit"], attrMap["other_active_conn_limit"], attrMap["tcp_half_open_conn_limit"], attrMap["udp_active_flow_limit"], attrMap["nat_active_conn_limit"], resourceName, context, name, resourceName)
}

func testAccNsxtPolicyGatewayFloodProtectionProfileMinimalistic(withContext bool, name string) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_gateway_flood_protection_profile" "test" {
%s
  display_name = "%s"

}`, context, name)
}
