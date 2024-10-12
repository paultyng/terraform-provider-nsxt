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

func TestAccResourceNsxtLogicalDhcpPort_basic(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_dhcp_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalPortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalDhcpPortCreateTemplate(portName, transportZoneName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_switch_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_server_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLogicalDhcpPortUpdateTemplate(updatePortName, transportZoneName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_switch_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_server_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalDhcpPort_importBasic(t *testing.T) {
	portName := getAccTestResourceName()
	testResourceName := "nsxt_logical_dhcp_port.test"
	transportZoneName := getOverlayTransportZoneName()
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalPortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalDhcpPortCreateTemplate(portName, transportZoneName, edgeClusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXDHCPServerCreateForPort(edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_dhcp_server_profile" "test" {
  display_name    = "dhcp_server"
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}
resource "nsxt_logical_dhcp_server" "test" {
  display_name     = "server1"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.test.id}"
  dhcp_server_ip   = "1.1.1.1/24"
  gateway_ip       = "1.1.1.2"
}`, edgeClusterName)
}

func testAccNSXLogicalDhcpPortCreateTemplate(portName string, transportZoneName string, edgeClusterName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + testAccNSXDHCPServerCreateForPort(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.test.id}"
  dhcp_server_id    = "${nsxt_logical_dhcp_server.test.id}"

  tag {
  	scope = "scope1"
    tag   = "tag1"
  }
}`, portName)
}

func testAccNSXLogicalDhcpPortUpdateTemplate(portUpdatedName string, transportZoneName string, edgeClusterName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + testAccNSXDHCPServerCreateForPort(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test Update"
  logical_switch_id = "${nsxt_logical_switch.test.id}"
  dhcp_server_id    = "${nsxt_logical_dhcp_server.test.id}"

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
