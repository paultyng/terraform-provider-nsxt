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

func TestAccResourceNsxtLbIcmpMonitor_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_icmp_monitor.test"
	port := "7887"
	updatedPort := "8778"
	count := "15"
	interval := "9"
	timeout := "20"
	dataLength := "100"
	updatedCount := "5"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbIcmpMonitorCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbIcmpMonitorCreateTemplate(name, count, interval, port, timeout, dataLength),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbIcmpMonitorExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", count),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", count),
					resource.TestCheckResourceAttr(testResourceName, "interval", interval),
					resource.TestCheckResourceAttr(testResourceName, "timeout", timeout),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", port),
					resource.TestCheckResourceAttr(testResourceName, "data_length", dataLength),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbIcmpMonitorCreateTemplate(updatedName, updatedCount, interval, updatedPort, timeout, dataLength),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbIcmpMonitorExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "fall_count", updatedCount),
					resource.TestCheckResourceAttr(testResourceName, "rise_count", updatedCount),
					resource.TestCheckResourceAttr(testResourceName, "interval", interval),
					resource.TestCheckResourceAttr(testResourceName, "timeout", timeout),
					resource.TestCheckResourceAttr(testResourceName, "monitor_port", updatedPort),
					resource.TestCheckResourceAttr(testResourceName, "data_length", dataLength),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbIcmpMonitor_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_icmp_monitor.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbIcmpMonitorCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbIcmpMonitorCreateTemplateTrivial(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbIcmpMonitorExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB icmp monitor resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB icmp monitor resource ID not set in resources ")
		}

		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerIcmpMonitor(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB icmp monitor with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB icmp monitor %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB icmp monitor %s wasn't found", displayName)
	}
}

func testAccNSXLbIcmpMonitorCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_icmp_monitor" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerIcmpMonitor(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB icmp monitor with ID %s. Error: %v", resourceID, err)
		}

		if displayName == monitor.DisplayName {
			return fmt.Errorf("NSX LB icmp monitor %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbIcmpMonitorCreateTemplate(name string, count string, interval string, port string, timeout string, dataLength string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_icmp_monitor" "test" {
  description  = "test description"
  display_name = "%s"
  fall_count   = "%s"
  interval     = "%s"
  monitor_port = "%s"
  rise_count   = "%s"
  timeout      = "%s"
  data_length  = "%s"
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, count, interval, port, count, timeout, dataLength)
}

func testAccNSXLbIcmpMonitorCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_icmp_monitor" "test" {
  description = "test description"
}
`
}
