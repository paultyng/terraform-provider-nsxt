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

var l2VpnResourceName = getAccTestResourceName()

var accTestPolicyL2VpnSessionCreateAttributes = map[string]string{
	"display_name":     l2VpnResourceName,
	"description":      "terraform created",
	"direction":        "NONE",
	"max_segment_size": "109",
	"local_address":    "10.12.23.45",
	"peer_address":     "10.12.34.56",
	"protocol":         "GRE",
}

var accTestPolicyL2VpnSessionUpdateAttributes = map[string]string{
	"display_name":     l2VpnResourceName,
	"description":      "terraform updated",
	"direction":        "BOTH",
	"max_segment_size": "128",
	"local_address":    "10.23.32.42",
	"peer_address":     "10.11.33.32",
	"protocol":         "GRE",
}

var testAccL2VpnSessionResourceName = "nsxt_policy_l2_vpn_session.test"

func TestAccResourceNsxtPolicyL2VpnSession_basic(t *testing.T) {
	testResourceName := testAccL2VpnSessionResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnSessionCheckDestroy(state, accTestPolicyL2VpnSessionCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnSessionTestTemplate(true, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnSessionExists(accTestPolicyL2VpnSessionCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnSessionCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnSessionCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "transport_tunnels.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "direction", accTestPolicyL2VpnSessionCreateAttributes["direction"]),
					resource.TestCheckResourceAttr(testResourceName, "max_segment_size", accTestPolicyL2VpnSessionCreateAttributes["max_segment_size"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyL2VpnSessionCreateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyL2VpnSessionCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "protocol", accTestPolicyL2VpnSessionCreateAttributes["protocol"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnSessionTestTemplate(false, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnSessionExists(accTestPolicyL2VpnSessionUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnSessionUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnSessionUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "transport_tunnels.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "direction", accTestPolicyL2VpnSessionUpdateAttributes["direction"]),
					resource.TestCheckResourceAttr(testResourceName, "max_segment_size", accTestPolicyL2VpnSessionUpdateAttributes["max_segment_size"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyL2VpnSessionUpdateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyL2VpnSessionUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "protocol", accTestPolicyL2VpnSessionUpdateAttributes["protocol"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnSessionMinimalistic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnSessionExists(accTestPolicyL2VpnSessionCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnSessionCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "transport_tunnels.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "local_address", ""),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", ""),
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

func TestAccResourceNsxtPolicyL2VpnSession_import(t *testing.T) {
	testResourceName := testAccL2VpnSessionResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnSessionCheckDestroy(state, accTestPolicyL2VpnSessionCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnSessionMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyL2VpnSessionImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyL2VpnSession_tier1(t *testing.T) {
	testResourceName := testAccL2VpnSessionResourceName

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnSessionCheckDestroy(state, accTestPolicyL2VpnSessionCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnSessionTestTemplate(true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnSessionExists(accTestPolicyL2VpnSessionCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnSessionCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnSessionCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "transport_tunnels.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "direction", accTestPolicyL2VpnSessionCreateAttributes["direction"]),
					resource.TestCheckResourceAttr(testResourceName, "max_segment_size", accTestPolicyL2VpnSessionCreateAttributes["max_segment_size"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyL2VpnSessionCreateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyL2VpnSessionCreateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "protocol", accTestPolicyL2VpnSessionCreateAttributes["protocol"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnSessionTestTemplate(false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnSessionExists(accTestPolicyL2VpnSessionUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnSessionUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnSessionUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "service_path"),
					resource.TestCheckResourceAttr(testResourceName, "transport_tunnels.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "direction", accTestPolicyL2VpnSessionUpdateAttributes["direction"]),
					resource.TestCheckResourceAttr(testResourceName, "max_segment_size", accTestPolicyL2VpnSessionUpdateAttributes["max_segment_size"]),
					resource.TestCheckResourceAttr(testResourceName, "local_address", accTestPolicyL2VpnSessionUpdateAttributes["local_address"]),
					resource.TestCheckResourceAttr(testResourceName, "peer_address", accTestPolicyL2VpnSessionUpdateAttributes["peer_address"]),
					resource.TestCheckResourceAttr(testResourceName, "protocol", accTestPolicyL2VpnSessionUpdateAttributes["protocol"]),

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

func testAccNSXPolicyL2VpnSessionImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccL2VpnSessionResourceName]
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

func testAccNsxtPolicyL2VpnSessionExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy L2VpnSession resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]
		isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
		if err != nil {
			return err
		}
		if resourceID == "" {
			return fmt.Errorf("Policy L2VpnSession resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyL2VpnSessionExists(isT0, gwID, localeServiceID, serviceID, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy L2VpnSession %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyL2VpnSessionCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_l2_vpn_session" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		servicePath := rs.Primary.Attributes["service_path"]
		isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
		if err != nil {
			return err
		}

		exists, err := resourceNsxtPolicyL2VpnSessionExists(isT0, gwID, localeServiceID, serviceID, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy L2VpnSession %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyL2VpnSessionPreConditionTemplate(isT0 bool, useGatewayPath bool) string {
	gatewayOrLocaleServicePath := getGatewayOrLocaleServicePath(useGatewayPath, isT0)
	var vpnServiceTemplate string
	if isT0 {
		vpnServiceTemplate = fmt.Sprintf(`
resource "nsxt_policy_l2_vpn_service" "test_l2_svc" {
	display_name          = "%s"
	%s
	enable_hub            = true
	mode                  = "SERVER"
	encap_ip_pool         = ["192.168.10.0/24"]
}

resource "nsxt_policy_ipsec_vpn_service" "test_ipsec_svc" {
	display_name          = "%s"
	%s
}
`, l2VpnResourceName, gatewayOrLocaleServicePath, l2VpnResourceName, gatewayOrLocaleServicePath)
	} else {
		vpnServiceTemplate = fmt.Sprintf(`
resource "nsxt_policy_l2_vpn_service" "test_l2_svc" {
	display_name          = "%s"
	%s
	enable_hub            = true
	mode                  = "SERVER"
	encap_ip_pool         = ["192.168.10.0/24"]
}
resource "nsxt_policy_ipsec_vpn_service" "test_ipsec_svc" {
	display_name          = "%s"
	%s
}
`, l2VpnResourceName, gatewayOrLocaleServicePath, l2VpnResourceName, gatewayOrLocaleServicePath)
	}
	return vpnServiceTemplate + fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_ike_profile" "test" {
	display_name          = "%s"
	description           = "Ike profile for ipsec vpn session"
	encryption_algorithms = ["AES_128"]
	digest_algorithms     = ["SHA2_256"]
	dh_groups             = ["GROUP14"]
	ike_version           = "IKE_V2"
}
resource "nsxt_policy_ipsec_vpn_dpd_profile" "test" {
	display_name       = "%s"
	description        = "Terraform provisioned IPSec VPN DPD Profile"
	dpd_probe_mode     = "ON_DEMAND"
	dpd_probe_interval = 1
	enabled            = true
	retry_count        = 8
}

resource "nsxt_policy_ipsec_vpn_local_endpoint" "test" {
	service_path  = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	display_name  = "%s"
	local_address = "20.20.0.20"
}

resource "nsxt_policy_ipsec_vpn_tunnel_profile" "test" {
	display_name          = "%s"
	description           = "Terraform provisioned IPSec VPN Ike Profile"
	df_policy             = "COPY"
	encryption_algorithms = ["AES_GCM_128"]
	dh_groups             = ["GROUP14"]
	sa_life_time          = 7200
}

resource "nsxt_policy_ipsec_vpn_session" "test" {
	display_name               = "%s"
	tunnel_profile_path        = nsxt_policy_ipsec_vpn_tunnel_profile.test.path
	local_endpoint_path        = nsxt_policy_ipsec_vpn_local_endpoint.test.path
	dpd_profile_path           = nsxt_policy_ipsec_vpn_dpd_profile.test.path
	ike_profile_path           = nsxt_policy_ipsec_vpn_ike_profile.test.path
	service_path               = nsxt_policy_ipsec_vpn_service.test_ipsec_svc.path
	vpn_type                   = "RouteBased"
	authentication_mode        = "PSK"
	peer_address               = "18.18.18.19"
	peer_id                    = "18.18.18.19"
	psk                        = "secret1"
	ip_addresses               = ["169.254.152.2"]
	prefix_length              = 24
}
`, l2VpnResourceName, l2VpnResourceName, l2VpnResourceName, l2VpnResourceName, l2VpnResourceName)
}

func testAccNsxtPolicyL2VpnSessionTestTemplate(createFlow bool, isT0 bool, useGatewayPath bool) string {
	return testAccNsxtPolicyGatewayTemplate(isT0) +
		testAccNsxtPolicyL2VpnSessionPreConditionTemplate(isT0, useGatewayPath) +
		testAccNsxtPolicyL2VpnSessionTemplate(createFlow)
}

func testAccNsxtPolicyL2VpnSessionTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyL2VpnSessionCreateAttributes
	} else {
		attrMap = accTestPolicyL2VpnSessionUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_l2_vpn_session" "test" {
	display_name      = "%s"
	description       = "%s"
	service_path      = nsxt_policy_l2_vpn_service.test_l2_svc.path
	transport_tunnels = [nsxt_policy_ipsec_vpn_session.test.path]
	direction         = "%s"
  	max_segment_size  = "%s"
	local_address     = "%s"
	peer_address      = "%s"
	protocol          = "%s"

	tag {
	  scope = "scope1"
	  tag   = "tag1"
	}
}

`, attrMap["display_name"], attrMap["description"], attrMap["direction"], attrMap["max_segment_size"], attrMap["local_address"], attrMap["peer_address"], attrMap["protocol"])
}

func testAccNsxtPolicyL2VpnSessionMinimalistic(useGatewayPath bool) string {
	attrMap := accTestPolicyL2VpnSessionCreateAttributes
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() +
		testAccNsxtPolicyL2VpnSessionPreConditionTemplate(true, useGatewayPath) + fmt.Sprintf(`
resource "nsxt_policy_l2_vpn_session" "test" {
	display_name      = "%s"
	service_path      = nsxt_policy_l2_vpn_service.test_l2_svc.path
	transport_tunnels = [nsxt_policy_ipsec_vpn_session.test.path]
}

`, attrMap["display_name"])
}
