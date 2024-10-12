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

var accTestVtepHAHostSwitchProfileCreateAttributes = map[string]string{
	"display_name":               getAccTestResourceName(),
	"description":                "terraform created",
	"auto_recovery":              "true",
	"auto_recovery_initial_wait": "3003",
	"auto_recovery_max_backoff":  "80000",
	"enabled":                    "true",
	"failover_timeout":           "40",
}

var accTestVtepHAHostSwitchProfileUpdateAttributes = map[string]string{
	"display_name":               getAccTestResourceName(),
	"description":                "terraform updated",
	"auto_recovery":              "true",
	"auto_recovery_initial_wait": "1001",
	"auto_recovery_max_backoff":  "80008",
	"enabled":                    "false",
	"failover_timeout":           "44",
}

func TestAccResourceNsxtPolicyVtepHAHostSwitchProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_vtep_ha_host_switch_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVtepHAHostSwitchProfileCheckDestroy(state, accTestVtepHAHostSwitchProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVtepHAHostSwitchProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVtepHAHostSwitchProfileExists(accTestVtepHAHostSwitchProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVtepHAHostSwitchProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVtepHAHostSwitchProfileCreateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "auto_recovery", accTestVtepHAHostSwitchProfileCreateAttributes["auto_recovery"]),
					resource.TestCheckResourceAttr(testResourceName, "auto_recovery_initial_wait", accTestVtepHAHostSwitchProfileCreateAttributes["auto_recovery_initial_wait"]),
					resource.TestCheckResourceAttr(testResourceName, "auto_recovery_max_backoff", accTestVtepHAHostSwitchProfileCreateAttributes["auto_recovery_max_backoff"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestVtepHAHostSwitchProfileCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "failover_timeout", accTestVtepHAHostSwitchProfileCreateAttributes["failover_timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVtepHAHostSwitchProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVtepHAHostSwitchProfileExists(accTestVtepHAHostSwitchProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVtepHAHostSwitchProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVtepHAHostSwitchProfileUpdateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "auto_recovery", accTestVtepHAHostSwitchProfileUpdateAttributes["auto_recovery"]),
					resource.TestCheckResourceAttr(testResourceName, "auto_recovery_initial_wait", accTestVtepHAHostSwitchProfileUpdateAttributes["auto_recovery_initial_wait"]),
					resource.TestCheckResourceAttr(testResourceName, "auto_recovery_max_backoff", accTestVtepHAHostSwitchProfileUpdateAttributes["auto_recovery_max_backoff"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestVtepHAHostSwitchProfileUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "failover_timeout", accTestVtepHAHostSwitchProfileUpdateAttributes["failover_timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVtepHAHostSwitchProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVtepHAHostSwitchProfileExists(accTestVtepHAHostSwitchProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyVtepHAHostSwitchProfile_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_vtep_ha_host_switch_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVtepHAHostSwitchProfileCheckDestroy(state, accTestVtepHAHostSwitchProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVtepHAHostSwitchProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtVtepHAHostSwitchProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("VtepHAHostSwitchProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("VtepHAHostSwitchProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtVtepHAHostSwitchProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("VtepHAHostSwitchProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVtepHAHostSwitchProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_vtep_ha_host_switch_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVtepHAHostSwitchProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("VtepHAHostSwitchProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVtepHAHostSwitchProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVtepHAHostSwitchProfileCreateAttributes
	} else {
		attrMap = accTestVtepHAHostSwitchProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_vtep_ha_host_switch_profile" "test" {
  display_name = "%s"
  description  = "%s"
  auto_recovery = "%s"
  auto_recovery_initial_wait = "%s"
  auto_recovery_max_backoff = "%s"
  enabled = "%s"
  failover_timeout = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
data "nsxt_policy_vtep_ha_host_switch_profile" "test" {
  display_name = "%s"
  depends_on = [nsxt_policy_vtep_ha_host_switch_profile.test]
}`, attrMap["display_name"], attrMap["description"], attrMap["auto_recovery"], attrMap["auto_recovery_initial_wait"], attrMap["auto_recovery_max_backoff"], attrMap["enabled"], attrMap["failover_timeout"], attrMap["display_name"])
}

func testAccNsxtVtepHAHostSwitchProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_vtep_ha_host_switch_profile" "test" {
  display_name = "%s"
}`, accTestVtepHAHostSwitchProfileUpdateAttributes["display_name"])
}
