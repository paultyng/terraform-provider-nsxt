// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
)

var accTestPolicyLBPoolCreateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform created",
	"algorithm":               "IP_HASH",
	"min_active_members":      "2",
	"tcp_multiplexing_number": "2",
	"active_monitor_path":     "/infra/lb-monitor-profiles/default-icmp-lb-monitor",
}

var accTestPolicyLBPoolUpdateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform updated",
	"algorithm":               "WEIGHTED_ROUND_ROBIN",
	"min_active_members":      "5",
	"tcp_multiplexing_number": "5",
	"active_monitor_paths":    "/infra/lb-monitor-profiles/default-http-lb-monitor",
}

func TestAccResourceNsxtPolicyLBPool_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBPoolCheckDestroy(state, accTestPolicyLBPoolUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBPoolMemberTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", accTestPolicyLBPoolCreateAttributes["algorithm"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.admin_state", "ENABLED"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.display_name", "member1"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.backup_member", "false"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.ip_address", "5.5.5.5"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.max_concurrent_connections", "7"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.port", "77"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.weight", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.1.admin_state", "DISABLED"),
					resource.TestCheckResourceAttr(testResourceName, "member.1.display_name", "member2"),
					resource.TestCheckResourceAttr(testResourceName, "member.1.ip_address", "5.5.5.3"),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", accTestPolicyLBPoolCreateAttributes["min_active_members"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_multiplexing_number", accTestPolicyLBPoolCreateAttributes["tcp_multiplexing_number"]),
					// In the 1st step we use the deprecated string attribute
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_path", "/infra/lb-monitor-profiles/default-icmp-lb-monitor"),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPoolMemberTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", accTestPolicyLBPoolUpdateAttributes["algorithm"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.admin_state", "ENABLED"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.display_name", "member1"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.backup_member", "false"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.ip_address", "5.5.5.5"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.max_concurrent_connections", "100"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.port", "88"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.weight", "1"),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", accTestPolicyLBPoolUpdateAttributes["min_active_members"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_multiplexing_number", accTestPolicyLBPoolUpdateAttributes["tcp_multiplexing_number"]),
					// In the 2nd step we switch to the current list attribute
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_paths.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_paths.0", "/infra/lb-monitor-profiles/default-http-lb-monitor"),
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_path", ""),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPoolGroupTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", accTestPolicyLBPoolUpdateAttributes["algorithm"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "member_group.0.group_path"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.allow_ipv4", "false"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.allow_ipv6", "true"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.max_ip_list_size", "10"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.port", "888"),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", accTestPolicyLBPoolUpdateAttributes["min_active_members"]),
					resource.TestCheckResourceAttr(testResourceName, "tcp_multiplexing_number", accTestPolicyLBPoolUpdateAttributes["tcp_multiplexing_number"]),
					// This step clears active_monitor_paths
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_paths.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_path", ""),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPoolMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
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

func TestAccResourceNsxtPolicyLBPool_snat(t *testing.T) {
	testResourceName := "nsxt_policy_lb_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBPoolCheckDestroy(state, accTestPolicyLBPoolCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBPoolSnatTemplate("DISABLED"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.type", "DISABLED"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.ip_pool_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPoolSnatTemplate("IPPOOL"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.type", "IPPOOL"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.ip_pool_addresses.#", "3"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyLBPoolSnatTemplate("AUTOMAP"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBPoolExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBPoolCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBPoolCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.type", "AUTOMAP"),
					resource.TestCheckResourceAttr(testResourceName, "snat.0.ip_pool_addresses.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBPool_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_pool.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBPoolMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBPoolExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		nsxClient := infra.NewLbPoolsClient(connector)

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBPool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBPool resource ID not set in resources")
		}

		_, err := nsxClient.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy LBPool ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyLBPoolCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	nsxClient := infra.NewLbPoolsClient(connector)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_pool" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		_, err := nsxClient.Get(resourceID)
		if err == nil {
			return fmt.Errorf("Policy LBPool %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBPoolMemberTemplate(createFlow bool) string {
	var attrMap map[string]string
	var members string
	if createFlow {
		attrMap = accTestPolicyLBPoolCreateAttributes
		members = `
  member {
    admin_state                = "ENABLED"
    backup_member              = false
    display_name               = "member1"
    ip_address                 = "5.5.5.5"
    max_concurrent_connections = 7
    port                       = "77"
    weight                     = 2
  }

  member {
    admin_state  = "DISABLED"
    display_name = "member2"
    ip_address   = "5.5.5.3"
  }
`
	} else {
		attrMap = accTestPolicyLBPoolUpdateAttributes
		members = `
  member {
    admin_state                = "ENABLED"
    backup_member              = false
    display_name               = "member1"
    ip_address                 = "5.5.5.5"
    max_concurrent_connections = 100
    port                       = "88"
    weight                     = 1
  }
`
	}

	monitorPaths := ""
	// Use either current or deprecated attribute for active monitors
	if attrMap["active_monitor_paths"] != "" {
		monitorPaths = fmt.Sprintf("active_monitor_paths = [\"%s\"]", attrMap["active_monitor_paths"])
	} else {
		if attrMap["active_monitor_path"] != "" {
			monitorPaths = fmt.Sprintf("active_monitor_path = \"%s\"", attrMap["active_monitor_path"])
		}
	}
	return fmt.Sprintf(`
resource "nsxt_policy_lb_pool" "test" {
  display_name = "%s"
  description  = "%s"
  algorithm = "%s"
  %s
  min_active_members = %s
  tcp_multiplexing_number = %s
  %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_pool.test.path
}`, attrMap["display_name"], attrMap["description"], attrMap["algorithm"], members, attrMap["min_active_members"], attrMap["tcp_multiplexing_number"], monitorPaths)
}

func testAccNsxtPolicyLBPoolGroupTemplate() string {
	attrMap := accTestPolicyLBPoolUpdateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
    display_name = "lb-test"
}

resource "nsxt_policy_lb_pool" "test" {
  display_name = "%s"
  description  = "%s"
  algorithm = "%s"
  min_active_members = %s
  tcp_multiplexing_number = %s

  member_group {
      group_path       = nsxt_policy_group.test.path
      allow_ipv4       = false
      allow_ipv6       = true
      max_ip_list_size = 10
      port             = 888
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_pool.test.path
}`, attrMap["display_name"], attrMap["description"], attrMap["algorithm"], attrMap["min_active_members"], attrMap["tcp_multiplexing_number"])
}

func testAccNsxtPolicyLBPoolSnatTemplate(snatType string) string {
	attrMap := accTestPolicyLBPoolCreateAttributes
	ipAddresses := ""
	if snatType == "IPPOOL" {
		ipAddresses = `
ip_pool_addresses = ["1.1.1.60-1.1.1.82", "1.1.2.0/24", "3.3.3.3"]
`
	}
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
    display_name = "lb-test"
}

resource "nsxt_policy_lb_pool" "test" {
  display_name = "%s"
  description  = "%s"

  member_group {
      allow_ipv4 = true
      allow_ipv6 = false
      group_path = nsxt_policy_group.test.path
  }

  snat {
      type = "%s"
      %s
  }
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_pool.test.path
}`, attrMap["display_name"], attrMap["description"], snatType, ipAddresses)
}

func testAccNsxtPolicyLBPoolMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
    display_name = "lb-test"
}

resource "nsxt_policy_lb_pool" "test" {
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
  path = nsxt_policy_lb_pool.test.path
}`, accTestPolicyLBPoolUpdateAttributes["display_name"])
}
