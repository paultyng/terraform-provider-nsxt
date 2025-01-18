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

func TestAccResourceNsxtLogicalDhcpServer_basic(t *testing.T) {
	prfName := getAccTestResourceName()
	updatePrfName := getAccTestResourceName()
	testResourceName := "nsxt_logical_dhcp_server.test"
	edgeClusterName := getEdgeClusterName()
	ip1 := "1.1.1.10/24"
	ip2 := "1.1.1.20"
	ip1upd := "2.1.1.10/24"
	ip2upd := "2.1.1.20"
	ip3 := "1.1.1.21"
	ip4 := "1.1.1.22"
	ip5 := "1.1.1.23"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalDhcpServerCheckDestroy(state, updatePrfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalDhcpServerCreateTemplate(edgeClusterName, prfName, ip1, ip2, ip3, ip4),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalDhcpServerExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_profile_id"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_ip", ip1),
					resource.TestCheckResourceAttr(testResourceName, "gateway_ip", ip2),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.0", ip3),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_option_121.0.network"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.0.next_hop", ip4),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.code", "119"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.values.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLogicalDhcpServerUpdateTemplate(edgeClusterName, updatePrfName, ip1upd, ip2upd, ip3, ip4, ip5),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalDhcpServerExists(updatePrfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePrfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_profile_id"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_ip", ip1upd),
					resource.TestCheckResourceAttr(testResourceName, "gateway_ip", ip2upd),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.code", "119"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.0.values.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalDhcpServer_noOpts(t *testing.T) {
	prfName := getAccTestResourceName()
	testResourceName := "nsxt_logical_dhcp_server.test"
	edgeClusterName := getEdgeClusterName()
	ip1 := "1.1.1.10/24"
	ip2 := "1.1.1.20"
	ip3 := "1.1.1.21"
	ip1upd := "2.1.1.10/24"
	ip2upd := "2.1.1.20"
	ip4 := "1.1.1.22"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalDhcpServerCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalDhcpServerCreateNoOptsTemplate(edgeClusterName, prfName, ip1, ip2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalDhcpServerExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_profile_id"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_ip", ip1),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.0", ip2),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLogicalDhcpServerUpdateNoOptsTemplate(edgeClusterName, prfName, ip1upd, ip2upd, ip3, ip4),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalDhcpServerExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_profile_id"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_ip", ip1upd),
					resource.TestCheckResourceAttr(testResourceName, "gateway_ip", ip2upd),
					resource.TestCheckResourceAttr(testResourceName, "dns_name_servers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_option_121.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_generic_option.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalDhcpServer_importBasic(t *testing.T) {
	prfName := getAccTestResourceName()
	testResourceName := "nsxt_logical_dhcp_server.test"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalDhcpServerCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalDhcpServerCreateTemplate(edgeClusterName, prfName, "1.1.1.1/24", "1.1.1.10", "1.1.1.100", "1.1.1.200"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalDhcpServerExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Logical Dhcp Server resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Logical Dhcp Server resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.ServicesApi.ReadDhcpServer(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving Logical Dhcp Server ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if Logical Dhcp Server %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("Logical Dhcp Server %s wasn't found", displayName)
	}
}

func testAccNSXLogicalDhcpServerCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_dhcp_server" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.ServicesApi.ReadDhcpServer(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving Logical Dhcp Server ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("Logical Dhcp Server %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXDhcpServerProfileCreateForServerTemplate(edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_dhcp_server_profile" "PRF" {
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}`, edgeClusterName)
}

func testAccNSXLogicalDhcpServerCreateTemplate(edgeClusterName string, name string, ip1 string, ip2 string, ip3 string, ip4 string) string {
	return testAccNSXDhcpServerProfileCreateForServerTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_server" "test" {
  display_name     = "%s"
  description      = "Acceptance Test"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.PRF.id}"
  dhcp_server_ip   = "%s"
  gateway_ip       = "%s"
  domain_name      = "abc.com"
  dns_name_servers = ["%s"]

  dhcp_option_121 {
    network  = "5.5.5.0/24"
    next_hop = "%s"
  }

  dhcp_generic_option {
    code = "119"
    values = ["abc"]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, ip1, ip2, ip3, ip4)
}

func testAccNSXLogicalDhcpServerUpdateTemplate(edgeClusterName string, updatedName string, ip1 string, ip2 string, ip3 string, ip4 string, ip5 string) string {
	return testAccNSXDhcpServerProfileCreateForServerTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_server" "test" {
  display_name                = "%s"
  description                 = "Acceptance Test Update"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.PRF.id}"
  dhcp_server_ip   = "%s"
  gateway_ip       = "%s"
  domain_name      = "abc.com"
  dns_name_servers = ["%s", "%s"]

  dhcp_option_121 {
    network  = "5.5.5.0/24"
    next_hop = "%s"
  }

  dhcp_option_121 {
    network  = "6.6.6.0/24"
    next_hop = "%s"
  }

  dhcp_generic_option {
    code = "119"
    values = ["abc", "def"]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName, ip1, ip2, ip3, ip4, ip5, ip5)
}

func testAccNSXLogicalDhcpServerCreateNoOptsTemplate(edgeClusterName string, name string, ip1 string, ip2 string) string {
	return testAccNSXDhcpServerProfileCreateForServerTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_server" "test" {
  display_name     = "%s"
  description      = "Acceptance Test"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.PRF.id}"
  dhcp_server_ip   = "%s"
  domain_name      = "abc.com"
  dns_name_servers = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, ip1, ip2)
}

func testAccNSXLogicalDhcpServerUpdateNoOptsTemplate(edgeClusterName string, updatedName string, ip1 string, ip2 string, ip3 string, ip4 string) string {
	return testAccNSXDhcpServerProfileCreateForServerTemplate(edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_dhcp_server" "test" {
  display_name     = "%s"
  description      = "Acceptance Test Update"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.PRF.id}"
  dhcp_server_ip   = "%s"
  gateway_ip       = "%s"
  domain_name      = "abc.com"
  dns_name_servers = ["%s", "%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName, ip1, ip2, ip3, ip4)
}
