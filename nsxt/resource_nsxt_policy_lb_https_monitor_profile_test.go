// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyLBHttpsMonitorProfileCreateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform created",
	"request_body":            "test-create",
	"request_method":          "POST",
	"request_url":             "test-create",
	"request_version":         "HTTP_VERSION_1_1",
	"response_body":           "test-create",
	"response_status_codes":   "200",
	"fall_count":              "2",
	"interval":                "2",
	"monitor_port":            "8080",
	"rise_count":              "2",
	"timeout":                 "2",
	"header_name":             "test-create",
	"header_value":            "test-create",
	"certificate_chain_depth": "2",
	"server_auth":             "IGNORE",
}

var accTestPolicyLBHttpsMonitorProfileUpdateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform updated",
	"request_body":            "test-update",
	"request_method":          "OPTIONS",
	"request_url":             "test-update",
	"request_version":         "HTTP_VERSION_1_0",
	"response_body":           "test-update",
	"response_status_codes":   "200",
	"fall_count":              "5",
	"interval":                "5",
	"monitor_port":            "8090",
	"rise_count":              "5",
	"timeout":                 "5",
	"header_name":             "test-update",
	"header_value":            "test-update",
	"certificate_chain_depth": "5",
	"server_auth":             "IGNORE",
}

func TestAccResourceNsxtPolicyLBHttpsMonitorProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_https_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpsMonitorProfileCheckDestroy(state, accTestPolicyLBHttpsMonitorProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpsMonitorProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpsMonitorProfileExists(accTestPolicyLBHttpsMonitorProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpsMonitorProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpsMonitorProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body", accTestPolicyLBHttpsMonitorProfileCreateAttributes["request_body"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", accTestPolicyLBHttpsMonitorProfileCreateAttributes["request_method"]),
					resource.TestCheckResourceAttr(testResourceName, "request_url", accTestPolicyLBHttpsMonitorProfileCreateAttributes["request_url"]),
					resource.TestCheckResourceAttr(testResourceName, "request_version", accTestPolicyLBHttpsMonitorProfileCreateAttributes["request_version"]),
					resource.TestCheckResourceAttr(testResourceName, "response_body", accTestPolicyLBHttpsMonitorProfileCreateAttributes["response_body"]),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.0", accTestPolicyLBHttpsMonitorProfileCreateAttributes["response_status_codes"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBHttpsMonitorProfileCreateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBHttpsMonitorProfileCreateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBHttpsMonitorProfileCreateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBHttpsMonitorProfileCreateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBHttpsMonitorProfileCreateAttributes["timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.name", accTestPolicyLBHttpsMonitorProfileCreateAttributes["header_name"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.value", accTestPolicyLBHttpsMonitorProfileCreateAttributes["header_value"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.certificate_chain_depth", accTestPolicyLBHttpsMonitorProfileCreateAttributes["certificate_chain_depth"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.server_auth", accTestPolicyLBHttpsMonitorProfileCreateAttributes["server_auth"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpsMonitorProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpsMonitorProfileExists(accTestPolicyLBHttpsMonitorProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["request_body"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["request_method"]),
					resource.TestCheckResourceAttr(testResourceName, "request_url", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["request_url"]),
					resource.TestCheckResourceAttr(testResourceName, "request_version", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["request_version"]),
					resource.TestCheckResourceAttr(testResourceName, "response_body", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["response_body"]),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.0", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["response_status_codes"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.name", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["header_name"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.value", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["header_value"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.certificate_chain_depth", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["certificate_chain_depth"]),
					resource.TestCheckResourceAttr(testResourceName, "server_ssl.0.server_auth", accTestPolicyLBHttpsMonitorProfileUpdateAttributes["server_auth"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpsMonitorProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpsMonitorProfileExists(accTestPolicyLBHttpsMonitorProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBHttpsMonitorProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_https_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpsMonitorProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpsMonitorProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBHttpsMonitorProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBHttpsMonitorProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBHttpsMonitorProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBHttpsMonitorProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBHttpsMonitorProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_https_monitor_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBHttpsMonitorProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBHttpsMonitorProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBHttpsMonitorProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBHttpsMonitorProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_https_monitor_profile" "test" {
  display_name = "%s"
  description  = "%s"
  request_body = "%s"


  request_method = "%s"
  request_url = "%s"
  request_version = "%s"
  response_body = "%s"
  response_status_codes = [%s]

  request_header {
    name = "%s"
    value = "%s"
  }

  fall_count = %s
  interval = %s
  monitor_port = %s
  rise_count = %s
  timeout = %s

  server_ssl {
    certificate_chain_depth = %s
    server_auth = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["request_body"], attrMap["request_method"], attrMap["request_url"], attrMap["request_version"], attrMap["response_body"], attrMap["response_status_codes"], attrMap["header_name"], attrMap["header_value"], attrMap["fall_count"], attrMap["interval"], attrMap["monitor_port"], attrMap["rise_count"], attrMap["timeout"], attrMap["certificate_chain_depth"], attrMap["server_auth"])
}

func testAccNsxtPolicyLBHttpsMonitorProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_lb_https_monitor_profile" "test" {
  display_name = "%s"

}`, accTestPolicyLBHttpsMonitorProfileUpdateAttributes["display_name"])
}
