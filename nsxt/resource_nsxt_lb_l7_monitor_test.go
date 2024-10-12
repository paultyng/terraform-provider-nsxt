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

var testLbMonitorCertID string
var testLbMonitorClientCertID string
var testLbMonitorCaCertID string

func TestAccResourceNsxtLbHTTPMonitor_basic(t *testing.T) {
	testAccResourceNsxtLbL7MonitorBasic(t, "http")
}

func TestAccResourceNsxtLbHTTPSMonitor_basic(t *testing.T) {
	testAccResourceNsxtLbL7MonitorBasic(t, "https")
}

func TestAccResourceNsxtLbHTTPMonitor_importBasic(t *testing.T) {
	testAccResourceNsxtLbL7MonitorImport(t, "http")
}

func TestAccResourceNsxtLbHTTPSMonitor_importBasic(t *testing.T) {
	testAccResourceNsxtLbL7MonitorImport(t, "https")
}

func testAccResourceNsxtLbL7MonitorBasic(t *testing.T, protocol string) {
	name := getAccTestResourceName()
	testResourceName := fmt.Sprintf("nsxt_lb_%s_monitor.test", protocol)
	requestMethod := "HEAD"
	updatedRequestMethod := "POST"
	body1 := "XXXXXXXXXXXXXXXXXXX"
	body2 := "YYYYYYYYYYYYYYYYYYY"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbL7MonitorCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbL7MonitorCreateTemplate(name, protocol, requestMethod, body1, body2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPSMonitorExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "request_url", "/healthcheck"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", requestMethod),
					resource.TestCheckResourceAttr(testResourceName, "request_body", body1),
					resource.TestCheckResourceAttr(testResourceName, "response_body", body2),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbL7MonitorCreateTemplate(name, protocol, updatedRequestMethod, body2, body1),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPSMonitorExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "request_url", "/healthcheck"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", updatedRequestMethod),
					resource.TestCheckResourceAttr(testResourceName, "request_body", body2),
					resource.TestCheckResourceAttr(testResourceName, "response_body", body1),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHTTPSMonitor_withAuth(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_https_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			testAccNSXDeleteCerts(t, testLbMonitorCertID, testLbMonitorClientCertID, testLbMonitorCaCertID)
			return testAccNSXLbL7MonitorCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testLbMonitorCertID, testLbMonitorClientCertID, testLbMonitorCaCertID = testAccNSXCreateCerts(t)
				},

				Config: testAccNSXLbHTTPSMonitorCreateTemplateWithAuth(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPSMonitorExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "certificate_chain_depth", "2"),
					resource.TestCheckResourceAttr(testResourceName, "server_auth", "REQUIRED"),
					resource.TestCheckResourceAttr(testResourceName, "server_auth_ca_ids.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "client_certificate_id"),
				),
			},
		},
	})
}

func testAccResourceNsxtLbL7MonitorImport(t *testing.T, protocol string) {
	name := getAccTestResourceName()
	testResourceName := fmt.Sprintf("nsxt_lb_%s_monitor.test", protocol)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbL7MonitorCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbL7MonitorCreateTemplateTrivial(protocol),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbHTTPSMonitorExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB monitor resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB monitor resource ID not set in resources ")
		}

		monitor, response, err := nsxClient.ServicesApi.ReadLoadBalancerMonitor(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB monitor with ID %s. Error: %v", resourceID, err)
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB monitor %s exists. HTTP return code was %d", resourceID, response.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB monitor %s wasn't found", displayName)
	}
}

func testAccNSXLbL7MonitorCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_https_monitor" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		monitor, response, err := nsxClient.ServicesApi.ReadLoadBalancerMonitor(nsxClient.Context, resourceID)
		if err != nil {
			if response.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB https monitor with ID %s. Error: %v", resourceID, err)
		}

		if displayName == monitor.DisplayName {
			return fmt.Errorf("NSX LB https monitor %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbL7MonitorCreateTemplate(name string, protocol string, requestMethod string, requestBody string, responseBody string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_%s_monitor" "test" {
  description       = "test description"
  display_name      = "%s"
  request_header {
      name = "header1"
      value = "value1"
  }
  request_header {
      name = "header2"
      value = "value2"
  }
  request_method = "%s"
  request_url = "/healthcheck"
  request_body = "%s"
  response_body = "%s"
  response_status_codes = [200, 304, 404]
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, protocol, name, requestMethod, requestBody, responseBody)
}

func testAccNSXLbL7MonitorCreateTemplateTrivial(protocol string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_%s_monitor" "test" {
  description = "test description"
}
`, protocol)
}

func testAccNSXLbHTTPSMonitorCreateTemplateWithAuth(name string) string {
	return fmt.Sprintf(`
data "nsxt_certificate" "ca" {
  display_name = "test_ca"
}

data "nsxt_certificate" "client" {
  display_name = "test_client"
}

resource "nsxt_lb_https_monitor" "test" {
  description             = "test description"
  display_name            = "%s"
  server_auth             = "REQUIRED"
  server_auth_ca_ids      = ["${data.nsxt_certificate.ca.id}"]
  certificate_chain_depth = 2
  client_certificate_id   = "${data.nsxt_certificate.client.id}"
  protocols               = ["TLS_V1_2"]
  ciphers                 = ["TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384"]
}
`, name)
}
