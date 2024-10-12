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

var testAccPolicyFixedSegmentResourceName = "nsxt_policy_fixed_segment.test"

func TestAccResourceNsxtPolicyFixedSegment_basicImport(t *testing.T) {
	name := getAccTestResourceName()
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentImportTemplate(tzName, name, false),
			},
			{
				ResourceName:      testAccPolicyFixedSegmentResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccNSXPolicyFixedSegmentImporterGetID,
			},
		},
	})
}

func TestAccResourceNsxtPolicyFixedSegment_basicImport_multitenancy(t *testing.T) {
	name := getAccTestResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentImportTemplate("", name, true),
			},
			{
				ResourceName:      testAccPolicyFixedSegmentResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccResourceNsxtPolicyImportIDRetriever(testAccPolicyFixedSegmentResourceName),
			},
		},
	})
}

func TestAccResourceNsxtPolicyFixedSegment_basicUpdate(t *testing.T) {
	testAccResourceNsxtPolicyFixedSegmentBasicUpdate(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyFixedSegment_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyFixedSegmentBasicUpdate(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func testAccResourceNsxtPolicyFixedSegmentBasicUpdate(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := testAccPolicyFixedSegmentResourceName
	tzName := ""
	if !withContext {
		tzName = getOverlayTransportZoneName()
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentBasicTemplate(tzName, name, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtPolicyFixedSegmentBasicUpdateTemplate(tzName, updatedName, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updatedName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "22.22.22.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest2.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyFixedSegment_connectivityPath(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := testAccPolicyFixedSegmentResourceName
	tzName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentBasicTemplate(tzName, name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "connectivity_path"),
				),
			},
			{
				Config: testAccNsxtPolicyFixedSegmentUpdateConnectivityTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "22.22.22.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest2.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttrSet(testResourceName, "connectivity_path"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentDeps(tzName, false),
			},
		},
	})
}

func TestAccResourceNsxtPolicyFixedSegment_updateAdvConfig(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := testAccPolicyFixedSegmentResourceName
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentBasicAdvConfigTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", "SOURCE"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "OFF"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "true"),
				),
			},
			{
				Config: testAccNsxtPolicyFixedSegmentBasicAdvConfigUpdateTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", "MTEP"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "ON"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "false"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyFixedSegment_withDhcp(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := testAccPolicyFixedSegmentResourceName
	leaseTimes := []string{"3600", "36000"}
	preferredTimes := []string{"3200", "32000"}
	dnsServersV4 := []string{"2.2.2.2", "3.3.3.3"}
	dnsServersV6 := []string{"2000::2", "3000::3"}
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyFixedSegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyFixedSegmentWithDhcpTemplate(tzName, name, dnsServersV4[0], dnsServersV6[0], leaseTimes[0], preferredTimes[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
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
				Config: testAccNsxtPolicyFixedSegmentWithDhcpTemplate(tzName, name, dnsServersV4[1], dnsServersV6[1], leaseTimes[1], preferredTimes[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyFixedSegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v6_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.lease_time", leaseTimes[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dns_servers.0", dnsServersV4[1]),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_option_121.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.dhcp_v4_config.0.dhcp_generic_option.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.cidr", "4012::1/64"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.1.dhcp_v6_config.#", "1"),
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

func testAccNsxtPolicyFixedSegmentExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Policy IP Pool resource %s not found in resources", resourceName)
		}
		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["connectivity_path"]
		context := testAccGetSessionContext()
		exists, err := resourceNsxtPolicySegmentExists(context, gwPath, true)(context, resourceID, connector)
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Policy resource %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyFixedSegmentCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_fixed_segment" {
			continue
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy resource ID not set in resources")
		}
		gwPath := rs.Primary.Attributes["connectivity_path"]
		context := testAccGetSessionContext()
		exists, err := resourceNsxtPolicySegmentExists(context, gwPath, true)(context, resourceID, connector)
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("Policy fixed segment %s still exists", displayName)
		}
	}
	return nil
}

