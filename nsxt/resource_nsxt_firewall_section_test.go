// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtFirewallSection_basic(t *testing.T) {
	sectionName := getAccTestResourceName()
	updateSectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	tags := singleTag
	updatedTags := doubleTags
	tos := ""

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, updateSectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateEmptyTemplate(sectionName, tags, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateEmptyTemplate(updateSectionName, updatedTags, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(updateSectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateSectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_withTos(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	tags := singleTag
	tos := `applied_to {
target_type = "NSGroup"
target_id   = "${nsxt_ns_group.grp1.id}"
}`
	updatedTos := `applied_to {
target_type = "NSGroup"
target_id   = "${nsxt_ns_group.grp1.id}"
}
applied_to {
target_type = "NSGroup"
target_id   = "${nsxt_ns_group.grp2.id}"
}`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateEmptyTemplate(sectionName, tags, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "1"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateEmptyTemplate(sectionName, tags, updatedTos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_withRules(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	ruleName := "rule1.0"
	updatedRuleName := "rule1.1"
	tags := singleTag
	tos := ""
	ruleTos := ""

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateTemplate(sectionName, ruleName, tags, tos, ruleTos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.applied_to.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateTemplate(sectionName, updatedRuleName, tags, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedRuleName),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_withRulesAndTags(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	ruleName := "rule1.0"
	updatedRuleName := "rule1.1"
	tags := singleTag
	updatedTags := doubleTags
	tos := ""

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateTemplate(sectionName, ruleName, tags, tos, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.applied_to.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateTemplate(sectionName, updatedRuleName, updatedTags, tos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedRuleName),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_withRulesAndTos(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	ruleName := "rule1.0"
	updatedRuleName := "rule1.1"
	tags := singleTag
	tos := `applied_to {
  target_type = "NSGroup"
  target_id   = "${nsxt_ns_group.grp1.id}"
}`
	ruleTos := `applied_to {
  target_type = "NSGroup"
  target_id   = "${nsxt_ns_group.grp5.id}"
}`
	updatedTos := `applied_to {
  target_type = "NSGroup"
  target_id   = "${nsxt_ns_group.grp1.id}"
}
applied_to {
  target_type = "NSGroup"
  target_id   = "${nsxt_ns_group.grp2.id}"
}`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateTemplate(sectionName, ruleName, tags, tos, ruleTos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.applied_to.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "1"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateTemplate(sectionName, updatedRuleName, tags, updatedTos),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "true"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", updatedRuleName),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_ordered(t *testing.T) {
	sectionNames := [4]string{getAccTestResourceName(), getAccTestResourceName(), getAccTestResourceName(), getAccTestResourceName()}
	testResourceNames := [4]string{"nsxt_firewall_section.test1", "nsxt_firewall_section.test2", "nsxt_firewall_section.test3", "nsxt_firewall_section.test4"}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			for i := 0; i <= 3; i++ {
				err := testAccNSXFirewallSectionCheckDestroy(state, sectionNames[i])
				if err != nil {
					return err
				}
			}

			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateOrderedTemplate(sectionNames),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionNames[0], testResourceNames[0]),
					resource.TestCheckResourceAttr(testResourceNames[0], "display_name", sectionNames[0]),
					resource.TestCheckResourceAttr(testResourceNames[0], "section_type", "LAYER3"),
					testAccNSXFirewallSectionExists(sectionNames[1], testResourceNames[1]),
					resource.TestCheckResourceAttr(testResourceNames[1], "display_name", sectionNames[1]),
					resource.TestCheckResourceAttr(testResourceNames[1], "section_type", "LAYER3"),
					testAccNSXFirewallSectionExists(sectionNames[2], testResourceNames[2]),
					resource.TestCheckResourceAttr(testResourceNames[2], "display_name", sectionNames[2]),
					resource.TestCheckResourceAttr(testResourceNames[2], "section_type", "LAYER3"),
				),
			},
			{
				Config: testAccNSXFirewallSectionUpdateOrderedTemplate(sectionNames),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionNames[0], testResourceNames[0]),
					resource.TestCheckResourceAttr(testResourceNames[0], "display_name", sectionNames[0]),
					resource.TestCheckResourceAttr(testResourceNames[0], "section_type", "LAYER3"),
					testAccNSXFirewallSectionExists(sectionNames[1], testResourceNames[1]),
					resource.TestCheckResourceAttr(testResourceNames[1], "display_name", sectionNames[1]),
					resource.TestCheckResourceAttr(testResourceNames[1], "section_type", "LAYER3"),
					testAccNSXFirewallSectionExists(sectionNames[2], testResourceNames[2]),
					resource.TestCheckResourceAttr(testResourceNames[2], "display_name", sectionNames[2]),
					resource.TestCheckResourceAttr(testResourceNames[2], "section_type", "LAYER3"),
					testAccNSXFirewallSectionExists(sectionNames[3], testResourceNames[3]),
					resource.TestCheckResourceAttr(testResourceNames[3], "display_name", sectionNames[3]),
					resource.TestCheckResourceAttr(testResourceNames[3], "section_type", "LAYER3"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_edge(t *testing.T) {
	sectionName := getAccTestResourceName()
	edgeClusterName := getEdgeClusterName()
	transportZoneName := getOverlayTransportZoneName()
	ruleName := "test"
	testResourceName := "nsxt_firewall_section.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccTestDeprecated(t)
			testAccNSXVersion(t, "2.4.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeFirewallSectionCreateTemplate(edgeClusterName, transportZoneName, sectionName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.applied_to.#", "1"),
				),
			},
			{
				Config: testAccNSXEdgeFirewallSectionCreateTemplate(edgeClusterName, transportZoneName, sectionName, ruleName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXFirewallSectionExists(sectionName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", sectionName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "section_type", "LAYER3"),
					resource.TestCheckResourceAttr(testResourceName, "stateful", "false"),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "applied_to.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.display_name", ruleName),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.applied_to.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_importBasic(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	tags := singleTag
	tos := string("")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateEmptyTemplate(sectionName, tags, tos),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_importWithRules(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	ruleName := "rule1.0"
	tags := singleTag
	tos := string("")

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateTemplate(sectionName, ruleName, tags, tos, tos),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtFirewallSection_importWithTos(t *testing.T) {
	sectionName := getAccTestResourceName()
	testResourceName := "nsxt_firewall_section.test"
	tags := singleTag
	tos := `applied_to {
target_type = "NSGroup"
target_id = "${nsxt_ns_group.grp1.id}"
}`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXFirewallSectionCheckDestroy(state, sectionName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXFirewallSectionCreateEmptyTemplate(sectionName, tags, tos),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXFirewallSectionExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Firewall Section resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Firewall Section resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.ServicesApi.GetSection(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving firewall section ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if firewall section %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("Firewall Section %s wasn't found", displayName)
	}
}

func testAccNSXFirewallSectionCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_firewall_section" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.ServicesApi.GetSection(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving firewall section ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("Firewall Section %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXFirewallSectionNSGroups() string {
	return `
resource "nsxt_ns_group" "grp1" {
  display_name = "grp1"
}

resource "nsxt_ns_group" "grp2" {
  display_name = "grp2"
}

resource "nsxt_ns_group" "grp3" {
  display_name = "grp3"
}

resource "nsxt_ns_group" "grp4" {
  display_name = "grp4"
}

resource "nsxt_ns_group" "grp5" {
  display_name = "grp5"
}

resource "nsxt_ip_protocol_ns_service" "test" {
  protocol = "6"
}`
}

func testAccNSXFirewallSectionCreateTemplate(name string, ruleName string, tags string, tos string, ruleTos string) string {
	return testAccNSXFirewallSectionNSGroups() + fmt.Sprintf(`
resource "nsxt_firewall_section" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  section_type = "LAYER3"
  stateful     = true
%s
%s

  rule {
    display_name          = "%s"
    description           = "rule1"
    action                = "ALLOW"
    logged                = "true"
    ip_protocol           = "IPV4"
    direction             = "IN"
    destinations_excluded = "false"
    sources_excluded      = "false"
    notes                 = "test rule"
    rule_tag              = "test rule tag"
    disabled              = "false"
    %s

    source {
      target_id   = "${nsxt_ns_group.grp1.id}"
      target_type = "NSGroup"
    }

    source {
      target_id   = "${nsxt_ns_group.grp2.id}"
      target_type = "NSGroup"
    }

    destination {
      target_id   = "${nsxt_ns_group.grp3.id}"
      target_type = "NSGroup"
    }

    destination {
      target_id   = "${nsxt_ns_group.grp4.id}"
      target_type = "NSGroup"
    }

    service {
      target_id   = "${nsxt_ip_protocol_ns_service.test.id}"
      target_type = "NSService"
    }
  }
}`, name, tags, tos, ruleName, ruleTos)
}

func testAccNSXFirewallSectionUpdateTemplate(updatedName string, updatedRuleName string, tags string, tos string) string {
	return testAccNSXFirewallSectionNSGroups() + fmt.Sprintf(`
resource "nsxt_firewall_section" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"
  section_type = "LAYER3"
  stateful     = true
  %s
  %s

  rule {
    display_name = "%s"
    description  = "rule1"
    action       = "ALLOW"
    logged       = "true"
    ip_protocol  = "IPV4"
    direction    = "IN"
    disabled     = "false"
  }

  rule {
    display_name = "rule2"
    description  = "rule2"
    action       = "ALLOW"
    logged       = "true"
    ip_protocol  = "IPV6"
    direction    = "OUT"
  }
}`, updatedName, tags, tos, updatedRuleName)
}

func testAccNSXFirewallSectionCreateEmptyTemplate(name string, tags string, tos string) string {
	return testAccNSXFirewallSectionNSGroups() + fmt.Sprintf(`
resource "nsxt_firewall_section" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  section_type = "LAYER3"
  stateful     = true
%s
%s
}`, name, tags, tos)
}

func testAccNSXFirewallSectionUpdateEmptyTemplate(updatedName string, tags string, tos string) string {
	return testAccNSXFirewallSectionNSGroups() + fmt.Sprintf(`
resource "nsxt_firewall_section" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"
  section_type = "LAYER3"
  stateful     = true
%s
%s
}`, updatedName, tags, tos)
}

func testAccNSXFirewallSectionCreateOrderedTemplate(names [4]string) string {
	return fmt.Sprintf(`
resource "nsxt_firewall_section" "test1" {
  display_name = "%s"
  section_type = "LAYER3"
  stateful     = true
}

resource "nsxt_firewall_section" "test2" {
  display_name  = "%s"
  section_type  = "LAYER3"
  insert_before = "${nsxt_firewall_section.test1.id}"
  stateful      = true

  rule {
    display_name = "test"
    action       = "ALLOW"
    logged       = "true"
    ip_protocol  = "IPV4"
    direction    = "IN"
  }
}

resource "nsxt_firewall_section" "test3" {
  display_name  = "%s"
  section_type  = "LAYER3"
  insert_before = "${nsxt_firewall_section.test2.id}"
  stateful      = true
}

`, names[0], names[1], names[2])
}

func testAccNSXFirewallSectionUpdateOrderedTemplate(names [4]string) string {
	return fmt.Sprintf(`
resource "nsxt_firewall_section" "test1" {
  display_name  = "%s"
  section_type  = "LAYER3"
  insert_before = "${nsxt_firewall_section.test4.id}"
  stateful      = true
}

resource "nsxt_firewall_section" "test2" {
  display_name  = "%s"
  section_type  = "LAYER3"
  insert_before = "${nsxt_firewall_section.test1.id}"
  stateful      = true
}

resource "nsxt_firewall_section" "test3" {
  display_name = "%s"
  section_type = "LAYER3"
  stateful     = true
}

resource "nsxt_firewall_section" "test4" {
  display_name = "%s"
  section_type = "LAYER3"
  stateful     = true
}

`, names[0], names[1], names[2], names[3])
}

func testAccNSXEdgeFirewallSectionCreateTemplate(edgeCluster string, transportZone string, name string, ruleName string) string {
	return fmt.Sprintf(`

data "nsxt_edge_cluster" "ec" {
  display_name = "%s"
}

data "nsxt_transport_zone" "tz" {
  display_name = "%s"
}

resource "nsxt_logical_tier1_router" "test" {
  display_name                = "test"
  edge_cluster_id             = "${data.nsxt_edge_cluster.ec.id}"
  enable_router_advertisement = true
}

resource "nsxt_logical_switch" "test" {
  display_name      = "test"
  transport_zone_id = "${data.nsxt_transport_zone.tz.id}"
}

resource "nsxt_logical_port" "test" {
  display_name      = "test"
  logical_switch_id = "${nsxt_logical_switch.test.id}"
}

resource "nsxt_logical_router_centralized_service_port" "test" {
  display_name                  = "test"
  logical_router_id             = "${nsxt_logical_tier1_router.test.id}"
  linked_logical_switch_port_id = "${nsxt_logical_port.test.id}"
  ip_address                    = "2.2.2.2/24"
}

resource "nsxt_firewall_section" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  section_type = "LAYER3"
  stateful     = false
  applied_to {
    target_type = "LogicalRouter"
    target_id   = "${nsxt_logical_tier1_router.test.id}"
  }

  rule {
    display_name = "%s"
    action       = "ALLOW"
    direction    = "IN"
    applied_to {
      target_id   = "${nsxt_logical_router_centralized_service_port.test.id}"
      target_type = "LogicalRouterPort"
    }
  }
}`, edgeCluster, transportZone, name, ruleName)
}
