// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var nsxtPolicyVMIDSchemaKeys = []string{
	"external_id",
	"bios_id",
	"instance_id",
}

func dataSourceNsxtPolicyVM() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyVMIDRead,

		Schema: map[string]*schema.Schema{
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"external_id":  getDataSourceStringSchema("External ID of the Virtual Machine"),
			"bios_id":      getDataSourceStringSchema("BIOS UUID of the Virtual Machine"),
			"instance_id":  getDataSourceStringSchema("Instance UUID of the Virtual Machine"),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
		},
	}
}

func getNsxtPolicyVMIDFromSchema(d *schema.ResourceData) string {
	for _, k := range nsxtPolicyVMIDSchemaKeys {
		v := d.Get(k).(string)
		if v != "" {
			return v
		}
	}
	return ""
}

func dataSourceNsxtPolicyVMIDRead(d *schema.ResourceData, m interface{}) error {
	var vmModel model.VirtualMachine
	connector := getPolicyConnector(m)

	objID := getNsxtPolicyVMIDFromSchema(d)
	context := getSessionContext(d, m)

	if objID != "" {
		vmObj, err := findNsxtPolicyVMByID(context, connector, objID, m)
		if err != nil {
			return fmt.Errorf("Error while reading Virtual Machine %s: %v", objID, err)
		}
		vmModel = vmObj
	} else {
		displayName := d.Get("display_name").(string)

		perfectMatch, prefixMatch, err := findNsxtPolicyVMByNamePrefix(context, connector, displayName, m)
		if err != nil {
			return err
		}

		foundLen := len(perfectMatch) + len(prefixMatch)
		if foundLen == 0 {
			return fmt.Errorf("Unable to find Virtual Machine with name prefix: %s", displayName)
		}
		if foundLen > 1 {
			return fmt.Errorf("Found %v Virtual Machines with name prefix: %s", foundLen, displayName)
		}
		if len(perfectMatch) > 0 {
			vmModel = perfectMatch[0]
		} else {
			vmModel = prefixMatch[0]
		}
	}

	computeIDMap := collectSeparatedStringListToMap(vmModel.ComputeIds, ":")
	if vmModel.ExternalId == nil {
		return fmt.Errorf("Unable to read external ID for Virtual Machine with name %s", *vmModel.DisplayName)
	}
	d.SetId(*vmModel.ExternalId)
	d.Set("display_name", vmModel.DisplayName)
	d.Set("description", vmModel.Description)
	d.Set("external_id", *vmModel.ExternalId)
	d.Set("bios_id", computeIDMap[nsxtPolicyBiosUUIDKey])
	d.Set("instance_id", computeIDMap[nsxtPolicyInstanceUUIDKey])
	setPolicyTagsInSchema(d, vmModel.Tags)

	return nil
}
