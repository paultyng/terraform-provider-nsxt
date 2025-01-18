// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/transport_nodes"
)

func dataSourceNsxtTransportNodeRealization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtTransportNodeRealizationRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique ID of this resource",
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
				Computed:    true,
				Description: "Overall state of desired configuration",
			},
		},
	}
}

func dataSourceNsxtTransportNodeRealizationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := transport_nodes.NewStateClient(connector)

	id := d.Get("id").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	pendingStates := []string{
		model.TransportNodeState_STATE_PENDING,
		model.TransportNodeState_STATE_IN_PROGRESS,
		model.TransportNodeState_STATE_IN_SYNC,
		model.TransportNodeState_STATE_UNKNOWN,
	}
	targetStates := []string{
		model.TransportNodeState_STATE_SUCCESS,
		model.TransportNodeState_STATE_FAILED,
		model.TransportNodeState_STATE_PARTIAL_SUCCESS,
		model.TransportNodeState_STATE_ORPHANED,
		model.TransportNodeState_STATE_ERROR,
	}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(id)
			if err != nil {
				return state, model.TransportNodeState_STATE_ERROR, logAPIError("Error while waiting for realization of Transport Node", err)
			}

			log.Printf("[DEBUG] Current realization state for Transport Node %s is %s", id, *state.State)
			if *state.State == model.TransportNodeState_STATE_FAILED || *state.State == model.TransportNodeState_STATE_ERROR {
				return state, *state.State, fmt.Errorf("transport node %s failed to realize with state %v: %v", id, state, state.FailureMessage)
			}

			d.Set("state", state.State)
			return state, *state.State, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      time.Duration(delay) * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return err
	}

	return nil
}
