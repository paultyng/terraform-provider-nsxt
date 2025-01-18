// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func TestAccResourceNsxtPolicyIPPool_minimal(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolCreateMinimalTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPool_basic(t *testing.T) {
	testAccResourceNsxtPolicyIPPoolBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccResourceNsxtPolicyIPPool_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIPPoolBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyIPPoolBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolCreateTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
				),
			},
			{
				Config: testAccNSXPolicyIPPoolUpdateTemplate(updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPool_import_basic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolCreateTemplate(name, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPool_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolCreateTemplate(name, true),
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

func testAccNSXPolicyIPPoolCheckExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		client := infra.NewIpPoolsClient(testAccGetSessionContext(), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Policy IP Pool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX Policy IP Pool resource ID not set in resources")
		}

		_, err := client.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Failed to find IP Pool %s", resourceID)
		}

		return nil
	}
}

func testAccNSXPolicyIPPoolCheckDestroy(state *terraform.State) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	client := infra.NewIpPoolsClient(testAccGetSessionContext(), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_policy_ip_pool" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		_, err := client.Get(resourceID)
		if err == nil {
			return fmt.Errorf("IP Pool still exists %s", resourceID)
		}

	}
	return nil
}

func testAccNSXPolicyIPPoolCreateMinimalTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "test" {
  display_name = "%s"
}`, name)
}

func testAccNSXPolicyIPPoolCreateTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, name)
}

func testAccNSXPolicyIPPoolUpdateTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, context, name)
}
