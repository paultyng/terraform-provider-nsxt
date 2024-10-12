// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
)

var accTestPolicyLBVirtualServerCreateAttributes = map[string]string{
	"display_name":                getAccTestResourceName(),
	"description":                 "terraform created",
	"access_log_enabled":          "true",
	"default_pool_member_ports":   "777-778",
	"enabled":                     "true",
	"ip_address":                  "1.1.1.1",
	"max_concurrent_connections":  "2",
	"max_new_connection_rate":     "2",
	"ports":                       "80-81",
	"certificate_chain_depth":     "3",
	"log_significant_event_only":  "true",
	"access_list_control_enabled": "true",
	"access_list_control_action":  "ALLOW",
}

var accTestPolicyLBVirtualServerUpdateAttributes = map[string]string{
	"display_name":                getAccTestResourceName(),
	"description":                 "terraform updated",
	"access_log_enabled":          "false",
	"default_pool_member_ports":   "555",
	"enabled":                     "false",
	"ip_address":                  "1.1.2.1",
	"max_concurrent_connections":  "5",
	"max_new_connection_rate":     "5",
	"ports":                       "443",
	"certificate_chain_depth":     "2",
	"log_significant_event_only":  "false",
	"access_list_control_enabled": "false",
	"access_list_control_action":  "DROP",
}

