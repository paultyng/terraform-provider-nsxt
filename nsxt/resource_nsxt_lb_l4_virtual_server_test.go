// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var accTestLbVirtualServerHelper1Name = getAccTestResourceName()
var accTestLbVirtualServerHelper2Name = getAccTestResourceName()

func TestAccResourceNsxtLbTCPVirtualServer_basic(t *testing.T) {
	testAccResourceNsxtLbL4VirtualServer(t, "tcp")
}

func TestAccResourceNsxtLbUDPVirtualServer_basic(t *testing.T) {
	testAccResourceNsxtLbL4VirtualServer(t, "udp")
}

func testAccResourceNsxtLbL4VirtualServer(t *testing.T, protocol string) {
	name := getAccTestResourceName()
	fullName := fmt.Sprintf("nsxt_lb_%s_virtual_server.test", protocol)
	port := "888-890"
	updatedPort := "999"
	memberPort := "786-788"
	updatedMemberPort := "10098"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbL4VirtualServerCheckDestroy(state, protocol, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbL4VirtualServerCreateTemplate(name, protocol, port, memberPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbL4VirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "false"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.#", "2"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.1", "1001"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.0", memberPort),
					resource.TestCheckResourceAttr(fullName, "enabled", "true"),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "ports.#", "2"),
					resource.TestCheckResourceAttr(fullName, "ports.0", port),
					resource.TestCheckResourceAttr(fullName, "ports.1", "2002"),
					resource.TestCheckResourceAttr(fullName, "max_concurrent_connections", "20"),
					resource.TestCheckResourceAttr(fullName, "max_new_connection_rate", "10"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(fullName, "pool_id"),
					resource.TestCheckResourceAttrSet(fullName, "sorry_pool_id"),
				),
			},
			{
				Config: testAccNSXLbL4VirtualServerCreateTemplate(name, protocol, updatedPort, updatedMemberPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbL4VirtualServerExists(name, fullName),
					resource.TestCheckResourceAttr(fullName, "display_name", name),
					resource.TestCheckResourceAttr(fullName, "description", "test description"),
					resource.TestCheckResourceAttr(fullName, "access_log_enabled", "false"),
					resource.TestCheckResourceAttrSet(fullName, "application_profile_id"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.#", "2"),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.0", updatedMemberPort),
					resource.TestCheckResourceAttr(fullName, "default_pool_member_ports.1", "1001"),
					resource.TestCheckResourceAttr(fullName, "enabled", "true"),
					resource.TestCheckResourceAttr(fullName, "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr(fullName, "ports.#", "2"),
					resource.TestCheckResourceAttr(fullName, "ports.0", updatedPort),
					resource.TestCheckResourceAttr(fullName, "ports.1", "2002"),
					resource.TestCheckResourceAttr(fullName, "max_concurrent_connections", "20"),
					resource.TestCheckResourceAttr(fullName, "max_new_connection_rate", "10"),
					resource.TestCheckResourceAttr(fullName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(fullName, "pool_id"),
					resource.TestCheckResourceAttrSet(fullName, "sorry_pool_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbTCPVirtualServer_importBasic(t *testing.T) {
	testAccResourceNsxtLbL4VirtualServerImport(t, "tcp")
}

func TestAccResourceNsxtLbUDPVirtualServer_importBasic(t *testing.T) {
	testAccResourceNsxtLbL4VirtualServerImport(t, "udp")
}

func testAccResourceNsxtLbL4VirtualServerImport(t *testing.T, protocol string) {
	name := getAccTestResourceName()
	resourceName := fmt.Sprintf("nsxt_lb_%s_virtual_server.test", protocol)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbL4VirtualServerCheckDestroy(state, protocol, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbL4VirtualServerCreateTemplateTrivial(protocol),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbL4VirtualServerExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB virtual server resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB virtual server resource ID not set in resources ")
		}

		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerVirtualServer(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB virtual server with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB virtual server %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB virtual server %s wasn't found", displayName)
	}
}

func testAccNSXLbL4VirtualServerCheckDestroy(state *terraform.State, protocol string, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != fmt.Sprintf("nsxt_lb_%s_virtual_server", protocol) {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		server, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerVirtualServer(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB virtual server with ID %s. Error: %v", resourceID, err)
		}

		if displayName == server.DisplayName {
			return fmt.Errorf("NSX LB virtual server %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbL4VirtualServerCreateTemplate(name string, protocol string, port string, memberPort string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_fast_%s_application_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_source_ip_persistence_profile" "test" {
  display_name = "%s"
}

resource "nsxt_lb_pool" "test" {
  display_name = "%s"
  algorithm    = "ROUND_ROBIN"
}

resource "nsxt_lb_pool" "sorry" {
  display_name = "%s"
  algorithm    = "ROUND_ROBIN"
}

resource "nsxt_lb_%s_virtual_server" "test" {
  display_name               = "%s"
  description                = "test description"
  access_log_enabled         = false
  application_profile_id     = "${nsxt_lb_fast_%s_application_profile.test.id}"
  enabled                    = true
  ip_address                 = "1.1.1.1"
  ports                      = ["%s", "2002"]
  default_pool_member_ports  = ["%s", "1001"]
  max_concurrent_connections = 20
  max_new_connection_rate    = 10
  persistence_profile_id     = "${nsxt_lb_source_ip_persistence_profile.test.id}"
  pool_id                    = "${nsxt_lb_pool.test.id}"
  sorry_pool_id              = "${nsxt_lb_pool.sorry.id}"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, protocol, name, name, accTestLbVirtualServerHelper1Name, accTestLbVirtualServerHelper2Name, protocol, name, protocol, port, memberPort)
}

func testAccNSXLbL4VirtualServerCreateTemplateTrivial(protocol string) string {
	return fmt.Sprintf(`

resource "nsxt_lb_fast_%s_application_profile" "test" {
  display_name = "lb virtual server test"
}

resource "nsxt_lb_%s_virtual_server" "test" {
  description            = "test description"
  application_profile_id = "${nsxt_lb_fast_%s_application_profile.test.id}"
  ip_address             = "2.2.2.2"
  ports                  = ["443"]
}
`, protocol, protocol, protocol)
}
