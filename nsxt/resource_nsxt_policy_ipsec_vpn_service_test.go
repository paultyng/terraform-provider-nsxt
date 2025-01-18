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

var accTestPolicyIPSecVpnServiceCreateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"enabled":       "true",
	"ha_sync":       "true",
	"ike_log_level": "INFO",
	"sources":       "192.168.10.0/24",
	"destinations":  "192.169.10.0/24",
	"action":        "BYPASS",
}

var accTestPolicyIPSecVpnServiceUpdateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform updated",
	"enabled":       "false",
	"ha_sync":       "true",
	"ike_log_level": "DEBUG",
	"sources":       "192.170.10.0/24",
	"destinations":  "192.171.10.0/24",
	"action":        "BYPASS",
}

func TestAccResourceNsxtPolicyIPSecVpnService_basic(t *testing.T) {
	testResourceName := "nsxt_policy_ipsec_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnServiceCheckDestroy(state, accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceCreateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceCreateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceCreateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceUpdateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceUpdateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceUpdateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceUpdateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceUpdateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnServiceMinimalistic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "0"),
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

func TestAccResourceNsxtPolicyIPSecVpnService_withGateway(t *testing.T) {
	testResourceName := "nsxt_policy_ipsec_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnServiceCheckDestroy(state, accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceCreateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceCreateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceCreateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceUpdateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceUpdateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceUpdateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceUpdateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceUpdateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnServiceMinimalistic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnService_updateFromlocaleServicePathToGatewayPath(t *testing.T) {
	testResourceName := "nsxt_policy_ipsec_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnServiceCheckDestroy(state, accTestPolicyIPSecVpnServiceCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceCreateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceCreateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceCreateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIPSecVpnServiceExists(accTestPolicyIPSecVpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyIPSecVpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyIPSecVpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", accTestPolicyIPSecVpnServiceCreateAttributes["enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "ha_sync", accTestPolicyIPSecVpnServiceCreateAttributes["ha_sync"]),
					resource.TestCheckResourceAttr(testResourceName, "ike_log_level", accTestPolicyIPSecVpnServiceCreateAttributes["ike_log_level"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.sources.0", accTestPolicyIPSecVpnServiceCreateAttributes["sources"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.destinations.0", accTestPolicyIPSecVpnServiceCreateAttributes["destinations"]),
					resource.TestCheckResourceAttr(testResourceName, "bypass_rule.0.action", accTestPolicyIPSecVpnServiceCreateAttributes["action"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIPSecVpnService_Import(t *testing.T) {
	resourceName := "nsxt_policy_ipsec_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIPSecVpnServiceCheckDestroy(state, accTestPolicyIPSecVpnServiceCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPSecVpnServiceTemplate(true, false),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNsxtPolicyIPSecVpnServiceImporterGetID,
			},
		},
	})
}

func testAccNsxtPolicyIPSecVpnServiceImporterGetID(s *terraform.State) (string, error) {
	resourceName := "nsxt_policy_ipsec_vpn_service.test"
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("Policy IPSecVpnService resource %s not found in resources", resourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("Policy IPSecVpnService resource ID not set in resources")
	}
	localeServicePath := rs.Primary.Attributes["locale_service_path"]
	gatewayPath := rs.Primary.Attributes["gateway_path"]
	if gatewayPath == "" && localeServicePath == "" {
		return "", fmt.Errorf("At least one of gateway path and locale service path should be provided for VPN resources")
	}
	if localeServicePath != "" {
		return fmt.Sprintf("%s/ipsec-vpn-services/%s", localeServicePath, resourceID), nil
	}
	return fmt.Sprintf("%s/ipsec-vpn-services/%s", gatewayPath, resourceID), nil
}

func testAccNsxtPolicyIPSecVpnServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy IPSecVpnService resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy IPSecVpnService resource ID not set in resources")
		}

		localeServicePath := rs.Primary.Attributes["locale_service_path"]
		gatewayPath := rs.Primary.Attributes["gateway_path"]
		if gatewayPath == "" && localeServicePath == "" {
			return fmt.Errorf("At least one of gateway path and locale service path should be provided for VPN resources")
		}
		isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
		if localeServiceID == "" {
			isT0, gwID = parseGatewayPolicyPath(gatewayPath)
		}
		if err != nil && gatewayPath == "" {
			return fmt.Errorf("Invalid locale service path %s", localeServicePath)
		}
		_, err1 := getNsxtPolicyIPSecVpnServiceByID(connector, gwID, isT0, localeServiceID, resourceID, testAccIsGlobalManager())
		if err1 != nil {
			return fmt.Errorf("Policy IPSecVpnService %s does not exist", displayName)
		}

		return nil
	}
}

func testAccNsxtPolicyIPSecVpnServiceCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ipsec_vpn_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		localeServicePath := rs.Primary.Attributes["locale_service_path"]
		gatewayPath := rs.Primary.Attributes["gateway_path"]
		isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
		if localeServiceID == "" {
			isT0, gwID = parseGatewayPolicyPath(gatewayPath)
		}
		if err != nil && gatewayPath == "" {
			return nil
		}

		_, err1 := getNsxtPolicyIPSecVpnServiceByID(connector, gwID, isT0, localeServiceID, resourceID, testAccIsGlobalManager())
		if err1 == nil {
			return fmt.Errorf("Policy IPSecVpnService %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIPSecVpnServiceTemplate(createFlow bool, useGatewayPath bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyIPSecVpnServiceCreateAttributes
	} else {
		attrMap = accTestPolicyIPSecVpnServiceUpdateAttributes
	}
	gatewayOrLocaleServicePath := getGatewayOrLocaleServicePath(useGatewayPath, true)
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
resource "nsxt_policy_ipsec_vpn_service" "test" {
	display_name                   = "%s"
	description                    = "%s"
	%s
	enabled                        = "%s"
	ha_sync                        = "%s"
	ike_log_level                  = "%s"
	bypass_rule {
		sources                    = ["%s"]
		destinations               = ["%s"]
		action                     = "%s"
	}

	tag {
	scope = "scope1"
	tag   = "tag1"
	}
   }`, attrMap["display_name"], attrMap["description"], gatewayOrLocaleServicePath, attrMap["enabled"], attrMap["ha_sync"], attrMap["ike_log_level"], attrMap["sources"], attrMap["destinations"], attrMap["action"])
}

func getGatewayOrLocaleServicePath(useGatewayPath bool, isT0 bool) string {
	if useGatewayPath {
		if isT0 {
			return "gateway_path = nsxt_policy_tier0_gateway.test.path"
		}
		return "gateway_path = nsxt_policy_tier1_gateway.test.path"
	}
	if isT0 {
		return "locale_service_path = one(nsxt_policy_tier0_gateway.test.locale_service).path"
	}
	return "locale_service_path = one(nsxt_policy_tier1_gateway.test.locale_service).path"

}

func testAccNsxtPolicyIPSecVpnServiceMinimalistic(useGatewayPath bool) string {
	gatewayOrLocaleServicePath := getGatewayOrLocaleServicePath(useGatewayPath, true)
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
   resource "nsxt_policy_ipsec_vpn_service" "test" {
	 display_name          = "%s"
	 %s
   }`, accTestPolicyIPSecVpnServiceUpdateAttributes["display_name"], gatewayOrLocaleServicePath)
}
