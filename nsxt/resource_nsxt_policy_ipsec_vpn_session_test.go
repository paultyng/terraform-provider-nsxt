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

var ipsecVpnResourceName = getAccTestResourceName()

var accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes = map[string]string{
	"display_name":               ipsecVpnResourceName,
	"description":                "terraform created",
	"enabled":                    "true",
	"vpn_type":                   "RouteBased",
	"authentication_mode":        "PSK",
	"compliance_suite":           "NONE",
	"ip_addresses":               "169.254.152.25",
	"prefix_length":              "24",
	"peer_address":               "18.18.18.22",
	"peer_id":                    "18.18.18.22",
	"psk":                        "secret1",
	"connection_initiation_mode": "RESPOND_ONLY",
}

var accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes = map[string]string{
	"display_name":               ipsecVpnResourceName,
	"description":                "terraform updated",
	"enabled":                    "false",
	"vpn_type":                   "RouteBased",
	"authentication_mode":        "PSK",
	"compliance_suite":           "NONE",
	"ip_addresses":               "169.254.152.26",
	"prefix_length":              "24",
	"peer_address":               "18.18.18.21",
	"peer_id":                    "18.18.18.21",
	"psk":                        "secret2",
	"connection_initiation_mode": "ON_DEMAND",
}

var accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes = map[string]string{
	"display_name":               ipsecVpnResourceName,
	"description":                "Terraform-provisioned IPsec Route-Based VPN",
	"enabled":                    "true",
	"vpn_type":                   "PolicyBased",
	"authentication_mode":        "PSK",
	"compliance_suite":           "NONE",
	"peer_address":               "18.18.18.21",
	"peer_id":                    "18.18.18.21",
	"psk":                        "VMware123!",
	"connection_initiation_mode": "RESPOND_ONLY",
	"sources":                    "192.170.10.0/24",
	"destinations":               "192.171.10.0/24",
	"action":                     "PROTECT",
}

var accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes = map[string]string{
	"display_name":               getAccTestResourceName(),
	"description":                "terraform updated",
	"enabled":                    "true",
	"vpn_type":                   "PolicyBased",
	"authentication_mode":        "PSK",
	"compliance_suite":           "NONE",
	"peer_address":               "18.18.18.22",
	"peer_id":                    "18.18.18.22",
	"psk":                        "secret1",
	"connection_initiation_mode": "RESPOND_ONLY",
	"sources":                    "192.172.10.0/24",
	"destinations":               "192.173.10.0/24",
	"action":                     "PROTECT",
}

var accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes = map[string]string{
	"display_name":               ipsecVpnResourceName,
	"description":                "Test compliance suite",
	"enabled":                    "true",
	"vpn_type":                   "RouteBased",
	"authentication_mode":        "CERTIFICATE",
	"compliance_suite":           "FIPS",
	"ip_addresses":               "169.254.152.26",
	"prefix_length":              "24",
	"peer_address":               "18.18.18.21",
	"peer_id":                    "18.18.18.21",
	"connection_initiation_mode": "ON_DEMAND",
}

var accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes = map[string]string{
	"display_name":               ipsecVpnResourceName,
	"description":                "Test compliance suite",
	"enabled":                    "true",
	"vpn_type":                   "PolicyBased",
	"authentication_mode":        "CERTIFICATE",
	"compliance_suite":           "FIPS",
	"peer_address":               "18.18.18.21",
	"peer_id":                    "18.18.18.21",
	"connection_initiation_mode": "RESPOND_ONLY",
	"sources":                    "192.170.10.0/24",
	"destinations":               "192.171.10.0/24",
	"action":                     "PROTECT",
}

var testAccIPSecVpnSessionResourceName = "nsxt_policy_ipsec_vpn_session.test"

