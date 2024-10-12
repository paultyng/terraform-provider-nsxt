// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	lm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var accTestPolicyTransportZoneCreateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform created",
	"transport_type": lm_model.PolicyTransportZone_TZ_TYPE_VLAN_BACKED,
}

var accTestPolicyTransportZoneUpdateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform updated",
	"transport_type": lm_model.PolicyTransportZone_TZ_TYPE_OVERLAY_BACKED,
}

func TestAccResourceNsxtPolicyTransportZone_basic(t *testing.T) {
	testResourceName := "nsxt_policy_transport_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTransportZoneCheckDestroy(state, accTestPolicyTransportZoneUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTransportZoneCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTransportZoneExists(accTestPolicyTransportZoneCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyTransportZoneCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyTransportZoneCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "transport_type", accTestPolicyTransportZoneCreateAttributes["transport_type"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "is_default"),
					resource.TestCheckResourceAttrSet(testResourceName, "enforcement_point"),
					resource.TestCheckResourceAttrSet(testResourceName, "site_path"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyTransportZoneUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyTransportZoneExists(accTestPolicyTransportZoneUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyTransportZoneUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyTransportZoneUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "transport_type", accTestPolicyTransportZoneUpdateAttributes["transport_type"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "is_default"),
					resource.TestCheckResourceAttrSet(testResourceName, "enforcement_point"),
					resource.TestCheckResourceAttrSet(testResourceName, "site_path"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTransportZone_import_basic(t *testing.T) {
	testResourceName := "nsxt_policy_transport_zone.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTransportZoneCheckDestroy(state, accTestPolicyTransportZoneCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTransportZoneCreate(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNsxtPolicyTransportZoneImporterGetID,
			},
		},
	})
}

func testAccNsxtPolicyTransportZoneImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["nsxt_policy_transport_zone.test"]
	if !ok {
		return "", fmt.Errorf("PolicyTransportZone resource %s not found in resources", "nsxt_policy_transport_zone.test")
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("PolicyTransportZone resource ID not set in resources ")
	}
	path := rs.Primary.Attributes["path"]
	if path == "" {
		return "", fmt.Errorf("PolicyTransportZone path not set in resources ")
	}
	return path, nil
}

func testAccNsxtPolicyTransportZoneExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("PolicyTransportZone resource %s not found in resources", resourceName)
		}

		tzID := rs.Primary.Attributes["id"]
		if tzID == "" {
			return fmt.Errorf("PolicyTransportZone resource ID not set in resources")
		}
		epID := rs.Primary.Attributes["enforcement_point"]
		sitePath := rs.Primary.Attributes["site_path"]
		siteID := getPolicyIDFromPath(sitePath)
		exists, err := resourceNsxtPolicyTransportZoneExists(siteID, epID, tzID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("PolicyTransportZone %s does not exist", tzID)
		}

		return nil
	}
}

func testAccNsxtPolicyTransportZoneCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_transport_zone" {
			continue
		}

		tzID := rs.Primary.Attributes["id"]
		epID := rs.Primary.Attributes["enforcement_point"]
		sitePath := rs.Primary.Attributes["site_path"]
		siteID := getPolicyIDFromPath(sitePath)
		exists, err := resourceNsxtPolicyTransportZoneExists(siteID, epID, tzID, connector)
		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("PolicyTransportZone %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyTransportZoneCreate() string {
	attrMap := accTestPolicyTransportZoneCreateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
  description  = "%s"
  transport_type = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["transport_type"])
}

func testAccNsxtPolicyTransportZoneUpdate() string {
	attrMap := accTestPolicyTransportZoneUpdateAttributes
	return fmt.Sprintf(`
resource "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
  description  = "%s"
  transport_type = "%s"
}`, attrMap["display_name"], attrMap["description"], attrMap["transport_type"])
}
