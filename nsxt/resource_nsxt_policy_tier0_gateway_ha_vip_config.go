// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var interfacePathLen = 8
var interfaceTier0Location = 3
var interfaceLocaleSrvLocation = 5

func resourceNsxtPolicyTier0GatewayHAVipConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyTier0GatewayHAVipConfigCreate,
		Read:   resourceNsxtPolicyTier0GatewayHAVipConfigRead,
		Update: resourceNsxtPolicyTier0GatewayHAVipConfigUpdate,
		Delete: resourceNsxtPolicyTier0GatewayHAVipConfigDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0GatewayHAVipConfigImport,
		},

		Schema: map[string]*schema.Schema{
			"config": {
				Type:        schema.TypeList,
				Description: "Tier0 HA VIP Config",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Flag to enable this HA VIP config",
							Optional:    true,
							Default:     true,
						},
						"external_interface_paths": {
							Type:        schema.TypeList,
							Description: "paths to Tier0 external interfaces which are to be paired to provide redundancy",
							Required:    true,
							MinItems:    2,
							MaxItems:    2,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validatePolicyPath(),
							},
						},
						"vip_subnets": {
							Type:        schema.TypeList,
							Description: "IP address subnets which will be used as floating IP addresses",
							Required:    true,
							MinItems:    1,
							MaxItems:    2,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIPCidr(),
							},
						},
					},
				},
			},
			"locale_service_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Gateway Locale Service on NSX",
				Computed:    true,
			},
			"tier0_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Tier0 Gateway on NSX",
				Computed:    true,
			},
		},
	}
}

func getHaVipInterfaceSubnetList(configData map[string]interface{}) []model.InterfaceSubnet {
	subnets := interface2StringList(configData["vip_subnets"].([]interface{}))
	var interfaceSubnetList []model.InterfaceSubnet
	for _, subnet := range subnets {
		result := strings.Split(subnet, "/")
		var ipAddresses []string
		ipAddresses = append(ipAddresses, result[0])
		prefix, _ := strconv.Atoi(result[1])
		prefix64 := int64(prefix)
		interfaceSubnet := model.InterfaceSubnet{
			IpAddresses: ipAddresses,
			PrefixLen:   &prefix64,
		}
		interfaceSubnetList = append(interfaceSubnetList, interfaceSubnet)
	}

	return interfaceSubnetList
}

func resourceNsxtPolicyTier0GatewayHAVipConfigCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Get the tier0 id and locale-service id from the interfaces and make sure
	// all the interfaces belong to the same locale service
	var tier0ID string
	var localeServiceID string
	configs := d.Get("config").([]interface{})
	var haVipConfigs []model.Tier0HaVipConfig
	for _, config := range configs {
		configData := config.(map[string]interface{})
		interfaces := configData["external_interface_paths"].([]interface{})
		var externalInterfacePaths []string
		for _, intf := range interfaces {
			splitPath := strings.Split(intf.(string), "/")
			if len(splitPath) != interfacePathLen {
				return fmt.Errorf("Error obtaining Tier0 id or Locale Service id from interface %s", intf)
			}
			if tier0ID == "" {
				tier0ID = splitPath[interfaceTier0Location]
			} else if tier0ID != splitPath[interfaceTier0Location] {
				return fmt.Errorf("Interface %s has a different tier0 instead of %s", intf, tier0ID)
			}
			if localeServiceID == "" {
				localeServiceID = splitPath[interfaceLocaleSrvLocation]
			} else if localeServiceID != splitPath[interfaceLocaleSrvLocation] {
				return fmt.Errorf("Interface %s has a different locale-service instead of %s", intf, localeServiceID)
			}
			externalInterfacePaths = append(externalInterfacePaths, intf.(string))
		}
		enabled := configData["enabled"].(bool)
		subnets := getHaVipInterfaceSubnetList(configData)
		elem := model.Tier0HaVipConfig{
			Enabled:                &enabled,
			ExternalInterfacePaths: externalInterfacePaths,
			VipSubnets:             subnets,
		}
		haVipConfigs = append(haVipConfigs, elem)
	}

	// Update the locale service id
	lsType := "LocaleServices"
	serviceStruct := model.LocaleServices{
		Id:           &localeServiceID,
		ResourceType: &lsType,
		HaVipConfigs: haVipConfigs,
	}

	id := newUUID()

	var err error
	if isPolicyGlobalManager(m) {
		// Use patch to only update the relevant fields
		rawObj, err1 := convertModelBindingType(serviceStruct, model.LocaleServicesBindingType(), gm_model.LocaleServicesBindingType())
		if err1 != nil {
			return err1
		}
		client := gm_tier0s.NewLocaleServicesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, rawObj.(gm_model.LocaleServices))

	} else {
		client := tier_0s.NewLocaleServicesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, serviceStruct)
	}
	if err != nil {
		return handleCreateError("Tier0 HA Vip config", id, err)
	}

	d.SetId(id)
	d.Set("tier0_id", tier0ID)
	d.Set("locale_service_id", localeServiceID)

	return resourceNsxtPolicyTier0GatewayHAVipConfigRead(d, m)
}

