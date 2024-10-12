// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func TestAccDataSourceNsxtPolicyTier0Gateway_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_tier0_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyTier0GatewayDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyTier0GatewayCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyTier0GatewayReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyTier0GatewayCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.Tier0{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	uuid, _ := uuid.NewRandom()
	id := uuid.String()

	if testAccIsGlobalManager() {
		gmObj, convErr := convertModelBindingType(obj, model.Tier0BindingType(), gm_model.Tier0BindingType())
		if convErr != nil {
			return convErr
		}

		client := gm_infra.NewTier0sClient(connector)
		err = client.Patch(id, gmObj.(gm_model.Tier0))

	} else {
		client := infra.NewTier0sClient(connector)
		err = client.Patch(id, obj)
	}
	if err != nil {
		return fmt.Errorf("Error during Tier0 creation: %v", err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyTier0GatewayDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	if testAccIsGlobalManager() {
		objID, err := testGetObjIDByName(name, "Tier0")
		if err == nil {
			client := gm_infra.NewTier0sClient(connector)
			err := client.Delete(objID)
			if err != nil {
				return handleDeleteError("Tier0", objID, err)
			}
			return nil
		}
	} else {
		client := infra.NewTier0sClient(connector)
		objList, err := client.List(nil, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("Error while reading Tier0s: %v", err)
		}
		for _, objInList := range objList.Results {
			if *objInList.DisplayName == name {
				err := client.Delete(*objInList.Id)
				if err != nil {
					return fmt.Errorf("Error during Tier0 deletion: %v", err)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("Error while deleting Tier0 '%s': resource not found", name)
}

func testAccNsxtPolicyTier0GatewayReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_tier0_gateway" "test" {
  display_name = "%s"
}`, name)
}
