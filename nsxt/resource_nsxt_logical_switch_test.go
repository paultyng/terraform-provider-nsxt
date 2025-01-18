// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtLogicalSwitch_basic(t *testing.T) {
	switchName := getAccTestResourceName()
	updateSwitchName := getAccTestResourceName()
	resourceName := "testoverlay"
	testResourceName := fmt.Sprintf("nsxt_logical_switch.%s", resourceName)
	novlan := "0"
	replicationMode := "MTEP"
	transportZoneName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalSwitchCheckDestroy(state, updateSwitchName, "nsxt_logical_switch")
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNSXLogicalSwitchNoTZIDTemplate(switchName),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config: testAccNSXLogicalSwitchCreateTemplate(resourceName, switchName, transportZoneName, novlan, replicationMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(switchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", switchName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", replicationMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "vlan", novlan),
					resource.TestCheckResourceAttr(testResourceName, "address_binding.#", "1"),
				),
			},
			{
				Config: testAccNSXLogicalSwitchUpdateTemplate(resourceName, updateSwitchName, transportZoneName, novlan, replicationMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(updateSwitchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateSwitchName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "DOWN"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", replicationMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "vlan", novlan),
					resource.TestCheckResourceAttr(testResourceName, "address_binding.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalSwitch_vlan(t *testing.T) {
	switchName := getAccTestResourceName()
	updateSwitchName := getAccTestResourceName()
	transportZoneName := getVlanTransportZoneName()
	resourceName := "testvlan"
	testResourceName := fmt.Sprintf("nsxt_logical_switch.%s", resourceName)

	origvlan := "1"
	updatedvlan := "2"
	replicationMode := ""

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalSwitchCheckDestroy(state, updateSwitchName, "nsxt_logical_switch")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalSwitchCreateTemplate(resourceName, switchName, transportZoneName, origvlan, replicationMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(switchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", switchName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttr(testResourceName, "vlan", origvlan),
				),
			},
			{
				Config: testAccNSXLogicalSwitchUpdateTemplate(resourceName, updateSwitchName, transportZoneName, updatedvlan, replicationMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(updateSwitchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateSwitchName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "DOWN"),
					resource.TestCheckResourceAttr(testResourceName, "vlan", updatedvlan),
				),
			},
		},
	})

}

