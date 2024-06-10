/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	realizedstate "github.com/vmware/terraform-provider-nsxt/api/infra/realized_state"
)

func dataSourceNsxtPolicyRealizationInfo() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyRealizationInfoRead,

		Schema: map[string]*schema.Schema{
			"id": getDataSourceIDSchema(),
			"path": {
				Type:         schema.TypeString,
				Description:  "The path for the policy resource",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"context": getContextSchema(false, false, false),
			"entity_type": {
				Type:        schema.TypeString,
				Description: "The entity type of the realized resource",
				Computed:    true,
				Optional:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The state of the realized resource",
				Computed:    true,
			},
			"realized_id": {
				Type:        schema.TypeString,
				Description: "The ID of the realized resource",
				Computed:    true,
			},
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path of the site this resource belongs to",
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
		},
	}
}

func dataSourceNsxtPolicyRealizationInfoRead(d *schema.ResourceData, m interface{}) error {
	// Read the realization info by the path, and wait till it is valid
	connector := getPolicyConnector(m)

	// Get the realization info of this resource
	path := d.Get("path").(string)
	entityType := d.Get("entity_type").(string)
	objSitePath := d.Get("site_path").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	// Site is mandatory got GM and irrelevant else
	if !isPolicyGlobalManager(m) && objSitePath != "" {
		return globalManagerOnlyError()
	}
	if isPolicyGlobalManager(m) {
		if objSitePath == "" {
			return attributeRequiredGlobalManagerError("site_path", "nsxt_policy_realization_info")
		}
	}

	// Dummy id, just because each data source needs one
	id := d.Get("id").(string)
	if id == "" {
		d.SetId(newUUID())
	}

	pendingStates := []string{"UNKNOWN", "UNREALIZED"}
	targetStates := []string{"REALIZED", "ERROR"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {

			var realizationError error
			var realizationResult model.GenericPolicyRealizedResourceListResult
			client := realizedstate.NewRealizedEntitiesClient(context, connector)
			realizationResult, realizationError = client.List(path, nil)
			state := "UNKNOWN"
			if realizationError == nil {
				// Find the right entry
				for _, objInList := range realizationResult.Results {
					if objInList.State != nil {
						state = *objInList.State
					}
					if entityType == "" {
						// Take the first one
						d.Set("state", state)
						if objInList.EntityType != nil {
							d.Set("entity_type", *objInList.EntityType)
						}
						if objInList.RealizationSpecificIdentifier == nil {
							d.Set("realized_id", "")
						} else {
							d.Set("realized_id", *objInList.RealizationSpecificIdentifier)
						}
						return realizationResult, state, nil
					} else if (objInList.EntityType != nil) && (*objInList.EntityType == entityType) {
						d.Set("state", state)
						if objInList.RealizationSpecificIdentifier == nil {
							d.Set("realized_id", "")
						} else {
							d.Set("realized_id", *objInList.RealizationSpecificIdentifier)
						}
						return realizationResult, state, nil
					}
				}
				// Realization info not found yet
				d.Set("state", "UNKNOWN")
				d.Set("realized_id", "")
				return realizationResult, "UNKNOWN", nil
			}
			return realizationResult, "", realizationError
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      time.Duration(delay) * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to get realization information for %s: %v", path, err)
	}
	return nil
}
