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

func TestAccResourceNsxtPolicyIPBlock_minimal(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_block.test"
	cidr := "192.168.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPBlockCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPBlockCreateMinimalTemplate(name, cidr, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPBlockCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPBlock_basic(t *testing.T) {
	testAccResourceNsxtPolicyIPBlockBasic(t, false, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccResourceNsxtPolicyIPBlock_visibility(t *testing.T) {
	testAccResourceNsxtPolicyIPBlockBasic(t, false, true, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "4.2.0")
	})
}

func TestAccResourceNsxtPolicyIPBlock_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIPBlockBasic(t, true, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyIPBlockBasic(t *testing.T, withContext bool, withVisibility bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_block.test"
	cidr := "192.168.1.0/24"
	cidr2 := "191.166.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPBlockCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPBlockCreateMinimalTemplate(name, cidr, withContext, withVisibility),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPBlockCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					testAccNSXPolicyIPBlockVisibility(testResourceName, withVisibility, "EXTERNAL"),
				),
			},
			{
				Config: testAccNSXPolicyIPBlockUpdateTemplate(name, cidr2, withContext, withVisibility),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPBlockCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "cidr", cidr2),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					testAccNSXPolicyIPBlockVisibility(testResourceName, withVisibility, "PRIVATE"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPBlock_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_block.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPBlockCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPBlockCreateMinimalTemplate(name, "192.191.1.0/24", false, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPBlock_importVisibility(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_block.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "4.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPBlockCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPBlockCreateMinimalTemplate(name, "192.191.1.0/24", false, true),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPBlock_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_block.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPBlockCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPBlockCreateMinimalTemplate(name, "192.191.1.0/24", true, false),
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

func testAccNSXPolicyIPBlockCheckExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		client := infra.NewIpBlocksClient(testAccGetSessionContext(), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Policy IP Block resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX Policy IP Block resource ID not set in resources ")
		}

		_, err := client.Get(resourceID, nil)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy IP Block ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNSXPolicyIPBlockVisibility(resourceName string, withVisibility bool, expected string) resource.TestCheckFunc {
	if !withVisibility {
		return func(state *terraform.State) error {
			return nil
		}
	}
	return resource.TestCheckResourceAttr(resourceName, "visibility", expected)
}

func testAccNSXPolicyIPBlockCheckDestroy(state *terraform.State) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	client := infra.NewIpBlocksClient(testAccGetSessionContext(), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_policy_ip_block" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		_, err := client.Get(resourceID, nil)
		if err == nil {
			return fmt.Errorf("Policy IP Block %s still exists", resourceID)
		}
	}
	return nil
}

func testAccNSXPolicyIPBlockCreateMinimalTemplate(displayName string, cidr string, withContext, withVisibility bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	visibility := ""
	if withVisibility {
		visibility = "  visibility   = \"EXTERNAL\""
	}

	return fmt.Sprintf(`
resource "nsxt_policy_ip_block" "test" {
%s
  display_name = "%s"
  cidr         = "%s"
%s
}`, context, displayName, cidr, visibility)
}

func testAccNSXPolicyIPBlockUpdateTemplate(displayName string, cidr string, withContext, withVisibility bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	visibility := ""
	if withVisibility {
		visibility = "  visibility   = \"PRIVATE\""
	}

	return fmt.Sprintf(`
resource "nsxt_policy_ip_block" "test" {
%s
  display_name = "%s"
  cidr         = "%s"
%s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, context, displayName, cidr, visibility)
}