func TestAccResourceNsxtLogicalSwitch_withProfiles(t *testing.T) {
	switchName := getAccTestResourceName()
	resourceName := "test_profiles"
	testResourceName := fmt.Sprintf("nsxt_logical_switch.%s", resourceName)
	transportZoneName := getOverlayTransportZoneName()
	customProfileName := "terraform_test_LS_profile"
	profileType := "SwitchSecuritySwitchingProfile"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			// Verify that the LS was deleted
			err := testAccNSXLogicalSwitchCheckDestroy(state, switchName, "nsxt_logical_switch")
			if err != nil {
				return err
			}
			// Delete the created switching profile
			return testAccDataSourceNsxtSwitchingProfileDeleteByName(customProfileName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					// Create a custom switching profile
					if err := testAccDataSourceNsxtSwitchingProfileCreate(customProfileName, profileType); err != nil {
						t.Error(err)
					}
				},
				// Create a logical switch to use the custom switching profile
				Config: testAccNSXLogicalSwitchCreateWithProfilesTemplate(resourceName, switchName, transportZoneName, customProfileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(switchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", switchName),
					resource.TestCheckResourceAttr(testResourceName, "switching_profile_id.#", "1"),
				),
			},
			{
				// Replace the custom switching profile with OOB one
				Config: testAccNSXLogicalSwitchUpdateWithProfilesTemplate(resourceName, switchName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(switchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", switchName),
					// Counting only custom profiles so count should be 0
					resource.TestCheckResourceAttr(testResourceName, "switching_profile_id.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalSwitch_withMacPool(t *testing.T) {
	switchName := getAccTestResourceName()
	resourceName := "test_mac_pool"
	testResourceName := fmt.Sprintf("nsxt_logical_switch.%s", resourceName)
	transportZoneName := getOverlayTransportZoneName()
	macPoolName := getMacPoolName()
	novlan := "0"
	replicationMode := "MTEP"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalSwitchCheckDestroy(state, switchName, "nsxt_logical_switch")
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNSXLogicalSwitchNoTZIDTemplate(switchName),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config: testAccNSXLogicalSwitchCreateWithMacTemplate(resourceName, switchName, transportZoneName, macPoolName, novlan, replicationMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalSwitchExists(switchName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", switchName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", replicationMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "vlan", novlan),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalSwitch_importBasic(t *testing.T) {
	switchName := getAccTestResourceName()
	resourceName := "testoverlay"
	testResourceName := fmt.Sprintf("nsxt_logical_switch.%s", resourceName)
	novlan := "0"
	replicationMode := "MTEP"
	transportZoneName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalSwitchCheckDestroy(state, switchName, "nsxt_logical_switch")
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNSXLogicalSwitchNoTZIDTemplate(switchName),
				ExpectError: regexp.MustCompile(`Missing required argument`),
			},
			{
				Config: testAccNSXLogicalSwitchCreateTemplate(resourceName, switchName, transportZoneName, novlan, replicationMode),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalSwitchExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical switch resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical switch resource ID not set in resources ")
		}

		logicalSwitch, responseCode, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical switch ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if logical switch %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == logicalSwitch.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical switch %s wasn't found", displayName)
	}
}

func testAccNSXLogicalSwitchCheckDestroy(state *terraform.State, resourceType string, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != resourceType {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		logicalSwitch, responseCode, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical switch ID %s. Error: %v", resourceID, err)
		}

		if displayName == logicalSwitch.DisplayName {
			return fmt.Errorf("NSX logical switch %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalSwitchNoTZIDTemplate(switchName string) string {
	return fmt.Sprintf(`
resource "nsxt_logical_switch" "error" {
  display_name     = "%s"
  admin_state      = "UP"
  description      = "Acceptance Test"
  replication_mode = "MTEP"
}`, switchName)
}

func testAccNSXLogicalSwitchCreateTemplate(resourceName string, switchName string, transportZoneName string, vlan string, replicationMode string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "%s" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"
  replication_mode  = "%s"
  vlan              = "%s"

  address_binding {
    ip_address = "1.1.1.1"
    mac_address = "00:11:22:33:44:55"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, transportZoneName, resourceName, switchName, replicationMode, vlan)
}

func testAccNSXLogicalSwitchUpdateTemplate(resourceName string, switchUpdateName string, transportZoneName string, vlan string, replicationMode string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "%s" {
  display_name      = "%s"
  admin_state       = "DOWN"
  description       = "Acceptance Test Update"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"
  replication_mode  = "%s"
  vlan              = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, transportZoneName, resourceName, switchUpdateName, replicationMode, vlan)
}

func testAccNSXLogicalSwitchCreateWithProfilesTemplate(resourceName string, switchName string, transportZoneName string, profileName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

data "nsxt_switching_profile" "test1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "%s" {
  display_name      = "%s"
  admin_state       = "UP"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"

  switching_profile_id {
    key   = "${data.nsxt_switching_profile.test1.resource_type}"
    value = "${data.nsxt_switching_profile.test1.id}"
  }
}`, transportZoneName, profileName, resourceName, switchName)
}

func testAccNSXLogicalSwitchUpdateWithProfilesTemplate(resourceName string, switchUpdateName string, transportZoneName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "%s" {
  display_name      = "%s"
  admin_state       = "UP"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"
}`, transportZoneName, resourceName, switchUpdateName)
}

func testAccNSXLogicalSwitchCreateWithMacTemplate(resourceName string, switchName string, transportZoneName string, macPoolName string, vlan string, replicationMode string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

data "nsxt_mac_pool" "MAC1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "%s" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"
  replication_mode  = "%s"
  vlan              = "%s"
  mac_pool_id       = "${data.nsxt_mac_pool.MAC1.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, transportZoneName, macPoolName, resourceName, switchName, replicationMode, vlan)
}
