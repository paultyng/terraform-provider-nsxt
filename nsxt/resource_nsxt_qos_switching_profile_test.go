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

func TestAccResourceNsxtQosSwitchingProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_qos_switching_profile.test"
	cos := "5"
	updatedCos := "2"
	peak := "700"
	updatedPeak := "400"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXQosSwitchingProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXQosSwitchingProfileBasicTemplate(name, cos, peak, "ingress"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXQosSwitchingProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "class_of_service", cos),
					resource.TestCheckResourceAttr(testResourceName, "dscp_trusted", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dscp_priority", "53"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.average_bw_mbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.peak_bw_mbps", peak),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.average_bw_kbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.peak_bw_kbps", peak),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXQosSwitchingProfileBasicTemplate(updatedName, updatedCos, updatedPeak, "egress"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXQosSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "class_of_service", updatedCos),
					resource.TestCheckResourceAttr(testResourceName, "dscp_trusted", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dscp_priority", "53"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.average_bw_mbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.peak_bw_mbps", updatedPeak),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.average_bw_kbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.peak_bw_kbps", updatedPeak),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXQosSwitchingProfileEmptyTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXQosSwitchingProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtQosSwitchingProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_qos_switching_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXQosSwitchingProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXQosSwitchingProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXQosSwitchingProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
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

		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetQosSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXQosSwitchingProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_qos_switching_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.LogicalSwitchingApi.GetQosSwitchingProfile(nsxClient.Context, resourceID)
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

func testAccNSXQosSwitchingProfileBasicTemplate(name string, cos string, peak string, direction string) string {
	return fmt.Sprintf(`
resource "nsxt_qos_switching_profile" "test" {
  display_name     = "%s"
  description      = "test description"
  class_of_service = %s
  dscp_trusted     = true
  dscp_priority    = 53

  %s_rate_shaper {
    average_bw_mbps = 111
    burst_size      = 222
    peak_bw_mbps    = "%s"
  }

  ingress_broadcast_rate_shaper {
    average_bw_kbps = 111
    burst_size      = 222
    peak_bw_kbps    = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, cos, direction, peak, peak)
}

func testAccNSXQosSwitchingProfileEmptyTemplate() string {
	return `
resource "nsxt_qos_switching_profile" "test" {
}`
}

func testAccNSXQosSwitchingProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_qos_switching_profile" "test" {
  display_name = "%s"
  description  = "test description"
  dscp_trusted = false

  egress_rate_shaper {
    enabled         = false
    peak_bw_mbps    = 800
    burst_size      = 222
    average_bw_mbps = 111
  }
}
`, name)
}
