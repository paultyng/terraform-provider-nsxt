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

func TestAccResourceNsxtLogicalRouterCentralizedServicePort_basic(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_centralized_service_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()
	routerObj := "nsxt_logical_tier1_router.rtr1.id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterCentralizedServicePortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterCentralizedServicePortCreateTemplate(portName, transportZoneName, edgeClusterName, routerObj),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterCentralizedServicePortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
				),
			},
			{
				Config: testAccNSXLogicalRouterCentralizedServicePortUpdateTemplate(updatePortName, transportZoneName, edgeClusterName, routerObj),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterCentralizedServicePortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "STRICT"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterCentralizedServicePort_onTier0(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_centralized_service_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()
	routerObj := "data.nsxt_logical_tier0_router.tier0rtr.id"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterCentralizedServicePortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterCentralizedServicePortCreateTemplate(portName, transportZoneName, edgeClusterName, routerObj),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterCentralizedServicePortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
				),
			},
			{
				Config: testAccNSXLogicalRouterCentralizedServicePortUpdateTemplate(updatePortName, transportZoneName, edgeClusterName, routerObj),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterCentralizedServicePortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "STRICT"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterCentralizedServicePort_importBasic(t *testing.T) {
	portName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_centralized_service_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()
	routerObj := "nsxt_logical_tier1_router.rtr1.id"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterCentralizedServicePortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterCentralizedServicePortCreateTemplate(portName, transportZoneName, edgeClusterName, routerObj),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterCentralizedServicePort_onTier1(t *testing.T) {
	portName := "test"
	testResourceName := "nsxt_logical_router_centralized_service_port.test"
	transportZoneName := getVlanTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterCentralizedServicePortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterCSPVlanCreateTemplate(transportZoneName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterCentralizedServicePortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
				),
			},
		},
	})
}

func testAccNSXLogicalRouterCentralizedServicePortExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical port resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical port resource ID not set in resources ")
		}

		logicalPort, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterCentralizedServicePort(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical port ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if logical port %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == logicalPort.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical port %s wasn't found", displayName)
	}
}

func testAccNSXLogicalRouterCentralizedServicePortCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_router_centralized_service_port" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		logicalPort, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterCentralizedServicePort(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical port ID %s. Error: %v", resourceID, err)
		}

		if displayName == logicalPort.DisplayName {
			return fmt.Errorf("NSX logical port %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalRouterCentralizedServicePortPreConditionTemplate(transportZoneName string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "tz1" {
  display_name = "%s"
}

data "nsxt_logical_tier0_router" "tier0rtr" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "ls1" {
  display_name      = "test_switch"
  admin_state       = "UP"
  replication_mode  = "MTEP"
  vlan              = "0"
  transport_zone_id = "${data.nsxt_transport_zone.tz1.id}"
}

resource "nsxt_logical_port" "port1" {
  display_name      = "LP"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.ls1.id}"
}

data "nsxt_edge_cluster" "ec" {
  display_name = "%s"
}

resource "nsxt_logical_tier1_router" "rtr1" {
  display_name    = "test_router"
  edge_cluster_id = "${data.nsxt_edge_cluster.ec.id}"
}`, transportZoneName, getTier0RouterName(), edgeClusterName)
}

func testAccNSXLogicalRouterCentralizedServicePortCreateTemplate(portName string, transportZoneName string, edgeClusterName string, routerObj string) string {
	return testAccNSXLogicalRouterCentralizedServicePortPreConditionTemplate(transportZoneName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_centralized_service_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${%s}"
  ip_address                    = "8.0.0.1/24"
  urpf_mode                     = "NONE"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, portName, routerObj)
}

func testAccNSXLogicalRouterCentralizedServicePortUpdateTemplate(portUpdatedName string, transportZoneName string, edgeClusterName string, routerObj string) string {
	return testAccNSXLogicalRouterCentralizedServicePortPreConditionTemplate(transportZoneName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_centralized_service_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test Update"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${%s}"
  ip_address                    = "8.0.0.1/24"
  urpf_mode                     = "STRICT"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, portUpdatedName, routerObj)
}

func testAccNSXLogicalRouterCSPVlanCreateTemplate(transportZoneName string, edgeClusterName string) string {
	return fmt.Sprintf(`

data "nsxt_transport_zone" "tz1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "ls1" {
  display_name      = "test_switch"
  replication_mode  = ""
  admin_state       = "UP"
  vlan              = "1"
  transport_zone_id = "${data.nsxt_transport_zone.tz1.id}"
}

resource "nsxt_logical_port" "port1" {
  display_name      = "LP"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.ls1.id}"
}

data "nsxt_edge_cluster" "ec" {
  display_name = "%s"
}

resource "nsxt_logical_tier1_router" "test" {
  display_name    = "test_router"
  edge_cluster_id = "${data.nsxt_edge_cluster.ec.id}"
}

resource "nsxt_logical_router_centralized_service_port" "test" {
  display_name                  = "test"
  description                   = "Acceptance Test"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_tier1_router.test.id}"
  ip_address                    = "8.0.0.1/24"
  urpf_mode                     = "NONE"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, transportZoneName, edgeClusterName)
}
