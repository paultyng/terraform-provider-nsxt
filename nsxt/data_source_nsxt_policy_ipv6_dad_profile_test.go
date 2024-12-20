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

func TestAccDataSourceNsxtPolicyIpv6DadProfile_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyIpv6DadProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyIpv6DadProfile_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyIpv6DadProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyIpv6DadProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_ipv6_dad_profile.test"
	var clientType tf_api.ClientType = tf_api.Local
	if withContext {
		clientType = tf_api.Multitenancy
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyIpv6DadProfileDeleteByName(name, clientType)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyIpv6DadProfileCreate(name, clientType); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyIpv6DadProfileReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyIpv6DadProfileCreate(name string, clientType tf_api.ClientType) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.Ipv6DadProfile{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	id := newUUID()
	client := infra.NewIpv6DadProfilesClient(testAccGetContextByType(clientType), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj, nil)

	if err != nil {
		return handleCreateError("Ipv6DadProfile", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyIpv6DadProfileDeleteByName(name string, clientType tf_api.ClientType) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	// Find the object by name and delete it
	client := infra.NewIpv6DadProfilesClient(testAccGetContextByType(clientType), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("Ipv6DadProfile", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(*objInList.Id, nil)
			if err != nil {
				return handleDeleteError("Ipv6DadProfile", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting Ipv6DadProfile '%s': resource not found", name)
}

func testAccNsxtPolicyIpv6DadProfileReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtProjectContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_ipv6_dad_profile" "test" {
%s
  display_name = "%s"
}`, context, name)
}
