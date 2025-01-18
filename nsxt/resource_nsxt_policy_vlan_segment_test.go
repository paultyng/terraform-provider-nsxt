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

func TestAccResourceNsxtPolicyVlanSegment_basicImport(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentImportTemplate(name, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyVlanSegment_basicImport_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentImportTemplate(name, true),
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

func TestAccResourceNsxtPolicyVlanSegment_basicUpdate(t *testing.T) {
	testAccResourceNsxtPolicyVlanSegmentBasicUpdate(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyVlanSegment_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyVlanSegmentBasicUpdate(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyVlanSegmentBasicUpdate(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyVlanSegmentCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentBasicTemplate(name, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVlanSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyVlanSegmentBasicUpdateTemplate(updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVlanSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test2"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest2.org"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyVlanSegment_updateAdvConfig(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyVlanSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentBasicAdvConfigTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVlanSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", "SOURCE"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "OFF"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "true"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.urpf_mode", "NONE"),
				),
			},
			{
				Config: testAccNsxtPolicyVlanSegmentBasicAdvConfigUpdateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyVlanSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", "MTEP"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "ON"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "false"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.urpf_mode", "STRICT"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyVlanSegment_noTransportZone(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"
	cidr := "4003::1/64"
	updatedCidr := "4004::1/64"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyGlobalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentNoTransportZoneTemplate(name, cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", cidr),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyVlanSegmentNoTransportZoneTemplate(name, updatedCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", updatedCidr),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyVlanSegment_withDhcp(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_vlan_segment.test"
	leaseTimes := []string{"3600", "36000"}
	preferredTimes := []string{"3200", "32000"}
	dnsServersV4 := []string{"2.2.2.2", "3.3.3.3"}
	dnsServersV6 := []string{"2000::2", "3000::3"}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyVlanSegmentWithDhcpTemplate(name, dnsServersV4[0], dnsServersV6[0], leaseTimes[0], preferredTimes[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v6_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.server_address", "12.12.2.2/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.lease_time", leaseTimes[0]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.0", dnsServersV4[0]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_option_121.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.cidr", "4012::1/64"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.server_address", "4012::2/64"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.lease_time", leaseTimes[0]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.preferred_time", preferredTimes[0]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.dns_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.dns_servers.0", dnsServersV6[0]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.sntp_servers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.excluded_range.#", "2"),
				),
			},
			{
				Config: testAccNsxtPolicyVlanSegmentWithDhcpTemplate(name, dnsServersV4[1], dnsServersV6[1], leaseTimes[1], preferredTimes[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v6_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.server_address", "12.12.2.2/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.lease_time", leaseTimes[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.0", dnsServersV4[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_option_121.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.cidr", "4012::1/64"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.server_address", "4012::2/64"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.lease_time", leaseTimes[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.preferred_time", preferredTimes[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.dns_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.dns_servers.0", dnsServersV6[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.sntp_servers.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.0.excluded_range.#", "2"),
				),
			},
		},
	})
}

// TODO: add tests for l2_extension; requires L2 VPN Session

func testAccNsxtPolicyVlanSegmentExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy VLAN Segment resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy VLAN Segment resource ID not set in resources")
		}
		context := testAccGetSessionContext()
		exists, err := resourceNsxtPolicySegmentExists(context, "", false)(context, resourceID, connector)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Error while retrieving policy VLAN Segment ID %s", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyVlanSegmentCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_vlan_segment" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		context := testAccGetSessionContext()
		exists, err := resourceNsxtPolicySegmentExists(context, "", false)(context, resourceID, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy VLAN Segment %s (%s) still exists", displayName, resourceID)
		}
	}
	return nil
}

func testAccNsxtPolicyVlanSegmentDeps() string {
	return testAccNSXPolicyTransportZoneReadTemplate(getVlanTransportZoneName(), true, true)
}

func testAccNsxtPolicyVlanSegmentImportTemplate(name string, withContext bool) string {
	context := ""
	tzSetting := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	} else {
		tzSetting = "transport_zone_path = data.nsxt_policy_transport_zone.test.path"
	}
	s := fmt.Sprintf(`
resource "nsxt_policy_vlan_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  vlan_ids            = ["101"]
  %s
}
`, context, name, tzSetting)
	if withContext {
		return s
	}
	return testAccNsxtPolicyVlanSegmentDeps() + s
}

func testAccNsxtPolicyVlanSegmentBasicTemplate(name string, withContext bool) string {
	context := ""
	deps := testAccNsxtPolicyVlanSegmentDeps()
	tzSpec := "transport_zone_path = data.nsxt_policy_transport_zone.test.path"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		deps = ""
		tzSpec = ""
	}
	return deps + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  %s
  domain_name         = "tftest.org"
  vlan_ids            = ["101"]

  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, context, name, tzSpec)
}

