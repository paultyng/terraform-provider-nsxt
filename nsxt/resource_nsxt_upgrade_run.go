/* Copyright © 2024 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade/plan"
)

// Order matters
var upgradeComponentList = []string{
	edgeUpgradeGroup,
	hostUpgradeGroup,
	mpUpgradeGroup,
}

var componentToGroupKey = map[string]string{
	edgeUpgradeGroup: "edge_group",
	hostUpgradeGroup: "host_group",
}

var componentToSettingKey = map[string]string{
	edgeUpgradeGroup: "edge_upgrade_setting",
	hostUpgradeGroup: "host_upgrade_setting",
}

var supportedUpgradeMode = []string{"maintenance_mode", "in_place", "stage_in_vlcm"}
var supportedMaintenanceModeConfigVsanMode = []string{"evacuate_all_data", "ensure_object_accessibility", "no_action"}

var (
	// Default waiting setup in seconds
	defaultUpgradeStatusCheckInterval = 30
	defaultUpgradeStatusCheckTimeout  = 3600
	defaultUpgradeStatusCheckDelay    = 30
)

var staticComponentUpgradeStatus = []string{
	model.ComponentUpgradeStatus_STATUS_FAILED,
	model.ComponentUpgradeStatus_STATUS_NOT_STARTED,
	model.ComponentUpgradeStatus_STATUS_PAUSED,
}

var inFlightComponentUpgradeStatus = []string{
	model.ComponentUpgradeStatus_STATUS_IN_PROGRESS,
	model.ComponentUpgradeStatus_STATUS_PAUSING,
}

type upgradeClientSet struct {
	GroupClient       upgrade.UpgradeUnitGroupsClient
	SettingClient     plan.SettingsClient
	PlanClient        upgrade.PlanClient
	StatusClient      upgrade.StatusSummaryClient
	UpgradeClient     nsx.UpgradeClient
	GroupStatusClient upgrade.UpgradeUnitGroupsStatusClient

	Timeout  int
	Delay    int
	Interval int
}

func newUpgradeClientSet(connector client.Connector, d *schema.ResourceData) *upgradeClientSet {
	return &upgradeClientSet{
		GroupClient:       upgrade.NewUpgradeUnitGroupsClient(connector),
		SettingClient:     plan.NewSettingsClient(connector),
		PlanClient:        upgrade.NewPlanClient(connector),
		StatusClient:      upgrade.NewStatusSummaryClient(connector),
		UpgradeClient:     nsx.NewUpgradeClient(connector),
		GroupStatusClient: upgrade.NewUpgradeUnitGroupsStatusClient(connector),

		Timeout:  d.Get("timeout").(int),
		Delay:    d.Get("delay").(int),
		Interval: d.Get("interval").(int),
	}
}

func resourceNsxtUpgradeRun() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtUpgradeRunCreate,
		Read:   resourceNsxtUpgradeRunRead,
		Update: resourceNsxtUpgradeRunUpdate,
		Delete: resourceNsxtUpgradeRunDelete,

		Schema: map[string]*schema.Schema{
			"upgrade_prepare_ready_id": {
				Type:        schema.TypeString,
				Description: "ID of corresponding nsxt_upgrade_prepare_ready resource",
				Required:    true,
				ForceNew:    true,
			},
			"edge_group":           getUpgradeGroupSchema(false),
			"host_group":           getUpgradeGroupSchema(true),
			"edge_upgrade_setting": getUpgradeSettingSchema(true),
			"host_upgrade_setting": getUpgradeSettingSchema(false),
			"timeout": {
				Type:         schema.TypeInt,
				Description:  "Upgrade status check timeout in seconds",
				Optional:     true,
				Default:      defaultUpgradeStatusCheckTimeout,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"interval": {
				Type:         schema.TypeInt,
				Description:  "Interval to check upgrade status in seconds",
				Optional:     true,
				Default:      defaultUpgradeStatusCheckInterval,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"delay": {
				Type:         schema.TypeInt,
				Description:  "Initial delay to start upgrade status checks in seconds",
				Optional:     true,
				Default:      defaultUpgradeStatusCheckDelay,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"upgrade_group_plan": getUpgradeGroupPlanSchema(),
			"state": {
				Type:        schema.TypeList,
				Description: "Upgrade states",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: "Component type",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Upgrade status of component",
							Computed:    true,
						},
						"target_version": {
							Type:        schema.TypeString,
							Description: "Target component version",
							Computed:    true,
						},
						"details": {
							Type:        schema.TypeString,
							Description: "Upgrade details",
							Computed:    true,
						},
						"group_state": {
							Type:        schema.TypeList,
							Description: "UpgradeGroup upgrade status",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_id": {
										Type:        schema.TypeString,
										Description: "Group ID",
										Computed:    true,
									},
									"group_name": {
										Type:        schema.TypeString,
										Description: "Group name",
										Computed:    true,
									},
									"status": {
										Type:        schema.TypeString,
										Description: "Upgrade status",
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func getUpgradeGroupPlanSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Upgrade plan for this upgrade",
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:        schema.TypeString,
					Description: "Component type",
					Computed:    true,
				},
				"id": {
					Type:        schema.TypeString,
					Description: "ID of upgrade unit group",
					Computed:    true,
				},
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Flag to indicate whether upgrade of this group is enabled or not",
					Computed:    true,
				},
				"parallel": {
					Type:        schema.TypeBool,
					Description: "Upgrade method to specify whether the upgrade is to be performed in parallel or serially",
					Computed:    true,
				},
				"pause_after_each_upgrade_unit": {
					Type:        schema.TypeBool,
					Description: "Flag to indicate whether upgrade should be paused after upgrade of each upgrade-unit",
					Computed:    true,
				},
				"extended_config": {
					Type:     schema.TypeMap,
					Optional: true,
					Computed: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func getUpgradeGroupSchema(isHostGroup bool) *schema.Schema {
	elemSchema := map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Description: "ID of upgrade unit group",
			Required:    true,
		},
		"enabled": {
			Type:        schema.TypeBool,
			Description: "Flag to indicate whether upgrade of this group is enabled or not",
			Optional:    true,
			Default:     true,
		},
		"parallel": {
			Type:        schema.TypeBool,
			Description: "Upgrade method to specify whether the upgrade is to be performed in parallel or serially",
			Optional:    true,
			Default:     true,
		},
		"pause_after_each_upgrade_unit": {
			Type:        schema.TypeBool,
			Description: "Flag to indicate whether upgrade should be paused after upgrade of each upgrade-unit",
			Optional:    true,
			Default:     false,
		},
	}

	if isHostGroup {
		elemSchema["upgrade_mode"] = &schema.Schema{
			Type:         schema.TypeString,
			Description:  "Upgrade mode",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(supportedUpgradeMode, false),
		}
		elemSchema["maintenance_mode_config_vsan_mode"] = &schema.Schema{
			Type:         schema.TypeString,
			Description:  "Maintenance mode config vsan mode",
			Optional:     true,
			ValidateFunc: validation.StringInSlice(supportedMaintenanceModeConfigVsanMode, false),
		}
		elemSchema["maintenance_mode_config_evacuate_powered_off_vms"] = &schema.Schema{
			Type:        schema.TypeBool,
			Description: "Maintenance mode config evacuate powered off vms",
			Optional:    true,
			Default:     false,
		}
		elemSchema["rebootless_upgrade"] = &schema.Schema{
			Type:        schema.TypeBool,
			Description: "Rebootless upgrade",
			Optional:    true,
			Default:     true,
		}
	}

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Upgrade group for this upgrade",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: elemSchema,
		},
	}
}

func getUpgradeSettingSchema(isEdge bool) *schema.Schema {
	elemSchema := map[string]*schema.Schema{
		"post_upgrade_check": {
			Type:        schema.TypeBool,
			Description: "Whether run post upgrade check",
			Optional:    true,
			Default:     true,
		},
		"parallel": {
			Type:        schema.TypeBool,
			Description: "Whether run upgrade parallel",
			Optional:    true,
			Default:     true,
		},
		"stop_on_error": {
			Type:        schema.TypeBool,
			Description: "Whether stop the upgrade when an error occur",
			Optional:    true,
			Default:     false,
		},
	}
	if isEdge {
		// Edge Upgrade setting is forced to stop on error.
		delete(elemSchema, "stop_on_error")
	}
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Upgrade plan setting for component",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: elemSchema,
		},
	}
}

func resourceNsxtUpgradeRunCreate(d *schema.ResourceData, m interface{}) error {
	return upgradeRunCreateOrUpdate(d, m)
}

func upgradeRunCreateOrUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		id = newUUID()
	}
	connector := getPolicyConnectorWithHeaders(m, nil, false, false)
	upgradeClientSet := newUpgradeClientSet(connector, d)

	log.Printf("[INFO] Updating UpgradeUnitGroup and UpgradePlanSetting.")
	err := prepareUpgrade(upgradeClientSet, d)
	if err != nil {
		return handleCreateError("NsxtUpgradeRun", id, err)
	}

	log.Printf("[INFO] Successfully update UpgradeUnitGroup and UpgradePlanSetting. Start Upgrade.")

	err = runUpgrade(upgradeClientSet, getPartialUpgradeMap(d))
	if err != nil {
		return handleCreateError("NsxtUpgradeRun", id, err)
	}

	runPostcheck(upgradeClientSet.UpgradeClient, d)

	d.SetId(id)
	return resourceNsxtUpgradeRunRead(d, m)
}

func prepareUpgrade(upgradeClientSet *upgradeClientSet, d *schema.ResourceData) error {
	for i := range upgradeComponentList {
		component := upgradeComponentList[i]
		// Customize MP upgrade is not allowed
		if component == mpUpgradeGroup {
			continue
		}

		if !d.HasChange(componentToGroupKey[component]) && !d.HasChange(componentToSettingKey[component]) {
			continue
		}

		status, err := getUpgradeStatus(upgradeClientSet.StatusClient, &component)
		if err != nil {
			return err
		}

		if status.Status == model.ComponentUpgradeStatus_STATUS_SUCCESS {
			log.Printf("[WARN] %s upgrade is already succeed. Any changes on it will be ignored.", component)
			continue
		}
		// If a component upgrade is in progress, to update either UpgradeUnitGroup or UpgradePlanSetting,
		// we should pause it first. Update an in-flight component will receive an error from API.
		if status.Status == model.ComponentUpgradeStatus_STATUS_IN_PROGRESS {
			upgradeClientSet.PlanClient.Pause()
		}
		err = waitUpgradeForStatus(upgradeClientSet, &component, inFlightComponentUpgradeStatus, staticComponentUpgradeStatus)
		if err != nil {
			return err
		}

		// Call reset regardless, because we don't know if UpgradeUnitGroup or UpgradePlanSetting has been changed.
		err = upgradeClientSet.PlanClient.Reset(component)
		if err != nil {
			return err
		}

		err = updateUpgradeUnitGroups(upgradeClientSet, d, component)
		if err != nil {
			return err
		}

		err = updateComponentUpgradePlanSetting(upgradeClientSet.SettingClient, d, component)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPartialUpgradeMap(d *schema.ResourceData) map[string]bool {
	isPartialUpgradeMap := map[string]bool{
		edgeUpgradeGroup: false,
		hostUpgradeGroup: false,
	}
	for _, component := range upgradeComponentList {
		if component == mpUpgradeGroup {
			continue
		}
		for _, groupI := range d.Get(componentToGroupKey[component]).([]interface{}) {
			group := groupI.(map[string]interface{})
			enabled := group["enabled"].(bool)
			pauseAfterEach := group["pause_after_each_upgrade_unit"].(bool)
			if !enabled || pauseAfterEach {
				isPartialUpgradeMap[component] = true
				break
			}
		}
	}
	return isPartialUpgradeMap
}

type upgradeStatusAndDetail struct {
	Status string
	Detail string
}

// Get component upgrade status. Using nil component for overall upgrade status.
func getUpgradeStatus(statusClient upgrade.StatusSummaryClient, component *string) (*upgradeStatusAndDetail, error) {
	status, err := statusClient.Get(component, nil, nil)
	if err != nil {
		return nil, err
	}
	if component == nil {
		return &upgradeStatusAndDetail{Status: *status.OverallUpgradeStatus}, nil
	}
	for _, componentStatus := range status.ComponentStatus {
		if *componentStatus.ComponentType == *component {
			detail := ""
			if componentStatus.Details != nil {
				detail = *componentStatus.Details
			}
			return &upgradeStatusAndDetail{Status: *componentStatus.Status, Detail: detail}, nil
		}
	}
	return nil, fmt.Errorf("couldn't find upgrade status of %s component", *component)
}

// Wait component upgrade status to become target status. Using nil component for overall upgrade status.
func waitUpgradeForStatus(upgradeClientSet *upgradeClientSet, component *string, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: func() (interface{}, string, error) {
			status, err := getUpgradeStatus(upgradeClientSet.StatusClient, component)
			if component != nil && *component == mpUpgradeGroup && (isServiceUnavailableError(err) || isTimeoutError(err)) {
				// After MP upgrade is completed, NSXT will restart and service_unavailable error or timeout error will be received depending on the request timing.
				// Keep polling for this case.
				return model.ComponentUpgradeStatus_STATUS_IN_PROGRESS, model.ComponentUpgradeStatus_STATUS_IN_PROGRESS, nil
			}
			if err != nil {
				return status, model.ComponentUpgradeStatus_STATUS_FAILED, err
			}
			log.Printf("[DEBUG] Current upgrade status: %s", status.Status)
			return status, status.Status, nil
		},
		Timeout:      time.Duration(upgradeClientSet.Timeout) * time.Second,
		PollInterval: time.Duration(upgradeClientSet.Interval) * time.Second,
		Delay:        time.Duration(upgradeClientSet.Delay) * time.Second,
	}
	statusI, err := stateConf.WaitForState()
	if err != nil {
		statusDetail := ""
		if statusI != nil {
			status := statusI.(*upgradeStatusAndDetail)
			statusDetail = fmt.Sprintf(" Current status: %s. Details: %s", status.Status, status.Detail)
		}
		return fmt.Errorf("failed to wait Upgrade to be %s: %v. %s", target, err, statusDetail)
	}
	return nil
}

func updateUpgradeUnitGroups(upgradeClientSet *upgradeClientSet, d *schema.ResourceData, component string) error {
	isBefore := false
	getReorderAfterReq := func(id string) model.ReorderRequest {
		return model.ReorderRequest{
			Id:       &id,
			IsBefore: &isBefore,
		}
	}

	preUpgradeGroupID := ""
	for _, groupI := range d.Get(componentToGroupKey[component]).([]interface{}) {
		group := groupI.(map[string]interface{})
		groupID := group["id"].(string)
		groupGet, err := upgradeClientSet.GroupClient.Get(groupID, nil)
		if err != nil {
			return err
		}

		enabled := group["enabled"].(bool)
		pause := group["pause_after_each_upgrade_unit"].(bool)
		groupGet.Enabled = &enabled
		groupGet.PauseAfterEachUpgradeUnit = &pause

		// Parallel can't be modified for EDGE upgrade unit group
		if component != edgeUpgradeGroup {
			parallel := group["parallel"].(bool)
			groupGet.Parallel = &parallel
		}

		if component == hostUpgradeGroup {
			upgradeMode := group["upgrade_mode"].(string)
			mmcVsanMode := group["maintenance_mode_config_vsan_mode"].(string)
			mmcEvacuateOffVms := group["maintenance_mode_config_evacuate_powered_off_vms"].(bool)
			rebootlessUpgrade := group["rebootless_upgrade"].(bool)

			var extendConfig []model.KeyValuePair
			if upgradeMode != "" {
				upgradeModeKey := "upgrade_mode"
				extendConfig = append(extendConfig, model.KeyValuePair{Key: &upgradeModeKey, Value: &upgradeMode})
			}
			if mmcVsanMode != "" {
				mmcVsanModeKey := "maintenance_mode_config_vsan_mode"
				extendConfig = append(extendConfig, model.KeyValuePair{Key: &mmcVsanModeKey, Value: &mmcVsanMode})
			}

			mmcEvacuateOffVmsStr := "false"
			mmcEvacuateOffVmsKey := "maintenance_mode_config_evacuate_powered_off_vms"
			if mmcEvacuateOffVms {
				mmcEvacuateOffVmsStr = "true"
			}
			extendConfig = append(extendConfig, model.KeyValuePair{Key: &mmcEvacuateOffVmsKey, Value: &mmcEvacuateOffVmsStr})

			rebootlessUpgradeStr := "false"
			rebootlessUpgradeKey := "rebootless_upgrade"
			if rebootlessUpgrade {
				rebootlessUpgradeStr = "true"
			}
			extendConfig = append(extendConfig, model.KeyValuePair{Key: &rebootlessUpgradeKey, Value: &rebootlessUpgradeStr})
			groupGet.ExtendedConfiguration = extendConfig
		}

		_, err = upgradeClientSet.GroupClient.Update(groupID, groupGet)
		if err != nil {
			return err
		}

		if preUpgradeGroupID != "" {
			err = upgradeClientSet.GroupClient.Reorder(groupID, getReorderAfterReq(preUpgradeGroupID))
			if err != nil {
				return err
			}
		}
		preUpgradeGroupID = groupID
	}
	return nil
}

func updateComponentUpgradePlanSetting(settingClient plan.SettingsClient, d *schema.ResourceData, component string) error {
	settingI := d.Get(componentToSettingKey[component]).([]interface{})
	if len(settingI) == 0 {
		return nil
	}

	settingGet, err := settingClient.Get(component)
	if err != nil {
		return err
	}

	setting := settingI[0].(map[string]interface{})
	parallel := setting["parallel"].(bool)
	settingGet.Parallel = &parallel

	// PauseOnError can't be modified for EDGE upgrade setting
	if component != edgeUpgradeGroup {
		stopOnErr := setting["stop_on_error"].(bool)
		settingGet.PauseOnError = &stopOnErr
	}

	_, err = settingClient.Update(component, settingGet)
	return err
}

func runUpgrade(upgradeClientSet *upgradeClientSet, partialUpgradeMap map[string]bool) error {
	partialUpgradeExist := false
	for i := range upgradeComponentList {
		// After one component upgrade is completed, although the status of our next component is NOT_STARTED,
		// there is a period that overall status is still IN_PROGRESS, which will prevent us to start the upgrade of next component.
		// Wait here for the overall status become stable. Because there is potential upgrade triggered before, we wait here also
		// for the first component for safety.
		err := waitUpgradeForStatus(upgradeClientSet, nil, inFlightComponentUpgradeStatus, staticComponentUpgradeStatus)
		if err != nil {
			return err
		}

		component := upgradeComponentList[i]

		if component == mpUpgradeGroup && partialUpgradeExist {
			log.Printf("[INFO] Some UpgradeUnitGroups haven't been upgraded. MP upgrade is skipped")
			continue
		}
		pendingStatus := []string{model.ComponentUpgradeStatus_STATUS_IN_PROGRESS}
		targetStatus := []string{model.ComponentUpgradeStatus_STATUS_SUCCESS}
		completeLog := fmt.Sprintf("[INFO] %s upgrade is completed.", component)
		if partialUpgradeMap[component] {
			// For partial upgrade, some groups are disabled, component upgrade status will be paused after all enabled groups upgraded.
			pendingStatus = append(pendingStatus, model.ComponentUpgradeStatus_STATUS_PAUSING)
			targetStatus = append(targetStatus, model.ComponentUpgradeStatus_STATUS_PAUSED)
			partialUpgradeExist = true
			completeLog = fmt.Sprintf("[INFO] %s upgrade is partially completed.", component)
		}
		upgradeClientSet.PlanClient.Upgrade(&component)
		err = waitUpgradeForStatus(upgradeClientSet, &component, pendingStatus, targetStatus)
		if err != nil {
			return err
		}
		log.Print(completeLog)
	}
	return nil
}

func runPostcheck(upgradeClient nsx.UpgradeClient, d *schema.ResourceData) {
	for i := range upgradeComponentList {
		component := upgradeComponentList[i]
		if component != mpUpgradeGroup {
			settingI := d.Get(componentToSettingKey[component]).([]interface{})
			if len(settingI) == 0 {
				continue
			}

			setting := settingI[0].(map[string]interface{})
			postCheck := setting["post_upgrade_check"].(bool)

			if postCheck {
				log.Printf("[INFO] Start %s upgrade postcheck. Please use data source nsxt_upgrade_postcheck for results.", component)
				upgradeClient.Executepostupgradechecks(component)
			}
		}
	}
}

func setUpgradeRunOutput(upgradeClientSet *upgradeClientSet, d *schema.ResourceData) error {
	results, err := upgradeClientSet.GroupClient.List(nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}

	var plans []map[string]interface{}
	for _, result := range results.Results {
		elem := make(map[string]interface{})
		elem["id"] = result.Id
		elem["parallel"] = result.Parallel
		elem["enabled"] = result.Enabled
		elem["pause_after_each_upgrade_unit"] = result.PauseAfterEachUpgradeUnit
		elem["type"] = result.Type_
		extConfig := make(map[string]string)
		for _, config := range result.ExtendedConfiguration {
			extConfig[*config.Key] = *config.Value
		}
		elem["extended_config"] = extConfig
		plans = append(plans, elem)
	}
	d.Set("upgrade_group_plan", plans)

	status, err := upgradeClientSet.StatusClient.Get(nil, nil, nil)
	if err != nil {
		return err
	}
	var states []map[string]interface{}
	for _, result := range status.ComponentStatus {
		elem := make(map[string]interface{})
		elem["type"] = *result.ComponentType
		elem["status"] = *result.Status
		if result.Details != nil {
			elem["details"] = *result.Details
		}
		groupStatusList, err := upgradeClientSet.GroupStatusClient.Getall(result.ComponentType, nil, nil, nil, nil, nil)
		if err != nil {
			return err
		}
		var groupStates []map[string]interface{}
		for _, groupStatusInList := range groupStatusList.Results {
			groupElem := make(map[string]interface{})
			groupElem["group_id"] = *groupStatusInList.GroupId
			groupElem["group_name"] = *groupStatusInList.GroupName
			groupElem["status"] = *groupStatusInList.Status
			groupStates = append(groupStates, groupElem)
		}
		elem["group_state"] = groupStates
		states = append(states, elem)
	}
	d.Set("state", states)
	return nil
}

func resourceNsxtUpgradeRunRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	connector := getPolicyConnector(m)
	upgradeClientSet := newUpgradeClientSet(connector, d)
	err := setUpgradeRunOutput(upgradeClientSet, d)
	if err != nil {
		return handleReadError(d, "NsxtUpgradeRun", id, err)
	}
	return nil
}

func resourceNsxtUpgradeRunUpdate(d *schema.ResourceData, m interface{}) error {
	return upgradeRunCreateOrUpdate(d, m)
}

func resourceNsxtUpgradeRunDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
