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

func TestAccResourceNsxtLbPool_basic(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"
	algorithm := "LEAST_CONNECTION"
	updatedAlgorithm := "WEIGHTED_ROUND_ROBIN"
	minActiveMembers := "3"
	updatedMinActiveMembers := "4"
	snatTranslationType := "TRANSPARENT"
	updatedSnatTranslationType := "SNAT_AUTO_MAP"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateTemplate(name, algorithm, minActiveMembers, snatTranslationType),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", algorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", minActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", snatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
			{
				Config: testAccNSXLbPoolUpdateTemplate(updatedName, updatedAlgorithm, updatedMinActiveMembers, updatedSnatTranslationType),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(updatedName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", updatedAlgorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", updatedMinActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", updatedSnatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbPool_withMonitors(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateWithMonitorsTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttrSet(testResourceName, "active_monitor_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "passive_monitor_id"),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", "TRANSPARENT"),
				),
			},
			{
				Config: testAccNSXLbPoolUpdateWithMonitorsTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "active_monitor_id", ""),
					resource.TestCheckResourceAttr(testResourceName, "passive_monitor_id", ""),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbPool_withIpSnat(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"
	algorithm := "LEAST_CONNECTION"
	updatedAlgorithm := "WEIGHTED_ROUND_ROBIN"
	minActiveMembers := "3"
	updatedMinActiveMembers := "4"
	snatTranslationType := "SNAT_IP_POOL"
	ipAddress := "1.1.1.1"
	updatedIPAddress := "1.1.1.2-1.1.1.20"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateWithSnatTemplate(name, algorithm, minActiveMembers, snatTranslationType, ipAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", algorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", minActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", snatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.ip", ipAddress),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
			{
				Config: testAccNSXLbPoolUpdateWithSnatTemplate(name, updatedAlgorithm, updatedMinActiveMembers, snatTranslationType, updatedIPAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", updatedAlgorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", updatedMinActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", snatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.ip", updatedIPAddress),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbPool_withMember(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"
	algorithm := "LEAST_CONNECTION"
	updatedAlgorithm := "WEIGHTED_ROUND_ROBIN"
	minActiveMembers := "3"
	updatedMinActiveMembers := "4"
	snatTranslationType := "TRANSPARENT"
	updatedSnatTranslationType := "SNAT_AUTO_MAP"
	memberIP := "1.1.1.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateWithMemberTemplate(name, algorithm, minActiveMembers, snatTranslationType, memberIP),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", algorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", minActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", snatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.display_name", name+"-member"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.ip_address", memberIP),
				),
			},
			{
				Config: testAccNSXLbPoolUpdateWithMemberTemplate(name, updatedAlgorithm, updatedMinActiveMembers, updatedSnatTranslationType, memberIP),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", updatedAlgorithm),
					resource.TestCheckResourceAttr(testResourceName, "min_active_members", updatedMinActiveMembers),
					resource.TestCheckResourceAttr(testResourceName, "snat_translation.0.type", updatedSnatTranslationType),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.display_name", name+"-member"),
					resource.TestCheckResourceAttr(testResourceName, "member.0.ip_address", memberIP),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbPool_withMemberGroup(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"
	algorithm := "LEAST_CONNECTION"
	size := "3"
	port := "50"
	updatedPort := "60"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateWithMemberGroupTemplate(name, algorithm, size, port),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", algorithm),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.limit_ip_list_size", "true"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.max_ip_list_size", size),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.ip_version_filter", "IPV4"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.port", port),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.grouping_object.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.grouping_object.0.target_type", "NSGroup"),
					resource.TestCheckResourceAttrSet(testResourceName, "member_group.0.grouping_object.0.target_id"),
				),
			},
			{
				Config: testAccNSXLbPoolUpdateWithMemberGroupTemplate(name, algorithm, updatedPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXLbPoolExists(name, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm", algorithm),
					resource.TestCheckResourceAttr(testResourceName, "member.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.limit_ip_list_size", "false"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.max_ip_list_size", "0"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.ip_version_filter", "IPV6"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.port", updatedPort),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.grouping_object.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "member_group.0.grouping_object.0.target_type", "NSGroup"),
					resource.TestCheckResourceAttrSet(testResourceName, "member_group.0.grouping_object.0.target_id"),
				),
			},
		},
	})
}

func TestAccResourceNsxtLbPool_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_lb_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccTestDeprecated(t)
			testAccPreCheck(t)
			testAccNSXVersion(t, "2.3.0")
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXLbPoolCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXLbPoolCreateTemplateTrivial(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXLbPoolExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX LB pool resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX LB pool resource ID not set in resources ")
		}

		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerPool(nsxClient.Context, resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving LB pool with ID %s. Error: %v", resourceID, err)
		}

		if responseCode.StatusCode != http.StatusOK {
			return fmt.Errorf("Error while checking if LB pool %s exists. HTTP return code was %d", resourceID, responseCode.StatusCode)
		}

		if displayName == monitor.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX LB pool %s wasn't found", displayName)
	}
}

