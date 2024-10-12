// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func TestAccDataSourceNsxtPolicyIpBlock_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyIPBlockBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.0.0")
	})
}

func TestAccDataSourceNsxtPolicyIpBlock_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyIPBlockBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyIPBlockBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_ip_block.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyIPBlockDeleteByName(testAccGetSessionContext(), name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyIPBlockCreate(testAccGetSessionContext(), name, newUUID(), "4001::/64", false); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyIPBlockReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyIPBlockCreate(context tf_api.SessionContext, name, id, cidr string, isPrivate bool) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewIpBlocksClient(context, connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	displayName := name
	description := name
	obj := model.IpAddressBlock{
		Description: &description,
		DisplayName: &displayName,
		Cidr:        &cidr,
	}
	if isPrivate {
		visibility := model.IpAddressBlock_VISIBILITY_PRIVATE
		obj.Visibility = &visibility
	}

	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("IpAddressBlock", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyIPBlockDeleteByName(context tf_api.SessionContext, name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewIpBlocksClient(context, connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("IpiAddressBlock", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(*objInList.Id)
			if err != nil {
				return handleDeleteError("IpAddressBlock", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting IpBlock '%s': resource not found", name)
}

func testAccNsxtPolicyIPBlockReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_ip_block" "test" {
%s
  display_name = "%s"
}`, context, name)
}
