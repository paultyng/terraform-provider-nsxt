// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var switchSecurityProfileBpduMacs = []string{
	"01:80:c2:00:00:00", "01:80:c2:00:00:01", "01:80:c2:00:00:02",
	"01:80:c2:00:00:03", "01:80:c2:00:00:04", "01:80:c2:00:00:05",
	"01:80:c2:00:00:06", "01:80:c2:00:00:07", "01:80:c2:00:00:08",
	"01:80:c2:00:00:09", "01:80:c2:00:00:0a", "01:80:c2:00:00:0b",
	"01:80:c2:00:00:0c", "01:80:c2:00:00:0d", "01:80:c2:00:00:0e",
	"01:80:c2:00:00:0f", "00:e0:2b:00:00:00", "00:e0:2b:00:00:04",
	"00:e0:2b:00:00:06", "01:00:0c:00:00:00", "01:00:0c:cc:cc:cc",
	"01:00:0c:cc:cc:cd", "01:00:0c:cd:cd:cd", "01:00:0c:cc:cc:c0",
	"01:00:0c:cc:cc:c1", "01:00:0c:cc:cc:c2", "01:00:0c:cc:cc:c3",
	"01:00:0c:cc:cc:c4", "01:00:0c:cc:cc:c5", "01:00:0c:cc:cc:c6",
	"01:00:0c:cc:cc:c7"}

func resourceNsxtSwitchSecuritySwitchingProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtSwitchSecuritySwitchingProfileCreate,
		Read:   resourceNsxtSwitchSecuritySwitchingProfileRead,
		Update: resourceNsxtSwitchSecuritySwitchingProfileUpdate,
		Delete: resourceNsxtSwitchSecuritySwitchingProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
			"block_non_ip": {
				Type:        schema.TypeBool,
				Description: "Block all traffic except IP/(G)ARP/BPDU",
				Optional:    true,
				Default:     false,
			},
			"block_client_dhcp": {
				Type:        schema.TypeBool,
				Description: "Indicates whether DHCP client blocking is enabled",
				Optional:    true,
				Default:     false,
			},
			"block_server_dhcp": {
				Type:        schema.TypeBool,
				Description: "Indicates whether DHCP server blocking is enabled",
				Optional:    true,
				Default:     false,
			},
			"bpdu_filter_enabled": {
				Type:        schema.TypeBool,
				Description: "Indicates whether BPDU filter is enabled",
				Optional:    true,
				Default:     false,
			},
			"bpdu_filter_whitelist": {
				Type:        schema.TypeSet,
				Description: "Set of allowed MAC addresses to be excluded from BPDU filtering",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(switchSecurityProfileBpduMacs, false),
				},
			},
			"rate_limits": getSwitchSecurityRateLimitsSchema(),
		},
	}
}

func getSwitchSecurityRateLimitsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Whether rate limiting is enabled",
					Optional:    true,
					Default:     true,
				},
				"rx_broadcast": {
					Type:        schema.TypeInt,
					Description: "Incoming broadcast traffic limit in packets per second",
					Optional:    true,
				},
				"rx_multicast": {
					Type:        schema.TypeInt,
					Description: "Incoming multicast traffic limit in packets per second",
					Optional:    true,
				},
				"tx_broadcast": {
					Type:        schema.TypeInt,
					Description: "Outgoing broadcast traffic limit in packets per second",
					Optional:    true,
				},
				"tx_multicast": {
					Type:        schema.TypeInt,
					Description: "Outgoing multicast traffic limit in packets per second",
					Optional:    true,
				},
			},
		},
	}
}

func getSwitchSecurityRateLimitsFromSchema(d *schema.ResourceData) *manager.RateLimits {
	rateLimitConfs := d.Get("rate_limits").([]interface{})
	for _, rateLimitConf := range rateLimitConfs {
		// only 1 is allowed
		data := rateLimitConf.(map[string]interface{})

		limits := manager.RateLimits{
			Enabled:     data["enabled"].(bool),
			RxBroadcast: int32(data["rx_broadcast"].(int)),
			RxMulticast: int32(data["rx_multicast"].(int)),
			TxBroadcast: int32(data["tx_broadcast"].(int)),
			TxMulticast: int32(data["tx_multicast"].(int)),
		}

		return &limits

	}

	return nil
}

func setSwitchSecurityRateLimitsInSchema(d *schema.ResourceData, rateLimitConf *manager.RateLimits) {
	var limits []map[string]interface{}
	if rateLimitConf == nil {
		return
	}

	elem := make(map[string]interface{})
	elem["enabled"] = rateLimitConf.Enabled
	elem["rx_broadcast"] = rateLimitConf.RxBroadcast
	elem["rx_multicast"] = rateLimitConf.RxMulticast
	elem["tx_broadcast"] = rateLimitConf.TxBroadcast
	elem["tx_multicast"] = rateLimitConf.TxMulticast

	limits = append(limits, elem)
	err := d.Set("rate_limits", limits)
	if err != nil {
		log.Printf("[WARNING] Failed to set rate limits in schema: %v", err)
	}
}

