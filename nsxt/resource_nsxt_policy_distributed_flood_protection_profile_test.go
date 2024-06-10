/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDistributedFloodProtectionProfileCreateAttributes = map[string]string{
	"description":              "terraform created",
	"icmp_active_flow_limit":   "2",
	"other_active_conn_limit":  "2",
	"tcp_half_open_conn_limit": "2",
	"udp_active_flow_limit":    "2",
	"enable_rst_spoofing":      "false",
	"enable_syncache":          "false",
}

var accTestPolicyDistributedFloodProtectionProfileUpdateAttributes = map[string]string{
	"description":              "terraform updated",
	"icmp_active_flow_limit":   "5",
	"other_active_conn_limit":  "5",
	"tcp_half_open_conn_limit": "5",
	"udp_active_flow_limit":    "5",
	"enable_rst_spoofing":      "true",
	"enable_syncache":          "true",
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_minimal(t *testing.T) {
	testResourceName := "nsxt_policy_distributed_flood_protection_profile.test"
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileMinimalistic(false, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedFloodProtectionProfileExists(testResourceName),
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

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyDistributedFloodProtectionProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_distributed_flood_protection_profile.test"
	if withContext {
		testResourceName = "nsxt_policy_distributed_flood_protection_profile.mttest"
	}
	name := getAccTestResourceName()
	updatedName := fmt.Sprintf("%s-updated", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileTemplate(true, withContext, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedFloodProtectionProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "icmp_active_flow_limit", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["icmp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "other_active_conn_limit", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["other_active_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_half_open_conn_limit", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["tcp_half_open_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "udp_active_flow_limit", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["udp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_rst_spoofing", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["enable_rst_spoofing"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_syncache", accTestPolicyDistributedFloodProtectionProfileCreateAttributes["enable_syncache"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileTemplate(false, withContext, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedFloodProtectionProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "icmp_active_flow_limit", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["icmp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "other_active_conn_limit", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["other_active_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_half_open_conn_limit", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["tcp_half_open_conn_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "udp_active_flow_limit", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["udp_active_flow_limit"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_rst_spoofing", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["enable_rst_spoofing"]),
					resource.TestCheckResourceAttr(testResourceName, "enable_syncache", accTestPolicyDistributedFloodProtectionProfileUpdateAttributes["enable_syncache"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileImportBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfile_importBasic_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileImportBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyDistributedFloodProtectionProfileImportBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_distributed_flood_protection_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedFloodProtectionProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileMinimalistic(withContext, name),
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

func testAccNsxtPolicyDistributedFloodProtectionProfileExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DistributedFloodProtectionProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy DistributedFloodProtectionProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyFloodProtectionProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy DistributedFloodProtectionProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDistributedFloodProtectionProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_distributed_flood_protection_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyFloodProtectionProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DistributedFloodProtectionProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDistributedFloodProtectionProfileTemplate(createFlow, withContext bool, name string) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDistributedFloodProtectionProfileCreateAttributes
	} else {
		attrMap = accTestPolicyDistributedFloodProtectionProfileUpdateAttributes
	}
	context := ""
	resourceName := "test"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		resourceName = "mttest"
	}
	return fmt.Sprintf(`
resource "nsxt_policy_distributed_flood_protection_profile" "%s" {
%s
  display_name = "%s"
  description  = "%s"
  icmp_active_flow_limit = %s
  other_active_conn_limit = %s
  tcp_half_open_conn_limit = %s
  udp_active_flow_limit = %s
  enable_rst_spoofing = %s
  enable_syncache = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_distributed_flood_protection_profile" "%s" {
%s
  display_name = "%s"
  depends_on = [nsxt_policy_distributed_flood_protection_profile.%s]
}`, resourceName, context, name, attrMap["description"], attrMap["icmp_active_flow_limit"], attrMap["other_active_conn_limit"], attrMap["tcp_half_open_conn_limit"], attrMap["udp_active_flow_limit"], attrMap["enable_rst_spoofing"], attrMap["enable_syncache"], resourceName, context, name, resourceName)
}

func testAccNsxtPolicyDistributedFloodProtectionProfileMinimalistic(withContext bool, name string) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_distributed_flood_protection_profile" "test" {
%s
  display_name = "%s"

}`, context, name)
}
