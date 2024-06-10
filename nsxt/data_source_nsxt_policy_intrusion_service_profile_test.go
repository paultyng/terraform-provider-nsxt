/* Copyright © 2021 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtPolicyIntrusionServiceProfile_default(t *testing.T) {
	name := "DefaultIDSProfile"
	testResourceName := "data.nsxt_policy_intrusion_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "3.0.0")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileReadTemplate(fmt.Sprintf("\"%s\"", name), false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyIntrusionServiceProfile_basic(t *testing.T) {
	testAccDataSourceNsxtPolicyIntrusionServiceProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
		testAccNSXVersion(t, "3.0.0")
	})
}

func TestAccDataSourceNsxtPolicyIntrusionServiceProfile_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyIntrusionServiceProfileBasic(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicyIntrusionServiceProfile_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicyIntrusionServiceProfileBasic(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccDataSourceNsxtPolicyIntrusionServiceProfileBasic(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_intrusion_service_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIntrusionServiceProfileMinimalistic(name, withContext) + testAccNsxtPolicyIntrusionServiceProfileReadTemplate("nsxt_policy_intrusion_service_profile.test.display_name", withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyIntrusionServiceProfileReadTemplate(name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "nsxt_policy_intrusion_service_profile" "test" {
%s
  display_name = %s
}`, context, name)
}
