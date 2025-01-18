// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceNsxtVPCGroup_basicImport(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_vpc_group"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVPCGroupCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVPCGroupAddressCreateTemplate(name),
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

func TestAccResourceNsxtVPCGroup_addressCriteria(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_vpc_group"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyVPC(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVPCGroupAddressCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "context.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "context.0.project_id", os.Getenv("NSXT_VPC_PROJECT_ID")),
					resource.TestCheckResourceAttr(testResourceName, "context.0.vpc_id", os.Getenv("NSXT_VPC_ID")),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
				),
			},
			{
				Config: testAccNsxtVPCGroupAddressUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "context.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "context.0.project_id", os.Getenv("NSXT_VPC_PROJECT_ID")),
					resource.TestCheckResourceAttr(testResourceName, "context.0.vpc_id", os.Getenv("NSXT_VPC_ID")),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.0.ip_addresses.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.macaddress_expression.#", "0"),
				),
			},
		},
	})
}

func testAccNsxtVPCGroupCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc_group" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyGroupExistsInDomain(testAccGetSessionContext(), resourceID, "", connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("policy Group %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtVPCGroupAddressCreateTemplate(name string) string {
	context := testAccNsxtPolicyMultitenancyContext()
	return fmt.Sprintf(`
resource "nsxt_vpc_group" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1", "222.2.2.2"]
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    ipaddress_expression {
	  ip_addresses = ["122.1.1.1"]
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}

data "nsxt_vpc_group" "test" {
%s
  display_name = "%s"

  depends_on = [nsxt_vpc_group.test]
}
`, context, name, context, name)
}

func testAccNsxtVPCGroupAddressUpdateTemplate(name string) string {
	context := testAccNsxtPolicyMultitenancyContext()
	return fmt.Sprintf(`
resource "nsxt_vpc_group" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["122.1.1.1"]
    }
  }
}
`, context, name)
}
