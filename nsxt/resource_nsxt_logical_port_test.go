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

func TestAccResourceNsxtLogicalPort_basic(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_port.test"
	transportZoneName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalPortCheckDestroy(state, updatePortName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalPortCreateTemplate(portName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLogicalPortUpdateTemplate(updatePortName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "admin_state", "UP"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalPort_withProfiles(t *testing.T) {
	portName := getAccTestResourceName()
	updatePortName := getAccTestResourceName()
	testResourceName := "nsxt_logical_port.test"
	transportZoneName := getOverlayTransportZoneName()
	customProfileName := "terraform_test_LP_profile"
	profileType := "QosSwitchingProfile"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			// Verify that the LP was deleted
			err := testAccNSXLogicalPortCheckDestroy(state, updatePortName)
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
				// Create a logical port to use the custom switching profile
				Config: testAccNSXLogicalPortCreateWithProfilesTemplate(portName, transportZoneName, customProfileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
					resource.TestCheckResourceAttr(testResourceName, "switching_profile_id.#", "1"),
				),
			},
			{
				// Remove custom switching profile
				Config: testAccNSXLogicalPortUpdateWithProfilesTemplate(updatePortName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(updatePortName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatePortName),
					resource.TestCheckResourceAttr(testResourceName, "switching_profile_id.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalPort_withNSGroup(t *testing.T) {
	// Verify port can be deleted if its an effective member
	// of existing NS Group
	portName := getAccTestResourceName()
	testResourceName := "nsxt_logical_port.test"
	transportZoneName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalPortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalPortCreateWithGroup(portName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalPortExists(portName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", portName),
				),
			},
			{
				Config: testAccNSXLogicalPortCreateNSGroup(),
				Check: func(state *terraform.State) error {
					return testAccNSXLogicalPortCheckDestroy(state, portName)
				},
			},
		},
	})
}

func TestAccResourceNsxtLogicalPort_importBasic(t *testing.T) {
	portName := getAccTestResourceName()
	testResourceName := "nsxt_logical_port.test"
	transportZoneName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalPortCheckDestroy(state, portName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalPortCreateTemplate(portName, transportZoneName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalPortExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical port resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical port resource ID not set in resources ")
		}

		logicalPort, responseCode, err := nsxClient.LogicalSwitchingApi.GetLogicalPort(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical port ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if logical port %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == logicalPort.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical port %s wasn't found", displayName)
	}
}

func testAccNSXLogicalPortCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_port" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		logicalPort, responseCode, err := nsxClient.LogicalSwitchingApi.GetLogicalPort(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical port ID %s. Error: %v", resourceID, err)
		}

		if displayName == logicalPort.DisplayName {
			return fmt.Errorf("NSX logical port %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalSwitchCreateForPort(transportZoneName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "TZ1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "test" {
  display_name      = "test-nsx-switch"
  admin_state       = "UP"
  replication_mode  = "MTEP"
  transport_zone_id = "${data.nsxt_transport_zone.TZ1.id}"
}`, transportZoneName)
}

func testAccNSXLogicalPortCreateTemplate(portName string, transportZoneName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + fmt.Sprintf(`
resource "nsxt_logical_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.test.id}"

  tag {
  	scope = "scope1"
    tag   = "tag1"
  }
}`, portName)
}

func testAccNSXLogicalPortUpdateTemplate(portUpdatedName string, transportZoneName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + fmt.Sprintf(`
resource "nsxt_logical_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test Update"
  logical_switch_id = "${nsxt_logical_switch.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, portUpdatedName)
}

func testAccNSXLogicalPortCreateWithProfilesTemplate(portName string, transportZoneName string, profileName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + fmt.Sprintf(`
data "nsxt_switching_profile" "test1" {
  display_name = "%s"
}

resource "nsxt_logical_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  logical_switch_id = "${nsxt_logical_switch.test.id}"

  switching_profile_id {
    key   = "${data.nsxt_switching_profile.test1.resource_type}"
    value = "${data.nsxt_switching_profile.test1.id}"
  }
}`, profileName, portName)
}

func testAccNSXLogicalPortUpdateWithProfilesTemplate(portUpdatedName string, transportZoneName string) string {
	return testAccNSXLogicalSwitchCreateForPort(transportZoneName) + fmt.Sprintf(`
resource "nsxt_logical_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  logical_switch_id = "${nsxt_logical_switch.test.id}"

}`, portUpdatedName)
}

func testAccNSXLogicalPortCreateNSGroup() string {
	return `
resource "nsxt_ns_group" "test" {
  membership_criteria {
      target_type = "LogicalPort"
      scope       = "scope1"
      tag         = "tag1"
  }
}`
}

func testAccNSXLogicalPortCreateWithGroup(portName string, transportZoneName string) string {
	return testAccNSXLogicalPortCreateNSGroup() + testAccNSXLogicalSwitchCreateForPort(transportZoneName) + fmt.Sprintf(`
resource "nsxt_logical_port" "test" {
  display_name      = "%s"
  admin_state       = "UP"
  description       = "Acceptance Test"
  logical_switch_id = "${nsxt_logical_switch.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, portName)
}
