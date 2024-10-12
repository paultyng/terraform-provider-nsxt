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

var accTestPolicyRoleCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
}

var accTestPolicyRoleUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
}

func TestAccResourceNsxtPolicyRole_basic(t *testing.T) {
	testResourceName := "nsxt_policy_user_management_role.test_role"
	roleName := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyRoleCheckDestroy(state, accTestPolicyRoleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRoleCreate(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleExists(accTestPolicyRoleCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "role", roleName),
					resource.TestCheckResourceAttr(testResourceName, "feature.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "feature.0.feature", "policy_grouping"),
					resource.TestCheckResourceAttr(testResourceName, "feature.1.feature", "vm_vm_info"),
					resource.TestCheckResourceAttr(testResourceName, "feature.2.feature", "policy_services"),
					resource.TestCheckResourceAttr(testResourceName, "feature.0.permission", "read"),
					resource.TestCheckResourceAttr(testResourceName, "feature.1.permission", "read"),
					resource.TestCheckResourceAttr(testResourceName, "feature.2.permission", "read"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.0.feature_description"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.1.feature_description"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.2.feature_description"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.0.feature_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.1.feature_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.2.feature_name"),
				),
			},
			{
				Config: testAccNsxtPolicyRoleUpdate(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleExists(accTestPolicyRoleUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "role", roleName),
					resource.TestCheckResourceAttr(testResourceName, "feature.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "feature.0.feature", "policy_grouping"),
					resource.TestCheckResourceAttr(testResourceName, "feature.1.feature", "vm_vm_info"),
					resource.TestCheckResourceAttr(testResourceName, "feature.0.permission", "crud"),
					resource.TestCheckResourceAttr(testResourceName, "feature.1.permission", "crud"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.0.feature_description"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.1.feature_description"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.0.feature_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "feature.1.feature_name"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyRole_import_basic(t *testing.T) {
	testResourceName := "nsxt_policy_user_management_role.test_role"
	roleName := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyRoleCheckDestroy(state, accTestPolicyRoleCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRoleCreate(roleName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyRoleExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Role resource %s not found in resources", resourceName)
		}

		roleID := rs.Primary.Attributes["id"]
		if roleID == "" {
			return fmt.Errorf("Role resource ID not set in resources")
		}
		exists, err := resourceNsxtPolicyUserManagementRoleExists(roleID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Role %s does not exist", roleID)
		}

		return nil
	}
}

func testAccNsxtPolicyRoleCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_user_management_role" {
			continue
		}

		roleID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyUserManagementRoleExists(roleID, connector)
		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Role %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyRoleCreate(role string) string {
	attrMap := accTestPolicyRoleCreateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_user_management_role" "test_role" {
  display_name = "%s"
  description = "%s"
  role = "%s"
  feature {
    feature = "policy_grouping"
    permission = "read"
  }

  feature {
    feature = "vm_vm_info"
    permission = "read"
  }

  feature {
    feature = "policy_services"
    permission = "read"
  }
}`, attrMap["display_name"], attrMap["description"], role)
}

func testAccNsxtPolicyRoleUpdate(role string) string {
	attrMap := accTestPolicyRoleUpdateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_user_management_role" "test_role" {
  display_name = "%s"
  description = "%s"
  role = "%s"
  feature {
    feature = "policy_grouping"
    permission = "crud"
  }

  feature {
    feature = "vm_vm_info"
    permission = "crud"
  }
}`, attrMap["display_name"], attrMap["description"], role)
}
