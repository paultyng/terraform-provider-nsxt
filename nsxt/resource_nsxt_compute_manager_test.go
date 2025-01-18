// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/fabric"
)

func TestAccResourceNsxtComputeManager_basic(t *testing.T) {
	computeManagerName := getAccTestResourceName()
	updateComputeManagerName := "updated-" + computeManagerName
	testResourceName := "nsxt_compute_manager.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestFabric(t)
			testAccTestVCCredentials(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXComputeManagerCheckDestroy(state, updateComputeManagerName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXComputeManagerCreateTemplate(computeManagerName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXComputeManagerExists(computeManagerName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", computeManagerName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test compute manager"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXComputeManagerCreateTemplate(updateComputeManagerName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXComputeManagerExists(updateComputeManagerName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateComputeManagerName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test compute manager"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtComputeManager_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	clusterName := getAccTestResourceName()
	testResourceName := "nsxt_compute_manager.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestFabric(t)
			testAccPreCheck(t)
			testAccTestVCCredentials(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXComputeManagerCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXComputeManagerCreateTemplate(clusterName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXComputeManagerExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Compute Manager resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX Compute Manager resource ID not set in resources ")
		}

		client := fabric.NewComputeManagersClient(connector)
		obj, err := client.Get(resourceID)
		if err != nil {
			return fmt.Errorf("error while retrieving Compute Manager ID %s. Error: %v", resourceID, err)
		}

		if displayName == *obj.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX Compute Manager %s wasn't found", displayName)
	}
}

func testAccNSXComputeManagerCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_compute_manager" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		client := fabric.NewComputeManagersClient(connector)
		obj, err := client.Get(resourceID)

		if isNotFoundError(err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while retrieving Compute Manager ID %s. Error: %v", resourceID, err)
		}

		if obj.DisplayName != nil && displayName == *obj.DisplayName {
			return fmt.Errorf("NSX Compute Manager %s still exists", displayName)
		}
	}

	return nil
}

func testAccNSXComputeManagerCreateTemplate(displayName string) string {
	return fmt.Sprintf(`
resource "nsxt_compute_manager" "test" {
  description  = "Terraform test compute manager"
  display_name = "%s"
  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  server = "%s"

  credential {
    username_password_login {
      username = "%s"
      password = "%s"
      thumbprint = "%s"
    }
  }
  origin_type = "vCenter"
}
`, displayName, getTestVCIPAddress(), getTestVCUsername(), getTestVCPassword(), getTestVCThumbprint())
}
