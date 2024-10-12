// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyIPSecVpnLocalEndpointCreateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"local_address": "20.20.0.10",
	"local_id":      "test-create",
}

var accTestPolicyIPSecVpnLocalEndpointUpdateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform updated",
	"local_address": "20.20.0.20",
	"local_id":      "test-update",
}

var testAccPolicyVPNLocalEndpointResourceName = "nsxt_policy_ipsec_vpn_local_endpoint.test"

func TestAccResourceNsxtPolicyIPSecVpnLocalEndpoint_basic(t *testing.T) {
	testResourceName := testAccPolicyVPNLocalEndpointResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnLocalEndpointCheckDestroy(state, accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnLocalEndpointExists(accTestPolicyIPSecVpnLocalEndpointCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnLocalEndpointCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnLocalEndpointCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyIPSecVpnLocalEndpointCreateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "local_id", accTestPolicyIPSecVpnLocalEndpointCreateAttributes["local_id"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnLocalEndpointExists(accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "local_id", accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["local_id"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnLocalEndpointExists(accTestPolicyIPSecVpnLocalEndpointCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(true),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnLocalEndpoint_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := testAccPolicyVPNLocalEndpointResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnLocalEndpointCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnLocalEndpointMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyIPSecVpnLocalEndpointImporterGetID,
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnLocalEndpointExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPSecVpnLocalEndpoint resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]

		exists, err := resourceNsxtPolicyIPSecVpnLocalEndpointExistsOnService(resourceID, connector, servicePath)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy IPSecVpnLocalEndpoint %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyIPSecVpnLocalEndpointCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ipsec_vpn_local_endpoint" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]
		exists, err := resourceNsxtPolicyIPSecVpnLocalEndpointExistsOnService(resourceID, connector, servicePath)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy IPSecVpnLocalEndpoint %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyIPSecVpnLocalEndpointImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccPolicyVPNLocalEndpointResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy VPN Local Endpoint resource %s not found in resources", testAccPolicyFixedSegmentResourceName)
	}
	path := rs.Primary.Attributes["path"]
	if path == "" {
		return "", fmt.Errorf("NSX Policy VPN Local Endpoint path not specified")
	}
	return path, nil
}

var accTestPolicyIPSecVpnGatewayTestName = getAccTestResourceName()
var accTestPolicyIPSecVpnServiceTestName = getAccTestResourceName()

func testAccNsxtPolicyIPSecVpnLocalEndpointPrerequisite() string {
	return fmt.Sprintf(`
   data "nsxt_policy_edge_cluster" "test" {
	 display_name = "%s"
   }
   resource "nsxt_policy_tier0_gateway" "test" {
	 display_name      = "%s"
	 ha_mode           = "ACTIVE_STANDBY"
	 edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
   }
   data "nsxt_policy_gateway_locale_service" "test" {
	 gateway_path = nsxt_policy_tier0_gateway.test.path
	 display_name = "default"
   }
   resource "nsxt_policy_ipsec_vpn_service" "test" {
	 display_name        = "%s"
	 locale_service_path = data.nsxt_policy_gateway_locale_service.test.path
   }`, getEdgeClusterName(), accTestPolicyIPSecVpnGatewayTestName, accTestPolicyIPSecVpnServiceTestName)
}

func testAccNsxtPolicyIPSecVpnLocalEndpointTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnLocalEndpointCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnLocalEndpointUpdateAttributes
	}
	return testAccNsxtPolicyIPSecVpnLocalEndpointPrerequisite() + fmt.Sprintf(`
   resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
	 service_path = nsxt_policy_ipsec_vpn_service.test.path
	 display_name = "%s"
	 description  = "%s"
	 local_address = "%s"
	 local_id = "%s"
	 tag {
	   scope = "scope1"
	   tag   = "tag1"
	 }
   }`, attrMap["display_name"], attrMap["description"], attrMap["local_address"], attrMap["local_id"])
}

func testAccNsxtPolicyIPSecVpnLocalEndpointMinimalistic() string {
	return testAccNsxtPolicyIPSecVpnLocalEndpointPrerequisite() + fmt.Sprintf(`
   resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
	 service_path  = nsxt_policy_ipsec_vpn_service.test.path
	 display_name  = "%s"
	 local_address = "%s"
   }`, accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["display_name"], accTestPolicyIPSecVpnLocalEndpointUpdateAttributes["local_address"])
}