func testAccNSXPolicyFixedSegmentImporterGetID(s *terraform.State) (string, error) {
	rs, ok := s.RootModule().Resources[testAccPolicyFixedSegmentResourceName]
	if !ok {
		return "", fmt.Errorf("NSX Policy Fixed Segment resource %s not found in resources", testAccPolicyFixedSegmentResourceName)
	}
	resourceID := rs.Primary.ID
	if resourceID == "" {
		return "", fmt.Errorf("NSX Policy Fixed Segment resource ID not set in resources")
	}
	gwPath := rs.Primary.Attributes["connectivity_path"]
	if gwPath == "" {
		return "", fmt.Errorf("NSX Policy Fixed Segment connectivity_path not specified")
	}
	_, gwID := parseGatewayPolicyPath(gwPath)
	return fmt.Sprintf("%s/%s", gwID, resourceID), nil
}

func testAccNsxtPolicyFixedSegmentImportTemplate(tzName string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicySegmentDeps(tzName, withContext) + fmt.Sprintf(`
resource "nsxt_policy_fixed_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }
}
`, context, name)
}

func testAccNsxtPolicyFixedSegmentBasicTemplate(tzName string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicySegmentDeps(tzName, withContext) + fmt.Sprintf(`

resource "nsxt_policy_fixed_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  overlay_id          = 1011
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, context, name)
}

func testAccNsxtPolicyFixedSegmentBasicUpdateTemplate(tzName string, name string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return testAccNsxtPolicySegmentDeps(tzName, withContext) + fmt.Sprintf(`
resource "nsxt_policy_fixed_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test2"
  domain_name         = "tftest2.org"
  overlay_id          = 1011
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "22.22.22.1/24"
  }

  tag {
    scope = "color"
    tag   = "green"
  }
  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, context, name)
}

func testAccNsxtPolicyFixedSegmentUpdateConnectivityTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_fixed_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test2"
  domain_name         = "tftest2.org"
  overlay_id          = 1011
  connectivity_path   = nsxt_policy_tier1_gateway.anotherTier1ForSegments.path

  subnet {
     cidr = "22.22.22.1/24"
  }

  tag {
    scope = "color"
    tag   = "green"
  }
  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, name)
}

func testAccNsxtPolicyFixedSegmentBasicAdvConfigTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_fixed_segment" "test" {
  display_name     = "%s"
  description      = "Acceptance Test"
  replication_mode = "SOURCE"
  domain_name      = "tftest.org"
  overlay_id       = 1011
  vlan_ids         = ["101", "102"]

  connectivity_path   = nsxt_policy_tier1_gateway.anotherTier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "orange"
  }

  advanced_config {
    connectivity = "OFF"
    local_egress = true
  }
}
`, name)
}

func testAccNsxtPolicyFixedSegmentBasicAdvConfigUpdateTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_fixed_segment" "test" {
  display_name     = "%s"
  description      = "Acceptance Test"
  replication_mode = "MTEP"
  domain_name      = "tftest.org"
  overlay_id       = 1011
  vlan_ids         = ["101-104"]

  connectivity_path   = nsxt_policy_tier1_gateway.anotherTier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

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

func testAccNsxtPolicyFixedSegmentWithDhcpTemplate(tzName string, name string, dnsServerV4 string, dnsServerV6 string, lease string, preferred string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) +
		testAccNsxtPolicyEdgeCluster(getEdgeClusterName()) + fmt.Sprintf(`

resource "nsxt_policy_dhcp_server" "test" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  display_name      = "segment-test"
}

resource "nsxt_policy_fixed_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.anotherTier1ForSegments.path

  subnet {
    cidr = "12.12.2.1/24"
    dhcp_v4_config {
        lease_time     = %s
        server_address = "12.12.2.2/24"
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
`, name, lease, dnsServerV4, lease, preferred, dnsServerV6)
}
