// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/aaa"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var accTestPolicyRoleBindingCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
}

var accTestPolicyRoleBindingUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
}

func TestAccResourceNsxtPolicyRoleBinding_basic(t *testing.T) {
	testResourceName := "nsxt_policy_user_management_role_binding.test"
	identType := nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_LDAP
	userType := nsxModel.RoleBinding_TYPE_REMOTE_USER

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_LDAP_USER")
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyRoleBindingCheckDestroy(state, accTestPolicyRoleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRoleBindingCreate(getTestLdapUser(), userType, identType, "false", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleBindingExists(accTestPolicyRoleBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleBindingCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleBindingCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "name", getTestLdapUser()),
					resource.TestCheckResourceAttr(testResourceName, "type", userType),
					resource.TestCheckResourceAttr(testResourceName, "identity_source_type", identType),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.path", "/"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.0", "network_engineer"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.path", "/orgs/default"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.roles.0", "org_admin"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyRoleBindingUpdate(getTestLdapUser(), userType, identType, "false", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleBindingExists(accTestPolicyRoleBindingUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleBindingUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleBindingUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "name", getTestLdapUser()),
					resource.TestCheckResourceAttr(testResourceName, "type", userType),
					resource.TestCheckResourceAttr(testResourceName, "identity_source_type", identType),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.path", "/"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.0", "security_engineer"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyRoleBinding_local_user(t *testing.T) {
	testResourceName := "nsxt_policy_user_management_role_binding.test"
	userType := nsxModel.RoleBinding_TYPE_LOCAL_USER
	testUsername := "terraform-" + getAccTestRandomString(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyRoleBindingCheckDestroy(state, accTestPolicyRoleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRoleBindingLocalOverwrite(testUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleBindingExists(accTestPolicyRoleBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleBindingCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleBindingCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "name", testUsername),
					resource.TestCheckResourceAttr(testResourceName, "type", userType),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.path", "/"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.0", "network_engineer"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.path", "/orgs/default"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.1.roles.0", "org_admin"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyRoleBindingLocalUpdate(testUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyRoleBindingExists(accTestPolicyRoleBindingUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyRoleBindingUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyRoleBindingUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "name", testUsername),
					resource.TestCheckResourceAttr(testResourceName, "type", userType),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.path", "/"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.0", "security_engineer"),

					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyRoleBinding_import_basic(t *testing.T) {
	testResourceName := "nsxt_policy_user_management_role_binding.test"
	identType := nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_LDAP
	userType := nsxModel.RoleBinding_TYPE_REMOTE_USER

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_LDAP_USER")
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyRoleBindingCheckDestroy(state, accTestPolicyRoleUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRoleBindingCreate(getTestLdapUser(), userType, identType, "false", ""),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyRoleBindingExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("RoleBinding resource %s not found in resources", resourceName)
		}

		rbID := rs.Primary.Attributes["id"]
		if rbID == "" {
			return fmt.Errorf("RoleBinding resource ID not set in resources")
		}
		rbClient := aaa.NewRoleBindingsClient(connector)
		_, err := rbClient.Get(rbID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			if isNotFoundError(err) {
				return fmt.Errorf("RoleBinding %s does not exist", rbID)
			}
		}

		return err
	}
}

func testAccNsxtPolicyRoleBindingCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_user_management_role_binding" {
			continue
		}

		rbID := rs.Primary.Attributes["id"]
		if rbID == "" {
			return fmt.Errorf("RoleBinding resource ID not set in resources")
		}
		rbClient := aaa.NewRoleBindingsClient(connector)
		_, err := rbClient.Get(rbID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("RoleBinding %s still exists", displayName)
	}
	return nil
}

func testAccNsxtPolicyRoleBindingLocalOverwrite(user string) string {
	return fmt.Sprintf("%s\n%s", testAccNodeUserCreate(user),
		testAccNsxtPolicyRoleBindingCreate(
			user,
			nsxModel.RoleBinding_TYPE_LOCAL_USER,
			"",
			"true",
			"nsxt_node_user.test"))
}

func testAccNsxtPolicyRoleBindingLocalUpdate(user string) string {
	return fmt.Sprintf("%s\n%s", testAccNodeUserCreate(user),
		testAccNsxtPolicyRoleBindingUpdate(
			user,
			nsxModel.RoleBinding_TYPE_LOCAL_USER,
			"",
			"true",
			"nsxt_node_user.test"))
}

func testAccNsxtPolicyRoleBindingCreate(user, userType, identType, overwrite, dependsOn string) string {
	attrMap := accTestPolicyRoleBindingCreateAttributes
	var identLine, dependsOnLine string
	if len(identType) > 0 {
		identLine = fmt.Sprintf("identity_source_type = \"%s\"", identType)
	}
	if len(dependsOn) > 0 {
		dependsOnLine = fmt.Sprintf("depends_on = [%s]", dependsOn)
	}

	return fmt.Sprintf(`
resource "nsxt_policy_user_management_role_binding" "test" {
    display_name         = "%s"
    description          = "%s"
    name                 = "%s"
    type                 = "%s"
    overwrite_local_user = %s
    %s
    %s

    roles_for_path {
        path  = "/"
        roles = ["network_engineer"]
    }

    roles_for_path {
        path  = "/orgs/default"
        roles = ["org_admin"]
    }
}`, attrMap["display_name"], attrMap["description"], user, userType, overwrite, identLine, dependsOnLine)
}

func testAccNsxtPolicyRoleBindingUpdate(user, userType, identType, overwrite, dependsOn string) string {
	attrMap := accTestPolicyRoleBindingUpdateAttributes
	var identLine, dependsOnLine string
	if len(identType) > 0 {
		identLine = fmt.Sprintf("identity_source_type = \"%s\"", identType)
	}
	if len(dependsOn) > 0 {
		dependsOnLine = fmt.Sprintf("depends_on = [%s]", dependsOn)
	}
	return fmt.Sprintf(`
resource "nsxt_policy_user_management_role_binding" "test" {
    display_name         = "%s"
    description          = "%s"
    name                 = "%s"
    type                 = "%s"
    overwrite_local_user = %s
    %s
    %s

    roles_for_path {
        path  = "/"
        roles = ["security_engineer"]
    }
}`, attrMap["display_name"], attrMap["description"], user, userType, overwrite, identLine, dependsOnLine)
}
