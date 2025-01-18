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

var accTestPolicyIPSecVpnTunnelProfileCreateAttributes = map[string]string{
	"display_name":                   getAccTestResourceName(),
	"description":                    "terraform created",
	"df_policy":                      "COPY",
	"dh_groups":                      "GROUP2",
	"digest_algorithms":              "SHA2_256",
	"enable_perfect_forward_secrecy": "true",
	"encryption_algorithms":          "AES_256",
	"sa_life_time":                   "46000",
}

var accTestPolicyIPSecVpnTunnelProfileUpdateAttributes = map[string]string{
	"display_name":                   getAccTestResourceName(),
	"description":                    "terraform updated",
	"df_policy":                      "CLEAR",
	"dh_groups":                      "GROUP5",
	"digest_algorithms":              "SHA1",
	"enable_perfect_forward_secrecy": "false",
	"encryption_algorithms":          "AES_128",
	"sa_life_time":                   "28000",
}

func TestAccResourceNsxtPolicyIPSecVpnTunnelProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_ipsec_vpn_tunnel_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnTunnelProfileCheckDestroy(state, accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnTunnelProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnTunnelProfileExists(accTestPolicyIPSecVpnTunnelProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "df_policy", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["df_policy"]),
					resource.TestCheckResourceAttr(testResourceName, "dh_groups.0", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["dh_groups"]),
					resource.TestCheckResourceAttr(testResourceName, "digest_algorithms.0", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["digest_algorithms"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_perfect_forward_secrecy", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["enable_perfect_forward_secrecy"]),
					resource.TestCheckResourceAttr(testResourceName, "encryption_algorithms.0", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["encryption_algorithms"]),
					resource.TestCheckResourceAttr(testResourceName, "sa_life_time", accTestPolicyIPSecVpnTunnelProfileCreateAttributes["sa_life_time"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnTunnelProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnTunnelProfileExists(accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "df_policy", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["df_policy"]),
					resource.TestCheckResourceAttr(testResourceName, "dh_groups.0", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["dh_groups"]),
					resource.TestCheckResourceAttr(testResourceName, "digest_algorithms.0", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["digest_algorithms"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_perfect_forward_secrecy", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["enable_perfect_forward_secrecy"]),
					resource.TestCheckResourceAttr(testResourceName, "encryption_algorithms.0", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["encryption_algorithms"]),
					resource.TestCheckResourceAttr(testResourceName, "sa_life_time", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["sa_life_time"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnTunnelProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnTunnelProfileExists(accTestPolicyIPSecVpnTunnelProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicyIPSecVpnTunnelProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ipsec_vpn_tunnel_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnTunnelProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnTunnelProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnTunnelProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPSecVpnTunnelProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy IPSecVpnTunnelProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIPSecVpnTunnelProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy IPSecVpnTunnelProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyIPSecVpnTunnelProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ipsec_vpn_tunnel_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyIPSecVpnTunnelProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy IPSecVpnTunnelProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIPSecVpnTunnelProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnTunnelProfileCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnTunnelProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_tunnel_profile" "test" {
  display_name                   = "%s"
  description                    = "%s"
  df_policy                      = "%s"
  dh_groups                      = ["%s"]
  digest_algorithms              = ["%s"]
  enable_perfect_forward_secrecy = %s
  encryption_algorithms          = ["%s"]
  sa_life_time                   = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["df_policy"], attrMap["dh_groups"], attrMap["digest_algorithms"], attrMap["enable_perfect_forward_secrecy"], attrMap["encryption_algorithms"], attrMap["sa_life_time"])
}

func testAccNsxtPolicyIPSecVpnTunnelProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_tunnel_profile" "test" {
  display_name          = "%s"
  encryption_algorithms = ["%s"]
  dh_groups             = ["%s"]
}`, accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["display_name"], "AES_GCM_192", accTestPolicyIPSecVpnTunnelProfileUpdateAttributes["dh_groups"])
}