func resourceNsxtSwitchSecuritySwitchingProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	blockNonIP := d.Get("block_non_ip").(bool)
	blockClientDHCP := d.Get("block_client_dhcp").(bool)
	blockServerDHCP := d.Get("block_server_dhcp").(bool)
	bpduFilterEnabled := d.Get("bpdu_filter_enabled").(bool)
	bpduFilterWhitelist := getStringListFromSchemaSet(d, "bpdu_filter_whitelist")
	rateLimits := getSwitchSecurityRateLimitsFromSchema(d)

	switchSecurityProfile := manager.SwitchSecuritySwitchingProfile{
		Description:       description,
		DisplayName:       displayName,
		Tags:              tags,
		BlockNonIpTraffic: blockNonIP,
		DhcpFilter: &manager.DhcpFilter{
			ClientBlockEnabled: blockClientDHCP,
			ServerBlockEnabled: blockServerDHCP,
		},
		BpduFilter: &manager.BpduFilter{
			Enabled:   bpduFilterEnabled,
			WhiteList: bpduFilterWhitelist,
		},
		RateLimits: rateLimits,
	}

	switchSecurityProfile, resp, err := nsxClient.LogicalSwitchingApi.CreateSwitchSecuritySwitchingProfile(nsxClient.Context, switchSecurityProfile)

	if err != nil {
		return fmt.Errorf("Error during SwitchSecurityProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during SwitchSecurityProfile create: %v", resp.StatusCode)
	}
	d.SetId(switchSecurityProfile.Id)

	return resourceNsxtSwitchSecuritySwitchingProfileRead(d, m)
}

func resourceNsxtSwitchSecuritySwitchingProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	switchSecurityProfile, resp, err := nsxClient.LogicalSwitchingApi.GetSwitchSecuritySwitchingProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] SwitchSecurityProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during SwitchSecurityProfile read: %v", err)
	}

	d.Set("revision", switchSecurityProfile.Revision)
	d.Set("description", switchSecurityProfile.Description)
	d.Set("display_name", switchSecurityProfile.DisplayName)
	setTagsInSchema(d, switchSecurityProfile.Tags)
	d.Set("block_non_ip", switchSecurityProfile.BlockNonIpTraffic)
	if switchSecurityProfile.DhcpFilter != nil {
		d.Set("block_client_dhcp", switchSecurityProfile.DhcpFilter.ClientBlockEnabled)
		d.Set("block_server_dhcp", switchSecurityProfile.DhcpFilter.ServerBlockEnabled)
	}
	if switchSecurityProfile.BpduFilter != nil {
		d.Set("bpdu_filter_enabled", switchSecurityProfile.BpduFilter.Enabled)
		d.Set("bpdu_filter_whitelist", switchSecurityProfile.BpduFilter.WhiteList)
	}
	setSwitchSecurityRateLimitsInSchema(d, switchSecurityProfile.RateLimits)

	return nil
}

func resourceNsxtSwitchSecuritySwitchingProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	blockNonIP := d.Get("block_non_ip").(bool)
	blockClientDHCP := d.Get("block_client_dhcp").(bool)
	blockServerDHCP := d.Get("block_server_dhcp").(bool)
	bpduFilterEnabled := d.Get("bpdu_filter_enabled").(bool)
	bpduFilterWhitelist := getStringListFromSchemaSet(d, "bpdu_filter_whitelist")
	rateLimits := getSwitchSecurityRateLimitsFromSchema(d)

	switchSecurityProfile := manager.SwitchSecuritySwitchingProfile{
		Description:       description,
		DisplayName:       displayName,
		Tags:              tags,
		BlockNonIpTraffic: blockNonIP,
		DhcpFilter: &manager.DhcpFilter{
			ClientBlockEnabled: blockClientDHCP,
			ServerBlockEnabled: blockServerDHCP,
		},
		BpduFilter: &manager.BpduFilter{
			Enabled:   bpduFilterEnabled,
			WhiteList: bpduFilterWhitelist,
		},
		RateLimits: rateLimits,
		Revision:   revision,
	}

	_, resp, err := nsxClient.LogicalSwitchingApi.UpdateSwitchSecuritySwitchingProfile(nsxClient.Context, id, switchSecurityProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during SwitchSecurityProfile update: %v", err)
	}

	return resourceNsxtSwitchSecuritySwitchingProfileRead(d, m)
}

func resourceNsxtSwitchSecuritySwitchingProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.LogicalSwitchingApi.DeleteSwitchingProfile(nsxClient.Context, id, nil)
	if err != nil {
		return fmt.Errorf("Error during SwitchSecurityProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] SwitchSecurityProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
