// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var lacpLoadBalanceAlgorithms = []string{
	model.Lag_LOAD_BALANCE_ALGORITHM_SRCMAC,
	model.Lag_LOAD_BALANCE_ALGORITHM_DESTMAC,
	model.Lag_LOAD_BALANCE_ALGORITHM_SRCDESTMAC,
	model.Lag_LOAD_BALANCE_ALGORITHM_SRCDESTIPVLAN,
	model.Lag_LOAD_BALANCE_ALGORITHM_SRCDESTMACIPPORT,
}

var lacpGroupModes = []string{
	model.Lag_MODE_ACTIVE,
	model.Lag_MODE_PASSIVE,
}

var lacpTimeoutTypes = []string{
	model.Lag_TIMEOUT_TYPE_SLOW,
	model.Lag_TIMEOUT_TYPE_FAST,
}

var uplinkTypes = []string{
	model.Uplink_UPLINK_TYPE_PNIC,
	model.Uplink_UPLINK_TYPE_LAG,
}

var teamingPolicies = []string{
	model.TeamingPolicy_POLICY_FAILOVER_ORDER,
	model.TeamingPolicy_POLICY_LOADBALANCE_SRCID,
	model.TeamingPolicy_POLICY_LOADBALANCE_SRC_MAC,
}

var overlayEncapModes = []string{
	model.PolicyUplinkHostSwitchProfile_OVERLAY_ENCAP_VXLAN,
	model.PolicyUplinkHostSwitchProfile_OVERLAY_ENCAP_GENEVE,
}

func resourceNsxtUplinkHostSwitchProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtUplinkHostSwitchProfileCreate,
		Read:   resourceNsxtUplinkHostSwitchProfileRead,
		Update: resourceNsxtUplinkHostSwitchProfileUpdate,
		Delete: resourceNsxtUplinkHostSwitchProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"lag": {
				Type:        schema.TypeList,
				Description: "List of LACP group",
				Optional:    true,
				MaxItems:    64,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "unique id",
							Required:    false,
							Optional:    false,
							Computed:    true,
						},
						"load_balance_algorithm": {
							Type:         schema.TypeString,
							Description:  "LACP load balance Algorithm",
							Required:     true,
							ValidateFunc: validation.StringInSlice(lacpLoadBalanceAlgorithms, false),
						},
						"mode": {
							Type:         schema.TypeString,
							Description:  "LACP group mode",
							Required:     true,
							ValidateFunc: validation.StringInSlice(lacpGroupModes, false),
						},
						"name": getRequiredStringSchema("Lag name"),
						"number_of_uplinks": {
							Type:         schema.TypeInt,
							Description:  "Number of uplinks",
							Required:     true,
							ValidateFunc: validation.IntBetween(2, 32),
						},
						"timeout_type": {
							Type:         schema.TypeString,
							Description:  "LACP timeout type",
							Optional:     true,
							Default:      model.Lag_TIMEOUT_TYPE_SLOW,
							ValidateFunc: validation.StringInSlice(lacpTimeoutTypes, false),
						},
						"uplink": {
							Type:        schema.TypeList,
							Description: "uplink names",
							Required:    false,
							Optional:    false,
							Computed:    true,
							Elem:        getUplinkSchemaResource(),
						},
					},
				},
			},
			"teaming": {
				Type:        schema.TypeList,
				Description: "Default TeamingPolicy associated with this UplinkProfile",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Type:        schema.TypeList,
							Description: "List of Uplinks used in active list",
							Required:    true,
							Elem:        getUplinkSchemaResource(),
						},
						"policy": {
							Type:         schema.TypeString,
							Description:  "Teaming policy",
							Required:     true,
							ValidateFunc: validation.StringInSlice(teamingPolicies, false),
						},
						"standby": {
							Type:        schema.TypeList,
							Description: "List of Uplinks used in standby list",
							Optional:    true,
							Elem:        getUplinkSchemaResource(),
						},
					},
				},
			},
			"named_teaming": {
				Type:        schema.TypeList,
				Description: "List of named uplink teaming policies that can be used by logical switches",
				Optional:    true,
				MaxItems:    32,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": getRequiredStringSchema("The name of the uplink teaming policy"),
						"active": {
							Type:        schema.TypeList,
							Description: "List of Uplinks used in active list",
							Required:    true,
							Elem:        getUplinkSchemaResource(),
						},
						"policy": {
							Type:         schema.TypeString,
							Description:  "Teaming policy",
							Required:     true,
							ValidateFunc: validation.StringInSlice(teamingPolicies, false),
						},
						"standby": {
							Type:        schema.TypeList,
							Description: "List of Uplinks used in standby list",
							Optional:    true,
							Elem:        getUplinkSchemaResource(),
						},
					},
				},
			},
			"overlay_encap": {
				Type:         schema.TypeString,
				Description:  "The protocol used to encapsulate overlay traffic",
				Optional:     true,
				Default:      model.PolicyUplinkHostSwitchProfile_OVERLAY_ENCAP_GENEVE,
				ValidateFunc: validation.StringInSlice(overlayEncapModes, false),
			},
			"transport_vlan": {
				Type:         schema.TypeInt,
				Description:  "VLAN used for tagging Overlay traffic of associated HostSwitch",
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 4094),
			},
			"mtu": {
				Type:         schema.TypeInt,
				Description:  "Maximum Transmission Unit used for uplinks",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1280),
			},
			"realized_id": {
				Type:        schema.TypeString,
				Description: "Computed ID of the realized object",
				Computed:    true,
			},
		},
	}
}

func getUplinkSchemaResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"uplink_name": getRequiredStringSchema("Name of this uplink"),
			"uplink_type": {
				Type:         schema.TypeString,
				Description:  "Type of the uplink",
				Required:     true,
				ValidateFunc: validation.StringInSlice(uplinkTypes, false),
			},
		},
	}
}

func setUplinkHostSwitchProfileSchema(d *schema.ResourceData, profile model.PolicyUplinkHostSwitchProfile) {
	uplinksToMap := func(uplinks []model.Uplink) []map[string]interface{} {
		var uplinksMap []map[string]interface{}
		for _, uplink := range uplinks {
			uplinkMap := make(map[string]interface{})
			uplinkMap["uplink_name"] = uplink.UplinkName
			uplinkMap["uplink_type"] = uplink.UplinkType
			uplinksMap = append(uplinksMap, uplinkMap)
		}
		return uplinksMap
	}

	var lagsSchema []map[string]interface{}
	for _, lag := range profile.Lags {
		lagMap := make(map[string]interface{})
		lagMap["id"] = lag.Id
		lagMap["load_balance_algorithm"] = lag.LoadBalanceAlgorithm
		lagMap["mode"] = lag.Mode
		lagMap["name"] = lag.Name
		lagMap["number_of_uplinks"] = lag.NumberOfUplinks
		lagMap["timeout_type"] = lag.TimeoutType
		lagMap["uplink"] = uplinksToMap(lag.Uplinks)
		lagsSchema = append(lagsSchema, lagMap)
	}
	d.Set("lag", lagsSchema)

	teamingSchema := make(map[string]interface{})
	teamingSchema["policy"] = profile.Teaming.Policy
	teamingSchema["active"] = uplinksToMap(profile.Teaming.ActiveList)
	teamingSchema["standby"] = uplinksToMap(profile.Teaming.StandbyList)
	d.Set("teaming", []map[string]interface{}{teamingSchema})

	var namedTeamingsSchema []map[string]interface{}
	for _, namedTeaming := range profile.NamedTeamings {
		namedTeamingMap := make(map[string]interface{})
		namedTeamingMap["name"] = namedTeaming.Name
		namedTeamingMap["policy"] = namedTeaming.Policy
		namedTeamingMap["active"] = uplinksToMap(namedTeaming.ActiveList)
		namedTeamingMap["standby"] = uplinksToMap(namedTeaming.StandbyList)
		namedTeamingsSchema = append(namedTeamingsSchema, namedTeamingMap)
	}
	d.Set("named_teaming", namedTeamingsSchema)

	d.Set("overlay_encap", profile.OverlayEncap)
	d.Set("transport_vlan", profile.TransportVlan)

	// Set MTU only when has a value
	if profile.Mtu != nil {
		d.Set("mtu", profile.Mtu)
	}
}

func resourceNsxtUplinkHostSwitchProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	client := infra.NewHostSwitchProfilesClient(connector)
	structValue, err := client.Get(id)
	if err == nil {
		converter := bindings.NewTypeConverter()
		baseInterface, errs := converter.ConvertToGolang(structValue, model.PolicyBaseHostSwitchProfileBindingType())
		if errs != nil {
			return false, errs[0]
		}
		base := baseInterface.(model.PolicyBaseHostSwitchProfile)

		resourceType := base.ResourceType
		if resourceType != infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE {
			return false, nil
		}
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtUplinkHostSwitchProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining UplinkHostSwitchProfile ID")
	}
	client := infra.NewHostSwitchProfilesClient(connector)
	structValue, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "UplinkHostSwitchProfile", id, err)
	}

	converter := bindings.NewTypeConverter()
	baseInterface, errs := converter.ConvertToGolang(structValue, model.PolicyBaseHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	base := baseInterface.(model.PolicyBaseHostSwitchProfile)

	resourceType := base.ResourceType
	if resourceType != infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE {
		return handleReadError(d, "UplinkHostSwitchProfile", id, err)
	}

	uplinkInterface, errs := converter.ConvertToGolang(structValue, model.PolicyUplinkHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	uplinkHostSwitchProfile := uplinkInterface.(model.PolicyUplinkHostSwitchProfile)
	d.Set("display_name", uplinkHostSwitchProfile.DisplayName)
	d.Set("description", uplinkHostSwitchProfile.Description)
	setPolicyTagsInSchema(d, uplinkHostSwitchProfile.Tags)
	d.Set("nsx_id", id)
	d.Set("path", uplinkHostSwitchProfile.Path)
	d.Set("revision", uplinkHostSwitchProfile.Revision)
	d.Set("realized_id", uplinkHostSwitchProfile.RealizationId)
	setUplinkHostSwitchProfileSchema(d, uplinkHostSwitchProfile)
	return nil
}

func uplinkSchemaToModelList(schema interface{}) []model.Uplink {
	uplinkSchemaList := schema.([]interface{})
	var uplinkList []model.Uplink
	for _, item := range uplinkSchemaList {
		data := item.(map[string]interface{})
		uplinkName := data["uplink_name"].(string)
		uplinkType := data["uplink_type"].(string)
		obj := model.Uplink{
			UplinkName: &uplinkName,
			UplinkType: &uplinkType,
		}
		uplinkList = append(uplinkList, obj)
	}
	return uplinkList
}

func uplinkHostSwitchProfileSchemaToModel(d *schema.ResourceData) model.PolicyUplinkHostSwitchProfile {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	var lags []model.Lag
	lagsList := d.Get("lag").([]interface{})
	for _, lagSchema := range lagsList {
		lagData := lagSchema.(map[string]interface{})
		loadBalanceAlgorithm := lagData["load_balance_algorithm"].(string)
		mode := lagData["mode"].(string)
		name := lagData["name"].(string)
		numberOfUplinks := int64(lagData["number_of_uplinks"].(int))
		timeoutType := lagData["timeout_type"].(string)
		lag := model.Lag{
			LoadBalanceAlgorithm: &loadBalanceAlgorithm,
			Mode:                 &mode,
			Name:                 &name,
			NumberOfUplinks:      &numberOfUplinks,
			TimeoutType:          &timeoutType,
		}
		lags = append(lags, lag)
	}

	var namedTeamings []model.NamedTeamingPolicy
	namedTeamingsList := d.Get("named_teaming").([]interface{})
	for _, namedTeamingsSchema := range namedTeamingsList {
		namedTeamingsSchemaData := namedTeamingsSchema.(map[string]interface{})
		activeList := uplinkSchemaToModelList(namedTeamingsSchemaData["active"])
		standbyList := uplinkSchemaToModelList(namedTeamingsSchemaData["standby"])
		policy := namedTeamingsSchemaData["policy"].(string)
		name := namedTeamingsSchemaData["name"].(string)
		namedTeaming := model.NamedTeamingPolicy{
			ActiveList:  activeList,
			Policy:      &policy,
			StandbyList: standbyList,
			Name:        &name,
		}
		namedTeamings = append(namedTeamings, namedTeaming)
	}

	teamingSchema := d.Get("teaming").([]interface{})[0]
	teamingSchemaData := teamingSchema.(map[string]interface{})
	activeList := uplinkSchemaToModelList(teamingSchemaData["active"])
	standbyList := uplinkSchemaToModelList(teamingSchemaData["standby"])
	policy := teamingSchemaData["policy"].(string)
	teaming := model.TeamingPolicy{
		ActiveList:  activeList,
		Policy:      &policy,
		StandbyList: standbyList,
	}

	overlayEncap := d.Get("overlay_encap").(string)
	mtu := int64(d.Get("mtu").(int))
	transportVlan := int64(d.Get("transport_vlan").(int))

	uplinkHostSwitchProfile := model.PolicyUplinkHostSwitchProfile{
		DisplayName:   &displayName,
		Description:   &description,
		Tags:          tags,
		ResourceType:  model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE,
		OverlayEncap:  &overlayEncap,
		TransportVlan: &transportVlan,
		Lags:          lags,
		NamedTeamings: namedTeamings,
		Teaming:       &teaming,
	}
	if mtu != 0 {
		uplinkHostSwitchProfile.Mtu = &mtu
	}
	return uplinkHostSwitchProfile
}

func resourceNsxtUplinkHostSwitchProfileCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtUplinkHostSwitchProfileExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	uplinkHostSwitchProfile := uplinkHostSwitchProfileSchemaToModel(d)
	profileValue, errs := converter.ConvertToVapi(uplinkHostSwitchProfile, model.PolicyUplinkHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)
	_, err = client.Patch(id, profileStruct)
	if err != nil {
		return handleCreateError("UplinkHostSwitchProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtUplinkHostSwitchProfileRead(d, m)
}

func resourceNsxtUplinkHostSwitchProfileUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining UplinkHostSwitchProfile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	uplinkHostSwitchProfile := uplinkHostSwitchProfileSchemaToModel(d)
	revision := int64(d.Get("revision").(int))
	uplinkHostSwitchProfile.Revision = &revision
	profileValue, errs := converter.ConvertToVapi(uplinkHostSwitchProfile, model.PolicyUplinkHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)
	_, err := client.Update(id, profileStruct)
	if err != nil {
		return handleUpdateError("UplinkHostSwitchProfile", id, err)
	}

	return resourceNsxtUplinkHostSwitchProfileRead(d, m)
}

func resourceNsxtUplinkHostSwitchProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining UplinkHostSwitchProfile ID")
	}
	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("UplinkHostSwitchProfile", id, err)
	}
	return nil
}
