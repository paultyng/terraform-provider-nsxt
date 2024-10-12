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
)

func TestAccDataSourceNsxtPolicyTier1Gateway_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyTier1GatewayBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyTier1Gateway_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyTier1GatewayBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyTier1GatewayBasic(t *testing.T, withContext bool, preCheck func()) {
	routerName := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_tier1_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyTier1GatewayDeleteByName(routerName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyTier1GatewayCreate(routerName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyTier1ReadTemplate(routerName, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", routerName),
					resource.TestCheckResourceAttr(testResourceName, "description", routerName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyTier1GatewayCreate(routerName string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := routerName
	description := routerName
	obj := model.Tier1{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	id := newUUID()
	client := infra.NewTier1sClient(testAccGetSessionContext(), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj)

	if err != nil {
		return handleCreateError("Tier1", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyTier1GatewayDeleteByName(routerName string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	client := infra.NewTier1sClient(testAccGetSessionContext(), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("Tier1", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == routerName {
			err := client.Delete(*objInList.Id)
			if err != nil {
				return handleDeleteError("Tier1", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting Tier1 '%s': resource not found", routerName)
}

func testAccNsxtPolicyTier1ReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_tier1_gateway" "test" {
%s
  display_name = "%s"
}`, context, name)
}
