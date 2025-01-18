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

var accTestPolicyLBClientSslProfileCreateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform created",
	"cipher_group_label":    "CUSTOM",
	"ciphers":               "TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	"prefer_server_ciphers": "true",
	"protocols":             "TLS_V1_2",
	"session_cache_enabled": "true",
	"session_cache_timeout": "2",
}

var accTestPolicyLBClientSslProfileUpdateAttributes = map[string]string{
	"display_name":          getAccTestResourceName(),
	"description":           "terraform updated",
	"cipher_group_label":    "CUSTOM",
	"ciphers":               "TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
	"prefer_server_ciphers": "false",
	"protocols":             "TLS_V1_2",
	"session_cache_enabled": "false",
	"session_cache_timeout": "5",
}

func TestAccResourceNsxtPolicyLBClientSslProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_lb_client_ssl_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBClientSslProfileCheckDestroy(state, accTestPolicyLBClientSslProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBClientSslProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBClientSslProfileExists(accTestPolicyLBClientSslProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBClientSslProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBClientSslProfileCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "cipher_group_label", accTestPolicyLBClientSslProfileCreateAttributes["cipher_group_label"]),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.0", accTestPolicyLBClientSslProfileCreateAttributes["ciphers"]),
					resource.TestCheckResourceAttrSet(testResourceName, "is_fips"),
					resource.TestCheckResourceAttrSet(testResourceName, "is_secure"),
					resource.TestCheckResourceAttr(testResourceName, "prefer_server_ciphers", accTestPolicyLBClientSslProfileCreateAttributes["prefer_server_ciphers"]),
					resource.TestCheckResourceAttr(testResourceName, "protocols.0", accTestPolicyLBClientSslProfileCreateAttributes["protocols"]),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", accTestPolicyLBClientSslProfileCreateAttributes["session_cache_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_timeout", accTestPolicyLBClientSslProfileCreateAttributes["session_cache_timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBClientSslProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBClientSslProfileExists(accTestPolicyLBClientSslProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestPolicyLBClientSslProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestPolicyLBClientSslProfileUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "cipher_group_label", accTestPolicyLBClientSslProfileUpdateAttributes["cipher_group_label"]),
					resource.TestCheckResourceAttr(testResourceName, "ciphers.0", accTestPolicyLBClientSslProfileUpdateAttributes["ciphers"]),
					resource.TestCheckResourceAttrSet(testResourceName, "is_fips"),
					resource.TestCheckResourceAttrSet(testResourceName, "is_secure"),
					resource.TestCheckResourceAttr(testResourceName, "prefer_server_ciphers", accTestPolicyLBClientSslProfileUpdateAttributes["prefer_server_ciphers"]),
					resource.TestCheckResourceAttr(testResourceName, "protocols.0", accTestPolicyLBClientSslProfileUpdateAttributes["protocols"]),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_enabled", accTestPolicyLBClientSslProfileUpdateAttributes["session_cache_enabled"]),
					resource.TestCheckResourceAttr(testResourceName, "session_cache_timeout", accTestPolicyLBClientSslProfileUpdateAttributes["session_cache_timeout"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyLBClientSslProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyLBClientSslProfileExists(accTestPolicyLBClientSslProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyLBClientSslProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_lb_client_ssl_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyLBClientSslProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyLBClientSslProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyLBClientSslProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy LBClientSslProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy LBClientSslProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyLBClientSslProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy LBClientSslProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyLBClientSslProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_lb_client_ssl_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyLBClientSslProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy LBClientSslProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyLBClientSslProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyLBClientSslProfileCreateAttributes
	} else {
		attrMap = accTestPolicyLBClientSslProfileUpdateAttributes
	}
	return fmt.Sprintf(`
   resource "nsxt_policy_lb_client_ssl_profile" "test" {
	 display_name = "%s"
	 description  = "%s"
	 cipher_group_label = "%s"
	 ciphers = ["%s"]
	 prefer_server_ciphers = %s
	 protocols = ["%s"]
	 session_cache_enabled = %s
	 session_cache_timeout = %s
	 tag {
	   scope = "scope1"
	   tag   = "tag1"
	 }
   }`, attrMap["display_name"], attrMap["description"], attrMap["cipher_group_label"], attrMap["ciphers"], attrMap["prefer_server_ciphers"], attrMap["protocols"], attrMap["session_cache_enabled"], attrMap["session_cache_timeout"])
}

func testAccNsxtPolicyLBClientSslProfileMinimalistic() string {
	return fmt.Sprintf(`
   resource "nsxt_policy_lb_client_ssl_profile" "test" {
	 display_name = "%s"
   }`, accTestPolicyLBClientSslProfileUpdateAttributes["display_name"])
}
