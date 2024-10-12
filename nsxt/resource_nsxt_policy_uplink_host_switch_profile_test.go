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

var accTestUplinkHostSwitchProfileCreateAttributes = map[string]string{
	"display_name":                 getAccTestResourceName(),
	"description":                  "terraform created",
	"mtu":                          "1400",
	"overlay_encap":                "GENEVE",
	"transport_vlan":               "1",
	"lag_load_balance_algorithm":   "SRCDESTIPVLAN",
	"lag_mode":                     "ACTIVE",
	"lag_name":                     "lag-name",
	"lag_number_of_uplinks":        "2",
	"lag_timeout_type":             "SLOW",
	"teaming_al_uplink_name":       "t-uplink-name1",
	"teaming_al_uplink_type":       "PNIC",
	"teaming_sl_uplink_name":       "t-uplink-name2",
	"teaming_sl_uplink_type":       "PNIC",
	"teaming_policy":               "FAILOVER_ORDER",
	"named_teaming_name":           "nt-name",
	"named_teaming_al_uplink_name": "nt-uplink-name1",
	"named_teaming_al_uplink_type": "PNIC",
	"named_teaming_sl_uplink_name": "nt-uplink-name2",
	"named_teaming_sl_uplink_type": "PNIC",
	"named_teaming_policy":         "FAILOVER_ORDER",
}

var accTestUplinkHostSwitchProfileUpdateAttributes = map[string]string{
	"display_name":                 getAccTestResourceName(),
	"description":                  "terraform updated",
	"mtu":                          "1500",
	"overlay_encap":                "VXLAN",
	"transport_vlan":               "1",
	"lag_load_balance_algorithm":   "SRCMAC",
	"lag_mode":                     "PASSIVE",
	"lag_name":                     "lag-name",
	"lag_number_of_uplinks":        "2",
	"lag_timeout_type":             "FAST",
	"teaming_al_uplink_name":       "t-uplink-name3",
	"teaming_al_uplink_type":       "PNIC",
	"teaming_sl_uplink_name":       "t-uplink-name4",
	"teaming_sl_uplink_type":       "PNIC",
	"teaming_policy":               "FAILOVER_ORDER",
	"named_teaming_name":           "nt-name",
	"named_teaming_al_uplink_name": "nt-uplink-name3",
	"named_teaming_al_uplink_type": "PNIC",
	"named_teaming_sl_uplink_name": "nt-uplink-name4",
	"named_teaming_sl_uplink_type": "PNIC",
	"named_teaming_policy":         "FAILOVER_ORDER",
}

