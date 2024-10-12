// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	gm_tier_0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var testAccResourcePolicyGWPrefixListName = "nsxt_policy_gateway_prefix_list.test"
var testAccDataSourcePolicyGWPrefixListName = "data.nsxt_policy_gateway_prefix_list.test"

// NOTE - to save test suite running time, this test also covers the data source
func TestAccResourceNsxtPolicyGatewayPrefixList_basic(t *testing.T) {
	name := getAccTestResourceName()
	action := model.PrefixEntry_ACTION_DENY
	actionUpdated := model.PrefixEntry_ACTION_PERMIT
	ge := "20"
	le := "23"
	network := "4.4.0.0/20"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGWPrefixListCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGWPrefixListCreateTemplate(name, action, ge, le, network),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGWPrefixListExists(testAccResourcePolicyGWPrefixListName),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "description", "test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.le", le),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.ge", ge),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.network", network),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "revision"),
					resource.TestCheckResourceAttr(testAccDataSourcePolicyGWPrefixListName, "display_name", name),
					resource.TestCheckResourceAttr(testAccDataSourcePolicyGWPrefixListName, "description", "test"),
				),
			},
			{
				Config: testAccNsxtPolicyGWPrefixListUpdateTemplate(name, actionUpdated, ge, le, network),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGWPrefixListExists(testAccResourcePolicyGWPrefixListName),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "description", "updated"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.action", actionUpdated),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.le", le),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.ge", ge),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.0.network", network),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.1.action", actionUpdated),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.1.le", "0"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.1.ge", "0"),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "prefix.1.network", ""),
					resource.TestCheckResourceAttr(testAccResourcePolicyGWPrefixListName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyGWPrefixListName, "revision"),
					resource.TestCheckResourceAttr(testAccDataSourcePolicyGWPrefixListName, "display_name", name),
					resource.TestCheckResourceAttr(testAccDataSourcePolicyGWPrefixListName, "description", "updated"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayPrefixList_import(t *testing.T) {
	name := getAccTestResourceName()
	action := model.PrefixEntry_ACTION_DENY
	ge := "0"
	le := "0"
	network := "4.4.0.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGWPrefixListCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGWPrefixListCreateTemplate(name, action, ge, le, network),
			},
			{
				ResourceName:      testAccResourcePolicyGWPrefixListName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyGetGatewayImporterIDGenerator(testAccResourcePolicyGWPrefixListName),
			},
		},
	})
}

func testAccNSXPolicyGetGatewayImporterIDGenerator(testResourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[testResourceName]
		if !ok {
			return "", fmt.Errorf("NSX Policy resource %s not found in resources", testResourceName)
		}
		resourceID := rs.Primary.ID
		if resourceID == "" {
			return "", fmt.Errorf("NSX Policy resource ID not set in resources ")
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		if gwPath == "" {
			return "", fmt.Errorf("NSX Policy Gateway Path not set in resources ")
		}
		_, gwID := parseGatewayPolicyPath(gwPath)
		return fmt.Sprintf("%s/%s", gwID, resourceID), nil
	}
}

func testAccNsxtPolicyGWPrefixListExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Gateway Prefix List resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Gateway Prefix List resource ID not set in resources")
		}

		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		var err error
		if testAccIsGlobalManager() {
			client := gm_tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, resourceID)
		} else {
			client := tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, resourceID)
		}
		if err != nil {
			return fmt.Errorf("Error while retrieving policy Gateway Prefix List ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyGWPrefixListCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_prefix_list" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		var err error
		if testAccIsGlobalManager() {
			client := gm_tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, resourceID)
		} else {
			client := tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, resourceID)
		}
		if err == nil {
			return fmt.Errorf("Policy Gateway Prefix List %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGWPrefixListCreateTemplate(name string, action string, ge string, le string, network string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"
  description  = "test"
  gateway_path = nsxt_policy_tier0_gateway.test.path

  prefix {
  	action  = "%s"
  	ge      = "%s"
  	le      = "%s"
  	network = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"

  gateway_path = nsxt_policy_tier0_gateway.test.path
  depends_on = [nsxt_policy_gateway_prefix_list.test]
}
`, name, action, ge, le, network, name)
}

func testAccNsxtPolicyGWPrefixListUpdateTemplate(name string, action string, ge string, le string, network string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"
  description  = "updated"
  gateway_path = nsxt_policy_tier0_gateway.test.path

  prefix {
  	action  = "%s"
  	ge      = "%s"
  	le      = "%s"
  	network = "%s"
  }

  prefix {
  	action  = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_policy_gateway_prefix_list" "test" {
  display_name = "%s"

  depends_on = [nsxt_policy_gateway_prefix_list.test]
}
`, name, action, ge, le, network, action, name)
}
