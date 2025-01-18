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

func TestAccResourceNsxtLbHTTPApplicationProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_http_application_profile.test"
	httpRedirect := "http://www.aaa.com"
	httpsRedirect := "false"
	idleTime := "10"
	bodySize := "11"
	headerSize := "12"
	respTimeout := "13"
	forward := "INSERT"
	ntlm := "true"

	upHTTPRedirect := ""
	upHTTPSRedirect := "true"
	upIdleTime := "110"
	upBodySize := "120"
	upHeaderSize := "130"
	upRespTimeout := "140"
	upForward := "REPLACE"
	upNtlm := "false"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPApplicationProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPApplicationCreateTemplate(name, httpRedirect, httpsRedirect, idleTime, bodySize, headerSize, respTimeout, forward, ntlm),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPApplicationProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "app profile"),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to", httpRedirect),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to_https", httpsRedirect),
					resource.TestCheckResourceAttr(testResourceName, "idle_timeout", idleTime),
					resource.TestCheckResourceAttr(testResourceName, "request_body_size", bodySize),
					resource.TestCheckResourceAttr(testResourceName, "request_header_size", headerSize),
					resource.TestCheckResourceAttr(testResourceName, "response_timeout", respTimeout),
					resource.TestCheckResourceAttr(testResourceName, "x_forwarded_for", forward),
					resource.TestCheckResourceAttr(testResourceName, "ntlm", ntlm),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbHTTPApplicationCreateTemplate(updatedName, upHTTPRedirect, upHTTPSRedirect, upIdleTime, upBodySize, upHeaderSize, upRespTimeout, upForward, upNtlm),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPApplicationProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "app profile"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHTTPApplicationProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_http_application_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPApplicationProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPApplicationCreateTemplateTrivial(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbHTTPApplicationProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX http application profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX http application profile resource ID not set in resources ")
		}

		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerHttpProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving http application profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if http application profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX http application profile %s wasn't found", displayName)
	}
}

func testAccNSXLbHTTPApplicationProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_http_application_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerHttpProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving http application profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX http application profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbHTTPApplicationCreateTemplate(name string, httpRedirect string, httpsRedirect string, idleTime string, bodySize string, headerSize string, respTimeout string, forward string, ntlm string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_http_application_profile" "test" {
  display_name           = "%s"
  description            = "app profile"
  http_redirect_to       = "%s"
  http_redirect_to_https = "%s"
  idle_timeout           = "%s"
  request_body_size      = "%s"
  request_header_size    = "%s"
  response_timeout       = "%s"
  x_forwarded_for        = "%s"
  ntlm                   = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, httpRedirect, httpsRedirect, idleTime, bodySize, headerSize, respTimeout, forward, ntlm)
}

func testAccNSXLbHTTPApplicationCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_http_application_profile" "test" {
  description = "test description"
}
`
}
