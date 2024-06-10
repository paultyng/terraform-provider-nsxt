/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	"github.com/vmware/terraform-provider-nsxt/api/infra/segments"
)

func dataSourceNsxtPolicySegmentRealization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicySegmentRealizationRead,

		Schema: map[string]*schema.Schema{
			"id":      getDataSourceIDSchema(),
			"context": getContextSchema(false, false, false),
			"path": {
				Type:         schema.TypeString,
				Description:  "The path for the policy segment",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The state of the realized resource on hypervisors",
				Computed:    true,
			},
			"network_name": {
				Type:        schema.TypeString,
				Description: "Network name on the hypervisors",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicySegmentRealizationRead(d *schema.ResourceData, m interface{}) error {
	// Read the realization info by the path, and wait till it is valid
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	commonProviderConfig := getCommonProviderConfig(m)

	// Get the realization info of this resource
	path := d.Get("path").(string)

	// Dummy id, just because each data source needs one
	id := d.Get("id").(string)
	if id == "" {
		d.SetId(newUUID())
	}

	// verifying segment realization on hypervisor
	segmentID := getPolicyIDFromPath(path)
	enforcementPointPath := getPolicyEnforcementPointPath(m)
	client := segments.NewStateClient(context, connector)
	pendingStates := []string{model.SegmentConfigurationState_STATE_PENDING,
		model.SegmentConfigurationState_STATE_IN_PROGRESS,
		model.SegmentConfigurationState_STATE_IN_SYNC,
		model.SegmentConfigurationState_STATE_UNKNOWN}
	targetStates := []string{model.SegmentConfigurationState_STATE_SUCCESS,
		model.SegmentConfigurationState_STATE_FAILED,
		model.SegmentConfigurationState_STATE_ERROR,
		model.SegmentConfigurationState_STATE_ORPHANED}
	if commonProviderConfig.ToleratePartialSuccess {
		targetStates = append(targetStates, model.SegmentConfigurationState_STATE_PARTIAL_SUCCESS)
	}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(segmentID, nil, nil, &enforcementPointPath, nil, nil, nil, nil, nil, nil, nil, nil)
			if err != nil {
				return state, model.SegmentConfigurationState_STATE_ERROR, logAPIError("Error while waiting for realization of segment", err)
			}

			log.Printf("[DEBUG] Current realization state for segment %s is %s", segmentID, *state.State)

			d.Set("state", state.State)
			return state, *state.State, nil
		},
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to get realization information for %s: %v", path, err)
	}

	// In some cases success state is returned a moment before VC actually sees the network
	// Adding a short sleep here prevents vsphere provider from erroring out
	time.Sleep(1 * time.Second)

	// We need to fetch network name to use in vpshere provider. However, state API does not
	// return it in details yet. For now, we'll use segment display name, since its always
	// translates to network name
	segClient := infra.NewSegmentsClient(context, connector)
	obj, err := segClient.Get(segmentID)
	if err != nil {
		return handleReadError(d, "Segment", segmentID, err)
	}

	d.Set("network_name", obj.DisplayName)

	return nil
}
