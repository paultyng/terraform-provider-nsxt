/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestPolicyDhcpV6StaticBindingCreateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform created",
	"ip_addresses":    "1001::2",
	"lease_time":      "3600",
	"preferred_time":  "360",
	"mac_address":     "10:0e:00:11:22:02",
	"dns_nameservers": "2002::157",
	"domain_names":    "aaa.com",
	"sntp_servers":    "2002::158",
}

var accTestPolicyDhcpV6StaticBindingUpdateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform updated",
	"lease_time":      "500",
	"preferred_time":  "380",
	"ip_addresses":    "1001::12",
	"mac_address":     "10:ff:22:11:cc:02",
	"dns_nameservers": "2003::157",
	"domain_names":    "bbb.com",
	"sntp_servers":    "2003::158",
}

var testAccPolicyDhcpV6StaticBindingResourceName = "nsxt_policy_dhcp_v6_static_binding.test"

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_basic(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, false, false, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyLocalManagerForFixedSegments(t, false)
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, false, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, false, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_fixed(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, true, false, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyLocalManagerForFixedSegments(t, true)
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_fixed_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, true, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_fixed_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t, true, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyDhcpV6StaticBindingBasic(t *testing.T, isFixed, withContext bool, preCheck func()) {
	testResourceName := testAccPolicyDhcpV6StaticBindingResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV6StaticBindingCheckDestroy(state, accTestPolicyDhcpV6StaticBindingUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingTemplate(isFixed, true, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV6StaticBindingExists(accTestPolicyDhcpV6StaticBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "segment_path"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpV6StaticBindingCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpV6StaticBindingCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyDhcpV6StaticBindingCreateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestPolicyDhcpV6StaticBindingCreateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "preferred_time", accTestPolicyDhcpV6StaticBindingCreateAttributes["preferred_time"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestPolicyDhcpV6StaticBindingCreateAttributes["mac_address"]),
					resource.TestCheckResourceAttr(testResourceName, "domain_names.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "domain_names.0", accTestPolicyDhcpV6StaticBindingCreateAttributes["domain_names"]),
					resource.TestCheckResourceAttr(testResourceName, "dns_nameservers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_nameservers.0", accTestPolicyDhcpV6StaticBindingCreateAttributes["dns_nameservers"]),
					resource.TestCheckResourceAttr(testResourceName, "sntp_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sntp_servers.0", accTestPolicyDhcpV6StaticBindingCreateAttributes["sntp_servers"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingTemplate(isFixed, false, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV6StaticBindingExists(accTestPolicyDhcpV6StaticBindingUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "segment_path"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyDhcpV6StaticBindingUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyDhcpV6StaticBindingUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyDhcpV6StaticBindingUpdateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestPolicyDhcpV6StaticBindingUpdateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "preferred_time", accTestPolicyDhcpV6StaticBindingUpdateAttributes["preferred_time"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestPolicyDhcpV6StaticBindingUpdateAttributes["mac_address"]),
					resource.TestCheckResourceAttr(testResourceName, "domain_names.0", accTestPolicyDhcpV6StaticBindingUpdateAttributes["domain_names"]),
					resource.TestCheckResourceAttr(testResourceName, "dns_nameservers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dns_nameservers.0", accTestPolicyDhcpV6StaticBindingUpdateAttributes["dns_nameservers"]),
					resource.TestCheckResourceAttr(testResourceName, "sntp_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "sntp_servers.0", accTestPolicyDhcpV6StaticBindingUpdateAttributes["sntp_servers"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingMinimalistic(isFixed, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyDhcpV6StaticBindingExists(accTestPolicyDhcpV6StaticBindingCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "segment_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_importBasic(t *testing.T) {
	name := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV6StaticBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingMinimalistic(false, false),
			},
			{
				ResourceName:      testAccPolicyDhcpV6StaticBindingResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyDhcpV6StaticBindingImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV6StaticBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingMinimalistic(false, true),
			},
			{
				ResourceName:      testAccPolicyDhcpV6StaticBindingResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testAccPolicyDhcpV6StaticBindingResourceName),
			},
		},
	})
}

func TestAccResourceNsxtPolicyDhcpV6StaticBinding_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDhcpV6StaticBindingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyDhcpV6StaticBindingMinimalistic(false, false),
			},
			{
				ResourceName:      testAccPolicyDhcpV6StaticBindingResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testAccPolicyDhcpV6StaticBindingResourceName),
			},
		},
	})
}

func testAccNsxtPolicyDhcpV6StaticBindingExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DhcpV6StaticBinding resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		segmentPath := rs.Primary.Attributes["segment_path"]
		if resourceID == "" {
			return fmt.Errorf("Policy DhcpV6StaticBinding resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(testAccGetSessionContext(), resourceID, segmentPath, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy DhcpV6StaticBinding %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyDhcpV6StaticBindingCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_dhcp_v6_static_binding" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		segmentPath := rs.Primary.Attributes["segment_path"]
		exists, err := resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(testAccGetSessionContext(), resourceID, segmentPath, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DhcpV6StaticBinding %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyDhcpV6StaticBindingImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccPolicyDhcpV6StaticBindingResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding resource %s not found in resources", testAccPolicyDhcpV6StaticBindingResourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding resource ID not set in resources ")
	}
	segmentPath := rs.Primary.Attributes["segment_path"]
	if segmentPath == "" {
		return "", fmt.Errorf("NSX Policy Dhcp Static Binding Segment Path not set in resources ")
	}
	segs := strings.Split(segmentPath, "/")
	return fmt.Sprintf("%s/%s", segs[len(segs)-1], resourceID), nil
}

func testAccNsxtPolicyDhcpV6StaticBindingTemplate(isFixed, createFlow, withContext bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyDhcpV6StaticBindingCreateAttributes
	} else {
		attrMap = accTestPolicyDhcpV6StaticBindingUpdateAttributes
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyDhcpStaticBindingPrerequisites(isFixed, true, withContext) + fmt.Sprintf(`

resource "nsxt_policy_dhcp_v6_static_binding" "test" {
%s
  nsx_id          = "terraform-test"
  segment_path    = %s.test.path
  display_name    = "%s"
  description     = "%s"
  ip_addresses    = ["%s"]
  lease_time      = %s
  preferred_time  = %s
  mac_address     = "%s"
  dns_nameservers = ["%s"]
  domain_names    = ["%s"]
  sntp_servers    = ["%s"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, testAccNsxtPolicyGetSegmentResourceName(isFixed), attrMap["display_name"], attrMap["description"], attrMap["ip_addresses"], attrMap["lease_time"], attrMap["preferred_time"], attrMap["mac_address"], attrMap["dns_nameservers"], attrMap["domain_names"], attrMap["sntp_servers"])
}

func testAccNsxtPolicyDhcpV6StaticBindingMinimalistic(isFixed, withContext bool) string {
	attrMap := accTestPolicyDhcpV6StaticBindingUpdateAttributes
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicyDhcpStaticBindingPrerequisites(isFixed, true, withContext) + fmt.Sprintf(`
resource "nsxt_policy_dhcp_v6_static_binding" "test" {
%s
  segment_path    = %s.test.path
  display_name = "%s"
  mac_address  = "%s"
}`, context, testAccNsxtPolicyGetSegmentResourceName(isFixed), attrMap["display_name"], attrMap["mac_address"])
}