func TestAccResourceNsxtPolicyUplinkHostSwitchProfile_basic(t *testing.T) {
	testResourceName := "nsxt_policy_uplink_host_switch_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtUplinkHostSwitchProfileCheckDestroy(state, accTestUplinkHostSwitchProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtUplinkHostSwitchProfileTemplate(true),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtUplinkHostSwitchProfileExists(accTestUplinkHostSwitchProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestUplinkHostSwitchProfileCreateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestUplinkHostSwitchProfileCreateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "mtu", accTestUplinkHostSwitchProfileCreateAttributes["mtu"]),
					resource.TestCheckResourceAttr(testResourceName, "overlay_encap", accTestUplinkHostSwitchProfileCreateAttributes["overlay_encap"]),
					resource.TestCheckResourceAttr(testResourceName, "transport_vlan", accTestUplinkHostSwitchProfileCreateAttributes["transport_vlan"]),

					resource.TestCheckResourceAttr(testResourceName, "lag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.load_balance_algorithm", accTestUplinkHostSwitchProfileCreateAttributes["lag_load_balance_algorithm"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.mode", accTestUplinkHostSwitchProfileCreateAttributes["lag_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.name", accTestUplinkHostSwitchProfileCreateAttributes["lag_name"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.number_of_uplinks", accTestUplinkHostSwitchProfileCreateAttributes["lag_number_of_uplinks"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.timeout_type", accTestUplinkHostSwitchProfileCreateAttributes["lag_timeout_type"]),

					resource.TestCheckResourceAttr(testResourceName, "teaming.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.0.uplink_name", accTestUplinkHostSwitchProfileCreateAttributes["teaming_al_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.0.uplink_type", accTestUplinkHostSwitchProfileCreateAttributes["teaming_al_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.0.uplink_name", accTestUplinkHostSwitchProfileCreateAttributes["teaming_sl_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.0.uplink_type", accTestUplinkHostSwitchProfileCreateAttributes["teaming_sl_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.policy", accTestUplinkHostSwitchProfileCreateAttributes["teaming_policy"]),

					resource.TestCheckResourceAttr(testResourceName, "named_teaming.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.0.uplink_name", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_al_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.0.uplink_type", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_al_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.0.uplink_name", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_sl_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.0.uplink_type", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_sl_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.policy", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_policy"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.name", accTestUplinkHostSwitchProfileCreateAttributes["named_teaming_name"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtUplinkHostSwitchProfileTemplate(false),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtUplinkHostSwitchProfileExists(accTestUplinkHostSwitchProfileUpdateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", accTestUplinkHostSwitchProfileUpdateAttributes["display_name"]),
					resource.TestCheckResourceAttr(testResourceName, "description", accTestUplinkHostSwitchProfileUpdateAttributes["description"]),

					resource.TestCheckResourceAttr(testResourceName, "mtu", accTestUplinkHostSwitchProfileUpdateAttributes["mtu"]),
					resource.TestCheckResourceAttr(testResourceName, "overlay_encap", accTestUplinkHostSwitchProfileUpdateAttributes["overlay_encap"]),
					resource.TestCheckResourceAttr(testResourceName, "transport_vlan", accTestUplinkHostSwitchProfileUpdateAttributes["transport_vlan"]),

					resource.TestCheckResourceAttr(testResourceName, "lag.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.load_balance_algorithm", accTestUplinkHostSwitchProfileUpdateAttributes["lag_load_balance_algorithm"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.mode", accTestUplinkHostSwitchProfileUpdateAttributes["lag_mode"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.name", accTestUplinkHostSwitchProfileUpdateAttributes["lag_name"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.number_of_uplinks", accTestUplinkHostSwitchProfileUpdateAttributes["lag_number_of_uplinks"]),
					resource.TestCheckResourceAttr(testResourceName, "lag.0.timeout_type", accTestUplinkHostSwitchProfileUpdateAttributes["lag_timeout_type"]),

					resource.TestCheckResourceAttr(testResourceName, "teaming.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.0.uplink_name", accTestUplinkHostSwitchProfileUpdateAttributes["teaming_al_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.active.0.uplink_type", accTestUplinkHostSwitchProfileUpdateAttributes["teaming_al_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.0.uplink_name", accTestUplinkHostSwitchProfileUpdateAttributes["teaming_sl_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.standby.0.uplink_type", accTestUplinkHostSwitchProfileUpdateAttributes["teaming_sl_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "teaming.0.policy", accTestUplinkHostSwitchProfileUpdateAttributes["teaming_policy"]),

					resource.TestCheckResourceAttr(testResourceName, "named_teaming.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.0.uplink_name", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_al_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.active.0.uplink_type", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_al_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.#", "1"),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.0.uplink_name", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_sl_uplink_name"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.standby.0.uplink_type", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_sl_uplink_type"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.policy", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_policy"]),
					resource.TestCheckResourceAttr(testResourceName, "named_teaming.0.name", accTestUplinkHostSwitchProfileUpdateAttributes["named_teaming_name"]),

					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNsxtUplinkHostSwitchProfileMinimalistic(),
				Check: resource.ComposeTestCheckFunc(
					testAccNsxtUplinkHostSwitchProfileExists(accTestUplinkHostSwitchProfileCreateAttributes["display_name"], testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "description", ""),
					resource.TestCheckResourceAttrSet(testResourceName, "nsx_id"),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
					resource.TestCheckResourceAttrSet(testResourceName, "revision"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "0"),

					resource.TestCheckNoResourceAttr(testResourceName, "lag.#"),
					resource.TestCheckNoResourceAttr(testResourceName, "named_teaming.#"),
					resource.TestCheckNoResourceAttr(testResourceName, "teaming.0.standby.#"),
					resource.TestCheckResourceAttr(testResourceName, "mtu", "0"),
					resource.TestCheckResourceAttr(testResourceName, "overlay_encap", "GENEVE"),
					resource.TestCheckResourceAttr(testResourceName, "transport_vlan", "0"),
				),
			},
		},
	})
}

func TestAccResourceNsxtPolicyUplinkHostSwitchProfile_importBasic(t *testing.T) {
	testResourceName := "nsxt_policy_uplink_host_switch_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccOnlyLocalManager(t)
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNsxtUplinkHostSwitchProfileCheckDestroy(state, accTestUplinkHostSwitchProfileUpdateAttributes["display_name"])
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNsxtUplinkHostSwitchProfileMinimalistic(),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNsxtUplinkHostSwitchProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("UplinkHostSwitchProfile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("UplinkHostSwitchProfile resource ID not set in resources")
		}

		exists, err := resourceNsxtUplinkHostSwitchProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("UplinkHostSwitchProfile %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtUplinkHostSwitchProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != "nsxt_policy_uplink_host_switch_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := resourceNsxtUplinkHostSwitchProfileExists(resourceID, connector, testAccIsGlobalManager())
		if err == nil {
			return err
		}

		if exists {
			return fmt.Errorf("UplinkHostSwitchProfile %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtUplinkHostSwitchProfileTemplate(createFlow bool) string {
	var attrMap map[string]string
	if createFlow {
		attrMap = accTestUplinkHostSwitchProfileCreateAttributes
	} else {
		attrMap = accTestUplinkHostSwitchProfileUpdateAttributes
	}
	return fmt.Sprintf(`
resource "nsxt_policy_uplink_host_switch_profile" "test" {
  display_name = "%s"
  description  = "%s"

  mtu            = %s
  transport_vlan = %s
  overlay_encap  = "%s"
  lag {
    name                   = "%s"
    load_balance_algorithm = "%s"
    mode                   = "%s"
    number_of_uplinks      = %s
    timeout_type           = "%s"
  }
  teaming {
    active {
	  uplink_name = "%s"
	  uplink_type = "%s"
    }
    standby {
	  uplink_name = "%s"
	  uplink_type = "%s"
    }
    policy        = "%s"
  }
  named_teaming {
    active {
  	  uplink_name = "%s"
	  uplink_type = "%s"
    }
    standby {
      uplink_name = "%s"
	  uplink_type = "%s"
    }
    policy        = "%s"
    name          = "%s"
  }

  tag {
    scope = "scope1"
    tag   = "tag1"
  }
}

data "nsxt_policy_uplink_host_switch_profile" "test" {
  display_name = "%s"
  depends_on = [nsxt_policy_uplink_host_switch_profile.test]
}`, attrMap["display_name"], attrMap["description"], attrMap["mtu"], attrMap["transport_vlan"], attrMap["overlay_encap"], attrMap["lag_name"], attrMap["lag_load_balance_algorithm"], attrMap["lag_mode"], attrMap["lag_number_of_uplinks"], attrMap["lag_timeout_type"], attrMap["teaming_al_uplink_name"], attrMap["teaming_al_uplink_type"], attrMap["teaming_sl_uplink_name"], attrMap["teaming_sl_uplink_type"], attrMap["teaming_policy"], attrMap["named_teaming_al_uplink_name"], attrMap["named_teaming_al_uplink_type"], attrMap["named_teaming_sl_uplink_name"], attrMap["named_teaming_sl_uplink_type"], attrMap["named_teaming_policy"], attrMap["named_teaming_name"], attrMap["display_name"])
}

func testAccNsxtUplinkHostSwitchProfileMinimalistic() string {
	return fmt.Sprintf(`
resource "nsxt_policy_uplink_host_switch_profile" "test" {
  display_name = "%s"
  teaming {
    active {
      uplink_name = "%s"
      uplink_type = "%s"
    }
    policy = "%s"
  }
}`, accTestUplinkHostSwitchProfileUpdateAttributes["display_name"], accTestUplinkHostSwitchProfileUpdateAttributes["teaming_al_uplink_name"], accTestUplinkHostSwitchProfileUpdateAttributes["teaming_al_uplink_type"], accTestUplinkHostSwitchProfileUpdateAttributes["teaming_policy"])
}
