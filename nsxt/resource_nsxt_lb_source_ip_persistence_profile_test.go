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

func TestAccResourceNsxtLbSourceIpPersistenceProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_source_ip_persistence_profile.test"
	timeout := "100"
	updatedTimeout := "200"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbSourceIPPersistenceProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbSourceIPPersistenceProfileBasicTemplate(name, timeout),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbSourceIPPersistenceProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "true"),
					resource.TestCheckResourceAttr(testResourceName, "ha_persistence_mirroring", "true"),
					resource.TestCheckResourceAttr(testResourceName, "purge_when_full", "false"),
					resource.TestCheckResourceAttr(testResourceName, "timeout", timeout),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbSourceIPPersistenceProfileBasicTemplate(updatedName, updatedTimeout),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbSourceIPPersistenceProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "true"),
					resource.TestCheckResourceAttr(testResourceName, "ha_persistence_mirroring", "true"),
					resource.TestCheckResourceAttr(testResourceName, "purge_when_full", "false"),
					resource.TestCheckResourceAttr(testResourceName, "timeout", updatedTimeout),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbSourceIpPersistenceProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_source_ip_persistence_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbSourceIPPersistenceProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbSourceIPPersistenceProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbSourceIPPersistenceProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB source ip persistence profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB source ip persistence profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerSourceIpPersistenceProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB source ip persistence profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB source ip persistence profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB source ip persistence profile %s wasn't found", displayName)
	}
}

func testAccNSXLbSourceIPPersistenceProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_source_ip_persistence_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerSourceIpPersistenceProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB source ip persistence profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("NSX LB source ip persistence profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbSourceIPPersistenceProfileBasicTemplate(name string, timeout string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_source_ip_persistence_profile" "test" {
  display_name             = "%s"
  description             = "test description"
  persistence_shared       = "true"
  ha_persistence_mirroring = "true"
  purge_when_full          = "false"
  timeout                  = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, timeout)
}

func testAccNSXLbSourceIPPersistenceProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_source_ip_persistence_profile" "test" {
  display_name = "%s"
  description  = "test description"
}
`, name)
}
