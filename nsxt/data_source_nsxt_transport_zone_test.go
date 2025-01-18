// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceNsxtTransportZone_basic(t *testing.T) {
	transportZoneName := getVlanTransportZoneName()
	testResourceName := "data.nsxt_transport_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNSXTransportZoneReadTemplate(transportZoneName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", transportZoneName),
					resource.TestCheckResourceAttrSet(testResourceName, "id"),
					resource.TestCheckResourceAttrSet(testResourceName, "host_switch_name"),
					resource.TestCheckResourceAttrSet(testResourceName, "transport_type"),
				),
			},
		},
	})
}

func testAccNSXTransportZoneReadTemplate(transportZoneName string) string {
	return fmt.Sprintf(`
data "nsxt_transport_zone" "test" {
  display_name = "%s"
}`, transportZoneName)
}
