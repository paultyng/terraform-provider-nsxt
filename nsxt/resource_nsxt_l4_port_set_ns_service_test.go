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

func TestAccResourceNsxtL4PortNsService_basic(t *testing.T) {
	serviceName := getAccTestResourceName()
	updateServiceName := getAccTestResourceName()
	testResourceName := "nsxt_l4_port_set_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXL4ServiceCheckDestroy(state, updateServiceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXserviceCreateTemplate(serviceName, "TCP", "99"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXL4ServiceExists(serviceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "l4 service"),
					resource.TestCheckResourceAttr(testResourceName, "protocol", "TCP"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXserviceCreateTemplate(updateServiceName, "UDP", "98"),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXL4ServiceExists(updateServiceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateServiceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "l4 service"),
					resource.TestCheckResourceAttr(testResourceName, "protocol", "UDP"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtL4PortNsService_importBasic(t *testing.T) {
	serviceName := getAccTestResourceName()
	testResourceName := "nsxt_l4_port_set_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXL4ServiceCheckDestroy(state, serviceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXserviceCreateTemplate(serviceName, "TCP", "99"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXL4ServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX L4 NS service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX L4 NS service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadL4PortSetNSService(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving L4 NS service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if L4 NS service %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX L4 NS service %s wasn't found", displayName)
	}
}

func testAccNSXL4ServiceCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_l4_port_set_ns_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadL4PortSetNSService(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving L4 NS service ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("NSX L4 NS service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXserviceCreateTemplate(serviceName string, protocol string, port string) string {
	return fmt.Sprintf(`
resource "nsxt_l4_port_set_ns_service" "test" {
  description       = "l4 service"
  display_name      = "%s"
  protocol          = "%s"
  destination_ports = [ "%s" ]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, protocol, port)
}
