/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDhcpRelayConfigCreateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform created",
	"server_addresses": "[\"2.2.2.3\", \"7001::23\"]",
}

var accTestPolicyDhcpRelayConfigUpdateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform updated",
	"server_addresses": "[\"4.1.1.23\"]",
}

func TestAccResourceNsxtPolicyDhcpRelayConfig_basic(t *testing.T) {
	testAccResourceNsxtPolicyDhcpRelayConfigBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDhcpRelayConfig_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDhcpRelayConfigBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDhcpRelayConfig_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDhcpRelayConfigBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

// NOTE: Realization is not tested here. Relay config is only realized when segment uses it.
func testAccResourceNsxtPolicyDhcpRelayConfigBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_dhcp_relay.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpRelayConfigCheckDestroy(state, accTestPolicyDhcpRelayConfigUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpRelayConfigTemplate(true, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpRelayConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpRelayConfigCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpRelayConfigCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "server_addresses.#", "2"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpRelayConfigTemplate(false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpRelayConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpRelayConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpRelayConfigUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "server_addresses.#", "1"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpRelayConfigMinimalistic(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpRelayConfigExists(testResourceName),
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

func TestAccResourceNsxtPolicyDhcpRelayConfig_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dhcp_relay.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpRelayConfigCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpRelayConfigMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyDhcpRelayConfig_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dhcp_relay.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpRelayConfigCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpRelayConfigMinimalistic(true),
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

func TestAccResourceNsxtPolicyDhcpRelayConfig_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dhcp_relay.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpRelayConfigCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpRelayConfigMinimalistic(false),
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

func testAccNsxtPolicyDhcpRelayConfigExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DhcpRelayConfig resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy DhcpRelayConfig resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDhcpRelayConfigExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy DhcpServer ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyDhcpRelayConfigCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_dhcp_relay" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyDhcpRelayConfigExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DhcpRelayConfig %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDhcpRelayConfigTemplate(createFlow bool, withContext bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDhcpRelayConfigCreateAttributes
	} else {
		attrMap = accTestPolicyDhcpRelayConfigUpdateAttributes
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_dhcp_relay" "test" {
%s
  display_name      = "%s"
  description      = "%s"
  server_addresses = %s


  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, attrMap["display_name"], attrMap["description"], attrMap["server_addresses"])
}

func testAccNsxtPolicyDhcpRelayConfigMinimalistic(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_dhcp_relay" "test" {
%s
  display_name     = "%s"
  server_addresses = %s
}`, context, accTestPolicyDhcpRelayConfigUpdateAttributes["display_name"], accTestPolicyDhcpRelayConfigUpdateAttributes["server_addresses"])
}
