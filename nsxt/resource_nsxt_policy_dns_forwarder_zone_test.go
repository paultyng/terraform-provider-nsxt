/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDNSForwarderZoneCreateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform created",
	"dns_domain_names": "email.mydomain.com",
	"source_ip":        "10.2.2.3",
	"upstream_servers": "92.68.2.15",
}

var accTestPolicyDNSForwarderZoneUpdateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform updated",
	"dns_domain_names": "stuff.mydomain.ca",
	"source_ip":        "20.0.0.1",
	"upstream_servers": "92.68.2.17",
}

func TestAccResourceNsxtPolicyDNSForwarderZone_basic(t *testing.T) {
	testAccResourceNsxtPolicyDNSForwarderZoneBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyDNSForwarderZone_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDNSForwarderZoneBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDNSForwarderZone_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDNSForwarderZoneBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyDNSForwarderZoneBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_dns_forwarder_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDNSForwarderZoneCheckDestroy(state, accTestPolicyDNSForwarderZoneUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDNSForwarderZoneTemplate(true, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDNSForwarderZoneExists(accTestPolicyDNSForwarderZoneCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDNSForwarderZoneCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDNSForwarderZoneCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dns_domain_names.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_domain_names.0", accTestPolicyDNSForwarderZoneCreateAttributes["dns_domain_names"]),
					resource.TestCheckResourceAttr(testResourceName, "source_ip", accTestPolicyDNSForwarderZoneCreateAttributes["source_ip"]),
					resource.TestCheckResourceAttr(testResourceName, "upstream_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "upstream_servers.0", accTestPolicyDNSForwarderZoneCreateAttributes["upstream_servers"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDNSForwarderZoneTemplate(false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDNSForwarderZoneExists(accTestPolicyDNSForwarderZoneUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDNSForwarderZoneUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDNSForwarderZoneUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dns_domain_names.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_domain_names.0", accTestPolicyDNSForwarderZoneUpdateAttributes["dns_domain_names"]),
					resource.TestCheckResourceAttr(testResourceName, "source_ip", accTestPolicyDNSForwarderZoneUpdateAttributes["source_ip"]),
					resource.TestCheckResourceAttr(testResourceName, "upstream_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "upstream_servers.0", accTestPolicyDNSForwarderZoneUpdateAttributes["upstream_servers"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDNSForwarderZoneMinimalistic(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDNSForwarderZoneExists(accTestPolicyDNSForwarderZoneCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "dns_domain_names.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "source_ip", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDNSForwarderZone_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dns_forwarder_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDNSForwarderZoneCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDNSForwarderZoneMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyDNSForwarderZone_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dns_forwarder_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDNSForwarderZoneCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDNSForwarderZoneMinimalistic(true),
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

func TestAccResourceNsxtPolicyDNSForwarderZone_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_dns_forwarder_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDNSForwarderZoneCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDNSForwarderZoneMinimalistic(false),
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

func testAccNsxtPolicyDNSForwarderZoneExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy PolicyDNSForwarderZone resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy PolicyDNSForwarderZone resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDNSForwarderZoneExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy PolicyDNSForwarderZone %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDNSForwarderZoneCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_dns_forwarder_zone" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyDNSForwarderZoneExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy PolicyDNSForwarderZone %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyDNSForwarderZoneTemplate(createFlow, withContext bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDNSForwarderZoneCreateAttributes
	} else {
		attrMap = accTestPolicyDNSForwarderZoneUpdateAttributes
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_dns_forwarder_zone" "test" {
%s
  display_name     = "%s"
  description      = "%s"
  dns_domain_names = ["%s"]
  source_ip        = "%s"
  upstream_servers = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, attrMap["display_name"], attrMap["description"], attrMap["dns_domain_names"], attrMap["source_ip"], attrMap["upstream_servers"])
}

func testAccNsxtPolicyDNSForwarderZoneMinimalistic(withContext bool) string {
	attrMap := accTestPolicyDNSForwarderZoneUpdateAttributes
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_dns_forwarder_zone" "test" {
%s
  display_name     = "%s"
  upstream_servers = ["%s"]
}`, context, attrMap["display_name"], attrMap["upstream_servers"])
}
