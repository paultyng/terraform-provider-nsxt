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

var accTestPolicyLBHttpMonitorProfileCreateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform created",
	"request_body":          "test-create",
	"request_method":        "HEAD",
	"request_url":           "test-create",
	"request_version":       "HTTP_VERSION_1_1",
	"response_body":         "test-create",
	"response_status_codes": "200",
	"fall_count":            "2",
	"interval":              "2",
	"monitor_port":          "8080",
	"rise_count":            "2",
	"timeout":               "2",
	"header_name":           "test-create",
	"header_value":          "test-create",
}

var accTestPolicyLBHttpMonitorProfileUpdateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform updated",
	"request_body":          "test-update",
	"request_method":        "GET",
	"request_url":           "test-update",
	"request_version":       "HTTP_VERSION_1_0",
	"response_body":         "test-update",
	"response_status_codes": "200",
	"fall_count":            "5",
	"interval":              "5",
	"monitor_port":          "8090",
	"rise_count":            "5",
	"timeout":               "5",
	"header_name":           "test-update",
	"header_value":          "test-update",
}

func TestAccResourceNsxtPolicyLBHttpMonitorProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_http_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpMonitorProfileCheckDestroy(state, accTestPolicyLBHttpMonitorProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpMonitorProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpMonitorProfileExists(accTestPolicyLBHttpMonitorProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpMonitorProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpMonitorProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body", accTestPolicyLBHttpMonitorProfileCreateAttributes["request_body"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", accTestPolicyLBHttpMonitorProfileCreateAttributes["request_method"]),
					resource.TestCheckResourceAttr(testResourceName, "request_url", accTestPolicyLBHttpMonitorProfileCreateAttributes["request_url"]),
					resource.TestCheckResourceAttr(testResourceName, "request_version", accTestPolicyLBHttpMonitorProfileCreateAttributes["request_version"]),
					resource.TestCheckResourceAttr(testResourceName, "response_body", accTestPolicyLBHttpMonitorProfileCreateAttributes["response_body"]),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.0", accTestPolicyLBHttpMonitorProfileCreateAttributes["response_status_codes"]),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBHttpMonitorProfileCreateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBHttpMonitorProfileCreateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBHttpMonitorProfileCreateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBHttpMonitorProfileCreateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBHttpMonitorProfileCreateAttributes["timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.name", accTestPolicyLBHttpMonitorProfileCreateAttributes["header_name"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.value", accTestPolicyLBHttpMonitorProfileCreateAttributes["header_value"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpMonitorProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpMonitorProfileExists(accTestPolicyLBHttpMonitorProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpMonitorProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpMonitorProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body", accTestPolicyLBHttpMonitorProfileUpdateAttributes["request_body"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "request_method", accTestPolicyLBHttpMonitorProfileUpdateAttributes["request_method"]),
					resource.TestCheckResourceAttr(testResourceName, "request_url", accTestPolicyLBHttpMonitorProfileUpdateAttributes["request_url"]),
					resource.TestCheckResourceAttr(testResourceName, "request_version", accTestPolicyLBHttpMonitorProfileUpdateAttributes["request_version"]),
					resource.TestCheckResourceAttr(testResourceName, "response_body", accTestPolicyLBHttpMonitorProfileUpdateAttributes["response_body"]),
					resource.TestCheckResourceAttr(testResourceName, "response_status_codes.0", accTestPolicyLBHttpMonitorProfileUpdateAttributes["response_status_codes"]),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBHttpMonitorProfileUpdateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBHttpMonitorProfileUpdateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBHttpMonitorProfileUpdateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBHttpMonitorProfileUpdateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBHttpMonitorProfileUpdateAttributes["timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.name", accTestPolicyLBHttpMonitorProfileUpdateAttributes["header_name"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header.0.value", accTestPolicyLBHttpMonitorProfileUpdateAttributes["header_value"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpMonitorProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpMonitorProfileExists(accTestPolicyLBHttpMonitorProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyLBHttpMonitorProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_http_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpMonitorProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpMonitorProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBHttpMonitorProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBHttpMonitorProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBHttpMonitorProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBHttpMonitorProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBHttpMonitorProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_http_monitor_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBHttpMonitorProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBHttpMonitorProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBHttpMonitorProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBHttpMonitorProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_http_monitor_profile" "test" {
  display_name = "%s"
  description  = "%s"
  request_body = "%s"

  request_method = "%s"
  request_url = "%s"
  request_version = "%s"
  response_body = "%s"
  response_status_codes = [%s]
  fall_count = %s
  interval = %s
  monitor_port = %s
  rise_count = %s
  timeout = %s

  request_header {
    name = "%s"
    value = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["request_body"], attrMap["request_method"], attrMap["request_url"], attrMap["request_version"], attrMap["response_body"], attrMap["response_status_codes"], attrMap["fall_count"], attrMap["interval"], attrMap["monitor_port"], attrMap["rise_count"], attrMap["timeout"], attrMap["header_name"], attrMap["header_value"])
}

func testAccNsxtPolicyLBHttpMonitorProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_lb_http_monitor_profile" "test" {
  display_name = "%s"

}`, accTestPolicyLBHttpMonitorProfileUpdateAttributes["display_name"])
}
