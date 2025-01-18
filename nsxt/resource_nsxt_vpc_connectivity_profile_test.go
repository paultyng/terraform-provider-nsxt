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

var accTestVpcConnectivityProfileCreateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform created",
	"enable":              "false",
	"enable_default_snat": "false",
}

var accTestVpcConnectivityProfileUpdateAttributes = map[string]string{
	"display_name":        getAccTestResourceName(),
	"description":         "terraform updated",
	"enable":              "false",
	"enable_default_snat": "false",
}

func TestAccResourceNsxtVpcConnectivityProfile_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_connectivity_profile.test"
	testDataSourceName := "nsxt_vpc_connectivity_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcConnectivityProfileCheckDestroy(state, accTestVpcConnectivityProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcConnectivityProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcConnectivityProfileExists(accTestVpcConnectivityProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcConnectivityProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcConnectivityProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "transit_gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.nat_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.enable", accTestVpcConnectivityProfileCreateAttributes["enable"]),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.nat_config.0.enable_default_snat", accTestVpcConnectivityProfileCreateAttributes["enable_default_snat"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcConnectivityProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcConnectivityProfileExists(accTestVpcConnectivityProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testDataSourceName, "path"),
					resource.TestCheckResourceAttrSet(testDataSourceName, "description"),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcConnectivityProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcConnectivityProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttrSet(testResourceName, "transit_gateway_path"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.nat_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.enable", accTestVpcConnectivityProfileUpdateAttributes["enable"]),
					resource.TestCheckResourceAttr(testResourceName, "service_gateway.0.nat_config.0.enable_default_snat", accTestVpcConnectivityProfileUpdateAttributes["enable_default_snat"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcConnectivityProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcConnectivityProfileExists(accTestVpcConnectivityProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVpcConnectivityProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_connectivity_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcConnectivityProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcConnectivityProfileMinimalistic(),
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

func testAccNsxtVpcConnectivityProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy VpcConnectivityProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy VpcConnectivityProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtVpcConnectivityProfileExists(testAccGetProjectContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy VpcConnectivityProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcConnectivityProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_connectivity_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVpcConnectivityProfileExists(testAccGetProjectContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy VpcConnectivityProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVpcConnectivityProfilePrerequisite() string {
	//TODO: replace datasource with resource when transit GW creation is enabled
	return fmt.Sprintf(`
data "nsxt_policy_transit_gateway" "test" {
%s
  id = "default"
}
`, testAccNsxtProjectContext())
}

func testAccNsxtVpcConnectivityProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcConnectivityProfileCreateAttributes
	} else {
		attrMap = accTestVpcConnectivityProfileUpdateAttributes
	}
	return testAccNsxtVpcConnectivityProfilePrerequisite() + fmt.Sprintf(`
resource "nsxt_vpc_connectivity_profile" "test" {
%s
  display_name = "%s"
  description  = "%s"
  transit_gateway_path = data.nsxt_policy_transit_gateway.test.path

  service_gateway {
    nat_config {
      enable_default_snat = %s
    }

    enable = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
data "nsxt_vpc_connectivity_profile" "test" {
%s
  display_name = "%s"

  depends_on = [nsxt_vpc_connectivity_profile.test]
}`, testAccNsxtProjectContext(), attrMap["display_name"], attrMap["description"], attrMap["enable_default_snat"], attrMap["enable"], testAccNsxtProjectContext(), attrMap["display_name"])
}

func testAccNsxtVpcConnectivityProfileMinimalistic() string {
	return testAccNsxtVpcConnectivityProfilePrerequisite() + fmt.Sprintf(`
resource "nsxt_vpc_connectivity_profile" "test" {
%s
  display_name         = "%s"
  transit_gateway_path = data.nsxt_policy_transit_gateway.test.path

}`, testAccNsxtProjectContext(), accTestVpcConnectivityProfileUpdateAttributes["display_name"])
}
