// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtLbCookiePersistenceProfile_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_cookie_persistence_profile.test"
	mode := "PREFIX"
	updatedMode := "REWRITE"
	cookieName := "my_cookie"
	updatedCookieName := "new_cookie"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbCookiePersistenceProfileCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbCookiePersistenceProfileBasicTemplate(name, mode, cookieName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbCookiePersistenceProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_mode", mode),
					resource.TestCheckResourceAttr(testResourceName, "cookie_name", cookieName),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_fallback", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_garble", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbCookiePersistenceProfileBasicTemplate(updatedName, updatedMode, updatedCookieName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbCookiePersistenceProfileExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_mode", updatedMode),
					resource.TestCheckResourceAttr(testResourceName, "cookie_name", updatedCookieName),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_fallback", "true"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_garble", "false"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbCookiePersistenceProfile_insertMode(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_cookie_persistence_profile.test"
	cookieName := "my_cookie"
	updatedCookieName := "new_cookie"
	cookieDomain := ".example.com"
	updatedCookieDomain := ".anotherexample.com"
	cookiePath := "/subfolder1"
	updatedCookiePath := "/subfolder1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbCookiePersistenceProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbCookiePersistenceProfileInsertTemplate(name, cookieName, cookieDomain, cookiePath),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbCookiePersistenceProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_mode", "INSERT"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_name", cookieName),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "false"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_fallback", "false"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_garble", "true"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_domain", cookieDomain),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_path", cookiePath),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_expiry_type", "SESSION_COOKIE_TIME"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.max_idle_time", "1000"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.max_life_time", "2000"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXLbCookiePersistenceProfileInsertTemplate(name, updatedCookieName, updatedCookieDomain, updatedCookiePath),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbCookiePersistenceProfileExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "test description"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_mode", "INSERT"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_name", updatedCookieName),
					resource.TestCheckResourceAttr(testResourceName, "persistence_shared", "false"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_fallback", "false"),
					resource.TestCheckResourceAttr(testResourceName, "cookie_garble", "true"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_domain", updatedCookieDomain),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_path", updatedCookiePath),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.cookie_expiry_type", "SESSION_COOKIE_TIME"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.max_idle_time", "1000"),
					resource.TestCheckResourceAttr(testResourceName, "insert_mode_params.0.max_life_time", "2000"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbCookiePersistenceProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_cookie_persistence_profile.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbCookiePersistenceProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbCookiePersistenceProfileCreateTemplateTrivial(name),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbCookiePersistenceProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB cookie persistence profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB cookie persistence profile resource ID not set in resources ")
		}

		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerCookiePersistenceProfile(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB cookie persistence profile with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB cookie persistence profile %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == profile.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB cookie persistence profile %s wasn't found", displayName)
	}
}

func testAccNSXLbCookiePersistenceProfileCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_cookie_persistence_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		profile, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerCookiePersistenceProfile(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB cookie persistence profile with ID %s. Error: %v", resourceID, err)
		}

		if displayName == profile.DisplayName {
			return fmt.Errorf("NSX LB cookie persistence profile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbCookiePersistenceProfileBasicTemplate(name string, mode string, cookieName string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_cookie_persistence_profile" "test" {
  description        = "test description"
  display_name       = "%s"
  persistence_shared = "true"
  cookie_fallback    = "true"
  cookie_garble      = "false"
  cookie_mode        = "%s"
  cookie_name        = "%s"

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, mode, cookieName)
}

func testAccNSXLbCookiePersistenceProfileInsertTemplate(name string, cookieName string, cookieDomain string, cookiePath string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_cookie_persistence_profile" "test" {
  description        = "test description"
  display_name       = "%s"
  persistence_shared = "false"
  cookie_fallback    = "false"
  cookie_garble      = "true"
  cookie_name        = "%s"
  cookie_mode        = "INSERT"

  insert_mode_params {
  	cookie_domain      = "%s"
  	cookie_path        = "%s"
  	cookie_expiry_type = "SESSION_COOKIE_TIME"
  	max_idle_time      = "1000"
  	max_life_time      = "2000"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, cookieName, cookieDomain, cookiePath)
}

func testAccNSXLbCookiePersistenceProfileCreateTemplateTrivial(name string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_cookie_persistence_profile" "test" {
  display_name = "%s"
  description  = "test description"
  cookie_name  = "my_cookie"
}
`, name)
}
