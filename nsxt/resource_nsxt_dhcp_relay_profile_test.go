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

func TestAccResourceNsxtDhcpRelayProfile_basic(t *testing.T) {
	prfName := getAccTestResourceName()
	updatePrfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_relay_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpRelayProfileCheckDestroy(state, updatePrfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpRelayProfileCreateTemplate(prfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayProfileExists(prfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", prfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "server_addresses.#", "1"),
				),
			},
			{
				Config: testAccNSXDhcpRelayProfileUpdateTemplate(updatePrfName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXDhcpRelayProfileExists(updatePrfName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePrfName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "server_addresses.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtDhcpRelayProfile_importBasic(t *testing.T) {
	prfName := getAccTestResourceName()
	testResourceName := "nsxt_dhcp_relay_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXDhcpRelayProfileCheckDestroy(state, prfName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXDhcpRelayProfileCreateTemplate(prfName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXDhcpRelayProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Dhcp Relay Profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Dhcp Relay Profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelayProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving dhcp relay profile ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if dhcp relay profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("Dhcp Relay Profile %s wasn't found", displayName)
	}
}

func testAccNSXDhcpRelayProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_dhcp_relay_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadDhcpRelayProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving dhcp relay profile ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("Dhcp Relay Profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXDhcpRelayProfileCreateTemplate(prfName string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
  display_name     = "%s"
  description      = "Acceptance Test"
  server_addresses = ["1.1.1.1"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, prfName)
}

func testAccNSXDhcpRelayProfileUpdateTemplate(prfUpdatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_dhcp_relay_profile" "test" {
  display_name     = "%s"
  description      = "Acceptance Test Update"
  server_addresses = ["1.1.1.1", "2.2.2.2"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, prfUpdatedName)
}
