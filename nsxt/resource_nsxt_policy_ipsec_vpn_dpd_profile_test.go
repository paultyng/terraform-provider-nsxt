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

var accTestPolicyIPSecVpnDpdProfileCreateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform created",
	"dpd_probe_interval": "2",
	"dpd_probe_mode":     "ON_DEMAND",
	"enabled":            "true",
	"retry_count":        "2",
}

var accTestPolicyIPSecVpnDpdProfileUpdateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform updated",
	"dpd_probe_interval": "5",
	"dpd_probe_mode":     "PERIODIC",
	"enabled":            "false",
	"retry_count":        "5",
}

func TestAccResourceNsxtPolicyIPSecVpnDpdProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_ipsec_vpn_dpd_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnDpdProfileCheckDestroy(state, accTestPolicyIPSecVpnDpdProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnDpdProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnDpdProfileExists(accTestPolicyIPSecVpnDpdProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnDpdProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnDpdProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dpd_probe_interval", accTestPolicyIPSecVpnDpdProfileCreateAttributes["dpd_probe_interval"]),
					resource.TestCheckResourceAttr(testResourceName, "dpd_probe_mode", accTestPolicyIPSecVpnDpdProfileCreateAttributes["dpd_probe_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnDpdProfileCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "retry_count", accTestPolicyIPSecVpnDpdProfileCreateAttributes["retry_count"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnDpdProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnDpdProfileExists(accTestPolicyIPSecVpnDpdProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dpd_probe_interval", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["dpd_probe_interval"]),
					resource.TestCheckResourceAttr(testResourceName, "dpd_probe_mode", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["dpd_probe_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "retry_count", accTestPolicyIPSecVpnDpdProfileUpdateAttributes["retry_count"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnDpdProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnDpdProfileExists(accTestPolicyIPSecVpnDpdProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyIPSecVpnDpdProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ipsec_vpn_dpd_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnDpdProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnDpdProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnDpdProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPSecVpnDpdProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy IPSecVpnDpdProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIPSecVpnDpdProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy IPSecVpnDpdProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyIPSecVpnDpdProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ipsec_vpn_dpd_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyIPSecVpnDpdProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy IPSecVpnDpdProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIPSecVpnDpdProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnDpdProfileCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnDpdProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_dpd_profile" "test" {
  display_name       = "%s"
  description        = "%s"
  dpd_probe_interval = %s
  dpd_probe_mode     = "%s"
  enabled            = %s
  retry_count        = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["dpd_probe_interval"], attrMap["dpd_probe_mode"], attrMap["enabled"], attrMap["retry_count"])
}

func testAccNsxtPolicyIPSecVpnDpdProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_dpd_profile" "test" {
  display_name = "%s"
}`, accTestPolicyIPSecVpnDpdProfileUpdateAttributes["display_name"])
}
