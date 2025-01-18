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

func TestAccResourceNsxtIpDiscoverySwitchingProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_ip_discovery_switching_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpDiscoverySwitchingProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpDiscoverySwitchingProfileCreateTemplate(name, "true", "true", "true", "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpDiscoverySwitchingProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "vm_tools_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "arp_snooping_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "arp_bindings_limit", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXIpDiscoverySwitchingProfileCreateTemplate(updatedName, "false", "false", "false", "3"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpDiscoverySwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "vm_tools_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "arp_snooping_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "arp_bindings_limit", "3"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXIpDiscoverySwitchingProfileEmptyTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpDiscoverySwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpDiscoverySwitchingProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_ip_discovery_switching_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpDiscoverySwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpDiscoverySwitchingProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXIpDiscoverySwitchingProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX switching profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX switching profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetIpDiscoverySwitchingProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving switching profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if switching profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX switching profile %s wasn't found", displayName)
	}
}

func testAccNSXIpDiscoverySwitchingProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ip_discovery_switching_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetIpDiscoverySwitchingProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving switching profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("NSX switching profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIpDiscoverySwitchingProfileCreateTemplate(name string, vmToolsEnabled string, arpSnoopingEnabled string, dhcpSnoopingEnabled string, arpBindingsLimit string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_discovery_switching_profile" "test" {
  display_name          = "%s"
  description           = "test description"
  vm_tools_enabled      = "%s"
  arp_snooping_enabled  = "%s"
  dhcp_snooping_enabled = "%s"
  arp_bindings_limit    = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, vmToolsEnabled, arpSnoopingEnabled, dhcpSnoopingEnabled, arpBindingsLimit)
}

func testAccNSXIpDiscoverySwitchingProfileEmptyTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_discovery_switching_profile" "test" {
  display_name          = "%s"
}
`, name)
}

func testAccNSXIpDiscoverySwitchingProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_discovery_switching_profile" "test" {
  display_name          = "%s"
}
`, name)
}