// TODO: add service_path to the test when backend bug is fixed
func TestAccResourceNsxtPolicyLBVirtualServer_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_virtual_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBVirtualServerCheckDestroy(state, accTestPolicyLBVirtualServerCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBVirtualServerTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBVirtualServerCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "access_log_enabled", accTestPolicyLBVirtualServerCreateAttributes["access_log_enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "application_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "default_pool_member_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "default_pool_member_ports.0", accTestPolicyLBVirtualServerCreateAttributes["default_pool_member_ports"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyLBVirtualServerCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "max_concurrent_connections", accTestPolicyLBVirtualServerCreateAttributes["max_concurrent_connections"]),
					resource.TestCheckResourceAttr(testResourceName, "max_new_connection_rate", accTestPolicyLBVirtualServerCreateAttributes["max_new_connection_rate"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", accTestPolicyLBVirtualServerCreateAttributes["ports"]),
					resource.TestCheckResourceAttrSet(testResourceName, "sorry_pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "persistence_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBVirtualServerUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "access_log_enabled", accTestPolicyLBVirtualServerUpdateAttributes["access_log_enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "application_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "default_pool_member_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyLBVirtualServerUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "max_concurrent_connections", accTestPolicyLBVirtualServerUpdateAttributes["max_concurrent_connections"]),
					resource.TestCheckResourceAttr(testResourceName, "max_new_connection_rate", accTestPolicyLBVirtualServerUpdateAttributes["max_new_connection_rate"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", accTestPolicyLBVirtualServerUpdateAttributes["ports"]),
					resource.TestCheckResourceAttrSet(testResourceName, "sorry_pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "persistence_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "pool_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "sorry_pool_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "persistence_profile_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

// TODO: add CRLs and other certificates
func TestAccResourceNsxtPolicyLBVirtualServer_withSSL(t *testing.T) {
	testResourceName := "nsxt_policy_lb_virtual_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_CERTIFICATE_NAME")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBVirtualServerCheckDestroy(state, accTestPolicyLBVirtualServerCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBVirtualServerSSLTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.0.client_auth", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.0.certificate_chain_depth", accTestPolicyLBVirtualServerCreateAttributes["certificate_chain_depth"]),
					resource.TestCheckResourceAttrSet(testResourceName, "client_ssl.0.default_certificate_path"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.server_auth", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.certificate_chain_depth", accTestPolicyLBVirtualServerCreateAttributes["certificate_chain_depth"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerSSLTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.0.client_auth", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.0.certificate_chain_depth", accTestPolicyLBVirtualServerUpdateAttributes["certificate_chain_depth"]),
					resource.TestCheckResourceAttrSet(testResourceName, "client_ssl.0.default_certificate_path"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.server_auth", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.certificate_chain_depth", accTestPolicyLBVirtualServerUpdateAttributes["certificate_chain_depth"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "client_ssl.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBVirtualServer_withAccessList(t *testing.T) {
	testResourceName := "nsxt_policy_lb_virtual_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBVirtualServerCheckDestroy(state, accTestPolicyLBVirtualServerCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBVirtualServerWithAccessList(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", accTestPolicyLBVirtualServerCreateAttributes["ports"]),
					resource.TestCheckResourceAttr(testResourceName, "log_significant_event_only", accTestPolicyLBVirtualServerCreateAttributes["log_significant_event_only"]),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.0.enabled", accTestPolicyLBVirtualServerCreateAttributes["access_list_control_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.0.action", accTestPolicyLBVirtualServerCreateAttributes["access_list_control_action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "access_list_control.0.group_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerWithAccessList(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", accTestPolicyLBVirtualServerUpdateAttributes["ports"]),
					resource.TestCheckResourceAttr(testResourceName, "log_significant_event_only", accTestPolicyLBVirtualServerUpdateAttributes["log_significant_event_only"]),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.0.enabled", accTestPolicyLBVirtualServerUpdateAttributes["access_list_control_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.0.action", accTestPolicyLBVirtualServerUpdateAttributes["access_list_control_action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "access_list_control.0.group_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "access_list_control.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "pool_path", ""),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBVirtualServer_withRules(t *testing.T) {
	testResourceName := "nsxt_policy_lb_virtual_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBVirtualServerCheckDestroy(state, accTestPolicyLBVirtualServerCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBVirtualServerWithRules(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", "80"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),

					resource.TestCheckResourceAttr(testResourceName, "rule.#", "14"),

					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "connection_drop_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action.0.connection_drop.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.condition.#", "0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "http_redirect_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.0.redirect_status", "301"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.0.redirect_url", "dummy"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.1.redirect_status", "302"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.1.redirect_url", "other_dummy"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.0.body_value", "xyz"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.2.display_name", "http_reject_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.0.reply_status", "404"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.1.reply_status", "400"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.cookie_name", "is"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.cookie_value", "brownie"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.3.display_name", "http_request_header_delete_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.header_name", "my_head"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.header_value", "has_hair"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.4.display_name", "http_request_header_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.0.header_value", "and"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.1.header_value", "!"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.0.uri_arguments", "foo=bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.0.uri", "xyz"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.5.display_name", "http_request_uri_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.0.uri", "uri_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.1.uri", "uri_test2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.1.uri_arguments", "foo=bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.0.http_request_version.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.0.http_request_version.0.version", "HTTP_VERSION_1_0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.6.display_name", "http_response_header_delete_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.header_value", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.7.display_name", "http_response_header_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.0.header_value", "and"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.1.header_value", "!"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.0.ip_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.0.ip_header.0.source_address", "1.1.1.1"),

					resource.TestCheckResourceAttr(testResourceName, "rule.8.display_name", "select_pool_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.phase", "HTTP_FORWARDING"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.action.0.select_pool.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.8.action.0.select_pool.0.pool_id"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.0.sni", "HELO"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.display_name", "variable_assignment_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.phase", "HTTP_ACCESS"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.0.variable_name", "foo"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.0.variable_value", "bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.variable_name", "my_var"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.variable_value", "my_value"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.10.display_name", "variable_persistence_learn_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.10.action.0.variable_persistence_learn.0.persistence_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.0.variable_hash_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.0.variable_name", "my_name"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.session_reused", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_protocol", "TLS_V1_2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_ssl_cipher", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.session_reused", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_protocol", "TLS_V1_2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.0.issuer_dn", "something"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.0.subject_dn", "something"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.0", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.1", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"),
					resource.TestCheckResourceAttr(testResourceName, "rule.11.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.11.condition.0.http_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.11.condition.0.http_ssl.0.session_reused", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.11.condition.0.http_ssl.0.used_protocol", "TLS_V1_2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.11.condition.0.http_ssl.0.used_ssl_cipher", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"),

					resource.TestCheckResourceAttr(testResourceName, "rule.12.display_name", "variable_persistence_on_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.phase", "HTTP_FORWARDING"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.12.action.0.variable_persistence_on.0.persistence_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.0.variable_hash_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.0.variable_name", "my_name"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.condition.#", "0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.13.display_name", "jwt_auth_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.phase", "HTTP_ACCESS"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.key.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.key.0.public_key_content", "xxx"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.pass_jwt_to_pool", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.realm", "realm"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.tokens.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.tokens.0", "a"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerWithRules(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBVirtualServerUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestPolicyLBVirtualServerUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ports.0", "80"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),

					resource.TestCheckResourceAttr(testResourceName, "rule.#", "14"),

					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "connection_drop_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action.0.connection_drop.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.condition.#", "0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "http_redirect_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.0.redirect_status", "301"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.0.redirect_url", "dummy"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.1.redirect_status", "302"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action.0.http_redirect.1.redirect_url", "other_dummy"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.0.body_value", "xyz"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.condition.0.http_request_body.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.2.display_name", "http_reject_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.0.reply_status", "404"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action.0.http_reject.1.reply_status", "400"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.cookie_name", "is"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.cookie_value", "brownie"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.condition.0.http_request_cookie.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.3.display_name", "http_request_header_delete_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action.0.http_request_header_delete.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.header_name", "my_head"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.header_value", "has_hair"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.condition.0.http_request_header.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.4.display_name", "http_request_header_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.0.header_value", "and"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action.0.http_request_header_rewrite.1.header_value", "!"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.0.uri_arguments", "foo=bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri_arguments.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.0.uri", "xyz"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.condition.0.http_request_uri.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.5.display_name", "http_request_uri_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.phase", "HTTP_REQUEST_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.0.uri", "uri_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.1.uri", "uri_test2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action.0.http_request_uri_rewrite.1.uri_arguments", "foo=bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.0.http_request_version.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.condition.0.http_request_version.0.version", "HTTP_VERSION_1_0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.6.display_name", "http_response_header_delete_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.action.0.http_response_header_delete.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.header_value", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.6.condition.0.http_response_header.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.7.display_name", "http_response_header_rewrite_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.0.header_name", "clever"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.0.header_value", "and"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.1.header_name", "smart"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.action.0.http_response_header_rewrite.1.header_value", "!"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.0.ip_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.7.condition.0.ip_header.0.source_address", "1.1.1.1"),

					resource.TestCheckResourceAttr(testResourceName, "rule.8.display_name", "select_pool_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.phase", "HTTP_FORWARDING"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.action.0.select_pool.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.8.action.0.select_pool.0.pool_id"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.8.condition.0.ssl_sni.0.sni", "HELO"),

					resource.TestCheckResourceAttr(testResourceName, "rule.9.display_name", "variable_assignment_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.phase", "HTTP_ACCESS"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.0.variable_name", "foo"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.action.0.variable_assignment.0.variable_value", "bar"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.variable_name", "my_var"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.variable_value", "my_value"),
					resource.TestCheckResourceAttr(testResourceName, "rule.9.condition.0.variable.0.match_type", "REGEX"),

					resource.TestCheckResourceAttr(testResourceName, "rule.10.display_name", "variable_persistence_learn_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.phase", "HTTP_RESPONSE_REWRITE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.10.action.0.variable_persistence_learn.0.persistence_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.0.variable_hash_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.action.0.variable_persistence_learn.0.variable_name", "my_name"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.session_reused", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_protocol", "TLS_V1_2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_ssl_cipher", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.session_reused", "IGNORE"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.used_protocol", "TLS_V1_2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.0.issuer_dn", "something"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_issuer_dn.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.0.subject_dn", "something"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_certificate_subject_dn.0.match_type", "REGEX"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.0", "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"),
					resource.TestCheckResourceAttr(testResourceName, "rule.10.condition.0.http_ssl.0.client_supported_ssl_ciphers.1", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384"),

					resource.TestCheckResourceAttr(testResourceName, "rule.12.display_name", "variable_persistence_on_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.phase", "HTTP_FORWARDING"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.12.action.0.variable_persistence_on.0.persistence_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.0.variable_hash_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.action.0.variable_persistence_on.0.variable_name", "my_name"),
					resource.TestCheckResourceAttr(testResourceName, "rule.12.condition.#", "0"),

					resource.TestCheckResourceAttr(testResourceName, "rule.13.display_name", "jwt_auth_test"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.phase", "HTTP_ACCESS"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.key.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.key.0.public_key_content", "xxx"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.pass_jwt_to_pool", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.realm", "realm"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.tokens.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.13.action.0.jwt_auth.0.tokens.0", "a"),
				),
			},
			{
				Config: testAccNsxtPolicyLBVirtualServerMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBVirtualServerExists(testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBVirtualServer_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_virtual_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBVirtualServerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBVirtualServerMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBVirtualServerExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		nsxClient := infra.NewLbVirtualServersClient(connector)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBVirtualServer resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBVirtualServer resource ID not set in resources")
		}

		_, err := nsxClient.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy LBVirtualServer ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyLBVirtualServerCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	nsxClient := infra.NewLbVirtualServersClient(connector)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_virtual_server" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		_, err := nsxClient.Get(resourceID)
		if err == nil {
			return fmt.Errorf("Policy LBVirtualServer %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBVirtualServerTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBVirtualServerCreateAttributes
	} else {
		attrMap = accTestPolicyLBVirtualServerUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "default_tcp"{
  type = "TCP"
  display_name = "default-tcp-lb-app-profile"
}

data "nsxt_policy_lb_persistence_profile" "default" {
  type = "SOURCE_IP"
}

resource "nsxt_policy_lb_pool" "pool" {
  display_name = "terraform-vs-test-pool"
}

resource "nsxt_policy_lb_pool" "sorry_pool" {
  display_name = "terraform-vs-test-sorry-pool"
}

resource "nsxt_policy_lb_virtual_server" "test" {
  display_name               = "%s"
  description                = "%s"
  access_log_enabled         = %s
  application_profile_path   = data.nsxt_policy_lb_app_profile.default_tcp.path
  default_pool_member_ports  = ["%s"]
  enabled                    = %s
  ip_address                 = "%s"
  max_concurrent_connections = %s
  max_new_connection_rate    = %s
  ports                      = ["%s"]
  pool_path                  = nsxt_policy_lb_pool.pool.path
  sorry_pool_path            = nsxt_policy_lb_pool.sorry_pool.path
  persistence_profile_path   = data.nsxt_policy_lb_persistence_profile.default.path

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_virtual_server.test.path
}`, attrMap["display_name"], attrMap["description"], attrMap["access_log_enabled"], attrMap["default_pool_member_ports"], attrMap["enabled"], attrMap["ip_address"], attrMap["max_concurrent_connections"], attrMap["max_new_connection_rate"], attrMap["ports"])
}

func testAccNsxtPolicyLBVirtualServerMinimalistic() string {
	attrMap := accTestPolicyLBVirtualServerCreateAttributes
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "default_http"{
    type = "HTTP"
	display_name = "default-http-lb-app-profile"
}

resource "nsxt_policy_lb_pool" "pool" {
    display_name = "terraform-vs-test-pool"
}

resource "nsxt_policy_lb_pool" "sorry_pool" {
    display_name = "terraform-vs-test-sorry-pool"
}

resource "nsxt_policy_lb_virtual_server" "test" {
  display_name             = "%s"
  application_profile_path = data.nsxt_policy_lb_app_profile.default_http.path
  ip_address               = "%s"
  ports                    = ["%s"]
}

resource "nsxt_policy_group" "group1" {
  display_name = "terraform-vs-test-group1"
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_virtual_server.test.path
}`, accTestPolicyLBVirtualServerUpdateAttributes["display_name"], attrMap["ip_address"], accTestPolicyLBVirtualServerUpdateAttributes["ports"])
}

func testAccNsxtPolicyLBVirtualServerSSLTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBVirtualServerCreateAttributes
	} else {
		attrMap = accTestPolicyLBVirtualServerUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_certificate" "test" {
    display_name = "%s"
}
data "nsxt_policy_lb_app_profile" "default_tcp"{
    type = "TCP"
	display_name = "default-tcp-lb-app-profile"
}

data "nsxt_policy_lb_client_ssl_profile" "default" {
     display_name = "default-balanced-client-ssl-profile"
}

data "nsxt_policy_lb_server_ssl_profile" "default" {
     display_name = "default-balanced-server-ssl-profile"
}

resource "nsxt_policy_lb_virtual_server" "test" {
  display_name             = "%s"
  application_profile_path = data.nsxt_policy_lb_app_profile.default_tcp.path
  ip_address               = "%s"
  ports                    = ["%s"]
  client_ssl {
      client_auth              = "IGNORE"
      certificate_chain_depth  = %s
      ssl_profile_path         = data.nsxt_policy_lb_client_ssl_profile.default.path
      default_certificate_path = data.nsxt_policy_certificate.test.path
  }

  server_ssl {
      server_auth             = "IGNORE"
      certificate_chain_depth = %s
      ssl_profile_path        = data.nsxt_policy_lb_server_ssl_profile.default.path
  }
}`, getTestCertificateName(false), attrMap["display_name"], attrMap["ip_address"], attrMap["ports"], attrMap["certificate_chain_depth"], attrMap["certificate_chain_depth"])
}

func testAccNsxtPolicyLBVirtualServerWithAccessList(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBVirtualServerCreateAttributes
	} else {
		attrMap = accTestPolicyLBVirtualServerUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "default_tcp"{
  type = "TCP"
  display_name = "default-tcp-lb-app-profile"
}

data "nsxt_policy_lb_persistence_profile" "default" {
  type = "SOURCE_IP"
}

resource "nsxt_policy_lb_pool" "pool" {
  display_name = "terraform-vs-test-pool"
}

resource "nsxt_policy_group" "group1" {
  display_name = "terraform-vs-test-group1"
}

resource "nsxt_policy_lb_virtual_server" "test" {
  display_name               = "%s"
  access_log_enabled         = true
  application_profile_path   = data.nsxt_policy_lb_app_profile.default_tcp.path
  enabled                    = true
  ip_address                 = "%s"
  ports                      = ["%s"]
  pool_path                  = nsxt_policy_lb_pool.pool.path
  persistence_profile_path   = data.nsxt_policy_lb_persistence_profile.default.path
  log_significant_event_only = %s
  access_list_control {
      action     = "%s"
      enabled    = %s
      group_path = nsxt_policy_group.group1.path
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_virtual_server.test.path
}`, attrMap["display_name"], attrMap["ip_address"], attrMap["ports"], attrMap["log_significant_event_only"], attrMap["access_list_control_action"], attrMap["access_list_control_enabled"])
}

func testAccNsxtPolicyLBVirtualServerWithRules(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBVirtualServerCreateAttributes
	} else {
		attrMap = accTestPolicyLBVirtualServerUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "default_http"{
  type = "HTTP"
  display_name = "default-http-lb-app-profile"
}

data "nsxt_policy_lb_persistence_profile" "default" {
  type = "SOURCE_IP"
}

data "nsxt_policy_lb_persistence_profile" "generic" {
  type = "GENERIC"
}

resource "nsxt_policy_lb_pool" "pool" {
  display_name = "terraform-vs-test-pool"
}

resource "nsxt_policy_group" "group1" {
  display_name = "terraform-vs-test-group1"
}

resource "nsxt_policy_lb_virtual_server" "test" {
  display_name             = "%s"
  application_profile_path = data.nsxt_policy_lb_app_profile.default_http.path
  ip_address               = "%s"
  ports                    = [ "80" ]
  pool_path                  = nsxt_policy_lb_pool.pool.path

  rule {
	display_name = "connection_drop_test"
	action {
	  connection_drop {}
	}
  }
  rule {
	display_name = "http_redirect_test"
	action {
	  http_redirect {
            redirect_status = "301"
            redirect_url = "dummy"
	  }
	  http_redirect {
            redirect_status = "302"
            redirect_url = "other_dummy"
	  }
	}
	condition {
	  http_request_body {
  	    body_value = "xyz"
	    match_type = "REGEX"
	  }
	}
  }
  rule {
	display_name = "http_reject_test"
	action {
	  http_reject {
  	    reply_status = "404"
	  }
	  http_reject {
  	    reply_status = "400"
	  }
	}
	condition {
	  http_request_cookie {
            cookie_name = "is"
	    cookie_value = "brownie"
	    match_type = "REGEX"
	  }
	}
  }
  rule {
	display_name = "http_request_header_delete_test"
	phase = "HTTP_REQUEST_REWRITE"
	action {
	  http_request_header_delete {
	    header_name = "clever"
	  }
	  http_request_header_delete {
	    header_name = "smart"
	  }
	}
	condition {
	  http_request_header {
	    header_name = "my_head"
	    header_value = "has_hair"
	    match_type = "REGEX"
	  }
	}
  }
  rule {
	display_name = "http_request_header_rewrite_test"
	phase = "HTTP_REQUEST_REWRITE"
	action {
	  http_request_header_rewrite {
	    header_name = "clever"
	    header_value = "and"
	  }
	  http_request_header_rewrite {
	    header_name = "smart"
	    header_value = "!"
	  }
	}
	condition {
	  http_request_uri_arguments {
	    uri_arguments = "foo=bar"
	    match_type = "REGEX"
	  }
	  http_request_uri {
	    uri = "xyz"
	    match_type = "REGEX"
	  }
	}
  }
  rule {
	display_name = "http_request_uri_rewrite_test"
	phase = "HTTP_REQUEST_REWRITE"
	action {
	  http_request_uri_rewrite {
	    uri = "uri_test"
	  }
	  http_request_uri_rewrite {
            uri = "uri_test2"
	    uri_arguments = "foo=bar"
	  }
	}
	condition {
	  http_request_version {
	    version = "HTTP_VERSION_1_0"
	  }
	}
  }
  rule {
	display_name = "http_response_header_delete_test"
	phase = "HTTP_RESPONSE_REWRITE"
	action {
	  http_response_header_delete {
	    header_name = "clever"
          }
          http_response_header_delete {
	    header_name = "smart"
          }
	}
	condition {
	  http_response_header {
	    header_name = "clever"
	    header_value = "smart"
	    match_type = "REGEX"
	  }
	}
  }
  rule {
	display_name = "http_response_header_rewrite_test"
	phase = "HTTP_RESPONSE_REWRITE"
	action {
	  http_response_header_rewrite {
	    header_name = "clever"
	    header_value = "and"
	  }
	  http_response_header_rewrite {
	    header_name = "smart"
	    header_value = "!"
	  }
	}
	condition {
	  ip_header {
	    source_address = "1.1.1.1"
	  }
	}
  }
  rule {
	display_name = "select_pool_test"
	phase = "HTTP_FORWARDING"
	action {
	  select_pool {
            pool_id = nsxt_policy_lb_pool.pool.path
	  }
	}
        condition {
	  ssl_sni {
	    match_type = "REGEX"
	    sni = "HELO"
	  }
	}
  }
  rule {
	display_name = "variable_assignment_test"
	phase = "HTTP_ACCESS"
	action {
	  variable_assignment {
            variable_name = "foo"
	    variable_value = "bar"
	  }
	}
	condition {
	  variable {
	    variable_name = "my_var"
	    variable_value = "my_value"
	    match_type = "REGEX"
	  }
	  tcp_header {
		source_port = "80"
	  }
	}
  }
  rule {
	display_name = "variable_persistence_learn_test"
	phase = "HTTP_RESPONSE_REWRITE"
	action {
	  variable_persistence_learn {
            persistence_profile_path = data.nsxt_policy_lb_persistence_profile.generic.path
	    variable_hash_enabled = true
	    variable_name = "my_name"
	  }
	}
	condition {
	  http_ssl {
	    session_reused = "IGNORE"
	    used_protocol = "TLS_V1_2"
	    used_ssl_cipher = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	    client_certificate_issuer_dn {
		  issuer_dn = "something"
		  match_type = "REGEX"
	    }
	    client_certificate_subject_dn {
		  subject_dn = "something"
		  match_type = "REGEX"
	    }
	    client_supported_ssl_ciphers = [ "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256", "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384" ]
	  }
	}
  }
  rule {
	display_name = "variable_persistence_learn_minimal_test"
	phase = "HTTP_RESPONSE_REWRITE"
	action {
	  variable_persistence_learn {
            persistence_profile_path = data.nsxt_policy_lb_persistence_profile.generic.path
	    variable_name = "my_name"
	  }
	}
	condition {
	  http_ssl {
	    used_protocol = "TLS_V1_2"
	    used_ssl_cipher = "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256"
	  }
	}
  }
  rule {
	display_name = "variable_persistence_on_test"
	phase = "HTTP_FORWARDING"
	action {
	  variable_persistence_on {
            persistence_profile_path = data.nsxt_policy_lb_persistence_profile.generic.path
	    variable_hash_enabled = false
	    variable_name = "my_name"
	  }
	}
  }
  rule {
	display_name = "jwt_auth_test"
	phase = "HTTP_ACCESS"
	action {
	  jwt_auth {
	    key {
	  	  // one of ...
		  public_key_content = "xxx"
		  //certificate_path = "/path/to/cert"
		  // this one only indicates presence, the value is discarded
		  //symmetric_key = "dummy_value"
	    }
	    pass_jwt_to_pool = true
	    realm = "realm"
	    // only one token allowed currently
	    tokens = [ "a" ]
	  }
	}
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_virtual_server.test.path
}`, attrMap["display_name"], attrMap["ip_address"])
}
