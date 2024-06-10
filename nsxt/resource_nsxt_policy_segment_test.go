package nsxt

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var testPolicySegmentHelper1Name = getAccTestResourceName()
var testPolicySegmentHelper2Name = getAccTestResourceName()

func TestAccResourceNsxtPolicySegment_basicImport(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentImportTemplate(tzName, name, false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicySegment_basicImport_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentImportTemplate("", name, true),
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

func TestAccResourceNsxtPolicySegment_basicImport_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentImportTemplate("", name, false),
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

func TestAccResourceNsxtPolicySegment_basicUpdate(t *testing.T) {
	name := getAccTestResourceName()
	updatedName := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, updatedName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentBasicTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
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
				Config: testAccNsxtPolicySegmentBasicUpdateTemplate(tzName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
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

func TestAccResourceNsxtPolicySegment_connectivityPath(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentBasicTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
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
				Config: testAccNsxtPolicySegmentUpdateConnectivityTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
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

func TestAccResourceNsxtPolicySegment_updateAdvConfig(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccNSXVersion(t, "3.1.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentBasicAdvConfigTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "OFF"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "true"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.urpf_mode", "STRICT"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentBasicAdvConfigUpdateTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", "12.12.2.1/24"),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.connectivity", "ON"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.local_egress", "false"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.0.urpf_mode", "NONE"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySegment_noTransportZone(t *testing.T) {
	testAccResourceNsxtPolicySegmentNoTransportZone(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyGlobalManager(t)
	})
}

func TestAccResourceNsxtPolicySegment_noTransportZone_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicySegmentNoTransportZone(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicySegment_noTransportZone_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicySegmentNoTransportZone(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicySegmentNoTransportZone(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	cidr := "4003::1/64"
	updatedCidr := "4004::1/64"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentNoTransportZoneTemplate(name, cidr, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", cidr),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentNoTransportZoneTemplate(name, updatedCidr, withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "subnet.0.cidr", updatedCidr),
					resource.TestCheckResourceAttr(testResourceName, "domain_name", "tftest.org"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_id", "1011"),
					resource.TestCheckResourceAttr(testResourceName, "vlan_ids.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "advanced_config.#", "0"),
				),
			},
		},
	})
}

// TODO: Rewrite this test based on profile resources when these are available.
var testAccSegmentQosProfileName = getAccTestResourceName()

func TestAccResourceNsxtPolicySegment_withProfiles(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			err := testAccNsxtPolicySegmentCheckDestroy(state, name)
			if err != nil {
				return err
			}

			return testAccDataSourceNsxtPolicyQosProfileDeleteByName(testAccSegmentQosProfileName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyQosProfileCreate(testAccSegmentQosProfileName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicySegmentWithProfilesTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "security_profile.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "security_profile.0.spoofguard_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "security_profile.0.security_profile_path", ""),
					resource.TestCheckResourceAttr(testResourceName, "qos_profile.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "qos_profile.0.qos_profile_path"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentWithProfilesUpdateTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "security_profile.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "security_profile.0.spoofguard_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "security_profile.0.security_profile_path"),
					resource.TestCheckResourceAttr(testResourceName, "qos_profile.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "discovery_profile.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "discovery_profile.0.ip_discovery_profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "discovery_profile.0.mac_discovery_profile_path"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentWithProfilesRemoveAll(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "security_profile.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "qos_profile.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "discovery_profile.#", "0"),
				),
			},
		},
	})
}

var testAccSegmentBridgeProfileName = getAccTestResourceName()

func TestAccResourceNsxtPolicySegment_withBridge(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()
	vlanTzName := getVlanTransportZoneName()
	vlan := "12"
	vlanUpdated := "20"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			err := testAccNsxtPolicySegmentCheckDestroy(state, name)
			if err != nil {
				return err
			}

			return testAccDataSourceNsxtPolicyBridgeProfileDeleteByName(testAccSegmentBridgeProfileName)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyBridgeProfileCreate(testAccSegmentBridgeProfileName); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicySegmentWithBridgeTemplate(tzName, vlanTzName, name, vlan),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "bridge_config.0.profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "bridge_config.0.transport_zone_path"),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.0.vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.0.vlan_ids.0", vlan),
				),
			},
			{
				Config: testAccNsxtPolicySegmentWithBridgeTemplate(tzName, vlanTzName, name, vlanUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.#", "1"),
					resource.TestCheckResourceAttrSet(testResourceName, "bridge_config.0.profile_path"),
					resource.TestCheckResourceAttrSet(testResourceName, "bridge_config.0.transport_zone_path"),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.0.vlan_ids.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.0.vlan_ids.0", vlanUpdated),
				),
			},
			{
				Config: testAccNsxtPolicySegmentWithBridgeRemoveAll(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "bridge_config.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicySegment_withDhcp(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	leaseTimes := []string{"3600", "36000"}
	preferredTimes := []string{"3200", "32000"}
	dnsServersV4 := []string{"2.2.2.2", "3.3.3.3"}
	dnsServersV6 := []string{"2000::2", "3000::3"}
	replicationModes := []string{"SOURCE", "MTEP"}
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccPreCheck(t); testAccNSXVersion(t, "3.0.0") },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentWithDhcpTemplate(tzName, name, dnsServersV4[0], dnsServersV6[0], leaseTimes[0], preferredTimes[0], replicationModes[0]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", replicationModes[0]),
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
				Config: testAccNsxtPolicySegmentWithDhcpTemplate(tzName, name, dnsServersV4[1], dnsServersV6[1], leaseTimes[1], preferredTimes[1], replicationModes[1]),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "replication_mode", replicationModes[1]),
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

func TestAccResourceNsxtPolicySegment_withIgnoreTags(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_segment.test"
	tzName := getOverlayTransportZoneName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicySegmentCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicySegmentIgnoreTagsCreateTemplate(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "3"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentIgnoreTagsUpdate1Template(tzName, name),
				// This is an "pretend" update of ignored tag on backend
				// Where backend changes are simulated with terraform update
				// The fake backend changes are still present in intent at this point,
				// Thus terraform will detect non-empty plan
				// Next step will delete those tags from intent, and we expect them to
				// remain in the ignored_tags section
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.scopes.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.0.scope", "color"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.0.tag", "orange"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.1.scope", "size"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.1.tag", "small"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentIgnoreTagsUpdate2Template(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.0.scope", "rack"),
					resource.TestCheckResourceAttr(testResourceName, "tag.0.tag", "5"),
					resource.TestCheckResourceAttr(testResourceName, "tag.1.scope", "shape"),
					resource.TestCheckResourceAttr(testResourceName, "tag.1.tag", "triangle"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.scopes.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.0.scope", "color"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.0.tag", "orange"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.1.scope", "size"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.0.detected.1.tag", "small"),
				),
			},
			{
				Config: testAccNsxtPolicySegmentIgnoreTagsUpdate3Template(tzName, name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicySegmentExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "tag.0.scope", "color"),
					resource.TestCheckResourceAttr(testResourceName, "tag.0.tag", "blue"),
					resource.TestCheckResourceAttr(testResourceName, "tag.1.scope", "rack"),
					resource.TestCheckResourceAttr(testResourceName, "tag.1.tag", "7"),
					resource.TestCheckResourceAttr(testResourceName, "ignore_tags.#", "0"),
				),
			},
		},
	})
}

