/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyShareCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
}

var accTestPolicyShareUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
}

func TestAccResourceNsxtPolicyShare_basic(t *testing.T) {
	testResourceName := "nsxt_policy_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyShareCheckDestroy(state, accTestPolicyShareUpdateAttributes["display_name"], tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyShareTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyShareExists(accTestPolicyShareCreateAttributes["display_name"], testResourceName, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyShareCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyShareCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyShareTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyShareExists(accTestPolicyShareUpdateAttributes["display_name"], testResourceName, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyShareUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyShareUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyShare_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyShareCheckDestroy(state, name, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyShareTemplate(true),
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

func TestAccResourceNsxtPolicyShare_multitenancy(t *testing.T) {
	testResourceName := "nsxt_policy_share.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyShareCheckDestroy(state, accTestPolicyShareUpdateAttributes["display_name"], tf_api.Multitenancy)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyShareWithMyselfTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyShareExists(accTestPolicyShareCreateAttributes["display_name"], testResourceName, tf_api.Multitenancy),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyShareCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyShareCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyShareWithMyselfTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyShareExists(accTestPolicyShareUpdateAttributes["display_name"], testResourceName, tf_api.Multitenancy),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyShareUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyShareUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func testAccNsxtPolicyShareExists(displayName string, resourceName string, clientType tf_api.ClientType) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Share resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Share resource ID not set in resources")
		}
		exists, err := resourceNsxtPolicyShareExists(testAccGetContextByType(clientType), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Share %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyShareCheckDestroy(state *terraform.State, displayName string, clientType tf_api.ClientType) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_share" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyShareExists(testAccGetContextByType(clientType), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Share %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyShareTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyShareCreateAttributes
	} else {
		attrMap = accTestPolicyShareUpdateAttributes
	}
	return testAccNsxtPolicyProjectMinimalistic() + fmt.Sprintf(`
resource "nsxt_policy_share" "test" {
  display_name = "%s"
  description  = "%s"

  shared_with = [nsxt_policy_project.test.path]
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"])
}

func testAccNsxtPolicyShareWithMyselfTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyShareCreateAttributes
	} else {
		attrMap = accTestPolicyShareUpdateAttributes
	}
	projectID := os.Getenv("NSXT_PROJECT_ID")
	return fmt.Sprintf(`
data "nsxt_policy_project" "test" {
  id = "%s"
}

resource "nsxt_policy_share" "test" {
  context {
    project_id = data.nsxt_policy_project.test.id
  }
  display_name = "%s"
  description  = "%s"

  sharing_strategy = "ALL_DESCENDANTS"
  shared_with      = [data.nsxt_policy_project.test.path]
}`, projectID, attrMap["display_name"], attrMap["description"])
}