func TestAccResourceNsxtPolicyIPSecVpnSessionRouteBased_basic(t *testing.T) {

	testResourceName := testAccIPSecVpnSessionResourceName
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplate(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["connection_initiation_mode"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplate(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["connection_initiation_mode"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", "NONE"),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", "INITIATOR"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(true),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionRouteBasedWithComplianceSuite(t *testing.T) {

	testResourceName := testAccIPSecVpnSessionResourceName
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_CERTIFICATE_NAME")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplateWithComplianceSuite(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes["connection_initiation_mode"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(true),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionRouteBased_import(t *testing.T) {
	testResourceName := testAccIPSecVpnSessionResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: testAccNSXPolicyIPSecVpnSessionImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionPolicyBased_basic(t *testing.T) {
	testResourceName := testAccIPSecVpnSessionResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplate(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["connection_initiation_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources.0", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations.0", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplate(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["connection_initiation_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources.0", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations.0", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(true),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionPolicyBasedWithComplianceSuite(t *testing.T) {

	testResourceName := testAccIPSecVpnSessionResourceName
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_CERTIFICATE_NAME")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplateWithComplianceSuite(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["connection_initiation_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources.0", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations.0", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes["action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(true),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionPolicyBased_tier1(t *testing.T) {
	testResourceName := testAccIPSecVpnSessionResourceName
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplate(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["connection_initiation_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources.0", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations.0", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes["action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplate(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["connection_initiation_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.sources.0", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.destinations.0", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "rule.0.action", accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes["action"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(false),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnSessionRouteBased_tier1(t *testing.T) {
	testResourceName := testAccIPSecVpnSessionResourceName
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state, accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplate(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes["connection_initiation_mode"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplate(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnSessionExists(accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "ike_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "tunnel_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "dpd_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "vpn_type", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["vpn_type"]),
					resource.TestCheckResourceAttr(testResourceName, "authentication_mode", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["authentication_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "compliance_suite", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["compliance_suite"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "prefix_length", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["prefix_length"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_id", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["peer_id"]),
					resource.TestCheckResourceAttr(testResourceName, "psk", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["psk"]),
					resource.TestCheckResourceAttr(testResourceName, "connection_initiation_mode", accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes["connection_initiation_mode"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGatewayTemplate(false),
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnSessionExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPSecVpnSession resource %s not found in resources", displayName)
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]
		if resourceID == "" {
			return fmt.Errorf("Policy IPSecVpnSession resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIPSecVpnSessionExists(servicePath, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy IPSecVpnSession %s does not exist", displayName)
		}

		return nil
	}
}

func testAccNsxtPolicyIPSecVpnSessionCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ipsec_vpn_session" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]

		exists, err := resourceNsxtPolicyIPSecVpnSessionExists(servicePath, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy IPSecVpnSession %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyIPSecVpnSessionImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccIPSecVpnSessionResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy VPN session %s not found in resources", testAccIPSecVpnSessionResourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy VPN session ID not set in resources ")
	}
	servicePath := rs.Primary.Attributes["service_path"]
	if servicePath == "" {
		return "", fmt.Errorf("NSX Policy VPN session service_path not set")
	}
	return fmt.Sprintf("%s/sessions/%s", servicePath, resourceID), nil
}

func testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(isT0 bool, useCert bool) string {
	certName := getTestCertificateName(false)
	var localEndpointTemplate string
	if useCert {
		localEndpointTemplate = fmt.Sprintf(`
data "nsxt_policy_certificate" "test" {
	display_name = "%s"
	}

resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
	service_path  = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	display_name  = "%s"
	local_address = "20.20.0.25"
	certificate_path = data.nsxt_policy_certificate.test.path
	trust_ca_paths = [data.nsxt_policy_certificate.test.path]
	}
		`, certName, ipsecVpnResourceName)
	} else {
		localEndpointTemplate = fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
	service_path  = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	display_name  = "%s"
	local_address = "20.20.0.25"
	}
	`, ipsecVpnResourceName)
	}
	var vpnServiceTemplate string
	if isT0 {
		vpnServiceTemplate = fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_service" "test_ipsec_svc" {
	display_name          = "%s"
	locale_service_path   = one(nsxt_policy_tier0_gateway.test.locale_service).path
	}
`, ipsecVpnResourceName)
	} else {
		vpnServiceTemplate = fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_service" "test_ipsec_svc" {
	display_name          = "%s"
	locale_service_path   = one(nsxt_policy_tier1_gateway.test.locale_service).path
	}
`, ipsecVpnResourceName)
	}
	return vpnServiceTemplate + localEndpointTemplate + fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_ike_profile" "test" {
	display_name          = "%s"
	description           = "Ike profile for ipsec vpn session"
	encryption_algorithms = ["AES_128"]
	digest_algorithms     = ["SHA2_256"]
	dh_groups             = ["GROUP14"]
	ike_version           = "IKE_V2"
}

resource "nsxt_policy_ipsec_vpn_tunnel_profile" "test" {
	display_name          = "%s"
	description           = "Terraform provisioned IPSec VPN Ike Profile"
	df_policy             = "COPY"
	encryption_algorithms = ["AES_128"]
	digest_algorithms     = ["SHA2_256"]
	dh_groups             = ["GROUP14"]
	sa_life_time          = 7200
}

resource "nsxt_policy_ipsec_vpn_dpd_profile" "test" {
	display_name       = "%s"
	description        = "Terraform provisioned IPSec VPN DPD Profile"
	dpd_probe_mode     = "ON_DEMAND"
	dpd_probe_interval = 1
	enabled            = true
	retry_count        = 8
  }
`, ipsecVpnResourceName, ipsecVpnResourceName, ipsecVpnResourceName)
}

func testAccNsxtPolicyIPSecVpnSessionRouteBasedMinimalistic() string {
	attrMap := accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() +
		testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(true, false) +
		fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name        = "%s"
	tunnel_profile_path = nsxt_policy_ipsec_vpn_tunnel_profile.test.path
	local_endpoint_path = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	service_path        = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type            = "%s"
	peer_address        = "%s"
	peer_id             = "%s"
	ip_addresses        = ["%s"]
	prefix_length       = %s
	psk                 = "%s"
}`, attrMap["display_name"], attrMap["vpn_type"], attrMap["peer_address"], attrMap["peer_id"], attrMap["ip_addresses"], attrMap["prefix_length"], attrMap["psk"])
}

func testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplate(createFlow bool, isT0 bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnSessionRouteBasedCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnSessionRouteBasedUpdateAttributes
	}
	return testAccNsxtPolicyGatewayTemplate(isT0) + testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(isT0, false) +
		fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name               = "%s"
	description                = "%s"
	ike_profile_path           = nsxt_policy_ipsec_vpn_ike_profile.test.path
	tunnel_profile_path        = nsxt_policy_ipsec_vpn_tunnel_profile.test.path
	dpd_profile_path           = nsxt_policy_ipsec_vpn_dpd_profile.test.path
	local_endpoint_path        = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	enabled                    = "%s"
	service_path               = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type                   = "%s"
	authentication_mode        = "%s"
	compliance_suite           = "%s"
	ip_addresses               = ["%s"]
	prefix_length              = "%s"
	peer_address               = "%s"
	peer_id                    = "%s"
	psk                        = "%s"
	connection_initiation_mode = "%s"

	tag {
	  scope = "scope1"
	  tag   = "tag1"
	}
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["vpn_type"],
			attrMap["authentication_mode"], attrMap["compliance_suite"], attrMap["ip_addresses"], attrMap["prefix_length"], attrMap["peer_address"], attrMap["peer_id"], attrMap["psk"], attrMap["connection_initiation_mode"])
}

func testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplate(createFlow bool, isT0 bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnSessionPolicyBasedCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnSessionPolicyBasedUpdateAttributes
	}
	return testAccNsxtPolicyGatewayTemplate(isT0) + testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(isT0, false) +
		fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name               = "%s"
	description                = "%s"
	ike_profile_path           = nsxt_policy_ipsec_vpn_ike_profile.test.path
	tunnel_profile_path        = nsxt_policy_ipsec_vpn_tunnel_profile.test.path
	dpd_profile_path		   = nsxt_policy_ipsec_vpn_dpd_profile.test.path
	local_endpoint_path		   = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	enabled                    = "%s"
	service_path               = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type                   = "%s"
	authentication_mode        = "%s"
	compliance_suite           = "%s"
	peer_address               = "%s"
	peer_id                    = "%s"
	psk                        = "%s"
	connection_initiation_mode = "%s"

	rule {
		sources             = ["%s"]
		destinations        = ["%s"]
		action              = "%s"
	  }

	tag {
	scope = "scope1"
	tag   = "tag1"
	  }
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["vpn_type"],
			attrMap["authentication_mode"], attrMap["compliance_suite"], attrMap["peer_address"], attrMap["peer_id"],
			attrMap["psk"], attrMap["connection_initiation_mode"], attrMap["sources"], attrMap["destinations"], attrMap["action"])
}

func testAccNsxtPolicyIPSecVpnSessionRouteBasedTemplateWithComplianceSuite(isT0 bool) string {
	attrMap := accTestPolicyIPSecVpnSessionRouteBasedComlianceSuiteAttributes
	return testAccNsxtPolicyGatewayTemplate(isT0) + testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(isT0, true) +
		fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name               = "%s"
	description                = "%s"
	dpd_profile_path		   = nsxt_policy_ipsec_vpn_dpd_profile.test.path
	local_endpoint_path		   = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	enabled                    = "%s"
	service_path               = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type                   = "%s"
	authentication_mode        = "%s"
	compliance_suite           = "%s"
	peer_address               = "%s"
	peer_id                    = "%s"
	connection_initiation_mode = "%s"
	ip_addresses               = ["%s"]
	prefix_length              = "%s"

	tag {
	scope = "scope1"
	tag   = "tag1"
	  }
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["vpn_type"],
			attrMap["authentication_mode"], attrMap["compliance_suite"], attrMap["peer_address"], attrMap["peer_id"],
			attrMap["connection_initiation_mode"], attrMap["ip_addresses"], attrMap["prefix_length"])
}

func testAccNsxtPolicyIPSecVpnSessionPolicyBasedTemplateWithComplianceSuite(isT0 bool) string {
	attrMap := accTestPolicyIPSecVpnSessionPolicyBasedComlianceSuiteAttributes
	return testAccNsxtPolicyGatewayTemplate(isT0) + testAccNsxtPolicyIPSecVpnSessionPreConditionTemplate(isT0, true) +
		fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name               = "%s"
	description                = "%s"
	dpd_profile_path		   = nsxt_policy_ipsec_vpn_dpd_profile.test.path
	local_endpoint_path		   = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	enabled                    = "%s"
	service_path               = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type                   = "%s"
	authentication_mode        = "%s"
	compliance_suite           = "%s"
	peer_address               = "%s"
	peer_id                    = "%s"
	connection_initiation_mode = "%s"

	rule {
		sources             = ["%s"]
		destinations        = ["%s"]
		action              = "%s"
	  }

	tag {
	scope = "scope1"
	tag   = "tag1"
	  }
}`, attrMap["display_name"], attrMap["description"], attrMap["enabled"], attrMap["vpn_type"],
			attrMap["authentication_mode"], attrMap["compliance_suite"], attrMap["peer_address"], attrMap["peer_id"],
			attrMap["connection_initiation_mode"], attrMap["sources"], attrMap["destinations"], attrMap["action"])
}
