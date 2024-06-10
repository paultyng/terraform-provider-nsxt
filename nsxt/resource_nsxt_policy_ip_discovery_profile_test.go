/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyIPDiscoveryProfileCreateAttributes = map[string]string{
	"display_name":                   getAccTestResourceName(),
	"description":                    "terraform created",
	"arp_nd_binding_timeout":         "15",
	"duplicate_ip_detection_enabled": "true",
	"arp_binding_limit":              "100",
	"arp_snooping_enabled":           "true",
	"dhcp_snooping_enabled":          "true",
	"vmtools_enabled":                "true",
	"dhcp_snooping_v6_enabled":       "true",
	"nd_snooping_enabled":            "true",
	"nd_snooping_limit":              "8",
	"vmtools_v6_enabled":             "true",
	"tofu_enabled":                   "true",
}

var accTestPolicyIPDiscoveryProfileUpdateAttributes = map[string]string{
	"display_name":                   getAccTestResourceName(),
	"description":                    "terraform updated",
	"arp_nd_binding_timeout":         "20",
	"duplicate_ip_detection_enabled": "false",
	"arp_binding_limit":              "140",
	"arp_snooping_enabled":           "false",
	"dhcp_snooping_enabled":          "false",
	"vmtools_enabled":                "false",
	"dhcp_snooping_v6_enabled":       "false",
	"nd_snooping_enabled":            "false",
	"nd_snooping_limit":              "12",
	"vmtools_v6_enabled":             "false",
	"tofu_enabled":                   "false",
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyIPDiscoveryProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIPDiscoveryProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyIPDiscoveryProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyIPDiscoveryProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPDiscoveryProfileCheckDestroy(state, accTestPolicyIPDiscoveryProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileTemplate(true, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPDiscoveryProfileExists(accTestPolicyIPDiscoveryProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPDiscoveryProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPDiscoveryProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_nd_binding_timeout", accTestPolicyIPDiscoveryProfileCreateAttributes["arp_nd_binding_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "duplicate_ip_detection_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["duplicate_ip_detection_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_binding_limit", accTestPolicyIPDiscoveryProfileCreateAttributes["arp_binding_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_snooping_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["arp_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["dhcp_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["vmtools_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_v6_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["dhcp_snooping_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["nd_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_limit", accTestPolicyIPDiscoveryProfileCreateAttributes["nd_snooping_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_v6_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["vmtools_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "tofu_enabled", accTestPolicyIPDiscoveryProfileCreateAttributes["tofu_enabled"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileTemplate(false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPDiscoveryProfileExists(accTestPolicyIPDiscoveryProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPDiscoveryProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPDiscoveryProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_nd_binding_timeout", accTestPolicyIPDiscoveryProfileUpdateAttributes["arp_nd_binding_timeout"]),
					resource.TestCheckResourceAttr(testResourceName, "duplicate_ip_detection_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["duplicate_ip_detection_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_binding_limit", accTestPolicyIPDiscoveryProfileUpdateAttributes["arp_binding_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "arp_snooping_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["arp_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["dhcp_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["vmtools_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_v6_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["dhcp_snooping_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["nd_snooping_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_limit", accTestPolicyIPDiscoveryProfileUpdateAttributes["nd_snooping_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_v6_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["vmtools_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "tofu_enabled", accTestPolicyIPDiscoveryProfileUpdateAttributes["tofu_enabled"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileMinimalistic(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPDiscoveryProfileExists(accTestPolicyIPDiscoveryProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "arp_nd_binding_timeout", "10"),
					resource.TestCheckResourceAttr(testResourceName, "duplicate_ip_detection_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "arp_binding_limit", "1"),
					resource.TestCheckResourceAttr(testResourceName, "arp_snooping_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_snooping_v6_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "nd_snooping_limit", "3"),
					resource.TestCheckResourceAttr(testResourceName, "vmtools_v6_enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tofu_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileMinimalistic(true),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testResourceName),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPDiscoveryProfile_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPDiscoveryProfileMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyIPDiscoveryProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPDiscoveryProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy IPDiscoveryProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIPDiscoveryProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy IPDiscoveryProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyIPDiscoveryProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ip_discovery_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyIPDiscoveryProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy IPDiscoveryProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIPDiscoveryProfileTemplate(createFlow, withContext bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPDiscoveryProfileCreateAttributes
	} else {
		attrMap = accTestPolicyIPDiscoveryProfileUpdateAttributes
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ip_discovery_profile" "test" {
%s
  display_name = "%s"
  description  = "%s"
  arp_nd_binding_timeout = %s
  duplicate_ip_detection_enabled = %s
  arp_binding_limit = %s
  arp_snooping_enabled = %s
  dhcp_snooping_enabled = %s
  vmtools_enabled = %s
  dhcp_snooping_v6_enabled = %s
  nd_snooping_enabled = %s
  nd_snooping_limit = %s
  vmtools_v6_enabled = %s
  tofu_enabled = %s
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, attrMap["display_name"], attrMap["description"], attrMap["arp_nd_binding_timeout"], attrMap["duplicate_ip_detection_enabled"], attrMap["arp_binding_limit"], attrMap["arp_snooping_enabled"], attrMap["dhcp_snooping_enabled"], attrMap["vmtools_enabled"], attrMap["dhcp_snooping_v6_enabled"], attrMap["nd_snooping_enabled"], attrMap["nd_snooping_limit"], attrMap["vmtools_v6_enabled"], attrMap["tofu_enabled"])
}

func testAccNsxtPolicyIPDiscoveryProfileMinimalistic(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ip_discovery_profile" "test" {
%s
  display_name = "%s"
}`, context, accTestPolicyIPDiscoveryProfileUpdateAttributes["display_name"])
}
