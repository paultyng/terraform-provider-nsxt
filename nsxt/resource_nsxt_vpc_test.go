// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var testAccNsxtVpcHelperName = getAccTestResourceName()

// shortId is limited to 8 chars
var testAccNsxtVpcShortID = testAccNsxtVpcHelperName[len(testAccNsxtVpcHelperName)-8:]
var accTestVpcCreateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform created",
	"private_ips":  "192.168.44.0/24",
	"short_id":     testAccNsxtVpcShortID,
	"enabled":      "false",
}

var accTestVpcUpdateAttributes = map[string]string{
	"display_name": getAccTestResourceName(),
	"description":  "terraform updated",
	"private_ips":  "192.168.44.0/24",
	"short_id":     testAccNsxtVpcShortID,
	"enabled":      "false",
}

func TestAccResourceNsxtVpc_basic(t *testing.T) {
	testResourceName := "nsxt_vpc.test"
	attachmentResourceName := "nsxt_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcCheckDestroy(state, accTestVpcUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcExists(accTestVpcCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcCreateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "private_ips.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "private_ips.0", accTestVpcCreateAttributes["private_ips"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address_type", model.Vpc_IP_ADDRESS_TYPE_IPV4),
					resource.TestCheckResourceAttr(testResourceName, "short_id", accTestVpcCreateAttributes["short_id"]),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.0.enabled", accTestVpcCreateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "vpc_service_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(attachmentResourceName, "parent_path"),
					resource.TestCheckResourceAttrSet(attachmentResourceName, "vpc_connectivity_profile"),
				),
			},
			{
				Config: testAccNsxtVpcTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcExists(accTestVpcUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestVpcUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestVpcUpdateAttributes["description"]),
					resource.TestCheckResourceAttr(testResourceName, "private_ips.0", accTestVpcUpdateAttributes["private_ips"]),
					resource.TestCheckResourceAttr(testResourceName, "ip_address_type", model.Vpc_IP_ADDRESS_TYPE_IPV4),
					resource.TestCheckResourceAttr(testResourceName, "short_id", accTestVpcUpdateAttributes["short_id"]),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.0.enabled", accTestVpcUpdateAttributes["enabled"]),
					resource.TestCheckResourceAttrSet(testResourceName, "vpc_service_profile"),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(attachmentResourceName, "parent_path"),
					resource.TestCheckResourceAttrSet(attachmentResourceName, "vpc_connectivity_profile"),
				),
			},
			{
				Config: testAccNsxtVpcMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcExists(accTestVpcCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "vpc_service_profile"),
					resource.TestCheckResourceAttr(testResourceName, "short_id", accTestVpcUpdateAttributes["short_id"]),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
			{
				Config: testAccNsxtVpcMinimalisticNoShortId(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtVpcExists(accTestVpcCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttrSet(testResourceName, "vpc_service_profile"),
					resource.TestCheckResourceAttr(testResourceName, "load_balancer_vpc_endpoint.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttr(testResourceName, "short_id", accTestVpcUpdateAttributes["short_id"]),
					resource.TestCheckResourceAttr(testResourceName, "nsx_id", accTestVpcUpdateAttributes["short_id"]),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtVpc_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyVPC(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtVpcCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtVpcMinimalistic(),
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

func testAccNsxtVpcExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Vpc resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Vpc resource ID not set in resources")
		}

		exists, err := resourceNsxtVpcExists(testAccGetProjectContext(), resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Vpc %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtVpcCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_vpc" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtVpcExists(testAccGetProjectContext(), resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Vpc %s still exists", displayName)
		}
	}
	return nil
}

var testAccNsxtVpcHelper = getAccTestResourceName()

func testAccNsxtVpcPrerequisites() string {
	return testAccNsxtVpcConnectivityProfilePrerequisite() + fmt.Sprintf(`
resource "nsxt_vpc_service_profile" "test" {
  %s
  display_name = "%s"
  dhcp_config {
    dhcp_server_config {
      lease_time = 15600
    }
  }
}

resource "nsxt_vpc_connectivity_profile" "test" {
  %s
  display_name = "%s"

  transit_gateway_path = data.nsxt_policy_transit_gateway.test.path
  service_gateway {
    enable = false
  }
}`, testAccNsxtProjectContext(), testAccNsxtVpcHelper, testAccNsxtProjectContext(), testAccNsxtVpcHelper)
}

func testAccNsxtVpcTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestVpcCreateAttributes
	} else {
		attrMap = accTestVpcUpdateAttributes
	}
	return testAccNsxtVpcPrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc" "test" {
  %s

  display_name = "%s"
  description  = "%s"
  private_ips  = ["%s"]
  short_id     = "%s"

  vpc_service_profile = nsxt_vpc_service_profile.test.path

  load_balancer_vpc_endpoint {
    enabled = %s
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

resource "nsxt_vpc_attachment" "test" {
  display_name             = "%s"
  parent_path              = nsxt_vpc.test.path
  vpc_connectivity_profile = nsxt_vpc_connectivity_profile.test.path
}
`, testAccNsxtProjectContext(), attrMap["display_name"], attrMap["description"], attrMap["private_ips"], attrMap["short_id"], attrMap["enabled"], attrMap["display_name"])
}

func testAccNsxtVpcMinimalistic() string {
	return testAccNsxtVpcPrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc" "test" {
  %s

  display_name = "%s"
  short_id     = "%s"
  # TODO - remove when default profiles are supported
  vpc_service_profile      = nsxt_vpc_service_profile.test.path
}`, testAccNsxtProjectContext(), accTestVpcUpdateAttributes["display_name"], accTestVpcUpdateAttributes["short_id"])
}

// We use short_id as nsx_id to make sure NSX populates the short_id correctly
func testAccNsxtVpcMinimalisticNoShortId() string {
	return testAccNsxtVpcPrerequisites() + fmt.Sprintf(`
resource "nsxt_vpc" "test" {
  %s
  nsx_id = "%s"
  display_name = "%s"
  # TODO - remove when default profiles are supported
  vpc_service_profile      = nsxt_vpc_service_profile.test.path
}`, testAccNsxtProjectContext(), accTestVpcUpdateAttributes["short_id"], accTestVpcUpdateAttributes["display_name"])
}
