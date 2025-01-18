// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func getGroupPrefix() string {
	groupPrefix := "/infra/domains/default/groups/"
	if testAccIsGlobalManager() {
		groupPrefix = "/global-infra/domains/default/groups/"
	}
	return groupPrefix
}

func TestAccResourceNsxtPolicyFirewallExcludeListMember_basic(t *testing.T) {
	testResourcePfx := "nsxt_policy_firewall_exclude_list_member."
	names := []string{getAccTestResourceName(), getAccTestResourceName(), getAccTestResourceName(), getAccTestResourceName()}
	groupPrefix := getGroupPrefix()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			for _, name := range names {
				err := testAccNsxtPolicyFirewallExcludeListMemberCheckDestroy(state, groupPrefix+name)
				if err != nil {
					return err
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				// Add resource
				Config: testAccNsxtPolicyFirewallExcludeListMemberTemplate(names[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFirewallExcludeListMemberExists(names[0], testResourcePfx+names[0]),
					resource.TestCheckResourceAttr(testResourcePfx+names[0], "member", groupPrefix+names[0]),
				),
			},
			{
				// Additional resource
				Config: testAccNsxtPolicyFirewallExcludeListMemberTemplate(names[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFirewallExcludeListMemberExists(names[1], testResourcePfx+names[1]),
					resource.TestCheckResourceAttr(testResourcePfx+names[1], "member", groupPrefix+names[1]),
				),
			},
			{
				// Two more concurrently
				Config: testAccNsxtPolicyFirewallExcludeListMemberTemplate(names[2]) + testAccNsxtPolicyFirewallExcludeListMemberTemplate(names[3]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFirewallExcludeListMemberExists(names[2], testResourcePfx+names[2]),
					testAccNsxtPolicyFirewallExcludeListMemberExists(names[3], testResourcePfx+names[3]),
					resource.TestCheckResourceAttr(testResourcePfx+names[2], "member", groupPrefix+names[2]),
					resource.TestCheckResourceAttr(testResourcePfx+names[3], "member", groupPrefix+names[3]),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyFirewallExcludeListMember_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_firewall_exclude_list_member." + name

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFirewallExcludeListMemberCheckDestroy(state, getGroupPrefix()+name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFirewallExcludeListMemberTemplate(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyFirewallExcludeListMemberExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("policy FirewallExcludeListMember resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("policy FirewallExcludeListMember resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyFirewallExcludeListMemberExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("policy FirewallExcludeListMember %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyFirewallExcludeListMemberCheckDestroy(state *terraform.State, member string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_firewall_exclude_list_member" {
			continue
		}

		exists, err := resourceNsxtPolicyFirewallExcludeListMemberExists(testAccGetSessionContext(), member, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("policy FirewallExcludeListMember %s still exists", member)
		}
	}
	return nil
}

func testAccNsxtPolicyFirewallExcludeListMemberTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "%s" {
  nsx_id = "%s"
  display_name = "%s"
  description  = "Acceptance Test"
}
resource "nsxt_policy_firewall_exclude_list_member" "%s" {
	member = nsxt_policy_group.%s.path
}
`, name, name, name, name, name)
}
