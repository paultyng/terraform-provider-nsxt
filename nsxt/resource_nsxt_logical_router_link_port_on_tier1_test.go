// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtLogicalRouterLinkPortOnTier1_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	tier0RouterName := getTier0RouterName()
	edgeClusterName := getEdgeClusterName()
	testResourceName := "nsxt_logical_router_link_port_on_tier1.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterLinkPortOnTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterLinkPortOnTier1CreateTemplate(name, tier0RouterName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterLinkPortOnTier1Exists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_router_port_id"),
				),
			},
			{
				Config: testAccNSXLogicalRouterLinkPortOnTier1UpdateTemplate(updateName, tier0RouterName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalRouterLinkPortOnTier1Exists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "linked_logical_router_port_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalRouterLinkPortOnTier1_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	tier0RouterName := getTier0RouterName()
	edgeClusterName := getEdgeClusterName()
	testResourceName := "nsxt_logical_router_link_port_on_tier1.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalRouterLinkPortOnTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalRouterLinkPortOnTier1CreateTemplate(name, tier0RouterName, edgeClusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalRouterLinkPortOnTier1Exists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical router link port on Tier1 resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical router link port on Tier1 resource ID not set in resources")
		}

		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier1(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical router link port on Tier1 ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking verifying logical router link port on Tier1 existence. HTTP returned %d", responseCode.StatusCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical router link port on Tier1 %s not found", displayName)
	}
}

func testAccNSXLogicalRouterLinkPortOnTier1CheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_router_link_port_on_tier1" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier1(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical router link port on Tier1 %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX logical router link port on Tier1 %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalRouterLinkPortOnTier1PreconditionsTemplate(tier0RouterName string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_logical_tier0_router" "tier0rtr" {
  display_name = "%s"
}

resource "nsxt_logical_router_link_port_on_tier0" "test" {
  display_name      = "port_on_tier0"
  logical_router_id = "${data.nsxt_logical_tier0_router.tier0rtr.id}"
}

data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_logical_tier1_router" "test" {
  display_name    = "tier1_router"
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}`, tier0RouterName, edgeClusterName)
}

func testAccNSXLogicalRouterLinkPortOnTier1CreateTemplate(name string, tier0RouterName string, edgeClusterName string) string {
	return testAccNSXLogicalRouterLinkPortOnTier1PreconditionsTemplate(tier0RouterName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_link_port_on_tier1" "test" {
  display_name                  = "%s"
  description                   = "Acceptance Test"
  logical_router_id             = "${nsxt_logical_tier1_router.test.id}"
  linked_logical_router_port_id = "${nsxt_logical_router_link_port_on_tier0.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXLogicalRouterLinkPortOnTier1UpdateTemplate(name string, tier0RouterName string, edgeClusterName string) string {
	return testAccNSXLogicalRouterLinkPortOnTier1PreconditionsTemplate(tier0RouterName, edgeClusterName) + fmt.Sprintf(`
resource "nsxt_logical_router_link_port_on_tier1" "test" {
  display_name                 = "%s"
  description                  = "Acceptance Test Update"
  logical_router_id            = "${nsxt_logical_tier1_router.test.id}"
  linked_logical_router_port_id = "${nsxt_logical_router_link_port_on_tier0.test.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, name)
}
