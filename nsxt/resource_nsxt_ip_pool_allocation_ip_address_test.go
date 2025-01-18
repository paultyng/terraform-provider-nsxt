// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const (
	// how long to wait for updated allocation IP address list (to check destroyment)
	waitSeconds = 150
)

var testAccIPAllocationName = "nsxt_ip_pool_allocation_ip_address.test"
var testAccIPPoolName = "data.nsxt_ip_pool.acceptance_test"

func TestAccResourceNsxtIPPoolAllocationIPAddress_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccEnvDefined(t, "NSXT_TEST_IP_POOL")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIPPoolAllocationIPAddressCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIPPoolAllocationIPAddressCreateTemplate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXIPPoolAllocationIPAddressExists(),
					resource.TestCheckResourceAttrSet(testAccIPAllocationName, "ip_pool_id"),
					resource.TestCheckResourceAttrSet(testAccIPAllocationName, "allocation_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtIPPoolAllocationIPAddress_import(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccTestDeprecated(t)
			testAccEnvDefined(t, "NSXT_TEST_IP_POOL")
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXIPPoolAllocationIPAddressCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXIPPoolAllocationIPAddressCreateTemplate(),
			},
			{
				ResourceName:      testAccIPAllocationName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXIPPoolAllocationIPAddressImporterGetID,
			},
		},
	})
}

func testAccNSXIPPoolAllocationIPAddressExists() resource.TestCheckFunc {
	return func(state *terraform.State) error {
		exists, err := checkAllocationIPAddressExists(state)
		if err != nil {
			return err
		}
		if !*exists {
			return fmt.Errorf("Allocation %s in IP Pool %s does not exist", testAccIPAllocationName, testAccIPPoolName)
		}
		return nil
	}
}

func getIPPoolIDByResourceName(state *terraform.State) (string, error) {
	rsPool, ok := state.RootModule().Resources[testAccIPPoolName]
	if !ok {
		return "", fmt.Errorf("IP Pool resource %s not found in resources", testAccIPPoolName)
	}

	poolID := rsPool.Primary.ID
	if poolID == "" {
		return "", fmt.Errorf("IP Pool resource ID not set in resources ")
	}
	return poolID, nil
}

func checkAllocationIPAddressExists(state *terraform.State) (*bool, error) {
	exists := false
	poolID, err := getIPPoolIDByResourceName(state)
	if err != nil {
		return &exists, nil
	}

	rs, ok := state.RootModule().Resources[testAccIPAllocationName]
	if !ok {
		return nil, fmt.Errorf("IP Pool allocation_ip_address resource %s not found in resources", testAccIPAllocationName)
	}

	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	if nsxClient == nil {
		return nil, resourceNotSupportedError()
	}
	listResult, responseCode, err := nsxClient.PoolManagementApi.ListIpPoolAllocations(nsxClient.Context, poolID)
	if err != nil {
		return nil, fmt.Errorf("Error while retrieving allocations for IP Pool ID %s. Error: %v", poolID, err)
	}

	if responseCode.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error while checking if allocations in IP Pool %s exists. HTTP return code was %d", poolID, responseCode.StatusCode)
	}

	for _, allocationIPAddress := range listResult.Results {
		if allocationIPAddress.AllocationId == rs.Primary.ID {
			exists = true
			break
		}
	}

	return &exists, nil
}

func testAccNSXIPPoolAllocationIPAddressCheckDestroy(state *terraform.State) error {
	// PoolManagementApi.ListIpPoolAllocations() call does not return updated list within two minutes. Need to wait...
	fmt.Printf("testAccNSXIPPoolAllocationIPAddressCheckDestroy: waiting up to %d seconds\n", waitSeconds)

	timeout := time.Now().Add(waitSeconds * time.Second)
	for time.Now().Before(timeout) {
		exists, err := checkAllocationIPAddressExists(state)
		if err != nil {
			return err
		}
		if !*exists {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Timeout on check destroy IP address allocation")
}

func testAccNSXIPPoolAllocationIPAddressCreateTemplate() string {
	return fmt.Sprintf(`
data "nsxt_ip_pool" "acceptance_test" {
  display_name = "%s"
}

resource "nsxt_ip_pool_allocation_ip_address" "test" {
  ip_pool_id = "${data.nsxt_ip_pool.acceptance_test.id}"
}`, getIPPoolName())
}

func testAccNSXIPPoolAllocationIPAddressImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccIPAllocationName]
	if !ok {
		return "", fmt.Errorf("NSX IP allocation resource %s not found in resources", testAccIPAllocationName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX IP allocation resource ID not set in resources ")
	}
	poolID := rs.Primary.Attributes["ip_pool_id"]
	if poolID == "" {
		return "", fmt.Errorf("NSX IP Pool ID not set in resources ")
	}
	return fmt.Sprintf("%s/%s", poolID, resourceID), nil
}
