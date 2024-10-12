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

var accTestTransitGatewayCreateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform created",
	"transit_subnets": "192.168.5.0/24",
}

var accTestTransitGatewayUpdateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform updated",
	"transit_subnets": "192.168.7.0/24",
}

func TestAccResourceNsxtPolicyTransitGateway_basic(t *testing.T) {
	testResourceName := "nsxt_policy_transit_gateway.test"
	testDataSourceName := "nsxt_policy_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
			testAccNSXVersion(t, "9.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTransitGatewayCheckDestroy(state, accTestTransitGatewayUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTransitGatewayTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTransitGatewayExists(accTestTransitGatewayCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestTransitGatewayCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestTransitGatewayCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "transit_subnets.0", accTestTransitGatewayCreateAttributes["transit_subnets"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),

					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
				),
			},
			{
				Config: testAccNsxtPolicyTransitGatewayTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTransitGatewayExists(accTestTransitGatewayUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestTransitGatewayUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestTransitGatewayUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "transit_subnets.0", accTestTransitGatewayUpdateAttributes["transit_subnets"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),

					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
				),
			},
			{
				Config: testAccNsxtPolicyTransitGatewayMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTransitGatewayExists(accTestTransitGatewayCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyTransitGateway_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
			testAccNSXVersion(t, "9.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTransitGatewayCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTransitGatewayMinimalistic(),
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

func testAccNsxtPolicyTransitGatewayExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy TransitGateway resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy TransitGateway resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyTransitGatewayExists(testAccGetSessionProjectContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy TransitGateway %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyTransitGatewayCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_transit_gateway" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyTransitGatewayExists(testAccGetSessionProjectContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy TransitGateway %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyTransitGatewayTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestTransitGatewayCreateAttributes
	} else {
		attrMap = accTestTransitGatewayUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_transit_gateway" "test" {
%s
  display_name    = "%s"
  description     = "%s"
  transit_subnets = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_transit_gateway" "test" {
%s
  display_name = "%s"
  depends_on = [nsxt_policy_transit_gateway.test]
}`, testAccNsxtProjectContext(), attrMap["display_name"], attrMap["description"], attrMap["transit_subnets"],
		testAccNsxtProjectContext(), attrMap["display_name"])
}

func testAccNsxtPolicyTransitGatewayMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_transit_gateway" "test" {
%s
  display_name    = "%s"
  transit_subnets = ["%s"]
}`, testAccNsxtProjectContext(), accTestTransitGatewayUpdateAttributes["display_name"], accTestTransitGatewayUpdateAttributes["transit_subnets"])
}