// TODO: add tests for l2_extension; requires L2 VPN Session

func testAccNsxtPolicySegmentExists(resourceName string) resource.TestCheckFunc {
	context := testAccGetSessionContext()
	return testAccNsxtPolicyResourceExists(context, resourceName, resourceNsxtPolicySegmentExists(context, "", false))
}

func testAccNsxtPolicySegmentCheckDestroy(state *terraform.State, displayName string) error {
	context := testAccGetSessionContext()
	return testAccNsxtPolicyResourceCheckDestroy(context, state, displayName, "nsxt_policy_segment", resourceNsxtPolicySegmentExists(context, "", false))
}

func testAccNsxtPolicySegmentDeps(tzName string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	s := fmt.Sprintf(`

resource "nsxt_policy_tier1_gateway" "tier1ForSegments" {
%s
  display_name              = "%s"
  description               = "Acceptance Test"
  default_rule_logging      = "true"
  enable_firewall           = "false"
  enable_standby_relocation = "false"
  force_whitelisting        = "false"
  failover_mode             = "NON_PREEMPTIVE"
  pool_allocation           = "ROUTING"
  route_advertisement_types = ["TIER1_STATIC_ROUTES", "TIER1_CONNECTED"]
}

resource "nsxt_policy_tier1_gateway" "anotherTier1ForSegments" {
%s
  display_name              = "%s"
  description               = "Another Tier1"
  default_rule_logging      = "true"
  enable_firewall           = "false"
  enable_standby_relocation = "false"
  force_whitelisting        = "false"
  failover_mode             = "NON_PREEMPTIVE"
  pool_allocation           = "ROUTING"
  route_advertisement_types = ["TIER1_STATIC_ROUTES", "TIER1_CONNECTED"]
}`, context, testPolicySegmentHelper1Name, context, testPolicySegmentHelper2Name)

	if tzName == "" {
		return s
	}
	return testAccNSXPolicyTransportZoneReadTemplate(tzName, false, true) + s
}

