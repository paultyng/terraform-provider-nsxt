/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
)

func TestAccDataSourceNsxtPolicyGroup_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyGroupBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccDataSourceNsxtPolicyGroup_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyGroupBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicyGroup_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicyGroupBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccDataSourceNsxtPolicyGroupBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestDataSourceName()
	domain := "default"
	testResourceName := "data.nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyGroupDeleteByName(domain, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyGroupCreate(domain, name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyGroupReadTemplate(domain, name, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyGroup_withSite(t *testing.T) {
	name := getAccTestDataSourceName()
	domain := getTestSiteName()
	testResourceName := "data.nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyGroupDeleteByName(domain, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyGroupCreate(domain, name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyGroupReadTemplate(domain, name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttr(testResourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccDataSourceNsxtPolicyGroupCreate(domain string, name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.Group{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	id := newUUID()

	client := domains.NewGroupsClient(testAccGetSessionContext(), connector)
	err = client.Patch(domain, id, obj)

	if err != nil {
		return handleCreateError("Group", id, err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyGroupDeleteByName(domain string, name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name and delete it
	client := domains.NewGroupsClient(testAccGetSessionContext(), connector)
	objList, err := client.List(domain, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("Group", err)
	}
	for _, objInList := range objList.Results {
		if *objInList.DisplayName == name {
			err := client.Delete(domain, *objInList.Id, nil, nil)
			if err != nil {
				return handleDeleteError("Group", *objInList.Id, err)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting Group '%s': resource not found", name)
}

func testAccNsxtPolicyGroupReadTemplate(domain string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_group" "test" {
%s
  display_name = "%s"
  domain       = "%s"
}`, context, name, domain)
}
