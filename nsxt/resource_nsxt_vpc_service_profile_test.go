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

var accTestPolicyVpcServiceProfileCreateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform created",
	"ntp_servers":      "5.5.5.5",
	"lease_time":       "50840",
	"dns_server_ips":   "7.7.7.7",
	"server_addresses": "11.11.11.11",
}

var accTestPolicyVpcServiceProfileUpdateAttributes = map[string]string{
	"display_name":     getAccTestResourceName(),
	"description":      "terraform updated",
	"ntp_servers":      "5.5.5.7",
	"lease_time":       "148000",
	"dns_server_ips":   "7.7.7.2",
	"server_addresses": "11.11.11.111",
}

func TestAccResourceNsxtVpcServiceProfile_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_service_profile.test"
	testDataSourceName := "data.nsxt_vpc_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcServiceProfileCheckDestroy(state, accTestPolicyVpcServiceProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcServiceProfileTemplate(true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcServiceProfileExists(accTestPolicyVpcServiceProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyVpcServiceProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyVpcServiceProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_relay_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.ntp_servers.0", accTestPolicyVpcServiceProfileCreateAttributes["ntp_servers"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.lease_time", accTestPolicyVpcServiceProfileCreateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.0.dns_server_ips.0", accTestPolicyVpcServiceProfileCreateAttributes["dns_server_ips"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				// add profiles
				Config: testAccNsxtVpcServiceProfileTemplate(true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcServiceProfileExists(accTestPolicyVpcServiceProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyVpcServiceProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyVpcServiceProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_relay_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.ntp_servers.0", accTestPolicyVpcServiceProfileCreateAttributes["ntp_servers"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.lease_time", accTestPolicyVpcServiceProfileCreateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.0.dns_server_ips.0", accTestPolicyVpcServiceProfileCreateAttributes["dns_server_ips"]),
					resource.TestCheckResourceAttrSet(testResourceName, "qos_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "spoof_guard_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "ip_discovery_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "mac_discovery_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "security_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				// update values, no profiles
				Config: testAccNsxtVpcServiceProfileTemplate(false, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcServiceProfileExists(accTestPolicyVpcServiceProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyVpcServiceProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyVpcServiceProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_relay_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.ntp_servers.0", accTestPolicyVpcServiceProfileUpdateAttributes["ntp_servers"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.lease_time", accTestPolicyVpcServiceProfileUpdateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_config.0.dns_client_config.0.dns_server_ips.0", accTestPolicyVpcServiceProfileUpdateAttributes["dns_server_ips"]),

					resource.TestCheckResourceAttr(testResourceName, "qos_profile", ""),
					resource.TestCheckResourceAttr(testResourceName, "spoof_guard_profile", ""),
					resource.TestCheckResourceAttr(testResourceName, "ip_discovery_profile", ""),
					resource.TestCheckResourceAttr(testResourceName, "mac_discovery_profile", ""),
					resource.TestCheckResourceAttr(testResourceName, "security_profile", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcServiceProfileDhcpRelay(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcServiceProfileExists(accTestPolicyVpcServiceProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_relay_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_relay_config.0.server_addresses.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcServiceProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcServiceProfileExists(accTestPolicyVpcServiceProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVpcServiceProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcServiceProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcServiceProfileMinimalistic(),
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

func testAccNsxtVpcServiceProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy VpcServiceProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy VpcServiceProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtVpcServiceProfileExists(testAccGetSessionProjectContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy VpcServiceProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcServiceProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_service_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVpcServiceProfileExists(testAccGetSessionProjectContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy VpcServiceProfile %s still exists", displayName)
		}
	}
	return nil
}

var testAccNsxtVpcServiceProfileHelper = getAccTestResourceName()

func testAccNsxtProjectProfiles() string {
	return fmt.Sprintf(`
resource "nsxt_policy_mac_discovery_profile" "test" {
  %s
  display_name = "%s"
}

resource "nsxt_policy_ip_discovery_profile" "test" {
  %s
  display_name = "%s"
}

resource "nsxt_policy_spoof_guard_profile" "test" {
  %s
  display_name = "%s"
}

resource "nsxt_policy_segment_security_profile" "test"{
  %s
  display_name = "%s"
}

resource "nsxt_policy_qos_profile" "test" {
  %s
  display_name = "%s"
}`, testAccNsxtProjectContext(), testAccNsxtVpcServiceProfileHelper, testAccNsxtProjectContext(), testAccNsxtVpcServiceProfileHelper, testAccNsxtProjectContext(), testAccNsxtVpcServiceProfileHelper, testAccNsxtProjectContext(), testAccNsxtVpcServiceProfileHelper, testAccNsxtProjectContext(), testAccNsxtVpcServiceProfileHelper)
}

func testAccNsxtVpcServiceProfileTemplate(createFlow bool, withProfiles bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyVpcServiceProfileCreateAttributes
	} else {
		attrMap = accTestPolicyVpcServiceProfileUpdateAttributes
	}
	profileAssignment := ""
	if withProfiles {
		profileAssignment = `
  qos_profile           = nsxt_policy_qos_profile.test.path
  security_profile      = nsxt_policy_segment_security_profile.test.path
  ip_discovery_profile  = nsxt_policy_ip_discovery_profile.test.path
  spoof_guard_profile   = nsxt_policy_spoof_guard_profile.test.path
  mac_discovery_profile = nsxt_policy_mac_discovery_profile.test.path`
	}
	return testAccNsxtProjectProfiles() + fmt.Sprintf(`
resource "nsxt_vpc_service_profile" "test" {
  %s
  %s
  display_name = "%s"
  description  = "%s"

  dhcp_config {
    dhcp_server_config {
      ntp_servers = ["%s"]
      lease_time = %s

      dns_client_config {
        dns_server_ips = ["%s"]
      }
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_vpc_service_profile" "test" {
  %s
  display_name = "%s"

  depends_on = [nsxt_vpc_service_profile.test]
}`, testAccNsxtProjectContext(), profileAssignment, attrMap["display_name"], attrMap["description"], attrMap["ntp_servers"], attrMap["lease_time"], attrMap["dns_server_ips"], testAccNsxtProjectContext(), attrMap["display_name"])
}

func testAccNsxtVpcServiceProfileDhcpRelay() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_service_profile" "test" {
  %s
  display_name = "%s"
  dhcp_config {
    dhcp_relay_config {
      server_addresses = ["%s"]
    }
  }
  
}`, testAccNsxtProjectContext(), accTestPolicyVpcServiceProfileUpdateAttributes["display_name"], accTestPolicyVpcServiceProfileUpdateAttributes["server_addresses"])
}

func testAccNsxtVpcServiceProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_service_profile" "test" {
  %s
  display_name = "%s"
  dhcp_config {
  }
  
}`, testAccNsxtProjectContext(), accTestPolicyVpcServiceProfileUpdateAttributes["display_name"])
}
