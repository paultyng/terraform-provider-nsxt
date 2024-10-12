// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
)

func TestAccResourceNsxtEdgeCluster_basic(t *testing.T) {
	clusterName := getAccTestResourceName()
	updateClusterName := "updated-" + clusterName
	testResourceName := "nsxt_edge_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestFabric(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXEdgeClusterCheckDestroy(state, updateClusterName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeClusterCreateTemplate(clusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXEdgeClusterExists(clusterName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", clusterName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test edge cluster"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXEdgeClusterCreateTemplate(updateClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXEdgeClusterExists(updateClusterName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateClusterName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test edge cluster"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtEdgeCluster_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	clusterName := getAccTestResourceName()
	testResourceName := "nsxt_edge_cluster.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestFabric(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXEdgeClusterCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeClusterCreateTemplate(clusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXEdgeClusterExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Edge Cluster resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX Edge Cluster resource ID not set in resources ")
		}

		client := nsx.NewEdgeClustersClient(connector)
		obj, err := client.Get(resourceID)
		if err != nil {
			return fmt.Errorf("error while retrieving Edge Cluster ID %s. Error: %v", resourceID, err)
		}

		if displayName == *obj.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX Edge Cluster %s wasn't found", displayName)
	}
}

func testAccNSXEdgeClusterCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

	// This addresses the fact that object is retrieved even though it had been deleted
	time.Sleep(10 * time.Second)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_edge_cluster" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		client := nsx.NewEdgeClustersClient(connector)
		obj, err := client.Get(resourceID)

		if isNotFoundError(err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while retrieving Edge Cluster ID %s. Error: %v", resourceID, err)
		}

		if obj.DisplayName != nil && displayName == *obj.DisplayName {
			return fmt.Errorf("NSX Edge Cluster %s still exists", displayName)
		}
	}

	return nil
}

func testAccNSXEdgeClusterCreateTemplate(displayName string) string {
	return fmt.Sprintf(`
resource "nsxt_edge_cluster" "test" {
    description  = "Terraform test edge cluster"
    display_name = "%s"
    tag {
    	scope = "scope1"
        tag   = "tag1"
    }
}
`, displayName)
}
