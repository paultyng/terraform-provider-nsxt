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

func TestAccResourceNsxtPolicyQosProfile_basic(t *testing.T) {
	testAccResourceNsxtPolicyQosProfileBasic(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyQosProfile_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyQosProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyQosProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_qos_profile.test"
	cos := "5"
	updatedCos := "2"
	peak := "700"
	updatedPeak := "400"

	// There is a NSX bug that messes up dependencies for QoS profiles
	// until the bug fixed, QoS tests need to avoid parallelization
	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXQosSwitchingProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyQosProfileBasicTemplate(name, cos, peak, "ingress", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyQosProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "class_of_service", cos),
					resource.TestCheckResourceAttr(testResourceName, "dscp_trusted", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dscp_priority", "53"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.average_bw_mbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_rate_shaper.0.peak_bw_mbps", peak),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.average_bw_kbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.peak_bw_kbps", peak),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXPolicyQosProfileBasicTemplate(updatedName, updatedCos, updatedPeak, "egress", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyQosProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "class_of_service", updatedCos),
					resource.TestCheckResourceAttr(testResourceName, "dscp_trusted", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dscp_priority", "53"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.average_bw_mbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.0.peak_bw_mbps", updatedPeak),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.average_bw_kbps", "111"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.burst_size", "222"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.peak_bw_kbps", updatedPeak),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXPolicyQosProfileUpdateTemplate(updatedName, updatedCos, updatedPeak, "egress", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyQosProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "class_of_service", updatedCos),
					resource.TestCheckResourceAttr(testResourceName, "dscp_trusted", "true"),
					resource.TestCheckResourceAttr(testResourceName, "dscp_priority", "53"),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igress_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.average_bw_kbps", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.burst_size", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.0.peak_bw_kbps", updatedPeak),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
			{
				Config: testAccNSXPolicyQosProfileEmptyTemplate(updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXPolicyQosProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "egress_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ingress_broadcast_rate_shaper.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyQosProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_qos_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyQosProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyQosProfileCreateTemplateTrivial(name, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyQosProfile_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_qos_profile.test"
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXPolicyQosProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXPolicyQosProfileCreateTemplateTrivial(name, true),
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

func testAccNSXPolicyQosProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy QosProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy QosProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyQosProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy QosProfile ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNSXPolicyQosProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_qos_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyQosProfileExists(testAccGetSessionContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy QosProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyQosProfileBasicTemplate(name string, cos string, peak string, direction string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_qos_profile" "test" {
%s
  display_name     = "%s"
  description      = "test description"
  class_of_service = %s
  dscp_trusted     = true
  dscp_priority    = 53

  %s_rate_shaper {
    average_bw_mbps = 111
    burst_size      = 222
    peak_bw_mbps    = "%s"
  }

  ingress_broadcast_rate_shaper {
    average_bw_kbps = 111
    burst_size      = 222
    peak_bw_kbps    = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, context, name, cos, direction, peak, peak)
}

func testAccNSXPolicyQosProfileUpdateTemplate(name string, cos string, peak string, direction string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_qos_profile" "test" {
%s
  display_name     = "%s"
  description      = "test description"
  class_of_service = %s
  dscp_trusted     = true
  dscp_priority    = 53

  ingress_broadcast_rate_shaper {
    peak_bw_kbps = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}
`, context, name, cos, peak)
}

func testAccNSXPolicyQosProfileEmptyTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_qos_profile" "test" {
%s
  display_name     = "%s"
}
`, context, name)
}

func testAccNSXPolicyQosProfileCreateTemplateTrivial(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_qos_profile" "test" {
%s
  display_name = "%s"
  description  = "test description"
  dscp_trusted = false

  egress_rate_shaper {
    enabled         = false
    peak_bw_mbps    = 800
    burst_size      = 222
    average_bw_mbps = 111
  }
}
`, context, name)
}
