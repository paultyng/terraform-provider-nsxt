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

var accTestPolicyL2VpnServiceCreateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"enable_hub":    "true",
	"mode":          "SERVER",
	"encap_ip_pool": "192.168.12.0/24",
}

var accTestPolicyL2VpnServiceUpdateAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"enable_hub":    "false",
	"mode":          "SERVER",
	"encap_ip_pool": "192.168.13.0/24",
}

var accTestPolicyL2VpnServiceCreateClientModeAttributes = map[string]string{
	"display_name":  getAccTestResourceName(),
	"description":   "terraform created",
	"enable_hub":    "false",
	"mode":          "CLIENT",
	"encap_ip_pool": "192.168.12.0/24",
}

func TestAccResourceNsxtPolicyL2VpnService_basic(t *testing.T) {
	testResourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceMinimalistic(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(false, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceUpdateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceUpdateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceUpdateAttributes["encap_ip_pool"]),
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

func TestAccResourceNsxtPolicyL2VpnService_updateFromlocaleServicePathToGatewayPath(t *testing.T) {
	testResourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(true, false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyL2VpnService_withGateway(t *testing.T) {
	testResourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceMinimalistic(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(true, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(false, false, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceUpdateAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceUpdateAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceUpdateAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyL2VpnService_ClientMode(t *testing.T) {
	testResourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(false, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateClientModeAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "locale_service_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateClientModeAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateClientModeAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateClientModeAttributes["encap_ip_pool"]),
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

func TestAccResourceNsxtPolicyL2VpnService_ClientMode_withGateway(t *testing.T) {
	testResourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.2.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(false, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyL2VpnServiceExists(accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyL2VpnServiceCreateClientModeAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyL2VpnServiceCreateClientModeAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "enable_hub", accTestPolicyL2VpnServiceCreateClientModeAttributes["enable_hub"]),
					resource.TestCheckResourceAttr(testResourceName, "mode", accTestPolicyL2VpnServiceCreateClientModeAttributes["mode"]),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "encap_ip_pool.0", accTestPolicyL2VpnServiceCreateClientModeAttributes["encap_ip_pool"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyL2VpnService_Import(t *testing.T) {
	resourceName := "nsxt_policy_l2_vpn_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyL2VpnServiceCheckDestroy(state, accTestPolicyL2VpnServiceCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyL2VpnServiceTemplate(true, false, false),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNsxtPolicyL2VpnServiceImporterGetID,
			},
		},
	})
}

func testAccNsxtPolicyL2VpnServiceImporterGetID(s *terraform.State) (string, error) {
	resourceName := "nsxt_policy_l2_vpn_service.test"
	rs, ok := s.RootModule().Resources[resourceName]
	if !ok {
		return "", fmt.Errorf("Policy L2VpnService resource %s not found in resources", resourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("Policy L2VpnService resource ID not set in resources")
	}
	localeServicePath := rs.Primary.Attributes["locale_service_path"]
	gatewayPath := rs.Primary.Attributes["gateway_path"]
	if gatewayPath == "" && localeServicePath == "" {
		return "", fmt.Errorf("At least one of gateway path and locale service path should be provided for VPN resources")
	}
	if localeServicePath != "" {
		return fmt.Sprintf("%s/l2vpn-services/%s", localeServicePath, resourceID), nil
	}
	return fmt.Sprintf("%s/l2vpn-services/%s", gatewayPath, resourceID), nil
}

func testAccNsxtPolicyL2VpnServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy L2VpnService resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy L2VpnService resource ID not set in resources")
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
		_, err1 := getNsxtPolicyL2VpnServiceByID(connector, gwID, isT0, localeServiceID, resourceID, testAccIsGlobalManager())
		if err1 != nil {
			return fmt.Errorf("Policy L2VpnService %s does not exist", displayName)
		}

		return nil
	}
}

func testAccNsxtPolicyL2VpnServiceCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_l2_vpn_service" {
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

		_, err1 := getNsxtPolicyL2VpnServiceByID(connector, gwID, isT0, localeServiceID, resourceID, testAccIsGlobalManager())
		if err1 == nil {
			return fmt.Errorf("Policy L2VpnService %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyL2VpnServiceTemplate(createFlow bool, clientMode bool, useGatewayPath bool) string {
	var attrMap map[string]string
	if clientMode {
		attrMap = accTestPolicyL2VpnServiceCreateClientModeAttributes
	} else {
		if createFlow {
			attrMap = accTestPolicyL2VpnServiceCreateAttributes
		} else {
			attrMap = accTestPolicyL2VpnServiceUpdateAttributes
		}
	}
	gatewayOrLocaleServicePath := getGatewayOrLocaleServicePath(useGatewayPath, true)
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
	  resource "nsxt_policy_l2_vpn_service" "test" {
		display_name                   = "%s"
		description                    = "%s"
		%s
		enable_hub                     = "%s"
		mode               			   = "%s"
		encap_ip_pool 				   = ["%s"]
		tag {
		  scope = "scope1"
		  tag   = "tag1"
		}
	  }`, attrMap["display_name"], attrMap["description"], gatewayOrLocaleServicePath, attrMap["enable_hub"], attrMap["mode"], attrMap["encap_ip_pool"])
}

func testAccNsxtPolicyL2VpnServiceMinimalistic(useGatewayPath bool) string {
	gatewayOrLocaleServicePath := getGatewayOrLocaleServicePath(useGatewayPath, true)
	return testAccNsxtPolicyTier0WithEdgeClusterForVPN() + fmt.Sprintf(`
	  resource "nsxt_policy_l2_vpn_service" "test" {
		display_name          = "%s"
		%s
	  }`, accTestPolicyL2VpnServiceUpdateAttributes["display_name"], gatewayOrLocaleServicePath)
}
