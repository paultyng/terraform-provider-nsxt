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

// NOTE - to save test suite running time, this test also covers the data source
func TestAccResourceNsxtPolicyGatewayRouteMap_basic(t *testing.T) {
	testResourceName := "nsxt_policy_gateway_route_map.test"
	displayName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayRouteMapCheckDestroy(state, displayName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayRouteMapCreateTemplate(displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayRouteMapExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", "terraform created"),
					resource.TestCheckResourceAttr(testResourceName, "entry.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.action", "PERMIT"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.0.local_preference", "100"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.0.med", "0"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.prefix_list_matches.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.action", "DENY"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.community_list_match.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.community_list_match.0.criteria", "11:22"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.local_preference", "1122"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.as_path_prepend", "100 100"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.community", "11:22"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.prefer_global_v6_next_hop", "true"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.weight", "12"),
					resource.TestCheckResourceAttr(testResourceName, "entry.1.set.0.med", "120"),
					resource.TestCheckResourceAttr(testResourceName, "entry.2.action", "DENY"),
					resource.TestCheckResourceAttr(testResourceName, "entry.2.set.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayRouteMapUpdateTemplate(displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayRouteMapExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", "terraform updated"),
					resource.TestCheckResourceAttr(testResourceName, "entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.action", "PERMIT"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.0.local_preference", "100"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.set.0.med", "5"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.community_list_match.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "entry.0.prefix_list_matches.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayRouteMapMinimalistic(displayName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayRouteMapExists(displayName, testResourceName),
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

func TestAccResourceNsxtPolicyGatewayRouteMap_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_route_map.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayRouteMapCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayRouteMapMinimalistic(name),
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

func testAccNsxtPolicyGatewayRouteMapExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Gateway Route Map resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Gateway Route Map resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyGatewayRouteMapExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Gateway Route Map %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyGatewayRouteMapCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_route_map" {
			continue
		}

		resourceID := rs.Primary.ID
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyGatewayRouteMapExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Gateway Route Map %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayRouteMapCreateTemplate(displayName string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"
  gateway_path = nsxt_policy_tier0_gateway.test.path

  prefix {
    action  = "DENY"
    ge      = "20"
    le      = "23"
    network = "4.4.0.0/20"
  }
}

resource "nsxt_policy_gateway_route_map" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  description  = "terraform created"

  entry {
    action              = "PERMIT"
    prefix_list_matches = [nsxt_policy_gateway_prefix_list.test.path]

    set {
      weight = 12
    }
  }

  entry {
    action = "DENY"
    community_list_match {
      criteria       = "11:22"
      match_operator = "MATCH_COMMUNITY_REGEX"
    }

    set {
      as_path_prepend           = "100 100"
      community                 = "11:22"
      local_preference          = 1122
      med                       = 120
      prefer_global_v6_next_hop = true
      weight                    = 12
    }
  }

  entry {
    action = "DENY"
    community_list_match {
      criteria       = "22:*"
      match_operator = "MATCH_COMMUNITY_REGEX"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_gateway_route_map" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"

  depends_on = [nsxt_policy_gateway_route_map.test]
}`, displayName, displayName, displayName)
}

func testAccNsxtPolicyGatewayRouteMapUpdateTemplate(displayName string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"
  gateway_path = nsxt_policy_tier0_gateway.test.path

  prefix {
    action  = "DENY"
    ge      = "20"
    le      = "23"
    network = "4.4.0.0/20"
  }
}

resource "nsxt_policy_gateway_route_map" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  description  = "terraform updated"

  entry {
    action = "PERMIT"
    community_list_match {
      criteria       = "11:22"
      match_operator = "MATCH_COMMUNITY_REGEX"
    }
    community_list_match {
      criteria       = "11:*"
      match_operator = "MATCH_LARGE_COMMUNITY_REGEX"
    }
    set {
      weight = 12
      med    = 5
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_gateway_route_map" "test" {
  display_name = "%s"

  depends_on = [nsxt_policy_gateway_route_map.test]
}`, displayName, displayName, displayName)
}

func testAccNsxtPolicyGatewayRouteMapMinimalistic(displayName string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_route_map" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  entry {
    action = "PERMIT"
    community_list_match {
      criteria       = "11:22"
      match_operator = "MATCH_COMMUNITY_REGEX"
    }
  }
}`, displayName)
}
