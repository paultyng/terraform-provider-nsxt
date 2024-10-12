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

func TestAccResourceNsxtPolicySecurityPolicyRule_basic(t *testing.T) {
	testAccResourceNsxtPolicySecurityPolicyRuleBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicySecurityPolicyRule_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicySecurityPolicyRuleBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicySecurityPolicyRuleBasic(t *testing.T, withContext bool, preCheck func()) {
	policyResourceName := "nsxt_policy_parent_security_policy.policy1"
	policyName := getAccTestResourceName()
	updatedPolicyName := getAccTestResourceName()
	locked := "true"
	updatedLocked := "false"

	ruleResourceName := "nsxt_policy_security_policy_rule.test"
	appendRuleResourceName := "nsxt_policy_security_policy_rule.test2"

	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	appendRuleName := getAccTestResourceName()
	direction := "IN"
	updatedDirection := "OUT"
	proto := "IPV4"
	updatedProto := "IPV4_IPV6"
	action := "ALLOW"
	updatedAction := "DROP"
	seqNum := "1"
	updatedSeqNum := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccNsxtPolicyParentSecurityPolicyCheckDestroy(state, updatedPolicyName); err != nil {
				return err
			}
			return testAccNsxtPolicySecurityPolicyRuleCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(withContext, policyName, locked) +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", name, action, direction, proto, seqNum, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(policyResourceName, defaultDomain),
					resource.TestCheckResourceAttr(policyResourceName, "display_name", policyName),
					resource.TestCheckResourceAttr(policyResourceName, "locked", locked),

					testAccNsxtPolicySecurityPolicyRuleExists(ruleResourceName),
					resource.TestCheckResourceAttr(ruleResourceName, "display_name", name),
					resource.TestCheckResourceAttr(ruleResourceName, "action", action),
					resource.TestCheckResourceAttr(ruleResourceName, "direction", direction),
					resource.TestCheckResourceAttr(ruleResourceName, "ip_version", proto),
					resource.TestCheckResourceAttr(ruleResourceName, "sequence_number", seqNum),
				),
			},
			{
				// Update Policy and Rule at the same time
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(withContext, updatedPolicyName, updatedLocked) +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", updatedName, updatedAction, updatedDirection, updatedProto, updatedSeqNum, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(policyResourceName, defaultDomain),
					resource.TestCheckResourceAttr(policyResourceName, "display_name", updatedPolicyName),
					resource.TestCheckResourceAttr(policyResourceName, "locked", updatedLocked),

					testAccNsxtPolicySecurityPolicyRuleExists(ruleResourceName),
					resource.TestCheckResourceAttr(ruleResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(ruleResourceName, "action", updatedAction),
					resource.TestCheckResourceAttr(ruleResourceName, "direction", updatedDirection),
					resource.TestCheckResourceAttr(ruleResourceName, "ip_version", updatedProto),
					resource.TestCheckResourceAttr(ruleResourceName, "sequence_number", updatedSeqNum),
				),
			},
			{
				// Update Policy and append another Rule at the same time
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(withContext, policyName, locked) +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", updatedName, updatedAction, updatedDirection, updatedProto, updatedSeqNum, false) +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test2", appendRuleName, action, direction, proto, seqNum, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(policyResourceName, defaultDomain),
					resource.TestCheckResourceAttr(policyResourceName, "display_name", policyName),
					resource.TestCheckResourceAttr(policyResourceName, "locked", locked),

					testAccNsxtPolicySecurityPolicyRuleExists(appendRuleResourceName),
					resource.TestCheckResourceAttr(appendRuleResourceName, "display_name", appendRuleName),
					resource.TestCheckResourceAttr(appendRuleResourceName, "action", action),
					resource.TestCheckResourceAttr(appendRuleResourceName, "direction", direction),
					resource.TestCheckResourceAttr(appendRuleResourceName, "ip_version", proto),
					resource.TestCheckResourceAttr(appendRuleResourceName, "sequence_number", seqNum),

					testAccNsxtPolicySecurityPolicyRuleExists(ruleResourceName),
					resource.TestCheckResourceAttr(ruleResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(ruleResourceName, "action", updatedAction),
					resource.TestCheckResourceAttr(ruleResourceName, "direction", updatedDirection),
					resource.TestCheckResourceAttr(ruleResourceName, "ip_version", updatedProto),
					resource.TestCheckResourceAttr(ruleResourceName, "sequence_number", updatedSeqNum),
				),
			},
			{
				// Update Policy and delete one of its Rule at the same time
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(withContext, updatedPolicyName, updatedLocked) +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", updatedName, updatedAction, updatedDirection, updatedProto, updatedSeqNum, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySecurityPolicyExists(policyResourceName, defaultDomain),
					resource.TestCheckResourceAttr(policyResourceName, "display_name", updatedPolicyName),
					resource.TestCheckResourceAttr(policyResourceName, "locked", updatedLocked),

					func(state *terraform.State) error {
						return testAccNsxtPolicySecurityPolicyRuleCheckDestroy(state, appendRuleName)
					},

					testAccNsxtPolicySecurityPolicyRuleExists(ruleResourceName),
					resource.TestCheckResourceAttr(ruleResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(ruleResourceName, "action", updatedAction),
					resource.TestCheckResourceAttr(ruleResourceName, "direction", updatedDirection),
					resource.TestCheckResourceAttr(ruleResourceName, "ip_version", updatedProto),
					resource.TestCheckResourceAttr(ruleResourceName, "sequence_number", updatedSeqNum),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySecurityPolicyRule_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_security_policy_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyRuleCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(false, "policyName", "true") +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", name, "ALLOW", "IN", "IPV4", "1", false),
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

func TestAccResourceNsxtPolicySecurityPolicyRule_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_security_policy_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySecurityPolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySecurityPolicyRuleDeps(true, "policyName", "true") +
					testAccNsxtPolicySecurityPolicyRuleTemplate("test", name, "ALLOW", "IN", "IPV4", "1", false),
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

func testAccNsxtPolicySecurityPolicyRuleExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy SecurityPolicyRule resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy SecurityPolicyRule resource ID not set in resources")
		}

		policyPath := rs.Primary.Attributes["policy_path"]
		exists, err := resourceNsxtPolicySecurityPolicyRuleExists(testAccGetSessionContext(), resourceID, policyPath, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy SecurityPolicyRule ID %s under SecurityPolicy %s", resourceID, policyPath)
		}
		return nil
	}
}

