// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyVMs_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyVMsBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccEnvDefined(t, "NSXT_TEST_VM_NAME")
	})
}

func TestAccDataSourceNsxtPolicyVMs_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyVMsBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
		testAccEnvDefined(t, "NSXT_TEST_VM_NAME")
	})
}

func testAccDataSourceNsxtPolicyVMsBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "data.nsxt_policy_vms.test"
	checkDataSourceName := "data.nsxt_policy_vm.check"
	checkResourceName := "nsxt_policy_group.check"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVMsTemplate("bios_id", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrPair(checkResourceName, "display_name", checkDataSourceName, "bios_id"),
				),
			},
			{
				Config: testAccNsxtPolicyVMsTemplate("external_id", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrPair(checkResourceName, "display_name", checkDataSourceName, "external_id"),
				),
			},
			{
				Config: testAccNsxtPolicyVMsTemplate("instance_id", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrPair(checkResourceName, "display_name", checkDataSourceName, "instance_id"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyVMs_filter(t *testing.T) {
	testResourceName := "data.nsxt_policy_vms.test"
	checkResourceName := "nsxt_policy_group.check"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVMsTemplateFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(checkResourceName, "display_name"),
				),
			},
		},
	})
}

func testAccNsxtPolicyVMsTemplate(valueType string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_vms" "test" {
%s
  value_type = "%s"
}

data "nsxt_policy_vm" "check" {
%s
  display_name = "%s"
}

resource "nsxt_policy_group" "check" {
%s
  display_name = data.nsxt_policy_vms.test.items["%s"]
}`, context, valueType, context, getTestVMName(), context, getTestVMName())
}

func testAccNsxtPolicyVMsTemplateFilter() string {
	return `
data "nsxt_policy_vms" "test" {
  state    = "running"
  guest_os = "ubuntu"
}

resource "nsxt_policy_group" "check" {
  display_name = length(data.nsxt_policy_vms.test.items)
}`
}
