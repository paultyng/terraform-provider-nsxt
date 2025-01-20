/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"bytes"
	"fmt"
	"strconv"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"
)

var shortID = getAccTestRandomString(6)

var accTestPolicyProjectCreateAttributes = map[string]string{
	"DisplayName": getAccTestResourceName(),
	"Description": "terraform created",
	"ShortId":     shortID,
}

var accTestPolicyProjectUpdateAttributes = map[string]string{
	"DisplayName": getAccTestResourceName(),
	"Description": "terraform updated",
	"ShortId":     shortID,
}

func getExpectedSiteInfoCount(t *testing.T) string {
	if util.NsxVersion == "" {
		connector, err := testAccGetPolicyConnector()
		if err != nil {
			t.Errorf("Failed to get policy connector")
			return "0"
		}

		err = initNSXVersion(connector)
		if err != nil {
			t.Errorf("Failed to retrieve NSX version")
			return "0"
		}
	}
	if util.NsxVersionHigherOrEqual("4.1.1") {
		return "1"
	}
	return "0"
}

func runChecksNsx410(testResourceName string, attributes map[string]string, expectedValues map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		testAccNsxtPolicyProjectExists(accTestPolicyProjectCreateAttributes["DisplayName"], testResourceName),
		resource.TestCheckResourceAttr(testResourceName, "display_name", attributes["DisplayName"]),
		resource.TestCheckResourceAttr(testResourceName, "description", attributes["Description"]),
		resource.TestCheckResourceAttr(testResourceName, "short_id", attributes["ShortId"]),
		//TODO: add site info validation
		resource.TestCheckResourceAttr(testResourceName, "site_info.#", expectedValues["site_count"]),
		resource.TestCheckResourceAttr(testResourceName, "tier0_gateway_paths.#", expectedValues["t0_count"]),

		resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
		resource.TestCheckResourceAttrSet(testResourceName, "path"),
		resource.TestCheckResourceAttrSet(testResourceName, "revision"),
		resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
	)
}

func runChecksNsx411(testResourceName string, expectedValues map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, "external_ipv4_blocks.#", expectedValues["ip_block_count"]),
		resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
	)
}

func runChecksNsx420(testResourceName string, expectedValues map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, "activate_default_dfw_rules", expectedValues["activate_default_dfw_rules"]),
		resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
	)
}

func runChecksNsx900(testResourceName string, expectedValues map[string]string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(testResourceName, "tgw_external_connections.#", expectedValues["tgw_ext_conn_count"]),
		resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
	)
}

