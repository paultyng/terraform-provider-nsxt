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

var accTestPolicyGatewayCommunityListCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"list":         `"NO_EXPORT_SUBCONFED", "65001:12"`,
}

var accTestPolicyGatewayCommunityListUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"list":         `"562:00:12"`,
}

func TestAccResourceNsxtPolicyGatewayCommunityList_basic(t *testing.T) {
	testResourceName := "nsxt_policy_gateway_community_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayCommunityListCheckDestroy(state, accTestPolicyGatewayCommunityListUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayCommunityListTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayCommunityListExists(accTestPolicyGatewayCommunityListCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyGatewayCommunityListCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayCommunityListCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "communities.#", "2"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayCommunityListTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayCommunityListExists(accTestPolicyGatewayCommunityListUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyGatewayCommunityListUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyGatewayCommunityListUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "communities.#", "1"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayCommunityListMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayCommunityListExists(accTestPolicyGatewayCommunityListCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyGatewayCommunityList_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_community_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayCommunityListCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayCommunityListMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyGetGatewayImporterIDGenerator(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyGatewayCommunityListExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Gateway Community List resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Gateway Community List resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyGatewayCommunityListExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Gateway Community List %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyGatewayCommunityListCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_community_list" {
			continue
		}

		resourceID := rs.Primary.ID
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyGatewayCommunityListExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Gateway Community List %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayCommunityListTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyGatewayCommunityListCreateAttributes
	} else {
		attrMap = accTestPolicyGatewayCommunityListUpdateAttributes
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_community_list" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  description  = "%s"
  communities  = [%s]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["list"])
}

func testAccNsxtPolicyGatewayCommunityListMinimalistic() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_community_list" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  communities  = [%s]
}`, accTestPolicyGatewayCommunityListUpdateAttributes["display_name"], accTestPolicyGatewayCommunityListUpdateAttributes["list"])
}
