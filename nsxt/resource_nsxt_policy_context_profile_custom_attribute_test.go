/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var accTestPolicyContextProfileCustomAttributeAttributes = map[string]string{
	"key":       model.PolicyCustomAttributes_KEY_DOMAIN_NAME,
	"attribute": "test.fqdn.org",
}

func TestAccResourceNsxtPolicyContextProfileCustomAttribute_basic(t *testing.T) {
	testAccResourceNsxtPolicyContextProfileCustomAttributeBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyContextProfileCustomAttribute_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyContextProfileCustomAttributeBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyContextProfileCustomAttributeBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_context_profile_custom_attribute.test"
	var clientType tf_api.ClientType = tf_api.Local
	if withContext {
		clientType = tf_api.Multitenancy
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyContextProfileCustomAttributeCheckDestroy(state, accTestPolicyContextProfileCustomAttributeAttributes["attribute"], clientType)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileCustomAttributeTemplate(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyContextProfileCustomAttributeExists(accTestPolicyContextProfileCustomAttributeAttributes["attribute"], testResourceName, clientType),
					resource.TestCheckResourceAttr(testResourceName, "key", accTestPolicyContextProfileCustomAttributeAttributes["key"]),
					resource.TestCheckResourceAttr(testResourceName, "attribute", accTestPolicyContextProfileCustomAttributeAttributes["attribute"]),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyContextProfileCustomAttribute_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_context_profile_custom_attribute.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyContextProfileCustomAttributeCheckDestroy(state, name, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileCustomAttributeTemplate(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyContextProfileCustomAttributeExists(displayName string, resourceName string, clientType tf_api.ClientType) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy ContextProfileCustomAttribute resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy ContextProfileCustomAttribute resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyContextProfileCustomAttributeExists(testAccGetContextByType(clientType), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy ContextProfileCustomAttribute %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyContextProfileCustomAttributeCheckDestroy(state *terraform.State, attribute string, clientType tf_api.ClientType) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_context_profile_custom_attribute" {
			continue
		}

		resourceID := makeCustomAttributeID(model.PolicyCustomAttributes_KEY_DOMAIN_NAME, rs.Primary.Attributes["attribute"])
		exists, err := resourceNsxtPolicyContextProfileCustomAttributeExists(testAccGetContextByType(clientType), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy ContextProfileCustomAttribute %s still exists", attribute)
		}
	}
	return nil
}

func testAccNsxtPolicyContextProfileCustomAttributeTemplate(withContext bool) string {
	return testAccNsxtPolicyContextProfileCustomAttributeArgTemplate(
		accTestPolicyContextProfileCustomAttributeAttributes["key"],
		accTestPolicyContextProfileCustomAttributeAttributes["attribute"],
		withContext)
}

func testAccNsxtPolicyContextProfileCustomAttributeArgTemplate(key string, attribute string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtProjectContext()
	}

	return fmt.Sprintf(`
resource "nsxt_policy_context_profile_custom_attribute" "test" {
%s
  key = "%s"
  attribute = "%s"
}`, context, key, attribute)
}
