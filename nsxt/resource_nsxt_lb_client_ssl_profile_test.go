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

func TestAccResourceNsxtLbClientSSLProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_client_ssl_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbClientSSLProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbClientSSLCreateTemplate(name, "TLS_V1"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbClientSSLProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.#", "4"),
					resource.TestCheckResourceAttr(testResourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_timeout", "100"),
					resource.TestCheckResourceAttr(testResourceName, "prefer_server_ciphers", "true"),
					resource.TestCheckResourceAttr(testResourceName, "is_secure", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbClientSSLCreateTemplate(updatedName, "TLS_V1_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbClientSSLProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.#", "4"),
					resource.TestCheckResourceAttr(testResourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_timeout", "100"),
					resource.TestCheckResourceAttr(testResourceName, "prefer_server_ciphers", "true"),
					resource.TestCheckResourceAttr(testResourceName, "is_secure", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbClientSSLProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_client_ssl_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbClientSSLProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbClientSSLCreateTemplateTrivial(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbClientSSLProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX client ssl profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX client ssl profile resource ID not set in resources ")
		}

		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerClientSslProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving client ssl profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if client ssl profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX client ssl profile %s wasn't found", displayName)
	}
}

func testAccNSXLbClientSSLProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_client_ssl_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerClientSslProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving client ssl profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX client ssl profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbClientSSLCreateTemplate(name string, protocol string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_client_ssl_profile" "test" {
  display_name          = "%s"
  ciphers               = ["TLS_RSA_WITH_AES_128_CBC_SHA256", "TLS_RSA_WITH_AES_256_CBC_SHA", "TLS_RSA_WITH_3DES_EDE_CBC_SHA", "TLS_ECDH_RSA_WITH_AES_256_CBC_SHA"]
  protocols             = ["%s"]
  session_cache_enabled = true
  session_cache_timeout = 100
  prefer_server_ciphers = true
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, protocol)
}

func testAccNSXLbClientSSLCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_client_ssl_profile" "test" {
  description = "test description"
}
`
}
