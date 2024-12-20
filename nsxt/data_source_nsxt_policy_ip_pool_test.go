/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func TestAccDataSourceNsxtPolicyIpPool_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyIPPoolBasic(t, false, func() {
		testAccOnlyLocalManager(t)
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyIpPool_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyIPPoolBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyIPPoolBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_ip_pool.test"
	var clientType tf_api.ClientType = tf_api.Local
	if withContext {
		clientType = tf_api.Multitenancy
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyIPPoolDeleteByName(name, clientType)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyIPPoolCreate(name, clientType); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyIPPoolReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyIPPoolCreate(name string, clientType tf_api.ClientType) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewIpPoolsClient(testAccGetContextByType(clientType), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	displayName := name
	description := name
	obj := model.IpAddressPool{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("IpAddressPool", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyIPPoolDeleteByName(name string, clientType tf_api.ClientType) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewIpPoolsClient(testAccGetContextByType(clientType), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("IpPool", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(*objInList.Id)
			if err != nil {
				return handleDeleteError("IpAddressPool", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting IpPool '%s': resource not found", name)
}

func testAccNsxtPolicyIPPoolReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtProjectContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_ip_pool" "test" {
%s
  display_name = "%s"
}`, context, name)
}
