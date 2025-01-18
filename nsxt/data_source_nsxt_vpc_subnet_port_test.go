// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtVpcSubnetPort_basic(t *testing.T) {
	subnetObjectName := "data.nsxt_vpc_subnet.test"
	portObjectName := "data.nsxt_vpc_subnet_port.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyVPC(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "9.0.0")
			testAccEnvDefined(t, "NSXT_TEST_VPC_VM_ID")
			testAccEnvDefined(t, "NSXT_TEST_VPC_SUBNET_NAME")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcSubnetPortReadTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(portObjectName, "display_name"),
					resource.TestCheckResourceAttrSet(portObjectName, "path"),
					resource.TestCheckResourceAttrSet(subnetObjectName, "path"),
				),
			},
		},
	})
}

func testAccNsxtVpcSubnetPortReadTemplate() string {
	context := testAccNsxtPolicyMultitenancyContext()
	subnetName := os.Getenv("NSXT_TEST_VPC_SUBNET_NAME")
	vmID := os.Getenv("NSXT_TEST_VPC_VM_ID")
	return fmt.Sprintf(`
data "nsxt_vpc_subnet" "test" {
%s
  display_name = "%s"
}

data "nsxt_vpc_subnet_port" "test" {
  subnet_path = data.nsxt_vpc_subnet.test.path
  vm_id       = "%s"
}`, context, subnetName, vmID)
}
