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

func TestAccResourceNsxtIpSet_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_ip_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersionLessThan(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpSetCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpSetCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpSetExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
				),
			},
			{
				Config: testAccNSXIpSetUpdateTemplate(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpSetExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpSet_basic_900(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpSetCheckDestroy(state, resourceName)
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNSXIpSetCreateTemplate(name),
				ExpectError: regexp.MustCompile("MP resource.*has been removed in NSX"),
			},
		},
	})
}

func TestAccResourceNsxtIpSet_noName(t *testing.T) {
	name := ""
	testResourceName := "nsxt_ip_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersionLessThan(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpSetCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpSetCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpSetExists(name, testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
				),
			},
			{
				Config: testAccNSXIpSetUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIpSetExists(name, testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIpSet_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_ip_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersionLessThan(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpSetCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIpSetCreateTemplate(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtIpSet_importBasic_900(t *testing.T) {
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIpSetCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNSXIpSetCreateTemplate(name),
				ExpectError: regexp.MustCompile("MP resource.*has been removed in NSX"),
			},
		},
	})
}

func testAccNSXIpSetExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("IP Set resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("IP Set resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.GroupingObjectsApi.ReadIPSet(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving IP Set ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if IP Set %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		// Ignore display name to support the 'no-name' test
		if displayName == "" || displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("IP Set %s wasn't found", displayName)
	}
}

func testAccNSXIpSetCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ip_set" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.GroupingObjectsApi.ReadIPSet(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving IP Set ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("IP Set %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIpSetCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_set" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  ip_addresses = ["1.1.1.1"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXIpSetUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ip_set" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"
  ip_addresses = ["1.1.1.1", "2.1.1.0/24", "3.1.1.1-3.1.1.10"]

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