func TestAccResourceNsxtPolicyProject_basic(t *testing.T) {
	testResourceName := "nsxt_policy_project.test"
	siteCount := getExpectedSiteInfoCount(t)
	expectedValuesStep1 := map[string]string{
		"t0_count":   "1",
		"site_count": siteCount,
	}
	expectedValuesStep2 := expectedValuesStep1
	expectedValuesStep3 := map[string]string{
		"t0_count":   "0",
		"site_count": siteCount,
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.0")
			testAccNSXVersionLessThan(t, "4.1.1")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, accTestPolicyProjectUpdateAttributes["DisplayName"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectTemplate410(true, true),
				Check:  runChecksNsx410(testResourceName, accTestPolicyProjectCreateAttributes, expectedValuesStep1),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate410(false, true),
				Check:  runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep2),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate410(false, false),
				Check:  runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep3),
			},
			{
				Config: testAccNsxtPolicyProjectMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectExists(accTestPolicyProjectCreateAttributes["DisplayName"], testResourceName),
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

func TestAccResourceNsxtPolicyProject_411basic(t *testing.T) {
	testResourceName := "nsxt_policy_project.test"
	siteCount := getExpectedSiteInfoCount(t)
	expectedValuesStep1 := map[string]string{
		"t0_count":       "1",
		"ip_block_count": "1",
		"site_count":     siteCount,
	}
	expectedValuesStep2 := expectedValuesStep1
	expectedValuesStep3 := map[string]string{
		"t0_count":       "0",
		"ip_block_count": "0",
		"site_count":     siteCount,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.1")
			testAccNSXVersionLessThan(t, "4.2.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, accTestPolicyProjectUpdateAttributes["DisplayName"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectTemplate411(true, true, true),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectCreateAttributes, expectedValuesStep1),
					runChecksNsx411(testResourceName, expectedValuesStep1),
				),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate411(false, true, true),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep2),
					runChecksNsx411(testResourceName, expectedValuesStep2),
				),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate411(false, false, false),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep3),
					runChecksNsx411(testResourceName, expectedValuesStep3),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyProject_420basic(t *testing.T) {
	testResourceName := "nsxt_policy_project.test"
	siteCount := getExpectedSiteInfoCount(t)
	expectedValuesStep1 := map[string]string{
		"t0_count":                   "1",
		"ip_block_count":             "1",
		"site_count":                 siteCount,
		"activate_default_dfw_rules": "false",
	}
	expectedValuesStep2 := expectedValuesStep1
	expectedValuesStep3 := map[string]string{
		"t0_count":                   "0",
		"ip_block_count":             "0",
		"site_count":                 siteCount,
		"activate_default_dfw_rules": "true",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.2.0")
			testAccNSXVersionLessThan(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, accTestPolicyProjectUpdateAttributes["DisplayName"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectTemplate420(true, true, true, false),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectCreateAttributes, expectedValuesStep1),
					runChecksNsx411(testResourceName, expectedValuesStep1),
					runChecksNsx420(testResourceName, expectedValuesStep1),
				),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate420(false, true, true, false),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep2),
					runChecksNsx411(testResourceName, expectedValuesStep2),
					runChecksNsx420(testResourceName, expectedValuesStep2),
				),
			},
			{
				Config: testAccNsxtPolicyProjectTemplate420(false, false, false, true),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep3),
					runChecksNsx411(testResourceName, expectedValuesStep3),
					runChecksNsx420(testResourceName, expectedValuesStep3),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyProject_900basic(t *testing.T) {
	testResourceName := "nsxt_policy_project.test"
	siteCount := getExpectedSiteInfoCount(t)
	expectedValuesStep1 := map[string]string{
		"t0_count":                   "1",
		"ip_block_count":             "1",
		"tgw_ext_conn_count":         "1",
		"activate_default_dfw_rules": "true",
		"site_count":                 siteCount,
	}
	expectedValuesStep2 := map[string]string{
		"t0_count":                   "1",
		"ip_block_count":             "0",
		"tgw_ext_conn_count":         "1",
		"activate_default_dfw_rules": "false",
		"site_count":                 siteCount,
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, accTestPolicyProjectUpdateAttributes["DisplayName"])
		},
		Steps: []resource.TestStep{
			{
				// Create: Set T0, Ext GW connection, Ext IPv4 Block, activate default DFW
				Config: testAccNsxtPolicyProjectTemplate900(true, true, true, true),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectCreateAttributes, expectedValuesStep1),
					runChecksNsx411(testResourceName, expectedValuesStep1),
					runChecksNsx420(testResourceName, expectedValuesStep1),
					runChecksNsx900(testResourceName, expectedValuesStep1),
				),
			},
			{
				// Update: Set T0, Ext GW connection, No Ext IPv4 Block, disable default DFW
				Config: testAccNsxtPolicyProjectTemplate900(false, true, false, false),
				Check: resource.ComposeTestCheckFunc(
					runChecksNsx410(testResourceName, accTestPolicyProjectUpdateAttributes, expectedValuesStep2),
					runChecksNsx411(testResourceName, expectedValuesStep2),
					runChecksNsx420(testResourceName, expectedValuesStep2),
					runChecksNsx900(testResourceName, expectedValuesStep2),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyProject_900defaultSecurityProfile(t *testing.T) {
	testResourceName := "nsxt_policy_project.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "9.0.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, accTestPolicyProjectCreateAttributes["DisplayName"])
		},
		Steps: []resource.TestStep{
			{
				// Create: Set T0, Ext GW connection, Ext IPv4 Block, activate default DFW
				Config: testAccNsxtPolicyProjectMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectGetSecurityProfileNSEnabled(testResourceName, true),
				),
			},
			{
				// Update: Set T0, Ext GW connection, No Ext IPv4 Block, disable default DFW
				Config: testAccNsxtPolicyProjectDefaultSecurityPolicy(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectGetSecurityProfileNSEnabled(testResourceName, false),
				),
			},
			{
				// Update: Set T0, Ext GW connection, No Ext IPv4 Block, disable default DFW
				Config: testAccNsxtPolicyProjectDefaultSecurityPolicy(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectGetSecurityProfileNSEnabled(testResourceName, true),
				),
			},
			{
				// Update: Set T0, Ext GW connection, No Ext IPv4 Block, disable default DFW
				Config: testAccNsxtPolicyProjectDefaultSecurityPolicy(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyProjectGetSecurityProfileNSEnabled(testResourceName, false),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyProject_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_project.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccOnlyLocalManager(t)
			testAccNSXVersion(t, "4.1.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyProjectCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyProjectMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtPolicyProjectGetSecurityProfileNSEnabled(resourceName string, expectedVal bool) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Project resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Project resource ID not set in resources")
		}

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		client := projects.NewVpcSecurityProfilesClient(connector)
		obj, err := client.Get(defaultOrgID, resourceID, "default")
		if err != nil {
			return err
		}
		if *obj.NorthSouthFirewall.Enabled != expectedVal {
			return fmt.Errorf("expected NorthSouthFirewall to be %v, isntead status is %v", *obj.NorthSouthFirewall.Enabled, expectedVal)
		}
		return nil
	}
}

func testAccNsxtPolicyProjectExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Project resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Project resource ID not set in resources")
		}

		exists, err := resourceNsxtPolicyProjectExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Policy Project %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyProjectCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_project" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtPolicyProjectExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy Project %s still exists", displayName)
		}
	}
	return nil
}

