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

func TestAccResourceNsxtLbHttpResponseRewriteRule_basic(t *testing.T) {
	name := getAccTestResourceName()
	fullName := "nsxt_lb_http_response_rewrite_rule.test"
	matchStrategy := "ANY"
	updatedMatchStrategy := "ALL"
	matchType := "CONTAINS"
	updatedMatchType := "REGEX"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPResponseRewriteRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPResponseRewriteRuleCreateTemplate(name, matchStrategy, matchType, "true", "false"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPResponseRewriteRuleExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "match_strategy", matchStrategy),
					testLbRuleConditionCount(fullName, "request_header", 1),
					testLbRuleConditionAttr(fullName, "request_header", "0", "name", "NAME1"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "value", "VALUE1"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "match_type", "EQUALS"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "case_sensitive", "true"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "response_header", 1),
					testLbRuleConditionAttr(fullName, "response_header", "0", "name", "NAME2"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "value", "VALUE2"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "match_type", matchType),
					testLbRuleConditionAttr(fullName, "response_header", "0", "case_sensitive", "true"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "cookie", 1),
					testLbRuleConditionAttr(fullName, "cookie", "0", "name", "NAME1"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "value", "VALUE1"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "match_type", matchType),
					testLbRuleConditionAttr(fullName, "cookie", "0", "case_sensitive", "true"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "method", 1),
					testLbRuleConditionAttr(fullName, "method", "0", "method", "POST"),
					testLbRuleConditionAttr(fullName, "method", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "version", 1),
					testLbRuleConditionAttr(fullName, "version", "0", "version", "HTTP_VERSION_1_0"),
					testLbRuleConditionAttr(fullName, "version", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "uri", 1),
					testLbRuleConditionAttr(fullName, "uri", "0", "uri", "/hello"),
					testLbRuleConditionAttr(fullName, "uri", "0", "match_type", matchType),
					testLbRuleConditionAttr(fullName, "uri", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "uri_arguments", 1),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "uri_arguments", "hello"),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "match_type", matchType),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "ip", 1),
					testLbRuleConditionAttr(fullName, "ip", "0", "source_address", "1.1.1.1"),
					testLbRuleConditionAttr(fullName, "ip", "0", "inverse", "false"),
					testLbRuleConditionCount(fullName, "tcp", 1),
					testLbRuleConditionAttr(fullName, "tcp", "0", "source_port", "7887"),
					testLbRuleConditionAttr(fullName, "tcp", "0", "inverse", "false"),

					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.#", "1"),
					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.0.name", "NAME1"),
					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.0.value", "VALUE1"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbHTTPResponseRewriteRuleCreateTemplate(name, updatedMatchStrategy, updatedMatchType, "false", "true"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbHTTPResponseRewriteRuleExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "match_strategy", updatedMatchStrategy),
					testLbRuleConditionCount(fullName, "request_header", 1),
					testLbRuleConditionAttr(fullName, "request_header", "0", "name", "NAME1"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "value", "VALUE1"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "match_type", "EQUALS"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "case_sensitive", "true"),
					testLbRuleConditionAttr(fullName, "request_header", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "response_header", 1),
					testLbRuleConditionAttr(fullName, "response_header", "0", "name", "NAME2"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "value", "VALUE2"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "match_type", updatedMatchType),
					testLbRuleConditionAttr(fullName, "response_header", "0", "case_sensitive", "false"),
					testLbRuleConditionAttr(fullName, "response_header", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "cookie", 1),
					testLbRuleConditionAttr(fullName, "cookie", "0", "name", "NAME1"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "value", "VALUE1"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "match_type", updatedMatchType),
					testLbRuleConditionAttr(fullName, "cookie", "0", "case_sensitive", "false"),
					testLbRuleConditionAttr(fullName, "cookie", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "method", 1),
					testLbRuleConditionAttr(fullName, "method", "0", "method", "POST"),
					testLbRuleConditionAttr(fullName, "method", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "version", 1),
					testLbRuleConditionAttr(fullName, "version", "0", "version", "HTTP_VERSION_1_0"),
					testLbRuleConditionAttr(fullName, "version", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "uri", 1),
					testLbRuleConditionAttr(fullName, "uri", "0", "uri", "/hello"),
					testLbRuleConditionAttr(fullName, "uri", "0", "match_type", updatedMatchType),
					testLbRuleConditionAttr(fullName, "uri", "0", "inverse", "true"),
					testLbRuleConditionAttr(fullName, "uri", "0", "case_sensitive", "true"),
					testLbRuleConditionCount(fullName, "uri_arguments", 1),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "uri_arguments", "hello"),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "match_type", updatedMatchType),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "inverse", "true"),
					testLbRuleConditionAttr(fullName, "uri_arguments", "0", "case_sensitive", "true"),
					testLbRuleConditionCount(fullName, "ip", 1),
					testLbRuleConditionAttr(fullName, "ip", "0", "source_address", "1.1.1.1"),
					testLbRuleConditionAttr(fullName, "ip", "0", "inverse", "true"),
					testLbRuleConditionCount(fullName, "tcp", 1),
					testLbRuleConditionAttr(fullName, "tcp", "0", "source_port", "7887"),
					testLbRuleConditionAttr(fullName, "tcp", "0", "inverse", "true"),

					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.#", "1"),
					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.0.name", "NAME1"),
					resource.TestCheckResourceAttr(fullName, "header_rewrite_action.0.value", "VALUE1"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbHttpResponseRewriteRule_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_lb_http_response_rewrite_rule.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbHTTPResponseRewriteRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbHTTPResponseRewriteRuleCreateTemplateTrivial(),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbHTTPResponseRewriteRuleExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB rule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB rule resource ID not set in resources ")
		}

		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerRule(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB rule with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB rule %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB rule %s wasn't found", displayName)
	}
}

