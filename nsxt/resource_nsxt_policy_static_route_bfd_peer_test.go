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

var accTestPolicyStaticRouteBfdPeerCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"enabled":      "true",
	"peer_address": "10.12.2.4",
}

var accTestPolicyStaticRouteBfdPeerUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"enabled":      "false",
	"peer_address": "10.12.2.5",
}

func TestAccResourceNsxtPolicyStaticRouteBfdPeer_basic(t *testing.T) {
	testResourceName := "nsxt_policy_static_route_bfd_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteBfdPeerCheckDestroy(state, accTestPolicyStaticRouteBfdPeerUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteBfdPeerTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteBfdPeerExists(accTestPolicyStaticRouteBfdPeerCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyStaticRouteBfdPeerCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyStaticRouteBfdPeerCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyStaticRouteBfdPeerCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyStaticRouteBfdPeerCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "0"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteBfdPeerTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteBfdPeerExists(accTestPolicyStaticRouteBfdPeerUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyStaticRouteBfdPeerUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyStaticRouteBfdPeerUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyStaticRouteBfdPeerUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyStaticRouteBfdPeerUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "0"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyStaticRouteBfdPeerMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyStaticRouteBfdPeerExists(accTestPolicyStaticRouteBfdPeerCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyStaticRouteBfdPeer_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_static_route_bfd_peer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyStaticRouteBfdPeerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyStaticRouteBfdPeerMinimalistic(),
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

func testAccNsxtPolicyStaticRouteBfdPeerExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Gateway Static Route BFD Peer resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Gateway Static Route BFD Peer resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyStaticRouteBfdPeerExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Gateway Static Route BFD Peer %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyStaticRouteBfdPeerCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_static_route_bfd_peer" {
			continue
		}

		resourceID := rs.Primary.ID
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		exists, err := resourceNsxtPolicyStaticRouteBfdPeerExists(gwID, resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Gateway Static Route BFD Peer %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyStaticRouteBfdPeerTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyStaticRouteBfdPeerCreateAttributes
	} else {
		attrMap = accTestPolicyStaticRouteBfdPeerUpdateAttributes
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
data "nsxt_policy_bfd_profile" "test" {
  display_name = "default"
}

resource "nsxt_policy_static_route_bfd_peer" "test" {
  gateway_path     = nsxt_policy_tier0_gateway.test.path
  bfd_profile_path = data.nsxt_policy_bfd_profile.test.path

  display_name = "%s"
  description  = "%s"
  enabled      = %s
  peer_address = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["peer_address"])
}

func testAccNsxtPolicyStaticRouteBfdPeerMinimalistic() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
data "nsxt_policy_bfd_profile" "test" {
  display_name = "default"
}

resource "nsxt_policy_static_route_bfd_peer" "test" {
  gateway_path     = nsxt_policy_tier0_gateway.test.path
  bfd_profile_path = data.nsxt_policy_bfd_profile.test.path

  display_name = "%s"
  peer_address = "%s"
}`, accTestPolicyStaticRouteBfdPeerUpdateAttributes["display_name"], accTestPolicyStaticRouteBfdPeerUpdateAttributes["peer_address"])
}
