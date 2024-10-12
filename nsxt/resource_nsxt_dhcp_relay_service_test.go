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

func TestAccResourceNsxtDhcpRelayService_basic(t *testing.T) {
	prfName := getAccTestResourceName()
	updatePrfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_relay_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpRelayServiceCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpRelayServiceCreateTemplate(prfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayServiceExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXDhcpRelayServiceUpdateTemplate(updatePrfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayServiceExists(updatePrfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePrfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtDhcpRelayService_importBasic(t *testing.T) {
	prfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_relay_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpRelayServiceCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpRelayServiceCreateTemplate(prfName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXDhcpRelayServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Dhcp Relay Service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Dhcp Relay Service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelay(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving Dhcp Relay Service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if Dhcp Relay Service %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("Dhcp Relay Service %s wasn't found", displayName)
	}
}

func testAccNSXDhcpRelayServiceCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_dhcp_relay_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelay(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving Dhcp Relay Service ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("Dhcp Relay Service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXDhcpRelayServiceCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
  display_name     = "prf"
  server_addresses = ["1.1.1.1"]
}

resource "nsxt_dhcp_relay_service" "test" {
  display_name          = "%s"
  description           = "Acceptance Test"
  dhcp_relay_profile_id = "${nsxt_dhcp_relay_profile.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXDhcpRelayServiceUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
  display_name     = "prf"
  server_addresses = ["1.1.1.1"]
}

resource "nsxt_dhcp_relay_service" "test" {
  display_name          = "%s"
  description           = "Acceptance Test Update"
  dhcp_relay_profile_id = "${nsxt_dhcp_relay_profile.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName)
}
