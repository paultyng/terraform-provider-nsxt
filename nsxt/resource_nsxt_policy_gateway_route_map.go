// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_tier0s "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyGatewayRouteMap() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayRouteMapCreate,
		Read:   resourceNsxtPolicyGatewayRouteMapRead,
		Update: resourceNsxtPolicyGatewayRouteMapUpdate,
		Delete: resourceNsxtPolicyGatewayRouteMapDelete,
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
			"entry": {
				Type:        schema.TypeList,
				Description: "List of Route Map entries",
				Required:    true,
				Elem:        getPolicyRouteMapEntrySchema(),
			},
		},
	}
}

var gatewayRouteMapActionValues = []string{
	model.RouteMapEntry_ACTION_PERMIT,
	model.RouteMapEntry_ACTION_DENY,
}

var gatewayRouteMapMatchCriteriaValues = []string{
	model.CommunityMatchCriteria_MATCH_OPERATOR_ANY,
	model.CommunityMatchCriteria_MATCH_OPERATOR_ALL,
	model.CommunityMatchCriteria_MATCH_OPERATOR_EXACT,
	model.CommunityMatchCriteria_MATCH_OPERATOR_COMMUNITY_REGEX,
	model.CommunityMatchCriteria_MATCH_OPERATOR_LARGE_COMMUNITY_REGEX,
}

func getPolicyRouteMapEntrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"action": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Action for the route map entry",
				Default:      model.RouteMapEntry_ACTION_PERMIT,
				ValidateFunc: validation.StringInSlice(gatewayRouteMapActionValues, false),
			},
			"community_list_match": {
				Type:        schema.TypeList,
				Description: "Prefix list match criteria for route map",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criteria": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Community list path or a regular expression",
						},
						"match_operator": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Match operator for community list entries",
							ValidateFunc: validation.StringInSlice(gatewayRouteMapMatchCriteriaValues, false),
						},
					},
				},
			},
			"prefix_list_matches": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of paths for prefix lists for route map",
				Elem:        getElemPolicyPathSchema(),
			},
			"set": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Set criteria for route map entry",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"as_path_prepend": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateASPath,
							Description:  "Autonomous System (AS) path prepend to influence route selection",
						},
						"community": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Set BGP regular or large community for matching routes",
							ValidateFunc: validatePolicyBGPCommunity,
						},
						"local_preference": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Local preference indicates the degree of preference for one BGP route over other BGP routes",
						},
						"med": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "A lower Multi exit descriminator (MED) is preferred over a higher value",
						},
						"prefer_global_v6_next_hop": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Indicator whether to prefer global address over link-local as the next hop",
						},
						"weight": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Weight is used to select a route when multiple routes are available to the same network",
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyGatewayRouteMapExists(tier0Id string, id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	if isGlobalManager {
		client := gm_tier0s.NewRouteMapsClient(connector)
		_, err = client.Get(tier0Id, id)
	} else {
		client := tier_0s.NewRouteMapsClient(connector)
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

func policyGatewayRouteMapBuildEntry(d *schema.ResourceData, entryNo int, schemaEntry map[string]interface{}) model.RouteMapEntry {

	action := schemaEntry["action"].(string)
	obj := model.RouteMapEntry{
		Action: &action,
	}

	commListMatches := schemaEntry["community_list_match"].([]interface{})
	prefixListMatches := schemaEntry["prefix_list_matches"].(*schema.Set).List()

	if len(commListMatches) > 0 {
		var commCriteriaList []model.CommunityMatchCriteria
		for _, comm := range commListMatches {
			data := comm.(map[string]interface{})
			criteria := data["criteria"].(string)
			operator := data["match_operator"].(string)
			commCriteria := model.CommunityMatchCriteria{
				Criteria:      &criteria,
				MatchOperator: &operator,
			}

			commCriteriaList = append(commCriteriaList, commCriteria)
		}

		obj.CommunityListMatches = commCriteriaList
	} else if len(prefixListMatches) > 0 {
		obj.PrefixListMatches = interfaceListToStringList(prefixListMatches)
	}

	if len(schemaEntry["set"].([]interface{})) > 0 {
		data := schemaEntry["set"].([]interface{})[0].(map[string]interface{})
		asPathPrepend := data["as_path_prepend"].(string)
		community := data["community"].(string)
		localPreferenceValue, lpSet := d.GetOk(fmt.Sprintf("entry.%d.set.0.local_preference", entryNo))
		medValue, medSet := d.GetOk(fmt.Sprintf("entry.%d.set.0.med", entryNo))
		globalIPv6 := data["prefer_global_v6_next_hop"].(bool)
		weight := int64(data["weight"].(int))

		entrySet := model.RouteMapEntrySet{
			PreferGlobalV6NextHop: &globalIPv6,
			Weight:                &weight,
		}

		if len(asPathPrepend) > 0 {
			entrySet.AsPathPrepend = &asPathPrepend
		}
		if len(community) > 0 {
			entrySet.Community = &community
		}
		if lpSet {
			localPreference := int64(localPreferenceValue.(int))
			entrySet.LocalPreference = &localPreference
		}
		if medSet {
			med := int64(medValue.(int))
			entrySet.Med = &med
		}
		obj.Set = &entrySet
	}

	return obj
}

func resourceNsxtPolicyGatewayRouteMapPatch(gwID string, id string, d *schema.ResourceData, isGlobalManager bool, connector client.Connector) error {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	schemaEntries := d.Get("entry").([]interface{})
	var entries []model.RouteMapEntry
	for entryNo, entry := range schemaEntries {
		entries = append(entries, policyGatewayRouteMapBuildEntry(d, entryNo, entry.(map[string]interface{})))
	}

	obj := model.Tier0RouteMap{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Entries:     entries,
	}

	if isGlobalManager {
		gmObj, convErr := convertModelBindingType(obj, model.Tier0RouteMapBindingType(), gm_model.Tier0RouteMapBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_tier0s.NewRouteMapsClient(connector)
		return client.Patch(gwID, id, gmObj.(gm_model.Tier0RouteMap))
	}
	client := tier_0s.NewRouteMapsClient(connector)
	return client.Patch(gwID, id, obj)
}

func resourceNsxtPolicyGatewayRouteMapCreate(d *schema.ResourceData, m interface{}) error {
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
			client := gm_tier0s.NewRouteMapsClient(connector)
			_, err = client.Get(gwID, id)
		} else {
			client := tier_0s.NewRouteMapsClient(connector)
			_, err = client.Get(gwID, id)
		}
		if err == nil {
			return fmt.Errorf("Route Map with ID '%s' already exists on Tier0 Gateway %s", id, gwID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	log.Printf("[INFO] Creating Gateway Route Map with ID %s", id)
	err := resourceNsxtPolicyGatewayRouteMapPatch(gwID, id, d, isPolicyGlobalManager(m), connector)
	if err != nil {
		return handleCreateError("Route Map", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayRouteMapRead(d, m)
}

func resourceNsxtPolicyGatewayRouteMapRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Route Map ID")
	}
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	var obj model.Tier0RouteMap
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewRouteMapsClient(connector)
		gmObj, err := client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Route Map", id, err)
		}

		lmObj, err := convertModelBindingType(gmObj, gm_model.Tier0RouteMapBindingType(), model.Tier0RouteMapBindingType())
		if err != nil {
			return err
		}
		obj = lmObj.(model.Tier0RouteMap)
	} else {
		client := tier_0s.NewRouteMapsClient(connector)
		var err error
		obj, err = client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway Route Map", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	var entryList []interface{}
	for _, entry := range obj.Entries {
		schemaEntry := make(map[string]interface{})
		schemaEntry["action"] = entry.Action

		var commList []map[string]interface{}
		if len(entry.CommunityListMatches) > 0 {
			for _, comm := range entry.CommunityListMatches {
				data := make(map[string]interface{})
				data["criteria"] = comm.Criteria
				data["match_operator"] = comm.MatchOperator

				commList = append(commList, data)
			}
		}

		schemaEntry["community_list_match"] = commList
		schemaEntry["prefix_list_matches"] = entry.PrefixListMatches

		var setList []map[string]interface{}
		if entry.Set != nil {
			setData := make(map[string]interface{})
			setData["as_path_prepend"] = entry.Set.AsPathPrepend
			setData["community"] = entry.Set.Community
			setData["med"] = entry.Set.Med
			setData["local_preference"] = entry.Set.LocalPreference
			setData["prefer_global_v6_next_hop"] = entry.Set.PreferGlobalV6NextHop
			setData["weight"] = entry.Set.Weight

			setList = append(setList, setData)
		}

		schemaEntry["set"] = setList

		entryList = append(entryList, schemaEntry)
	}

	d.Set("entry", entryList)

	return nil
}

func resourceNsxtPolicyGatewayRouteMapUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayRouteMap ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	log.Printf("[INFO] Updating Gateway Route Map with ID %s", id)
	err := resourceNsxtPolicyGatewayRouteMapPatch(gwID, id, d, isPolicyGlobalManager(m), connector)
	if err != nil {
		return handleCreateError("Gateway Route Map", id, err)
	}

	return resourceNsxtPolicyGatewayRouteMapRead(d, m)
}

func resourceNsxtPolicyGatewayRouteMapDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway Route Map ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	connector := getPolicyConnector(m)
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_tier0s.NewRouteMapsClient(connector)
		err = client.Delete(gwID, id)
	} else {
		client := tier_0s.NewRouteMapsClient(connector)
		err = client.Delete(gwID, id)
	}

	if err != nil {
		return handleDeleteError("Gateway Route Map", id, err)
	}

	return nil
}
