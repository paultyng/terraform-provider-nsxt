/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"testing"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TODO - include display name in metadata
var createDisplayName = getAccTestResourceName()
var updateDisplayName = getAccTestResourceName()

func TestAccResourceNsxtPolicyMacDiscoveryProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyMacDiscoveryProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
	})
}

func TestAccResourceNsxtPolicyMacDiscoveryProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyMacDiscoveryProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyMacDiscoveryProfile_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyMacDiscoveryProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyMultitenancyProvider(t)
	})
}

func getMacDiscoveryProfileTestCheckFunc(testResourceName string, create bool) []resource.TestCheckFunc {
	displayName := createDisplayName
	if !create {
		displayName = updateDisplayName
	}
	result := []resource.TestCheckFunc{testAccNsxtPolicyMacDiscoveryProfileExists(displayName, testResourceName),
		resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
		resource.TestCheckResourceAttrSet(testResourceName, "path"),
		resource.TestCheckResourceAttrSet(testResourceName, "revision"),
		resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
	}
	for key, item := range macDiscoveryProfileSchema {
		if item.Metadata.Skip {
			continue
		}
		if item.Metadata.IntroducedInVersion != "" && util.NsxVersionLower(item.Metadata.IntroducedInVersion) {
			continue
		}

		value := item.Metadata.TestData.CreateValue
		if !create {
			value = item.Metadata.TestData.UpdateValue
		}
		result = append(result, resource.TestCheckResourceAttr(testResourceName, key, value.(string)))
	}

	return result
}

func getMacDiscoveryProfileTestConfigAttributes(create bool) string {
	result := ""
	schemaDef := resourceNsxtPolicyMacDiscoveryProfile().Schema
	for key, item := range macDiscoveryProfileSchema {
		if item.Metadata.Skip {
			continue
		}
		log.Printf("[INFO] inspecting schema test key %s", key)
		if item.Metadata.IntroducedInVersion != "" && util.NsxVersionLower(item.Metadata.IntroducedInVersion) {
			continue
		}

		value := item.Metadata.TestData.CreateValue
		if !create {
			value = item.Metadata.TestData.UpdateValue
		}

		if schemaDef[key].Type == schema.TypeString {
			result += fmt.Sprintf("\n %s = \"%s\"", key, value.(string))
		} else {
			result += fmt.Sprintf("\n %s = %s", key, value.(string))
		}
	}

	return result
}

func testAccResourceNsxtPolicyMacDiscoveryProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMacDiscoveryProfileCheckDestroy(state, updateDisplayName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileTemplate(true, withContext),
				Check:  resource.ComposeTestCheckFunc(getMacDiscoveryProfileTestCheckFunc(testResourceName, true)...),
			},
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileTemplate(false, withContext),
				Check:  resource.ComposeTestCheckFunc(getMacDiscoveryProfileTestCheckFunc(testResourceName, false)...),
			},
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileMinimalistic(withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyMacDiscoveryProfileExists(createDisplayName, testResourceName),
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

func TestAccResourceNsxtPolicyMacDiscoveryProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMacDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileMinimalistic(false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyMacDiscoveryProfile_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMacDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileMinimalistic(true),
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

func TestAccResourceNsxtPolicyMacDiscoveryProfile_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_mac_discovery_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMacDiscoveryProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMacDiscoveryProfileMinimalistic(false),
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

func testAccNsxtPolicyMacDiscoveryProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy MacDiscoveryProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy MacDiscoveryProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyMacDiscoveryProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy MacDiscoveryProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyMacDiscoveryProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_mac_discovery_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyMacDiscoveryProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy MacDiscoveryProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyMacDiscoveryProfileTemplate(createFlow, withContext bool) string {
	displayName := createDisplayName
	if !createFlow {
		displayName = updateDisplayName
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_mac_discovery_profile" "test" {
%s
  display_name = "%s"
  %s
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, displayName, getMacDiscoveryProfileTestConfigAttributes(createFlow))
}

func testAccNsxtPolicyMacDiscoveryProfileMinimalistic(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_mac_discovery_profile" "test" {
%s
  display_name = "%s"
}`, context, updateDisplayName)
}
