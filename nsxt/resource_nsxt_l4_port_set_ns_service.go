// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var protocolValues = []string{"TCP", "UDP"}

func resourceNsxtL4PortSetNsService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtL4PortSetNsServiceCreate,
		Read:   resourceNsxtL4PortSetNsServiceRead,
		Update: resourceNsxtL4PortSetNsServiceUpdate,
		Delete: resourceNsxtL4PortSetNsServiceDelete,
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
			"default_service": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether this is a default NSServices which can't be modified/deleted",
				Computed:    true,
			},
			"destination_ports": {
				Type:        schema.TypeSet,
				Description: "Set of destination ports",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
				Optional: true,
			},
			"source_ports": {
				Type:        schema.TypeSet,
				Description: "Set of source ports",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePortRange(),
				},
				Optional: true,
			},
			"protocol": {
				Type:         schema.TypeString,
				Description:  "L4 Protocol",
				Required:     true,
				ValidateFunc: validation.StringInSlice(protocolValues, false),
			},
		},
	}
}

func resourceNsxtL4PortSetNsServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	l4Protocol := d.Get("protocol").(string)
	sourcePorts := getStringListFromSchemaSet(d, "source_ports")
	destinationPorts := getStringListFromSchemaSet(d, "destination_ports")

	nsService := manager.L4PortSetNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
		},
		NsserviceElement: manager.L4PortSetNsServiceEntry{
			ResourceType:     "L4PortSetNSService",
			L4Protocol:       l4Protocol,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
		},
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.CreateL4PortSetNSService(nsxClient.Context, nsService)

	if err != nil {
		return fmt.Errorf("Error during NsService create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during NsService create: %v", resp.StatusCode)
	}
	d.SetId(nsService.Id)
	return resourceNsxtL4PortSetNsServiceRead(d, m)
}

func resourceNsxtL4PortSetNsServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.ReadL4PortSetNSService(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsService %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during NsService read: %v", err)
	}

	nsserviceElement := nsService.NsserviceElement

	d.Set("revision", nsService.Revision)
	d.Set("description", nsService.Description)
	d.Set("display_name", nsService.DisplayName)
	setTagsInSchema(d, nsService.Tags)
	d.Set("default_service", nsService.DefaultService)
	d.Set("protocol", nsserviceElement.L4Protocol)
	d.Set("destination_ports", nsserviceElement.DestinationPorts)
	d.Set("source_ports", nsserviceElement.SourcePorts)

	return nil
}

func resourceNsxtL4PortSetNsServiceUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	l4Protocol := d.Get("protocol").(string)
	sourcePorts := getStringListFromSchemaSet(d, "source_ports")
	destinationPorts := getStringListFromSchemaSet(d, "destination_ports")
	revision := int64(d.Get("revision").(int))

	nsService := manager.L4PortSetNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			Revision:    revision,
		},
		NsserviceElement: manager.L4PortSetNsServiceEntry{
			ResourceType:     "L4PortSetNSService",
			L4Protocol:       l4Protocol,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
		},
	}

	_, resp, err := nsxClient.GroupingObjectsApi.UpdateL4PortSetNSService(nsxClient.Context, id, nsService)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during NsService update: %v %v", err, resp)
	}

	return resourceNsxtL4PortSetNsServiceRead(d, m)
}

func resourceNsxtL4PortSetNsServiceDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	localVarOptionals := make(map[string]interface{})
	localVarOptionals["force"] = true
	resp, err := nsxClient.GroupingObjectsApi.DeleteNSService(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during NsService delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsService %s not found", id)
		d.SetId("")
	}
	return nil
}