// TODO: Visibility should not be set for 410 tests (attribute exists, different enum values since 411)
var templateData = `
data "nsxt_policy_tier0_gateway" "test" {
  display_name = "{{.T0Name}}"
}

resource "nsxt_policy_ip_block" "test_ip_block" {
  display_name = "test_ip_block"
  cidr         = "10.20.0.0/16"
  {{if .SetIpBlockVisibility}}visibility   = "EXTERNAL"{{end}}
}

{{if .ExternalTGWConnectionPath}}resource "nsxt_policy_gateway_connection" "test_gw_conn" {
  display_name     = "test_gw_conn"
  tier0_path       = {{.T0GwPath}}
  aggregate_routes = ["192.168.240.0/24"]
}{{end}}

resource "nsxt_policy_project" "test" {
  display_name               = "{{.DisplayName}}"
  description                = "{{.Description}}"
  {{if .ShortId}}short_id                   = "{{.ShortId}}"{{end}}
  {{if .T0GwPath}}tier0_gateway_paths        = [{{.T0GwPath}}]{{end}}
  {{if .ExternalIPv4BlockPath}}external_ipv4_blocks       = [{{.ExternalIPv4BlockPath}}]{{end}}
  {{if .ExternalTGWConnectionPath}}tgw_external_connections   = [{{.ExternalTGWConnectionPath}}]{{end}}
  {{if .ActivateDefaultDfwRules}}activate_default_dfw_rules = {{.ActivateDefaultDfwRules}}{{end}}
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`

func getBasicAttrMap(createFlow, includeT0GW bool) map[string]string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestPolicyProjectCreateAttributes
	} else {
		// do not pass short_id on update
		attrMap = make(map[string]string)
		attrMap["DisplayName"] = accTestPolicyProjectUpdateAttributes["DisplayName"]
		attrMap["Description"] = accTestPolicyProjectUpdateAttributes["Description"]
	}
	if includeT0GW {
		attrMap["T0GwPath"] = "data.nsxt_policy_tier0_gateway.test.path"
	}
	attrMap["T0Name"] = getTier0RouterName()
	return attrMap
}

func testAccNsxtPolicyProjectTemplate410(createFlow, includeT0GW bool) string {
	attrMap := getBasicAttrMap(createFlow, includeT0GW)
	buffer := new(bytes.Buffer)
	tmpl, err := template.New("testAaccNsxtPolicyProject").Parse(templateData)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buffer, attrMap)
	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func testAccNsxtPolicyProjectTemplate411(createFlow, includeT0GW, includeExternalIPv4Block bool) string {
	attrMap := getBasicAttrMap(createFlow, includeT0GW)
	if includeExternalIPv4Block {
		attrMap["ExternalIPv4BlockPath"] = "nsxt_policy_ip_block.test_ip_block.path"
	}
	attrMap["SetIpBlockVisibility"] = "true"
	buffer := new(bytes.Buffer)
	tmpl, err := template.New("testAaccNsxtPolicyProject").Parse(templateData)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buffer, attrMap)
	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func testAccNsxtPolicyProjectTemplate420(createFlow, includeT0GW, includeExternalIPv4Block, activateDefaultDfwRules bool) string {
	attrMap := getBasicAttrMap(createFlow, includeT0GW)
	if includeExternalIPv4Block {
		attrMap["ExternalIPv4BlockPath"] = "nsxt_policy_ip_block.test_ip_block.path"
	}
	attrMap["SetIpBlockVisibility"] = "true"
	attrMap["ActivateDefaultDfwRules"] = strconv.FormatBool(activateDefaultDfwRules)
	buffer := new(bytes.Buffer)
	tmpl, err := template.New("testAaccNsxtPolicyProject").Parse(templateData)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buffer, attrMap)
	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func testAccNsxtPolicyProjectTemplate900(createFlow, includeT0GW, includeExternalIPv4Block, activateDefaultDfwRules bool) string {
	attrMap := getBasicAttrMap(createFlow, includeT0GW)
	if includeExternalIPv4Block {
		attrMap["ExternalIPv4BlockPath"] = "nsxt_policy_ip_block.test_ip_block.path"
	}
	attrMap["SetIpBlockVisibility"] = "true"
	attrMap["ActivateDefaultDfwRules"] = strconv.FormatBool(activateDefaultDfwRules)
	attrMap["ExternalTGWConnectionPath"] = "nsxt_policy_gateway_connection.test_gw_conn.path"
	buffer := new(bytes.Buffer)
	tmpl, err := template.New("testAaccNsxtPolicyProject").Parse(templateData)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(buffer, attrMap)
	if err != nil {
		panic(err)
	}
	return buffer.String()
}

func testAccNsxtPolicyProjectMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_project" "test" {
  display_name = "%s"

}`, accTestPolicyProjectUpdateAttributes["DisplayName"])
}

func testAccNsxtPolicyProjectDefaultSecurityPolicy(enabled bool) string {
	return fmt.Sprintf(`
resource "nsxt_policy_project" "test" {
  display_name = "%s"
  default_security_profile {
    north_south_firewall {
      enabled = %s
    }
  }

}`, accTestPolicyProjectCreateAttributes["DisplayName"], strconv.FormatBool(enabled))
}
