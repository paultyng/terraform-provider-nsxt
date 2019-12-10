/* Copyright © 2018 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/vmware/go-vmware-nsxt"
	"net/http"
	"testing"
)

func TestAccResourceNsxtSpoofGuardSwitchingProfile_basic(t *testing.T) {
	name := "test-nsx-switching-profile"
	updatedName := fmt.Sprintf("%s-update", name)
	testResourceName := "nsxt_spoofguard_switching_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXSpoofguardSwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSpoofGuardSwitchingProfileCreateTemplate(name, "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSpoofGuardSwitchingProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "address_binding_whitelist_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXSpoofGuardSwitchingProfileCreateTemplate(updatedName, "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSpoofGuardSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "address_binding_whitelist_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXSpoofGuardSwitchingProfileEmptyTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXSpoofGuardSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtSpoofGuardSwitchingProfile_importBasic(t *testing.T) {
	name := "test-nsx-application-profile"
	testResourceName := "nsxt_spoofguard_switching_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXSpoofguardSwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXSpoofGuardSwitchingProfileCreateTemplate(name, "true"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXSpoofGuardSwitchingProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(*nsxt.APIClient)
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX switching profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX switching profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetSpoofGuardSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXSpoofguardSwitchingProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(*nsxt.APIClient)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_spoofguard_switching_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetSpoofGuardSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXSpoofGuardSwitchingProfileCreateTemplate(name string, enableWhitelist string) string {
	return fmt.Sprintf(`
resource "nsxt_spoofguard_switching_profile" "test" {
  display_name                      = "%s"
  description                       = "test description"
  address_binding_whitelist_enabled = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, enableWhitelist)
}

func testAccNSXSpoofGuardSwitchingProfileEmptyTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_spoofguard_switching_profile" "test" {
  display_name                      = "%s"
}
`, name)
}
