/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func testAccDataSourceNsxtPolicySegmentRealization(t *testing.T, vlan bool, withContext bool, preCheck func()) {
	testResourceName := "data.nsxt_policy_segment_realization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentRealizationTemplate(vlan, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "success"),
					resource.TestCheckResourceAttr(testResourceName, "network_name", "terra-test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicySegmentRealization_basic(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentRealization(t, false, false, func() {
		testAccOnlyLocalManager(t)
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
	})
}

func TestAccDataSourceNsxtPolicySegmentRealization_vlan(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentRealization(t, true, false, func() {
		testAccOnlyLocalManager(t)
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
	})
}

func TestAccDataSourceNsxtPolicySegmentRealization_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentRealization(t, false, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccDataSourceNsxtPolicySegmentRealization_multitenancyProvider(t *testing.T) {
	testAccDataSourceNsxtPolicySegmentRealization(t, false, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccNsxtPolicySegmentRealizationTemplate(vlan, withContext bool) string {
	resource := "nsxt_policy_segment"
	tz := getOverlayTransportZoneName()
	extra := ""
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	if vlan {
		resource = "nsxt_policy_vlan_segment"
		tz = getVlanTransportZoneName()
		extra = "vlan_ids = [12]"
	}
	tzSpec := ""
	tzDatasource := ""
	if !withContext {
		tzSpec = "transport_zone_path = data.nsxt_policy_transport_zone.test.path"
		tzDatasource = fmt.Sprintf(`
data "nsxt_policy_transport_zone" "test" {
  display_name = "%s"
}
`, tz)
	}
	return fmt.Sprintf(`
%s
resource "%s" "test" {
%s
  display_name        = "terra-test"
  %s
  %s
}

data "nsxt_policy_segment_realization" "test" {
%s
  path = %s.test.path
}`, tzDatasource, resource, context, tzSpec, extra, context, resource)
}
