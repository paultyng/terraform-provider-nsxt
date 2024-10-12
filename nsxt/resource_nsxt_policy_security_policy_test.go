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

func TestAccResourceNsxtPolicySecurityPolicy_basic(t *testing.T) {
	testAccResourceNsxtPolicySecurityPolicyBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicySecurityPolicy_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicySecurityPolicyBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicySecurityPolicyBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	resourceName := "nsxt_policy_security_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)
	comments1 := "Acceptance test create"
	comments2 := "Acceptance test update"
	direction1 := "IN"
	direction2 := "OUT"
	proto1 := "IPV4"
	proto2 := "IPV4_IPV6"
	defaultAction := "ALLOW"
	tag1 := "abc"
	tag2 := "def"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, updatedName, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, name, comments1, defaultDomain, withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments1),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, updatedName, comments2, defaultDomain, withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments2),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithRule(resourceName, updatedName, direction1, proto1, tag1, defaultDomain, "", withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag1),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.0.rule_id"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithRule(resourceName, updatedName, direction2, proto2, tag2, defaultDomain, "", withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithProfiles(resourceName, updatedName, direction2, proto2, tag2, defaultDomain, withContext, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.profiles.#", "1"),
				),
			},
		},
	})
}
func TestAccResourceNsxtPolicySecurityPolicy_EthernetRule(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_security_policy.test"
	direction1 := "IN"
	proto1 := "NONE"
	defaultAction := "ALLOW"
	tag1 := "abc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyWithEthernetRule(name, direction1, proto1, tag1, defaultDomain, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Ethernet"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag1),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.0.rule_id"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySecurityPolicy_withDependencies(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_security_policy.test"
	defaultAction := "ALLOW"
	defaultDirection := "IN_OUT"
	defaultProtocol := "IPV4_IPV6"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyWithDepsCreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Environment"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sequence_number", "20"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources_excluded", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations_excluded", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.services.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "rule2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.sequence_number", "30"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.source_groups.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.disabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.services.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "rule.1.rule_id"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithDepsUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Environment"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", "JUMP_TO_APPLICATION"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.disabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.services.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySecurityPolicy_withIPCidrRange(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_security_policy.test"
	defaultDirection := "IN_OUT"
	defaultProtocol := "IPV4_IPV6"
	policyIP := "10.10.20.5"
	policyCidr := "10.10.20.0/22"
	policyRange := "10.10.20.6-10.10.20.7"
	updatedPolicyIP := "10.10.40.5"
	updatedPolicyCidr := "10.10.40.0/22"
	updatedPolicyRange := "10.10.40.6-10.10.40.7"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyWithIPCidrRange(name, policyIP, policyCidr, policyRange, policyIP, policyCidr, policyRange),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "6"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.0", policyIP),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "rule2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.0", policyCidr),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.display_name", "rule3"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.destination_groups.0", policyRange),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.display_name", "rule4"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.source_groups.0", policyIP),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.display_name", "rule5"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.source_groups.0", policyCidr),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.display_name", "rule6"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.source_groups.0", policyRange),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.destination_groups.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithIPCidrRange(name, updatedPolicyIP, updatedPolicyCidr, updatedPolicyRange, updatedPolicyIP, updatedPolicyCidr, updatedPolicyRange),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "6"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.0", updatedPolicyIP),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "rule2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.0", updatedPolicyCidr),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.display_name", "rule3"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.2.destination_groups.0", updatedPolicyRange),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.display_name", "rule4"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.source_groups.0", updatedPolicyIP),
					resource.TestCheckResourceAttr(testResourceName, "rule.3.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.display_name", "rule5"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.source_groups.0", updatedPolicyCidr),
					resource.TestCheckResourceAttr(testResourceName, "rule.4.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.display_name", "rule6"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.action", "ALLOW"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.source_groups.0", updatedPolicyRange),
					resource.TestCheckResourceAttr(testResourceName, "rule.5.destination_groups.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySecurityPolicy_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_security_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, name, "import", defaultDomain, false, true),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicySecurityPolicy_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_security_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, name, "import", defaultDomain, true, true),
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

func TestAccResourceNsxtGlobalPolicySecurityPolicy_withSite(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	resourceName := "nsxt_policy_security_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)
	comments1 := "Acceptance test create"
	comments2 := "Acceptance test update"
	direction1 := "IN"
	direction2 := "OUT"
	proto1 := "IPV4"
	proto2 := "IPV4_IPV6"
	defaultAction := "ALLOW"
	tag1 := "abc"
	tag2 := "def"
	domain := getTestSiteName()
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, updatedName, domain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, name, comments1, domain, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, domain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments1),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyBasic(resourceName, updatedName, comments2, domain, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, domain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments2),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithRule(resourceName, updatedName, direction1, proto1, tag1, domain, "", false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, domain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithRule(resourceName, updatedName, direction2, proto2, tag2, domain, "", false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, domain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicySecurityPolicyWithProfiles(resourceName, updatedName, direction2, proto2, tag2, domain, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(testResourceName, domain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "Application"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "scope.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.profiles.#", "1"),
				),
			},
		},
	})
}

