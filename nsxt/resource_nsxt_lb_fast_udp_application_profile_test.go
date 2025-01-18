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

func TestAccResourceNsxtLbFastUDPApplicationProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_fast_udp_application_profile.test"
	idleTimeout := "100"
	updatedIdleTimeout := "200"
	mirroring := "true"
	updatedMirroring := "false"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbFastUDPApplicationProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbFastUDPApplicationProfileBasicTemplate(name, idleTimeout, mirroring),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbFastUDPApplicationProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "idle_timeout", idleTimeout),
					resource.TestCheckResourceAttr(testResourceName, "ha_flow_mirroring", mirroring),
				),
			},
			{
				Config: testAccNSXLbFastUDPApplicationProfileBasicTemplate(updatedName, updatedIdleTimeout, updatedMirroring),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbFastUDPApplicationProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "idle_timeout", updatedIdleTimeout),
					resource.TestCheckResourceAttr(testResourceName, "ha_flow_mirroring", updatedMirroring),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbFastUDPApplicationProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_fast_udp_application_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbFastUDPApplicationProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbFastUDPApplicationProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbFastUDPApplicationProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB fast udp application profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB fast udp application profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerFastUdpProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB fast udp application profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB fast udp application profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB fast udp application profile %s wasn't found", displayName)
	}
}

func testAccNSXLbFastUDPApplicationProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_fast_udp_application_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerFastUdpProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB fast udp application profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("NSX LB fast udp application profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbFastUDPApplicationProfileBasicTemplate(name string, idleTimeout string, mirroring string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_fast_udp_application_profile" "test" {
  display_name      = "%s"
  description       = "test description"
  idle_timeout      = "%s"
  ha_flow_mirroring = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, idleTimeout, mirroring)
}

func testAccNSXLbFastUDPApplicationProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_fast_udp_application_profile" "test" {
  display_name = "%s"
  description  = "test description"
}
`, name)
}
