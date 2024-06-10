/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO: replace with profile data source when available
var policyDefaultIdsProfilePath = "/infra/settings/firewall/security/intrusion-services/profiles/DefaultIDSProfile"

func TestAccResourceNsxtPolicyIntrusionServicePolicy_basic(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.1.0")
	})
}

func TestAccResourceNsxtPolicyIntrusionServicePolicy_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyIntrusionServicePolicy_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyIntrusionServicePolicyBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_intrusion_service_policy.test"
	comments1 := "Acceptance test create"
	comments2 := "Acceptance test update"
	direction1 := "IN"
	direction2 := "OUT"
	proto1 := "IPV4"
	proto2 := "IPV4_IPV6"
	defaultAction := "DETECT"
	tag1 := "abc"
	tag2 := "def"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIntrusionServicePolicyCheckDestroy(state, updatedName, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyBasic(name, comments1, defaultDomain, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments1),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyBasic(updatedName, comments2, defaultDomain, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", comments2),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyWithRule(updatedName, direction1, proto1, tag1, defaultDomain, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag1),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ids_profiles.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyWithRule(updatedName, direction2, proto2, tag2, defaultDomain, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", direction2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", proto2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.log_label", tag2),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ids_profiles.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIntrusionServicePolicy_withDependencies(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_intrusion_service_policy.test"
	defaultAction := "DETECT"
	defaultDirection := "IN_OUT"
	defaultProtocol := "IPV4_IPV6"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIntrusionServicePolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyWithDepsCreate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "false"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.source_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destination_groups.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources_excluded", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations_excluded", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.services.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ids_profiles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.display_name", "rule2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.action", defaultAction),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.source_groups.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destination_groups.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.sources_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.destinations_excluded", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.disabled", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.services.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "rule.1.ids_profiles.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyWithDepsUpdate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServicePolicyExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "comments", ""),
					resource.TestCheckResourceAttr(testResourceName, "locked", "true"),
					resource.TestCheckResourceAttr(testResourceName, "sequence_number", "3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", "rule1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.direction", defaultDirection),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.ip_version", defaultProtocol),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", "DETECT_PREVENT"),
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

func TestAccResourceNsxtPolicyIntrusionServicePolicy_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyImportBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.1.0")
	})
}

func TestAccResourceNsxtPolicyIntrusionServicePolicy_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyIntrusionServicePolicy_importBasic_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServicePolicyImportBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyIntrusionServicePolicyImportBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_intrusion_service_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIntrusionServicePolicyCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServicePolicyBasic(name, "import", defaultDomain, withContext),
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

func testAccNsxtPolicyIntrusionServicePolicyExists(resourceName string, domainName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIntrusionServicePolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy resource ID %s", resourceID)
		}
		return nil
	}
}

func testAccNsxtPolicyIntrusionServicePolicyCheckDestroy(state *terraform.State, displayName string, domainName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_intrusion_service_policy" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyIntrusionServicePolicyExistsInDomain(testAccGetSessionContext(), resourceID, domainName, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy resource %s still exists", displayName)
		}
	}
	return nil
}

// This resource is not supported on GM yet, non-default domain here is for future use
func testAccNsxtPolicyIntrusionServicePolicyBasic(name string, comments string, domainName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	if domainName == defaultDomain {
		return fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_policy" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = true

  tag {
    scope = "color"
    tag   = "orange"
  }

}`, context, name, comments)
	}
	return testAccNsxtGlobalPolicySite(domainName) + fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_policy" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  comments        = "%s"
  locked          = true
  sequence_number = 3
  stateful        = true
  domain          = data.nsxt_policy_site.test.id

  tag {
    scope = "color"
    tag   = "orange"
  }

}`, context, name, comments)
}

func testAccNsxtPolicyIntrusionServicePolicyWithRule(name string, direction string, protocol string, ruleTag string, domainName string, withContext bool) string {
	context := ""
	profile := ""
	profilePath := fmt.Sprintf("\"%s\"", policyDefaultIdsProfilePath)
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		profile = testAccNsxtPolicyIntrusionServiceProfileMinimalistic(name, withContext)
		profilePath = "nsxt_policy_intrusion_service_profile.test.path"
	}
	if domainName == defaultDomain {
		return fmt.Sprintf(`
%s
resource "nsxt_policy_intrusion_service_policy" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  locked          = false
  sequence_number = 3
  stateful        = true

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name = "%s"
    direction    = "%s"
    ip_version   = "%s"
    log_label    = "%s"
    ids_profiles = [%s]

    tag {
      scope = "color"
      tag   = "blue"
    }
  }
}`, profile, context, name, name, direction, protocol, ruleTag, profilePath)
	}
	return testAccNsxtGlobalPolicyGroupIPAddressCreateTemplate("group", domainName) + fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_policy" "test" {
%s
  display_name    = "%s"
  description     = "Acceptance Test"
  locked          = false
  sequence_number = 3
  stateful        = true
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
    ids_profiles = ["%s"]

    tag {
      scope = "color"
      tag   = "blue"
    }
  }
}`, context, name, name, direction, protocol, ruleTag, policyDefaultIdsProfilePath)
}

func testAccNsxtPolicyIntrusionServicePolicyDeps() string {
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

func testAccNsxtPolicyIntrusionServicePolicyWithDepsCreate(name string) string {
	return testAccNsxtPolicyIntrusionServicePolicyDeps() + fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  locked          = false
  sequence_number = 3
  stateful        = true

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name          = "rule1"
    source_groups         = [nsxt_policy_group.group1.path]
    destination_groups    = [nsxt_policy_group.group2.path]
    sources_excluded      = true
    destinations_excluded = true
    services              = [nsxt_policy_service.icmp.path, nsxt_policy_service.tcp778.path]
    ids_profiles          = ["%s"]
  }

  rule {
    display_name          = "rule2"
    source_groups         = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    sources_excluded      = false
    destinations_excluded = false
    ids_profiles          = ["%s"]
  }
}`, name, policyDefaultIdsProfilePath, policyDefaultIdsProfilePath)
}

func testAccNsxtPolicyIntrusionServicePolicyWithDepsUpdate(name string) string {
	return testAccNsxtPolicyIntrusionServicePolicyDeps() + fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_policy" "test" {
  display_name    = "%s"
  description     = "Acceptance Test"
  locked          = true
  sequence_number = 3
  stateful        = true

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
    action                = "DETECT_PREVENT"
    logged                = true
    ids_profiles          = ["%s"]
  }

}`, name, policyDefaultIdsProfilePath)
}
