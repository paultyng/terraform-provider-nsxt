// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	ippools "github.com/vmware/terraform-provider-nsxt/api/infra/ip_pools"
)

// TODO: remove extra test step config once IP Blocks don't need a delay to delete
func TestAccResourceNsxtPolicyIPPoolBlockSubnet_minimal(t *testing.T) {
	poolName := getAccTestResourceName()
	name := "blocksubnet1"
	testResourceName := "nsxt_policy_ip_pool_block_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolBlockSubnetCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetCreateMinimalTemplate(poolName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolBlockSubnetCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "size", "4"),
					resource.TestCheckResourceAttr(testResourceName, "auto_assign_gateway", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "block_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
				),
			},
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(false),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPoolBlockSubnet_basic(t *testing.T) {
	testAccResourceNsxtPolicyIPPoolBlockSubnetBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccResourceNsxtPolicyIPPoolBlockSubnet_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIPPoolBlockSubnetBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyIPPoolBlockSubnetBasic(t *testing.T, withContext bool, preCheck func()) {
	poolName := getAccTestResourceName()
	name := getAccTestResourceName()
	updatedName := fmt.Sprintf("%s-updated", name)
	testResourceName := "nsxt_policy_ip_pool_block_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolBlockSubnetCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetCreateTemplate(poolName, name, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolBlockSubnetCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "size", "4"),
					resource.TestCheckResourceAttr(testResourceName, "auto_assign_gateway", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "block_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
				),
			},
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetUpdateTemplate(poolName, updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyIPPoolBlockSubnetCheckExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "size", "4"),
					resource.TestCheckResourceAttr(testResourceName, "auto_assign_gateway", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "block_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
				),
			},
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(withContext),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPoolBlockSubnet_import_basic(t *testing.T) {
	poolName := getAccTestResourceName()
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool_block_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolBlockSubnetCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetCreateTemplate(poolName, name, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyIPPoolBlockSubnetImporterGetID,
			},
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(false),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPPoolBlockSubnet_importBasic_multitenancy(t *testing.T) {
	poolName := getAccTestResourceName()
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ip_pool_block_subnet.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyIPPoolBlockSubnetCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetCreateTemplate(poolName, name, true),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testResourceName),
			},
			{
				Config: testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(true),
			},
		},
	})
}

func testAccNSXPolicyIPPoolBlockSubnetImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["nsxt_policy_ip_pool_block_subnet.test"]
	if !ok {
		return "", fmt.Errorf("NSX Policy Block Subnet resource %s not found in resources", "nsxt_policy_ip_pool_block_subnet.test")
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Block Subnet resource ID not set in resources ")
	}
	poolPath := rs.Primary.Attributes["pool_path"]
	if poolPath == "" {
		return "", fmt.Errorf("NSX Policy Block Subnet pool_path not set in resources ")
	}
	poolID := getPolicyIDFromPath(poolPath)
	return fmt.Sprintf("%s/%s", poolID, resourceID), nil
}

func testAccNSXPolicyIPPoolBlockSubnetCheckExists(resourceName string) resource.TestCheckFunc {

	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		client := ippools.NewIpSubnetsClient(testAccGetSessionContext(), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Policy Block Subnet resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		poolID := getPolicyIDFromPath(rs.Primary.Attributes["pool_path"])
		if resourceID == "" {
			return fmt.Errorf("NSX Policy Block Subnet resource ID not set in resources")
		}

		_, err := client.Get(poolID, resourceID)
		if err != nil {
			return fmt.Errorf("Failed to find Block Subnet %s", resourceID)
		}

		return nil
	}
}

func testAccNSXPolicyIPPoolBlockSubnetCheckDestroy(state *terraform.State) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	client := ippools.NewIpSubnetsClient(testAccGetSessionContext(), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_policy_ip_pool_block_subnet" {
			continue
		}

		resourceID := rs.Primary.ID
		poolID := getPolicyIDFromPath(rs.Primary.Attributes["pool_path"])
		_, err := client.Get(poolID, resourceID)
		if err == nil {
			return fmt.Errorf("Block Subnet still exists %s", resourceID)
		}

	}
	return nil
}

func testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	return fmt.Sprintf(`
resource "nsxt_policy_ip_block" "block1" {
%s
  display_name = "tfblock2"
  cidr         = "11.11.12.0/24"
}`, context)
}

func testAccNSXPolicyIPPoolBlockSubnetCreateMinimalTemplate(poolName string, name string) string {
	return testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(false) + fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "pool1" {
  display_name = "%s"
}

resource "nsxt_policy_ip_pool_block_subnet" "test" {
  display_name = "%s"
  size         = 4
  pool_path    = nsxt_policy_ip_pool.pool1.path
  block_path   = nsxt_policy_ip_block.block1.path
}`, poolName, name)
}

func testAccNSXPolicyIPPoolBlockSubnetCreateTemplate(poolName string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	return testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(withContext) + fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "pool1" {
%s
  display_name = "%s"
}

resource "nsxt_policy_ip_pool_block_subnet" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  size                = 4
  auto_assign_gateway = false
  pool_path           = nsxt_policy_ip_pool.pool1.path
  block_path          = nsxt_policy_ip_block.block1.path
  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, context, poolName, context, name)
}

func testAccNSXPolicyIPPoolBlockSubnetUpdateTemplate(poolName string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNSXPolicyIPPoolBlockSubnetIPBlockTemplate(withContext) + fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "pool1" {
%s
  display_name = "%s"
}

resource "nsxt_policy_ip_pool_block_subnet" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  size                = 4
  auto_assign_gateway = true
  pool_path           = nsxt_policy_ip_pool.pool1.path
  block_path          = nsxt_policy_ip_block.block1.path
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, context, poolName, context, name)
}
