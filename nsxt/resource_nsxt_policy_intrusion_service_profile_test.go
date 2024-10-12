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

func TestAccResourceNsxtPolicyIntrusionServiceProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServiceProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.1.0")
	})
}

func TestAccResourceNsxtPolicyIntrusionServiceProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServiceProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyIntrusionServiceProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_intrusion_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIntrusionServiceProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileCreate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServiceProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "severities.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_types.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_targets.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.cvss.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.products_affected.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "overridden_signature.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "overridden_signature.0.signature_id"),
					resource.TestCheckResourceAttr(testResourceName, "overridden_signature.0.action", "REJECT"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileUpdate(updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServiceProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "severities.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_types.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_targets.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.cvss.#", "4"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.products_affected.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "overridden_signature.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileMinimalistic(updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyIntrusionServiceProfileExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "severities.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_types.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.attack_targets.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.cvss.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.products_affected.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "overridden_signature.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyIntrusionServiceProfile_importBasic(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServiceProfileImportBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.1.0")
	})
}

func TestAccResourceNsxtPolicyIntrusionServiceProfile_importBasic_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyIntrusionServiceProfileImportBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyIntrusionServiceProfileImportBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_intrusion_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyIntrusionServiceProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileMinimalistic(name, withContext),
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

func testAccNsxtPolicyIntrusionServiceProfileExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyIntrusionServiceProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy resource ID %s", resourceID)
		}
		return nil
	}
}

func testAccNsxtPolicyIntrusionServiceProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_intrusion_service_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyIntrusionServiceProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy resource %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIntrusionServiceProfileCreate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_profile" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"
  severities   = ["HIGH", "CRITICAL"]

  criteria {
    attack_types   = ["trojan-activity", "policy-violation", "attempted-admin"]
    attack_targets = ["SERVER", "SMTP Server"]
  }

  overridden_signature {
    signature_id = "2030240"
    action       = "REJECT"
    enabled      = true
  }

  tag {
    scope = "color"
    tag   = "orange"
  }
}`, context, name)
}

func testAccNsxtPolicyIntrusionServiceProfileUpdate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_profile" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"
  severities   = ["HIGH", "CRITICAL", "LOW"]

  criteria {
    attack_types      = ["trojan-activity", "policy-violation", "attempted-admin"]
    cvss              = ["NONE", "MEDIUM", "HIGH", "CRITICAL"]
    products_affected = ["Linux", "VMware"]
  }

  overridden_signature {
    signature_id = "2030240"
    action       = "REJECT"
    enabled      = true
  }

  overridden_signature {
    signature_id = "2030241"
    action       = "REJECT"
    enabled      = true
  }

  tag {
    scope = "color"
    tag   = "orange"
  }
}`, context, name)
}

func testAccNsxtPolicyIntrusionServiceProfileMinimalistic(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_intrusion_service_profile" "test" {
%s
  display_name = "%s"
  severities   = ["HIGH", "CRITICAL"]

  criteria {
    products_affected = ["Linux"]
  }

}`, context, name)
}
