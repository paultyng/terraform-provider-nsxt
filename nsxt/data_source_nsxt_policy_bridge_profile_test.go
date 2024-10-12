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
	ep "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func TestAccDataSourceNsxtPolicyBridgeProfile_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_bridge_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyBridgeProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyBridgeProfileCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyBridgeProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyBridgeProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.L2BridgeEndpointProfile{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	uuid, _ := uuid.NewRandom()
	id := uuid.String()

	client := ep.NewEdgeBridgeProfilesClient(connector)
	err = client.Patch(defaultSite, defaultEnforcementPoint, id, obj)

	if err != nil {
		return fmt.Errorf("Error during Bridge Profile creation: %v", err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyBridgeProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objID, err := testGetObjIDByName(name, "L2BridgeEndpointProfile")
	if err != nil {
		return nil
	}
	client := ep.NewEdgeBridgeProfilesClient(connector)
	err = client.Delete(defaultSite, defaultEnforcementPoint, objID)
	if err != nil {
		return fmt.Errorf("Error during Bridge Profile deletion: %v", err)
	}
	return nil
}

func testAccNsxtPolicyBridgeProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_bridge_profile" "test" {
  display_name = "%s"
}`, name)
}
