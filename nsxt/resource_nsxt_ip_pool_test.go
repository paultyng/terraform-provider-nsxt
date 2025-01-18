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

func TestAccResourceNsxtIpPool_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_ip_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpPoolCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpPoolCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "1.1.1.0/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.allocation_ranges.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.gateway_ip", "1.1.1.12"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dns_suffix", "abc"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dns_nameservers.#", "1"),
				),
			},
			{
				Config: testAccNSXIpPoolUpdateTemplate(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpPoolExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpPool_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_ip_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpPoolCreateTemplate(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXIpPoolExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("IP Pool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("IP Pool resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpPool(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving IP Pool ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if IP Pool %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		// Ignore display name to support the 'no-name' test
		if displayName == "" || displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("IP Pool %s wasn't found", displayName)
	}
}

func testAccNSXIpPoolCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ip_pool" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpPool(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving IP Pool ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("IP Pool %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIpPoolCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_pool" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  subnet {
    allocation_ranges = ["1.1.1.1-1.1.1.11", "1.1.1.21-1.1.1.100"]
    cidr              = "1.1.1.0/24"
    gateway_ip        = "1.1.1.12"
    dns_suffix        = "abc"
    dns_nameservers   = ["3.3.3.3"]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXIpPoolUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_pool" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"

  subnet {
    allocation_ranges = ["1.1.1.1-1.1.1.11", "1.1.1.21-1.1.1.99"]
    cidr              = "1.1.1.0/24"
    gateway_ip        = "1.1.1.12"
    dns_suffix        = "abc"
    dns_nameservers   = ["3.3.3.3"]
  }

  subnet {
    allocation_ranges = ["2.1.1.1-2.1.1.11", "2.1.1.21-2.1.1.100"]
    cidr              = "2.1.1.0/24"
    gateway_ip        = "2.1.1.12"
    dns_suffix        = "abc"
    dns_nameservers   = ["33.33.33.33"]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName)
}
