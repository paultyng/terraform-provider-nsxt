/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestVpcSubnetCreateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform created",
	"access_mode":        "Isolated",
	"enabled":            "true",
	"ipv4_subnet_size":   "32",
	"ip_addresses":       "192.168.240.0/26",
	"reserved_ip_ranges": "11.11.50.7",
}

var accTestVpcSubnetUpdateAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform updated",
	"access_mode":        "Isolated",
	"enabled":            "true",
	"ipv4_subnet_size":   "32",
	"ip_addresses":       "192.168.240.0/26",
	"reserved_ip_ranges": "11.11.50.4-11.11.50.6",
}

var accTestVpcSubnetReservedIPRangeAttributes = map[string]string{
	"display_name":       getAccTestResourceName(),
	"description":        "terraform created",
	"access_mode":        "Isolated",
	"enabled":            "true",
	"ipv4_subnet_size":   "32",
	"ip_addresses":       "10.110.1.0/26",
	"reserved_ip_ranges": "10.110.1.3-10.110.1.15",
}

func TestAccResourceNsxtVpcSubnet_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetCheckDestroy(state, accTestVpcSubnetUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestVpcSubnetCreateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "access_mode", accTestVpcSubnetCreateAttributes["access_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.mode", "DHCP_DEACTIVATED"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.0", accTestVpcSubnetUpdateAttributes["ip_addresses"]),
					resource.TestCheckResourceAttr(testResourceName, "access_mode", accTestVpcSubnetUpdateAttributes["access_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.mode", "DHCP_DEACTIVATED"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetMinimalisticIsolated(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVpcSubnet_subnetSize(t *testing.T) {
	testResourceName := "nsxt_vpc_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetCheckDestroy(state, accTestVpcSubnetUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetSizeTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ipv4_subnet_size", accTestVpcSubnetCreateAttributes["ipv4_subnet_size"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.option121.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.other.#", "1"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetSizeTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ipv4_subnet_size", accTestVpcSubnetCreateAttributes["ipv4_subnet_size"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.option121.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.other.#", "1"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetMinimalisticPublic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtVpcSubnet_reservedIPRange(t *testing.T) {
	testResourceName := "nsxt_vpc_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetCheckDestroy(state, accTestVpcSubnetUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetReservedIPRangeTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetExists(accTestVpcSubnetCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetReservedIPRangeAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetReservedIPRangeAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.reserved_ip_ranges.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.reserved_ip_ranges.0", accTestVpcSubnetReservedIPRangeAttributes["reserved_ip_ranges"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.option121.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "dhcp_config.0.dhcp_server_additional_config.0.options.0.other.#", "1"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVpcSubnet_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetMinimalisticIsolated(),
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

func testAccNsxtVpcSubnetExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("VpcSubnet resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("VpcSubnet resource ID not set in resources")
		}

		exists, err := resourceNsxtVpcSubnetExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("VpcSubnet %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcSubnetCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_subnet" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVpcSubnetExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("VpcSubnet %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVpcSubnetTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcSubnetCreateAttributes
	} else {
		attrMap = accTestVpcSubnetUpdateAttributes
	}

	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
  description  = "%s"

  ip_addresses = ["%s"]
  access_mode = "%s"
  dhcp_config {
    mode = "DHCP_DEACTIVATED"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, testAccNsxtPolicyMultitenancyContext(), attrMap["display_name"], attrMap["description"], attrMap["ip_addresses"], attrMap["access_mode"])
}

func testAccNsxtVpcSubnetSizeTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcSubnetCreateAttributes
	} else {
		attrMap = accTestVpcSubnetUpdateAttributes
	}

	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
  description  = "%s"

  ipv4_subnet_size = %s
  access_mode      = "Public"

  dhcp_config {
    mode = "DHCP_SERVER"

    dhcp_server_additional_config {
      options {
        option121 {
	  static_route {
	    network  = "2.1.1.0/24"
	    next_hop = "2.3.1.3"
	  }
        }
        other {
	  code   = "119"
	  values = ["abc", "def"]
        }
      }
    }
  }
}`, testAccNsxtPolicyMultitenancyContext(), attrMap["display_name"], attrMap["description"], attrMap["ipv4_subnet_size"])
}

func testAccNsxtVpcSubnetMinimalisticIsolated() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
  ip_addresses = ["%s"]
  access_mode  = "Isolated"
}`, testAccNsxtPolicyMultitenancyContext(), accTestVpcSubnetUpdateAttributes["display_name"], accTestVpcSubnetUpdateAttributes["ip_addresses"])
}

func testAccNsxtVpcSubnetMinimalisticPublic() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
  access_mode  = "Public"
}`, testAccNsxtPolicyMultitenancyContext(), accTestVpcSubnetUpdateAttributes["display_name"])
}

func testAccNsxtVpcSubnetReservedIPRangeTemplate(createFlow bool) string {
	attrMap := accTestVpcSubnetReservedIPRangeAttributes

	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
  description  = "%s"
  ip_addresses = ["%s"]

  access_mode      = "Public"

  dhcp_config {
    mode = "DHCP_SERVER"

    dhcp_server_additional_config {
      options {
        option121 {
	  static_route {
	    network  = "2.1.1.0/24"
	    next_hop = "2.3.1.3"
	  }
        }
        other {
	  code   = "119"
	  values = ["abc", "def"]
        }
      }
      reserved_ip_ranges = ["%s"]
    }
  }
}`, testAccNsxtPolicyMultitenancyContext(), attrMap["display_name"], attrMap["description"], attrMap["ip_addresses"], attrMap["reserved_ip_ranges"])
}
