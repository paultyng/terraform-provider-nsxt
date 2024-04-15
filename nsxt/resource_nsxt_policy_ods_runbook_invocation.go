/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sha"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyODSRunbookInvocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyODSRunbookInvocationCreate,
		Read:   resourceNsxtPolicyODSRunbookInvocationRead,
		Update: resourceNsxtPolicyODSRunbookInvocationUpdate,
		Delete: resourceNsxtPolicyODSRunbookInvocationDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id": getNsxIDSchema(),
			"path":   getPathSchema(),
			// Due to a bug, invocations with a display_name specification fail, and when there's none set, NSX assigns
			// the id value to the display name attribute. This should work around that bug.
			"display_name": {
				Type:        schema.TypeString,
				Description: "Display name for this resource",
				Optional:    true,
				Computed:    true,
			},
			"revision": getRevisionSchema(),
			"argument": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Arguments for runbook invocation",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Key",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value",
						},
					},
				},
			},
			"runbook_path": getPolicyPathSchema(true, true, "Path of runbook object"),
			"target_node": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Identifier of an appliance node or transport node",
				ForceNew:    true,
			},
		},
	}
}

func getODSRunbookInvocationFromSchema(id string, d *schema.ResourceData) model.OdsRunbookInvocation {
	displayName := d.Get("display_name").(string)
	runbookPath := d.Get("runbook_path").(string)
	targetNode := d.Get("target_node").(string)

	var arguments []model.UnboundedKeyValuePair
	for _, arg := range d.Get("argument").(*schema.Set).List() {
		argMap := arg.(map[string]interface{})
		key := argMap["key"].(string)
		value := argMap["value"].(string)
		item := model.UnboundedKeyValuePair{
			Key:   &key,
			Value: &value,
		}
		arguments = append(arguments, item)
	}

	obj := model.OdsRunbookInvocation{
		Id:          &id,
		RunbookPath: &runbookPath,
		Arguments:   arguments,
		TargetNode:  &targetNode,
	}
	if displayName != "" {
		obj.DisplayName = &displayName
	}

	return obj
}

func resourceNsxtPolicyODSRunbookInvocationCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyODSRunbookInvocationExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := sha.NewRunbookInvocationsClient(connector)

	obj := getODSRunbookInvocationFromSchema(id, d)
	err = client.Create(id, obj)
	if err != nil {
		return handleCreateError("OdsRunbookInvocation", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyODSRunbookInvocationRead(d, m)
}

func resourceNsxtPolicyODSRunbookInvocationExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	client := sha.NewRunbookInvocationsClient(connector)
	_, err = client.Get(id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyODSRunbookInvocationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining OdsRunbookInvocation ID")
	}

	client := sha.NewRunbookInvocationsClient(connector)
	var err error
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "OdsRunbookInvocation", id, err)
	}

	if obj.DisplayName != nil && *obj.DisplayName != "" {
		d.Set("display_name", obj.DisplayName)
	}
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("runbook_path", obj.RunbookPath)
	d.Set("target_node", obj.TargetNode)

	var argList []map[string]interface{}
	for _, arg := range obj.Arguments {
		argData := make(map[string]interface{})
		argData["key"] = arg.Key
		argData["value"] = arg.Value
		argList = append(argList, argData)
	}
	d.Set("argument", argList)

	return nil
}

func resourceNsxtPolicyODSRunbookInvocationUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyODSRunbookInvocationRead(d, m)
}

func resourceNsxtPolicyODSRunbookInvocationDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining OdsRunbookInvocation ID")
	}

	connector := getPolicyConnector(m)
	var err error
	client := sha.NewRunbookInvocationsClient(connector)
	err = client.Delete(id)

	if err != nil {
		return handleDeleteError("OdsRunbookInvocation", id, err)
	}

	return nil
}
