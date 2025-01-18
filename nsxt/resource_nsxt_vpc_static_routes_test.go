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

var accTestStaticRoutesCreateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform created",
	"network":        "2.2.2.0/24",
	"ip_address":     "3.1.1.1",
	"admin_distance": "2",
}

var accTestStaticRoutesUpdateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform updated",
	"network":        "3.3.3.0/24",
	"ip_address":     "4.1.1.1",
	"admin_distance": "5",
}

func TestAccResourceNsxtVpcStaticRoutes_basic(t *testing.T) {
	testResourceName := "nsxt_vpc_static_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcStaticRoutesCheckDestroy(state, accTestStaticRoutesUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcStaticRoutesTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcStaticRoutesExists(accTestStaticRoutesCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestStaticRoutesCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestStaticRoutesCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "network", accTestStaticRoutesCreateAttributes["network"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.0.ip_address", accTestStaticRoutesCreateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.0.admin_distance", accTestStaticRoutesCreateAttributes["admin_distance"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcStaticRoutesTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcStaticRoutesExists(accTestStaticRoutesUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestStaticRoutesUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestStaticRoutesUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "network", accTestStaticRoutesUpdateAttributes["network"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.0.ip_address", accTestStaticRoutesUpdateAttributes["ip_address"]),
					resource.TestCheckResourceAttr(testResourceName, "next_hop.0.admin_distance", accTestStaticRoutesUpdateAttributes["admin_distance"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtVpcStaticRoutesMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcStaticRoutesExists(accTestStaticRoutesCreateAttributes["display_name"], testResourceName),
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

func TestAccResourceNsxtVpcStaticRoutes_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc_static_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcStaticRoutesCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcStaticRoutesMinimalistic(),
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

func testAccNsxtVpcStaticRoutesExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy StaticRoutes resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy StaticRoutes resource ID not set in resources")
		}

		exists, err := resourceNsxtVpcStaticRoutesExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy StaticRoutes %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcStaticRoutesCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_static_route" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVpcStaticRoutesExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy StaticRoutes %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVpcStaticRoutesTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestStaticRoutesCreateAttributes
	} else {
		attrMap = accTestStaticRoutesUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_vpc_static_route" "test" {
%s
  display_name = "%s"
  description  = "%s"

  network = "%s"

  next_hop {
    ip_address     = "%s"
    admin_distance = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, testAccNsxtPolicyMultitenancyContext(), attrMap["display_name"], attrMap["description"], attrMap["network"], attrMap["ip_address"], attrMap["admin_distance"])
}

func testAccNsxtVpcStaticRoutesMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_vpc_static_route" "test" {
%s
  display_name = "%s"
  network = "%s"
  next_hop {
    ip_address     = "%s"
  }
}`, testAccNsxtPolicyMultitenancyContext(), accTestStaticRoutesUpdateAttributes["display_name"], accTestStaticRoutesUpdateAttributes["network"], accTestStaticRoutesUpdateAttributes["ip_address"])
}
