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

func TestAccDataSourceNsxtPolicyBfdProfile_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_bfd_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyBfdProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyBfdProfileCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyBfdProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyBfdProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.BfdProfile{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	uuid, _ := uuid.NewRandom()
	id := uuid.String()

	if testAccIsGlobalManager() {
		gmObj, convErr := convertModelBindingType(obj, model.BfdProfileBindingType(), gm_model.BfdProfileBindingType())
		if convErr != nil {
			return convErr
		}

		client := gm_infra.NewBfdProfilesClient(connector)
		err = client.Patch(id, gmObj.(gm_model.BfdProfile), nil)

	} else {
		client := infra.NewBfdProfilesClient(connector)
		err = client.Patch(id, obj, nil)
	}
	if err != nil {
		return fmt.Errorf("Error during Bfd Profile creation: %v", err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyBfdProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objID, err := testGetObjIDByName(name, "BfdProfile")
	if err != nil {
		return nil
	}
	if testAccIsGlobalManager() {
		client := gm_infra.NewBfdProfilesClient(connector)
		err = client.Delete(objID, nil)
	} else {
		client := infra.NewBfdProfilesClient(connector)
		err = client.Delete(objID, nil)
	}
	if err != nil {
		return fmt.Errorf("Error during Bfd Profile deletion: %v", err)
	}
	return nil
}

func testAccNsxtPolicyBfdProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_bfd_profile" "test" {
  display_name = "%s"
}`, name)
}