func testAccNsxtPolicySegmentImportTemplate(tzName string, name string, withContext bool) string {
	context := ""
	tzSetting := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	} else {
		tzSetting = "transport_zone_path = data.nsxt_policy_transport_zone.test.path"
	}

	return testAccNsxtPolicySegmentDeps(tzName, withContext) + fmt.Sprintf(`
resource "nsxt_policy_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path
  %s

  subnet {
     cidr = "12.12.2.1/24"
  }
}
`, context, name, tzSetting)
}

func testAccNsxtPolicySegmentBasicTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "orange"
  }
}
`, name)
}

func testAccNsxtPolicySegmentIgnoreTagsCreateTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "orange"
  }

  tag {
    scope = "shape"
    tag   = "triangle"
  }

  tag {
    scope = "size"
    tag   = "small"
  }
}
`, name)
}

func testAccNsxtPolicySegmentIgnoreTagsUpdate1Template(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "orange"
  }

  tag {
    scope = "shape"
    tag   = "triangle"
  }

  tag {
    scope = "size"
    tag   = "small"
  }

  ignore_tags {
    scopes = ["color", "size"]
  }
}
`, name)
}

func testAccNsxtPolicySegmentIgnoreTagsUpdate2Template(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "rack"
    tag   = "5"
  }

  tag {
    scope = "shape"
    tag   = "triangle"
  }

  ignore_tags {
    scopes = ["color", "size"]
  }
}
`, name)
}

func testAccNsxtPolicySegmentIgnoreTagsUpdate3Template(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "12.12.2.1/24"
  }

  tag {
    scope = "color"
    tag   = "blue"
  }
 
  tag {
    scope = "rack"
    tag   = "7"
  }
}
`, name)
}

func testAccNsxtPolicySegmentBasicUpdateTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test2"
  domain_name         = "tftest2.org"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
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
`, name)
}

func testAccNsxtPolicySegmentUpdateConnectivityTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  description         = "Acceptance Test2"
  domain_name         = "tftest2.org"
  overlay_id          = 1011
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
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

func testAccNsxtPolicySegmentWithProfileDeps(tzName string) string {
	return testAccNSXPolicyTransportZoneReadTemplate(tzName, false, true) + fmt.Sprintf(`
data "nsxt_policy_qos_profile" "test" {
    display_name = "%s"
}

data "nsxt_policy_segment_security_profile" "test" {
    display_name = "default-segment-security-profile"
}

data "nsxt_policy_spoofguard_profile" "test" {
    display_name = "default-spoofguard-profile"
}

data "nsxt_policy_ip_discovery_profile" "test" {
    display_name = "default-ip-discovery-profile"
}

data "nsxt_policy_mac_discovery_profile" "test" {
    display_name = "default-mac-discovery-profile"
}

`, testAccSegmentQosProfileName)
}

