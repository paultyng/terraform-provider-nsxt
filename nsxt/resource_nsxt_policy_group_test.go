package nsxt

import (
	"fmt"
	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
)

func TestAccResourceNsxtPolicyGroup_basicImport(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_group"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupIPAddressImportTemplate(name, resourceName, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_basicImport_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_group"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, tf_api.Multitenancy)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupIPAddressImportTemplate(name, resourceName, true),
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

func TestAccResourceNsxtPolicyGroup_empty(t *testing.T) {
	testAccResourceNsxtPolicyGroupEmpty(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyGroup_empty_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGroupEmpty(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyGroupEmpty(t *testing.T, withContext bool, preCheck func()) {
	// A simple tests that verify successful creation of a group with no criteria
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_group"

	var clientType tf_api.ClientType = tf_api.Local
	if withContext {
		clientType = tf_api.Multitenancy
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, clientType)
		},
		Steps: []resource.TestStep{
			{
				Config:      testAccNsxtPolicyEmptyCreateTemplate(name, resourceName, withContext),
				ExpectError: regexp.MustCompile("found empty criteria block in configuration"),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_addressCriteria(t *testing.T) {
	testAccResourceNsxtPolicyGroupAddressCriteria(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyGroup_addressCriteria_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyGroupAddressCriteria(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyGroupAddressCriteria(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	resourceName := "nsxt_policy_group"
	testResourceName := fmt.Sprintf("%s.test", resourceName)

	var clientType tf_api.ClientType = tf_api.Local
	if withContext {
		clientType = tf_api.Multitenancy
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, clientType)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupAddressCreateTemplate(name, resourceName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, clientType),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.macaddress_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.macaddress_expression.0.mac_addresses.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupAddressUpdateTemplate(name, resourceName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, clientType),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
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

func TestAccResourceNsxtPolicyGroup_groupTypeIPAddressCriteria(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"
	testResourceName2 := "nsxt_policy_group.test-2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNSXVersion(t, "3.2.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupIPAddressCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					testAccNsxtPolicyGroupExists(testResourceName2, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.0.ip_addresses.#", "2"),
					resource.TestCheckResourceAttr(testResourceName2, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName2, "criteria.0.condition.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupIPAddressUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.ipaddress_expression.0.ip_addresses.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtGlobalPolicyGroup_singleIPAddressCriteria(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, getTestSiteName(), tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtGlobalPolicyGroupIPAddressCreateTemplate(name, getTestSiteName()),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, getTestSiteName(), tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", getTestSiteName()),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
				),
			},
			{
				Config: testAccNsxtGlobalPolicyGroupIPAddressUpdateTemplate(updatedName, getTestSiteName()),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, getTestSiteName(), tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", getTestSiteName()),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtGlobalPolicyGroup_externalIDCriteria(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupExternalIDCreateTemplate(name, defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupExternalIDUpdateTemplate(updatedName, defaultDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtGlobalPolicyGroup_withDomain(t *testing.T) {
	name := "test-nsx-global-policy-group-domain"
	testResourceName := "nsxt_policy_group.test"
	domainName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccEnvDefined(t, "NSXT_TEST_SITE_NAME")
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyDomainCheckDestroy(state, name)

		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtGlobalPolicyGroupCreateTemplateWithDomain(name, domainName, getTestSiteName()),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, domainName, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", domainName),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_multipleIPAddressCriteria(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupIPAddressMultipleCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupIPAddressMultipleUpdateTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_pathCriteria(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccNSXGlobalManagerSitePrecheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupPathsCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.path_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.path_expression.0.member_paths.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupPathsUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.path_expression.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.path_expression.0.member_paths.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupPathsPrerequisites(),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_nestedCriteria(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupNestedConditionCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupNestedConditionUpdateTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_multipleCriteria(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupMultipleConditionCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.condition.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupMultipleConditionUpdateTemplate(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.condition.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_multipleNestedCriteria(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, name, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupMultipleNestedCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.condition.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.2.ipaddress_expression.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteIPAddrs(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.condition.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteNestedConds(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.1.condition.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteLastCriteria(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.0.condition.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyGroup_identityGroup(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_group.test"
	updatedName := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyGroupCheckDestroy(state, updatedName, defaultDomain, tf_api.Local)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyGroupIdentityGroup(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.distinguished_name", "test-dn"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.domain_base_distinguished_name", "test-dbdn"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupIdentityGroupUpdateChangeAD(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.distinguished_name", "test-dn-update"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.domain_base_distinguished_name", "test-dbdn-update"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.sid", "test-sid"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupIdentityGroupUpdateMultipleAD(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.distinguished_name", "test-dn-1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.domain_base_distinguished_name", "test-dbdn-1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.0.sid", "test-sid-1"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.1.distinguished_name", "test-dn-2"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.1.domain_base_distinguished_name", "test-dbdn-2"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.0.identity_group.1.sid", "test-sid-2"),
				),
			},
			{
				Config: testAccNsxtPolicyGroupIdentityGroupUpdateDeleteAD(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyGroupExists(testResourceName, defaultDomain, tf_api.Local),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "domain", defaultDomain),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "conjunction.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "criteria.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "extended_criteria.#", "0"),
				),
			},
		},
	})
}

func testAccNsxtPolicyGroupExists(resourceName string, domainName string, clientType tf_api.ClientType) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy Group resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy Group resource ID not set in resources")
		}

		nsxClient := domains.NewGroupsClient(testAccGetContextByType(clientType), connector)
		if nsxClient == nil {
			return policyResourceNotSupportedError()
		}
		_, err := nsxClient.Get(domainName, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy Group ID %s domain %s. Error: %v", resourceID, domainName, err)
		}

		return nil
	}
}

func testAccNsxtPolicyGroupCheckDestroy(state *terraform.State, displayName string, domainName string, clientType tf_api.ClientType) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_group" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		isPolicyGlobalManager := isPolicyGlobalManager(testAccProvider.Meta())
		domainID := domainName
		if isPolicyGlobalManager && domainName != "default" {
			// get the non default domain id
			var errDomain error
			domainID, errDomain = testGetObjIDByName(domainName, "Domain")
			if errDomain != nil {
				return fmt.Errorf("Error while retrieving policy domain %s. Error: %v", domainName, errDomain)
			}
		}
		exists, err := resourceNsxtPolicyGroupExistsInDomain(testAccGetContextByType(clientType), resourceID, domainID, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy Group %s still exists in domain %s", displayName, domainName)
		}
	}
	return nil
}

