// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var accTestPolicyVniPoolName = getAccTestDataSourceName()

func TestAccDataSourceNsxtPolicyVniPoolConfig_basic(t *testing.T) {
	name := accTestPolicyVniPoolName
	testResourceName := "data.nsxt_policy_vni_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyVniPoolConfigDelete()
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyVniPoolConfigReadTemplate(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "start"),
					resource.TestCheckResourceAttrSet(testResourceName, "end"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyVniPoolConfigCreate() error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewVniPoolsClient(connector)

	displayName := accTestPolicyVniPoolName
	description := accTestPolicyVniPoolName
	start := int64(75502)
	end := int64(75550)
	obj := model.VniPoolConfig{
		Description: &description,
		DisplayName: &displayName,
		Start:       &start,
		End:         &end,
	}

	// Generate a random ID for the resource
	id := newUUID()

	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("VniPoolConfig", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyVniPoolConfigDelete() error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}
	client := infra.NewVniPoolsClient(connector)

	// Find the object by name
	objList, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("VniPoolConfig", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == accTestPolicyVniPoolName {
			err := client.Delete(*objInList.Id)
			if err != nil {
				return handleDeleteError("VniPoolConfig", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting VniPoolConfig '%s': resource not found", accTestPolicyVniPoolName)
}

func testAccNsxtPolicyVniPoolConfigReadTemplate() string {
	return fmt.Sprintf(`
data "nsxt_policy_vni_pool" "test" {
  display_name = "%s"
}`, accTestPolicyVniPoolName)
}
