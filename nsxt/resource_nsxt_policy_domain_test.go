// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
)

func TestAccResourceNsxtPolicyDomain_basic(t *testing.T) {
	name := "test-nsx-policy-domain-basic"
	updateName := fmt.Sprintf("%s-update", name)
	testResourceName := "nsxt_policy_domain.test"
	locationName1 := getTestSiteName()
	locationName2 := getTestAnotherSiteName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccEnvDefined(t, "NSXT_TEST_ANOTHER_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDomainCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDomainOneLocationTemplate(name, locationName1),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDomainExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sites.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sites.0", locationName1),
				),
			},
			{
				// Add another location
				Config: testAccNsxtPolicyDomainTwoLocationTemplate(updateName, locationName1, locationName2),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDomainExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "sites.#", "2"),
				),
			},
			{
				// Remove 1 location
				Config: testAccNsxtPolicyDomainOneLocationTemplate(updateName, locationName2),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDomainExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sites.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sites.0", locationName2),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDomain_importBasic(t *testing.T) {
	name := "test-nsx-policy-domain-import"
	testResourceName := "nsxt_policy_domain.test"
	locationName := getTestSiteName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDomainCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDomainOneLocationTemplate(name, locationName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyDomainExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy domain resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy domain resource ID not set in resources")
		}

		nsxClient := gm_infra.NewDomainsClient(connector)
		_, err := nsxClient.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy domain ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyDomainCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_domains" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		nsxClient := gm_infra.NewDomainsClient(connector)
		_, err := nsxClient.Get(resourceID)
		if err == nil {
			return fmt.Errorf("Policy domain %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDomainOneLocationTemplate(name string, locationName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_domain" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  sites        = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, locationName)
}

func testAccNsxtPolicyDomainTwoLocationTemplate(name string, locationName1 string, locationName2 string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_domain" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  sites        = ["%s", "%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, name, locationName1, locationName2)
}
