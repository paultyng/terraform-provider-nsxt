/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var valueTypeValues = []string{"bios_id", "external_id", "instance_id"}

var stateMap = map[string]string{
	"running":   model.VirtualMachine_POWER_STATE_VM_RUNNING,
	"stopped":   model.VirtualMachine_POWER_STATE_VM_STOPPED,
	"suspended": model.VirtualMachine_POWER_STATE_VM_SUSPENDED,
	"unknown":   model.VirtualMachine_POWER_STATE_UNKNOWN,
}

func dataSourceNsxtPolicyVMs() *schema.Resource {
	stateMapKeys := []string{}
	for k := range stateMap {
		stateMapKeys = append(stateMapKeys, k)
	}

	return &schema.Resource{
		Read: dataSourceNsxtPolicyVMsRead,

		Schema: map[string]*schema.Schema{
			// TODO: add option to filter by display name regex
			"value_type": {
				Type:         schema.TypeString,
				Description:  "Type of data populated in map value",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(valueTypeValues, false),
				Default:      "bios_id",
			},
			"state": {
				Type:         schema.TypeString,
				Description:  "Power state of the VM",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(stateMapKeys, false),
			},
			"guest_os": {
				Type:        schema.TypeString,
				Description: "Operating system",
				Optional:    true,
			},
			"items": {
				Type:        schema.TypeMap,
				Description: "Mapping of VM instance ID by display name",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"context": getContextSchema(false, false, false),
		},
	}
}

func dataSourceNsxtPolicyVMsRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	valueType := d.Get("value_type").(string)
	state := d.Get("state").(string)
	osPrefix := d.Get("guest_os").(string)
	vmMap := make(map[string]interface{})

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	allVMs, err := listAllPolicyVirtualMachines(context, connector, m)
	if err != nil {
		return fmt.Errorf("Error reading Virtual Machines: %v", err)
	}

	for _, vm := range allVMs {
		if state != "" {
			if vm.PowerState != nil && *vm.PowerState != stateMap[state] {
				continue
			}
		}

		if osPrefix != "" {
			if vm.GuestInfo != nil && vm.GuestInfo.OsName != nil {
				osName := strings.ToLower(*vm.GuestInfo.OsName)
				if !strings.HasPrefix(osName, strings.ToLower(osPrefix)) {
					continue
				}
			}
		}
		if vm.DisplayName == nil {
			continue
		}
		computeIDMap := collectSeparatedStringListToMap(vm.ComputeIds, ":")
		if valueType == "instance_id" {
			vmMap[*vm.DisplayName] = computeIDMap[nsxtPolicyInstanceUUIDKey]
		} else if valueType == "bios_id" {
			vmMap[*vm.DisplayName] = computeIDMap[nsxtPolicyBiosUUIDKey]
		} else if valueType == "external_id" {
			if vm.ExternalId == nil {
				// skip this vm
				continue
			}
			vmMap[*vm.DisplayName] = *vm.ExternalId
		}
	}

	d.SetId(newUUID())
	d.Set("items", vmMap)

	return nil
}
