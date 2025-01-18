// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
)

var accTestFailureDomainCreateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform created",
	"preferred_edge_services": "active",
}

var accTestFailureDomainUpdateAttributes = map[string]string{
	"display_name":            getAccTestResourceName(),
	"description":             "terraform created",
	"preferred_edge_services": "standby",
}

func TestAccResourceNsxtFailureDomain_basic(t *testing.T) {
	testResourceName := "nsxt_failure_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtFailureDomainCheckDestroy(state, accTestFailureDomainUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtFailureDomainTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtFailureDomainExists(accTestFailureDomainCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestFailureDomainCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestFailureDomainCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "preferred_edge_services", accTestFailureDomainCreateAttributes["preferred_edge_services"]),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtFailureDomainTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtFailureDomainExists(accTestFailureDomainUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestFailureDomainUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestFailureDomainUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "preferred_edge_services", accTestFailureDomainUpdateAttributes["preferred_edge_services"]),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtFailureDomainMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtFailureDomainExists(accTestFailureDomainCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtFailureDomain_importBasic(t *testing.T) {
	testResourceName := "nsxt_failure_domain.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtFailureDomainCheckDestroy(state, accTestFailureDomainUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtFailureDomainMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtFailureDomainExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("FailureDomain resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("FailureDomain resource ID not set in resources")
		}

		client := nsx.NewFailureDomainsClient(connector)
		_, err := client.Get(resourceID)

		if isNotFoundError(err) {
			return fmt.Errorf("FailureDomain %s does not exist", resourceID)
		}

		if err != nil {
			return fmt.Errorf("error while retrieving Failure Domain ID %s, error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtFailureDomainCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_failure_domain" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		client := nsx.NewFailureDomainsClient(connector)
		obj, err := client.Get(resourceID)

		if isNotFoundError(err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while retrieving Failure Domain ID %s. Error: %v", resourceID, err)
		}

		if obj.DisplayName != nil && displayName == *obj.DisplayName {
			return fmt.Errorf("FailureDomain %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtFailureDomainTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestFailureDomainCreateAttributes
	} else {
		attrMap = accTestFailureDomainUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_failure_domain" "test" {
  display_name = "%s"
  description  = "%s"
  preferred_edge_services = "%s"
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_failure_domain" "test" {
  display_name = "%s"
  depends_on = [nsxt_failure_domain.test]
}`, attrMap["display_name"], attrMap["description"], attrMap["preferred_edge_services"], attrMap["display_name"])
}

func testAccNsxtFailureDomainMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_failure_domain" "test" {
  display_name = "%s"
}`, accTestUplinkHostSwitchProfileUpdateAttributes["display_name"])
}
