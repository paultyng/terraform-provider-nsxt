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

var accTestPolicyOspfAreaCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"area_id":      "15",
	"area_type":    "NORMAL",
	"auth_mode":    "MD5",
	"secret_key":   "banana",
	"key_id":       "15",
}

var accTestPolicyOspfAreaUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"area_id":      "17.2.2.2",
	"area_type":    "NSSA",
	"auth_mode":    "PASSWORD",
	"secret_key":   "banana",
	"key_id":       "",
}

var accTestPolicyOspfAreaHelperName = getAccTestResourceName()

func TestAccResourceNsxtPolicyOspfArea_basic(t *testing.T) {
	testResourceName := "nsxt_policy_ospf_area.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.1.1") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyOspfAreaCheckDestroy(state, accTestPolicyOspfAreaCreateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyOspfAreaTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyOspfAreaExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfAreaCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyOspfAreaCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "area_id", accTestPolicyOspfAreaCreateAttributes["area_id"]),
					resource.TestCheckResourceAttr(testResourceName, "area_type", accTestPolicyOspfAreaCreateAttributes["area_type"]),
					resource.TestCheckResourceAttr(testResourceName, "auth_mode", accTestPolicyOspfAreaCreateAttributes["auth_mode"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyOspfAreaTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyOspfAreaExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfAreaUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyOspfAreaUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "area_id", accTestPolicyOspfAreaUpdateAttributes["area_id"]),
					resource.TestCheckResourceAttr(testResourceName, "area_type", accTestPolicyOspfAreaUpdateAttributes["area_type"]),
					resource.TestCheckResourceAttr(testResourceName, "auth_mode", accTestPolicyOspfAreaUpdateAttributes["auth_mode"]),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyOspfAreaMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyOspfAreaExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyOspfAreaUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "area_id", accTestPolicyOspfAreaUpdateAttributes["area_id"]),
					resource.TestCheckResourceAttr(testResourceName, "auth_mode", "NONE"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyOspfArea_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_ospf_area.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.1.1") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyOspfAreaCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyOspfAreaMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyOspfAreaImporterGetIDs,
			},
		},
	})
}

func testAccNSXPolicyOspfAreaImporterGetIDs(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources["nsxt_policy_ospf_area.test"]
	if !ok {
		return "", fmt.Errorf("NSX Policy OSPF Area resource %s not found in resources", "nsxt_policy_ospf_area.test")
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy OSPF Area resource ID not set in resources ")
	}
	ospfPath := rs.Primary.Attributes["ospf_path"]
	if ospfPath == "" {
		return "", fmt.Errorf("NSX Policy OSPF Area ospf_path not set in resources ")
	}
	gwID, serviceID := parseOspfConfigPath(ospfPath)
	return fmt.Sprintf("%s/%s/%s", gwID, serviceID, resourceID), nil
}

func testAccNsxtPolicyOspfAreaExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Ospf Area resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Ospf Area resource ID not set in resources")
		}

		ospfPath := rs.Primary.Attributes["ospf_path"]
		gwID, serviceID := parseOspfConfigPath(ospfPath)

		exists, err := resourceNsxtPolicyOspfAreaExists(gwID, serviceID, resourceID, testAccIsGlobalManager(), connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Ospf Area ID %s does not exist on backend", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyOspfAreaCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_ospf_area" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		ospfPath := rs.Primary.Attributes["ospf_path"]
		gwID, serviceID := parseOspfConfigPath(ospfPath)
		exists, err := resourceNsxtPolicyOspfAreaExists(gwID, serviceID, resourceID, testAccIsGlobalManager(), connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy Ospf Area %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyOspfAreaPrerequisites() string {
	return testAccNsxtPolicyOspfConfigPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_ospf_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
}`, accTestPolicyOspfAreaHelperName)
}

func testAccNsxtPolicyOspfAreaTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyOspfAreaCreateAttributes
	} else {
		attrMap = accTestPolicyOspfAreaUpdateAttributes
	}
	extraConfig := ""
	if attrMap["key_id"] != "" {
		extraConfig = fmt.Sprintf(`key_id = "%s"`, attrMap["key_id"])
	}
	return testAccNsxtPolicyOspfAreaPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_ospf_area" "test" {
  display_name = "%s"
  description  = "%s"
  ospf_path    = nsxt_policy_ospf_config.test.path
  area_id      = "%s"
  area_type    = "%s"
  auth_mode    = "%s"
  secret_key   = "%s"
  %s

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"], attrMap["area_id"], attrMap["area_type"], attrMap["auth_mode"], attrMap["secret_key"], extraConfig)
}

func testAccNsxtPolicyOspfAreaMinimalistic() string {
	return testAccNsxtPolicyOspfAreaPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_ospf_area" "test" {
  ospf_path    = nsxt_policy_ospf_config.test.path
  display_name = "%s"
  area_id      = "%s"
}`, accTestPolicyOspfAreaUpdateAttributes["display_name"], accTestPolicyOspfAreaUpdateAttributes["area_id"])
}