func testAccNsxtPolicySegmentWithProfilesTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentWithProfileDeps(tzName) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  security_profile {
    spoofguard_profile_path = data.nsxt_policy_spoofguard_profile.test.path
  }

  qos_profile {
    qos_profile_path = data.nsxt_policy_qos_profile.test.path
  }

}
`, name)
}

func testAccNsxtPolicySegmentWithProfilesUpdateTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentWithProfileDeps(tzName) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  security_profile {
    spoofguard_profile_path = data.nsxt_policy_spoofguard_profile.test.path
    security_profile_path   = data.nsxt_policy_segment_security_profile.test.path
  }

  discovery_profile {
    ip_discovery_profile_path = data.nsxt_policy_ip_discovery_profile.test.path
    mac_discovery_profile_path   = data.nsxt_policy_mac_discovery_profile.test.path
  }
}
`, name)
}

func testAccNsxtPolicySegmentWithProfilesRemoveAll(tzName string, name string) string {
	return testAccNsxtPolicySegmentWithProfileDeps(tzName) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
}
`, name)
}

func testAccNsxtPolicySegmentWithBridgeTemplate(tzName string, bridgeTzName string, name string, vlan string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`
data "nsxt_policy_bridge_profile" "test" {
  display_name = "%s"
}

data "nsxt_policy_transport_zone" "bridge" {
  display_name = "%s"
}

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path

  bridge_config {
    profile_path        = data.nsxt_policy_bridge_profile.test.path
    transport_zone_path = data.nsxt_policy_transport_zone.bridge.path
    vlan_ids            = ["%s"]
  }

}
`, testAccSegmentBridgeProfileName, bridgeTzName, name, vlan)
}

func testAccNsxtPolicySegmentWithBridgeRemoveAll(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`
resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
}
`, name)
}

func testAccNsxtPolicySegmentBasicAdvConfigTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  domain_name  = "tftest.org"
  overlay_id   = 1011
  vlan_ids     = ["101", "102"]

  transport_zone_path = data.nsxt_policy_transport_zone.test.path

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
    urpf_mode    = "STRICT"
  }
}
`, name)
}

func testAccNsxtPolicySegmentBasicAdvConfigUpdateTemplate(tzName string, name string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) + fmt.Sprintf(`

resource "nsxt_policy_segment" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  domain_name  = "tftest.org"
  overlay_id   = 1011
  vlan_ids     = ["101-104"]

  transport_zone_path = data.nsxt_policy_transport_zone.test.path

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
    urpf_mode    = "NONE"
  }
}
`, name)
}

func testAccNsxtPolicyEdgeCluster(name string) string {
	if testAccIsGlobalManager() {
		return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
  site_path    = data.nsxt_policy_site.test.path
}`, name)
	}
	return fmt.Sprintf(`
data "nsxt_policy_edge_cluster" "EC" {
  display_name = "%s"
}`, name)
}

func testAccNsxtPolicySegmentWithDhcpTemplate(tzName string, name string, dnsServerV4 string, dnsServerV6 string, lease string, preferred string, replicationMode string) string {
	return testAccNsxtPolicySegmentDeps(tzName, false) +
		testAccNsxtPolicyEdgeCluster(getEdgeClusterName()) + fmt.Sprintf(`

resource "nsxt_policy_dhcp_server" "test" {
  edge_cluster_path = data.nsxt_policy_edge_cluster.EC.path
  display_name      = "segment-test"
}

resource "nsxt_policy_segment" "test" {
  display_name        = "%s"
  transport_zone_path = data.nsxt_policy_transport_zone.test.path
  replication_mode    = "%s"

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
          network  = "3.1.1.21/24"
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
`, name, replicationMode, lease, dnsServerV4, lease, preferred, dnsServerV6)
}

func testAccNsxtPolicySegmentNoTransportZoneTemplate(name string, cidr string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`

resource "nsxt_policy_tier1_gateway" "tier1ForSegments" {
%s
  display_name              = "terraform-segment-test"
  failover_mode             = "NON_PREEMPTIVE"
}

resource "nsxt_policy_segment" "test" {
%s
  display_name        = "%s"
  description         = "Acceptance Test"
  domain_name         = "tftest.org"
  overlay_id          = 1011
  connectivity_path   = nsxt_policy_tier1_gateway.tier1ForSegments.path

  subnet {
     cidr = "%s"
  }
}
`, context, context, name, cidr)
}
