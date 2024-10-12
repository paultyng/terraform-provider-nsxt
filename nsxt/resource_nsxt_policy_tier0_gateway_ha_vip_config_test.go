// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
)

// Note that this test depends on having at least 2 edge nodes in the cluster
func TestAccResourceNsxtPolicyTier0GatewayHaVipConfig_basic(t *testing.T) {
	subnet1 := "1.1.12.1/24"
	subnet2 := "1.1.12.2/24"
	vipSubnet := "1.1.12.4/24"
	updatedVipSubnet := "1.1.12.5/24"
	tier0Name := getAccTestResourceName()
	updatedTier0Name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier0_gateway_ha_vip_config.test"

	resource.Test(t, resource.TestCase{
		// This test works with local manager but only with 2 edge nodes
		// 3.2.0 added an edge node constraint, meaning we need big edge
		// topology for this test
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccNSXVersionLessThan(t, "3.2.0")
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier0HAVipConfigCheckDestroy(state, "")
		},
		Steps: []resource.TestStep{
			{
				// create the vip config with all the dependencies
				Config: testAccNsxtPolicyTier0HAVipConfigTemplate(tier0Name, "true", subnet1, subnet2, vipSubnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier0HAVipConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.external_interface_paths.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.0", vipSubnet),
					resource.TestCheckResourceAttrSet(testResourceName, "tier0_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_id"),
				),
			},
			{
				// update the vip config
				Config: testAccNsxtPolicyTier0HAVipConfigTemplate(tier0Name, "false", subnet1, subnet2, updatedVipSubnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier0HAVipConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.external_interface_paths.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.0", updatedVipSubnet),
					resource.TestCheckResourceAttrSet(testResourceName, "tier0_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_id"),
				),
			},
			{
				// update the tier0 only
				Config: testAccNsxtPolicyTier0HAVipConfigTemplate(updatedTier0Name, "false", subnet1, subnet2, updatedVipSubnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier0HAVipConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.enabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.external_interface_paths.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.0", updatedVipSubnet),
					resource.TestCheckResourceAttrSet(testResourceName, "tier0_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_id"),
				),
			},
			{
				// Update Tier0 and HA vip config at the same time (This doesn't work currently)
				Config: testAccNsxtPolicyTier0HAVipConfigTemplate(tier0Name, "true", subnet1, subnet2, vipSubnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier0HAVipConfigExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.enabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.external_interface_paths.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "config.0.vip_subnets.0", vipSubnet),
					resource.TestCheckResourceAttrSet(testResourceName, "tier0_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_id"),
				),
			},
		},
	})
}

func testAccNSXPolicyTier0HAVipConfigImporterGetID(s *terraform.State) (string, error) {
	testResourceName := "nsxt_policy_tier0_gateway_ha_vip_config.test"
	rs, ok := s.RootModule().Resources[testResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy Tier0 HA Vip config resource %s not found in resources", testResourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Tier0 HA Vip config resource ID not set in resources ")
	}
	gwID := rs.Primary.Attributes["tier0_id"]
	if gwID == "" {
		return "", fmt.Errorf("NSX Policy HA Vip config Tier0 Gateway ID not set in resources ")
	}
	localeServiceID := rs.Primary.Attributes["locale_service_id"]
	if localeServiceID == "" {
		return "", fmt.Errorf("NSX Policy HA Vip config Tier0 Gateway locale service ID not set in resources ")
	}

	return fmt.Sprintf("%s/%s", gwID, localeServiceID), nil
}

func TestAccResourceNsxtPolicyTier0GatewayHaVipConfig_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_tier0_gateway_ha_vip_config.test"
	subnet1 := "1.1.12.1/24"
	subnet2 := "1.1.12.2/24"
	vipSubnet := "1.1.12.4/24"
	tier0Name := "ha_tier0"

	resource.Test(t, resource.TestCase{
		// This test works with local manager but only with 2 edge nodes
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccNSXVersionLessThan(t, "3.2.0")
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier0HAVipConfigCheckDestroy(state, "")
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier0HAVipConfigTemplate(tier0Name, "true", subnet1, subnet2, vipSubnet),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: testAccNSXPolicyTier0HAVipConfigImporterGetID,
			},
		},
	})
}

