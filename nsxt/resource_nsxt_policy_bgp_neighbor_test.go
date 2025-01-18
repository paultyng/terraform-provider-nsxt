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

var accTestPolicyBgpNeighborConfigCreateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform created",
	"allow_as_in":           "true",
	"graceful_restart_mode": "HELPER_ONLY",
	"hold_down_time":        "900",
	"keep_alive_time":       "200",
	"maximum_hop_limit":     "3",
	"neighbor_address":      "12.12.12.12",
	"password":              "test-create",
	"remote_as_num":         "12000012",
}

var accTestPolicyBgpNeighborConfigUpdateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform updated",
	"allow_as_in":           "false",
	"graceful_restart_mode": "GR_AND_HELPER",
	"hold_down_time":        "950",
	"keep_alive_time":       "250",
	"maximum_hop_limit":     "5",
	"neighbor_address":      "12.12.12.13",
	"password":              "test-update",
	"remote_as_num":         "12.013",
}

func TestAccResourceNsxtPolicyBgpNeighbor_basic(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"
	subnet := "1.1.12.2/24"
	sourceAddress := "1.1.12.2"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, accTestPolicyBgpNeighborConfigCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborTemplate(true, subnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyBgpNeighborConfigCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyBgpNeighborConfigCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "allow_as_in", accTestPolicyBgpNeighborConfigCreateAttributes["allow_as_in"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpNeighborConfigCreateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "hold_down_time", accTestPolicyBgpNeighborConfigCreateAttributes["hold_down_time"]),
					resource.TestCheckResourceAttr(testResourceName, "keep_alive_time", accTestPolicyBgpNeighborConfigCreateAttributes["keep_alive_time"]),
					resource.TestCheckResourceAttr(testResourceName, "maximum_hop_limit", accTestPolicyBgpNeighborConfigCreateAttributes["maximum_hop_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "neighbor_address", accTestPolicyBgpNeighborConfigCreateAttributes["neighbor_address"]),
					resource.TestCheckResourceAttr(testResourceName, "remote_as_num", accTestPolicyBgpNeighborConfigCreateAttributes["remote_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.0", sourceAddress),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "password"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyBgpNeighborTemplate(false, subnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyBgpNeighborConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyBgpNeighborConfigUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "allow_as_in", accTestPolicyBgpNeighborConfigUpdateAttributes["allow_as_in"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpNeighborConfigUpdateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "hold_down_time", accTestPolicyBgpNeighborConfigUpdateAttributes["hold_down_time"]),
					resource.TestCheckResourceAttr(testResourceName, "keep_alive_time", accTestPolicyBgpNeighborConfigUpdateAttributes["keep_alive_time"]),
					resource.TestCheckResourceAttr(testResourceName, "maximum_hop_limit", accTestPolicyBgpNeighborConfigUpdateAttributes["maximum_hop_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "neighbor_address", accTestPolicyBgpNeighborConfigUpdateAttributes["neighbor_address"]),
					resource.TestCheckResourceAttr(testResourceName, "remote_as_num", accTestPolicyBgpNeighborConfigUpdateAttributes["remote_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.0", sourceAddress),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "password"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_globalManager(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"
	subnet := "1.1.12.2/24"
	sourceAddress := "1.1.12.2"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyGlobalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, accTestPolicyBgpNeighborConfigCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborGMTemplate(true, subnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyBgpNeighborConfigCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyBgpNeighborConfigCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "allow_as_in", accTestPolicyBgpNeighborConfigCreateAttributes["allow_as_in"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpNeighborConfigCreateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "hold_down_time", accTestPolicyBgpNeighborConfigCreateAttributes["hold_down_time"]),
					resource.TestCheckResourceAttr(testResourceName, "keep_alive_time", accTestPolicyBgpNeighborConfigCreateAttributes["keep_alive_time"]),
					resource.TestCheckResourceAttr(testResourceName, "maximum_hop_limit", accTestPolicyBgpNeighborConfigCreateAttributes["maximum_hop_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "neighbor_address", accTestPolicyBgpNeighborConfigCreateAttributes["neighbor_address"]),
					resource.TestCheckResourceAttr(testResourceName, "remote_as_num", accTestPolicyBgpNeighborConfigCreateAttributes["remote_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.0", sourceAddress),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "password"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyBgpNeighborGMTemplate(false, subnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyBgpNeighborConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyBgpNeighborConfigUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "allow_as_in", accTestPolicyBgpNeighborConfigUpdateAttributes["allow_as_in"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpNeighborConfigUpdateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "hold_down_time", accTestPolicyBgpNeighborConfigUpdateAttributes["hold_down_time"]),
					resource.TestCheckResourceAttr(testResourceName, "keep_alive_time", accTestPolicyBgpNeighborConfigUpdateAttributes["keep_alive_time"]),
					resource.TestCheckResourceAttr(testResourceName, "maximum_hop_limit", accTestPolicyBgpNeighborConfigUpdateAttributes["maximum_hop_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "neighbor_address", accTestPolicyBgpNeighborConfigUpdateAttributes["neighbor_address"]),
					resource.TestCheckResourceAttr(testResourceName, "remote_as_num", accTestPolicyBgpNeighborConfigUpdateAttributes["remote_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "source_addresses.0", sourceAddress),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "password"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_minimalistic(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, accTestPolicyBgpNeighborConfigCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_subConfig(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, "tfbgp")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborSubConfigCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.interval", "1000"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.multiple", "4"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.address_family", "IPV4"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.maximum_routes", "20"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.1.address_family", "L2VPN_EVPN"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.1.maximum_routes", "20"),
				),
			},
			{
				Config: testAccNsxtPolicyBgpNeighborSubConfigUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.interval", "2000"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.multiple", "3"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.address_family", "IPV4"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.maximum_routes", "20"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_subConfigSingleRoute(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, "tfbgp")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborSubConfigCreateSingleRouteFilter(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.interval", "2000"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.multiple", "3"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.address_family", "IPV4"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_subConfigPrefixList(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, "tfbgp")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborSubConfigCreatePrefixList(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyBgpNeighborExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.interval", "2000"),
					resource.TestCheckResourceAttr(testResourceName, "bfd_config.0.multiple", "3"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_filtering.0.address_family", "IPV4"),
					resource.TestCheckResourceAttrSet(testResourceName, "route_filtering.0.in_route_filter"),
					resource.TestCheckResourceAttrSet(testResourceName, "route_filtering.0.out_route_filter"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_importGlobalManager(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyGlobalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborGMImportTemplate(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyBgpNeighborImporterGetIDs,
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpNeighbor_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_bgp_neighbor.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyBgpNeighborCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpNeighborMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyBgpNeighborImporterGetIDs,
			},
		},
	})
}

func testAccNSXPolicyBgpNeighborImporterGetIDs(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["nsxt_policy_bgp_neighbor.test"]
	if !ok {
		return "", fmt.Errorf("NSX Policy BGP Neighbor resource %s not found in resources", "nsxt_policy_bgp_neighbor.test")
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy BGP Neighbor resource ID not set in resources ")
	}
	bgpPath := rs.Primary.Attributes["bgp_path"]
	if bgpPath == "" {
		return "", fmt.Errorf("NSX Policy BGP Neighbor bgp_path not set in resources")
	}
	t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
	return fmt.Sprintf("%s/%s/%s", t0ID, serviceID, resourceID), nil
}

func testAccNsxtPolicyBgpNeighborExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy BgpNeighbor resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy BgpNeighbor resource ID not set in resources")
		}

		bgpPath := rs.Primary.Attributes["bgp_path"]
		t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)

		exists, err := resourceNsxtPolicyBgpNeighborExists(t0ID, serviceID, resourceID, testAccIsGlobalManager(), connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("BgpNeighbor ID %s does not exist on backend", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyBgpNeighborCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_bgp_neighbor" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		bgpPath := rs.Primary.Attributes["bgp_path"]
		t0ID, serviceID := resourceNsxtPolicyBgpNeighborParseIDs(bgpPath)
		exists, err := resourceNsxtPolicyBgpNeighborExists(t0ID, serviceID, resourceID, testAccIsGlobalManager(), connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy BgpNeighbor %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyBgpNeighborTemplate(createFlow bool, subnet string) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyBgpNeighborConfigCreateAttributes
	} else {
		attrMap = accTestPolicyBgpNeighborConfigUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

data "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
}

resource "nsxt_policy_vlan_segment" "test" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "Acceptance Test"
  vlan_ids            = [11]
  subnet {
      cidr = "10.2.2.2/24"
  }
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_tier0_gateway_interface" "test" {
  display_name   = "terraformt0gwintf"
  type           = "EXTERNAL"
  description    = "Acceptance Test"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test.path
  subnets        = ["%s"]
}

resource "nsxt_policy_bgp_neighbor" "test" {
  display_name          = "%s"
  description           = "%s"
  bgp_path              = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  allow_as_in           = %s
  graceful_restart_mode = "%s"
  hold_down_time        = %s
  keep_alive_time       = %s
  maximum_hop_limit     = %s
  neighbor_address      = "%s"
  remote_as_num         = "%s"
  password              = "%s"
  source_addresses      = nsxt_policy_tier0_gateway_interface.test.ip_addresses

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, getEdgeClusterName(), getVlanTransportZoneName(), subnet, attrMap["display_name"], attrMap["description"], attrMap["allow_as_in"], attrMap["graceful_restart_mode"], attrMap["hold_down_time"], attrMap["keep_alive_time"], attrMap["maximum_hop_limit"], attrMap["neighbor_address"], attrMap["remote_as_num"], attrMap["password"])
}

func testAccNsxtPolicyBgpNeighborMinimalistic() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  display_name     = "%s"
  neighbor_address = "%s"
  remote_as_num    = "%s"

}
`, getEdgeClusterName(), accTestPolicyBgpNeighborConfigCreateAttributes["display_name"], accTestPolicyBgpNeighborConfigCreateAttributes["neighbor_address"], accTestPolicyBgpNeighborConfigCreateAttributes["remote_as_num"])
}

func testAccNsxtPolicyBgpNeighborSubConfigCreate() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  display_name     = "tfbgp"
  neighbor_address = "12.12.12.12"
  remote_as_num    = "60000"

  bfd_config {
    enabled  = true
    interval = 1000
    multiple = 4
  }

  route_filtering {
    address_family = "IPV4"
    maximum_routes = 20
  }

  route_filtering {
    address_family = "L2VPN_EVPN"
    maximum_routes = 20
  }
}`, getEdgeClusterName())
}

func testAccNsxtPolicyBgpNeighborSubConfigUpdate() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  display_name     = "tfbgp"
  neighbor_address = "12.12.12.12"
  remote_as_num    = "60000"

  bfd_config {
    enabled  = false
    interval = 2000
    multiple = 3
  }

  route_filtering {
    address_family = "IPV4"
    maximum_routes = 20
  }

}`, getEdgeClusterName())
}

func testAccNsxtPolicyBgpNeighborSubConfigCreateSingleRouteFilter() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  display_name     = "tfbgp"
  neighbor_address = "12.12.12.12"
  remote_as_num    = "60000"

  bfd_config {
    enabled  = false
    interval = 2000
    multiple = 3
  }

  route_filtering {
    address_family = "IPV4"
  }

}`, getEdgeClusterName())
}

func testAccNsxtPolicyBgpNeighborSubConfigCreatePrefixList() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  bgp_config {
    local_as_num    = "60000"
    multipath_relax = true
    route_aggregation {
      prefix = "12.12.12.0/24"
    }
  }
}

resource "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "prefix_list"
  gateway_path = nsxt_policy_tier0_gateway.test.path

  prefix {
  	action  = "DENY"
  	ge      = "20"
  	le      = "23"
  	network = "4.4.0.0/20"
  }
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_tier0_gateway.test.bgp_config.0.path
  display_name     = "tfbgp"
  neighbor_address = "12.12.12.12"
  remote_as_num    = "60000"

  bfd_config {
    enabled  = false
    interval = 2000
    multiple = 3
  }

  route_filtering {
    address_family  = "IPV4"
    in_route_filter = nsxt_policy_gateway_prefix_list.test.path
    out_route_filter = nsxt_policy_gateway_prefix_list.test.path
  }
}`, getEdgeClusterName())
}

func testAccNsxtPolicyBgpNeighborGMTemplate(createFlow bool, subnet string) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyBgpNeighborConfigCreateAttributes
	} else {
		attrMap = accTestPolicyBgpNeighborConfigUpdateAttributes
	}
	return testAccNsxtGlobalPolicyEdgeClusterReadTemplate() + testAccNSXGlobalPolicyTransportZoneReadTemplate(true, false) + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "Acceptance Test"
  vlan_ids            = [11]
  subnet {
      cidr = "10.2.2.2/24"
  }
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  locale_service {
    edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  }
}

resource "nsxt_policy_bgp_config" "test" {
  gateway_path    = nsxt_policy_tier0_gateway.test.path
  site_path       = data.nsxt_policy_site.test.path
  local_as_num    = "60000"
  multipath_relax = true
  route_aggregation {
    prefix = "12.12.12.0/24"
  }
}

resource "nsxt_policy_tier0_gateway_interface" "test" {
  display_name   = "terraformt0gwintf"
  type           = "EXTERNAL"
  description    = "Acceptance Test"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test.path
  subnets        = ["%s"]
  site_path      = data.nsxt_policy_site.test.path
}

resource "nsxt_policy_bgp_neighbor" "test" {
  display_name          = "%s"
  description           = "%s"
  bgp_path              = nsxt_policy_bgp_config.test.path
  allow_as_in           = %s
  graceful_restart_mode = "%s"
  hold_down_time        = %s
  keep_alive_time       = %s
  maximum_hop_limit     = %s
  neighbor_address      = "%s"
  remote_as_num         = "%s"
  password              = "%s"
  source_addresses      = nsxt_policy_tier0_gateway_interface.test.ip_addresses

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, subnet, attrMap["display_name"], attrMap["description"], attrMap["allow_as_in"], attrMap["graceful_restart_mode"], attrMap["hold_down_time"], attrMap["keep_alive_time"], attrMap["maximum_hop_limit"], attrMap["neighbor_address"], attrMap["remote_as_num"], attrMap["password"])
}

func testAccNsxtPolicyBgpNeighborGMImportTemplate() string {
	attrMap := accTestPolicyBgpNeighborConfigCreateAttributes
	return testAccNsxtGlobalPolicyEdgeClusterReadTemplate() + testAccNSXGlobalPolicyTransportZoneReadTemplate(true, false) + fmt.Sprintf(`
resource "nsxt_policy_vlan_segment" "test" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "Acceptance Test"
  vlan_ids            = [11]
  subnet {
      cidr = "10.2.2.2/24"
  }
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "terraformt0gw"
  description       = "Acceptance Test"
  locale_service {
    edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  }
}

resource "nsxt_policy_bgp_config" "test" {
  gateway_path    = nsxt_policy_tier0_gateway.test.path
  site_path       = data.nsxt_policy_site.test.path
  local_as_num    = "60000"
  multipath_relax = true
  route_aggregation {
    prefix = "12.12.12.0/24"
  }
}

resource "nsxt_policy_tier0_gateway_interface" "test" {
  display_name   = "terraformt0gwintf"
  type           = "EXTERNAL"
  description    = "Acceptance Test"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test.path
  subnets        = ["12.12.12.1/24"]
  site_path      = data.nsxt_policy_site.test.path
}

resource "nsxt_policy_bgp_neighbor" "test" {
  bgp_path         = nsxt_policy_bgp_config.test.path
  display_name     = "%s"
  neighbor_address = "%s"
  remote_as_num    = "%s"
  source_addresses = nsxt_policy_tier0_gateway_interface.test.ip_addresses
}`, attrMap["display_name"], attrMap["neighbor_address"], attrMap["remote_as_num"])
}
