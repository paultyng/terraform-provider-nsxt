/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtPolicyTier1Gateway_basic(t *testing.T) {
	testAccResourceNsxtPolicyTier1GatewayBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyTier1GatewayBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyTier1GatewayBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyTier1GatewayBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	failoverMode := "NON_PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1CreateTemplate(name, failoverMode, false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "true"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1UpdateTemplate(updateName, failoverMode, false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "false"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1Update2Template(updateName, failoverMode, false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "false"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withPoolAllocation(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	failoverMode := "NON_PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1CreateTemplate(name, failoverMode, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "true"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "pool_allocation", "ROUTING"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1UpdateTemplate(updateName, failoverMode, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "false"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "pool_allocation", "ROUTING"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				// ForceNew due to pool allocation change
				Config: testAccNsxtPolicyTier1Update2Template(updateName, failoverMode, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "failover_mode", failoverMode),
					resource.TestCheckResourceAttr(testResourceName, "default_rule_logging", "false"),
					resource.TestCheckResourceAttr(testResourceName, "enable_firewall", "true"),
					resource.TestCheckResourceAttr(testResourceName, "enable_standby_relocation", "false"),
					resource.TestCheckResourceAttr(testResourceName, "force_whitelisting", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_types.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_ndra_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "ipv6_dad_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "pool_allocation", "LB_SMALL"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withDHCP(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1CreateWithDHCPTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttrSet(testResourceName, "dhcp_config_path"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1CreateWithDHCPRemovedTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config_path", ""),
				),
			},
			{
				Config: testAccNsxtPolicyTier1CreateDHCPTemplate(),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withEdgeCluster(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	edgeClusterName := getEdgeClusterName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1CreateWithEcTemplate(name, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_path"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1CreateWithEcTemplate(updateName, edgeClusterName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "edge_cluster_path"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withId(t *testing.T) {
	name := getAccTestResourceName()
	id := "test-id"
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1SetTemplateWithID(name, id),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "id", id),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "type", "ISOLATED"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1SetTemplateWithID(updateName, id),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "id", id),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "ha_mode", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "type", "ISOLATED"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
				),
			},
		},
	})
}
func TestAccResourceNsxtPolicyTier1Gateway_withQos(t *testing.T) {
	name := getAccTestResourceName()
	profileName := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyGatewayQosProfileDeleteByName(profileName); err != nil {
				return err
			}
			return testAccNsxtPolicyTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyGatewayQosProfileCreate(profileName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyTier1TemplateWithQos(name, profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "ingress_qos_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "egress_qos_profile_path"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1TemplateRemoveQos(updateName, profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "ingress_qos_profile_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "egress_qos_profile_path", ""),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withRules(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1CreateTemplateWithRules(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.0.name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.0.prefix_operator", "EQ"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
				),
			},
			{
				Config: testAccNsxtPolicyTier1UpdateTemplateWithRules(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "route_advertisement_rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_withTier0(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	tier0Name := getAccTestResourceName()
	failoverMode := "NON_PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			err := testAccDataSourceNsxtPolicyTier0GatewayDeleteByName(tier0Name)
			err2 := testAccNsxtPolicyTier1CheckDestroy(state, updateName)
			if err != nil {
				return err
			}

			if err2 != nil {
				return err2
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyTier0GatewayCreate(tier0Name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyTier1CreateWithTier0Template(name, tier0Name, failoverMode),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "tier0_path"),
					resource.TestCheckResourceAttr(realizationResourceName, "state", "REALIZED"),
				),
			},
			{
				Config: testAccNsxtPolicyTier1UpdateTemplate(updateName, failoverMode, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTier1Exists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "tier0_path", ""),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	failoverMode := "PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1ImportTemplate(name, failoverMode, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier1Gateway_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	failoverMode := "PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1ImportTemplate(name, failoverMode, true),
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

func TestAccResourceNsxtPolicyTier1Gateway_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier1_gateway.test"
	failoverMode := "PREEMPTIVE"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier1CheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier1ImportTemplate(name, failoverMode, false),
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

func testAccNsxtPolicyTier1Exists(resourceName string) resource.TestCheckFunc {
	return testAccNsxtPolicyResourceExists(testAccGetSessionContext(), resourceName, resourceNsxtPolicyTier1GatewayExists)
}

func testAccNsxtPolicyTier1CheckDestroy(state *terraform.State, displayName string) error {
	return testAccNsxtPolicyResourceCheckDestroy(testAccGetSessionContext(), state, displayName, "nsxt_policy_tier1_gateway", resourceNsxtPolicyTier1GatewayExists)
}

func testAccNsxtPolicyTier1CreateDHCPTemplate() string {
	return `
resource "nsxt_policy_dhcp_relay" "test" {
  display_name      = "terraform-dhcp-relay"
  server_addresses  = ["88.9.9.2"]
}
`
}

func testAccNsxtPolicyTier1CreateWithDHCPTemplate(name string) string {
	return testAccNsxtPolicyTier1CreateDHCPTemplate() + fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  dhcp_config_path  = nsxt_policy_dhcp_relay.test.path
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, getEdgeClusterName(), name)
}

func testAccNsxtPolicyTier1CreateWithDHCPRemovedTemplate(name string) string {
	return testAccNsxtPolicyTier1CreateDHCPTemplate() + fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name      = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, getEdgeClusterName(), name)
}

func testAccNsxtPolicyTier1CreateTemplate(name string, failoverMode string, withPoolAllocation, withContext bool) string {
	poolAllocation := ""
	if withPoolAllocation {
		poolAllocation = `pool_allocation = "ROUTING"`
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	config := fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test"
  failover_mode             = "%s"
  default_rule_logging      = "true"
  enable_firewall           = "false"
  enable_standby_relocation = "true"
  force_whitelisting        = "false"
  route_advertisement_types = ["TIER1_STATIC_ROUTES", "TIER1_CONNECTED"]
  ipv6_ndra_profile_path    = "/infra/ipv6-ndra-profiles/default"
  ipv6_dad_profile_path     = "/infra/ipv6-dad-profiles/default"
  %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, context, name, failoverMode, poolAllocation)
	return testAccAdjustPolicyInfraConfig(config)
}

func testAccNsxtPolicyTier1UpdateTemplate(name string, failoverMode string, withPoolAllocation, withContext bool) string {
	poolAllocation := ""
	if withPoolAllocation {
		poolAllocation = `pool_allocation = "ROUTING"`
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	config := fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test Update"
  failover_mode             = "%s"
  default_rule_logging      = "false"
  enable_firewall           = "true"
  enable_standby_relocation = "false"
  force_whitelisting        = "true"
  route_advertisement_types = ["TIER1_CONNECTED"]
  ipv6_ndra_profile_path    = "/infra/ipv6-ndra-profiles/default"
  ipv6_dad_profile_path     = "/infra/ipv6-dad-profiles/default"
  %s

  tag {
    scope = "scope3"
    tag   = "tag3"
  }
}`, context, name, failoverMode, poolAllocation)
	return testAccAdjustPolicyInfraConfig(config)
}

func testAccNsxtPolicyTier1Update2Template(name string, failoverMode string, withPoolAllocation, withContext bool) string {
	poolAllocation := ""
	if withPoolAllocation {
		poolAllocation = `pool_allocation = "LB_SMALL"`
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	config := fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test Update"
  failover_mode             = "%s"
  default_rule_logging      = "false"
  enable_firewall           = "true"
  enable_standby_relocation = "false"
  force_whitelisting        = "true"
  route_advertisement_types = ["TIER1_CONNECTED"]
  ipv6_ndra_profile_path    = "/infra/ipv6-ndra-profiles/default"
  ipv6_dad_profile_path     = "/infra/ipv6-dad-profiles/default"
  %s

  tag {
    scope = "scope3"
    tag   = "tag3"
  }
}`, context, name, failoverMode, poolAllocation)

	return testAccAdjustPolicyInfraConfig(config)
}

func testAccNsxtPolicyTier1ImportTemplate(name string, failoverMode string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
%s
  display_name              = "%s"
  description               = "Acceptance Test"
  failover_mode             = "%s"
  default_rule_logging      = "true"
  enable_firewall           = "false"
  enable_standby_relocation = "true"
  force_whitelisting        = "false"
  route_advertisement_types = ["TIER1_STATIC_ROUTES", "TIER1_CONNECTED"]
  ipv6_ndra_profile_path    = "/infra/ipv6-ndra-profiles/default"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
%s
  path = nsxt_policy_tier1_gateway.test.path
}`, context, name, failoverMode, context)
}

func testAccNsxtPolicyTier1CreateWithEcTemplate(name string, edgeClusterName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name      = "%s"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, edgeClusterName, name)
}

func testAccNsxtPolicyTier1CreateWithTier0Template(name string, tier0Name string, failoverMode string) string {
	return fmt.Sprintf(`
data "nsxt_policy_tier0_gateway" "T0" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name  = "%s"
  description   = "Acceptance Test"
  tier0_path    = data.nsxt_policy_tier0_gateway.T0.path
  failover_mode = "%s"
  type          = "ROUTED"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, tier0Name, name, failoverMode)
}

