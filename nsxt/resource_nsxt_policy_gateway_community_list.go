// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyGatewayCommunityList() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayCommunityListCreate,
		Read:   resourceNsxtPolicyGatewayCommunityListRead,
		Update: resourceNsxtPolicyGatewayCommunityListUpdate,
		Delete: resourceNsxtPolicyGatewayCommunityListDelete,
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
			"communities": {
				Type:        schema.TypeSet,
				Description: "List of BGP community entries",
				Required:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePolicyBGPCommunity,
				},
			},
		},
	}
}

func resourceNsxtPolicyGatewayCommunityListExists(tier0Id string, id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	if isGlobalManager {
		client := gm_tier0s.NewCommunityListsClient(connector)
		_, err = client.Get(tier0Id, id)
	} else {
		client := tier_0s.NewCommunityListsClient(connector)
		_, err = client.Get(tier0Id, id)
	}
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyGatewayCommunityListCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id := d.Get("nsx_id").(string)
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	if id == "" {
		id = newUUID()
	} else {
		var err error
		if isPolicyGlobalManager(m) {
			client := gm_tier0s.NewCommunityListsClient(connector)
			_, err = client.Get(gwID, id)
		} else {
			client := tier_0s.NewCommunityListsClient(connector)
			_, err = client.Get(gwID, id)
		}
		if err == nil {
			return fmt.Errorf("Community List with ID '%s' already exists on Tier0 Gateway %s", id, gwID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	communities := getStringListFromSchemaSet(d, "communities")
	tags := getPolicyTagsFromSchema(d)

	obj := model.CommunityList{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Communities: communities,
	}

	var err error
	// Create the resource using PATCH
	log.Printf("[INFO] Creating Gateway Community List with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.CommunityListBindingType(), gm_model.CommunityListBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_tier0s.NewCommunityListsClient(connector)
		err = client.Patch(gwID, id, gmObj.(gm_model.CommunityList))
	} else {
		client := tier_0s.NewCommunityListsClient(connector)
		err = client.Patch(gwID, id, obj)
	}
	if err != nil {
		return handleCreateError("Community List", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayCommunityListRead(d, m)
}

func resourceNsxtPolicyGatewayCommunityListRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Community List ID")
	}
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	var obj model.CommunityList
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewCommunityListsClient(connector)
		gmObj, err := client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Community List", id, err)
		}

		lmObj, err := convertModelBindingType(gmObj, gm_model.CommunityListBindingType(), model.CommunityListBindingType())
		if err != nil {
			return err
		}
		obj = lmObj.(model.CommunityList)
	} else {
		client := tier_0s.NewCommunityListsClient(connector)
		var err error
		obj, err = client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Community List", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("communities", obj.Communities)

	return nil
}

func resourceNsxtPolicyGatewayCommunityListUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayCommunityList ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	communities := getStringListFromSchemaSet(d, "communities")
	revision := int64(d.Get("revision").(int))

	obj := model.CommunityList{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Communities: communities,
		Revision:    &revision,
	}

	var err error
	log.Printf("[INFO] Updating Gateway Community List with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.CommunityListBindingType(), gm_model.CommunityListBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_tier0s.NewCommunityListsClient(connector)
		_, err = client.Update(gwID, id, gmObj.(gm_model.CommunityList))
	} else {
		client := tier_0s.NewCommunityListsClient(connector)
		_, err = client.Update(gwID, id, obj)
	}
	if err != nil {
		return handleCreateError("Gateway Community List", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayCommunityListRead(d, m)
}

func resourceNsxtPolicyGatewayCommunityListDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayCommunityList ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	connector := getPolicyConnector(m)
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewCommunityListsClient(connector)
		err = client.Delete(gwID, id)
	} else {
		client := tier_0s.NewCommunityListsClient(connector)
		err = client.Delete(gwID, id)
	}

	if err != nil {
		return handleDeleteError("GatewayCommunityList", id, err)
	}

	return nil
}
