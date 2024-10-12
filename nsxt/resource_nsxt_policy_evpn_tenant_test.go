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

var accTestPolicyEvpnTenantCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
}

var accTestPolicyEvpnTenantUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
}

func TestAccResourceNsxtPolicyEvpnTenant_basic(t *testing.T) {
	testResourceName := "nsxt_policy_evpn_tenant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnTenantCheckDestroy(state, accTestPolicyEvpnTenantUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnTenantCreate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnTenantExists(accTestPolicyEvpnTenantCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyEvpnTenantCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyEvpnTenantCreateAttributes["description"]),

					resource.TestCheckResourceAttrSet(testResourceName, "transport_zone_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "vni_pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "mapping.#", "2"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyEvpnTenantUpdate(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnTenantExists(accTestPolicyEvpnTenantUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyEvpnTenantUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyEvpnTenantUpdateAttributes["description"]),

					resource.TestCheckResourceAttrSet(testResourceName, "transport_zone_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "vni_pool_path"),
					resource.TestCheckResourceAttr(testResourceName, "mapping.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyEvpnTenant_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_evpn_tenant.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnTenantCheckDestroy(state, accTestPolicyEvpnTenantUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnTenantUpdate(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyEvpnTenantExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Evpn Tenant resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Evpn Tenant resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyEvpnTenantExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Evpn Tenant %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyEvpnTenantCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_evpn_tenant" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyEvpnTenantExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Evpn Tenant %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyEvpnTenantPrerequisites() string {
	return testAccNsxtPolicyVniPoolConfigReadTemplate() + fmt.Sprintf(`
    data "nsxt_policy_transport_zone" "test" {
      display_name = "%s"
    }
    `, getOverlayTransportZoneName())
}

func testAccNsxtPolicyEvpnTenantCreate() string {
	attrMap := accTestPolicyEvpnTenantCreateAttributes
	return testAccNsxtPolicyEvpnTenantPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_evpn_tenant" "test" {
  display_name = "%s"
  description  = "%s"

  vni_pool_path       = data.nsxt_policy_vni_pool.test.path
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  mapping {
    vnis  = "75503"
    vlans = "103"
  }

  mapping {
    vnis  = "75512-75515"
    vlans = "112-115"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, attrMap["display_name"], attrMap["description"])
}

func testAccNsxtPolicyEvpnTenantUpdate() string {
	attrMap := accTestPolicyEvpnTenantUpdateAttributes
	return testAccNsxtPolicyEvpnTenantPrerequisites() + fmt.Sprintf(`
resource "nsxt_policy_evpn_tenant" "test" {
  display_name = "%s"
  description  = "%s"

  vni_pool_path       = data.nsxt_policy_vni_pool.test.path
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  mapping {
    vnis  = "75505"
    vlans = "100"
  }
}`, attrMap["display_name"], attrMap["description"])
}
