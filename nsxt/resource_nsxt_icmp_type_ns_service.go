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

var icmpProtocolValues = []string{"ICMPv4", "ICMPv6"}

func resourceNsxtIcmpTypeNsService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIcmpTypeNsServiceCreate,
		Read:   resourceNsxtIcmpTypeNsServiceRead,
		Update: resourceNsxtIcmpTypeNsServiceUpdate,
		Delete: resourceNsxtIcmpTypeNsServiceDelete,
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
			"icmp_code": {
				Type:         schema.TypeInt,
				Description:  "ICMP message code",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 255),
			},
			"icmp_type": {
				Type:         schema.TypeInt,
				Description:  "ICMP message type",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 255),
			},
			"protocol": {
				Type:         schema.TypeString,
				Description:  "Version of ICMP protocol (ICMPv4/ICMPv6)",
				Required:     true,
				ValidateFunc: validation.StringInSlice(icmpProtocolValues, false),
			},
		},
	}
}

func resourceNsxtIcmpTypeNsServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	icmpCode := int64(d.Get("icmp_code").(int))
	icmpType := int64(d.Get("icmp_type").(int))
	protocol := d.Get("protocol").(string)

	nsService := manager.IcmpTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
		},
		NsserviceElement: manager.IcmpTypeNsServiceEntry{
			ResourceType: "ICMPTypeNSService",
			IcmpCode:     icmpCode,
			IcmpType:     icmpType,
			Protocol:     protocol,
		},
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.CreateIcmpTypeNSService(nsxClient.Context, nsService)
	if resp != nil && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during NsService create: %v", resp.StatusCode)
	}
	if err != nil {
		return fmt.Errorf("Error during NsService create: %v", err)
	}

	d.SetId(nsService.Id)
	return resourceNsxtIcmpTypeNsServiceRead(d, m)
}

func resourceNsxtIcmpTypeNsServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.ReadIcmpTypeNSService(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during NsService read: %v", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsService %s not found", id)
		d.SetId("")
		return nil
	}

	nsserviceElement := nsService.NsserviceElement

	d.Set("revision", nsService.Revision)
	d.Set("description", nsService.Description)
	d.Set("display_name", nsService.DisplayName)
	setTagsInSchema(d, nsService.Tags)
	d.Set("default_service", nsService.DefaultService)
	d.Set("icmp_type", nsserviceElement.IcmpType)
	d.Set("icmp_code", nsserviceElement.IcmpCode)
	d.Set("protocol", nsserviceElement.Protocol)

	return nil
}

func resourceNsxtIcmpTypeNsServiceUpdate(d *schema.ResourceData, m interface{}) error {
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
	icmpCode := int64(d.Get("icmp_code").(int))
	icmpType := int64(d.Get("icmp_type").(int))
	protocol := d.Get("protocol").(string)
	revision := int64(d.Get("revision").(int))

	nsService := manager.IcmpTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			Revision:    revision,
		},
		NsserviceElement: manager.IcmpTypeNsServiceEntry{
			ResourceType: "ICMPTypeNSService",
			IcmpCode:     icmpCode,
			IcmpType:     icmpType,
			Protocol:     protocol,
		},
	}

	_, resp, err := nsxClient.GroupingObjectsApi.UpdateIcmpTypeNSService(nsxClient.Context, id, nsService)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during NsService update: %v %v", err, resp)
	}

	return resourceNsxtIcmpTypeNsServiceRead(d, m)
}

func resourceNsxtIcmpTypeNsServiceDelete(d *schema.ResourceData, m interface{}) error {
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
