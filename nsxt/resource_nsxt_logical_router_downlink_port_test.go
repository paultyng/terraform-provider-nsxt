// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtLogicalRouterDownlinkPort_basic(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_downlink_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterDownlinkPortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterDownlinkPortCreateTemplate(portName, transportZoneName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterDownlinkPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
					resource.TestCheckResourceAttrSet(testResourceName, "mac_address"),
				),
			},
			{
				Config: testAccNSXLogicalRouterDownlinkPortUpdateTemplate(updatePortName, transportZoneName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterDownlinkPortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "urpf_mode", "STRICT"),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", "8.0.0.1/24"),
					resource.TestCheckResourceAttrSet(testResourceName, "mac_address"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterDownlinkPort_withRelay(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_downlink_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	// resource type changed to DhcpRelayService in NSX 2.5
	resourceType := "DhcpRelayService"
	// this is needed to init the version
	testAccNSXVersion(t, "2.2.0")
	if util.NsxVersionLower("2.5.0") {
		resourceType = "LogicalService"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterDownlinkPortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterDownlinkPortCreateWithRelayTemplate(portName, transportZoneName, edgeClusterName, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterDownlinkPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.0.target_type", resourceType),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.0.target_display_name", "srv"),
				),
			},
			{
				Config: testAccNSXLogicalRouterDownlinkPortUpdateWithRelayTemplate(updatePortName, transportZoneName, edgeClusterName, resourceType),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterDownlinkPortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_switch_port_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.0.target_type", resourceType),
					resource.TestCheckResourceAttr(testResourceName, "service_binding.0.target_display_name", "srv"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterDownlinkPort_importBasic(t *testing.T) {
	portName := getAccTestResourceName()
	testResourceName := "nsxt_logical_router_downlink_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterDownlinkPortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterDownlinkPortCreateTemplate(portName, transportZoneName, edgeClusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalRouterDownlinkPortExists(displayName string, resourceName string) resource.TestCheckFunc {
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

		logicalPort, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterDownLinkPort(nsxClient.Context, resourceID)
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

func testAccNSXLogicalRouterDownlinkPortCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_router_downlink_port" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		logicalPort, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterDownLinkPort(nsxClient.Context, resourceID)
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

func testAccNSXLogicalRouterDownlinkPortPreConditionTemplate(transportZoneName string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "tz1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "ls1" {
  display_name     = "downlink_test_switch"
  admin_state      = "UP"
  replication_mode = "MTEP"
  vlan             = "0"
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
  display_name    = "downlink_test_router"
  edge_cluster_id = "${data.nsxt_edge_cluster.ec.id}"
}`, transportZoneName, edgeClusterName)
}

func testAccNSXLogicalRouterDownlinkPortRelayTemplate() string {
	return `
resource "nsxt_dhcp_relay_profile" "drp1" {
  display_name     = "prf"
  server_addresses = ["1.1.1.1"]
}

resource "nsxt_dhcp_relay_service" "drs1" {
	display_name = "srv"
	description = "Acceptance Test"
	dhcp_relay_profile_id = "${nsxt_dhcp_relay_profile.drp1.id}"
}`
}

func testAccNSXLogicalRouterDownlinkPortCreateTemplate(portName string, transportZoneName string, edgeClusterName string) string {
	return testAccNSXLogicalRouterDownlinkPortPreConditionTemplate(transportZoneName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_downlink_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_tier1_router.rtr1.id}"
  ip_address                    = "8.0.0.1/24"
  urpf_mode                     = "NONE"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, portName)
}

func testAccNSXLogicalRouterDownlinkPortUpdateTemplate(portUpdatedName string, transportZoneName string, edgeClusterName string) string {
	return testAccNSXLogicalRouterDownlinkPortPreConditionTemplate(transportZoneName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_downlink_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test Update"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_tier1_router.rtr1.id}"
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
}`, portUpdatedName)
}

func testAccNSXLogicalRouterDownlinkPortCreateWithRelayTemplate(portName string, transportZoneName string, edgeClusterName string, resourceType string) string {
	return testAccNSXLogicalRouterDownlinkPortPreConditionTemplate(transportZoneName, edgeClusterName) +
		testAccNSXLogicalRouterDownlinkPortRelayTemplate() +
		fmt.Sprintf(`
resource "nsxt_logical_router_downlink_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_tier1_router.rtr1.id}"
  ip_address                    = "8.0.0.1/24"

  service_binding {
    target_id   = "${nsxt_dhcp_relay_service.drs1.id}"
	target_type = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, portName, resourceType)
}

func testAccNSXLogicalRouterDownlinkPortUpdateWithRelayTemplate(portUpdatedName string, transportZoneName string, edgeClusterName string, resourceType string) string {
	return testAccNSXLogicalRouterDownlinkPortPreConditionTemplate(transportZoneName, edgeClusterName) +
		testAccNSXLogicalRouterDownlinkPortRelayTemplate() +
		fmt.Sprintf(`
resource "nsxt_logical_router_downlink_port" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test Update"
  linked_logical_switch_port_id = "${nsxt_logical_port.port1.id}"
  logical_router_id             = "${nsxt_logical_tier1_router.rtr1.id}"
  ip_address                    = "8.0.0.1/24"

  service_binding {
    target_id   = "${nsxt_dhcp_relay_service.drs1.id}"
    target_type = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, portUpdatedName, resourceType)
}
