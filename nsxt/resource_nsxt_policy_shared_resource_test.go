/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/terraform-provider-nsxt/api/infra/shares"
)

var accTestPolicySharedResourceCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
}

var accTestPolicySharedResourceUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
}

func TestAccResourceNsxtPolicySharedResource_basic(t *testing.T) {
	testResourceName := "nsxt_policy_shared_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNotGlobalManager(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySharedResourceCheckDestroy(state, accTestPolicySharedResourceUpdateAttributes["display_name"], tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySharedResourceTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySharedResourceExists(accTestPolicySharedResourceCreateAttributes["display_name"], testResourceName, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySharedResourceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySharedResourceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySharedResourceTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySharedResourceExists(accTestPolicySharedResourceUpdateAttributes["display_name"], testResourceName, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySharedResourceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySharedResourceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySharedResource_multitenancy(t *testing.T) {
	testResourceName := "nsxt_policy_shared_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySharedResourceCheckDestroy(state, accTestPolicySharedResourceUpdateAttributes["display_name"], tf_api.Multitenancy)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySharedResourceMultitenancyTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySharedResourceExists(accTestPolicySharedResourceCreateAttributes["display_name"], testResourceName, tf_api.Multitenancy),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySharedResourceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySharedResourceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicySharedResourceMultitenancyTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySharedResourceExists(accTestPolicySharedResourceUpdateAttributes["display_name"], testResourceName, tf_api.Multitenancy),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicySharedResourceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicySharedResourceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySharedResource_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_shared_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNotGlobalManager(t)
			testAccNSXVersion(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySharedResourceCheckDestroy(state, name, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySharedResourceTemplate(true),
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

func testAccNsxtPolicySharedResourceExists(displayName string, resourceName string, clientType tf_api.ClientType) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Shared Resource resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Shared Resource resource ID not set in resources")
		}
		shareID := getPolicyIDFromPath(rs.Primary.Attributes["share_path"])
		context := testAccGetContextByType(clientType)
		client := shares.NewResourcesClient(context, connector)
		_, err := client.Get(shareID, resourceID)

		if isNotFoundError(err) {
			return fmt.Errorf("Policy Shared Resource %s does not exist", resourceID)
		}
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccNsxtPolicySharedResourceCheckDestroy(state *terraform.State, displayName string, clientType tf_api.ClientType) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_shared_resource" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		shareID := getPolicyIDFromPath(rs.Primary.Attributes["share_path"])
		context := testAccGetContextByType(clientType)
		client := shares.NewResourcesClient(context, connector)
		_, err := client.Get(shareID, resourceID)

		if !isNotFoundError(err) {
			return err
		}

		if err == nil {
			return fmt.Errorf("Policy Shared Resource %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicySharedResourceTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicySharedResourceCreateAttributes
	} else {
		attrMap = accTestPolicySharedResourceUpdateAttributes
	}
	return testAccNsxtPolicyShareTemplate(createFlow) + testAccNsxtPolicyContextProfileReadTemplate("AMQP") + fmt.Sprintf(`
resource "nsxt_policy_shared_resource" "test" {
  display_name = "%s"
  description  = "%s"
  
  share_path   = nsxt_policy_share.test.path
  resource_object {
    resource_path = data.nsxt_policy_context_profile.test.path
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"])
}

func testAccNsxtPolicySharedResourceMultitenancyTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicySharedResourceCreateAttributes
	} else {
		attrMap = accTestPolicySharedResourceUpdateAttributes
	}
	return testAccNsxtPolicyShareWithMyselfTemplate(true) + testAccNsxtPolicyContextProfileReadTemplate("AMQP") + fmt.Sprintf(`
resource "nsxt_policy_shared_resource" "test" {
  display_name = "%s"
  description  = "%s"

  share_path   = nsxt_policy_share.test.path
  resource_object {
    resource_path = data.nsxt_policy_context_profile.test.path
  }
}`, attrMap["display_name"], attrMap["description"])
}
