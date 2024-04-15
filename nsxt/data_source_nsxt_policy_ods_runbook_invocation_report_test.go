/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceNsxtPolicyODSRunbookInvocationReport_basic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "data.nsxt_policy_ods_runbook_invocation_report.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.2.0")
			testAccEnvDefined(t, "NSXT_TEST_HOST_TRANSPORT_NODE")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyODSRunbookInvocationReportReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(testResourceName, "target_node"),
					resource.TestCheckResourceAttrSet(testResourceName, "request_status"),
					resource.TestCheckResourceAttrSet(testResourceName, "operation_state"),
				),
			},
		},
	})
}

func testAccNsxtPolicyODSRunbookInvocationReportReadTemplate(name string) string {
	return testAccNsxtPolicyODSRunbookInvocationCreateTemplate(name, "ControllerConn", "") + `
data "nsxt_policy_ods_runbook_invocation_report" "test" {
  invocation_id = nsxt_policy_ods_runbook_invocation.test.id
}`
}
