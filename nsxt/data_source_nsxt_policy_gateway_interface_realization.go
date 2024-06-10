/* Copyright © 2024 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	realizedstate "github.com/vmware/terraform-provider-nsxt/api/infra/realized_state"
)

func dataSourceNsxtPolicyGatewayInterfaceRealization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayInterfaceRealizationRead,

		Schema: map[string]*schema.Schema{
			"id":      getDataSourceIDSchema(),
			"context": getContextSchema(false, false, false),
			"gateway_path": {
				Type:         schema.TypeString,
				Description:  "The path for the gateway",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of the gateway interface",
				Optional:    true,
			},
			"timeout": {
				Type:         schema.TypeInt,
				Description:  "Realization timeout in seconds",
				Optional:     true,
				Default:      600,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"delay": {
				Type:         schema.TypeInt,
				Description:  "Initial delay to start realization checks in seconds",
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The state of the realized gateway interface",
				Computed:    true,
			},
			"ip_address": {
				Type:        schema.TypeList,
				Description: "IP addresses of the gateway interface",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mac_address": {
				Type:        schema.TypeString,
				Description: "MAC address of the gateway interface",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyGatewayInterfaceRealizationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := realizedstate.NewRealizedEntitiesClient(context, connector)

	id := d.Get("id").(string)
	gatewayPath := d.Get("gateway_path").(string)
	displayName := d.Get("display_name").(string)
	delay := d.Get("delay").(int)
	timeout := d.Get("timeout").(int)

	pendingStates := []string{"UNKNOWN", "UNREALIZED"}
	targetStates := []string{"REALIZED", "ERROR"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			result, err := client.List(gatewayPath, nil)
			if err != nil {
				return result, "", err
			}

			var perfectMatch []model.GenericPolicyRealizedResource
			var containsMatch []model.GenericPolicyRealizedResource
			for _, objInList := range result.Results {
				if id != "" && objInList.Id != nil && id == *objInList.Id {
					if objInList.State == nil {
						return result, "UNKNOWN", nil
					}
					setGatewayInterfaceRealizationInSchema(objInList, d)
					return result, *objInList.State, nil
				} else if displayName != "" && objInList.DisplayName != nil {
					if displayName == *objInList.DisplayName {
						perfectMatch = append(perfectMatch, objInList)
					} else if strings.Contains(*objInList.DisplayName, displayName) {
						containsMatch = append(containsMatch, objInList)
					}
				} else {
					// If neither ID nor displayName is provided, return the one with IPAddresses, considering
					// it is the most common case.
					for _, ext := range objInList.ExtendedAttributes {
						if ext.Key != nil && *ext.Key == "IpAddresses" {
							setGatewayInterfaceRealizationInSchema(objInList, d)
							return result, *objInList.State, nil
						}
					}
				}
			}

			if id == "" && displayName == "" && len(result.Results) > 0 {
				// If neither ID nor displayName is provided and there is no entity contains IpAddresses,
				// return the first one.
				obj := result.Results[0]
				setGatewayInterfaceRealizationInSchema(obj, d)
				return result, *obj.State, nil
			}

			if len(perfectMatch) > 1 {
				return result, "", fmt.Errorf("Found multiple gateway interfaces with name '%s'", displayName)
			}
			if len(perfectMatch) > 0 {
				if perfectMatch[0].State == nil {
					return result, "UNKNOWN", nil
				}
				setGatewayInterfaceRealizationInSchema(perfectMatch[0], d)
				return result, *perfectMatch[0].State, nil
			}

			if len(containsMatch) > 1 {
				return result, "", fmt.Errorf("Found multiple gateway interfaces whose name contains '%s'", displayName)
			}
			if len(containsMatch) > 0 {
				if containsMatch[0].State == nil {
					return result, "UNKNOWN", nil
				}
				setGatewayInterfaceRealizationInSchema(containsMatch[0], d)
				return result, *containsMatch[0].State, nil
			}

			return result, "UNKNOWN", nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      time.Duration(delay) * time.Second,
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to get gateway interface realization information for %s: %v", gatewayPath, err)
	}
	return nil
}

func setGatewayInterfaceRealizationInSchema(realizedResource model.GenericPolicyRealizedResource, d *schema.ResourceData) {
	for _, ext := range realizedResource.ExtendedAttributes {
		if *ext.Key == "IpAddresses" && len(ext.Values) > 0 {
			d.Set("ip_address", ext.Values)
		}
		if *ext.Key == "MacAddress" && len(ext.Values) > 0 {
			d.Set("mac_address", ext.Values[0])
		}
	}
	d.Set("state", *realizedResource.State)
	d.SetId(*realizedResource.Id)
}
