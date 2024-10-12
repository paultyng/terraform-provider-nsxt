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

var accTestPolicyVniPoolCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"start":        "75001",
	"end":          "75100",
}

var accTestPolicyVniPoolUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"start":        "75101",
	"end":          "75200",
}

func TestAccResourceNsxtPolicyVniPool_basic(t *testing.T) {
	testResourceName := "nsxt_policy_vni_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyVniPoolCheckDestroy(state, accTestPolicyVniPoolUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVniPoolCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVniPoolExists(accTestPolicyVniPoolCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyVniPoolCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyVniPoolCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "start", accTestPolicyVniPoolCreateAttributes["start"]),
					resource.TestCheckResourceAttr(testResourceName, "end", accTestPolicyVniPoolCreateAttributes["end"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyVniPoolUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVniPoolExists(accTestPolicyVniPoolUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyVniPoolUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyVniPoolUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "start", accTestPolicyVniPoolUpdateAttributes["start"]),
					resource.TestCheckResourceAttr(testResourceName, "end", accTestPolicyVniPoolUpdateAttributes["end"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyVniPool_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_vni_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyVniPoolCheckDestroy(state, accTestPolicyVniPoolUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVniPoolUpdate(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyVniPoolExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy VNI Pool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy VNI Pool resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyVniPoolExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy VNI Pool %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyVniPoolCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_vni_pool" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyVniPoolExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Evpn Tenant %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyVniPoolCreate() string {
	attrMap := accTestPolicyVniPoolCreateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_vni_pool" "test" {
  display_name = "%s"
  description  = "%s"

  start = %s
  end = %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["start"], attrMap["end"])
}

func testAccNsxtPolicyVniPoolUpdate() string {
	attrMap := accTestPolicyVniPoolUpdateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_vni_pool" "test" {
  display_name = "%s"
  description  = "%s"

  start = %s
  end = %s
}`, attrMap["display_name"], attrMap["description"], attrMap["start"], attrMap["end"])
}
