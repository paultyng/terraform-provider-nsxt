// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func TestAccDataSourceNsxtPolicyLBAppProfile_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_lb_app_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyLBAppProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyLBAppProfileCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyLBAppProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch test profile by name only */
				Config: testAccNsxtPolicyLBAppProfileNameOnlyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch default HTTP profile */
				Config: testAccNsxtPolicyLBAppProfileTypeOnlyTemplate("HTTP"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "type", "HTTP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyLBAppProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewLbAppProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	displayName := name
	description := name
	profileType := "LBFastTcpProfile"
	obj := model.LBFastTcpProfile{
		Description:  &description,
		DisplayName:  &displayName,
		ResourceType: profileType,
	}

	dataValue, errs := converter.ConvertToVapi(obj, model.LBFastTcpProfileBindingType())
	if errs != nil {
		return fmt.Errorf("Error during conversion of LBFastTcpProfile: %v", errs[0])
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(id, dataValue.(*data.StructValue))
	if err != nil {
		return handleCreateError("LBAppProfile", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyLBAppProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewLbAppProfilesClient(connector)

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("LBAppProfile", err)
	}
	force := true
	for _, objInList := range objList.Results {
		result, err := policyLbAppProfileConvert(objInList, "ANY")
		if err != nil {
			return fmt.Errorf("Error during LBAppProfile conversion: %v", err)
		}
		if result != nil && *result.DisplayName == name {
			err := client.Delete(*result.Id, &force)
			if err != nil {
				return handleDeleteError("LBAppProfile", *result.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting LBAppProfile '%s': resource not found", name)
}

func testAccNsxtPolicyLBAppProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  type = "TCP"
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyLBAppProfileTypeOnlyTemplate(pType string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  type = "%s"
}`, pType)
}

func testAccNsxtPolicyLBAppProfileNameOnlyTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_app_profile" "test" {
  display_name = "%s"
}`, name)
}
