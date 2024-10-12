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

func resourceNsxtEtherTypeNsService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtEtherTypeNsServiceCreate,
		Read:   resourceNsxtEtherTypeNsServiceRead,
		Update: resourceNsxtEtherTypeNsServiceUpdate,
		Delete: resourceNsxtEtherTypeNsServiceDelete,
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
			"ether_type": {
				Type:        schema.TypeInt,
				Description: "Type of the encapsulated protocol",
				Required:    true,
			},
		},
	}
}

func resourceNsxtEtherTypeNsServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	etherType := int64(d.Get("ether_type").(int))

	nsService := manager.EtherTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
		},
		NsserviceElement: manager.EtherTypeNsServiceEntry{
			ResourceType: "EtherTypeNSService",
			EtherType:    etherType,
		},
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.CreateEtherTypeNSService(nsxClient.Context, nsService)

	if err != nil {
		return fmt.Errorf("Error during NsService create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during NsService create: %v", resp.StatusCode)
	}
	d.SetId(nsService.Id)
	return resourceNsxtEtherTypeNsServiceRead(d, m)
}

func resourceNsxtEtherTypeNsServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.ReadEtherTypeNSService(nsxClient.Context, id)
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
	d.Set("ether_type", nsserviceElement.EtherType)

	return nil
}

func resourceNsxtEtherTypeNsServiceUpdate(d *schema.ResourceData, m interface{}) error {
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
	revision := int64(d.Get("revision").(int))
	etherType := int64(d.Get("ether_type").(int))

	nsService := manager.EtherTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			Revision:    revision,
		},
		NsserviceElement: manager.EtherTypeNsServiceEntry{
			ResourceType: "EtherTypeNSService",
			EtherType:    etherType,
		},
	}

	_, resp, err := nsxClient.GroupingObjectsApi.UpdateEtherTypeNSService(nsxClient.Context, id, nsService)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during NsService update: %v %v", err, resp)
	}

	return resourceNsxtEtherTypeNsServiceRead(d, m)
}

func resourceNsxtEtherTypeNsServiceDelete(d *schema.ResourceData, m interface{}) error {
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
