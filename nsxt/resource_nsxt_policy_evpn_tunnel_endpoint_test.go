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

var accTestEvpnTunnelEndpointHelperName = getAccTestResourceName()

var accTestPolicyEvpnTunnelEndpointCreateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"mtu":           "1500",
	"local_address": "10.2.0.12",
}

var accTestPolicyEvpnTunnelEndpointUpdateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform update",
	"mtu":           "1200",
	"local_address": "10.2.0.15",
}

func TestAccResourceNsxtPolicyEvpnTunnelEndpoint_basic(t *testing.T) {
	testResourceName := "nsxt_policy_evpn_tunnel_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnTunnelEndpointCheckDestroy(state, accTestPolicyEvpnTunnelEndpointUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnTunnelEndpointBasic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnTunnelEndpointExists(accTestPolicyEvpnTunnelEndpointCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyEvpnTunnelEndpointCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyEvpnTunnelEndpointCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyEvpnTunnelEndpointCreateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "mtu", accTestPolicyEvpnTunnelEndpointCreateAttributes["mtu"]),
					resource.TestCheckResourceAttrSet(testResourceName, "external_interface_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_node_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyEvpnTunnelEndpointBasic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnTunnelEndpointExists(accTestPolicyEvpnTunnelEndpointUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyEvpnTunnelEndpointUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyEvpnTunnelEndpointUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyEvpnTunnelEndpointUpdateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "mtu", accTestPolicyEvpnTunnelEndpointUpdateAttributes["mtu"]),
					resource.TestCheckResourceAttrSet(testResourceName, "external_interface_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_node_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyEvpnTunnelEndpoint_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_evpn_tunnel_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnTunnelEndpointCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnTunnelEndpointBasic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyEvpnTunnelEndpointIDGenerator(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyEvpnTunnelEndpointExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Evpn Tunnel Endpoint resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Evpn Tunnel Endpoint resource ID not set in resources")
		}
		interfacePath := rs.Primary.Attributes["external_interface_path"]
		_, gwID, localeServiceID, _ := parseGatewayInterfacePolicyPath(interfacePath)

		exists, err := resourceNsxtPolicyEvpnTunnelEndpointExists(connector, gwID, localeServiceID, resourceID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Evpn Tunnel Endpoint %s does not exist on backend", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyEvpnTunnelEndpointCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_evpn_tunnel_endpoint" {
			continue
		}

		resourceID := rs.Primary.ID
		interfacePath := rs.Primary.Attributes["external_interface_path"]
		_, gwID, localeServiceID, _ := parseGatewayInterfacePolicyPath(interfacePath)

		exists, err := resourceNsxtPolicyEvpnTunnelEndpointExists(connector, gwID, localeServiceID, resourceID)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy Evpn Tunnel Endpoint %s still exists on backend", resourceID)
		}

	}
	return nil
}

func testAccEvpnTunnelEndpointPrerequisites() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) +
		testAccNsxtPolicyVniPoolConfigReadTemplate() + fmt.Sprintf(`
data "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
}

data "nsxt_policy_edge_node" "test" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  member_index      = 0
}

resource "nsxt_policy_vlan_segment" "test" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "%s"
  vlan_ids            = [11]
  subnet {
    cidr = "10.2.2.2/24"
  }
}

resource "nsxt_policy_tier0_gateway_interface" "test" {
  display_name   = "%s"
  type           = "EXTERNAL"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test.path
  edge_node_path = data.nsxt_policy_edge_node.test.path
  subnets        = ["10.2.2.14/24"]
}

resource "nsxt_policy_evpn_config" "test" {
  display_name  = "%s"
  gateway_path  = nsxt_policy_tier0_gateway.test.path
  vni_pool_path = data.nsxt_policy_vni_pool.test.path
  mode          = "INLINE"
}
`, getVlanTransportZoneName(), accTestEvpnTunnelEndpointHelperName, accTestEvpnTunnelEndpointHelperName, accTestEvpnTunnelEndpointHelperName)
}

func testAccNsxtPolicyEvpnTunnelEndpointBasic(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyEvpnTunnelEndpointCreateAttributes
	} else {
		attrMap = accTestPolicyEvpnTunnelEndpointUpdateAttributes
	}

	return testAccEvpnTunnelEndpointPrerequisites() + fmt.Sprintf(`

resource "nsxt_policy_evpn_tunnel_endpoint" "test" {
  display_name = "%s"
  description  = "%s"

  external_interface_path = nsxt_policy_tier0_gateway_interface.test.path
  edge_node_path          = data.nsxt_policy_edge_node.test.path

  local_address  = "%s"
  mtu            = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["local_address"], attrMap["mtu"])
}

func testAccNSXPolicyEvpnTunnelEndpointIDGenerator(testResourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[testResourceName]
		if !ok {
			return "", fmt.Errorf("NSX Policy resource %s not found in resources", testResourceName)
		}
		id := rs.Primary.ID
		interfacePath := rs.Primary.Attributes["external_interface_path"]
		if interfacePath == "" {
			return "", fmt.Errorf("NSX Policy Interface Path not set in resources ")
		}

		_, gwID, localeServiceID, interfaceID := parseGatewayInterfacePolicyPath(interfacePath)

		return fmt.Sprintf("%s/%s/%s/%s", gwID, localeServiceID, interfaceID, id), nil
	}
}