func testAccNsxtPolicySecurityPolicyRuleCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_security_policy_rule" {
			continue
		}

		if rs.Primary.Attributes["display_name"] != displayName {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		policyPath := rs.Primary.Attributes["policy_path"]
		exists, err := resourceNsxtPolicySecurityPolicyRuleExists(testAccGetSessionContext(), resourceID, policyPath, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy SecurityPolicyRule %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicySecurityPolicyRuleDeps(withContext bool, displayName, locked string) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_parent_security_policy" "policy1" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  category        = "Application"
  locked          = %s
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }
}`, context, displayName, locked)
}

func testAccNsxtPolicySecurityPolicyRuleTemplate(resourceName, displayName, action, direction, ipVersion, seqNum string, withID bool) string {
	idString := ""
	if withID {
		idString = fmt.Sprintf(`nsx_id = "%s"`, displayName)
	}
	return fmt.Sprintf(`
resource "nsxt_policy_security_policy_rule" "%s" {
%s
  display_name    = "%s"
  policy_path     = nsxt_policy_parent_security_policy.policy1.path
  action          = "%s"
  direction       = "%s"
  ip_version      = "%s"
  sequence_number = %s

  tag {
    scope = "color"
    tag   = "orange"
  }
}`, resourceName, idString, displayName, action, direction, ipVersion, seqNum)
}
