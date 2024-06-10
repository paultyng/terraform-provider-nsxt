/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func TestAccDataSourceNsxtPolicySegment_basic(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicySegment_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicySegment_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccDataSourceNsxtPolicySegmentBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicySegmentDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicySegmentCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicySegmentReadTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicySegmentCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.Segment{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	uuid, _ := uuid.NewRandom()
	id := uuid.String()

	client := infra.NewSegmentsClient(testAccGetSessionContext(), connector)
	err = client.Patch(id, obj)
	if err != nil {
		return fmt.Errorf("Error during Segment creation: %v", err)
	}
	return nil
}

func testAccDataSourceNsxtPolicySegmentDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objID, err := testGetObjIDByName(name, "Segment")
	if err != nil {
		return nil
	}
	client := infra.NewSegmentsClient(testAccGetSessionContext(), connector)
	err = client.Delete(objID)

	if err != nil {
		return fmt.Errorf("Error during Segment deletion: %v", err)
	}
	return nil
}

func testAccNsxtPolicySegmentReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_segment" "test" {
%s
  display_name = "%s"
}`, context, name)
}
