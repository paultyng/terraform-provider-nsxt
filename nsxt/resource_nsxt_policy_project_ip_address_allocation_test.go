// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestProjectIpAddressAllocationCreateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform created",
	"allocation_size": "1",
}

var accTestProjectIpAddressAllocationUpdateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform updated",
	"allocation_size": "1",
}

func TestAccResourceNsxtPolicyProjectIpAddressAllocation_basic(t *testing.T) {
	testResourceName := "nsxt_policy_project_ip_address_allocation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectIpAddressAllocationCheckDestroy(state, accTestProjectIpAddressAllocationUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectIpAddressAllocationTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectIpAddressAllocationExists(accTestProjectIpAddressAllocationCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestProjectIpAddressAllocationCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestProjectIpAddressAllocationCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ips"),
					resource.TestCheckResourceAttr(testResourceName, "allocation_size", accTestProjectIpAddressAllocationCreateAttributes["allocation_size"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyProjectIpAddressAllocationTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectIpAddressAllocationExists(accTestProjectIpAddressAllocationUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestProjectIpAddressAllocationUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestProjectIpAddressAllocationUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "allocation_ips"),
					resource.TestCheckResourceAttr(testResourceName, "allocation_size", accTestProjectIpAddressAllocationUpdateAttributes["allocation_size"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyProjectIpAddressAllocationMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectIpAddressAllocationExists(accTestProjectIpAddressAllocationCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyProjectIpAddressAllocation_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_project_ip_address_allocation.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectIpAddressAllocationCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectIpAddressAllocationMinimalistic(),
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

func testAccNsxtPolicyProjectIpAddressAllocationExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy ProjectIpAddressAllocation resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy ProjectIpAddressAllocation resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyProjectIpAddressAllocationExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy ProjectIpAddressAllocation %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyProjectIpAddressAllocationCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_project_ip_address_allocation" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyProjectIpAddressAllocationExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy ProjectIpAddressAllocation %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyProjectIpAddressAllocationTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestProjectIpAddressAllocationCreateAttributes
	} else {
		attrMap = accTestProjectIpAddressAllocationUpdateAttributes
	}
	return fmt.Sprintf(`
data "nsxt_policy_project" "test" {
  id = "%s"
}

resource "nsxt_policy_project_ip_address_allocation" "test" {
  %s
  display_name    = "%s"
  description     = "%s"
  allocation_size = %s
  ip_block        = data.nsxt_policy_project.test.external_ipv4_blocks[0]
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_project_ip_address_allocation" "test" {
  %s
  allocation_ips = nsxt_policy_project_ip_address_allocation.test.allocation_ips
}`, os.Getenv("NSXT_VPC_PROJECT_ID"), testAccNsxtProjectContext(), attrMap["display_name"], attrMap["description"], attrMap["allocation_size"], testAccNsxtProjectContext())
}

func testAccNsxtPolicyProjectIpAddressAllocationMinimalistic() string {
	return fmt.Sprintf(`
data "nsxt_policy_project" "test" {
  id = "%s"
}

resource "nsxt_policy_project_ip_address_allocation" "test" {
  %s
  display_name    = "%s"
  allocation_size = %s
  ip_block        = data.nsxt_policy_project.test.external_ipv4_blocks[0]
}

data "nsxt_policy_project_ip_address_allocation" "test" {
  %s
  allocation_ips = nsxt_policy_project_ip_address_allocation.test.allocation_ips
}`, os.Getenv("NSXT_VPC_PROJECT_ID"), testAccNsxtProjectContext(), accTestProjectIpAddressAllocationUpdateAttributes["display_name"], accTestProjectIpAddressAllocationUpdateAttributes["allocation_size"], testAccNsxtProjectContext())
}
