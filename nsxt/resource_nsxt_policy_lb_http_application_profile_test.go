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

var accTestPolicyLBHttpApplicationProfileCreateAttributes = map[string]string{
	"display_name":           getAccTestResourceName(),
	"description":            "terraform created",
	"http_redirect_to":       "http://www.test-create.com",
	"http_redirect_to_https": "false",
	"idle_timeout":           "15",
	"request_body_size":      "256",
	"request_header_size":    "1024",
	"response_buffering":     "true",
	"response_header_size":   "2048",
	"response_timeout":       "20",
	"server_keep_alive":      "true",
	"x_forwarded_for":        "REPLACE",
}

var accTestPolicyLBHttpApplicationProfileUpdateAttributes = map[string]string{
	"display_name":           getAccTestResourceName(),
	"description":            "terraform updated",
	"http_redirect_to":       "http://www.test-update.com",
	"http_redirect_to_https": "false",
	"idle_timeout":           "20",
	"request_body_size":      "512",
	"request_header_size":    "2048",
	"response_buffering":     "false",
	"response_header_size":   "4096",
	"response_timeout":       "10",
	"server_keep_alive":      "false",
	"x_forwarded_for":        "INSERT",
}

func TestAccResourceNsxtPolicyLBHttpApplicationProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_http_application_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpApplicationProfileCheckDestroy(state, accTestPolicyLBHttpApplicationProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpApplicationProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpApplicationProfileExists(accTestPolicyLBHttpApplicationProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpApplicationProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpApplicationProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to", accTestPolicyLBHttpApplicationProfileCreateAttributes["http_redirect_to"]),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to_https", accTestPolicyLBHttpApplicationProfileCreateAttributes["http_redirect_to_https"]),
					resource.TestCheckResourceAttr(testResourceName, "idle_timeout", accTestPolicyLBHttpApplicationProfileCreateAttributes["idle_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body_size", accTestPolicyLBHttpApplicationProfileCreateAttributes["request_body_size"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header_size", accTestPolicyLBHttpApplicationProfileCreateAttributes["request_header_size"]),
					resource.TestCheckResourceAttr(testResourceName, "response_buffering", accTestPolicyLBHttpApplicationProfileCreateAttributes["response_buffering"]),
					resource.TestCheckResourceAttr(testResourceName, "response_header_size", accTestPolicyLBHttpApplicationProfileCreateAttributes["response_header_size"]),
					resource.TestCheckResourceAttr(testResourceName, "response_timeout", accTestPolicyLBHttpApplicationProfileCreateAttributes["response_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "server_keep_alive", accTestPolicyLBHttpApplicationProfileCreateAttributes["server_keep_alive"]),
					resource.TestCheckResourceAttr(testResourceName, "x_forwarded_for", accTestPolicyLBHttpApplicationProfileCreateAttributes["x_forwarded_for"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpApplicationProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpApplicationProfileExists(accTestPolicyLBHttpApplicationProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBHttpApplicationProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBHttpApplicationProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to", accTestPolicyLBHttpApplicationProfileUpdateAttributes["http_redirect_to"]),
					resource.TestCheckResourceAttr(testResourceName, "http_redirect_to_https", accTestPolicyLBHttpApplicationProfileUpdateAttributes["http_redirect_to_https"]),
					resource.TestCheckResourceAttr(testResourceName, "idle_timeout", accTestPolicyLBHttpApplicationProfileUpdateAttributes["idle_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "request_body_size", accTestPolicyLBHttpApplicationProfileUpdateAttributes["request_body_size"]),
					resource.TestCheckResourceAttr(testResourceName, "request_header_size", accTestPolicyLBHttpApplicationProfileUpdateAttributes["request_header_size"]),
					resource.TestCheckResourceAttr(testResourceName, "response_buffering", accTestPolicyLBHttpApplicationProfileUpdateAttributes["response_buffering"]),
					resource.TestCheckResourceAttr(testResourceName, "response_header_size", accTestPolicyLBHttpApplicationProfileUpdateAttributes["response_header_size"]),
					resource.TestCheckResourceAttr(testResourceName, "response_timeout", accTestPolicyLBHttpApplicationProfileUpdateAttributes["response_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "server_keep_alive", accTestPolicyLBHttpApplicationProfileUpdateAttributes["server_keep_alive"]),
					resource.TestCheckResourceAttr(testResourceName, "x_forwarded_for", accTestPolicyLBHttpApplicationProfileUpdateAttributes["x_forwarded_for"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBHttpApplicationProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBHttpApplicationProfileExists(accTestPolicyLBHttpApplicationProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyLBHttpApplicationProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_http_application_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBHttpApplicationProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBHttpApplicationProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBHttpApplicationProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBHttpProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBHttpProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBHttpApplicationProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBHttpProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBHttpApplicationProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_http_application_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBHttpApplicationProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBHttpProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBHttpApplicationProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBHttpApplicationProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBHttpApplicationProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_http_application_profile" "test" {
  display_name = "%s"
  description  = "%s"
  http_redirect_to = "%s"
  http_redirect_to_https = %s
  idle_timeout = %s
  request_body_size = %s
  request_header_size = %s
  response_buffering = %s
  response_header_size = %s
  response_timeout = %s
  server_keep_alive = %s
  x_forwarded_for = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["http_redirect_to"], attrMap["http_redirect_to_https"], attrMap["idle_timeout"], attrMap["request_body_size"], attrMap["request_header_size"], attrMap["response_buffering"], attrMap["response_header_size"], attrMap["response_timeout"], attrMap["server_keep_alive"], attrMap["x_forwarded_for"])
}

func testAccNsxtPolicyLBHttpApplicationProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_lb_http_application_profile" "test" {
  display_name = "%s"

}`, accTestPolicyLBHttpApplicationProfileUpdateAttributes["display_name"])
}