func testAccNsxtPolicyTier0HAVipConfigExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Tier0 HA vip config resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Tier0 HA vip config resource ID not set in resources")
		}

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		localeServiceID := rs.Primary.Attributes["locale_service_id"]
		gwID := rs.Primary.Attributes["tier0_id"]
		localeService := policyTier0GetLocaleService(gwID, localeServiceID, connector, testAccIsGlobalManager())
		if localeService == nil {
			return fmt.Errorf("Error while retrieving Locale Service %s for Gateway %s", localeServiceID, gwID)
		}

		if localeService.HaVipConfigs == nil {
			return fmt.Errorf("Error while retrieving policy Tier0 HA vip config %s. HaVipConfigs is empty", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyTier0HAVipConfigCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy'_tier0_gateway_interface" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		localeServiceID := rs.Primary.Attributes["locale_service_id"]
		gwID := rs.Primary.Attributes["tier0_id"]
		if testAccIsGlobalManager() {
			nsxClient := gm_tier0s.NewLocaleServicesClient(connector)
			obj, err := nsxClient.Get(gwID, localeServiceID)
			if err == nil && obj.HaVipConfigs != nil {
				return fmt.Errorf("Policy Tier0 HA vip config %s still exists", resourceID)
			}
		} else {
			nsxClient := tier_0s.NewLocaleServicesClient(connector)
			obj, err := nsxClient.Get(gwID, localeServiceID)
			if err == nil && obj.HaVipConfigs != nil {
				return fmt.Errorf("Policy Tier0 HA vip config %s still exists", resourceID)
			}
		}
	}
	return nil
}

func testAccNsxtPolicyTier0HAVipConfigSiteTemplate() string {
	if testAccIsGlobalManager() {
		return "site_path = data.nsxt_policy_site.test.path"
	}
	return ""
}

func testAccNsxtPolicyTier0HAVipConfigTemplate(tier0Name string, enabled string, subnet1 string, subnet2 string, vipSubnet string) string {
	return testAccNsxtPolicyGatewayFabricDeps(true) + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test1" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "interface_test1"
  vlan_ids            = [11]
  subnet {
      cidr = "10.2.2.2/24"
  }
}

resource "nsxt_policy_vlan_segment" "test2" {
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "interface_test2"
  vlan_ids            = [11]
  subnet {
      cidr = "10.2.2.2/24"
  }
}

data "nsxt_policy_edge_node" "EN0" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  member_index      = 0
}

data "nsxt_policy_edge_node" "EN1" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  member_index      = 1
}

resource "nsxt_policy_tier0_gateway" "test" {
  display_name      = "%s"
  ha_mode           = "ACTIVE_STANDBY"
  %s
}

resource "nsxt_policy_tier0_gateway_interface" "test1" {
  display_name   = "interface1"
  description    = "Acceptance Test"
  type           = "EXTERNAL"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test1.path
  edge_node_path = data.nsxt_policy_edge_node.EN0.path
  subnets        = ["%s"]
  urpf_mode      = "STRICT"
  %s
}

resource "nsxt_policy_tier0_gateway_interface" "test2" {
  display_name   = "interface2"
  description    = "Acceptance Test"
  type           = "EXTERNAL"
  gateway_path   = nsxt_policy_tier0_gateway.test.path
  segment_path   = nsxt_policy_vlan_segment.test2.path
  edge_node_path = data.nsxt_policy_edge_node.EN1.path
  subnets        = ["%s"]
  urpf_mode      = "STRICT"
  %s
  depends_on     = ["nsxt_policy_tier0_gateway_interface.test1"]
}

resource "nsxt_policy_tier0_gateway_ha_vip_config" "test" {
	config {
		enabled                  = %s
		external_interface_paths = [nsxt_policy_tier0_gateway_interface.test1.path, nsxt_policy_tier0_gateway_interface.test2.path]
		vip_subnets              = ["%s"]
	}
}`, tier0Name, testAccNsxtPolicyLocaleServiceECTemplate(),
		subnet1, testAccNsxtPolicyTier0HAVipConfigSiteTemplate(),
		subnet2, testAccNsxtPolicyTier0HAVipConfigSiteTemplate(), enabled, vipSubnet)
}
