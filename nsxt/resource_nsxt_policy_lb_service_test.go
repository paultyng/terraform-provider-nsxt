// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
)

var accTestPolicyLBServiceCreateAttributes = map[string]string{
	"display_name":      getAccTestResourceName(),
	"description":       "terraform created",
	"connectivity_path": "nsxt_policy_tier1_gateway.test1.path",
	"enabled":           "true",
	"error_log_level":   "ERROR",
	"size":              "SMALL",
}

var accTestPolicyLBServiceUpdateAttributes = map[string]string{
	"display_name":      getAccTestResourceName(),
	"description":       "terraform updated",
	"connectivity_path": "nsxt_policy_tier1_gateway.test2.path",
	"enabled":           "false",
	"error_log_level":   "EMERGENCY",
	"size":              "SMALL",
}

var accTestPolicyLBServiceGateway1Name = getAccTestResourceName()
var accTestPolicyLBServiceGateway2Name = getAccTestResourceName()

func TestAccResourceNsxtPolicyLBService_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBServiceCheckDestroy(state, accTestPolicyLBServiceCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBServiceTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "connectivity_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyLBServiceCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", accTestPolicyLBServiceCreateAttributes["error_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "size", accTestPolicyLBServiceCreateAttributes["size"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBServiceTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBServiceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBServiceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "connectivity_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyLBServiceUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", accTestPolicyLBServiceUpdateAttributes["error_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "size", accTestPolicyLBServiceUpdateAttributes["size"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBServiceMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttr(testResourceName, "connectivity_path", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBService_importBasic(t *testing.T) {
	name := accTestPolicyLBServiceUpdateAttributes["display_name"]
	testResourceName := "nsxt_policy_lb_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBServiceMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBServiceExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		nsxClient := infra.NewLbServicesClient(connector)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBService resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBService resource ID not set in resources")
		}

		_, err := nsxClient.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy LBService ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyLBServiceCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	nsxClient := infra.NewLbServicesClient(connector)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		_, err := nsxClient.Get(resourceID)
		if err == nil {
			return fmt.Errorf("Policy LBService %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBServiceTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBServiceCreateAttributes
	} else {
		attrMap = accTestPolicyLBServiceUpdateAttributes
	}
	return testAccNsxtPolicyLBServiceDeps() + fmt.Sprintf(`

resource "nsxt_policy_tier1_gateway" "test1" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  tier0_path        = nsxt_policy_tier0_gateway.test.path
}

resource "nsxt_policy_tier1_gateway" "test2" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  tier0_path        = nsxt_policy_tier0_gateway.test.path
}

resource "nsxt_policy_lb_service" "test" {
  display_name      = "%s"
  description       = "%s"
  connectivity_path = %s
  enabled           = %s
  error_log_level   = "%s"
  size              = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_service.test.path
}`, accTestPolicyLBServiceGateway1Name, accTestPolicyLBServiceGateway2Name, attrMap["display_name"], attrMap["description"], attrMap["connectivity_path"], attrMap["enabled"], attrMap["error_log_level"], attrMap["size"])
}

// Terraform test does not respect T1-T0 dependency upon destroy,
// so we need workaround that detaches T1 from T0 to avoid destroy error
func testAccNsxtPolicyLBServiceMinimalistic() string {
	return testAccNsxtPolicyLBServiceDeps() + fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test1" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
}

resource "nsxt_policy_tier1_gateway" "test2" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
}
resource "nsxt_policy_lb_service" "test" {
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_service.test.path
}`, accTestPolicyLBServiceGateway1Name, accTestPolicyLBServiceGateway2Name, accTestPolicyLBServiceUpdateAttributes["display_name"])
}

func testAccNsxtPolicyLBServiceDeps() string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name = "terraform-lb-test"
}`, getEdgeClusterName())

}
