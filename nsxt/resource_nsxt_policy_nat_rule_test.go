// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var testAccResourcePolicyNATRuleName = "nsxt_policy_nat_rule.test"
var testAccResourcePolicyNATRuleSourceNet = "14.1.1.3"
var testAccResourcePolicyNATRuleDestNet = "15.1.1.3"
var testAccResourcePolicyNATRuleTransNet = "16.1.1.3"

func TestAccResourceNsxtPolicyNATRule_minimalT0(t *testing.T) {
	name := getAccTestResourceName()
	action := model.PolicyNatRule_ACTION_REFLEXIVE

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier0MinimalCreateTemplate(name, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleTransNet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", testAccResourcePolicyNATRuleSourceNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", testAccResourcePolicyNATRuleTransNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_basic_T1(t *testing.T) {
	testAccResourceNsxtPolicyNATRuleBasicT1(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyNATRule_basicT1_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyNATRuleBasicT1(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyNATRuleBasicT1(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	snet := "22.1.1.2"
	dnet := "33.1.1.2"
	tnet := "44.1.1.2"
	action := model.PolicyNatRule_ACTION_DNAT

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleDestNet, testAccResourcePolicyNATRuleTransNet, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", testAccResourcePolicyNATRuleDestNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", testAccResourcePolicyNATRuleSourceNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", testAccResourcePolicyNATRuleTransNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(updateName, action, snet, dnet, tnet, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleTier1UpdateMultipleSourceNetworksTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, snet, dnet, tnet, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRuleT1_natType(t *testing.T) {
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleNatTypeTemplate(name, "DEFAULT", model.PolicyNatRule_ACTION_DNAT, "22.1.1.14", "33.1.1.14"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "DEFAULT"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleNatTypeTemplate(name, "USER", model.PolicyNatRule_ACTION_DNAT, "22.1.1.14", "33.1.1.14"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleNatTypeTemplate(name, "NAT64", model.PolicyNatRule_ACTION_NAT64, "2201::0014", "3301:1122::1280:0014"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "NAT64"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_withPolicyBasedVpnMode(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	snet := "22.1.1.2"
	dnet := "33.1.1.2"
	tnet := "44.1.1.2"
	action := model.PolicyNatRule_ACTION_DNAT

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "4.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplateWithPolicyBasedVpnMode(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleDestNet, testAccResourcePolicyNATRuleTransNet, model.PolicyNatRule_POLICY_BASED_VPN_MODE_BYPASS, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", testAccResourcePolicyNATRuleDestNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", testAccResourcePolicyNATRuleSourceNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", testAccResourcePolicyNATRuleTransNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "policy_based_vpn_mode", model.PolicyNatRule_POLICY_BASED_VPN_MODE_BYPASS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplateWithPolicyBasedVpnMode(updateName, action, snet, dnet, tnet, model.PolicyNatRule_POLICY_BASED_VPN_MODE_MATCH, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "policy_based_vpn_mode", model.PolicyNatRule_POLICY_BASED_VPN_MODE_MATCH),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_basicT0(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	snet := "22.1.1.2"
	tnet := "44.1.1.2"
	action := model.PolicyNatRule_ACTION_REFLEXIVE

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, updateName, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier0CreateTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleTransNet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", testAccResourcePolicyNATRuleSourceNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", testAccResourcePolicyNATRuleTransNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "scope.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleTier0CreateTemplate(updateName, action, snet, tnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "scope.#", "2"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_basicT1Import(t *testing.T) {
	name := getAccTestResourceName()
	action := model.PolicyNatRule_ACTION_DNAT

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleDestNet, testAccResourcePolicyNATRuleTransNet, false),
			},
			{
				ResourceName:      testAccResourcePolicyNATRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyNATRuleImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_basicT1Import_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	action := model.PolicyNatRule_ACTION_DNAT

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleDestNet, testAccResourcePolicyNATRuleTransNet, true),
			},
			{
				ResourceName:      testAccResourcePolicyNATRuleName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testAccResourcePolicyNATRuleName),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRule_nat64T1(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	snet := "2201::100:11:11:0"
	dnet := "2001:db8:122:344::/96"
	tnet := "44.1.1.2"
	tnet1 := "44.1.1.3"
	action := model.PolicyNatRule_ACTION_NAT64

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, name, true)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(name, action, snet, dnet, tnet, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "NAT64"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyNATRuleTier1CreateTemplate(updateName, action, snet, dnet, tnet1, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "NAT64"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "translated_networks.0", tnet1),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_BYPASS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyNATRuleNoSnatWithoutTNet(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	snet := "22.1.1.2"
	dnet := "33.1.1.2"
	action := model.PolicyNatRule_ACTION_NO_SNAT
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyNATRuleCheckDestroy(state, updateName, false)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxPolicyNatRuleNoTranslatedNetworkTemplate(name, action, testAccResourcePolicyNATRuleSourceNet, testAccResourcePolicyNATRuleDestNet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", name),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", testAccResourcePolicyNATRuleSourceNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", testAccResourcePolicyNATRuleDestNet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
			{
				Config: testAccNsxPolicyNatRuleNoTranslatedNetworkTemplate(updateName, action, snet, dnet),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyNATRuleExists(testAccResourcePolicyNATRuleName, "USER"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "display_name", updateName),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.#", "1"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "source_networks.0", snet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "destination_networks.0", dnet),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "action", action),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "logging", "false"),
					resource.TestCheckResourceAttr(testAccResourcePolicyNATRuleName, "firewall_match", model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "path"),
					resource.TestCheckResourceAttrSet(testAccResourcePolicyNATRuleName, "revision"),
				),
			},
		},
	})
}

func testAccNSXPolicyNATRuleImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccResourcePolicyNATRuleName]
	if !ok {
		return "", fmt.Errorf("NSX Policy NAT Rule resource %s not found in resources", testAccResourcePolicyNATRuleName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy NAT Rule resource ID not set in resources ")
	}
	gwPath := rs.Primary.Attributes["gateway_path"]
	if gwPath == "" {
		return "", fmt.Errorf("NSX Policy NAT Rule Gateway Policy Path not set in resources ")
	}
	_, gwID := parseGatewayPolicyPath(gwPath)
	return fmt.Sprintf("%s/%s", gwID, resourceID), nil
}

func testAccNsxtPolicyNATRuleExists(resourceName string, natType string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy NAT Rule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy NAT Rule resource ID not set in resources")
		}

		gwPath := rs.Primary.Attributes["gateway_path"]
		isT0, gwID := parseGatewayPolicyPath(gwPath)
		_, err := getNsxtPolicyNATRuleByID(testAccGetSessionContext(), connector, gwID, isT0, natType, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy NAT Rule ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyNATRuleCheckDestroy(state *terraform.State, displayName string, isNat bool) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_nat_rule" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		gwPath := rs.Primary.Attributes["gateway_path"]
		isT0, gwID := parseGatewayPolicyPath(gwPath)
		natType := model.PolicyNat_NAT_TYPE_USER
		if isNat {
			natType = model.PolicyNat_NAT_TYPE_NAT64
		}
		_, err := getNsxtPolicyNATRuleByID(testAccGetSessionContext(), connector, gwID, isT0, natType, resourceID)
		if err == nil {
			return fmt.Errorf("Policy NAT Rule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyNATRuleTier0MinimalCreateTemplate(name, sourceNet, translatedNet string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) + fmt.Sprintf(`
resource "nsxt_policy_nat_rule" "test" {
  display_name         = "%s"
  gateway_path         = nsxt_policy_tier0_gateway.test.path
  action               = "%s"
  source_networks      = ["%s"]
  translated_networks  = ["%s"]
}
`, name, model.PolicyNatRule_ACTION_REFLEXIVE, sourceNet, translatedNet)
}

func testAccNsxtPolicyNATRuleNatTypeTemplate(name string, natType string, action string, srcIP string, dstIP string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier1WithEdgeClusterTemplate("test", false, false) + fmt.Sprintf(`
data "nsxt_policy_service" "test" {
  display_name = "DNS-UDP"
}

resource "nsxt_policy_nat_rule" "test" {
  display_name          = "%s"
  type                  = "%s"
  description           = "Acceptance Test"
  gateway_path          = nsxt_policy_tier1_gateway.test.path
  action                = "%s"
  source_networks       = ["%s"]
  destination_networks  = ["%s"]
  translated_networks   = ["44.11.11.2"]
  service               = data.nsxt_policy_service.test.path
}
`, name, natType, action, srcIP, dstIP)
}

func testAccNsxtPolicyNATRuleTier1CreateTemplate(name string, action string, sourceNet string, destNet string, translatedNet string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier1WithEdgeClusterTemplate("test", false, withContext) + fmt.Sprintf(`
data "nsxt_policy_service" "test" {
  display_name = "DNS-UDP"
}

resource "nsxt_policy_nat_rule" "test" {
%s
  display_name          = "%s"
  description           = "Acceptance Test"
  gateway_path          = nsxt_policy_tier1_gateway.test.path
  action                = "%s"
  source_networks       = ["%s"]
  destination_networks  = ["%s"]
  translated_networks   = ["%s"]
  logging               = false
  firewall_match        = "%s"
  service               = data.nsxt_policy_service.test.path

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, name, action, sourceNet, destNet, translatedNet, model.PolicyNatRule_FIREWALL_MATCH_BYPASS)
}

func testAccNsxtPolicyNATRuleTier1CreateTemplateWithPolicyBasedVpnMode(name string, action string, sourceNet string, destNet string, translatedNet string, policyBasedVpnMode string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier1WithEdgeClusterTemplate("test", false, withContext) + fmt.Sprintf(`
data "nsxt_policy_service" "test" {
  display_name = "DNS-UDP"
}

resource "nsxt_policy_nat_rule" "test" {
%s
  display_name          = "%s"
  description           = "Acceptance Test"
  gateway_path          = nsxt_policy_tier1_gateway.test.path
  action                = "%s"
  source_networks       = ["%s"]
  destination_networks  = ["%s"]
  translated_networks   = ["%s"]
  logging               = false
  firewall_match        = "%s"
  service               = data.nsxt_policy_service.test.path
  policy_based_vpn_mode = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, name, action, sourceNet, destNet, translatedNet, model.PolicyNatRule_FIREWALL_MATCH_BYPASS, policyBasedVpnMode)
}

func testAccNsxtPolicyNATRuleTier1UpdateMultipleSourceNetworksTemplate(name string, action string, sourceNet1 string, sourceNet2 string, destNet string, translatedNet string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier1WithEdgeClusterTemplate("test", false, withContext) + fmt.Sprintf(`
resource "nsxt_policy_nat_rule" "test" {
%s
  display_name          = "%s"
  description           = "Acceptance Test"
  gateway_path          = nsxt_policy_tier1_gateway.test.path
  action                = "%s"
  source_networks       = ["%s", "%s"]
  destination_networks  = ["%s"]
  translated_networks   = ["%s"]
  logging               = false
  firewall_match        = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, name, action, sourceNet1, sourceNet2, destNet, translatedNet, model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS)
}

func testAccNsxtPolicyNATRuleTier0CreateTemplate(name string, action string, sourceNet string, translatedNet string) string {

	var transportZone string
	interfaceSite := ""
	if testAccIsGlobalManager() {
		transportZone = `
data "nsxt_policy_transport_zone" "test" {
  transport_type = "VLAN_BACKED"
  is_default     = true
  site_path      = data.nsxt_policy_site.test.path
}`
		interfaceSite = "site_path = data.nsxt_policy_site.test.path"
	} else {
		transportZone = fmt.Sprintf(`
data "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
}`, getVlanTransportZoneName())
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		transportZone + testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", true) + fmt.Sprintf(`
resource "nsxt_policy_vlan_segment" "test" {
  count               = 2
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  display_name        = "interface-test-${count.index}"
  vlan_ids            = [10 + count.index]
}

resource "nsxt_policy_tier0_gateway_interface" "test" {
  count        = 2
  display_name = "gwinterface-test-${count.index}"
  type         = "SERVICE"
  gateway_path = nsxt_policy_tier0_gateway.test.path
  segment_path = nsxt_policy_vlan_segment.test[count.index].path
  subnets      = ["1.1.${count.index}.2/24"]
  %s
}

resource "nsxt_policy_nat_rule" "test" {
  display_name          = "%s"
  description           = "Acceptance Test"
  gateway_path          = nsxt_policy_tier0_gateway.test.path
  action                = "%s"
  source_networks       = ["%s"]
  translated_networks   = ["%s"]
  logging               = false
  firewall_match        = "%s"
  scope                 = [nsxt_policy_tier0_gateway_interface.test[1].path, nsxt_policy_tier0_gateway_interface.test[0].path]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

`, interfaceSite, name, action, sourceNet, translatedNet, model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS)
}

func testAccNsxPolicyNatRuleNoTranslatedNetworkTemplate(name string, action string, sourceNet string, destNet string) string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier1WithEdgeClusterTemplate("test", false, false) + fmt.Sprintf(`
	resource "nsxt_policy_nat_rule" "test" {
	  display_name          = "%s"
	  description           = "Acceptance Test"
	  gateway_path          = nsxt_policy_tier1_gateway.test.path
	  action                = "%s"
	  source_networks       = ["%s"]
	  destination_networks  = ["%s"]
	  logging               = false
	  firewall_match        = "%s"

	  tag {
		scope = "scope1"
		tag   = "tag1"
	  }

	  tag {
		scope = "scope2"
		tag   = "tag2"
	  }
	}
	`, name, action, sourceNet, destNet, model.PolicyNatRule_FIREWALL_MATCH_MATCH_EXTERNAL_ADDRESS)
}
