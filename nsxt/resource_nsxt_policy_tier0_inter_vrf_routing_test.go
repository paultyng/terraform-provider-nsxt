// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
)

func TestAccResourceNsxtPolicyTier0InterVRFRouting_basic(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier0_inter_vrf_routing.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier0InterVRFRoutingCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier0InterVRFRoutingTemplate(name, "[\"192.168.240.0/24\", \"192.168.241.0/24\"]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "target_path"),
					resource.TestCheckResourceAttr(testResourceName, "bgp_route_leaking.0.address_family", "IPV4"),
					resource.TestCheckResourceAttr(testResourceName, "static_route_advertisement.0.advertisement_rule.0.name", "test"),
					resource.TestCheckResourceAttr(testResourceName, "static_route_advertisement.0.advertisement_rule.0.subnets.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyTier0InterVRFRoutingTemplate(updateName, "[\"192.168.240.0/24\", \"192.168.241.0/24\", \"192.168.242.0/24\"]"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "target_path"),
					resource.TestCheckResourceAttr(testResourceName, "bgp_route_leaking.0.address_family", "IPV4"),
					resource.TestCheckResourceAttr(testResourceName, "static_route_advertisement.0.advertisement_rule.0.name", "test"),
					resource.TestCheckResourceAttr(testResourceName, "static_route_advertisement.0.advertisement_rule.0.subnets.#", "3"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyTier0InterVRFRouting_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_tier0_inter_vrf_routing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyTier0InterVRFRoutingCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyTier0InterVRFRoutingTemplate(name, "[\"192.168.240.0/24\", \"192.168.241.0/24\"]"),
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

func testAccNsxtPolicyTier0InterVRFRoutingCheckDestroy(state *terraform.State, name string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	client := tier_0s.NewInterVrfRoutingClient(connector)
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_tier0_inter_vrf_routing" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		gwID := getPolicyIDFromPath(rs.Primary.Attributes["gateway_path"])
		_, err := client.Get(gwID, resourceID)
		if err == nil {
			return fmt.Errorf("policy Tier0InterVRFRouting %s still exists", name)
		}
	}
	return nil
}

func testAccNsxtPolicyTier0InterVRFRoutingTemplate(name, subnets string) string {
	rName := getAccTestResourceName()

	return testAccNsxtPolicyTier0WithVRFTemplate(rName, false, false, false) + fmt.Sprintf(`
resource "nsxt_policy_tier0_inter_vrf_routing" "test" {
  display_name = "%s"
  gateway_path = nsxt_policy_tier0_gateway.parent.path
  target_path = nsxt_policy_tier0_gateway.test.path

  bgp_route_leaking {
    address_family = "IPV4"
  }
  static_route_advertisement {
    advertisement_rule {
      name = "test"
      action = "PERMIT"
      prefix_operator = "GE"
      route_advertisement_types = ["TIER0_CONNECTED", "TIER0_NAT"]
      subnets = %s
    }
    in_filter_prefix_list = [
      join("/", [nsxt_policy_tier0_gateway.parent.path, "prefix-lists/IPSEC_LOCAL_IP"]),
      join("/", [nsxt_policy_tier0_gateway.parent.path, "prefix-lists/DNS_FORWARDER_IP"])
    ]
  }
}
`, name, subnets)
}
