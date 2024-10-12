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

func TestAccResourceNsxtPolicyGatewayPolicy_basic(t *testing.T) {
	testAccResourceNsxtPolicyGatewayPolicyBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyGatewayPolicy_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGatewayPolicyBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyGatewayPolicyBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	resourceName := "nsxt_policy_gateway_policy"
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
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, updatedName, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyBasic(resourceName, name, comments1, withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments1),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyBasic(resourceName, updatedName, comments2, withContext, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments2),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyWithRule(resourceName, updatedName, direction1, proto1, tag1, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
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
					resource.TestCheckResourceAttrSet(testResourceName, "rule.0.rule_id"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyWithRule(resourceName, updatedName, direction2, proto2, tag2, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
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
		},
	})
}

func TestAccResourceNsxtPolicyGatewayPolicy_withDependencies(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_policy.test"
	defaultAction := "ALLOW"
	defaultDirection := "IN_OUT"
	defaultProtocol := "IPV4_IPV6"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyWithMultipleRulesCreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.services.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "rule2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.source_groups.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.disabled", "false"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayPolicyWithMultipleRulesUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", "DROP"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.disabled", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.services.#", "1"),
				),
			},
		},
	})
}
func TestAccResourceNsxtPolicyGatewayPolicy_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_gateway_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyBasic(resourceName, name, "import", false, true),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayPolicy_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_gateway_policy"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyBasic(resourceName, name, "import", true, true),
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

func TestAccResourceNsxtPolicyGatewayPolicy_importNoTcpStrict(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyBasicNoTCPStrict(name, "import"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtGlobalPolicyGatewayPolicy_withSite(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	domainName := getTestSiteName()
	testResourceName := "nsxt_policy_gateway_policy.test"
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
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, updatedName, domainName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtGlobalPolicyGatewayPolicyBasic(name, comments1, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, domainName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments1),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtGlobalPolicyGatewayPolicyBasic(updatedName, comments2, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, domainName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments2),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
			{
				Config: testAccNsxtGlobalPolicyGatewayPolicyWithRule(updatedName, direction1, proto1, tag1, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, domainName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
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
				),
			},
			{
				Config: testAccNsxtGlobalPolicyGatewayPolicyWithRule(updatedName, direction2, proto2, tag2, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, domainName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
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
		},
	})
}

func TestAccResourceNsxtGlobalPolicyGatewayPolicy_withDomain(t *testing.T) {
	name := "terraform-test"
	siteName := getTestSiteName()
	domainName := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_policy.test"
	comments := "Acceptance test create"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, domainName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtGlobalPolicyGatewayPolicyWithDomain(name, comments, domainName, siteName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, domainName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "tcp_strict", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGatewayPolicy_withIPCidrRange(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_gateway_policy.test"
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
			return testAccNsxtPolicyGatewayPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGatewayPolicyWithIPCidrRange(name, policyIP, policyCidr, policyRange, policyIP, policyCidr, policyRange),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
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
				Config: testAccNsxtPolicyGatewayPolicyWithIPCidrRange(name, updatedPolicyIP, updatedPolicyCidr, updatedPolicyRange, updatedPolicyIP, updatedPolicyCidr, updatedPolicyRange),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGatewayPolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "category", "LocalGatewayRules"),
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

func testAccNsxtPolicyGatewayPolicyExists(resourceName string, domainName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy GatewayPolicy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy GatewayPolicy resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyGatewayPolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy GatewayPolicy ID %s", resourceID)
		}
		return nil
	}
}

func testAccNsxtPolicyGatewayPolicyCheckDestroy(state *terraform.State, displayName string, domainName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_gateway_policy" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyGatewayPolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy GatewayPolicy %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyGatewayPolicyBasicNoTCPStrict(name string, comments string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = false

}`, name, comments)
}

func testAccNsxtPolicyGatewayPolicyBasic(resourceName, name, comments string, withContext, withCategory bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	category := ""
	if withCategory {
		category = `
  category        = "LocalGatewayRules"
`
	}
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

func testAccNsxtPolicyGatewayPolicyWithRule(resourceName, name, direction, protocol, ruleTag string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "gwt1test" {
%s
  display_name      = "tf-t1-gw"
  description       = "Acceptance Test"
}

resource "%s" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
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
    scope        = [nsxt_policy_tier1_gateway.gwt1test.path]
    log_label    = "%s"

    tag {
      scope = "color"
      tag   = "blue"
    }
  }
}`, context, resourceName, context, name, name, direction, protocol, ruleTag)
}

// TODO: add  profiles when available
func testAccNsxtPolicyGatewayPolicyWithMultipleRulesCreate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "gwt1test" {
  display_name      = "tf-t1-gw"
  description       = "Acceptance Test"
}

resource "nsxt_policy_group" "group1" {
  display_name = "terraform testacc 1"
}

resource "nsxt_policy_group" "group2" {
  display_name = "terraform testacc 2"
}

resource "nsxt_policy_service" "icmp" {
    display_name = "gateway-policy-test-icmp"
    icmp_entry {
        protocol = "ICMPv4"
    }
}

