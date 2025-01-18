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

var accTestPolicyLBTcpMonitorProfileCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"receive":      "test-create",
	"send":         "test-create",
	"fall_count":   "2",
	"interval":     "2",
	"monitor_port": "8080",
	"rise_count":   "2",
	"timeout":      "2",
}

var accTestPolicyLBTcpMonitorProfileUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"receive":      "test-update",
	"send":         "test-update",
	"fall_count":   "5",
	"interval":     "5",
	"monitor_port": "8090",
	"rise_count":   "5",
	"timeout":      "5",
}

func TestAccResourceNsxtPolicyLBTcpMonitorProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_tcp_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBTcpMonitorProfileCheckDestroy(state, accTestPolicyLBTcpMonitorProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBTcpMonitorProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBTcpMonitorProfileExists(accTestPolicyLBTcpMonitorProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBTcpMonitorProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBTcpMonitorProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "receive", accTestPolicyLBTcpMonitorProfileCreateAttributes["receive"]),
					resource.TestCheckResourceAttr(testResourceName, "send", accTestPolicyLBTcpMonitorProfileCreateAttributes["send"]),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBTcpMonitorProfileCreateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBTcpMonitorProfileCreateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBTcpMonitorProfileCreateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBTcpMonitorProfileCreateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBTcpMonitorProfileCreateAttributes["timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBTcpMonitorProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBTcpMonitorProfileExists(accTestPolicyLBTcpMonitorProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBTcpMonitorProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBTcpMonitorProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "receive", accTestPolicyLBTcpMonitorProfileUpdateAttributes["receive"]),
					resource.TestCheckResourceAttr(testResourceName, "send", accTestPolicyLBTcpMonitorProfileUpdateAttributes["send"]),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", accTestPolicyLBTcpMonitorProfileUpdateAttributes["fall_count"]),
					resource.TestCheckResourceAttr(testResourceName, "interval", accTestPolicyLBTcpMonitorProfileUpdateAttributes["interval"]),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", accTestPolicyLBTcpMonitorProfileUpdateAttributes["monitor_port"]),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", accTestPolicyLBTcpMonitorProfileUpdateAttributes["rise_count"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBTcpMonitorProfileUpdateAttributes["timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBTcpMonitorProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBTcpMonitorProfileExists(accTestPolicyLBTcpMonitorProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyLBTcpMonitorProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_tcp_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBTcpMonitorProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBTcpMonitorProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBTcpMonitorProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBTcpMonitorProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBTcpMonitorProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBTcpMonitorProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBTcpMonitorProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_tcp_monitor_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBTcpMonitorProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBTcpMonitorProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBTcpMonitorProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBTcpMonitorProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_tcp_monitor_profile" "test" {
  display_name = "%s"
  description  = "%s"
  receive = "%s"
  send = "%s"
  fall_count = %s
  interval = %s
  monitor_port = %s
  rise_count = %s
  timeout = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["receive"], attrMap["send"], attrMap["fall_count"], attrMap["interval"], attrMap["monitor_port"], attrMap["rise_count"], attrMap["timeout"])
}

func testAccNsxtPolicyLBTcpMonitorProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_lb_tcp_monitor_profile" "test" {
  display_name = "%s"

}`, accTestPolicyLBTcpMonitorProfileUpdateAttributes["display_name"])
}
