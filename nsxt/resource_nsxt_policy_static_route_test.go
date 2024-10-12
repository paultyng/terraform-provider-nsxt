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

var testAccResourcePolicyStaticRouteName = "nsxt_policy_static_route.test"
var testAccResourcePolicyStaticRouteGatewayName = getAccTestResourceName()

func TestAccResourceNsxtPolicyStaticRoute_basicT0(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	network := "14.1.1.0/24"
	updateNetwork := "15.1.1.0/24"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteTier0CreateTemplate(name, network),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", network),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteTier0CreateTemplate(updateName, updateNetwork),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", updateNetwork),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteMultipleHopsTier0CreateTemplate(updateName, updateNetwork),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", updateNetwork),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "3"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyStaticRoute_basicT1(t *testing.T) {
	testAccResourceNsxtPolicyStaticRouteBasicT1(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccResourceNsxtPolicyStaticRoute_basicT1_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyStaticRouteBasicT1(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyStaticRouteBasicT1(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	network := "14.1.1.0/24"
	updateNetwork := "15.1.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteTier1CreateTemplate(name, network, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", network),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteTier1CreateTemplate(updateName, updateNetwork, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", updateNetwork),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteMultipleHopsTier1CreateTemplate(updateName, updateNetwork, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteExists(testAccResourcePolicyStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "network", updateNetwork),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "next_hop.#", "3"),
					resource.TestCheckResourceAttr(testAccResourcePolicyStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyStaticRouteName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyStaticRoute_basicT0Import(t *testing.T) {
	name := getAccTestResourceName()
	network := "14.1.1.0/24"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteTier0CreateTemplate(name, network),
			},
			{
				ResourceName:      testAccResourcePolicyStaticRouteName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyStaticRouteImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyStaticRoute_basicT1Import(t *testing.T) {
	name := getAccTestResourceName()
	network := "14.1.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteTier1CreateTemplate(name, network, false),
			},
			{
				ResourceName:      testAccResourcePolicyStaticRouteName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyStaticRouteImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyStaticRoute_basicT1Import_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	network := "14.1.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteTier1CreateTemplate(name, network, true),
			},
			{
				ResourceName:      testAccResourcePolicyStaticRouteName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testAccResourcePolicyStaticRouteName),
			},
		},
	})
}

func testAccNSXPolicyStaticRouteImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccResourcePolicyStaticRouteName]
	if !ok {
		return "", fmt.Errorf("NSX Policy Static Route resource %s not found in resources", testAccResourcePolicyStaticRouteName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Static Route resource ID not set in resources ")
	}
	gwPath := rs.Primary.Attributes["gateway_path"]
	if gwPath == "" {
		return "", fmt.Errorf("NSX Policy Static Route Gateway Policy Path not set in resources ")
	}
	_, gwID := parseGatewayPolicyPath(gwPath)
	return fmt.Sprintf("%s/%s", gwID, resourceID), nil
}

func testAccNsxtPolicyStaticRouteExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Static Route resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Static Route resource ID not set in resources")
		}

		gwPath := rs.Primary.Attributes["gateway_path"]
		isT0, gwID := parseGatewayPolicyPath(gwPath)
		_, err := getNsxtPolicyStaticRouteByID(testAccGetSessionContext(), connector, gwID, isT0, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy Static Route ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyStaticRouteCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_static_route" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		gwPath := rs.Primary.Attributes["gateway_path"]
		isT0, gwID := parseGatewayPolicyPath(gwPath)
		_, err := getNsxtPolicyStaticRouteByID(testAccGetSessionContext(), connector, gwID, isT0, resourceID)
		if err == nil {
			return fmt.Errorf("Policy Static Route %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyStaticRouteTier0CreateTemplate(name string, network string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier0_gateway" "t0test" {
  display_name              = "terraform-t0-gw"
  description               = "Acceptance Test"

}

resource "nsxt_policy_static_route" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  gateway_path        = "${nsxt_policy_tier0_gateway.t0test.path}"
  network             = "%s"
  next_hop {
    ip_address = "9.10.10.1"
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, name, network)
}

func testAccNsxtPolicyStaticRouteMultipleHopsTier0CreateTemplate(name string, network string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier0_gateway" "t0test" {
  display_name              = "terraform-t0-gw"
  description               = "Acceptance Test"

}

resource "nsxt_policy_static_route" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  gateway_path = "${nsxt_policy_tier0_gateway.t0test.path}"
  network             = "%s"
  next_hop {
    ip_address = "9.10.10.1"
  }
  next_hop {
    ip_address = "10.10.10.1"
  }
  next_hop {
    ip_address = "11.10.10.1"
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, name, network)
}

func testAccNsxtPolicyStaticRouteTier1CreateTemplate(name string, network string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "t1test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test"

}

resource "nsxt_policy_static_route" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  gateway_path        = "${nsxt_policy_tier1_gateway.t1test.path}"
  network             = "%s"
  next_hop {
    ip_address = "9.10.10.1"
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, testAccResourcePolicyStaticRouteGatewayName, context, name, network)
}

func testAccNsxtPolicyStaticRouteMultipleHopsTier1CreateTemplate(name string, network string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "t1test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test"

}

resource "nsxt_policy_static_route" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  gateway_path        = "${nsxt_policy_tier1_gateway.t1test.path}"
  network             = "%s"
  next_hop {
    ip_address = "9.10.10.1"
  }
  next_hop {
    ip_address = "10.10.10.1"
  }
  next_hop {
    admin_distance = 2
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, testAccResourcePolicyStaticRouteGatewayName, context, name, network)
}
