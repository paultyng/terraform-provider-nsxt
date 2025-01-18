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

func TestAccDataSourceNsxtPolicyLBMonitor_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_lb_monitor.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyLBMonitorDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyLBMonitorCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyLBMonitorReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch test profile by name only */
				Config: testAccNsxtPolicyLBMonitorNameOnlyTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "type", "TCP"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				/* fetch default HTTP profile */
				Config: testAccNsxtPolicyLBMonitorTypeOnlyTemplate("PASSIVE"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "display_name"),
					resource.TestCheckResourceAttr(testResourceName, "type", "PASSIVE"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyLBMonitorCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewLbMonitorProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	displayName := name
	description := name
	profileType := model.LBMonitorProfile_RESOURCE_TYPE_LBTCPMONITORPROFILE
	obj := model.LBFastTcpProfile{
		Description:  &description,
		DisplayName:  &displayName,
		ResourceType: profileType,
	}

	dataValue, errs := converter.ConvertToVapi(obj, model.LBTcpMonitorProfileBindingType())
	if errs != nil {
		return fmt.Errorf("Error during conversion of LBTcpMonitor: %v", errs[0])
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(id, dataValue.(*data.StructValue))
	if err != nil {
		return handleCreateError("LBMonitor", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyLBMonitorDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewLbMonitorProfilesClient(connector)

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("LBMonitor", err)
	}
	force := true
	for _, objInList := range objList.Results {
		result, err := policyLbMonitorConvert(objInList, "ANY")
		if err != nil {
			return fmt.Errorf("Error during LBMonitor conversion: %v", err)
		}
		if result != nil && *result.DisplayName == name {
			err := client.Delete(*result.Id, &force)
			if err != nil {
				return handleDeleteError("LBMonitor", *result.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting LBMonitor '%s': resource not found", name)
}

func testAccNsxtPolicyLBMonitorReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_monitor" "test" {
  type = "TCP"
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyLBMonitorTypeOnlyTemplate(pType string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_monitor" "test" {
  type = "%s"
}`, pType)
}

func testAccNsxtPolicyLBMonitorNameOnlyTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_lb_monitor" "test" {
  display_name = "%s"
}`, name)
}
