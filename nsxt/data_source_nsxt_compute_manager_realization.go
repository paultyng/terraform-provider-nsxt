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
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/fabric/compute_managers"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func dataSourceNsxtComputeManagerRealization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtComputeManagerRealizationRead,

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
			"check_registration": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Check if registration of compute manager is complete",
			},
			"registration_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Overall registration status of desired configuration",
			},
		},
	}
}

func dataSourceNsxtComputeManagerRealizationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	checkRegistration := d.Get("check_registration").(bool)

	err := dataSourceNsxtComputeManagerRealizationWait(d, connector)

	if !checkRegistration {
		return err
	}

	return dataSourceNsxtComputeManagerRegistrationWait(d, connector)
}

func dataSourceNsxtComputeManagerRealizationWait(d *schema.ResourceData, connector client.Connector) error {
	id := d.Get("id").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	client := compute_managers.NewStateClient(connector)

	pendingStates := []string{
		model.ConfigurationState_STATE_PENDING,
		model.ConfigurationState_STATE_IN_PROGRESS,
		model.ConfigurationState_STATE_UNKNOWN,
		model.ConfigurationState_STATE_IN_SYNC,
	}
	targetStates := []string{
		model.ConfigurationState_STATE_SUCCESS,
		model.ConfigurationState_STATE_FAILED,
		model.ConfigurationState_STATE_PARTIAL_SUCCESS,
		model.ConfigurationState_STATE_ORPHANED,
		model.ConfigurationState_STATE_ERROR,
	}

	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(id)
			if err != nil {
				return state, model.ConfigurationState_STATE_ERROR, logAPIError("Error while waiting for realization of Compute Manager", err)
			}

			log.Printf("[DEBUG] Current realization state for Compute Manager %s is %s", id, *state.State)
			if *state.State == model.ConfigurationState_STATE_ERROR || *state.State == model.ConfigurationState_STATE_FAILED {
				return state, *state.State, fmt.Errorf("compute manager %s failed to realize with state %v", id, *state.State)
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

func dataSourceNsxtComputeManagerRegistrationWait(d *schema.ResourceData, connector client.Connector) error {
	id := d.Get("id").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	client := compute_managers.NewStatusClient(connector)

	pendingStates := []string{
		model.ComputeManagerStatus_CONNECTION_STATUS_CONNECTING,
		model.ComputeManagerStatus_REGISTRATION_STATUS_REGISTERING,
	}
	targetStates := []string{
		model.ComputeManagerStatus_REGISTRATION_STATUS_REGISTERED,
		model.ComputeManagerStatus_REGISTRATION_STATUS_UNREGISTERED,
		model.ComputeManagerStatus_REGISTRATION_STATUS_REGISTERED_WITH_ERRORS,
	}

	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			status, err := client.Get(id)
			if err != nil {
				return status, model.ComputeManagerStatus_REGISTRATION_STATUS_REGISTERED_WITH_ERRORS, logAPIError("Error while waiting for realization of Compute Manager", err)
			}

			log.Printf("[DEBUG] Current registration status for Compute Manager %s is %s", id, *status.RegistrationStatus)

			d.Set("registration_status", status.RegistrationStatus)
			return status, *status.RegistrationStatus, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      time.Duration(delay) * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("failed to get registration information for %s: %v", id, err)
	}
	return nil
}