func resourceNsxtPolicyTier0GatewayHAVipConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0ID := d.Get("tier0_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || tier0ID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	var obj model.LocaleServices
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewLocaleServicesClient(connector)
		gmObj, err1 := client.Get(tier0ID, localeServiceID)
		if err1 != nil {
			return handleReadError(d, "Tier0 HA Vip config", id, err1)
		}
		lmObj, err2 := convertModelBindingType(gmObj, model.LocaleServicesBindingType(), model.LocaleServicesBindingType())
		if err2 != nil {
			return err2
		}
		obj = lmObj.(model.LocaleServices)
	} else {
		var err error
		client := tier_0s.NewLocaleServicesClient(connector)
		obj, err = client.Get(tier0ID, localeServiceID)
		if err != nil {
			return handleReadError(d, "Tier0 HA Vip config", id, err)
		}
	}

	// Set the ha config values in the scheme
	var configList []map[string]interface{}
	for _, config := range obj.HaVipConfigs {
		elem := make(map[string]interface{})
		elem["enabled"] = config.Enabled

		var subnetList []string
		for _, subnet := range config.VipSubnets {
			cidr := fmt.Sprintf("%s/%d", subnet.IpAddresses[0], *subnet.PrefixLen)
			subnetList = append(subnetList, cidr)
		}
		elem["vip_subnets"] = subnetList
		elem["external_interface_paths"] = config.ExternalInterfacePaths

		configList = append(configList, elem)
	}
	d.Set("config", configList)

	return nil
}

func resourceNsxtPolicyTier0GatewayHAVipConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0ID := d.Get("tier0_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || tier0ID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	// Make sure all the interfaces belong to the same locale service
	configs := d.Get("config").([]interface{})
	var haVipConfigs []model.Tier0HaVipConfig
	for _, config := range configs {
		configData := config.(map[string]interface{})
		interfaces := configData["external_interface_paths"].([]interface{})
		var externalInterfacePaths []string
		for _, intf := range interfaces {
			splitPath := strings.Split(intf.(string), "/")
			if len(splitPath) != interfacePathLen {
				return fmt.Errorf("Error obtaining Tier0 id or Locale Service id from interface %s", intf)
			}
			if tier0ID != splitPath[interfaceTier0Location] {
				return fmt.Errorf("Interface %s has a different tier0 instead of %s", intf, tier0ID)
			}
			if localeServiceID != splitPath[interfaceLocaleSrvLocation] {
				return fmt.Errorf("Interface %s has a different locale-service instead of %s", intf, localeServiceID)
			}
			externalInterfacePaths = append(externalInterfacePaths, intf.(string))
		}
		enabled := configData["enabled"].(bool)
		subnets := getHaVipInterfaceSubnetList(configData)
		elem := model.Tier0HaVipConfig{
			Enabled:                &enabled,
			ExternalInterfacePaths: externalInterfacePaths,
			VipSubnets:             subnets,
		}
		haVipConfigs = append(haVipConfigs, elem)
	}

	// Update the locale service
	lsType := "LocaleServices"
	serviceStruct := model.LocaleServices{
		Id:           &localeServiceID,
		ResourceType: &lsType,
		HaVipConfigs: haVipConfigs,
	}

	var err error
	if isPolicyGlobalManager(m) {
		// Use patch to only update the relevant fields
		rawObj, err1 := convertModelBindingType(serviceStruct, model.LocaleServicesBindingType(), gm_model.LocaleServicesBindingType())
		if err1 != nil {
			return err1
		}
		client := gm_tier0s.NewLocaleServicesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, rawObj.(gm_model.LocaleServices))

	} else {
		client := tier_0s.NewLocaleServicesClient(connector)
		err = client.Patch(tier0ID, localeServiceID, serviceStruct)
	}
	if err != nil {
		return handleUpdateError("Tier0 HA Vip config", id, err)
	}

	return resourceNsxtPolicyTier0GatewayHAVipConfigRead(d, m)
}

func resourceNsxtPolicyTier0GatewayHAVipConfigDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	tier0ID := d.Get("tier0_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || tier0ID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	// Update the locale service with empty HaVipConfigs using get/post
	doUpdate := func() error {
		if isPolicyGlobalManager(m) {
			client := gm_tier0s.NewLocaleServicesClient(connector)
			gmObj, err := client.Get(tier0ID, localeServiceID)
			if err != nil {
				return err
			}
			gmObj.HaVipConfigs = nil
			_, err = client.Update(tier0ID, localeServiceID, gmObj)
			return err
		}
		client := tier_0s.NewLocaleServicesClient(connector)
		obj, err := client.Get(tier0ID, localeServiceID)
		if err != nil {
			return err
		}
		obj.HaVipConfigs = nil
		_, err = client.Update(tier0ID, localeServiceID, obj)
		return err
	}

	commonProviderConfig := getCommonProviderConfig(m)
	err := retryUponPreconditionFailed(doUpdate, commonProviderConfig.MaxRetries)
	if err != nil {
		return handleDeleteError("Tier0 HA Vip config", id, err)
	}

	return nil
}

func resourceNsxtPolicyTier0GatewayHAVipConfigImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <gateway-id>/<locale-service-id> as an input")
	}

	tier0ID := s[0]
	localeServiceID := s[1]
	connector := getPolicyConnector(m)
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewLocaleServicesClient(connector)
		obj, err := client.Get(tier0ID, localeServiceID)
		if err != nil || obj.HaVipConfigs == nil {
			return nil, err
		}
	} else {
		var err error
		client := tier_0s.NewLocaleServicesClient(connector)
		obj, err := client.Get(tier0ID, localeServiceID)
		if err != nil || obj.HaVipConfigs == nil {
			return nil, err
		}
	}

	d.Set("tier0_id", tier0ID)
	d.Set("locale_service_id", localeServiceID)
	d.SetId(newUUID())

	return []*schema.ResourceData{d}, nil
}
