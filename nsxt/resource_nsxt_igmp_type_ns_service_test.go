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

func TestAccResourceNsxtIgmpTypeNsService_basic(t *testing.T) {
	serviceName := getAccTestResourceName()
	updateServiceName := getAccTestResourceName()
	testResourceName := "nsxt_igmp_type_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIgmpServiceCheckDestroy(state, updateServiceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIgmpServiceCreateTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIgmpServiceExists(serviceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "igmp service"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXIgmpServiceCreateTemplate(updateServiceName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIgmpServiceExists(updateServiceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateServiceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "igmp service"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIgmpTypeNsService_importBasic(t *testing.T) {
	serviceName := getAccTestResourceName()
	testResourceName := "nsxt_igmp_type_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIgmpServiceCheckDestroy(state, serviceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIgmpServiceCreateTemplate(serviceName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXIgmpServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX igmp service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX IGMP NS service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadIgmpTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving IGMP NS service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if IGMP NS service %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX IGMP NS service %s wasn't found", displayName)
	}
}

func testAccNSXIgmpServiceCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_igmp_set_ns_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadIgmpTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving IGMP NS service ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("NSX IGMP NS service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXIgmpServiceCreateTemplate(serviceName string) string {
	return fmt.Sprintf(`
resource "nsxt_igmp_type_ns_service" "test" {
  description  = "igmp service"
  display_name = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName)
}
