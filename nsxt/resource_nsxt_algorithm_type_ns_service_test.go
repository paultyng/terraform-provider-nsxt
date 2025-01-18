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

func TestAccResourceNsxtAlgorithmTypeNsService_basic(t *testing.T) {
	serviceName := getAccTestResourceName()
	updateServiceName := getAccTestResourceName()
	testResourceName := "nsxt_algorithm_type_ns_service.test"
	destPort := "21"
	updatedDestPort := "21"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXAlgServiceCheckDestroy(state, updateServiceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXAlgServiceCreateTemplate(serviceName, "FTP", "9000-9001", destPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXAlgServiceExists(serviceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "alg service"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", "FTP"),
					resource.TestCheckResourceAttr(testResourceName, "destination_port", destPort),
					resource.TestCheckResourceAttr(testResourceName, "source_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXAlgServiceCreateTemplate(updateServiceName, "ORACLE_TNS", "600", updatedDestPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXAlgServiceExists(updateServiceName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateServiceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "alg service"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", "ORACLE_TNS"),
					resource.TestCheckResourceAttr(testResourceName, "destination_port", updatedDestPort),
					resource.TestCheckResourceAttr(testResourceName, "source_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtAlgorithmTypeNsService_importBasic(t *testing.T) {
	serviceName := getAccTestResourceName()
	testResourceName := "nsxt_algorithm_type_ns_service.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXAlgServiceCheckDestroy(state, serviceName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXAlgServiceCreateTemplate(serviceName, "FTP", "9000-9001", "21"),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXAlgServiceExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX alg ns service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX alg ns service resource ID not set in resources ")
		}

		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadAlgTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving alg ns service ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if alg ns service %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == service.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX alg ns service %s wasn't found", displayName)
	}
}

func testAccNSXAlgServiceCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	if nsxClient == nil {
		return fmt.Errorf("Failed to initialize the client")
	}
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_algorithm_type_ns_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		service, responseCode, err := nsxClient.GroupingObjectsApi.ReadAlgTypeNSService(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving alg ns service ID %s. Error: %v", resourceID, err)
		}

		if displayName == service.DisplayName {
			return fmt.Errorf("NSX alg ns service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXAlgServiceCreateTemplate(serviceName string, protocol string, sourcePorts string, destPort string) string {
	return fmt.Sprintf(`
resource "nsxt_algorithm_type_ns_service" "test" {
  description       = "alg service"
  display_name      = "%s"
  algorithm         = "%s"
  source_ports      = ["%s"]
  destination_port  = "%s"
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, serviceName, protocol, sourcePorts, destPort)
}
