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
)

var testAccNsxtNSGroupHelperName = getAccTestResourceName()

func TestAccResourceNsxtNSGroup_basic(t *testing.T) {
	grpName := getAccTestResourceName()
	updateGrpName := getAccTestResourceName()
	testResourceName := "nsxt_ns_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, updateGrpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupCreateTemplate(grpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
			{
				Config: testAccNSXNSGroupUpdateTemplate(updateGrpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(updateGrpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateGrpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNSGroup_nested(t *testing.T) {
	grpName := getAccTestResourceName()
	updateGrpName := getAccTestResourceName()
	testResourceName := "nsxt_ns_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, updateGrpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupNestedCreateTemplate(grpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "1"),
				),
			},
			{
				Config: testAccNSXNSGroupNestedUpdateTemplate(updateGrpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(updateGrpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateGrpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNSGroup_withCriteria(t *testing.T) {
	grpName := getAccTestResourceName()
	testResourceName := "nsxt_ns_group.test"
	transportZoneName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, grpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupCriteriaCreateTemplate(grpName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "membership_criteria.#", "2"),
				),
			},
			{
				Config: testAccNSXNSGroupCriteriaUpdateTemplate(grpName, transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXNSGroupExists(grpName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", grpName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test Update"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "membership_criteria.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtNSGroup_importBasic(t *testing.T) {
	grpName := getAccTestResourceName()
	testResourceName := "nsxt_ns_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, grpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupCreateTemplate(grpName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtNSGroup_importWithCriteria(t *testing.T) {
	grpName := getAccTestResourceName()
	testResourceName := "nsxt_ns_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestDeprecated(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXNSGroupCheckDestroy(state, grpName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXNSGroupCriteriaCreateTemplate(grpName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXNSGroupExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NS Group resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NS Group resource ID not set in resources ")
		}

		localVarOptionals := make(map[string]interface{})
		group, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, resourceID, localVarOptionals)
		if err != nil {
			return fmt.Errorf("Error while retrieving NS Group ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if NS Group %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == group.DisplayName {
			return nil
		}
		return fmt.Errorf("NS Group %s wasn't found", displayName)
	}
}

func testAccNSXNSGroupCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_ns_group" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		localVarOptionals := make(map[string]interface{})
		group, responseCode, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, resourceID, localVarOptionals)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving NS Group ID %s. Error: %v", resourceID, err)
		}

		if displayName == group.DisplayName {
			return fmt.Errorf("NS Group %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXNSGroupCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name)
}

func testAccNSXNSGroupUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, updatedName)
}

func testAccNSXNSGroupNestedCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "grp1" {
  display_name = "grp1"
}

resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  member {
    target_type = "NSGroup"
    value       = "${nsxt_ns_group.grp1.id}"
  }
}`, name)
}

func testAccNSXNSGroupNestedUpdateTemplate(updatedName string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "grp1" {
  display_name = "grp1"
}

resource "nsxt_ns_group" "grp2" {
  display_name = "grp2"
}

resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test Update"

  member {
    target_type = "NSGroup"
    value       = "${nsxt_ns_group.grp1.id}"
  }

  member {
    target_type = "NSGroup"
    value       = "${nsxt_ns_group.grp2.id}"
  }
}`, updatedName)
}

func testAccNSXNSGroupCriteriaCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  membership_criteria {
    target_type = "LogicalSwitch"
    scope       = "XXX"
  }

  membership_criteria {
    target_type = "LogicalPort"
    scope       = "XXX"
    tag         = "YYY"
  }
}`, name)
}

func testAccNSXNSGroupCriteriaUpdateTemplate(name string, tzName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "tz1" {
  display_name = "%s"
}

resource "nsxt_logical_switch" "test" {
  display_name      = "%s"
  admin_state       = "DOWN"
  replication_mode  = "MTEP"
  transport_zone_id = "${data.nsxt_transport_zone.tz1.id}"
}

resource "nsxt_ns_group" "test" {
  display_name = "%s"
  description = "Acceptance Test Update"

  membership_criteria {
	target_type = "LogicalSwitch"
    scope       = "XXX"
  }

  member {
    target_type = "LogicalSwitch"
    value = "${nsxt_logical_switch.test.id}"
  }
}`, tzName, testAccNsxtNSGroupHelperName, name)
}