func testAccNsxtPolicySecurityPolicyExists(resourceName string, domainName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy SecurityPolicy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy SecurityPolicy resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicySecurityPolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy SecurityPolicy ID %s", resourceID)
		}
		return nil
	}
}

func testAccNsxtPolicySecurityPolicyCheckDestroy(state *terraform.State, displayName string, domainName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_security_policy" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicySecurityPolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy SecurityPolicy %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicySecurityPolicyBasic(resourceName, name, comments, domainName string, withContext, withCategory bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	category := ""
	if withCategory {
		category = `
  category        = "Application"
`
	}
	if domainName == defaultDomain {
		return fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
%s
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

}`, resourceName, context, name, category, comments)
	}
	return testAccNsxtGlobalPolicySite(domainName) + fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Application"
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  domain          = data.nsxt_policy_site.test.id

  tag {
    scope = "color"
    tag   = "orange"
  }

}`, resourceName, context, name, comments)
}

func testAccNsxtPolicySecurityPolicyWithRule(resourceName, name, direction, protocol, ruleTag, domainName, profiles string, withContext, withCategory bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	category := ""
	if withCategory {
		category = `
  category        = "Application"
`
	}
	if domainName == defaultDomain {
		return fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
%s
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name = "%s"
    direction    = "%s"
    ip_version   = "%s"
    log_label    = "%s"

    tag {
      scope = "color"
      tag   = "blue"
    }
    %s
  }
}`, resourceName, context, name, category, name, direction, protocol, ruleTag, profiles)
	}
	return testAccNsxtGlobalPolicyGroupIPAddressCreateTemplate("group", domainName) + fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Application"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  domain          = data.nsxt_policy_site.test.id

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name = "%s"
    direction    = "%s"
    ip_version   = "%s"
    log_label    = "%s"
    source_groups = [nsxt_policy_group.test.path]

    tag {
      scope = "color"
      tag   = "blue"
    }
    %s
  }
}`, resourceName, context, name, name, direction, protocol, ruleTag, profiles)
}

func testAccNsxtPolicySecurityPolicyWithEthernetRule(name string, direction string, protocol string, ruleTag string, domainName string, profiles string) string {
	if domainName == defaultDomain {
		return fmt.Sprintf(`
resource "nsxt_policy_security_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Ethernet"
  locked          = false
  stateful        = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name = "%s"
    direction    = "%s"
    ip_version   = "%s"
    log_label    = "%s"

    tag {
      scope = "color"
      tag   = "blue"
    }
    %s
  }
}`, name, name, direction, protocol, ruleTag, profiles)
	}
	return testAccNsxtGlobalPolicyGroupIPAddressCreateTemplate("group", domainName) + fmt.Sprintf(`
resource "nsxt_policy_security_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Application"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  domain          = data.nsxt_policy_site.test.id

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name = "%s"
    direction    = "%s"
    ip_version   = "%s"
    log_label    = "%s"
    source_groups = [nsxt_policy_group.test.path]

    tag {
      scope = "color"
      tag   = "blue"
    }
    %s
  }
}`, name, name, direction, protocol, ruleTag, profiles)
}

func testAccNsxtPolicySecurityPolicyDeps() string {
	return `
resource "nsxt_policy_group" "group1" {
  display_name = "terraform testacc 1"
}

resource "nsxt_policy_group" "group2" {
  display_name = "terraform testacc 2"
}

resource "nsxt_policy_service" "icmp" {
    display_name = "security-policy-test-icmp"
    icmp_entry {
        protocol = "ICMPv4"
    }
}

resource "nsxt_policy_service" "tcp778" {
    display_name = "security-policy-test-tcp"
    l4_port_set_entry {
        protocol          = "TCP"
        destination_ports = [ "778" ]
    }
}`
}

