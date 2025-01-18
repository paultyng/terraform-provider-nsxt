// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var accTestPolicyOspfConfigCreateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform created",
	"enabled":               "false",
	"ecmp":                  "true",
	"prefix":                "20.0.1.0/24",
	"advertise":             "true",
	"default_originate":     "true",
	"graceful_restart_mode": "HELPER_ONLY",
}

var accTestPolicyOspfConfigUpdateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform updated",
	"enabled":               "true",
	"ecmp":                  "false",
	"prefix":                "20.0.2.0/24",
	"advertise":             "false",
	"default_originate":     "false",
	"graceful_restart_mode": "DISABLE",
}

var accTestPolicyOspfConfigHelperName = getAccTestResourceName()

func TestAccResourceNsxtPolicyOspfConfig_basic(t *testing.T) {
	testResourceName := "nsxt_policy_ospf_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.1.1") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyOspfConfigTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfConfigCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyOspfConfigCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyOspfConfigCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ecmp", accTestPolicyOspfConfigCreateAttributes["ecmp"]),
					resource.TestCheckResourceAttr(testResourceName, "default_originate", accTestPolicyOspfConfigCreateAttributes["default_originate"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyOspfConfigCreateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.0.prefix", accTestPolicyOspfConfigCreateAttributes["prefix"]),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.0.advertise", accTestPolicyOspfConfigCreateAttributes["advertise"]),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyOspfConfigTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyOspfConfigUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyOspfConfigUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ecmp", accTestPolicyOspfConfigUpdateAttributes["ecmp"]),
					resource.TestCheckResourceAttr(testResourceName, "default_originate", accTestPolicyOspfConfigUpdateAttributes["default_originate"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyOspfConfigUpdateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.0.prefix", accTestPolicyOspfConfigUpdateAttributes["prefix"]),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.0.advertise", accTestPolicyOspfConfigUpdateAttributes["advertise"]),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyOspfConfigMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "summary_address.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyOspfConfig_minimalistic(t *testing.T) {
	testResourceName := "nsxt_policy_ospf_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.1.1") },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyOspfConfigMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func testAccNsxtPolicyOspfConfigPrerequisites() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
}`, getEdgeClusterName(), accTestPolicyOspfConfigHelperName)
}

func testAccNsxtPolicyOspfConfigTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyOspfConfigCreateAttributes
	} else {
		attrMap = accTestPolicyOspfConfigUpdateAttributes
	}
	return testAccNsxtPolicyOspfConfigPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_ospf_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path

  display_name          = "%s"
  description           = "%s"
  enabled               = %s
  ecmp                  = %s
  default_originate     = %s
  graceful_restart_mode = "%s"

  summary_address {
    prefix    = "%s"
    advertise = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["ecmp"], attrMap["default_originate"], attrMap["graceful_restart_mode"], attrMap["prefix"], attrMap["advertise"])
}

func testAccNsxtPolicyOspfConfigMinimalistic() string {
	return testAccNsxtPolicyOspfConfigPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_ospf_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
}`, accTestPolicyOspfConfigUpdateAttributes["display_name"])
}
