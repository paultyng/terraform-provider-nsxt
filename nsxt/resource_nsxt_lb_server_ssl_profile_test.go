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

func TestAccResourceNsxtLbServerSSLProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_server_ssl_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbServerSSLProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbServerSSLCreateTemplate(name, "TLS_V1_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServerSSLProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "is_secure", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbServerSSLCreateTemplate(updatedName, "TLS_V1_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServerSSLProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "is_secure", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbServerSSLProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_server_ssl_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbServerSSLProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbServerSSLCreateTemplateTrivial(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbServerSSLProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX server ssl profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX server ssl profile resource ID not set in resources ")
		}

		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerServerSslProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving server ssl profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if server ssl profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX server ssl profile %s wasn't found", displayName)
	}
}

func testAccNSXLbServerSSLProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_server_ssl_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerServerSslProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving server ssl profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX server ssl profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbServerSSLCreateTemplate(name string, protocol string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_server_ssl_profile" "test" {
  display_name          = "%s"
  ciphers               = ["TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384", "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384"]
  protocols             = ["%s"]
  session_cache_enabled = false
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, protocol)
}

func testAccNSXLbServerSSLCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_server_ssl_profile" "test" {
  description = "test description"
}
`
}
