// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/trust_management"
)

var accTestPrincipalIdentityCreateAttributes = map[string]string{
	"is_protected": "false",
	"name":         getAccTestResourceName(),
	"node_id":      "node-2",
	"role_path":    "/orgs/default",
	"role":         "org_admin",
}

func TestAccResourceNsxtPrincipalIdentity_basic(t *testing.T) {
	testResourceName := "nsxt_principal_identity.test"
	certPem, _, err := testAccGenerateTLSKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPrincipalIdentityCheckDestroy(state, accTestPrincipalIdentityCreateAttributes["name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPrincipalIdentityCreate(certPem),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPrincipalIdentityExists(accTestPrincipalIdentityCreateAttributes["name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "is_protected", accTestPrincipalIdentityCreateAttributes["is_protected"]),
					resource.TestCheckResourceAttr(testResourceName, "name", accTestPrincipalIdentityCreateAttributes["name"]),
					resource.TestCheckResourceAttr(testResourceName, "node_id", accTestPrincipalIdentityCreateAttributes["node_id"]),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.path", accTestPrincipalIdentityCreateAttributes["role_path"]),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "roles_for_path.0.roles.0", accTestPrincipalIdentityCreateAttributes["role"]),

					resource.TestCheckResourceAttrSet(testResourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "certificate_pem"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPrincipalIdentity_import_basic(t *testing.T) {
	testResourceName := "nsxt_principal_identity.test"
	certPem, _, err := testAccGenerateTLSKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPrincipalIdentityCheckDestroy(state, accTestPrincipalIdentityCreateAttributes["name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPrincipalIdentityCreate(certPem),
			},
			{
				ResourceName:            testResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_pem"},
			},
		},
	})
}

func testAccNsxtPrincipalIdentityExists(name string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("PrincipalIdentity resource %s not found in resources", resourceName)
		}

		piID := rs.Primary.Attributes["id"]
		if piID == "" {
			return fmt.Errorf("PrincipalIdentity resource ID not set in resources")
		}
		tmClient := trust_management.NewPrincipalIdentitiesClient(connector)
		_, err := tmClient.Get(piID)
		if err != nil {
			if isNotFoundError(err) {
				return fmt.Errorf("PrincipalIdentity %s does not exist", name)
			}
		}

		return err
	}
}

func testAccNsxtPrincipalIdentityCheckDestroy(state *terraform.State, name string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_principal_identity" {
			continue
		}

		piID := rs.Primary.Attributes["id"]
		if piID == "" {
			return fmt.Errorf("PrincipalIdentity resource ID not set in resources")
		}
		tmClient := trust_management.NewPrincipalIdentitiesClient(connector)
		_, err := tmClient.Get(piID)
		if err != nil {
			if isNotFoundError(err) {
				return nil
			}
			return err
		}
		return fmt.Errorf("PrincipalIdentity %s still exists", name)
	}
	return nil
}

func testAccNsxtPrincipalIdentityCreate(certPem string) string {
	attrMap := accTestPrincipalIdentityCreateAttributes
	return fmt.Sprintf(`
resource "nsxt_principal_identity" "test" {
    certificate_pem = <<-EOT
%s
    EOT
    is_protected    = %s
    name            = "%s"
    node_id         = "%s"

    roles_for_path {
        path  = "%s"
        roles = ["%s"]
    }
}`, certPem, attrMap["is_protected"], attrMap["name"], attrMap["node_id"], attrMap["role_path"], attrMap["role"])
}
