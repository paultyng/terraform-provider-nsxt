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

func TestAccResourceNsxtLbService_basic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbServiceCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServiceExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "size", "SMALL"),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", "DEBUG"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbServiceUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServiceExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "size", "SMALL"),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", "ERROR"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbService_withServers(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_service.test"
	logLevel := "EMERGENCY"
	updatedLogLevel := "INFO"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbServiceCreateTemplateWithServers(name, logLevel),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServiceExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "size", "SMALL"),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", logLevel),
					resource.TestCheckResourceAttr(testResourceName, "virtual_server_ids.#", "2"),
				),
			},
			{
				Config: testAccNSXLbServiceCreateTemplateWithServers(name, updatedLogLevel),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbServiceExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "enabled", "true"),
					resource.TestCheckResourceAttrSet(testResourceName, "logical_router_id"),
					resource.TestCheckResourceAttr(testResourceName, "size", "SMALL"),
					resource.TestCheckResourceAttr(testResourceName, "error_log_level", updatedLogLevel),
					resource.TestCheckResourceAttr(testResourceName, "virtual_server_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbService_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbServiceCreateTemplate(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX load balancer service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX load balancer service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerService(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving load balancer service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if load balancer service %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX load balancer service %s wasn't found", displayName)
	}
}

func testAccNSXLbServiceCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerService(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving load balancer service ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("NSX load balancer service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbCreateTopology() string {
	edgeClusterName := getEdgeClusterName()
	tier0Name := getTier0RouterName()
	return fmt.Sprintf(`
data "nsxt_edge_cluster" "EC" {
  display_name = "%s"
}

data "nsxt_logical_tier0_router" "test" {
  display_name = "%s"
}

resource "nsxt_logical_router_link_port_on_tier0" "test" {
  display_name      = "port_on_tier0"
  logical_router_id = "${data.nsxt_logical_tier0_router.test.id}"
}

resource "nsxt_logical_tier1_router" "test" {
  display_name    = "test"
  edge_cluster_id = "${data.nsxt_edge_cluster.EC.id}"
}

resource "nsxt_logical_router_link_port_on_tier1" "test" {
  logical_router_id             = "${nsxt_logical_tier1_router.test.id}"
  linked_logical_router_port_id = "${nsxt_logical_router_link_port_on_tier0.test.id}"
}`, edgeClusterName, tier0Name)
}

func testAccNSXLbServiceCreateTemplate(name string) string {
	return testAccNSXLbCreateTopology() + fmt.Sprintf(`
resource "nsxt_lb_service" "test" {
  display_name      = "%s"
  enabled           = true
  description       = "Acceptance Test"
  logical_router_id = "${nsxt_logical_tier1_router.test.id}"
  size              = "SMALL"
  error_log_level   = "DEBUG"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  depends_on = ["nsxt_logical_router_link_port_on_tier1.test"]
}`, name)
}

func testAccNSXLbServiceUpdateTemplate(name string) string {
	return testAccNSXLbCreateTopology() + fmt.Sprintf(`
resource "nsxt_lb_service" "test" {
  display_name      = "%s"
  enabled           = false
  description       = "Acceptance Test Update"
  logical_router_id = "${nsxt_logical_tier1_router.test.id}"
  size              = "SMALL"
  error_log_level   = "ERROR"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }

  depends_on = ["nsxt_logical_router_link_port_on_tier1.test"]
}`, name)
}

func testAccNSXLbServiceCreateTemplateWithServers(name string, logLevel string) string {
	return testAccNSXLbCreateTopology() + fmt.Sprintf(`

resource "nsxt_lb_fast_tcp_application_profile" "test" {
  display_name = "lb service test"
}

resource "nsxt_lb_fast_udp_application_profile" "test" {
  display_name = "lb service test"
}

resource "nsxt_lb_tcp_virtual_server" "test"{
  application_profile_id = "${nsxt_lb_fast_tcp_application_profile.test.id}"
  ip_address             = "1.1.1.2"
  ports                  = ["7887"]
}

resource "nsxt_lb_udp_virtual_server" "test"{
  application_profile_id = "${nsxt_lb_fast_udp_application_profile.test.id}"
  ip_address             = "1.1.1.2"
  ports                  = ["7888"]
}

resource "nsxt_lb_service" "test" {
  display_name       = "%s"
  logical_router_id  = "${nsxt_logical_tier1_router.test.id}"
  error_log_level    = "%s"
  virtual_server_ids = ["${nsxt_lb_tcp_virtual_server.test.id}", "${nsxt_lb_udp_virtual_server.test.id}"]

  depends_on = ["nsxt_logical_router_link_port_on_tier1.test"]
}`, name, logLevel)
}
