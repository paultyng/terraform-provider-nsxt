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
	orgs "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs"
)

func TestAccDataSourceNsxtPolicyProject_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyProjectDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyProjectCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyProjectReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttr(testResourceName, "tier0_gateway_paths.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyProjectCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	var tier0s = []string{getTier0RouterPath(connector)}
	client := orgs.NewProjectsClient(connector)

	displayName := name
	description := name
	obj := model.Project{
		Description: &description,
		DisplayName: &displayName,
		Tier0s:      tier0s,
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(defaultOrgID, id, obj)
	if err != nil {
		return handleCreateError("Project", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyProjectDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := orgs.NewProjectsClient(connector)

	// Find the object by name
	objList, err := client.List(defaultOrgID, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("Project", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(defaultOrgID, *objInList.Id, nil)
			if err != nil {
				return handleDeleteError("Project", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting Project '%s': resource not found", name)
}

func testAccNsxtPolicyProjectReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_project" "test" {
  display_name = "%s"
}`, name)
}
