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

func TestAccResourceNsxtMacManagementSwitchingProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_mac_management_switching_profile.test"
	limit := "100"
	updatedLimit := "101"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXMacManagementSwitchingProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXMacManagementSwitchingProfileCreateTemplate(name, "true", "true", limit, "ALLOW", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXMacManagementSwitchingProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "mac_change_allowed", "true"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.limit", limit),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.limit_policy", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.unicast_flooding_allowed", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXMacManagementSwitchingProfileCreateTemplate(updatedName, "false", "false", updatedLimit, "DROP", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXMacManagementSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "mac_change_allowed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.limit", updatedLimit),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.limit_policy", "DROP"),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.0.unicast_flooding_allowed", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXMacManagementSwitchingProfileBasicTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXMacManagementSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "mac_learning.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtMacManagementSwitchingProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_mac_management_switching_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXMacManagementSwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXMacManagementSwitchingProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXMacManagementSwitchingProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
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

		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetMacManagementSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXMacManagementSwitchingProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_mac_management_switching_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetMacManagementSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXMacManagementSwitchingProfileCreateTemplate(name string, macChange string, macLearnEnabled string, limit string, limitPolicy string, unicast string) string {
	return fmt.Sprintf(`
resource "nsxt_mac_management_switching_profile" "test" {
  display_name       = "%s"
  description        = "test description"
  mac_change_allowed = "%s"

  mac_learning {
  	enabled                  = "%s"
  	limit                    = "%s"
  	limit_policy             = "%s"
  	unicast_flooding_allowed = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, macChange, macLearnEnabled, limit, limitPolicy, unicast)
}

func testAccNSXMacManagementSwitchingProfileBasicTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_mac_management_switching_profile" "test" {
  display_name       = "%s"
}
`, name)
}

func testAccNSXMacManagementSwitchingProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_mac_management_switching_profile" "test" {
  display_name       = "%s"
  mac_change_allowed = "true"
}
`, name)
}
