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

func TestAccResourceNsxtIpBlockSubnet_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_ip_block_subnet.test"
	size := 16
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.4.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpBlockSubnetCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpBlockSubnetCreateTemplate(name, size),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpBlockSubnetExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "size", fmt.Sprint(size)),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "allocation_ranges.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ranges.0.start"),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ranges.0.end"),
					resource.TestCheckResourceAttrSet(testResourceName, "block_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "cidr"),
				),
			},
			{
				// Test updating the subnet (ForceNew)
				Config: testAccNSXIpBlockSubnetUpdateTemplate(updateName, size),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpBlockSubnetExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "size", fmt.Sprint(size)),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "allocation_ranges.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ranges.0.start"),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ranges.0.end"),
					resource.TestCheckResourceAttrSet(testResourceName, "block_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "cidr"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpBlockSubnet_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_ip_block_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.4.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpBlockSubnetCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpBlockSubnetCreateTemplate(name, 32),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXIpBlockSubnetExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("IP Block Subnet resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("IP Block Subnet resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpBlockSubnet(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving IP Block Subnet ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if IP Block Subnet %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		// Ignore display name to support the 'no-name' test
		if displayName == "" || displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("IP Block Subnet %s wasn't found", displayName)
	}
}

func testAccNSXIpBlockSubnetCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ip_block_subnet" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.PoolManagementApi.ReadIpBlockSubnet(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving IP Block Subnet ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("IP Block Subnet %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIpBlockTemplate() string {
	return `
resource "nsxt_ip_block" "ip_block" {
  display_name = "block1"
  cidr         = "55.0.0.0/24"
}`
}

func testAccNSXIpBlockSubnetCreateTemplate(name string, size int) string {
	return testAccNSXIpBlockTemplate() + fmt.Sprintf(`
resource "nsxt_ip_block_subnet" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  block_id     = "${nsxt_ip_block.ip_block.id}"
  size         = "%d"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, size)
}

func testAccNSXIpBlockSubnetUpdateTemplate(updatedName string, size int) string {
	return testAccNSXIpBlockTemplate() + fmt.Sprintf(`
resource "nsxt_ip_block_subnet" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"
  block_id     = "${nsxt_ip_block.ip_block.id}"
  size         = "%d"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName, size)
}
