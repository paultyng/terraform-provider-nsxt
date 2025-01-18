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

func resourceNsxtMacManagementSwitchingProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtMacManagementSwitchingProfileCreate,
		Read:   resourceNsxtMacManagementSwitchingProfileRead,
		Update: resourceNsxtMacManagementSwitchingProfileUpdate,
		Delete: resourceNsxtMacManagementSwitchingProfileDelete,
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
			"mac_change_allowed": {
				Type:        schema.TypeBool,
				Description: "Allowing source MAC address change",
				Optional:    true,
				Default:     false,
			},
			"mac_learning": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Allowing source MAC address learning",
							Optional:    true,
							Default:     false,
						},
						"unicast_flooding_allowed": {
							Type:        schema.TypeBool,
							Description: "Allowing flooding for unlearned MAC for ingress traffic",
							Optional:    true,
							Default:     false,
						},
						"limit": {
							Type:         schema.TypeInt,
							Description:  "The maximum number of MAC addresses that can be learned on this port",
							Optional:     true,
							Default:      4096,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"limit_policy": {
							Type:         schema.TypeString,
							Description:  "The policy after MAC Limit is exceeded",
							Optional:     true,
							Default:      "ALLOW",
							ValidateFunc: validation.StringInSlice([]string{"ALLOW", "DROP"}, false),
						},
					},
				},
			},
		},
	}
}

func getMacLearningFromSchema(d *schema.ResourceData) *manager.MacLearningSpec {
	macLearningConfs := d.Get("mac_learning").([]interface{})
	for _, macLearningConf := range macLearningConfs {
		// only 1 is allowed
		data := macLearningConf.(map[string]interface{})
		return &manager.MacLearningSpec{
			Enabled:                data["enabled"].(bool),
			UnicastFloodingAllowed: data["unicast_flooding_allowed"].(bool),
			Limit:                  int32(data["limit"].(int)),
			LimitPolicy:            data["limit_policy"].(string),
		}
	}

	return nil
}

func setMacLearningInSchema(d *schema.ResourceData, macLearning *manager.MacLearningSpec) error {
	var resultList []map[string]interface{}
	if macLearning != nil {
		elem := make(map[string]interface{})
		elem["enabled"] = macLearning.Enabled
		elem["unicast_flooding_allowed"] = macLearning.UnicastFloodingAllowed
		elem["limit"] = macLearning.Limit
		elem["limit_policy"] = macLearning.LimitPolicy
		resultList = append(resultList, elem)
	}
	err := d.Set("mac_learning", resultList)
	return err
}

func resourceNsxtMacManagementSwitchingProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	macChangeAllowed := d.Get("mac_change_allowed").(bool)
	macLearning := getMacLearningFromSchema(d)

	switchingProfile := manager.MacManagementSwitchingProfile{
		Description:      description,
		DisplayName:      displayName,
		Tags:             tags,
		MacChangeAllowed: macChangeAllowed,
		MacLearning:      macLearning,
	}

	switchingProfile, resp, err := nsxClient.LogicalSwitchingApi.CreateMacManagementSwitchingProfile(nsxClient.Context, switchingProfile)

	if err != nil {
		return fmt.Errorf("Error during MacManagementSwitchingProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during MacManagementSwitchingProfile create: %v", resp.StatusCode)
	}
	d.SetId(switchingProfile.Id)

	return resourceNsxtMacManagementSwitchingProfileRead(d, m)
}

func resourceNsxtMacManagementSwitchingProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	switchingProfile, resp, err := nsxClient.LogicalSwitchingApi.GetMacManagementSwitchingProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] MacManagementSwitchingProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during MacManagementSwitchingProfile read: %v", err)
	}

	d.Set("revision", switchingProfile.Revision)
	d.Set("description", switchingProfile.Description)
	d.Set("display_name", switchingProfile.DisplayName)
	d.Set("mac_change_allowed", switchingProfile.MacChangeAllowed)
	setTagsInSchema(d, switchingProfile.Tags)
	err = setMacLearningInSchema(d, switchingProfile.MacLearning)
	if err != nil {
		return fmt.Errorf("Error during setting MacManagementSwitchingProfile MacLearning: %v", err)
	}

	return nil
}

func resourceNsxtMacManagementSwitchingProfileUpdate(d *schema.ResourceData, m interface{}) error {
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
	macChangeAllowed := d.Get("mac_change_allowed").(bool)
	macLearning := getMacLearningFromSchema(d)

	switchingProfile := manager.MacManagementSwitchingProfile{
		Description:      description,
		DisplayName:      displayName,
		Tags:             tags,
		MacChangeAllowed: macChangeAllowed,
		MacLearning:      macLearning,
		Revision:         revision,
	}

	_, resp, err := nsxClient.LogicalSwitchingApi.UpdateMacManagementSwitchingProfile(nsxClient.Context, id, switchingProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during MacManagementSwitchingProfile update: %v", err)
	}

	return resourceNsxtMacManagementSwitchingProfileRead(d, m)
}

func resourceNsxtMacManagementSwitchingProfileDelete(d *schema.ResourceData, m interface{}) error {
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
		return fmt.Errorf("Error during MacManagementSwitchingProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] MacManagementSwitchingProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
