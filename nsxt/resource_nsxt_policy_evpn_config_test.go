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

var accTestEvpnConfigHelperName = getAccTestResourceName()

func TestAccResourceNsxtPolicyEvpnConfig_inline(t *testing.T) {
	testResourceName := "nsxt_policy_evpn_config.test"
	displayName := getAccTestResourceName()
	description := "terraform created"
	updatedDescription := "terraform updated"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnConfigCheckDestroy(state, displayName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnConfigInline(displayName, description, true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnConfigExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttr(testResourceName, "mode", "INLINE"),

					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "vni_pool_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyEvpnConfigInline(displayName, updatedDescription, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnConfigExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(testResourceName, "mode", "INLINE"),

					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "vni_pool_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyEvpnConfig_routeServer(t *testing.T) {
	testResourceName := "nsxt_policy_evpn_config.test"
	displayName := getAccTestResourceName()
	description := "terraform created"
	updatedDescription := "terraform updated"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnConfigCheckDestroy(state, displayName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnConfigRouteServer(displayName, description),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnConfigExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", description),
					resource.TestCheckResourceAttr(testResourceName, "mode", "ROUTE_SERVER"),

					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "evpn_tenant_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyEvpnConfigRouteServer(displayName, updatedDescription),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyEvpnConfigExists(displayName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", displayName),
					resource.TestCheckResourceAttr(testResourceName, "description", updatedDescription),
					resource.TestCheckResourceAttr(testResourceName, "mode", "ROUTE_SERVER"),

					resource.TestCheckResourceAttrSet(testResourceName, "gateway_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "evpn_tenant_path"),

					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyEvpnConfig_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_evpn_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if err := testAccDataSourceNsxtPolicyVniPoolConfigDelete(); err != nil {
				t.Error(err)
			}
			return testAccNsxtPolicyEvpnConfigCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyVniPoolConfigCreate(); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyEvpnConfigInline(name, "", false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyEvpnConfigIDGenerator(testResourceName),
			},
		},
	})
}

func testAccNsxtPolicyEvpnConfigExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Gateway Route Map resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Gateway Route Map resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		_, err := policyEvpnConfigGet(connector, gwID)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccNsxtPolicyEvpnConfigCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_evpn_config" {
			continue
		}

		gwPath := rs.Primary.Attributes["gateway_path"]
		_, gwID := parseGatewayPolicyPath(gwPath)

		_, err := policyEvpnConfigGet(connector, gwID)
		if err == nil {
			return fmt.Errorf("Policy Gateway Route Map %s still exists", displayName)
		}
	}
	return nil
}

func testAccEvpnConfigPrerequisites() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) +
		testAccNsxtPolicyTier0WithEdgeClusterTemplate("test", false) +
		testAccNsxtPolicyEvpnTenantPrerequisites()
}

func testAccNsxtPolicyEvpnConfigInline(displayName string, description string, withTags bool) string {
	tags := ""
	if withTags {
		tags = `
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
            `
	}
	return testAccEvpnConfigPrerequisites() + fmt.Sprintf(`

resource "nsxt_policy_evpn_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  description  = "%s"

  mode          = "INLINE"
  vni_pool_path = data.nsxt_policy_vni_pool.test.path
  %s
}`, displayName, description, tags)
}

func testAccNsxtPolicyEvpnConfigRouteServer(displayName string, description string) string {
	return testAccEvpnConfigPrerequisites() + fmt.Sprintf(`

resource "nsxt_policy_evpn_tenant" "test" {
  display_name        = "%s"
  vni_pool_path       = data.nsxt_policy_vni_pool.test.path
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  mapping {
    vnis  = "75503"
    vlans = "103"
  }
}

resource "nsxt_policy_evpn_config" "test" {
  gateway_path = nsxt_policy_tier0_gateway.test.path
  display_name = "%s"
  description  = "%s"

  mode             = "ROUTE_SERVER"
  evpn_tenant_path = nsxt_policy_evpn_tenant.test.path
}`, accTestEvpnConfigHelperName, displayName, description)
}

func testAccNSXPolicyEvpnConfigIDGenerator(testResourceName string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[testResourceName]
		if !ok {
			return "", fmt.Errorf("NSX Policy resource %s not found in resources", testResourceName)
		}
		gwPath := rs.Primary.Attributes["gateway_path"]
		if gwPath == "" {
			return "", fmt.Errorf("NSX Policy Gateway Path not set in resources ")
		}

		return gwPath, nil
	}
}
