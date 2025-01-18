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
	"github.com/vmware/go-vmware-nsxt/manager"
)

var testNsxtDhcpServerIPPoolName = getAccTestResourceName()
var testNsxtDhcpServerIPPoolResourceName = "nsxt_dhcp_server_ip_pool.test"

func TestAccResourceNsxtDhcpServerIPPool_basic(t *testing.T) {
	name := testNsxtDhcpServerIPPoolName
	updatedName := getAccTestResourceName()
	testResourceName := testNsxtDhcpServerIPPoolResourceName
	edgeClusterName := getEdgeClusterName()
	gateway := "1.1.1.1"
	updatedGateway := "1.1.1.8"
	leaseTime := "999999"
	updatedLeaseTime := "1000000"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpServerIPPoolCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpServerIPPoolTemplate(edgeClusterName, name, gateway, leaseTime),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpIPPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_dhcp_server_id"),
					resource.TestCheckResourceAttr(testResourceName, "gateway_ip", gateway),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", leaseTime),
					resource.TestCheckResourceAttr(testResourceName, "error_threshold", "99"),
					resource.TestCheckResourceAttr(testResourceName, "warning_threshold", "70"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_option_121.0.network"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.0.next_hop", "5.2.2.1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.code", "119"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.values.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.start", "1.1.1.20"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.end", "1.1.1.40"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXDhcpServerIPPoolTemplate(edgeClusterName, updatedName, updatedGateway, updatedLeaseTime),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpIPPoolExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_dhcp_server_id"),
					resource.TestCheckResourceAttr(testResourceName, "gateway_ip", updatedGateway),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", updatedLeaseTime),
					resource.TestCheckResourceAttr(testResourceName, "error_threshold", "99"),
					resource.TestCheckResourceAttr(testResourceName, "warning_threshold", "70"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_option_121.0.network"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.0.next_hop", "5.2.2.1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.code", "119"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.values.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.start", "1.1.1.20"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.end", "1.1.1.40"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtDhcpServerIPPool_noOpts(t *testing.T) {
	name := testNsxtDhcpServerIPPoolName
	updatedName := getAccTestResourceName()
	testResourceName := testNsxtDhcpServerIPPoolResourceName
	edgeClusterName := getEdgeClusterName()
	start1 := "1.1.1.100"
	end1 := "1.1.1.120"
	start2 := "1.1.1.140"
	end2 := "1.1.1.160"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpServerIPPoolCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpServerIPPoolNoOptsTemplate(edgeClusterName, name, start1, end1, start2, end2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpIPPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.start", start1),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.end", end1),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.1.start", start2),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.1.end", end2),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNSXDhcpServerIPPoolNoOptsTemplate(edgeClusterName, updatedName, start2, end2, start1, end1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpIPPoolExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.start", start2),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.0.end", end2),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.1.start", start1),
					resource.TestCheckResourceAttr(testResourceName, "ip_range.1.end", end1),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtDhcpServerIPPool_Import(t *testing.T) {
	name := testNsxtDhcpServerIPPoolName
	testResourceName := testNsxtDhcpServerIPPoolResourceName
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNATRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpServerIPPoolNoOptsTemplate(edgeClusterName, name, "2.0.0.3", "2.0.0.16", "2.0.0.140", "2.0.0.156"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXNDhcpIPPoolImporterGetID,
			},
		},
	})
}

func testAccNSXNDhcpIPPoolImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testNsxtDhcpServerIPPoolResourceName]
	if !ok {
		return "", fmt.Errorf("DHCP IP Pool %s not found in resources", testNsxtDhcpServerIPPoolName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("DHCP IP Pool resource ID not set in resources")
	}
	serverID := rs.Primary.Attributes["logical_dhcp_server_id"]
	if serverID == "" {
		return "", fmt.Errorf("DHCP IP Pool logical_dhcp_server_id not set in resources")
	}
	return fmt.Sprintf("%s/%s", serverID, resourceID), nil
}

func findAccNSXDhcpIPPool(resourceID string) (*manager.DhcpIpPool, error) {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

	servers, responseCode, err := nsxClient.ServicesApi.ListDhcpServers(nsxClient.Context, nil)
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving Dhcp Servers: %v", err)
	}

	if responseCode.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code %d while retrieving Dhcp Servers", responseCode.StatusCode)
	}

	serverID := ""
	for _, server := range servers.Results {
		if server.DisplayName == "Acceptance Test" {
			serverID = server.Id
			break
		}
	}

	if serverID == "" {
		return nil, nil
	}

	pool, responseCode, err := nsxClient.ServicesApi.ReadDhcpIpPool(nsxClient.Context, serverID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving Dhcp IP Pool %s, error %v", resourceID, err)
	}

	if responseCode.StatusCode == http.StatusOK {
		return &pool, nil
	}

	if responseCode.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	return nil, fmt.Errorf("Unexpected status code %d when looking for Dhcp IP pool %s", responseCode.StatusCode, resourceID)
}

func testAccNSXDhcpIPPoolExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Dhcp Server IP Pool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Dhcp Server IP Pool resource ID not set in resources ")
		}

		pool, err := findAccNSXDhcpIPPool(resourceID)
		if err != nil {
			return err
		}

		if pool == nil {
			return fmt.Errorf("Dhcp Server IP Pool %s wasn't found", displayName)
		}

		return nil
	}
}

func testAccNSXDhcpServerIPPoolCheckDestroy(state *terraform.State, displayName string) error {
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_dhcp_server_ip_pool" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		pool, err := findAccNSXDhcpIPPool(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving Dhcp Server IP Pool ID %s. Error: %v", resourceID, err)
		}

		if pool != nil && pool.DisplayName == displayName {
			return fmt.Errorf("Dhcp Server IP Pool %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXCreateDhcpIPPoolPrerequisites(edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_dhcp_server_profile" "PRF" {
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}

resource "nsxt_logical_dhcp_server" "DS" {
    display_name     = "Acceptance Test"
    description      = "Acceptance Test"
    dhcp_profile_id  = "${nsxt_dhcp_server_profile.PRF.id}"
    dhcp_server_ip   = "1.1.1.10/24"
    gateway_ip       = "1.1.1.1"
}`, edgeClusterName)
}

func testAccNSXDhcpServerIPPoolTemplate(edgeClusterName string, name string, ip string, lease string) string {
	return testAccNSXCreateDhcpIPPoolPrerequisites(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_dhcp_server_ip_pool" "test" {
  display_name           = "%s"
  description            = "test"
  logical_dhcp_server_id = "${nsxt_logical_dhcp_server.DS.id}"
  gateway_ip             = "%s"
  lease_time             = %s
  error_threshold        = 99
  warning_threshold      = 70

  dhcp_option_121 {
    network  = "5.5.5.0/24"
    next_hop = "5.2.2.1"
  }

  dhcp_generic_option {
    code = "119"
    values = ["abc"]
  }

  ip_range {
      start = "1.1.1.20"
      end   = "1.1.1.40"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, ip, lease)
}

func testAccNSXDhcpServerIPPoolNoOptsTemplate(edgeClusterName string, name string, start1 string, end1 string, start2 string, end2 string) string {
	return testAccNSXCreateDhcpIPPoolPrerequisites(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_dhcp_server_ip_pool" "test" {
  display_name           = "%s"
  logical_dhcp_server_id = "${nsxt_logical_dhcp_server.DS.id}"

  ip_range {
      start = "%s"
      end   = "%s"
  }

  ip_range {
      start = "%s"
      end   = "%s"
  }

}`, name, start1, end1, start2, end2)
}
