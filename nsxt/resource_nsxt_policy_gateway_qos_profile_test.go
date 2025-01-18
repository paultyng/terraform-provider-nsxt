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

var accTestPolicyGatewayQosProfileCreateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform created",
	"burst_size":          "2",
	"committed_bandwidth": "2",
	"excess_action":       "DROP",
}

var accTestPolicyGatewayQosProfileUpdateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform updated",
	"burst_size":          "5",
	"committed_bandwidth": "5",
	"excess_action":       "DROP",
}

func TestAccResourceNsxtPolicyGatewayQosProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_gateway_qos_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayQosProfileCheckDestroy(state, accTestPolicyGatewayQosProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayQosProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayQosProfileExists(accTestPolicyGatewayQosProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyGatewayQosProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayQosProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "burst_size", accTestPolicyGatewayQosProfileCreateAttributes["burst_size"]),
					resource.TestCheckResourceAttr(testResourceName, "committed_bandwidth", accTestPolicyGatewayQosProfileCreateAttributes["committed_bandwidth"]),
					resource.TestCheckResourceAttr(testResourceName, "excess_action", accTestPolicyGatewayQosProfileCreateAttributes["excess_action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayQosProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayQosProfileExists(accTestPolicyGatewayQosProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyGatewayQosProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayQosProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "burst_size", accTestPolicyGatewayQosProfileUpdateAttributes["burst_size"]),
					resource.TestCheckResourceAttr(testResourceName, "committed_bandwidth", accTestPolicyGatewayQosProfileUpdateAttributes["committed_bandwidth"]),
					resource.TestCheckResourceAttr(testResourceName, "excess_action", accTestPolicyGatewayQosProfileUpdateAttributes["excess_action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayQosProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayQosProfileExists(accTestPolicyGatewayQosProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyGatewayQosProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_qos_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayQosProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayQosProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyGatewayQosProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy GatewayQosProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy GatewayQosProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyGatewayQosProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy GatewayQosProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyGatewayQosProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_qos_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyGatewayQosProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy GatewayQosProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayQosProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyGatewayQosProfileCreateAttributes
	} else {
		attrMap = accTestPolicyGatewayQosProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_gateway_qos_profile" "test" {
  display_name = "%s"
  description  = "%s"
  burst_size = %s
  committed_bandwidth = %s
  excess_action = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["burst_size"], attrMap["committed_bandwidth"], attrMap["excess_action"])
}

func testAccNsxtPolicyGatewayQosProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_gateway_qos_profile" "test" {
  display_name = "%s"
}`, accTestPolicyGatewayQosProfileUpdateAttributes["display_name"])
}
