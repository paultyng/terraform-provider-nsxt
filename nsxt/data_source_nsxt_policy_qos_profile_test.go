/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func TestAccDataSourceNsxtPolicyQosProfile_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyQosProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyQosProfile_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyQosProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicyQosProfile_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicyQosProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccDataSourceNsxtPolicyQosProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_qos_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyQosProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyQosProfileCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyQosProfileReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyQosProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.QosProfile{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	id := newUUID()

	client := infra.NewQosProfilesClient(testAccGetSessionContext(), connector)
	err = client.Patch(id, obj, nil)

	if err != nil {
		return handleCreateError("QosProfile", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyQosProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name and delete it
	client := infra.NewQosProfilesClient(testAccGetSessionContext(), connector)
	objList, err := client.List(nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("QosProfile", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(*objInList.Id, nil)
			if err != nil {
				return handleDeleteError("QosProfile", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting QosProfile '%s': resource not found", name)
}

func testAccNsxtPolicyQosProfileReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_qos_profile" "test" {
%s
  display_name = "%s"
}`, context, name)
}