func testAccNSXLbHTTPResponseRewriteRuleCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_http_response_rewrite_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerRule(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB rule with ID %s. Error: %v", resourceID, err)
		}

		if displayName == monitor.DisplayName {
			return fmt.Errorf("NSX LB rule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbHTTPResponseRewriteRuleCreateTemplate(name string, matchStrategy string, matchType string, caseSensitive string, inverse string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_http_response_rewrite_rule" "test" {
  display_name   = "%s"
  description    = "test description"
  match_strategy = "%s"
`, name, matchStrategy) +
		testAccNSXLbRuleNameValueConditionTemplate("request_header", "EQUALS", "true", "true", 1) +
		testAccNSXLbRuleNameValueConditionTemplate("response_header", matchType, caseSensitive, inverse, 2) +
		testAccNSXLbRuleNameValueConditionTemplate("cookie", matchType, caseSensitive, inverse, 1) +
		testAccNSXLbRuleSimpleConditionTemplate("method", "POST", inverse) +
		testAccNSXLbRuleSimpleConditionTemplate("version", "HTTP_VERSION_1_0", inverse) +
		testAccNSXLbRuleSimpleMatchConditionTemplate("uri", "/hello", matchType, inverse) +
		testAccNSXLbRuleSimpleMatchConditionTemplate("uri_arguments", "hello", matchType, inverse) +
		fmt.Sprintf(`
  ip_condition {
    source_address = "1.1.1.1"
    inverse        = %s
  }

  tcp_condition {
    source_port = "7887"
    inverse     = %s
  }

  header_rewrite_action {
    name  = "NAME1"
    value = "VALUE1"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, inverse, inverse)
}

func testAccNSXLbHTTPResponseRewriteRuleCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_http_response_rewrite_rule" "test" {
  description    = "test description"
  match_strategy = "ANY"

  response_header_condition {
    name = "NAME1"
    value = "VALUE1"
    match_type = "EQUALS"
  }

  header_rewrite_action {
    name  = "NAME1"
    value = "VALUE2"
  }
}
`
}
