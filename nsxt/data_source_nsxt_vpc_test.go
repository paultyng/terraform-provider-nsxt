// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects"
)

func TestAccDataSourceNsxtVPC_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyMultitenancy(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.2")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtVPCDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtVPCCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtVPCReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "short_id"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtVPCCreate(name string) error {

	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("error during test client initialization: %v", err)
	}

	ipBlockID := newUUID()
	err = testAccDataSourceNsxtPolicyIPBlockCreate(testAccGetProjectContext(), name, ipBlockID, "192.168.240.0/24", true)
	if err != nil {
		return err
	}

	client := projects.NewVpcsClient(connector)
	projID := os.Getenv("NSXT_PROJECT_ID")

	displayName := name
	description := name
	addrType := model.Vpc_IP_ADDRESS_TYPE_IPV4
	enableDhcp := false
	disableGateway := true
	obj := model.Vpc{
		Description:       &description,
		DisplayName:       &displayName,
		IpAddressType:     &addrType,
		DhcpConfig:        &model.DhcpConfig{EnableDhcp: &enableDhcp},
		ServiceGateway:    &model.ServiceGateway{Disable: &disableGateway},
		PrivateIpv4Blocks: []string{fmt.Sprintf("/orgs/default/projects/%s/infra/ip-blocks/%s", projID, ipBlockID)},
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(defaultOrgID, projID, id, obj)
	if err != nil {
		return handleCreateError("VPC", id, err)
	}
	return nil
}

func testAccDataSourceNsxtVPCDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("error during test client initialization: %v", err)
	}

	client := projects.NewVpcsClient(connector)
	projID := os.Getenv("NSXT_PROJECT_ID")

	// Find the object by name
	objList, err := client.List(defaultOrgID, projID, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("VPC", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(defaultOrgID, projID, *objInList.Id, nil)
			if err != nil {
				return handleDeleteError("VPC", *objInList.Id, err)
			}
			return testAccDataSourceNsxtPolicyIPBlockDeleteByName(testAccGetProjectContext(), name)
		}
	}
	return fmt.Errorf("error while deleting VPC '%s': resource not found", name)
}

func testAccNsxtVPCReadTemplate(name string) string {
	// We just need the project context as VPC is not under VPC, but the VPC itself
	context := testAccNsxtProjectContext()
	return fmt.Sprintf(`
data "nsxt_policy_ip_block" "test" {
%s
  display_name = "%s"
}
data "nsxt_vpc" "test" {
%s
  display_name = "%s"
}`, context, name, context, name)
}
