/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtPolicyODSRunbookInvocation_basic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ods_runbook_invocation.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.2.0")
			testAccEnvDefined(t, "NSXT_TEST_HOST_TRANSPORT_NODE")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyODSRunbookInvocationCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyODSRunbookInvocationCreateTemplate(name, "OverlayTunnel", `
  argument {
    key = "src"
    value = "192.168.0.11"
  }
  argument {
    key = "dst"
    value = "192.168.0.10" 
  }
`),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyODSRunbookInvocationExists(name, testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "target_node"),
					resource.TestCheckResourceAttrSet(testResourceName, "runbook_path"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyODSRunbookInvocation_import(t *testing.T) {

	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ods_runbook_invocation.test"
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.2.0")
			testAccOnlyLocalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_HOST_TRANSPORT_NODE")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyODSRunbookInvocationCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyODSRunbookInvocationCreateTemplate(name, "OverlayTunnel", `
  argument {
    key = "src"
    value = "192.168.0.11"
  }
  argument {
    key = "dst"
    value = "192.168.0.10" 
  }
`),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyODSRunbookInvocationCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ods_runbook_invocation" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyODSRunbookInvocationExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("policy ODSRunbookInvocation %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyODSRunbookInvocationExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("policy ODSRunbookInvocation resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("policy ODSRunbookInvocation resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyODSRunbookInvocationExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("policy ODSRunbookInvocation %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyODSRunbookInvocationCreateTemplate(name, runbook, arguments string) string {
	htnName := getHostTransportNodeName()
	return testAccNsxtPolicyODSPredefinedRunbookReadTemplate(runbook) + fmt.Sprintf(`
data "nsxt_policy_host_transport_node" "test" {
  display_name = "%s"
}

resource "nsxt_policy_ods_runbook_invocation" "test" {
  display_name = "%s"
  runbook_path = data.nsxt_policy_ods_pre_defined_runbook.test.path
%s
  target_node = data.nsxt_policy_host_transport_node.test.unique_id
}
`, htnName, name, arguments)
}
