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

var accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform created",
	"gateway_address": "192.168.240.1",
	"host_name":       "test-create",
	"mac_address":     "10:0e:00:11:22:02",
	"lease_time":      "162",
	"ip_address":      "192.168.240.41",
	"opt121-1-gw":     "192.168.240.6",
	"opt121-1-net":    "192.168.244.0/24",
	"opt121-2-gw":     "192.168.240.7",
	"opt121-2-net":    "192.168.245.0/24",
	"opt-other-code":  "35",
	"opt-other-value": "50",
}

var accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes = map[string]string{
	"display_name":    getAccTestResourceName(),
	"description":     "terraform updated",
	"gateway_address": "192.168.240.2",
	"host_name":       "test-update",
	"mac_address":     "10:ff:22:11:cc:02",
	"lease_time":      "500",
	"ip_address":      "192.168.240.42",
	"opt121-1-gw":     "192.168.240.8",
	"opt121-1-net":    "192.168.244.0/24",
	"opt121-2-gw":     "192.168.240.9",
	"opt121-2-net":    "192.168.245.0/24",
	"opt-other-code":  "35",
	"opt-other-value": "45",
}

func TestAccResourceNsxtVpcSubnetDhcpV4StaticBindingConfig_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_dhcp_v4_static_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetDhcpV4StaticBindingConfigCheckDestroy(state, accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetDhcpV4StaticBindingConfigTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetDhcpV4StaticBindingConfigExists(accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_address", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["gateway_address"]),
					resource.TestCheckResourceAttr(testResourceName, "host_name", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["host_name"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["mac_address"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.0.next_hop", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt121-1-gw"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.0.network", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt121-1-net"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.1.next_hop", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt121-2-gw"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.1.network", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt121-2-net"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.0.code", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt-other-code"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.0.values.0", accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["opt-other-value"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetDhcpV4StaticBindingConfigTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetDhcpV4StaticBindingConfigExists(accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "gateway_address", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["gateway_address"]),
					resource.TestCheckResourceAttr(testResourceName, "host_name", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["host_name"]),
					resource.TestCheckResourceAttr(testResourceName, "mac_address", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["mac_address"]),
					resource.TestCheckResourceAttr(testResourceName, "lease_time", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["lease_time"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.0.next_hop", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt121-1-gw"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.0.network", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt121-1-net"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.1.next_hop", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt121-2-gw"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.option121.0.static_route.1.network", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt121-2-net"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.0.code", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt-other-code"]),
					resource.TestCheckResourceAttr(testResourceName, "options.0.other.0.values.0", accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["opt-other-value"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcSubnetDhcpV4StaticBindingConfigMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcSubnetDhcpV4StaticBindingConfigExists(accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "parent_path"),
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

func TestAccResourceNsxtVpcSubnetDhcpV4StaticBindingConfig_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_dhcp_v4_static_binding.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcSubnetDhcpV4StaticBindingConfigCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetDhcpV4StaticBindingConfigMinimalistic(),
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

func testAccNsxtVpcSubnetDhcpV4StaticBindingConfigExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy DhcpV4StaticBindingConfig resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy DhcpV4StaticBindingConfig resource ID not set in resources")
		}

		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtVpcSubnetDhcpV4StaticBindingConfigExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy DhcpV4StaticBindingConfig %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcSubnetDhcpV4StaticBindingConfigCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_dhcp_v4_static_binding" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		parentPath := rs.Primary.Attributes["parent_path"]

		exists, err := resourceNsxtVpcSubnetDhcpV4StaticBindingConfigExists(testAccGetSessionContext(), parentPath, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy DhcpV4StaticBindingConfig %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVpcSubnetDhcpV4StaticBindingConfigPrerequisites() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_subnet" "test" {
%s
  display_name = "test-subnet"
  ip_addresses = ["192.168.240.0/26"]
  access_mode = "Isolated"
  dhcp_config {
    mode = "DHCP_SERVER"
    dhcp_server_additional_config {
      reserved_ip_ranges = ["192.168.240.40-192.168.240.60"]
    }
  }
}
`, testAccNsxtPolicyMultitenancyContext())
}

func testAccNsxtVpcSubnetDhcpV4StaticBindingConfigTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes
	} else {
		attrMap = accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes
	}
	return testAccNsxtVpcSubnetDhcpV4StaticBindingConfigPrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_dhcp_v4_static_binding" "test" {
  parent_path = nsxt_vpc_subnet.test.path
  display_name = "%s"
  description  = "%s"
  gateway_address = "%s"
  host_name = "%s"
  mac_address = "%s"
  lease_time = %s
  ip_address = "%s"

  options {
    option121 {
      static_route {
        next_hop = "%s"
        network = "%s"
      }
      static_route {
        next_hop = "%s"
        network = "%s"
      }
    }
    other {
      code = %s
      values = ["%s"]
    }
  }
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["gateway_address"], attrMap["host_name"],
		attrMap["mac_address"], attrMap["lease_time"], attrMap["ip_address"], attrMap["opt121-1-gw"], attrMap["opt121-1-net"],
		attrMap["opt121-2-gw"], attrMap["opt121-2-net"], attrMap["opt-other-code"], attrMap["opt-other-value"])
}

func testAccNsxtVpcSubnetDhcpV4StaticBindingConfigMinimalistic() string {
	attrMap := accTestVpcSubnetDhcpV4StaticBindingConfigCreateAttributes
	return testAccNsxtVpcSubnetDhcpV4StaticBindingConfigPrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc_dhcp_v4_static_binding" "test" {
  display_name = "%s"
  parent_path = nsxt_vpc_subnet.test.path
  ip_address   = "%s"
  mac_address  = "%s"
}`, accTestVpcSubnetDhcpV4StaticBindingConfigUpdateAttributes["display_name"], attrMap["ip_address"], attrMap["mac_address"])
}
