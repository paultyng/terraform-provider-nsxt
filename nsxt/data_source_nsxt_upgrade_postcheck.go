// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade/upgrade_unit_groups"
)

var (
	// Default waiting setup in seconds
	defaultUpgradePostcheckInterval = 30
	defaultUpgradePostcheckTimeout  = 1800
	defaultUpgradePostcheckDelay    = 30
)

func dataSourceNsxtUpgradePostCheck() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtUpgradePostCheckRead,

		Schema: map[string]*schema.Schema{
			"upgrade_run_id": {
				Type:        schema.TypeString,
				Description: "ID of corresponding nsxt_upgrade_run resource",
				Required:    true,
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "Component Type",
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{edgeUpgradeGroup, hostUpgradeGroup}, false),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Description:  "Post-check status checks timeout in seconds",
				Optional:     true,
				Default:      defaultUpgradePostcheckTimeout,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"interval": {
				Type:         schema.TypeInt,
				Description:  "Interval to check Post-check status in seconds",
				Optional:     true,
				Default:      defaultUpgradePostcheckInterval,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"delay": {
				Type:         schema.TypeInt,
				Description:  "Initial delay to start Post-check status checks in seconds",
				Optional:     true,
				Default:      defaultUpgradePostcheckDelay,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"failed_group": {
				Type:        schema.TypeList,
				Description: "Upgrade unit groups that failed in upgrade post-checks",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Unique identifier of upgrade unit group",
							Required:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Status of execution of upgrade post-checks",
							Computed:    true,
						},
						"details": {
							Type:        schema.TypeString,
							Description: "Details about current execution of upgrade post-checks",
							Computed:    true,
						},
						"failure_count": {
							Type:        schema.TypeInt,
							Description: "Total count of generated failures or warnings in last execution of upgrade post-checks",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceNsxtUpgradePostCheckRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	aggregateInfoClient := upgrade_unit_groups.NewAggregateInfoClient(connector)
	component := d.Get("type").(string)
	timeout := d.Get("timeout").(int)
	interval := d.Get("interval").(int)
	delay := d.Get("delay").(int)

	stateConf := &resource.StateChangeConf{
		Pending: []string{model.UpgradeChecksExecutionStatus_STATUS_IN_PROGRESS},
		Target:  []string{model.UpgradeChecksExecutionStatus_STATUS_COMPLETED},
		Refresh: func() (interface{}, string, error) {
			listRes, err := aggregateInfoClient.List(&component, nil, nil, nil, nil, nil, nil, nil)
			if err != nil {
				return nil, "", fmt.Errorf("failed to retrieve %s upgrade post-check results: %v", component, err)
			}
			for _, res := range listRes.Results {
				if *res.PostUpgradeStatus.Status == model.UpgradeChecksExecutionStatus_STATUS_IN_PROGRESS {
					return res.PostUpgradeStatus, model.UpgradeChecksExecutionStatus_STATUS_IN_PROGRESS, nil
				}
			}
			// Return UpgradeChecksExecutionStatus_STATUS_COMPLETED here doesn't mean all UpgradeUnitGroup post-checks are completed.
			// Just to stop polling the states as none of UpgradeUnitGroup is in in_progress state.
			return listRes, model.UpgradeChecksExecutionStatus_STATUS_COMPLETED, nil
		},
		Timeout:      time.Duration(timeout) * time.Second,
		PollInterval: time.Duration(interval) * time.Second,
		Delay:        time.Duration(delay) * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	listRes, err := aggregateInfoClient.List(&component, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to retrieve %s upgrade post-check result: %v", component, err)
	}
	var failedGroup []map[string]interface{}
	for _, res := range listRes.Results {
		failureCounts := 0
		if res.PostUpgradeStatus.FailureCount != nil {
			failureCounts = int(*res.PostUpgradeStatus.FailureCount)
		} else if res.PostUpgradeStatus.NodeWithIssuesCount != nil {
			failureCounts = int(*res.PostUpgradeStatus.NodeWithIssuesCount)
		}
		if res.PostUpgradeStatus.Status != nil && *res.PostUpgradeStatus.Status != model.UpgradeChecksExecutionStatus_STATUS_COMPLETED || failureCounts > 0 {
			elem := make(map[string]interface{}, 4)
			elem["status"] = *res.PostUpgradeStatus.Status
			elem["failure_count"] = failureCounts
			if res.Id != nil {
				elem["id"] = *res.Id
			}
			if res.PostUpgradeStatus.Details != nil {
				elem["details"] = *res.PostUpgradeStatus.Details
			}
			failedGroup = append(failedGroup, elem)
		}
	}
	if len(failedGroup) > 0 {
		d.Set("failed_group", failedGroup)
	}

	return nil
}