resource "nsxt_policy_service" "tcp778" {
    display_name = "gateway-policy-test-tcp"
    l4_port_set_entry {
        protocol          = "TCP"
        destination_ports = [ "778" ]
    }
}

resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name          = "rule1"
    source_groups         = [nsxt_policy_group.group1.path]
    destination_groups    = [nsxt_policy_group.group2.path]
    scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
    services              = [nsxt_policy_service.icmp.path]
  }

  rule {
    display_name          = "rule2"
    source_groups         = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
  }
}`, name)
}

func testAccNsxtPolicyGatewayPolicyWithMultipleRulesUpdate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "gwt1test" {
  display_name      = "tf-t1-gw"
  description       = "Acceptance Test"
}

resource "nsxt_policy_group" "group1" {
  display_name = "terraform testacc 1"
}

resource "nsxt_policy_group" "group2" {
  display_name = "terraform testacc 2"
}

resource "nsxt_policy_service" "icmp" {
    display_name = "gateway-policy-test-icmp"
    icmp_entry {
        protocol = "ICMPv4"
    }
}

resource "nsxt_policy_service" "tcp778" {
    display_name = "gateway-policy-test-tcp"
    l4_port_set_entry {
        protocol          = "TCP"
        destination_ports = [ "778" ]
    }
}

resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
  locked          = true
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name          = "rule1"
    destination_groups    = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    disabled              = true
    action                = "DROP"
    logged                = true
    scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
    services              = [nsxt_policy_service.tcp778.path]
  }

}`, name)
}

func testAccNsxtGlobalPolicyGatewayPolicyBasic(name string, comments string, domainName string) string {
	return testAccNsxtGlobalPolicySite(domainName) + fmt.Sprintf(`
resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
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

}`, name, comments)
}

func testAccNsxtGlobalPolicyGatewayPolicyWithDomain(name string, comments string, domainName string, siteName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_domain" "domain1" {
  display_name = "%s"
  sites        = ["%s"]
  nsx_id       = "%s"
}

resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = true
  tcp_strict      = false
  domain          = nsxt_policy_domain.domain1.id

  tag {
    scope = "color"
    tag   = "orange"
  }

}`, domainName, siteName, domainName, name, comments)
}

func testAccNsxtGlobalPolicyGatewayPolicyWithRule(name string, direction string, protocol string, ruleTag string, domainName string) string {
	return testAccNsxtGlobalPolicySite(domainName) + fmt.Sprintf(`
resource "nsxt_policy_tier1_gateway" "gwt1test" {
  display_name      = "tf-t1-gw"
  description       = "Acceptance Test"
}

resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "LocalGatewayRules"
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
    scope        = [nsxt_policy_tier1_gateway.gwt1test.path]
    log_label    = "%s"

    tag {
      scope = "color"
      tag   = "blue"
    }
  }
}`, name, name, direction, protocol, ruleTag)
}

func testAccNsxtPolicyGatewayPolicyDeps() string {
	return `
resource "nsxt_policy_tier1_gateway" "gwt1test" {
	display_name      = "tf-t1-gw"
	description       = "Acceptance Test"
}

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
}`
}

func testAccNsxtPolicyGatewayPolicyWithIPCidrRange(name string, destIP string, destCidr string, destIPRange string, sourceIP string, sourceCidr string, sourceIPRange string) string {
	return testAccNsxtPolicyGatewayPolicyDeps() + fmt.Sprintf(`
	resource "nsxt_policy_gateway_policy" "test" {
		display_name    = "%s"
		description     = "Acceptance Test"
		category        = "LocalGatewayRules"
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
		  scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
		  action                = "ALLOW"
		}

		rule {
			display_name          = "rule2"
			source_groups         = [nsxt_policy_group.group1.path]
			destination_groups    = ["%s"]
			services              = [nsxt_policy_service.icmp.path]
			scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
			action                = "ALLOW"
		  }

		  rule {
			display_name          = "rule3"
			source_groups         = [nsxt_policy_group.group1.path]
			destination_groups    = ["%s"]
			services              = [nsxt_policy_service.icmp.path]
			scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
			action                = "ALLOW"
		  }

		  rule {
			display_name          = "rule4"
			source_groups         = ["%s"]
			destination_groups    = [nsxt_policy_group.group2.path]
			services              = [nsxt_policy_service.icmp.path]
			scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
			action                = "ALLOW"
		  }

		  rule {
			  display_name          = "rule5"
			  source_groups         = ["%s"]
			  destination_groups    = [nsxt_policy_group.group2.path]
			  services              = [nsxt_policy_service.icmp.path]
			  scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
			  action                = "ALLOW"
			}

			rule {
			  display_name          = "rule6"
			  source_groups         = ["%s"]
			  destination_groups    = [nsxt_policy_group.group2.path]
			  services              = [nsxt_policy_service.icmp.path]
			  scope                 = [nsxt_policy_tier1_gateway.gwt1test.path]
			  action                = "ALLOW"
			}
}`, name, destIP, destCidr, destIPRange, sourceIP, sourceCidr, sourceIPRange)
}
