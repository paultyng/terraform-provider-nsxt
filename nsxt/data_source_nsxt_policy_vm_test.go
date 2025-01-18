// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyVM_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyVMBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccEnvDefined(t, "NSXT_TEST_VM_ID")
		testAccEnvDefined(t, "NSXT_TEST_VM_NAME")
	})
}

func TestAccDataSourceNsxtPolicyVM_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyVMBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
		testAccEnvDefined(t, "NSXT_TEST_VM_ID")
		testAccEnvDefined(t, "NSXT_TEST_VM_NAME")
	})
}
func testAccDataSourceNsxtPolicyVMBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "data.nsxt_policy_vm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVMReadByNameTemplate(withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "bios_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "external_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "tag.#"),
				),
			},
			{
				Config: testAccNsxtPolicyVMReadByIDTemplate(withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "bios_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "external_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "tag.#"),
				),
			},
		},
	})
}

func testAccNsxtPolicyVMReadByNameTemplate(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_vm" "test" {
%s
  display_name = "%s"
}`, context, getTestVMName())
}

func testAccNsxtPolicyVMReadByIDTemplate(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_vm" "test" {
%s
  external_id = "%s"
}`, context, getTestVMID())
}
