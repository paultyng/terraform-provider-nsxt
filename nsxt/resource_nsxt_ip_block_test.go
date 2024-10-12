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

func TestAccResourceNsxtIpBlock_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_ip_block.test"
	cidr1 := "1.1.1.0/24"
	cidr2 := "2.2.2.0/24"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpBlockCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpBlockCreateTemplate(name, cidr1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpBlockExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr1),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXIpBlockUpdateTemplate(updateName, cidr1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpBlockExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr1),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
			{
				// Test updating the Cidr (ForceNew)
				Config: testAccNSXIpBlockUpdateTemplate(updateName, cidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpBlockExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr2),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpBlock_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_ip_block.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpBlockCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpBlockCreateTemplate(name, "1.1.1.0/24"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXIpBlockExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("IP Block resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("IP Block resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpBlock(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving IP Block ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if IP Block %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		// Ignore display name to support the 'no-name' test
		if displayName == "" || displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("IP Block %s wasn't found", displayName)
	}
}

func testAccNSXIpBlockCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ip_block" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpBlock(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving IP Block ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("IP Block %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIpBlockCreateTemplate(name string, cidr string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_block" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  cidr         = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, cidr)
}

func testAccNSXIpBlockUpdateTemplate(updatedName string, cidr string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_block" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"
  cidr         = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName, cidr)
}
