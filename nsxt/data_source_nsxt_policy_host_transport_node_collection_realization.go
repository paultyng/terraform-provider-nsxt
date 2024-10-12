// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points/transport_node_collections"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyHostTransportNodeCollectionRealization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyHostTransportNodeCollectionRealizationRead,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:         schema.TypeString,
				Description:  "Path of this Transport Node Collection",
				Optional:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"timeout": {
				Type:         schema.TypeInt,
				Description:  "Realization timeout in seconds",
				Optional:     true,
				Default:      1200,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"delay": {
				Type:         schema.TypeInt,
				Description:  "Initial delay to start realization checks in seconds",
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"state": {
				Type:        schema.TypeString,
				Description: "Application state of transport node profile on compute collection",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyHostTransportNodeCollectionRealizationRead(d *schema.ResourceData, m interface{}) error {
	if isPolicyGlobalManager(m) {
		return localManagerOnlyError()
	}
	connector := getPolicyConnector(m)
	client := transport_node_collections.NewStateClient(connector)

	path := d.Get("path").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	pendingStates := []string{model.TransportNodeCollectionState_STATE_IN_PROGRESS}
	targetStates := []string{
		model.TransportNodeCollectionState_STATE_FAILED_TO_CREATE,
		model.TransportNodeCollectionState_STATE_FAILED_TO_REALIZE,
		model.TransportNodeCollectionState_STATE_PROFILE_MISMATCH,
		model.TransportNodeCollectionState_STATE_SUCCESS,
	}

	site, err := getParameterFromPolicyPath("/sites/", "/enforcement-points/", path)
	if err != nil {
		return fmt.Errorf("Invalid transport node collection path %s", path)
	}

	ep, err1 := getParameterFromPolicyPath("/enforcement-points/", "/transport-node-collections/", path)
	if err1 != nil {
		return fmt.Errorf("Invalid transport node collection path %s", path)
	}

	objID := getPolicyIDFromPath(path)

	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(site, ep, objID)
			if err != nil {
				return state, model.TransportNodeCollectionState_STATE_IN_PROGRESS, logAPIError("Error getting collection state", err)
			}

			log.Printf("[DEBUG] Current realization state for Transport Node Collection %s is %s", objID, *state.State)

			if *state.State != model.TransportNodeCollectionState_STATE_SUCCESS && *state.State != model.TransportNodeCollectionState_STATE_IN_PROGRESS {
				return state, *state.State, getErrorFromState(&state)
			}

			d.SetId(objID)
			d.Set("state", state.State)
			return state, *state.State, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      time.Duration(delay) * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}
	return nil
}

func getErrorFromState(state *model.TransportNodeCollectionState) error {
	result := fmt.Sprintf("state: %s\n", *state.State)
	if state.ClusterLevelError != nil {
		result += fmt.Sprintf("cluster level error: %v\n", *state.ClusterLevelError)
	}
	for _, item := range state.ValidationErrors {
		result += fmt.Sprintf("validation error for node %s: %v\n", *item.DiscoveredNodeId, *item.ErrorMessage)
	}
	if state.VlcmTransitionError != nil {
		result += fmt.Sprintf("VCLM transition error: %v\n", *state.VlcmTransitionError)
	}

	return errors.New(result)
}
