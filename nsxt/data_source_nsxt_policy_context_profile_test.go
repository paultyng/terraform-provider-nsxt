/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyContextProfile_basic(t *testing.T) {
	// Use existing system defined profile
	name := "AMQP"
	testResourceName := "data.nsxt_policy_context_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyContextProfile_prefix(t *testing.T) {
	// Use existing system defined profile
	name := "DIAMETER"
	namePrefix := name[0:5]
	testResourceName := "data.nsxt_policy_context_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileReadTemplate(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyContextProfile_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_context_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancy(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileMultitenancyTemplate(name, true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyContextProfile_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_context_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyMultitenancyProvider(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyContextProfileMultitenancyTemplate(name, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "description"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyContextProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_context_profile" "test" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicyContextProfileMultitenancyTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	s := fmt.Sprintf(`
resource "nsxt_policy_context_profile" "test" {
%s
  display_name = "%s"
  description  = "Terraform provisioned ContextProfile"
  domain_name {
    description = "test-domain-name-attribute"
    value       = ["*-myfiles.sharepoint.com"]
  }
  app_id {
    description = "test-app-id-attribute"
    value       = ["SSL"]
    sub_attribute {
      tls_version = ["SSL_V3"]
    }
  }
}

data "nsxt_policy_context_profile" "test" {
%s
  display_name = nsxt_policy_context_profile.test.display_name
}`, context, name, context)
	return s
}
