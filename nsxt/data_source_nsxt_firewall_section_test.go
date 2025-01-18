// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func TestAccDataSourceNsxtFirewallSection_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_firewall_section.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtFirewallSectionDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtFirewallSectionCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNSXFirewallSectionReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
				),
			},
		},
	})
}

func testAccDataSourceNsxtFirewallSectionCreate(name string) error {
	nsxClient, err := testAccGetClient()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	section := manager.FirewallSection{
		Description: name,
		DisplayName: name,
		SectionType: "LAYER3",
	}

	_, responseCode, err := nsxClient.ServicesApi.AddSection(nsxClient.Context, section, nil)
	if err != nil {
		return fmt.Errorf("Error during firewall section creation: %v", err)
	}

	if responseCode.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during firewall section creation: %v", responseCode.StatusCode)
	}
	return nil
}

func testAccDataSourceNsxtFirewallSectionDeleteByName(name string) error {
	nsxClient, err := testAccGetClient()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objList, _, err := nsxClient.ServicesApi.ListSections(nsxClient.Context, nil)
	if err != nil {
		return fmt.Errorf("Error while reading Firewall sections: %v", err)
	}
	// go over the list to find the correct one
	for _, objInList := range objList.Results {
		if objInList.DisplayName == name {
			responseCode, err := nsxClient.ServicesApi.DeleteSection(nsxClient.Context, objInList.Id, nil)
			if err != nil {
				return fmt.Errorf("Error during firewall section deletion: %v", err)
			}

			if responseCode.StatusCode != http.StatusOK {
				return fmt.Errorf("Unexpected status returned during firewall section deletion: %v", responseCode.StatusCode)
			}
			return nil
		}
	}
	return fmt.Errorf("Error while deleting Firewall section '%s': firewall section not found", name)
}

func testAccNSXFirewallSectionReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_firewall_section" "test" {
  display_name = "%s"
}`, name)
}
