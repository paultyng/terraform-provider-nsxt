// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testAccResourceNatRuleName = "nsxt_nat_rule.test"

func TestAccResourceNsxtNatRule_snat(t *testing.T) {
	ruleName := getAccTestResourceName()
	updateRuleName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, updateRuleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSNATRuleCreateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testAccResourceNatRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testAccResourceNatRuleName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "logging", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "nat_pass", "false"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "action", "SNAT"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "translated_network", "4.4.4.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_destination_network", "3.3.3.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_source_network", "5.5.5.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "rule_priority", "5"),
				),
			},
			{
				Config: testAccNSXSNATRuleUpdateTemplate(updateRuleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(updateRuleName, testAccResourceNatRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "display_name", updateRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testAccResourceNatRuleName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "enabled", "false"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "logging", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "action", "SNAT"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "translated_network", "4.4.4.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_destination_network", "3.3.3.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_source_network", "6.6.6.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "rule_priority", "5"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNatRule_snatImport(t *testing.T) {
	ruleName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSNATRuleCreateTemplate(ruleName, edgeClusterName),
			},
			{
				ResourceName:      testAccResourceNatRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXNATRuleImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtNatRule_dnat(t *testing.T) {
	ruleName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDNATRuleCreateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testAccResourceNatRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testAccResourceNatRuleName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "logging", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "action", "DNAT"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "translated_network", "4.4.4.4"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_destination_network", "3.3.3.0/24"),
				),
			},
			{
				Config: testAccNSXDNATRuleUpdateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testAccResourceNatRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testAccResourceNatRuleName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "logging", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "action", "DNAT"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "translated_network", "4.4.4.4"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_destination_network", "7.7.7.0/24"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNatRule_dnatImport(t *testing.T) {
	ruleName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDNATRuleCreateTemplate(ruleName, edgeClusterName),
			},
			{
				ResourceName:      testAccResourceNatRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXNATRuleImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtNatRule_noNnat(t *testing.T) {
	ruleName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersionLessThan(t, "3.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, ruleName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNoNATRuleCreateTemplate(ruleName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNATRuleCheckExists(ruleName, testAccResourceNatRuleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "display_name", ruleName),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testAccResourceNatRuleName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "enabled", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "logging", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "nat_pass", "true"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "action", "NO_NAT"),
					resource.TestCheckResourceAttr(testAccResourceNatRuleName, "match_destination_network", "3.3.3.0/24"),
				),
			},
		},
	})
}

func testAccNSXNATRuleCheckExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX nat rule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX nat rule resource ID not set in resources ")
		}
		routerID := rs.Primary.Attributes["logical_router_id"]
		if routerID == "" {
			return fmt.Errorf("NSX nat rule routerID not set in resources ")
		}

		natRule, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.GetNatRule(nsxClient.Context, routerID, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving nat rule ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if nat rule %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == natRule.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX nat rule %s wasn't found", displayName)
	}
}

func testAccNSXNATRuleCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_nat_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		routerID := rs.Primary.Attributes["logical_router_id"]
		natRule, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.GetNatRule(nsxClient.Context, routerID, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving nat rule ID %s. Error: %v", resourceID, err)
		}

		if displayName == natRule.DisplayName {
			return fmt.Errorf("NSX nat rule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXNATRuleImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccResourceNatRuleName]
	if !ok {
		return "", fmt.Errorf("NSX nat rule resource %s not found in resources", testAccResourceNatRuleName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX nat rule resource ID not set in resources ")
	}
	routerID := rs.Primary.Attributes["logical_router_id"]
	if routerID == "" {
		return "", fmt.Errorf("NSX nat rule routerID not set in resources ")
	}
	return fmt.Sprintf("%s/%s", routerID, resourceID), nil
}

func testAccNSXNATRulePreConditionTemplate(edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_logical_tier1_router" "rtr1" {
  display_name    = "tier1_router"
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}`, edgeClusterName)
}

func testAccNSXSNATRuleCreateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
  logical_router_id         = "${nsxt_logical_tier1_router.rtr1.id}"
  display_name              = "%s"
  description               = "Acceptance Test"
  action                    = "SNAT"
  translated_network        = "4.4.4.0/24"
  match_destination_network = "3.3.3.0/24"
  match_source_network      = "5.5.5.0/24"
  enabled                   = "true"
  logging                   = "true"
  nat_pass                  = "false"
  rule_priority             = 5

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXSNATRuleUpdateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
  logical_router_id         = "${nsxt_logical_tier1_router.rtr1.id}"
  display_name              = "%s"
  description               = "Acceptance Test Update"
  action                    = "SNAT"
  translated_network        = "4.4.4.0/24"
  match_destination_network = "3.3.3.0/24"
  match_source_network      = "6.6.6.0/24"
  enabled                   = "false"
  logging                   = "true"
  nat_pass                  = "true"
  rule_priority             = 5

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, name)
}

func testAccNSXDNATRuleCreateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
  logical_router_id         = "${nsxt_logical_tier1_router.rtr1.id}"
  display_name              = "%s"
  description               = "Acceptance Test"
  action                    = "DNAT"
  translated_network        = "4.4.4.4"
  match_destination_network = "3.3.3.0/24"
  enabled                   = "true"
  logging                   = "true"
  nat_pass                  = "true"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXDNATRuleUpdateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
  logical_router_id         = "${nsxt_logical_tier1_router.rtr1.id}"
  display_name              = "%s"
  description               = "Acceptance Test Update"
  action                    = "DNAT"
  translated_network        = "4.4.4.4"
  match_destination_network = "7.7.7.0/24"
  enabled                   = "true"
  logging                   = "true"
  nat_pass                  = "true"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, name)
}

func testAccNSXNoNATRuleCreateTemplate(name string, edgeClusterName string) string {
	return testAccNSXNATRulePreConditionTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_nat_rule" "test" {
  logical_router_id         = "${nsxt_logical_tier1_router.rtr1.id}"
  display_name              = "%s"
  description               = "Acceptance Test"
  action                    = "NO_NAT"
  match_destination_network = "3.3.3.0/24"
  match_source_network      = "0.0.0.0/0"
  enabled                   = "true"
  logging                   = "true"
  nat_pass                  = "true"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}
