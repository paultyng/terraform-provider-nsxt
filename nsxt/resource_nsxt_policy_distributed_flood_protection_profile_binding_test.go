// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

// This test file tests both distributed_flood_protection_profile and distributed_flood_protection_profile_binding
package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDistributedFloodProtectionProfileBindingCreateAttributes = map[string]string{
	"description":      "terraform created",
	"profile_res_name": "test1",
	"seq_num":          "10",
}

var accTestPolicyDistributedFloodProtectionProfileBindingUpdateAttributes = map[string]string{
	"description":      "terraform updated",
	"profile_res_name": "test2",
	"seq_num":          "12",
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfileBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfileBinding_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_distributed_flood_protection_profile_binding.test"
	if withContext {
		testResourceName = "nsxt_policy_distributed_flood_protection_profile_binding.mttest"
	}
	name := getAccTestResourceName()
	updatedName := fmt.Sprintf("%s-updated", name)

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedFloodProtectionProfileBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileBindingTemplate(true, withContext, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedFloodProtectionProfileBindingExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedFloodProtectionProfileBindingCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestPolicyDistributedFloodProtectionProfileBindingCreateAttributes["seq_num"]),
					resource.TestCheckResourceAttrSet(testResourceName, "profile_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileBindingTemplate(false, withContext, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDistributedFloodProtectionProfileBindingExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDistributedFloodProtectionProfileBindingUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", accTestPolicyDistributedFloodProtectionProfileBindingUpdateAttributes["seq_num"]),
					resource.TestCheckResourceAttrSet(testResourceName, "profile_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfileBinding_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingImportBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDistributedFloodProtectionProfileBinding_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyDistributedFloodProtectionProfileBindingImportBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_distributed_flood_protection_profile_binding.test"
	if withContext {
		testResourceName = "nsxt_policy_distributed_flood_protection_profile_binding.mttest"
	}
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDistributedFloodProtectionProfileBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDistributedFloodProtectionProfileBindingTemplate(true, withContext, name),
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

func testAccNsxtPolicyDistributedFloodProtectionProfileBindingExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DistributedFloodProtectionProfileBinding resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy DistributedFloodProtectionProfileBinding resource ID not set in resources")
		}
		groupPath := rs.Primary.Attributes["group_path"]
		if groupPath == "" {
			return fmt.Errorf("Policy DistributedFloodProtectionProfileBinding resource group_path not set in resources")
		}

		exists, err := resourceNsxtPolicyDistributedFloodProtectionProfileBindingExists(testAccGetSessionContext(), connector, groupPath, resourceID)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy DistributedFloodProtectionProfileBinding %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDistributedFloodProtectionProfileBindingCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_distributed_flood_protection_profile_binding" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		groupPath := rs.Primary.Attributes["group_path"]
		exists, err := resourceNsxtPolicyDistributedFloodProtectionProfileBindingExists(testAccGetSessionContext(), connector, groupPath, resourceID)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DistributedFloodProtectionProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDistributedFloodProtectionProfileBindingTemplate(createFlow, withContext bool, name string) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDistributedFloodProtectionProfileBindingCreateAttributes
	} else {
		attrMap = accTestPolicyDistributedFloodProtectionProfileBindingUpdateAttributes
	}
	context := ""
	resourceName := "test"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		resourceName = "mttest"
	}
	return testAccNsxtPolicyDistributedFloodProtectionProfileBindingDeps(withContext) + fmt.Sprintf(`
resource "nsxt_policy_distributed_flood_protection_profile_binding" "%s" {
%s
 display_name = "%s"
 description  = "%s"
 profile_path    = nsxt_policy_distributed_flood_protection_profile.%s.path
 group_path      = nsxt_policy_group.test.path
 sequence_number = %s

 tag {
   scope = "scope1"
   tag   = "tag1"
 }
}
`, resourceName, context, name, attrMap["description"], attrMap["profile_res_name"], attrMap["seq_num"])
}

func testAccNsxtPolicyDistributedFloodProtectionProfileBindingDeps(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
%s
  display_name = "testgroup"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
    }
  }
}

resource "nsxt_policy_distributed_flood_protection_profile" "test1" {
%s
  display_name = "dfpp1"
  description  = "Acceptance Test"
  icmp_active_flow_limit = 3
  other_active_conn_limit = 3
  tcp_half_open_conn_limit = 3
  udp_active_flow_limit = 3
  enable_rst_spoofing = false
  enable_syncache = false
}

resource "nsxt_policy_distributed_flood_protection_profile" "test2" {
%s
  display_name = "dfpp2"
  description  = "Acceptance Test"
  icmp_active_flow_limit = 4
  other_active_conn_limit = 4
  tcp_half_open_conn_limit = 4
  udp_active_flow_limit = 4
  enable_rst_spoofing = false
  enable_syncache = false
}
`, context, context, context)
}
