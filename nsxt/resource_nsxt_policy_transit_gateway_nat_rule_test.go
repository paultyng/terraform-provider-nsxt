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

var accTestTransitGatewayNatRuleHelperName = getAccTestResourceName()
var accTestTransitGatewayNatRuleCreateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform created",
	"firewall_match":      "MATCH_EXTERNAL_ADDRESS",
	"logging":             "true",
	"action":              "DNAT",
	"source_network":      "2.2.2.14",
	"destination_network": "2.1.1.14",
	"translated_network":  "2.3.3.24",
	"enabled":             "false",
	"sequence_number":     "16",
}

var accTestTransitGatewayNatRuleUpdateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform updated",
	"firewall_match":      "MATCH_INTERNAL_ADDRESS",
	"logging":             "false",
	"action":              "DNAT",
	"source_network":      "3.3.3.14",
	"destination_network": "3.1.1.14",
	"translated_network":  "30.3.3.14",
	"enabled":             "true",
	"sequence_number":     "3",
}

func TestAccResourceNsxtTransitGatewayNatRule_basic(t *testing.T) {
	testResourceName := "nsxt_policy_transit_gateway_nat_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtTransitGatewayNatRuleCheckDestroy(state, accTestTransitGatewayNatRuleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtTransitGatewayNatRuleTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtTransitGatewayNatRuleExists(accTestTransitGatewayNatRuleCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestTransitGatewayNatRuleCreateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestTransitGatewayNatRuleCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", accTestTransitGatewayNatRuleCreateAttributes["translated_network"]),
					resource.TestCheckResourceAttr(testResourceName, "logging", accTestTransitGatewayNatRuleCreateAttributes["logging"]),
					resource.TestCheckResourceAttrSet(testResourceName, "destination_network"),
					resource.TestCheckResourceAttr(testResourceName, "action", accTestTransitGatewayNatRuleCreateAttributes["action"]),
					resource.TestCheckResourceAttr(testResourceName, "firewall_match", accTestTransitGatewayNatRuleCreateAttributes["firewall_match"]),
					resource.TestCheckResourceAttr(testResourceName, "source_network", accTestTransitGatewayNatRuleCreateAttributes["source_network"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestTransitGatewayNatRuleCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestTransitGatewayNatRuleCreateAttributes["sequence_number"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtTransitGatewayNatRuleTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtTransitGatewayNatRuleExists(accTestTransitGatewayNatRuleUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestTransitGatewayNatRuleUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestTransitGatewayNatRuleUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", accTestTransitGatewayNatRuleUpdateAttributes["translated_network"]),
					resource.TestCheckResourceAttr(testResourceName, "logging", accTestTransitGatewayNatRuleUpdateAttributes["logging"]),
					resource.TestCheckResourceAttrSet(testResourceName, "destination_network"),
					resource.TestCheckResourceAttr(testResourceName, "action", accTestTransitGatewayNatRuleUpdateAttributes["action"]),
					resource.TestCheckResourceAttr(testResourceName, "firewall_match", accTestTransitGatewayNatRuleUpdateAttributes["firewall_match"]),
					resource.TestCheckResourceAttr(testResourceName, "source_network", accTestTransitGatewayNatRuleUpdateAttributes["source_network"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestTransitGatewayNatRuleUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestTransitGatewayNatRuleUpdateAttributes["sequence_number"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtTransitGatewayNatRuleMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtTransitGatewayNatRuleExists(accTestTransitGatewayNatRuleCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtTransitGatewayNatRule_changeTypes(t *testing.T) {
	testResourceName := "nsxt_policy_transit_gateway_nat_rule.test"
	sourceIP := "2.2.2.34"
	translatedNetwork := "2.2.3.34"
	ruleName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtTransitGatewayNatRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtTransitGatewayNatRuleSnatTemplate(ruleName, translatedNetwork),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtTransitGatewayNatRuleExists(ruleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", ruleName),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "translated_network"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "false"),
					resource.TestCheckNoResourceAttr(testResourceName, "destination_network"),
					resource.TestCheckNoResourceAttr(testResourceName, "source_network"),
					resource.TestCheckResourceAttr(testResourceName, "action", "SNAT"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtTransitGatewayNatRuleReflexiveTemplate(ruleName, sourceIP, translatedNetwork),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtTransitGatewayNatRuleExists(ruleName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", ruleName),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "translated_network"),
					resource.TestCheckResourceAttr(testResourceName, "logging", "false"),
					resource.TestCheckNoResourceAttr(testResourceName, "destination_network"),
					resource.TestCheckResourceAttr(testResourceName, "source_network", sourceIP),
					resource.TestCheckResourceAttr(testResourceName, "action", "REFLEXIVE"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtTransitGatewayNatRule_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_transit_gateway_nat_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtTransitGatewayNatRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtTransitGatewayNatRuleMinimalistic(),
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

func testAccNsxtTransitGatewayNatRuleExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy TransitGatewayNatRule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy TransitGatewayNatRule resource ID not set in resources")
		}

		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtPolicyTransitGatewayNatRuleExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("TransitGatewayNatRule %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtTransitGatewayNatRuleCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_transit_gateway_nat_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtPolicyTransitGatewayNatRuleExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("TransitGatewayNatRule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtTransitGatewayNatRulePrerequisites() string {
	return fmt.Sprintf(`
resource "nsxt_policy_transit_gateway" "test" {
  %s
  display_name    = "%s"
  transit_subnets = ["192.168.5.0/24"]
}

data "nsxt_policy_transit_gateway_nat" "test" {
  transit_gateway_path = data.nsxt_policy_transit_gateway.test.path
}
`, testAccNsxtProjectContext(), accTestTransitGatewayNatRuleHelperName)
}

func testAccNsxtTransitGatewayNatRuleTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestTransitGatewayNatRuleCreateAttributes
	} else {
		attrMap = accTestTransitGatewayNatRuleUpdateAttributes
	}
	return testAccNsxtTransitGatewayNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_transit_gateway_nat_rule" "test" {
  parent_path         = data.nsxt_policy_transit_gateway_nat.test.path
  display_name        = "%s"
  description         = "%s"
  translated_network  = "%s"
  logging             = %s
  destination_network = "%s"
  action              = "%s"
  firewall_match      = "%s"
  source_network      = "%s"
  enabled             = %s
  sequence_number     = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["translated_network"], attrMap["logging"], attrMap["destination_network"], attrMap["action"], attrMap["firewall_match"], attrMap["source_network"], attrMap["enabled"], attrMap["sequence_number"])
}

func testAccNsxtTransitGatewayNatRuleMinimalistic() string {
	return testAccNsxtTransitGatewayNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_transit_gateway_nat_rule" "test" {
  parent_path         = data.nsxt_policy_transit_gateway_nat.test.path
  display_name        = "%s"
  destination_network = "%s"
  translated_network  = "%s"
  action              = "%s"
}`, accTestTransitGatewayNatRuleUpdateAttributes["display_name"], accTestTransitGatewayNatRuleUpdateAttributes["translated_network"], accTestTransitGatewayNatRuleUpdateAttributes["destination_network"], accTestTransitGatewayNatRuleUpdateAttributes["action"])
}

func testAccNsxtTransitGatewayNatRuleSnatTemplate(name string, translatedNetwork string) string {
	return testAccNsxtTransitGatewayNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_transit_gateway_nat_rule" "test" {
  parent_path         = data.nsxt_policy_transit_gateway_nat.test.path
  display_name        = "%s"
  translated_network  = "%s"
  action              = "SNAT"
}`, name, translatedNetwork)
}

func testAccNsxtTransitGatewayNatRuleReflexiveTemplate(name string, sourceIP string, translatedNetwork string) string {
	return testAccNsxtTransitGatewayNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_transit_gateway_nat_rule" "test" {
  parent_path         = data.nsxt_policy_transit_gateway_nat.test.path
  display_name        = "%s"
  source_network      = "%s"
  translated_network  = "%s"
  action              = "REFLEXIVE"
}`, name, sourceIP, translatedNetwork)
}
