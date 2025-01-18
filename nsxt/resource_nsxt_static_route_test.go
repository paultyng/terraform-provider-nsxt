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

var testAccResourceStaticRouteName = "nsxt_static_route.test"
var testAccResourceStaticRouteHelperName = getAccTestResourceName()

func TestAccResourceNsxtStaticRoute_basic(t *testing.T) {
	testAccResourceNsxtStaticRoute(t, "tier1")
}

func TestAccResourceNsxtStaticRoute_onT0(t *testing.T) {
	testAccResourceNsxtStaticRoute(t, "tier0")
}

func testAccResourceNsxtStaticRoute(t *testing.T, tier string) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()
	transportZoneName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXStaticRouteCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXStaticRouteCreateTemplate(tier, name, edgeClusterName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXStaticRouteCheckExists(name, testAccResourceStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testAccResourceStaticRouteName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "network", "4.4.4.0/24"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "next_hop.0.administrative_distance", "1"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "next_hop.0.ip_address", "8.0.0.10"),
					resource.TestCheckResourceAttrSet(testAccResourceStaticRouteName, "next_hop.0.logical_router_port_id"),
				),
			},
			{
				Config: testAccNSXStaticRouteUpdateTemplate(tier, updateName, edgeClusterName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXStaticRouteCheckExists(updateName, testAccResourceStaticRouteName),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testAccResourceStaticRouteName, "logical_router_id"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "network", "5.5.5.0/24"),
					resource.TestCheckResourceAttr(testAccResourceStaticRouteName, "next_hop.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtStaticRoute_importBasic(t *testing.T) {
	testAccResourceNsxtStaticRouteImport(t, "tier1")
}

func TestAccResourceNsxtStaticRoute_importOnTier0(t *testing.T) {
	testAccResourceNsxtStaticRouteImport(t, "tier0")
}

func testAccResourceNsxtStaticRouteImport(t *testing.T, tier string) {
	name := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()
	transportZoneName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXStaticRouteCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXStaticRouteCreateTemplate(tier, name, edgeClusterName, transportZoneName),
			},
			{
				ResourceName:      testAccResourceStaticRouteName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXStaticRouteImporterGetID,
			},
		},
	})
}

func testAccNSXStaticRouteCheckExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX static route resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX static route resource ID not set in resources ")
		}
		routerID := rs.Primary.Attributes["logical_router_id"]
		if routerID == "" {
			return fmt.Errorf("NSX static route routerID not set in resources ")
		}

		staticRoute, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadStaticRoute(nsxClient.Context, routerID, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving static route ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if static route %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == staticRoute.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX static route %s wasn't found", displayName)
	}
}

func testAccNSXStaticRouteCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_static_route" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		routerID := rs.Primary.Attributes["logical_router_id"]
		staticRoute, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadStaticRoute(nsxClient.Context, routerID, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving static route ID %s. Error: %v", resourceID, err)
		}

		if displayName == staticRoute.DisplayName {
			return fmt.Errorf("NSX static route %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXStaticRouteImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccResourceStaticRouteName]
	if !ok {
		return "", fmt.Errorf("NSX static route resource %s not found in resources", testAccResourceStaticRouteName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX static route resource ID not set in resources ")
	}
	routerID := rs.Primary.Attributes["logical_router_id"]
	if routerID == "" {
		return "", fmt.Errorf("NSX static route routerID not set in resources ")
	}
	return fmt.Sprintf("%s/%s", routerID, resourceID), nil
}

func testAccNSXStaticRoutePreConditionTemplate(tier string, edgeClusterName string, tzName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_logical_%s_router" "rtr1" {
  display_name    = "static route test"
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}

data "nsxt_transport_zone" "tz1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "ls1" {
  display_name      = "%s"
  admin_state       = "UP"
  replication_mode  = "MTEP"
  vlan              = "0"
  transport_zone_id = "${data.nsxt_transport_zone.tz1.id}"
}

resource "nsxt_logical_port" "port1" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.ls1.id}"
}

resource "nsxt_logical_router_downlink_port" "lrp1" {
  display_name                  = "LRP"
  description                   = "Acceptance Test"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_%s_router.rtr1.id}"
  ip_address                    = "8.0.0.1/24"
}`, edgeClusterName, tier, tzName, testAccResourceStaticRouteHelperName, testAccResourceStaticRouteHelperName, tier)
}

func testAccNSXStaticRouteCreateTemplate(tier string, name string, edgeClusterName string, tzName string) string {
	return testAccNSXStaticRoutePreConditionTemplate(tier, edgeClusterName, tzName) + fmt.Sprintf(`
resource "nsxt_static_route" "test" {
  logical_router_id = "${nsxt_logical_%s_router.rtr1.id}"
  display_name      = "%s"
  description       = "Acceptance Test"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  network = "4.4.4.0/24"

  next_hop {
    ip_address              = "8.0.0.10"
    administrative_distance = "1"
    logical_router_port_id  = "${nsxt_logical_router_downlink_port.lrp1.id}"
  }
}`, tier, name)
}

func testAccNSXStaticRouteUpdateTemplate(tier string, name string, edgeClusterName string, tzName string) string {
	return testAccNSXStaticRoutePreConditionTemplate(tier, edgeClusterName, tzName) + fmt.Sprintf(`
resource "nsxt_static_route" "test" {
  logical_router_id = "${nsxt_logical_%s_router.rtr1.id}"
  display_name      = "%s"
  description       = "Acceptance Test Update"
  network           = "5.5.5.0/24"

  next_hop {
    ip_address              = "8.0.0.10"
    administrative_distance = "1"
    logical_router_port_id  = "${nsxt_logical_router_downlink_port.lrp1.id}"
  }

  next_hop {
    ip_address              = "2.2.2.2"
    administrative_distance = "2"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, tier, name)
}
