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

func TestAccResourceNsxtNsServiceGroup_basic(t *testing.T) {
	serviceName := getAccTestResourceName()
	updateServiceName := getAccTestResourceName()
	testResourceName := "nsxt_ns_service_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXServiceGroupCheckDestroy(state, updateServiceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXServiceGroupCreateTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXServiceGroupExists(serviceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "service group"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "1"),
				),
			},
			{
				Config: testAccNSXServiceGroupUpdateTemplate(updateServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXServiceGroupExists(updateServiceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateServiceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "service group"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "members.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNsServiceGroup_importBasic(t *testing.T) {
	serviceName := getAccTestResourceName()
	testResourceName := "nsxt_ns_service_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXServiceGroupCheckDestroy(state, serviceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXServiceGroupCreateTemplate(serviceName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXServiceGroupExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX service group resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX service group resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSServiceGroup(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving service group ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if service group %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX NS service group %s wasn't found", displayName)
	}
}

func testAccNSXServiceGroupCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ns_service_group" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSServiceGroup(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving NS service group ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("NSX NS service group %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXServiceGroupCreateServices() string {
	return `
resource "nsxt_ip_protocol_ns_service" "srv1" {
  display_name = "test_ip_protocol_service"
  protocol     = "17"
}

resource "nsxt_l4_port_set_ns_service" "srv2" {
  display_name      = "test_l4_service"
  protocol          = "TCP"
  destination_ports = [ "8080" ]
}`
}

func testAccNSXServiceGroupCreateTemplate(serviceName string) string {
	return testAccNSXServiceGroupCreateServices() + fmt.Sprintf(`
resource "nsxt_ns_service_group" "test" {
  description  = "service group"
  display_name = "%s"
  members      = ["${nsxt_ip_protocol_ns_service.srv1.id}"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName)
}

func testAccNSXServiceGroupUpdateTemplate(serviceName string) string {
	return testAccNSXServiceGroupCreateServices() + fmt.Sprintf(`
resource "nsxt_ns_service_group" "test" {
  description  = "service group"
  display_name = "%s"
  members      = ["${nsxt_ip_protocol_ns_service.srv1.id}", "${nsxt_l4_port_set_ns_service.srv2.id}"]

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, serviceName)
}
