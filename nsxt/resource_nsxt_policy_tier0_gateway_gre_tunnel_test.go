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

func TestAccResourceNsxtPolicyTier0GatewayGRETunnel_basic(t *testing.T) {
	testResourceName := "nsxt_policy_tier0_gateway_gre_tunnel.test"
	name := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.2")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtNsxtPolicyTier0GatewayGRETunnelCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtNsxtPolicyTier0GatewayGRETunnelTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtNsxtPolicyTier0GatewayGRETunnelExists(testResourceName),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier0GatewayGRETunnel_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier0_gateway_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.2")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtNsxtPolicyTier0GatewayGRETunnelCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtNsxtPolicyTier0GatewayGRETunnelTemplate(name),
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

func testAccNsxtNsxtPolicyTier0GatewayGRETunnelExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("policy Tier0 Gateway GRE Tunnel resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		localeServicePath := rs.Primary.Attributes["locale_service_path"]
		if resourceID == "" {
			return fmt.Errorf("policy Tier0 Gateway GRE Tunnel resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyTier0GatewayGRETunnelExists(resourceID, localeServicePath, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("policy Tier0 Gateway GRE Tunnel %s does not exist", resourceID)
		}
		return nil
	}

}

func testAccNsxtNsxtPolicyTier0GatewayGRETunnelCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_tier0_gateway_gre_tunnel" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		localeServicePath := rs.Primary.Attributes["locale_service_path"]
		exists, err := resourceNsxtPolicyTier0GatewayGRETunnelExists(resourceID, localeServicePath, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("policy GRE Tunnel %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtNsxtPolicyTier0GatewayGRETunnelTemplate(name string) string {
	return testAccNsxtPolicyTier0InterfaceExternalTemplate(name, "192.168.240.10/24", "1500", "true", false) + fmt.Sprintf(`
data "nsxt_policy_gateway_locale_service" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
}

resource "nsxt_policy_tier0_gateway_gre_tunnel" "test" {
  display_name = "%s"
  description  = "Acceptance test GRE tunnel"
  locale_service_path = data.nsxt_policy_gateway_locale_service.test.path
  destination_address = "192.168.221.221"
  tunnel_address {
    source_address = nsxt_policy_tier0_gateway_interface.test.ip_addresses[0]
    edge_path = data.nsxt_policy_edge_node.EN.path
    tunnel_interface_subnet {
       ip_addresses = ["192.168.243.243"]
       prefix_len = 24
    }
  }
  tunnel_keepalive {
    enabled = false
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}
