// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtLogicalPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalPortCreate,
		Read:   resourceNsxtLogicalPortRead,
		Update: resourceNsxtLogicalPortUpdate,
		Delete: resourceNsxtLogicalPortDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"logical_switch_id": {
				Type:        schema.TypeString,
				Description: "Id of the Logical switch that this port belongs to",
				Required:    true,
				ForceNew:    true, // Cannot change the logical switch of a logical port
			},
			"admin_state":          getAdminStateSchema(),
			"switching_profile_id": getSwitchingProfileIdsSchema(),
			"tag":                  getTagsSchema(),
		},
	}
}

func resourceNsxtLogicalPortCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	name := d.Get("display_name").(string)
	description := d.Get("description").(string)
	lsID := d.Get("logical_switch_id").(string)
	adminState := d.Get("admin_state").(string)
	profilesList := getSwitchingProfileIdsFromSchema(d)
	tagList := getTagsFromSchema(d)

	lp := manager.LogicalPort{
		DisplayName:         name,
		Description:         description,
		LogicalSwitchId:     lsID,
		AdminState:          adminState,
		SwitchingProfileIds: profilesList,
		Tags:                tagList}

	lp, resp, err := nsxClient.LogicalSwitchingApi.CreateLogicalPort(nsxClient.Context, lp)

	if err != nil {
		return fmt.Errorf("Error while creating logical port %s: %v", lp.DisplayName, err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during Logical port create: %v", resp.StatusCode)
	}

	resourceID := lp.Id
	d.SetId(resourceID)

	return resourceNsxtLogicalPortRead(d, m)
}

func resourceNsxtLogicalPortRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical port ID from state during read")
	}
	logicalPort, resp, err := nsxClient.LogicalSwitchingApi.GetLogicalPort(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		d.SetId("")
		log.Printf("[DEBUG] Logical port %s not found", id)
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error while reading logical port %s: %v", id, err)
	}

	d.Set("revision", logicalPort.Revision)
	d.Set("display_name", logicalPort.DisplayName)
	d.Set("description", logicalPort.Description)
	d.Set("logical_switch_id", logicalPort.LogicalSwitchId)
	d.Set("admin_state", logicalPort.AdminState)
	err = setSwitchingProfileIdsInSchema(d, nsxClient, logicalPort.SwitchingProfileIds)
	if err != nil {
		return fmt.Errorf("Error during logical port switching profiles set in schema: %v", err)
	}
	setTagsInSchema(d, logicalPort.Tags)

	return nil
}

func resourceNsxtLogicalPortUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	name := d.Get("display_name").(string)
	description := d.Get("description").(string)
	adminState := d.Get("admin_state").(string)
	profilesList := getSwitchingProfileIdsFromSchema(d)
	tagList := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))

	// Some of the port attributes (attachment) are not exposed to terraform.
	// If we try to update port based on terraform attributes only, apply will fail
	// due to missing info.
	// We don't expose attachment to terraform, since it will become out of sync
	// once attachment info is updated/remove outside the scope of port management.

	lp, resp, err := nsxClient.LogicalSwitchingApi.GetLogicalPort(nsxClient.Context, id)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Logical port %s was not found", id)
	}
	if err != nil {
		return fmt.Errorf("Error while reading logical port %s: %v", id, err)
	}

	lp.DisplayName = name
	lp.Description = description
	lp.AdminState = adminState
	lp.SwitchingProfileIds = profilesList
	lp.Tags = tagList
	lp.Revision = revision

	lp, resp, err = nsxClient.LogicalSwitchingApi.UpdateLogicalPort(nsxClient.Context, id, lp)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error while updating logical port %s: %v", id, err)
	}
	return resourceNsxtLogicalPortRead(d, m)
}

func resourceNsxtLogicalPortDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	lpID := d.Id()
	if lpID == "" {
		return fmt.Errorf("Error obtaining logical port ID from state during delete")
	}
	localVarOptionals := make(map[string]interface{})
	localVarOptionals["detach"] = true

	resp, err := nsxClient.LogicalSwitchingApi.DeleteLogicalPort(nsxClient.Context, lpID, localVarOptionals)

	if err != nil {
		return fmt.Errorf("Error while deleting logical port %s: %v", lpID, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] Logical port %s was not found", lpID)
		d.SetId("")
	}

	return nil
}
