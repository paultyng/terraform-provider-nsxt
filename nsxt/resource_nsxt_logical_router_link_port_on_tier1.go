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

func resourceNsxtLogicalRouterLinkPortOnTier1() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalRouterLinkPortOnTier1Create,
		Read:   resourceNsxtLogicalRouterLinkPortOnTier1Read,
		Update: resourceNsxtLogicalRouterLinkPortOnTier1Update,
		Delete: resourceNsxtLogicalRouterLinkPortOnTier1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
			"logical_router_id": {
				Type:        schema.TypeString,
				Description: "Identifier for logical router on which this port is created",
				Required:    true,
				ForceNew:    true,
			},
			"linked_logical_router_port_id": {
				Type:        schema.TypeString,
				Description: "Identifier for port on logical router to connect to",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceNsxtLogicalRouterLinkPortOnTier1Create(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	linkedLogicalRouterPortID := d.Get("linked_logical_router_port_id").(string)
	logicalRouterLinkPort := manager.LogicalRouterLinkPortOnTier1{
		Description:               description,
		DisplayName:               displayName,
		Tags:                      tags,
		LogicalRouterId:           logicalRouterID,
		LinkedLogicalRouterPortId: makeResourceReference("LogicalPort", linkedLogicalRouterPortID),
	}

	logicalRouterLinkPort, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouterLinkPortOnTier1(nsxClient.Context, logicalRouterLinkPort)
	if resp != nil && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during LogicalRouterLinkPortOnTier1 create: %v", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterLinkPortOnTier1 create: %v", err)
	}

	d.SetId(logicalRouterLinkPort.Id)

	return resourceNsxtLogicalRouterLinkPortOnTier1Read(d, m)
}

func resourceNsxtLogicalRouterLinkPortOnTier1Read(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router link port on tier1 id")
	}

	logicalRouterLinkPort, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterLinkPortOnTier1(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalRouterLinkPortOnTier1 %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterLinkPortOnTier1 read: %v", err)
	}

	d.Set("revision", logicalRouterLinkPort.Revision)
	d.Set("description", logicalRouterLinkPort.Description)
	d.Set("display_name", logicalRouterLinkPort.DisplayName)
	setTagsInSchema(d, logicalRouterLinkPort.Tags)
	d.Set("logical_router_id", logicalRouterLinkPort.LogicalRouterId)
	d.Set("linked_logical_router_port_id", logicalRouterLinkPort.LinkedLogicalRouterPortId.TargetId)

	return nil
}

func resourceNsxtLogicalRouterLinkPortOnTier1Update(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router link port on tier1 id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	linkedLogicalRouterPortID := d.Get("linked_logical_router_port_id").(string)
	logicalRouterLinkPort := manager.LogicalRouterLinkPortOnTier1{
		Revision:                  revision,
		Description:               description,
		DisplayName:               displayName,
		Tags:                      tags,
		LogicalRouterId:           logicalRouterID,
		LinkedLogicalRouterPortId: makeResourceReference("LogicalPort", linkedLogicalRouterPortID),
		ResourceType:              "LogicalRouterLinkPortOnTIER1",
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateLogicalRouterLinkPortOnTier1(nsxClient.Context, id, logicalRouterLinkPort)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalRouterLinkPortOnTier1 %v update: %v (%+v)", id, err, resp)
	}

	return resourceNsxtLogicalRouterLinkPortOnTier1Read(d, m)
}

func resourceNsxtLogicalRouterLinkPortOnTier1Delete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router link port on tier1 id")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouterPort(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterLinkPortOnTier1 delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalRouterLinkPortOnTier1 %s not found", id)
		d.SetId("")
	}

	return nil
}
