// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/node"
)

var accTestNodeUserCreateAttributes = map[string]string{
	"full_name":                 getAccTestRandomString(10),
	"password":                  "Q5&WfLqv9Zd5",
	"active":                    "true",
	"password_change_frequency": "180",
	"password_change_warning":   "30",
}

var accTestNodeUserUpdateAttributes = map[string]string{
	"full_name":                 getAccTestRandomString(10),
	"password":                  "Q5&WfLqv9Zd5",
	"active":                    "false",
	"password_change_frequency": "75",
	"password_change_warning":   "20",
}

func TestAccResourceNsxtNodeUser_basic(t *testing.T) {
	testResourceName := "nsxt_node_user.test"
	testUsername := getAccTestRandomString(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNodeUserCheckDestroy(state, testUsername)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNodeUserCreate(testUsername),
				Check: resource.ComposeTestCheckFunc(
					testAccNodeUserExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "full_name", accTestNodeUserCreateAttributes["full_name"]),
					resource.TestCheckResourceAttr(testResourceName, "active", accTestNodeUserCreateAttributes["active"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_frequency", accTestNodeUserCreateAttributes["password_change_frequency"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_warning", accTestNodeUserCreateAttributes["password_change_warning"]),
					resource.TestCheckResourceAttr(testResourceName, "username", testUsername),
					resource.TestCheckResourceAttr(testResourceName, "status", nsxModel.NodeUserProperties_STATUS_ACTIVE),
					resource.TestCheckResourceAttrSet(testResourceName, "password_reset_required"),
					resource.TestCheckResourceAttrSet(testResourceName, "last_password_change"),
					resource.TestCheckResourceAttrSet(testResourceName, "user_id"),
				),
			},
			{
				Config: testAccNodeUserUpdate(testUsername, accTestNodeUserCreateAttributes["password"]),
				Check: resource.ComposeTestCheckFunc(
					testAccNodeUserExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "full_name", accTestNodeUserUpdateAttributes["full_name"]),
					resource.TestCheckResourceAttr(testResourceName, "active", accTestNodeUserUpdateAttributes["active"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_frequency", accTestNodeUserUpdateAttributes["password_change_frequency"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_warning", accTestNodeUserUpdateAttributes["password_change_warning"]),
					resource.TestCheckResourceAttr(testResourceName, "username", testUsername),
					resource.TestCheckResourceAttr(testResourceName, "status", nsxModel.NodeUserProperties_STATUS_NOT_ACTIVATED),
					resource.TestCheckResourceAttrSet(testResourceName, "password_reset_required"),
					resource.TestCheckResourceAttrSet(testResourceName, "last_password_change"),
					resource.TestCheckResourceAttrSet(testResourceName, "user_id"),
				),
			},
			{
				// Test password change
				Config: testAccNodeUserUpdate(testUsername, accTestNodeUserUpdateAttributes["password"]),
				Check: resource.ComposeTestCheckFunc(
					testAccNodeUserExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "full_name", accTestNodeUserUpdateAttributes["full_name"]),
					resource.TestCheckResourceAttr(testResourceName, "active", accTestNodeUserUpdateAttributes["active"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_frequency", accTestNodeUserUpdateAttributes["password_change_frequency"]),
					resource.TestCheckResourceAttr(testResourceName, "password_change_warning", accTestNodeUserUpdateAttributes["password_change_warning"]),
					resource.TestCheckResourceAttr(testResourceName, "username", testUsername),
					resource.TestCheckResourceAttr(testResourceName, "status", nsxModel.NodeUserProperties_STATUS_NOT_ACTIVATED),
					resource.TestCheckResourceAttrSet(testResourceName, "password_reset_required"),
					resource.TestCheckResourceAttrSet(testResourceName, "last_password_change"),
					resource.TestCheckResourceAttrSet(testResourceName, "user_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNodeUser_import_basic(t *testing.T) {
	testResourceName := "nsxt_node_user.test"
	testUsername := getAccTestRandomString(10)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNodeUserCheckDestroy(state, testUsername)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNodeUserCreate(testUsername),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       testAccNodeUserImporterGetID,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testAccNodeUserExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("User resource %s not found in resources", resourceName)
		}

		userID := rs.Primary.Attributes["id"]
		if userID == "" {
			return fmt.Errorf("User ID not set in resources")
		}
		client := node.NewUsersClient(connector)
		_, err := client.Get(userID)
		if err != nil {
			if isNotFoundError(err) {
				return fmt.Errorf("User %s does not exist", userID)
			}
			return err
		}

		return nil
	}
}

func testAccNodeUserCheckDestroy(state *terraform.State, username string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_node_user" {
			continue
		}

		userID := rs.Primary.Attributes["id"]
		client := node.NewUsersClient(connector)
		_, err := client.Get(userID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
		}

		return fmt.Errorf("user %s still exists", username)
	}
	return nil
}

func testAccNodeUserImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["nsxt_node_user.test"]
	if !ok {
		return "", fmt.Errorf("User %s not found in resources", "nsxt_node_user.test")
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("User ID not set in resources")
	}
	return resourceID, nil
}

func testAccNodeUserCreate(username string) string {
	attrMap := accTestNodeUserCreateAttributes
	return fmt.Sprintf(`
resource "nsxt_node_user" "test" {
  active = %s
  full_name = "%s"
  password = "%s"
  username = "%s"
  password_change_frequency = %s
  password_change_warning = %s
}`, attrMap["active"], attrMap["full_name"], attrMap["password"], username, attrMap["password_change_frequency"], attrMap["password_change_warning"])
}

func testAccNodeUserUpdate(username, password string) string {
	attrMap := accTestNodeUserUpdateAttributes
	return fmt.Sprintf(`
resource "nsxt_node_user" "test" {
  active = %s
  full_name = "%s"
  password = "%s"
  username = "%s"
  password_change_frequency = %s
  password_change_warning = %s
}`, attrMap["active"], attrMap["full_name"], password, username, attrMap["password_change_frequency"], attrMap["password_change_warning"])
}
