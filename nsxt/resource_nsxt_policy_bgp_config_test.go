// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var accTestPolicyBgpConfigCreateAttributes = map[string]string{
	"enabled":                            "false",
	"inter_sr_ibgp":                      "false",
	"local_as_num":                       "1200.12",
	"multipath_relax":                    "true",
	"prefix":                             "20.0.1.0/24",
	"summary_only":                       "true",
	"graceful_restart_mode":              "HELPER_ONLY",
	"graceful_restart_timer":             "2000",
	"graceful_restart_stale_route_timer": "2100",
}

var accTestPolicyBgpConfigUpdateAttributes = map[string]string{
	"enabled":                            "true",
	"inter_sr_ibgp":                      "true",
	"local_as_num":                       "1200.13",
	"multipath_relax":                    "false",
	"prefix":                             "20.0.2.0/24",
	"summary_only":                       "false",
	"graceful_restart_mode":              "DISABLE",
	"graceful_restart_timer":             "3000",
	"graceful_restart_stale_route_timer": "3100",
}

func TestAccResourceNsxtPolicyBgpConfig_basic(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpConfigTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyBgpConfigCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "inter_sr_ibgp", accTestPolicyBgpConfigCreateAttributes["inter_sr_ibgp"]),
					resource.TestCheckResourceAttr(testResourceName, "local_as_num", accTestPolicyBgpConfigCreateAttributes["local_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "multipath_relax", accTestPolicyBgpConfigCreateAttributes["multipath_relax"]),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.0.prefix", accTestPolicyBgpConfigCreateAttributes["prefix"]),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.0.summary_only", accTestPolicyBgpConfigCreateAttributes["summary_only"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpConfigCreateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_timer", accTestPolicyBgpConfigCreateAttributes["graceful_restart_timer"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_stale_route_timer", accTestPolicyBgpConfigCreateAttributes["graceful_restart_stale_route_timer"]),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyBgpConfigTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyBgpConfigUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "inter_sr_ibgp", accTestPolicyBgpConfigUpdateAttributes["inter_sr_ibgp"]),
					resource.TestCheckResourceAttr(testResourceName, "local_as_num", accTestPolicyBgpConfigUpdateAttributes["local_as_num"]),
					resource.TestCheckResourceAttr(testResourceName, "multipath_relax", accTestPolicyBgpConfigUpdateAttributes["multipath_relax"]),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.0.prefix", accTestPolicyBgpConfigUpdateAttributes["prefix"]),
					resource.TestCheckResourceAttr(testResourceName, "route_aggregation.0.summary_only", accTestPolicyBgpConfigUpdateAttributes["summary_only"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_mode", accTestPolicyBgpConfigUpdateAttributes["graceful_restart_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_timer", accTestPolicyBgpConfigUpdateAttributes["graceful_restart_timer"]),
					resource.TestCheckResourceAttr(testResourceName, "graceful_restart_stale_route_timer", accTestPolicyBgpConfigUpdateAttributes["graceful_restart_stale_route_timer"]),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyBgpConfig_minimalistic(t *testing.T) {
	testResourceName := "nsxt_policy_bgp_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyBgpConfigMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func testAccNsxtPolicyBgpConfigTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyBgpConfigCreateAttributes
	} else {
		attrMap = accTestPolicyBgpConfigUpdateAttributes
	}
	extraConfig := ""
	if testAccIsGlobalManager() {
		extraConfig = `site_path    = data.nsxt_policy_site.test.path`
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) + testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_bgp_config" "test" {
  enabled         = %s
  inter_sr_ibgp   = %s
  local_as_num    = "%s"
  multipath_relax = %s

  route_aggregation {
    prefix       = "%s"
    summary_only = "%s"
  }

  graceful_restart_mode              = "%s"
  graceful_restart_timer             = %s
  graceful_restart_stale_route_timer = %s

  gateway_path = nsxt_policy_tier0_gateway.test.path
  %s
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}


data "nsxt_policy_realization_info" "bgp_realization_info" {
  path      = nsxt_policy_bgp_config.test.path
  %s
}`, attrMap["enabled"], attrMap["inter_sr_ibgp"], attrMap["local_as_num"], attrMap["multipath_relax"], attrMap["prefix"], attrMap["summary_only"], attrMap["graceful_restart_mode"], attrMap["graceful_restart_timer"], attrMap["graceful_restart_stale_route_timer"], extraConfig, extraConfig)
}

func testAccNsxtPolicyBgpConfigMinimalistic() string {
	extraConfig := ""
	if testAccIsGlobalManager() {
		extraConfig = `site_path    = data.nsxt_policy_site.test.path`
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) + testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_bgp_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  %s
  local_as_num = 65001
}

data "nsxt_policy_realization_info" "realization_info" {
  path      = nsxt_policy_bgp_config.test.path
  %s
}`, extraConfig, extraConfig)
}