func testAccNsxtPolicyGroupIPAddressImportTemplate(name string, resourceName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"
}
`, resourceName, context, name)
}

func testAccNsxtPolicyEmptyCreateTemplate(name, resourceName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "%s" "test_empty" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
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
`, resourceName, context, name)
}

func testAccNsxtPolicyGroupAddressCreateTemplate(name, resourceName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "%s" "test" {
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
    macaddress_expression {
        mac_addresses = ["a2:54:00:68:b0:83", "fa:10:3e:01:49:5e"]
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
`, resourceName, context, name)
}

func testAccNsxtPolicyGroupAddressUpdateTemplate(name, resourceName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "%s" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1"]
    }
  }
}
`, resourceName, context, name)
}

func testAccNsxtGlobalPolicyGroupIPAddressCreateTemplate(name string, siteName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_site" "test" {
  display_name = "%s"
}
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  domain       = data.nsxt_policy_site.test.id

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1", "222.2.2.2"]
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
`, siteName, name)
}

func testAccNsxtGlobalPolicyGroupIPAddressUpdateTemplate(name string, siteName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_site" "test" {
  display_name = "%s"
}
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  domain       = data.nsxt_policy_site.test.id

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.2.1.1", "232.2.2.2"]
	}
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.3", "222.2.2.4"]
	}
  }
}
`, siteName, name)
}

func testAccNsxtGlobalPolicyGroupCreateTemplateWithDomain(name string, domainName string, siteName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_domain" "%s" {
  display_name = "%s"
  sites        = ["%s"]
  nsx_id       = "%s"
}

resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  domain       = nsxt_policy_domain.%s.id

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1", "222.2.2.2"]
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
`, domainName, domainName, siteName, domainName, name, domainName)
}

func testAccNsxtPolicyGroupIPAddressMultipleCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
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
	  ip_addresses = ["111.1.1.3", "222.2.2.3"]
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
`, name)
}

