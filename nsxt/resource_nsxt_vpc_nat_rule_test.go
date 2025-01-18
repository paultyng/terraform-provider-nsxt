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

var accTestVpcNatIPAllocationTestName = getAccTestResourceName()
var accTestVpcNatRuleCreateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform created",
	"firewall_match":     "MATCH_EXTERNAL_ADDRESS",
	"logging":            "true",
	"action":             "DNAT",
	"source_network":     "2.2.2.14",
	"translated_network": "2.3.3.24",
	"enabled":            "false",
	"sequence_number":    "16",
}

var accTestVpcNatRuleUpdateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform updated",
	"firewall_match":     "MATCH_INTERNAL_ADDRESS",
	"logging":            "false",
	"action":             "DNAT",
	"source_network":     "3.3.3.14",
	"translated_network": "30.3.3.14",
	"enabled":            "true",
	"sequence_number":    "3",
}

func TestAccResourceNsxtVpcNatRule_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_nat_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcNatRuleCheckDestroy(state, accTestVpcNatRuleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcNatRuleTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcNatRuleExists(accTestVpcNatRuleCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcNatRuleCreateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcNatRuleCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", accTestVpcNatRuleCreateAttributes["translated_network"]),
					resource.TestCheckResourceAttr(testResourceName, "logging", accTestVpcNatRuleCreateAttributes["logging"]),
					resource.TestCheckResourceAttrSet(testResourceName, "destination_network"),
					resource.TestCheckResourceAttr(testResourceName, "action", accTestVpcNatRuleCreateAttributes["action"]),
					resource.TestCheckResourceAttr(testResourceName, "firewall_match", accTestVpcNatRuleCreateAttributes["firewall_match"]),
					resource.TestCheckResourceAttr(testResourceName, "source_network", accTestVpcNatRuleCreateAttributes["source_network"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestVpcNatRuleCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestVpcNatRuleCreateAttributes["sequence_number"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcNatRuleTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcNatRuleExists(accTestVpcNatRuleUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcNatRuleUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcNatRuleUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "translated_network", accTestVpcNatRuleUpdateAttributes["translated_network"]),
					resource.TestCheckResourceAttr(testResourceName, "logging", accTestVpcNatRuleUpdateAttributes["logging"]),
					resource.TestCheckResourceAttrSet(testResourceName, "destination_network"),
					resource.TestCheckResourceAttr(testResourceName, "action", accTestVpcNatRuleUpdateAttributes["action"]),
					resource.TestCheckResourceAttr(testResourceName, "firewall_match", accTestVpcNatRuleUpdateAttributes["firewall_match"]),
					resource.TestCheckResourceAttr(testResourceName, "source_network", accTestVpcNatRuleUpdateAttributes["source_network"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestVpcNatRuleUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestVpcNatRuleUpdateAttributes["sequence_number"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcNatRuleMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcNatRuleExists(accTestVpcNatRuleCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtVpcNatRule_changeTypes(t *testing.T) {
	testResourceName := "nsxt_vpc_nat_rule.test"
	sourceIP := "2.2.2.34"
	ruleName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcNatRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcNatRuleSnatTemplate(ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcNatRuleExists(ruleName, testResourceName),
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
				Config: testAccNsxtVpcNatRuleReflexiveTemplate(ruleName, sourceIP),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcNatRuleExists(ruleName, testResourceName),
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

func TestAccResourceNsxtVpcNatRule_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_nat_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcNatRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcNatRuleMinimalistic(),
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

func testAccNsxtVpcNatRuleExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy VpcNatRule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy VpcNatRule resource ID not set in resources")
		}

		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtPolicyVpcNatRuleExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("VpcNatRule %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcNatRuleCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_nat_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtPolicyVpcNatRuleExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("VpcNatRule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVpcNatRulePrerequisites() string {
	return fmt.Sprintf(`
data "nsxt_vpc_nat" "test" {
  %s
  nat_type = "USER"
}

resource "nsxt_vpc_ip_address_allocation" "test" {
  %s
  allocation_size = 1
  display_name    = "%s"
}
`, testAccNsxtPolicyMultitenancyContext(), testAccNsxtPolicyMultitenancyContext(), accTestVpcNatIPAllocationTestName)
}

func testAccNsxtVpcNatRuleTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcNatRuleCreateAttributes
	} else {
		attrMap = accTestVpcNatRuleUpdateAttributes
	}
	return testAccNsxtVpcNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_nat_rule" "test" {
  parent_path         = data.nsxt_vpc_nat.test.path
  display_name        = "%s"
  description         = "%s"
  translated_network  = "%s"
  logging             = %s
  destination_network = nsxt_vpc_ip_address_allocation.test.allocation_ips
  action              = "%s"
  firewall_match      = "%s"
  source_network      = "%s"
  enabled             = %s
  sequence_number     = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["translated_network"], attrMap["logging"], attrMap["action"], attrMap["firewall_match"], attrMap["source_network"], attrMap["enabled"], attrMap["sequence_number"])
}

func testAccNsxtVpcNatRuleMinimalistic() string {
	return testAccNsxtVpcNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_nat_rule" "test" {
  parent_path         = data.nsxt_vpc_nat.test.path
  display_name        = "%s"
  destination_network = nsxt_vpc_ip_address_allocation.test.allocation_ips
  translated_network  = "%s"
  action              = "%s"
}`, accTestVpcNatRuleUpdateAttributes["display_name"], accTestVpcNatRuleUpdateAttributes["translated_network"], accTestVpcNatRuleUpdateAttributes["action"])
}

func testAccNsxtVpcNatRuleSnatTemplate(name string) string {
	return testAccNsxtVpcNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_nat_rule" "test" {
  parent_path         = data.nsxt_vpc_nat.test.path
  display_name        = "%s"
  translated_network  = nsxt_vpc_ip_address_allocation.test.allocation_ips
  action              = "SNAT"
}`, name)
}

func testAccNsxtVpcNatRuleReflexiveTemplate(name string, sourceIP string) string {
	return testAccNsxtVpcNatRulePrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_nat_rule" "test" {
  parent_path         = data.nsxt_vpc_nat.test.path
  display_name        = "%s"
  source_network      = "%s"
  translated_network  = nsxt_vpc_ip_address_allocation.test.allocation_ips
  action              = "REFLEXIVE"
}`, name, sourceIP)
}