func testAccNsxtPolicyTier1SetTemplateWithID(name string, id string) string {
	return fmt.Sprintf(`

resource "nsxt_policy_tier1_gateway" "test" {
  nsx_id       = "%s"
  display_name = "%s"
  description  = "Acceptance Test"
  ha_mode      = "NONE"
  type         = "ISOLATED"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, id, name)
}

func testAccNsxtPolicyTier1CreateTemplateWithRules(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  route_advertisement_rule {
    name                      = "rule1"
    action                    = "DENY"
    subnets                   = ["20.0.0.0/24", "21.0.0.0/24"]
    route_advertisement_types = ["TIER1_CONNECTED"]
    prefix_operator           = "EQ"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_tier1_gateway.test.path
}`, name)
}

func testAccNsxtPolicyTier1UpdateTemplateWithRules(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  route_advertisement_rule {
    name                      = "rule1"
    action                    = "DENY"
    subnets                   = ["20.0.0.0/24", "21.0.0.0/24"]
    route_advertisement_types = ["TIER1_CONNECTED"]
  }

  route_advertisement_rule {
    name            = "rule2"
    action          = "PERMIT"
    subnets         = ["30.0.0.0/24", "31.0.0.0/24"]
    prefix_operator = "GE"
  }
}`, name)
}

func testAccNsxtPolicyTier1TemplateWithQos(name string, profileName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_gateway_qos_profile" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name             = "%s"
  ingress_qos_profile_path = data.nsxt_policy_gateway_qos_profile.test.path
  egress_qos_profile_path  = data.nsxt_policy_gateway_qos_profile.test.path
}`, profileName, name)
}

func testAccNsxtPolicyTier1TemplateRemoveQos(name string, profileName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_gateway_qos_profile" "test" {
  display_name = "%s"
}

resource "nsxt_policy_tier1_gateway" "test" {
  display_name             = "%s"
}`, profileName, name)
}
