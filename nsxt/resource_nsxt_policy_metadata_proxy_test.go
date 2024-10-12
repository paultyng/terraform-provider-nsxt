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

var accTestPolicyMetadataProxyCreateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform created",
	"secret":         "topsecret!!",
	"server_address": "http://1.1.1.120:6000",
}

var accTestPolicyMetadataProxyUpdateAttributes = map[string]string{
	"display_name":   getAccTestResourceName(),
	"description":    "terraform updated",
	"secret":         "donottell:)",
	"server_address": "http://1.1.1.123:6000",
}

func TestAccResourceNsxtPolicyMetadataProxy_basic(t *testing.T) {
	testResourceName := "nsxt_policy_metadata_proxy.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMetadataProxyCheckDestroy(state, accTestPolicyMetadataProxyUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMetadataProxyTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyMetadataProxyExists(accTestPolicyMetadataProxyCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyMetadataProxyCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyMetadataProxyCreateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "server_address", accTestPolicyMetadataProxyCreateAttributes["server_address"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyMetadataProxyTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyMetadataProxyExists(accTestPolicyMetadataProxyUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyMetadataProxyUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyMetadataProxyUpdateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "server_address", accTestPolicyMetadataProxyUpdateAttributes["server_address"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyMetadataProxy_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_metadata_proxy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyMetadataProxyCheckDestroy(state, accTestPolicyMetadataProxyUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyMetadataProxyTemplate(true),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret"},
			},
		},
	})
}

func testAccNsxtPolicyMetadataProxyExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("PolicyMetadataProxy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("PolicyMetadataProxy resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyMetadataProxyExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("PolicyMetadataProxy %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyMetadataProxyCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_metadata_proxy" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyMetadataProxyExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("PolicyMetadataProxy %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyMetadataProxyTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyMetadataProxyCreateAttributes
	} else {
		attrMap = accTestPolicyMetadataProxyUpdateAttributes
	}
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) + fmt.Sprintf(`
resource "nsxt_policy_metadata_proxy" "test" {
  display_name      = "%s"
  description       = "%s"
  edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  secret            = "%s"
  server_address    = "%s"
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, attrMap["display_name"], attrMap["description"], attrMap["secret"], attrMap["server_address"])
}
