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

func TestAccResourceNsxtDhcpServerProfile_basic(t *testing.T) {
	prfName := getAccTestResourceName()
	updatePrfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_server_profile.test"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpServerProfileCheckDestroy(state, updatePrfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpServerProfileCreateTemplate(edgeClusterName, prfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpServerProfileExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXDhcpServerProfileUpdateTemplate(edgeClusterName, updatePrfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpServerProfileExists(updatePrfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePrfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_id"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtDhcpServerProfile_importBasic(t *testing.T) {
	prfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_server_profile.test"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpServerProfileCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpServerProfileCreateTemplate(edgeClusterName, prfName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXDhcpServerProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Dhcp Server Profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Dhcp Server Profile resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.ServicesApi.ReadDhcpProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving Dhcp Server Profile ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if Dhcp Server Profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("Dhcp Server Profile %s wasn't found", displayName)
	}
}

func testAccNSXDhcpServerProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_dhcp_server_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.ServicesApi.ReadDhcpProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving Dhcp Server Profile ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("Dhcp Server Profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXDhcpServerProfileCreateTemplate(edgeClusterName string, name string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_dhcp_server_profile" "test" {
  display_name                = "%s"
  description                 = "Acceptance Test"
  edge_cluster_id             = "${data.nsxt_edge_cluster.EC.id}"
  edge_cluster_member_indexes = [0]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, edgeClusterName, name)
}

func testAccNSXDhcpServerProfileUpdateTemplate(edgeClusterName string, updatedName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_dhcp_server_profile" "test" {
  display_name                = "%s"
  description                 = "Acceptance Test Update"
  edge_cluster_id             = "${data.nsxt_edge_cluster.EC.id}"
  edge_cluster_member_indexes = [0]
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, edgeClusterName, updatedName)
}
