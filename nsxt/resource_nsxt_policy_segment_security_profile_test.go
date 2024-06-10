/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicySegmentSecurityProfileCreateAttributes = map[string]string{
	"display_name":                 getAccTestResourceName(),
	"description":                  "terraform created",
	"bpdu_filter_allow":            "01:80:c2:00:00:08",
	"bpdu_filter_enable":           "true",
	"dhcp_client_block_enabled":    "true",
	"dhcp_client_block_v6_enabled": "true",
	"dhcp_server_block_enabled":    "true",
	"dhcp_server_block_v6_enabled": "true",
	"non_ip_traffic_block_enabled": "true",
	"ra_guard_enabled":             "true",
	"rate_limits_enabled":          "true",
	"rx_broadcast":                 "2",
	"rx_multicast":                 "2",
	"tx_broadcast":                 "2",
	"tx_multicast":                 "2",
}

var accTestPolicySegmentSecurityProfileUpdateAttributes = map[string]string{
	"display_name":                 getAccTestResourceName(),
	"description":                  "terraform updated",
	"bpdu_filter_allow":            "01:80:c2:00:00:05",
	"bpdu_filter_enable":           "true",
	"dhcp_client_block_enabled":    "false",
	"dhcp_client_block_v6_enabled": "false",
	"dhcp_server_block_enabled":    "false",
	"dhcp_server_block_v6_enabled": "false",
	"non_ip_traffic_block_enabled": "false",
	"ra_guard_enabled":             "false",
	"rate_limits_enabled":          "false",
	"rx_broadcast":                 "5",
	"rx_multicast":                 "5",
	"tx_broadcast":                 "5",
	"tx_multicast":                 "5",
}

func TestAccResourceNsxtPolicySegmentSecurityProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicySegmentSecurityProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicySegmentSecurityProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicySegmentSecurityProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicySegmentSecurityProfile_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicySegmentSecurityProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicySegmentSecurityProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_segment_security_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentSecurityProfileCheckDestroy(state, accTestPolicySegmentSecurityProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileTemplate(true, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentSecurityProfileExists(accTestPolicySegmentSecurityProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySegmentSecurityProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySegmentSecurityProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_allow.0", accTestPolicySegmentSecurityProfileCreateAttributes["bpdu_filter_allow"]),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_enable", accTestPolicySegmentSecurityProfileCreateAttributes["bpdu_filter_enable"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_client_block_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["dhcp_client_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_client_block_v6_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["dhcp_client_block_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_block_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["dhcp_server_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_block_v6_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["dhcp_server_block_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "non_ip_traffic_block_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["non_ip_traffic_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ra_guard_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["ra_guard_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits_enabled", accTestPolicySegmentSecurityProfileCreateAttributes["rate_limits_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.rx_broadcast", accTestPolicySegmentSecurityProfileCreateAttributes["rx_broadcast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.rx_multicast", accTestPolicySegmentSecurityProfileCreateAttributes["rx_multicast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.tx_broadcast", accTestPolicySegmentSecurityProfileCreateAttributes["tx_broadcast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.tx_multicast", accTestPolicySegmentSecurityProfileCreateAttributes["tx_multicast"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentSecurityProfileTemplate(false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentSecurityProfileExists(accTestPolicySegmentSecurityProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySegmentSecurityProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySegmentSecurityProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_allow.0", accTestPolicySegmentSecurityProfileUpdateAttributes["bpdu_filter_allow"]),
					resource.TestCheckResourceAttr(testResourceName, "bpdu_filter_enable", accTestPolicySegmentSecurityProfileUpdateAttributes["bpdu_filter_enable"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_client_block_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["dhcp_client_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_client_block_v6_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["dhcp_client_block_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_block_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["dhcp_server_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_server_block_v6_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["dhcp_server_block_v6_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "non_ip_traffic_block_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["non_ip_traffic_block_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ra_guard_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["ra_guard_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rate_limits_enabled", accTestPolicySegmentSecurityProfileUpdateAttributes["rate_limits_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.rx_broadcast", accTestPolicySegmentSecurityProfileUpdateAttributes["rx_broadcast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.rx_multicast", accTestPolicySegmentSecurityProfileUpdateAttributes["rx_multicast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.tx_broadcast", accTestPolicySegmentSecurityProfileUpdateAttributes["tx_broadcast"]),
					resource.TestCheckResourceAttr(testResourceName, "rate_limit.0.tx_multicast", accTestPolicySegmentSecurityProfileUpdateAttributes["tx_multicast"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentSecurityProfileMinimalistic(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentSecurityProfileExists(accTestPolicySegmentSecurityProfileCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtPolicySegmentSecurityProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment_security_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentSecurityProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicySegmentSecurityProfile_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment_security_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentSecurityProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileMinimalistic(true),
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

func TestAccResourceNsxtPolicySegmentSecurityProfile_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment_security_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentSecurityProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentSecurityProfileMinimalistic(false),
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

func testAccNsxtPolicySegmentSecurityProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy SegmentSecurityProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy SegmentSecurityProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicySegmentSecurityProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy SegmentSecurityProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicySegmentSecurityProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_segment_security_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicySegmentSecurityProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy SegmentSecurityProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicySegmentSecurityProfileTemplate(createFlow, withContext bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicySegmentSecurityProfileCreateAttributes
	} else {
		attrMap = accTestPolicySegmentSecurityProfileUpdateAttributes
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_segment_security_profile" "test" {
%s
  display_name = "%s"
  description  = "%s"
  bpdu_filter_allow = ["%s"]
  bpdu_filter_enable = %s
  dhcp_client_block_enabled = %s
  dhcp_client_block_v6_enabled = %s
  dhcp_server_block_enabled = %s
  dhcp_server_block_v6_enabled = %s
  non_ip_traffic_block_enabled = %s
  ra_guard_enabled = %s
  rate_limits_enabled = %s

  rate_limit {
    rx_broadcast = %s
    rx_multicast = %s
    tx_broadcast = %s
    tx_multicast = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, attrMap["display_name"], attrMap["description"], attrMap["bpdu_filter_allow"], attrMap["bpdu_filter_enable"], attrMap["dhcp_client_block_enabled"], attrMap["dhcp_client_block_v6_enabled"], attrMap["dhcp_server_block_enabled"], attrMap["dhcp_server_block_v6_enabled"], attrMap["non_ip_traffic_block_enabled"], attrMap["ra_guard_enabled"], attrMap["rate_limits_enabled"], attrMap["rx_broadcast"], attrMap["rx_multicast"], attrMap["tx_broadcast"], attrMap["tx_multicast"])
}

func testAccNsxtPolicySegmentSecurityProfileMinimalistic(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_segment_security_profile" "test" {
%s
  display_name = "%s"
}`, context, accTestPolicySegmentSecurityProfileUpdateAttributes["display_name"])
}