func testAccNsxtPolicyGroupExternalIDCreateTemplate(name string, siteName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    external_id_expression {
      member_type = "VirtualMachine"
      external_ids = ["520ba7b0-d9f8-87b1-6f44-15bbeb7935c7", "52748a9e-d61d-e29b-d54b-07f169ff0ee8-4000"]
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
`, name)
}

func testAccNsxtPolicyGroupExternalIDUpdateTemplate(name string, siteName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    external_id_expression {
      member_type = "VirtualMachine"
      external_ids = ["520ba7b0-d9f8-87b1-6f44-15bbeb7935c7", "52748a9e-d61d-e29b-d54b-07f169ff0ee8-4000"]
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    external_id_expression {
      member_type = "VirtualNetworkInterface"
      external_ids = ["520ba7b0-d9f8-87b1-6f44-15bbeb7935c7", "52748a9e-d61d-e29b-d54b-07f169ff0ee8-4000"]
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupPathsPrerequisites() string {
	var preRequisites string
	if testAccIsGlobalManager() {
		preRequisites = testNsxtGlobalPolicyGroupPathsTransportZone()
	} else {
		preRequisites = testNsxtPolicyGroupPathsTransportZone()
	}
	return preRequisites + `
resource "nsxt_policy_segment" "test-1" {
  display_name        = "group-test-1"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
}

resource "nsxt_policy_segment" "test-2" {
  display_name        = "group-test-1"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
}`

}

func testNsxtGlobalPolicyGroupPathsTransportZone() string {
	return fmt.Sprintf(`
data "nsxt_policy_site" "test" {
  display_name = "%s"
}
data "nsxt_policy_transport_zone" "test"{
  site_path      = data.nsxt_policy_site.test.path
  transport_type = "OVERLAY_STANDARD"
  is_default     = true
}`, getTestSiteName())
}

func testNsxtPolicyGroupPathsTransportZone() string {
	return fmt.Sprintf(`
data "nsxt_policy_transport_zone" "test"{
  display_name = "%s"
}`, getOverlayTransportZoneName())
}

func testAccNsxtPolicyGroupPathsCreateTemplate(name string) string {
	return testAccNsxtPolicyGroupPathsPrerequisites() + fmt.Sprintf(`

resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    path_expression {
      member_paths = [nsxt_policy_segment.test-1.path, nsxt_policy_segment.test-2.path]
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupPathsUpdateTemplate(name string) string {
	return testAccNsxtPolicyGroupPathsPrerequisites() + fmt.Sprintf(`

resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    path_expression {
      member_paths = [nsxt_policy_segment.test-1.path]
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupIPAddressMultipleUpdateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1-111.1.1.10", "222.2.2.0/24"]
	}
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.2.4", "222.2.3.4"]
	}
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupNestedConditionCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
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
`, name)
}

func testAccNsxtPolicyGroupNestedConditionUpdateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "green"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupMultipleConditionCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
  }

  conjunction {
    operator = "AND"
  }

  criteria {
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
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
`, name)
}

func testAccNsxtPolicyGroupMultipleConditionUpdateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
  }

  conjunction {
    operator = "OR"
  }

  criteria {
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "public"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

}
`, name)
}

func testAccNsxtPolicyGroupMultipleNestedCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "green"
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
    }
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "public"
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.4", "222.2.2.4"]
	}
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteIPAddrs(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Name"
      member_type = "VirtualMachine"
      operator    = "STARTSWITH"
      value       = "publicVM"
    }
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "green"
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
    }
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "public"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteNestedConds(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "green"
    }
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    condition {
      key         = "OSName"
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      value       = "Ubuntu"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupMultipleNestedUpdateTemplateDeleteLastCriteria(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  criteria {
    condition {
      key         = "Tag"
      member_type = "VirtualMachine"
      operator    = "EQUALS"
      value       = "green"
    }
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name)
}

func testAccNsxtPolicyGroupIdentityGroup(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  extended_criteria {
    identity_group {
      distinguished_name             = "test-dn"
      domain_base_distinguished_name = "test-dbdn"
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupIdentityGroupUpdateChangeAD(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  extended_criteria {
    identity_group {
          distinguished_name             = "test-dn-update"
          domain_base_distinguished_name = "test-dbdn-update"
          sid                            = "test-sid"
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupIdentityGroupUpdateMultipleAD(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  extended_criteria {
    identity_group {
          distinguished_name             = "test-dn-1"
          domain_base_distinguished_name = "test-dbdn-1"
          sid                            = "test-sid-1"
    }

    identity_group {
          distinguished_name             = "test-dn-2"
          domain_base_distinguished_name = "test-dbdn-2"
          sid                            = "test-sid-2"
    }
  }
}
`, name)
}

func testAccNsxtPolicyGroupIdentityGroupUpdateDeleteAD(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
}
`, name)
}

func testAccNsxtPolicyGroupIPAddressCreateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  group_type   = "IPAddress"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.1", "222.2.2.2"]
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

resource "nsxt_policy_group" "test-2" {
  display_name = "%s"

  criteria {
    condition {
      key         = "GroupType"
      member_type = "Group"
      operator    = "EQUALS"
      value       = "IPAddress"
    }
    condition {
      key         = "Tag"
      member_type = "Group"
      operator    = "EQUALS"
      value       = "orange"
    }
  }
}`, name, getAccTestResourceName())
}

func testAccNsxtPolicyGroupIPAddressUpdateTemplate(name string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_group" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  group_type   = "IPAddress"

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.2.1.1", "232.2.2.2"]
	}
  }

  conjunction {
	operator = "OR"
  }

  criteria {
    ipaddress_expression {
	  ip_addresses = ["111.1.1.3", "222.2.2.4"]
	}
  }
}
`, name)
}
