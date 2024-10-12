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
	"github.com/vmware/go-vmware-nsxt/manager"
)

func TestAccDataSourceNsxtNsService_basic(t *testing.T) {
	serviceName := getAccTestDataSourceName()
	testResourceName := "data.nsxt_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtNsServiceDeleteByName(serviceName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtNsServiceCreate(serviceName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNSXNsServiceReadTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtNsService_systemOwned(t *testing.T) {
	serviceName := "WINS"
	testResourceName := "data.nsxt_ns_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNsServiceReadTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName),
				),
			},
		},
	})
}

func testAccDataSourceNsxtNsServiceCreate(serviceName string) error {
	nsxClient, err := testAccGetClient()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	nsService := manager.IgmpTypeNsService{
		NsService: manager.NsService{
			DisplayName:    serviceName,
			Description:    serviceName,
			DefaultService: false,
		},
		NsserviceElement: manager.IgmpTypeNsServiceEntry{
			ResourceType: "IGMPTypeNSService",
		},
	}
	_, responseCode, err := nsxClient.GroupingObjectsApi.CreateIgmpTypeNSService(nsxClient.Context, nsService)
	if err != nil {
		return fmt.Errorf("Error during nsService creation: %v", err)
	}

	if responseCode.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during nsService creation: %v", responseCode.StatusCode)
	}
	return nil
}

func testAccDataSourceNsxtNsServiceDeleteByName(serviceName string) error {
	nsxClient, err := testAccGetClient()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objList, _, err := nsxClient.GroupingObjectsApi.ListNSServices(nsxClient.Context, nil)
	if err != nil {
		return fmt.Errorf("Error while reading NS services: %v", err)
	}
	// go over the list to find the correct one
	for _, objInList := range objList.Results {
		if objInList.DisplayName == serviceName {
			localVarOptionals := make(map[string]interface{})
			responseCode, err := nsxClient.GroupingObjectsApi.DeleteNSService(nsxClient.Context, objInList.Id, localVarOptionals)
			if err != nil {
				return fmt.Errorf("Error during nsService deletion: %v", err)
			}

			if responseCode.StatusCode != http.StatusOK {
				return fmt.Errorf("Unexpected status returned during nsService deletion: %v", responseCode.StatusCode)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting NS service '%s': service not found", serviceName)
}

func testAccNSXNsServiceReadTemplate(serviceName string) string {
	return fmt.Sprintf(`
data "nsxt_ns_service" "test" {
  display_name = "%s"
}`, serviceName)
}
