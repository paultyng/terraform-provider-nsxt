// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyService_basic(t *testing.T) {
	serviceName := "ICMPv6-ALL"
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceReadTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyService_byId(t *testing.T) {
	serviceID := "ICMPv6-ALL"
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceReadIDTemplate(serviceID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceID),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceID),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyService_byPrefix(t *testing.T) {
	serviceName := "Heartbeat"
	servicePrefix := "Heart"
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceReadTemplate(servicePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyService_spaces(t *testing.T) {
	serviceName := "Enterprise Manager Servlet port SSL"
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceReadTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyService_specialChars(t *testing.T) {
	serviceName1 := "IKE (Key Exchange)"
	serviceName2 := "EdgeSync service/ADAM"
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceReadTemplate(serviceName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName1),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName1),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
			{
				Config: testAccNsxtPolicyServiceReadTemplate(serviceName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName2),
					resource.TestCheckResourceAttr(testResourceName, "description", serviceName2),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyService_multitenancy(t *testing.T) {
	serviceName := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyServiceMultitenancyTemplate(serviceName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", serviceName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyServiceReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_service" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyServiceReadIDTemplate(id string) string {
	return fmt.Sprintf(`
data "nsxt_policy_service" "test" {
  id = "%s"
}`, id)
}

func testAccNsxtPolicyServiceMultitenancyTemplate(name string) string {
	context := testAccNsxtPolicyMultitenancyContext()
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  icmp_entry {
	display_name = "%s"
	description  = "Entry"
	icmp_type    = "3"
	icmp_code    = "1"
	protocol     = "ICMPv4"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_service" "test" {
%s
  display_name = nsxt_policy_service.test.display_name
}`, context, name, name, context)
}
