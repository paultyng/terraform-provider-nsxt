/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func TestAccResourceNsxtPolicyService_icmp(t *testing.T) {
	testAccResourceNsxtPolicyServiceIcmp(t, false, func() {
		testAccPreCheck(t)
	})
}

func TestAccResourceNsxtPolicyService_icmp_multitenancy(t *testing.T) {
	testAccResourceNsxtPolicyServiceIcmp(t, true, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancy(t)
	})
}

func TestAccResourceNsxtPolicyService_icmp_multitenancyProvider(t *testing.T) {
	testAccResourceNsxtPolicyServiceIcmp(t, false, func() {
		testAccPreCheck(t)
		testAccOnlyMultitenancyProvider(t)
	})
}

func testAccResourceNsxtPolicyServiceIcmp(t *testing.T, withContext bool, preCheck func()) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  preCheck,
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				// Step 0: Single ICMP entry
				Config: testAccNsxtPolicyIcmpTypeServiceCreateTypeCodeTemplate(name, "3", "1", "ICMPv4", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", "3"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 1: Update service & entry
				Config: testAccNsxtPolicyIcmpTypeServiceCreateTypeOnlyTemplate(updateName, "34", "ICMPv4", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", "34"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 2: Update service & entry for ICMPv6
				Config: testAccNsxtPolicyIcmpTypeServiceCreateTypeOnlyTemplate(name, "3", "ICMPv6", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", "3"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv6"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 3: Update service & entry with no type & code
				Config: testAccNsxtPolicyIcmpTypeServiceCreateNoTypeCodeTemplate(name, "ICMPv4", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 4: Add another ICMP service entry
				Config: testAccNsxtPolicyIcmpTypeServiceCreate2Template(name, "3", "1", "ICMPv4", withContext),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Updated Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.1.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.1.description", "Updated Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.1.icmp_type", "3"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.1.icmp_code", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.1.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", "no-type-code"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Updated Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_icmpNoEntryDisplayName(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIcmpTypeServiceCreateNoDisplayNameTemplate(name, "3", "1", "ICMPv4"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Acceptance Test"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", ""),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_type", "3"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.icmp_code", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_l4PortSet(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				// Step 0: Single TCP L4 port set entry
				Config: testAccNsxtPolicyL4PortSetTypeServiceCreateTemplate(name, "TCP", "99-101"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "l4 service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.source_ports.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 1: Single UDP L4 port set entry
				Config: testAccNsxtPolicyL4PortSetTypeServiceCreateTemplate(name, "UDP", "88"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "l4 service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "UDP"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.source_ports.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 2: 2 service Entries
				Config: testAccNsxtPolicyL4PortSetTypeServiceUpdateTemplate(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "updated l4 service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", "entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.1.display_name", "entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.1.description", "Entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.1.protocol", "UDP"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 2: Update back to only one entry
				Config: testAccNsxtPolicyL4PortSetTypeServiceCreateTemplate(name, "UDP", "88"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "l4 service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "UDP"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.destination_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.source_ports.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_mixedServices(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				// Step 0: mixed service entries
				Config: testAccNsxtPolicyMixedServiceCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", "entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.display_name", "entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.description", "Entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.0.protocol", "ICMPv4"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				// Step 1: Single UDP L4 port set entry
				Config: testAccNsxtPolicyL4PortSetTypeServiceCreateTemplate(name, "UDP", "88"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_igmp(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIgmpTypeServiceCreateTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "IGMP entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyIgmpTypeServiceCreateTemplate(updateName),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "IGMP entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.0.display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_etherType(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyEtherTypeServiceCreateTemplate(name, "1536"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Ether type entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.ether_type", "1536"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyEtherTypeServiceCreateTemplate(updateName, "2001"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Ether type entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.0.ether_type", "2001"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_ipProtocolType(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIPProtocolTypeServiceCreateTemplate(name, "6"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "IP Protocol type entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.protocol", "6"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyIPProtocolTypeServiceCreateTemplate(updateName, "17"),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "IP Protocol type entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.0.protocol", "17"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_algType(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"
	alg := "SUN_RPC_UDP"
	destPort := "210"
	updatedDestPort := "211"
	sourcePorts := "9000-9001"
	updatedSourcePorts := "9000-9001\", \"500-504"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyAlgServiceCreateTemplate(name, alg, sourcePorts, destPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Algorithm entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.algorithm", alg),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.destination_port", destPort),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.source_ports.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
				),
			},
			{
				Config: testAccNsxtPolicyAlgServiceCreateTemplate(updateName, alg, updatedSourcePorts, updatedDestPort),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Algorithm entry"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.display_name", updateName),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.description", "Entry-1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.algorithm", alg),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.source_ports.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.0.destination_port", updatedDestPort),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_nestedServiceType(t *testing.T) {
	name := getAccTestResourceName()
	updateName := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"
	testNestedService1Name := "HTTP"
	testNestedService2Name := "HTTPS"
	regexpService1Name, err := regexp.Compile("/.*/" + testNestedService1Name)

	if err != nil {
		fmt.Printf("Error while compiling regexp: %v", err)
	}

	regexpService2Name, err := regexp.Compile("/.*/" + testNestedService2Name)

	if err != nil {
		fmt.Printf("Error while compiling regexp: %v", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, updateName)
		},
		Steps: []resource.TestStep{
			{
				// Step 0: Create a nested service
				Config: testAccNsxtPolicyNestedServiceCreateTemplate(name, testNestedService1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Nested service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.display_name", testNestedService1Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.description", "Entry-1"),
					resource.TestMatchResourceAttr(testResourceName, "nested_service_entry.0.nested_service_path", regexpService1Name),
				),
			},
			{
				// Step 1: Add another nested service
				Config: testAccNsxtPolicyNestedServiceUpdateTemplate(name, testNestedService1Name, testNestedService2Name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Nested service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.#", "2"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.display_name", testNestedService1Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.description", "Entry-1"),
					resource.TestMatchResourceAttr(testResourceName, "nested_service_entry.0.nested_service_path", regexpService1Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.1.display_name", testNestedService2Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.1.description", "Entry-2"),
					resource.TestMatchResourceAttr(testResourceName, "nested_service_entry.1.nested_service_path", regexpService2Name),
				),
			},
			{
				// Step 2: Remove nested service 2
				Config: testAccNsxtPolicyNestedServiceCreateTemplate(name, testNestedService1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Nested service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.display_name", testNestedService1Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.description", "Entry-1"),
					resource.TestMatchResourceAttr(testResourceName, "nested_service_entry.0.nested_service_path", regexpService1Name),
				),
			},
			{
				// Step 3: Mix with other service types
				Config: testAccNsxtPolicyNestedServiceMixedTemplate(name, testNestedService1Name),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtPolicyServiceExists(testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", "Nested service"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "algorithm_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.display_name", "entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.description", "Entry-2"),
					resource.TestCheckResourceAttr(testResourceName, "l4_port_set_entry.0.protocol", "TCP"),
					resource.TestCheckResourceAttr(testResourceName, "icmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "igmp_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ether_type_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "ip_protocol_entry.#", "0"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.display_name", testNestedService1Name),
					resource.TestCheckResourceAttr(testResourceName, "nested_service_entry.0.description", "Entry-1"),
					resource.TestMatchResourceAttr(testResourceName, "nested_service_entry.0.nested_service_path", regexpService1Name),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIcmpTypeServiceCreateNoTypeCodeTemplate(name, "ICMPv4", false),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceNsxtPolicyService_importBasic_multitenancy(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancy(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIcmpTypeServiceCreateNoTypeCodeTemplate(name, "ICMPv4", true),
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

func TestAccResourceNsxtPolicyService_importBasic_multitenancyProvider(t *testing.T) {
	name := getAccTestResourceName()
	testResourceName := "nsxt_policy_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyMultitenancyProvider(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtPolicyServiceCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtPolicyIcmpTypeServiceCreateNoTypeCodeTemplate(name, "ICMPv4", false),
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

func testAccNsxtPolicyServiceExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy service resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy service resource ID not set in resources")
		}

		nsxClient := infra.NewServicesClient(testAccGetSessionContext(), connector)
		_, err := nsxClient.Get(resourceID)
		if err != nil {
			return fmt.Errorf("Error while retrieving policy service ID %s. Error: %v", resourceID, err)
		}

		return nil
	}
}

func testAccNsxtPolicyServiceCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_service" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		nsxClient := infra.NewServicesClient(testAccGetSessionContext(), connector)
		_, err := nsxClient.Get(resourceID)
		if err == nil {
			return fmt.Errorf("Policy service %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyIcmpTypeServiceCreateTypeCodeTemplate(name string, icmpType string, icmpCode string, protocol string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}

	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  icmp_entry {
	display_name = "%s"
	description  = "Entry"
	icmp_type    = "%s"
	icmp_code    = "%s"
	protocol     = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, name, name, icmpType, icmpCode, protocol)
}

func testAccNsxtPolicyIcmpTypeServiceCreateTypeOnlyTemplate(name string, icmpType string, protocol string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  icmp_entry {
	display_name = "%s"
	description  = "Entry"
	icmp_type    = "%s"
	protocol     = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, name, name, icmpType, protocol)
}

func testAccNsxtPolicyIcmpTypeServiceCreateNoTypeCodeTemplate(name string, protocol string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
%s
  display_name = "%s"
  description  = "Acceptance Test"

  icmp_entry {
    display_name = "%s"
    description  = "Entry"
    protocol     = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, context, name, name, protocol)
}

func testAccNsxtPolicyIcmpTypeServiceCreateNoDisplayNameTemplate(name string, icmpType string, icmpCode string, protocol string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  display_name = "%s"
  description  = "Acceptance Test"

  icmp_entry {
	description  = "Entry"
	icmp_type    = "%s"
	icmp_code    = "%s"
	protocol     = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, name, icmpType, icmpCode, protocol)
}

func testAccNsxtPolicyIcmpTypeServiceCreate2Template(name string, icmpType string, icmpCode string, protocol string, withContext bool) string {
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
%s
  display_name = "%s"
  description  = "Updated Acceptance Test"

  icmp_entry {
    display_name = "%s"
    description  = "Updated Entry"
    icmp_type    = "%s"
    icmp_code    = "%s"
    protocol     = "%s"
  }

  icmp_entry {
    display_name = "no-type-code"
    description  = "Updated Entry"
    protocol     = "ICMPv4"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }

}`, context, name, name, icmpType, icmpCode, protocol)
}

func testAccNsxtPolicyL4PortSetTypeServiceCreateTemplate(serviceName string, protocol string, port string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description       = "l4 service"
  display_name      = "%s"

  l4_port_set_entry {
    display_name      = "%s"
    description       = "Entry"
    protocol          = "%s"
    destination_ports = [ "%s" ]
    source_ports      = [ "100", "200-300" ]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, serviceName, protocol, port)
}

func testAccNsxtPolicyL4PortSetTypeServiceUpdateTemplate(serviceName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description       = "updated l4 service"
  display_name      = "%s"

  l4_port_set_entry {
    display_name      = "entry-1"
    description       = "Entry-1"
    protocol          = "TCP"
    destination_ports = [ "100" ]
  }

  l4_port_set_entry {
    display_name      = "entry-2"
    description       = "Entry-2"
    protocol          = "UDP"
    destination_ports = [ "101" ]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }

  tag {
    scope = "scope2"
    tag   = "tag2"
  }
}`, serviceName)
}

func testAccNsxtPolicyMixedServiceCreateTemplate(serviceName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description       = "mixed services"
  display_name      = "%s"

  l4_port_set_entry {
    display_name      = "entry-1"
    description       = "Entry-1"
    protocol          = "TCP"
    destination_ports = [ "80" ]
  }

  icmp_entry {
    display_name = "entry-2"
    description  = "Entry-2"
    protocol     = "ICMPv4"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName)
}

func testAccNsxtPolicyIgmpTypeServiceCreateTemplate(serviceName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description  = "IGMP entry"
  display_name = "%s"

  igmp_entry {
    display_name = "%s"
    description  = "Entry-1"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, serviceName)
}

func testAccNsxtPolicyEtherTypeServiceCreateTemplate(serviceName string, etherType string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description  = "Ether type entry"
  display_name = "%s"

  ether_type_entry {
    display_name = "%s"
    description  = "Entry-1"
    ether_type   = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, serviceName, etherType)
}

func testAccNsxtPolicyIPProtocolTypeServiceCreateTemplate(serviceName string, protocol string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description  = "IP Protocol type entry"
  display_name = "%s"

  ip_protocol_entry {
    display_name = "%s"
    description  = "Entry-1"
    protocol     = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, serviceName, protocol)
}

func testAccNsxtPolicyAlgServiceCreateTemplate(serviceName string, alg string, sourcePorts string, destPort string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description  = "Algorithm entry"
  display_name = "%s"

  algorithm_entry {
    display_name      = "%s"
    description       = "Entry-1"
    algorithm         = "%s"
    source_ports      = ["%s"]
    destination_port  = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, serviceName, alg, sourcePorts, destPort)
}

func testAccNsxtPolicyNestedServiceCreateTemplate(serviceName string, nestedServiceEntryName string) string {
	return fmt.Sprintf(`
resource "nsxt_policy_service" "test" {
  description  = "Nested service"
  display_name = "%s"

  nested_service_entry {
    display_name        = "%s"
    description         = "Entry-1"
	nested_service_path = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, nestedServiceEntryName, testAccAdjustPolicyInfraConfig("/infra/services/"+nestedServiceEntryName))
}

func testAccNsxtPolicyNestedServiceUpdateTemplate(serviceName string, nestedServiceEntry1Name string, nestedServiceEntry2Name string) string {
	return fmt.Sprintf(`resource "nsxt_policy_service" "test" {
  description  = "Nested service"
  display_name = "%s"

  nested_service_entry {
    display_name        = "%s"
    description         = "Entry-1"
	nested_service_path	= "%s"
  }

  nested_service_entry {
    display_name        = "%s"
    description         = "Entry-2"
	nested_service_path = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, nestedServiceEntry1Name, testAccAdjustPolicyInfraConfig("/infra/services/"+nestedServiceEntry1Name), nestedServiceEntry2Name, testAccAdjustPolicyInfraConfig("/infra/services/"+nestedServiceEntry2Name))

}

func testAccNsxtPolicyNestedServiceMixedTemplate(serviceName string, nestedServiceEntryName string) string {
	return fmt.Sprintf(`resource "nsxt_policy_service" "test" {
  description  = "Nested service"
  display_name = "%s"

  nested_service_entry {
    display_name        = "%s"
    description         = "Entry-1"
	nested_service_path = "%s"
  }

  l4_port_set_entry {
    display_name      = "entry-2"
    description       = "Entry-2"
    protocol          = "TCP"
    destination_ports = [ "443" ]
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}`, serviceName, nestedServiceEntryName, testAccAdjustPolicyInfraConfig("/infra/services/"+nestedServiceEntryName))
}
