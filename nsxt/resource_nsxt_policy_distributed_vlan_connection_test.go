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

var accTestPolicyDistributedVlanConnectionCreateAttributes = map[string]string{
	"display_name":      getAccTestResourceName(),
	"description":       "terraform created",
	"vlan_id":           "12",
	"gateway_addresses": "3.3.3.1/24",
}

var accTestPolicyDistributedVlanConnectionUpdateAttributes = map[string]string{
	"display_name":      getAccTestResourceName(),
	"description":       "terraform updated",
	"vlan_id":           "22",
	"gateway_addresses": "2.2.2.1/24",
}

func TestAccResourceNsxtPolicyDistributedVlanConnection_basic(t *testing.T) {
	testResourceName := "nsxt_policy_distributed_vlan_connection.test"
	testDataSourceName := "data.nsxt_policy_distributed_vlan_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedVlanConnectionCheckDestroy(state, accTestPolicyDistributedVlanConnectionUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedVlanConnectionTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedVlanConnectionExists(accTestPolicyDistributedVlanConnectionCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDistributedVlanConnectionCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedVlanConnectionCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "vlan_id", accTestPolicyDistributedVlanConnectionCreateAttributes["vlan_id"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_addresses.0", accTestPolicyDistributedVlanConnectionCreateAttributes["gateway_addresses"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyDistributedVlanConnectionTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedVlanConnectionExists(accTestPolicyDistributedVlanConnectionUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDistributedVlanConnectionUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedVlanConnectionUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "vlan_id", accTestPolicyDistributedVlanConnectionUpdateAttributes["vlan_id"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_addresses.0", accTestPolicyDistributedVlanConnectionUpdateAttributes["gateway_addresses"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyDistributedVlanConnectionMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedVlanConnectionExists(accTestPolicyDistributedVlanConnectionCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDistributedVlanConnection_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_distributed_vlan_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedVlanConnectionCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedVlanConnectionMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyDistributedVlanConnectionExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy GatewayConnection resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy GatewayConnection resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDistributedVlanConnectionExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy GatewayConnection %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDistributedVlanConnectionCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_distributed_vlan_connection" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyDistributedVlanConnectionExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy GatewayConnection %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDistributedVlanConnectionTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDistributedVlanConnectionCreateAttributes
	} else {
		attrMap = accTestPolicyDistributedVlanConnectionUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_distributed_vlan_connection" "test" {
  display_name = "%s"
  description  = "%s"

  vlan_id           = %s
  gateway_addresses = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_distributed_vlan_connection" "test" {
  display_name = "%s"

  depends_on = [nsxt_policy_distributed_vlan_connection.test]
}`, attrMap["display_name"], attrMap["description"], attrMap["vlan_id"], attrMap["gateway_addresses"], attrMap["display_name"])
}

func testAccNsxtPolicyDistributedVlanConnectionMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_distributed_vlan_connection" "test" {
  display_name = "%s"
}

data "nsxt_policy_distributed_vlan_connection" "test" {
  display_name = "%s"

  depends_on = [nsxt_policy_distributed_vlan_connection.test]
}`, accTestPolicyDistributedVlanConnectionUpdateAttributes["display_name"], accTestPolicyDistributedVlanConnectionUpdateAttributes["display_name"])
}
