// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_tier_0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var gatewayPrefixListActionTypeValues = []string{
	model.PrefixEntry_ACTION_PERMIT,
	model.PrefixEntry_ACTION_DENY,
}

func resourceNsxtPolicyGatewayPrefixList() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayPrefixListCreate,
		Read:   resourceNsxtPolicyGatewayPrefixListRead,
		Update: resourceNsxtPolicyGatewayPrefixListUpdate,
		Delete: resourceNsxtPolicyGatewayPrefixListDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0GatewayImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"gateway_path": getPolicyPathSchema(true, true, "Policy path for Tier0 gateway"),
			"prefix": {
				Type:        schema.TypeList,
				Description: "Ordered list of network prefixes",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Action for the prefix list",
							ValidateFunc: validation.StringInSlice(gatewayPrefixListActionTypeValues, false),
							Default:      model.PrefixEntry_ACTION_PERMIT,
						},
						"ge": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Prefix length greater than or equal to",
							ValidateFunc: validation.IntBetween(0, 128),
							Default:      0,
						},
						"le": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Prefix length less than or equal to",
							ValidateFunc: validation.IntBetween(0, 128),
							Default:      0,
						},
						"network": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Network prefix in CIDR format. If not set it will match ANY network",
							ValidateFunc: validateCidr(),
						},
					},
				},
			},
		},
	}
}

func setPrefixesInSchema(d *schema.ResourceData, prefixes []model.PrefixEntry) {
	var entriesList []map[string]interface{}
	for _, prefix := range prefixes {
		elem := make(map[string]interface{})
		elem["action"] = *prefix.Action
		elem["ge"] = prefix.Ge
		elem["le"] = prefix.Le
		if *prefix.Network == "ANY" || *prefix.Network == "any" {
			elem["network"] = ""
		} else {
			elem["network"] = prefix.Network
		}

		entriesList = append(entriesList, elem)
	}

	err := d.Set("prefix", entriesList)
	if err != nil {
		log.Printf("[WARNING] Failed to set prefix in schema: %v", err)
	}
}

func getPrefixesFromSchema(d *schema.ResourceData) []model.PrefixEntry {
	prefixes := d.Get("prefix").([]interface{})
	var entriesList []model.PrefixEntry
	for _, prefix := range prefixes {
		data := prefix.(map[string]interface{})
		action := data["action"].(string)
		network := data["network"].(string)

		if network == "" {
			network = "ANY"
		}

		elem := model.PrefixEntry{
			Action:  &action,
			Network: &network,
		}

		ge := int64(data["ge"].(int))
		if ge > 0 {
			elem.Ge = &ge
		}
		le := int64(data["le"].(int))
		if le > 0 {
			elem.Le = &le
		}

		entriesList = append(entriesList, elem)
	}

	return entriesList
}

func resourceNsxtPolicyGatewayPrefixListDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	isGlobalManager := isPolicyGlobalManager(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Prefix List ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	var err error
	if isGlobalManager {
		client := gm_tier_0s.NewPrefixListsClient(connector)
		err = client.Delete(gwID, id)
	} else {
		client := tier_0s.NewPrefixListsClient(connector)
		err = client.Delete(gwID, id)
	}
	if err != nil {
		return handleDeleteError("Gateway Prefix List", id, err)
	}

	return nil
}

func resourceNsxtPolicyGatewayPrefixListRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	isGlobalManager := isPolicyGlobalManager(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Prefix List ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	var obj model.PrefixList
	var err error
	if isGlobalManager {
		var gmObj gm_model.PrefixList
		var rawObj interface{}
		client := gm_tier_0s.NewPrefixListsClient(connector)
		gmObj, err = client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Prefix List", id, err)
		}
		rawObj, err = convertModelBindingType(gmObj, gm_model.PrefixListBindingType(), model.PrefixListBindingType())
		if err != nil {
			return handleReadError(d, "Gateway Prefix List", id, err)
		}
		obj = rawObj.(model.PrefixList)
	} else {
		client := tier_0s.NewPrefixListsClient(connector)
		obj, err = client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Prefix List", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPrefixesInSchema(d, obj.Prefixes)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.SetId(id)

	return nil
}

func patchNsxtPolicyGatewayPrefixList(connector client.Connector, gwID string, prefixList model.PrefixList, isGlobalManager bool) error {
	if isGlobalManager {
		rawObj, err := convertModelBindingType(prefixList, model.PrefixListBindingType(), gm_model.PrefixListBindingType())
		if err != nil {
			return err
		}
		client := gm_tier_0s.NewPrefixListsClient(connector)
		return client.Patch(gwID, *prefixList.Id, rawObj.(gm_model.PrefixList))
	}
	client := tier_0s.NewPrefixListsClient(connector)
	return client.Patch(gwID, *prefixList.Id, prefixList)
}

func resourceNsxtPolicyGatewayPrefixListCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	isGlobalManager := isPolicyGlobalManager(m)

	// Initialize resource Id and verify this ID is not yet used
	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		var err error
		if isGlobalManager {
			client := gm_tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, id)
		} else {
			client := tier_0s.NewPrefixListsClient(connector)
			_, err = client.Get(gwID, id)
		}
		if err == nil {
			return fmt.Errorf("Gateway Prefix List with nsx_id '%s' already exists", id)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	prefixes := getPrefixesFromSchema(d)
	tags := getPolicyTagsFromSchema(d)

	prefixListStruct := model.PrefixList{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Prefixes:    prefixes,
	}

	log.Printf("[INFO] Creating Gateway Prefix List with ID %s", id)

	err := patchNsxtPolicyGatewayPrefixList(connector, gwID, prefixListStruct, isGlobalManager)
	if err != nil {
		return handleCreateError("Gateway Prefix List", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayPrefixListRead(d, m)
}

func resourceNsxtPolicyGatewayPrefixListUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Prefix List ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	prefixes := getPrefixesFromSchema(d)
	tags := getPolicyTagsFromSchema(d)

	prefixListStruct := model.PrefixList{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Prefixes:    prefixes,
	}

	log.Printf("[INFO] Updating Gateway Prefix List with ID %s", id)
	err := patchNsxtPolicyGatewayPrefixList(connector, gwID, prefixListStruct, isPolicyGlobalManager(m))
	if err != nil {
		return handleUpdateError("Gateway Prefix List", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayPrefixListRead(d, m)
}

func resourceNsxtPolicyTier0GatewayImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <gateway-id>/<id> as an input")
	}

	gwID := s[0]
	connector := getPolicyConnector(m)
	if isPolicyGlobalManager(m) {
		t0Client := gm_infra.NewTier0sClient(connector)
		t0gw, err := t0Client.Get(gwID)
		if err != nil {
			return nil, err
		}
		d.Set("gateway_path", t0gw.Path)
	} else {
		t0Client := infra.NewTier0sClient(connector)
		t0gw, err := t0Client.Get(gwID)
		if err != nil {
			return nil, err
		}
		d.Set("gateway_path", t0gw.Path)
	}
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