func testAccNsxtPolicyVlanSegmentBasicUpdateTemplate(name string, withContext bool) string {
	context := ""
	deps := testAccNsxtPolicyVlanSegmentDeps()
	tzSpec := "transport_zone_path = data.nsxt_policy_transport_zone.test.path"
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
		deps = ""
		tzSpec = ""
	}
	return deps + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test2"
  %s
  domain_name         = "tftest2.org"
  vlan_ids            = ["101", "104-110"]

  tag {
    scope = "color"
    tag   = "green"
  }
  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, context, name, tzSpec)
}

func testAccNsxtPolicyVlanSegmentBasicAdvConfigTemplate(name string) string {
	return testAccNsxtPolicyVlanSegmentDeps() + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  replication_mode    = "SOURCE"
  vlan_ids            = ["101", "102"]

  tag {
    scope = "color"
    tag   = "orange"
  }

  advanced_config {
    connectivity = "OFF"
    local_egress = true
    urpf_mode    = "NONE"
  }
}
`, name)
}

func testAccNsxtPolicyVlanSegmentBasicAdvConfigUpdateTemplate(name string) string {
	return testAccNsxtPolicyVlanSegmentDeps() + fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  vlan_ids            = ["101", "102"]

  tag {
    scope = "color"
    tag   = "orange"
  }

  advanced_config {
    connectivity = "ON"
    local_egress = false
  }
}
`, name)
}

func testAccNsxtPolicyVlanSegmentWithDhcpTemplate(name string, dnsServerV4 string, dnsServerV6 string, lease string, preferred string) string {
	return testAccNsxtPolicyVlanSegmentDeps() + fmt.Sprintf(`

data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}

resource "nsxt_policy_dhcp_server" "test" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  display_name      = "segment-test"
}

resource "nsxt_policy_vlan_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  vlan_ids            = ["101", "102"]

  subnet {
    cidr = "12.12.2.1/24"
    dhcp_v4_config {
        server_address = "12.12.2.2/24"
        lease_time     = %s
        dns_servers    = ["%s"]
        dhcp_option_121 {
          network  = "2.1.1.0/24"
          next_hop = "2.3.1.3"
        }
        dhcp_option_121 {
          network  = "3.1.1.0/24"
          next_hop = "3.3.1.3"
        }
        dhcp_generic_option {
          code   = "119"
          values = ["abc", "def"]
        }
    }
  }

  subnet {
    cidr = "4012::1/64"
    dhcp_v6_config {
        server_address = "4012::2/64"
        lease_time     = %s
        preferred_time = %s
        dns_servers    = ["%s"]
        sntp_servers   = ["3001::1", "3001::2"]
        excluded_range {
            start = "4012::400"
            end   = "4012::500"
        }
        excluded_range {
            start = "4012::a00"
            end   = "4012::b00"
        }
    }
  }

  dhcp_config_path = nsxt_policy_dhcp_server.test.path

}
`, getEdgeClusterName(), name, lease, dnsServerV4, lease, preferred, dnsServerV6)
}

func testAccNsxtPolicyVlanSegmentNoTransportZoneTemplate(name string, cidr string) string {
	return fmt.Sprintf(`

resource "nsxt_policy_vlan_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  vlan_ids            = ["101"]

  subnet {
     cidr = "%s"
  }
}
`, name, cidr)
}
