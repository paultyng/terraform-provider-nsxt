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

func TestAccResourceNsxtSwitchSecuritySwitchingProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_switch_security_switching_profile.test"
	limit := "700"
	updatedLimit := "400"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccTestDeprecated(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXSwitchSecuritySwitchingProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSwitchSecuritySwitchingProfileBasicTemplate(name, limit),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSwitchSecuritySwitchingProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "block_non_ip", "true"),
					resource.TestCheckResourceAttr(testResourceName, "block_server_dhcp", "false"),
					resource.TestCheckResourceAttr(testResourceName, "block_client_dhcp", "true"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_whitelist.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_whitelist.0", "01:80:c2:00:00:01"),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.rx_broadcast", limit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.rx_multicast", limit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.tx_broadcast", limit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.tx_multicast", limit),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXSwitchSecuritySwitchingProfileBasicTemplate(updatedName, updatedLimit),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSwitchSecuritySwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "block_non_ip", "true"),
					resource.TestCheckResourceAttr(testResourceName, "block_server_dhcp", "false"),
					resource.TestCheckResourceAttr(testResourceName, "block_client_dhcp", "true"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_whitelist.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_whitelist.0", "01:80:c2:00:00:01"),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.rx_broadcast", updatedLimit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.rx_multicast", updatedLimit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.tx_broadcast", updatedLimit),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.0.tx_multicast", updatedLimit),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXSwitchSecuritySwitchingProfileEmptyTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSwitchSecuritySwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_whitelist.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtSwitchSecuritySwitchingProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_switch_security_switching_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccTestDeprecated(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXSwitchSecuritySwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSwitchSecuritySwitchingProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXSwitchSecuritySwitchingProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
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

		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetSwitchSecuritySwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXSwitchSecuritySwitchingProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_switch_security_switching_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetSwitchSecuritySwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXSwitchSecuritySwitchingProfileBasicTemplate(name string, limit string) string {
	return fmt.Sprintf(`
resource "nsxt_switch_security_switching_profile" "test" {
  display_name          = "%s"
  description           = "test description"
  block_non_ip          = true
  block_client_dhcp     = true
  block_server_dhcp     = false
  bpdu_filter_enabled   = true
  bpdu_filter_whitelist = ["01:80:c2:00:00:01"]

  rate_limits {
    enabled = true
    rx_broadcast = %s
    rx_multicast = %s
    tx_broadcast = %s
    tx_multicast = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, limit, limit, limit, limit)
}
func testAccNSXSwitchSecuritySwitchingProfileEmptyTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_switch_security_switching_profile" "test" {
  display_name          = "%s"
}
`, name)
}

func testAccNSXSwitchSecuritySwitchingProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_switch_security_switching_profile" "test" {
  display_name = "%s"
  description  = "test description"
  rate_limits {
    enabled = false
  }
}
`, name)
}