func testAccNSXLbPoolCheckDestroy(state *terraform.State, displayName string) error {
	nsxClient := testAccProvider.Meta().(nsxtClients).NsxtClient
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_lb_icmp_monitor" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		monitor, responseCode, err := nsxClient.ServicesApi.ReadLoadBalancerPool(nsxClient.Context, resourceID)
		if err != nil {
			if responseCode.StatusCode != http.StatusOK {
				return nil
			}
			return fmt.Errorf("Error while retrieving LB pool with ID %s. Error: %v", resourceID, err)
		}

		if displayName == monitor.DisplayName {
			return fmt.Errorf("NSX LB pool %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXLbPoolCreateTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  algorithm          = "%s"
  description        = "Acceptance Test"
  min_active_members = "%s"

  snat_translation {
  	type = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, algorithm, minActiveMembers, snatTranslationType)
}

func testAccNSXLbPoolUpdateTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name          = "%s"
  algorithm             = "%s"
  description           = "Updated Acceptance Test"
  min_active_members    = "%s"

  snat_translation {
  	type = "%s"
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
`, name, algorithm, minActiveMembers, snatTranslationType)
}

func testAccNSXLbPoolMonitorsTemplate() string {
	return `
resource "nsxt_lb_icmp_monitor" "lb_icmp_monitor" {
  display_name = "lb_icmp_monitor"
  fall_count   = 3
  interval     = 5
}

resource "nsxt_lb_passive_monitor" "lb_passive_monitor" {
  display_name = "lb_passive_monitor"
  max_fails    = 3
  timeout      = 10
}
`
}

func testAccNSXLbPoolCreateWithMonitorsTemplate(name string) string {
	return testAccNSXLbPoolMonitorsTemplate() + fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  description        = "Acceptance Test"
  active_monitor_id  = "${nsxt_lb_icmp_monitor.lb_icmp_monitor.id}"
  passive_monitor_id = "${nsxt_lb_passive_monitor.lb_passive_monitor.id}"
}
`, name)
}

func testAccNSXLbPoolUpdateWithMonitorsTemplate(name string) string {
	return testAccNSXLbPoolMonitorsTemplate() + fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  description        = "Updated Acceptance Test"
}
`, name)
}

func testAccNSXLbPoolCreateWithSnatTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string, snatTranslationIP string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  algorithm          = "%s"
  description        = "Acceptance Test"
  min_active_members = "%s"

  snat_translation {
  	type          = "%s"
  	ip            = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}
`, name, algorithm, minActiveMembers, snatTranslationType, snatTranslationIP)
}

func testAccNSXLbPoolUpdateWithSnatTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string, snatTranslationIP string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  algorithm          = "%s"
  description        = "Updated Acceptance Test"
  min_active_members = "%s"

  snat_translation {
  	type          = "%s"
  	ip            = "%s"
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
`, name, algorithm, minActiveMembers, snatTranslationType, snatTranslationIP)
}

func testAccNSXLbPoolCreateTemplateTrivial() string {
	return `
resource "nsxt_lb_pool" "test" {
  description = "test description"
}
`
}

func testAccNSXLbPoolCreateWithMemberTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string, memberIP string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name       = "%s"
  algorithm          = "%s"
  description        = "Acceptance Test"
  min_active_members = "%s"

  snat_translation {
  	type = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  member {
    admin_state                = "ENABLED"
    backup_member              = "false"
    display_name               = "%s-member"
    ip_address                 = "%s"
    max_concurrent_connections = "7"
    port                       = "77"
    weight                     = "1"
  }
}
`, name, algorithm, minActiveMembers, snatTranslationType, name, memberIP)
}

func testAccNSXLbPoolUpdateWithMemberTemplate(name string, algorithm string, minActiveMembers string, snatTranslationType string, memberIP string) string {
	return fmt.Sprintf(`
resource "nsxt_lb_pool" "test" {
  display_name          = "%s"
  algorithm             = "%s"
  description           = "Updated Acceptance Test"
  min_active_members    = "%s"

  snat_translation {
  	type = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
  tag {
    scope = "scope2"
    tag   = "tag2"
  }

  member {
    admin_state                = "ENABLED"
    backup_member              = "false"
    display_name               = "%s-member"
    ip_address                 = "%s"
    max_concurrent_connections = "7"
    port                       = "77"
    weight                     = "1"
  }
  member {
    admin_state                = "DISABLED"
    backup_member              = "true"
    display_name               = "2nd-member"
    ip_address                 = "7.7.7.7"
    max_concurrent_connections = "8"
    port                       = "88"
    weight                     = "8"
  }
}
`, name, algorithm, minActiveMembers, snatTranslationType, name, memberIP)
}

func testAccNSXLbPoolCreateWithMemberGroupTemplate(name string, algorithm string, size string, port string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "grp1" {
  display_name = "grp1"
}

resource "nsxt_lb_pool" "test" {
  display_name          = "%s"
  algorithm             = "%s"
  description           = "Acceptance Test"

  member_group {
    ip_version_filter  = "IPV4"
    limit_ip_list_size = true
    max_ip_list_size   = %s
    port               = "%s"

    grouping_object {
      target_type = "NSGroup"
      target_id   = "${nsxt_ns_group.grp1.id}"
    }
  }
}
`, name, algorithm, size, port)
}

func testAccNSXLbPoolUpdateWithMemberGroupTemplate(name string, algorithm string, port string) string {
	return fmt.Sprintf(`
resource "nsxt_ns_group" "grp1" {
  display_name = "grp1"
}

resource "nsxt_lb_pool" "test" {
  display_name          = "%s"
  algorithm             = "%s"
  description           = "Updated Acceptance Test"

  member_group {
    ip_version_filter  = "IPV6"
    limit_ip_list_size = false
    port               = "%s"

    grouping_object {
      target_type = "NSGroup"
      target_id   = "${nsxt_ns_group.grp1.id}"
    }
  }
}
`, name, algorithm, port)
}
