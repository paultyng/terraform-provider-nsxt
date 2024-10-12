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

func TestAccResourceNsxtLogicalTier0Router_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_logical_tier0_router.test"
	haMode := "ACTIVE_STANDBY"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalTier0RouterCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalTier0RouterCreateTemplate(name, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier0RouterExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
			{
				Config: testAccNSXLogicalTier0RouterUpdateTemplate(updateName, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier0RouterExists(updateName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalTier0Router_active(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_logical_tier0_router.test"
	haMode := "ACTIVE_ACTIVE"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalTier0RouterCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalTier0RouterCreateTemplate(name, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier0RouterExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
			{
				Config: testAccNSXLogicalTier0RouterUpdateTemplate(name, haMode, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLogicalTier0RouterExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "high_availability_mode", haMode),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLogicalTier0Router_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_logical_tier0_router.test"
	haMode := "ACTIVE_STANDBY"
	edgeClusterName := getEdgeClusterName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLogicalTier0RouterCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLogicalTier0RouterCreateTemplate(name, haMode, edgeClusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLogicalTier0RouterExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX logical tier0 router resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX logical tier0 router resource ID not set in resources")
		}

		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving logical tier0 router ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking verifying logical tier0 router existence. HTTP returned %d", responseCode.StatusCode)
		}

		if displayName == resource.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX logical tier0 router %s not found", displayName)
	}
}

func testAccNSXLogicalTier0RouterCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_logical_tier0_router" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		resource, responseCode, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving logical tier0 router ID %s. Error: %v", resourceID, err)
		}

		if displayName == resource.DisplayName {
			return fmt.Errorf("NSX logical tier0 router %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLogicalTier0RouterCreateTemplate(name string, haMode string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_logical_tier0_router" "test" {
  display_name           = "%s"
  description            = "Acceptance Test"
  high_availability_mode = "%s"
  edge_cluster_id        = "${data.nsxt_edge_cluster.EC.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, edgeClusterName, name, haMode)
}

func testAccNSXLogicalTier0RouterUpdateTemplate(name string, haMode string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
     display_name = "%s"
}

resource "nsxt_logical_tier0_router" "test" {
  display_name           = "%s"
  description            = "Acceptance Test Update"
  high_availability_mode = "%s"
  edge_cluster_id        = "${data.nsxt_edge_cluster.EC.id}"

  tag {
    scope = "scope3"
    tag   = "tag3"
  }
}`, edgeClusterName, name, haMode)
}
