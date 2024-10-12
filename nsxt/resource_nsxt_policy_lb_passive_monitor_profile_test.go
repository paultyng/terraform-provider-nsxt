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

var accTestPolicyLBPassiveMonitorProfileCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"max_fails":    "2",
	"timeout":      "2",
}

var accTestPolicyLBPassiveMonitorProfileUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"max_fails":    "5",
	"timeout":      "5",
}

func TestAccResourceNsxtPolicyLBPassiveMonitorProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_passive_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBPassiveMonitorProfileCheckDestroy(state, accTestPolicyLBPassiveMonitorProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBPassiveMonitorProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPassiveMonitorProfileExists(accTestPolicyLBPassiveMonitorProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPassiveMonitorProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPassiveMonitorProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "max_fails", accTestPolicyLBPassiveMonitorProfileCreateAttributes["max_fails"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBPassiveMonitorProfileCreateAttributes["timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPassiveMonitorProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPassiveMonitorProfileExists(accTestPolicyLBPassiveMonitorProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPassiveMonitorProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPassiveMonitorProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "max_fails", accTestPolicyLBPassiveMonitorProfileUpdateAttributes["max_fails"]),
					resource.TestCheckResourceAttr(testResourceName, "timeout", accTestPolicyLBPassiveMonitorProfileUpdateAttributes["timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPassiveMonitorProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPassiveMonitorProfileExists(accTestPolicyLBPassiveMonitorProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyLBPassiveMonitorProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_passive_monitor_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBPassiveMonitorProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBPassiveMonitorProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBPassiveMonitorProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBPassiveMonitorProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBPassiveMonitorProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBPassiveMonitorProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBPassiveMonitorProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_passive_monitor_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBMonitorProfileExistsWrapper(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBPassiveMonitorProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBPassiveMonitorProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBPassiveMonitorProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBPassiveMonitorProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_passive_monitor_profile" "test" {
  display_name = "%s"
  description  = "%s"
  max_fails = %s
  timeout = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["max_fails"], attrMap["timeout"])
}

func testAccNsxtPolicyLBPassiveMonitorProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_lb_passive_monitor_profile" "test" {
  display_name = "%s"

}`, accTestPolicyLBPassiveMonitorProfileUpdateAttributes["display_name"])
}
