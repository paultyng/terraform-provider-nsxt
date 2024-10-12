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

var testLbVirtualServerCertID string
var testLbVirtualServerClientCertID string
var testLbVirtualServerCaCertID string

var testLbVirtualServerHelper1Name = getAccTestResourceName()
var testLbVirtualServerHelper2Name = getAccTestResourceName()
var testLbVirtualServerHelper3Name = getAccTestResourceName()

func TestAccResourceNsxtLbHttpVirtualServer_basic(t *testing.T) {
	name := getAccTestResourceName()
	fullName := "nsxt_lb_http_virtual_server.test"
	port := "888"
	updatedPort := "999"
	enabled := "true"
	updatedEnabled := "false"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPVirtualServerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplate(name, port, enabled),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "true"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_port", "9090"),
					resource.TestCheckResourceAttr(fullName, "enabled", enabled),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", port),
					resource.TestCheckResourceAttr(fullName, "max_concurrent_connections", "20"),
					resource.TestCheckResourceAttr(fullName, "max_new_connection_rate", "10"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(fullName, "pool_id"),
					resource.TestCheckResourceAttrSet(fullName, "sorry_pool_id"),
				),
			},
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplate(name, updatedPort, updatedEnabled),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "true"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_port", "9090"),
					resource.TestCheckResourceAttr(fullName, "enabled", updatedEnabled),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", updatedPort),
					resource.TestCheckResourceAttr(fullName, "max_concurrent_connections", "20"),
					resource.TestCheckResourceAttr(fullName, "max_new_connection_rate", "10"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(fullName, "pool_id"),
					resource.TestCheckResourceAttrSet(fullName, "sorry_pool_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHttpVirtualServer_withRules(t *testing.T) {
	name := getAccTestResourceName()
	fullName := "nsxt_lb_http_virtual_server.test"
	rule1 := "rule1"
	rule2 := "rule3"
	updatedRule1 := "rule2"
	updatedRule2 := "rule1"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPVirtualServerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplateWithRules(name, rule1, rule2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", "443"),
					resource.TestCheckResourceAttr(fullName, "rule_ids.#", "2"),
				),
			},
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplateWithRules(name, updatedRule1, updatedRule2),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", "443"),
					resource.TestCheckResourceAttr(fullName, "rule_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHttpVirtualServer_withSSL(t *testing.T) {
	name := getAccTestResourceName()
	fullName := "nsxt_lb_http_virtual_server.test"
	depth := "2"
	updatedDepth := "4"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			testAccNSXDeleteCerts(t, testLbVirtualServerCertID, testLbVirtualServerClientCertID, testLbVirtualServerCaCertID)
			return testAccNSXLbHTTPVirtualServerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					testLbVirtualServerCertID, testLbVirtualServerClientCertID, testLbVirtualServerCaCertID = testAccNSXCreateCerts(t)
				},
				Config: testAccNSXLbHTTPVirtualServerCreateTemplateWithSSL(name, depth),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "false"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "enabled", "true"),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", "443"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.#", "1"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.certificate_chain_depth", depth),
					resource.TestCheckResourceAttrSet(fullName, "client_ssl.0.client_ssl_profile_id"),
					resource.TestCheckResourceAttrSet(fullName, "client_ssl.0.default_certificate_id"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.client_auth", "true"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.ca_ids.#", "1"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.certificate_chain_depth", depth),
					resource.TestCheckResourceAttrSet(fullName, "server_ssl.0.server_ssl_profile_id"),
					resource.TestCheckResourceAttrSet(fullName, "server_ssl.0.client_certificate_id"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.server_auth", "true"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.ca_ids.#", "1"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplateWithSSL(name, updatedDepth),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPVirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "false"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "enabled", "true"),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "port", "443"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.#", "1"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.certificate_chain_depth", updatedDepth),
					resource.TestCheckResourceAttrSet(fullName, "client_ssl.0.client_ssl_profile_id"),
					resource.TestCheckResourceAttrSet(fullName, "client_ssl.0.default_certificate_id"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.client_auth", "true"),
					resource.TestCheckResourceAttr(fullName, "client_ssl.0.ca_ids.#", "1"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.certificate_chain_depth", updatedDepth),
					resource.TestCheckResourceAttrSet(fullName, "server_ssl.0.server_ssl_profile_id"),
					resource.TestCheckResourceAttrSet(fullName, "server_ssl.0.client_certificate_id"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.server_auth", "true"),
					resource.TestCheckResourceAttr(fullName, "server_ssl.0.ca_ids.#", "1"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHttpVirtualServer_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_lb_http_virtual_server.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPVirtualServerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPVirtualServerCreateTemplateTrivial(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbHTTPVirtualServerExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB virtual server resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB virtual server resource ID not set in resources ")
		}

		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerVirtualServer(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB virtual server with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB virtual server %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB virtual server %s wasn't found", displayName)
	}
}

func testAccNSXLbHTTPVirtualServerCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_http_virtual_server" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerVirtualServer(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB virtual server with ID %s. Error: %v", resourceID, err)
		}

		if displayName == monitor.DisplayName {
			return fmt.Errorf("NSX LB virtual server %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbHTTPVirtualServerCreateTemplate(name string, port string, enabled string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_http_application_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_cookie_persistence_profile" "test" {
  display_name = "%s"
  cookie_name  = "test"
}

resource "nsxt_lb_pool" "test" {
  display_name = "%s"
  algorithm    = "ROUND_ROBIN"
}

resource "nsxt_lb_pool" "sorry" {
  display_name = "%s"
  algorithm    = "ROUND_ROBIN"
}

resource "nsxt_lb_http_virtual_server" "test" {
  display_name               = "%s"
  description                = "test description"
  access_log_enabled         = true
  application_profile_id     = "${nsxt_lb_http_application_profile.test.id}"
  default_pool_member_port   = "9090"
  enabled                    = %s
  ip_address                 = "1.1.1.1"
  port                       = "%s"
  max_concurrent_connections = 20
  max_new_connection_rate    = 10
  persistence_profile_id     = "${nsxt_lb_cookie_persistence_profile.test.id}"
  pool_id                    = "${nsxt_lb_pool.test.id}"
  sorry_pool_id              = "${nsxt_lb_pool.sorry.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, name, testLbVirtualServerHelper1Name, testLbVirtualServerHelper2Name, name, enabled, port)
}

// TODO: add other types of rules
func testAccNSXLbHTTPVirtualServerCreateTemplateWithRules(name string, rule1 string, rule2 string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_http_application_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_http_request_rewrite_rule" "rule1" {
  display_name = "%s"
  method_condition {
    method = "HEAD"
  }

  header_rewrite_action {
    name  = "NAME1"
    value = "VALUE1"
  }
}

resource "nsxt_lb_http_request_rewrite_rule" "rule2" {
  display_name = "%s"
  uri_condition {
    uri        = "/hello"
    match_type = "STARTS_WITH"
  }

  header_rewrite_action {
    name  = "NAME1"
    value = "VALUE1"
  }
}

resource "nsxt_lb_http_request_rewrite_rule" "rule3" {
  display_name = "%s"
  uri_condition {
    uri = "html"
    match_type = "ENDS_WITH"
  }

  header_rewrite_action {
    name  = "NAME1"
    value = "VALUE1"
  }
}

resource "nsxt_lb_http_virtual_server" "test" {
  display_name           = "%s"
  application_profile_id = "${nsxt_lb_http_application_profile.test.id}"
  ip_address             = "1.1.1.1"
  port                   = "443"
  rule_ids               = ["${nsxt_lb_http_request_rewrite_rule.%s.id}", "${nsxt_lb_http_request_rewrite_rule.%s.id}"]
}
`, testLbVirtualServerHelper1Name, testLbVirtualServerHelper1Name, testLbVirtualServerHelper2Name, testLbVirtualServerHelper3Name, name, rule1, rule2)
}

func testAccNSXLbHTTPVirtualServerCreateTemplateWithSSL(name string, depth string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_http_application_profile" "test" {
  display_name = "%s"
}

data "nsxt_certificate" "ca" {
  display_name = "test_ca"
}

data "nsxt_certificate" "client" {
  display_name = "test_client"
}

data "nsxt_certificate" "mine" {
  display_name = "test"
}

resource "nsxt_lb_client_ssl_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_server_ssl_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_http_virtual_server" "test" {
  display_name           = "%s"
  description            = "test description"
  application_profile_id = "${nsxt_lb_http_application_profile.test.id}"
  ip_address             = "1.1.1.1"
  port                   = "443"

  client_ssl {
    client_ssl_profile_id   = "${nsxt_lb_client_ssl_profile.test.id}"
    default_certificate_id  = "${data.nsxt_certificate.mine.id}"
    certificate_chain_depth = %s
    client_auth             = true
    ca_ids                  = ["${data.nsxt_certificate.ca.id}"]
  }

  server_ssl {
    server_ssl_profile_id   = "${nsxt_lb_server_ssl_profile.test.id}"
    client_certificate_id   = "${data.nsxt_certificate.client.id}"
    certificate_chain_depth = %s
    server_auth             = true
    ca_ids                  = ["${data.nsxt_certificate.ca.id}"]
  }

}
`, name, name, name, name, depth, depth)
}

func testAccNSXLbHTTPVirtualServerCreateTemplateTrivial() string {
	return fmt.Sprintf(`

resource "nsxt_lb_http_application_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_http_virtual_server" "test" {
  description            = "test description"
  application_profile_id = "${nsxt_lb_http_application_profile.test.id}"
  ip_address             = "2.2.2.2"
  port                   = "443"
}
`, testLbVirtualServerHelper1Name)
}
