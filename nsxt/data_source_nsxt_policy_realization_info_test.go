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

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1DataSource(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1DataSource(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1DataSource_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1DataSource(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyRealizationInfoTier1DataSource(t *testing.T, withContext bool, preCheck func()) {
	resourceDataType := "nsxt_policy_tier1_gateway"
	resourceName := getAccTestDataSourceName()
	entityType := ""
	testResourceName := "data.nsxt_policy_realization_info.realization_info"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyTier1GatewayDeleteByName(resourceName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyTier1GatewayCreate(resourceName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyRealizationInfoReadDataSourceTemplate(resourceDataType, resourceName, entityType, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttrSet(testResourceName, "entity_type"),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1DataSourceEntity(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1DataSourceEntity(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1DataSourceEntity_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1DataSourceEntity(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyRealizationInfoTier1DataSourceEntity(t *testing.T, withContext bool, preCheck func()) {
	resourceDataType := "nsxt_policy_tier1_gateway"
	resourceName := getAccTestDataSourceName()
	entityType := "RealizedLogicalRouter"
	testResourceName := "data.nsxt_policy_realization_info.realization_info"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyTier1GatewayDeleteByName(resourceName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyTier1GatewayCreate(resourceName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyRealizationInfoReadDataSourceTemplate(resourceDataType, resourceName, entityType, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttr(testResourceName, "entity_type", entityType),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1Resource(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1Resource(t, false, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyLocalManager(t)
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_tier1Resource_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoTier1Resource(t, true, func() {
		testAccPreCheck(t)
		testAccNSXVersion(t, "3.0.0")
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyRealizationInfoTier1Resource(t *testing.T, withContext bool, preCheck func()) {
	resourceType := "nsxt_policy_tier1_gateway"
	resourceName := getAccTestDataSourceName()
	entityType := "RealizedLogicalRouter"
	testResourceName := "data.nsxt_policy_realization_info.realization_info"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRealizationInfoReadResourceTemplate(resourceType, resourceName, entityType, withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttr(testResourceName, "entity_type", entityType),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_errorState(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoErrorState(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyLocalManager(t)
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_errorState_multitenancy(t *testing.T) {
	testAccDataSourceNsxtPolicyRealizationInfoErrorState(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccDataSourceNsxtPolicyRealizationInfoErrorState(t *testing.T, withContext bool, preCheck func()) {
	testResourceName := "data.nsxt_policy_realization_info.realization_info"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyRealizationInfoReadDataSourceErrorTemplate(withContext),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "ERROR"),
					resource.TestCheckResourceAttrSet(testResourceName, "entity_type"),
					resource.TestCheckResourceAttr(testResourceName, "realized_id", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func TestAccDataSourceNsxtPolicyRealizationInfo_gmServiceDataSource(t *testing.T) {
	resourceDataType := "nsxt_policy_service"
	resourceName := "ICMPv6-ALL"
	entityType := ""
	testResourceName := "data.nsxt_policy_realization_info.realization_info"
	site := getTestSiteName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGMRealizationInfoReadDataSourceTemplate(resourceDataType, resourceName, entityType, site),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "state", "REALIZED"),
					resource.TestCheckResourceAttrSet(testResourceName, "entity_type"),
					resource.TestCheckResourceAttrSet(testResourceName, "realized_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "site_path"),
				),
			},
		},
	})
}

func testAccNsxtPolicyGMRealizationInfoReadDataSourceTemplate(resourceDataType string, resourceName string, entityType string, site string) string {
	return fmt.Sprintf(`
data "%s" "policy_resource" {
  display_name = "%s"
}

data "nsxt_policy_site" "test" {
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
  path = data.%s.policy_resource.path
  entity_type = "%s"
  site_path = data.nsxt_policy_site.test.path
}`, resourceDataType, resourceName, site, resourceDataType, entityType)
}

func testAccNsxtPolicyRealizationInfoReadDataSourceErrorTemplate(withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_ip_pool" "test" {
%s
  display_name = "tfippool1"
}

resource "nsxt_policy_ip_pool_static_subnet" "test" {
%s
  display_name = "tfssnet1"
  pool_path    = nsxt_policy_ip_pool.test.path
  cidr         = "12.12.12.0/24"
  allocation_range {
    start = "12.12.12.10"
    end   = "12.12.12.20"
  }
}

resource "nsxt_policy_ip_address_allocation" "test" {
%s
  display_name  = "tfipallocationerror"
  pool_path     = nsxt_policy_ip_pool.test.path
  allocation_ip = "12.12.12.21"
  depends_on    = [nsxt_policy_ip_pool_static_subnet.test]
}

data "nsxt_policy_realization_info" "realization_info" {
%s
  path = nsxt_policy_ip_address_allocation.test.path
}
`, context, context, context, context)
}

func testAccNsxtPolicyRealizationInfoReadDataSourceTemplate(resourceDataType string, resourceName string, entityType string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
data "%s" "policy_resource" {
%s
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
%s
  path = data.%s.policy_resource.path
  entity_type = "%s"
}`, resourceDataType, context, resourceName, context, resourceDataType, entityType)
}

func testAccNsxtPolicyRealizationInfoReadResourceTemplate(resourceType string, resourceName string, entityType string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "%s" "policy_resource" {
%s
  display_name = "%s"
}

data "nsxt_policy_realization_info" "realization_info" {
%s
  path = %s.policy_resource.path
  entity_type = "%s"
}`, resourceType, context, resourceName, context, resourceType, entityType)
}
