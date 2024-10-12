// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var vmTagsResourceName = "test"
var vmTagsFullResourceName = "nsxt_vm_tags." + vmTagsResourceName

func TestAccResourceNsxtVMTags_basic(t *testing.T) {
	vmID := getTestVMID()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccTestDeprecated(t)
			testAccEnvDefined(t, "NSXT_TEST_VM_ID")
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXVMTagsCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXVMTagsCreateTemplate(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXVMTagsCheckExists(),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.#", "1"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.934310497.scope", "a"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.934310497.tag", "b"),
				),
			},
			{
				Config: testAccNSXVMTagsUpdateTemplate(vmID),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXVMTagsCheckExists(),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.#", "1"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.600822426.scope", "c"),
					resource.TestCheckResourceAttr(vmTagsFullResourceName, "logical_port_tag.600822426.tag", "d"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVMTags_import_basic(t *testing.T) {
	vmID := getTestVMID()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccTestDeprecated(t)
			testAccEnvDefined(t, "NSXT_TEST_VM_ID")
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXVMTagsCheckDestroy(state)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXVMTagsCreateTemplate(vmID),
			},
			{
				ResourceName:            vmTagsFullResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instance_id"},
			},
		},
	})
}

func testAccNSXVMTagsCheckExists() resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[vmTagsFullResourceName]
		if !ok {
			return fmt.Errorf("NSX vm tags resource %s not found in resources", vmTagsFullResourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX vm tags resource ID not set in resources ")
		}

		_, err := findVMByExternalID(nsxClient, resourceID)
		if err != nil {
			return fmt.Errorf("Failed to find VM %s", resourceID)
		}

		return nil
	}
}

func testAccNSXVMTagsCheckDestroy(state *terraform.State) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vm_tags" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		vm, err := findVMByExternalID(nsxClient, resourceID)
		if err != nil {
			return fmt.Errorf("Failed to find VM %s", resourceID)
		}

		if len(vm.Tags) > 0 {
			return fmt.Errorf("VM %s still has tags, although nsxt_vm_tags was deleted", resourceID)
		}
	}
	return nil
}

func testAccNSXVMTagsCreateTemplate(instanceID string) string {
	return fmt.Sprintf(`
resource "nsxt_vm_tags" "%s" {
  instance_id = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  logical_port_tag {
    scope = "a"
    tag   = "b"
  }
}`, vmTagsResourceName, instanceID)
}

func testAccNSXVMTagsUpdateTemplate(instanceID string) string {
	return fmt.Sprintf(`
resource "nsxt_vm_tags" "%s" {
  instance_id = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }

  logical_port_tag {
    scope = "c"
    tag   = "d"
  }
}`, vmTagsResourceName, instanceID)
}