// TODO: add profiles when available
func testAccNsxtPolicySecurityPolicyWithDepsCreate(name string) string {
	return testAccNsxtPolicySecurityPolicyDeps() + fmt.Sprintf(`
resource "nsxt_policy_security_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Environment"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  scope           = [nsxt_policy_group.group1.path]

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name          = "rule1"
    sequence_number       = 20
    source_groups         = [nsxt_policy_group.group1.path]
    destination_groups    = [nsxt_policy_group.group2.path]
    sources_excluded      = true
    destinations_excluded = true
    services              = [nsxt_policy_service.icmp.path, nsxt_policy_service.tcp778.path]
  }

  rule {
    display_name          = "rule2"
    sequence_number       = 30
    source_groups         = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    sources_excluded      = false
    destinations_excluded = false
  }
}`, name)
}

func testAccNsxtPolicySecurityPolicyWithDepsUpdate(name string) string {
	return testAccNsxtPolicySecurityPolicyDeps() + fmt.Sprintf(`
resource "nsxt_policy_security_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Environment"
  locked          = true
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  scope           = [nsxt_policy_group.group2.path]

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name          = "rule1"
    destination_groups    = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    sources_excluded      = false
    destinations_excluded = false
    disabled              = true
    action                = "JUMP_TO_APPLICATION"
    logged                = true
  }

}`, name)
}

func testAccNsxtPolicySecurityPolicyWithIPCidrRange(name string, destIP string, destCidr string, destIPRange string, sourceIP string, sourceCidr string, sourceIPRange string) string {
	return testAccNsxtPolicySecurityPolicyDeps() + fmt.Sprintf(`
	resource "nsxt_policy_security_policy" "test" {
		display_name    = "%s"
		description     = "Acceptance Test"
		category        = "Application"
		sequence_number = 3
		locked          = false
		stateful        = true
		tcp_strict      = false

		tag {
		  scope = "color"
		  tag   = "orange"
		}

		rule {
		  display_name          = "rule1"
		  source_groups         = [nsxt_policy_group.group1.path]
		  destination_groups    = ["%s"]
		  services              = [nsxt_policy_service.icmp.path]
		  action                = "ALLOW"
		}

		rule {
			display_name          = "rule2"
			source_groups         = [nsxt_policy_group.group1.path]
			destination_groups    = ["%s"]
			services              = [nsxt_policy_service.icmp.path]
			action                = "ALLOW"
		}

		rule {
			display_name          = "rule3"
			source_groups         = [nsxt_policy_group.group1.path]
			destination_groups    = ["%s"]
			services              = [nsxt_policy_service.icmp.path]
			action                = "ALLOW"
			sequence_number       = 50
		}
		rule {
			display_name          = "rule4"
			source_groups         = ["%s"]
			destination_groups    = [nsxt_policy_group.group2.path]
			services              = [nsxt_policy_service.icmp.path]
			action                = "ALLOW"
		}

		rule {
			  display_name          = "rule5"
			  source_groups         = ["%s"]
			  destination_groups    = [nsxt_policy_group.group2.path]
			  services              = [nsxt_policy_service.icmp.path]
			  action                = "ALLOW"
			  sequence_number       = 105
                }

		rule {
			  display_name          = "rule6"
			  source_groups         = ["%s"]
			  destination_groups    = [nsxt_policy_group.group2.path]
			  services              = [nsxt_policy_service.icmp.path]
			  action                = "ALLOW"
		}
}`, name, destIP, destCidr, destIPRange, sourceIP, sourceCidr, sourceIPRange)
}

func testAccNsxtPolicySecurityPolicyWithProfiles(resourceName, name, direction, protocol, ruleTag, domainName string, withContext bool, isVpc bool) string {
	vpcShare := ""
	withCategory := true
	if isVpc {
		// this is VPC rule, we need to share context profile with the VPC
		// we do this by sharing with project and all its descendants
		withCategory = false
		vpcShare = testAccNsxtProjectShareAll("nsxt_policy_context_profile.test.path")
	}
	profiles := `
profiles = [nsxt_policy_context_profile.test.path]
`
	return testAccNsxtPolicyContextProfileTemplate("security-policy-test-profile", testAccNsxtPolicyContextProfileAttributeDomainNameTemplate(testSystemDomainName), withContext) + vpcShare + testAccNsxtPolicySecurityPolicyWithRule(resourceName, name, direction, protocol, ruleTag, domainName, profiles, withContext, withCategory)
}
