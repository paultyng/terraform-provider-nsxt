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
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

var lbServiceLogLevels = []string{"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL", "ALERT", "EMERGENCY"}
var lbServiceSizes = []string{"SMALL", "MEDIUM", "LARGE"}

func resourceNsxtLbService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbServiceCreate,
		Read:   resourceNsxtLbServiceRead,
		Update: resourceNsxtLbServiceUpdate,
		Delete: resourceNsxtLbServiceDelete,
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
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether the load balancer service is enabled",
				Optional:    true,
				Default:     true,
			},
			"error_log_level": {
				Type:         schema.TypeString,
				Description:  "Load balancer engine writes information about encountered issues of different severity levels to the error log. This setting is used to define the severity level of the error log",
				Optional:     true,
				Default:      "INFO",
				ValidateFunc: validation.StringInSlice(lbServiceLogLevels, false),
			},
			"size": {
				Type:         schema.TypeString,
				Description:  "Size of load balancer service",
				Optional:     true,
				Default:      "SMALL",
				ValidateFunc: validation.StringInSlice(lbServiceSizes, false),
			},
			// TODO: LB service creation will error out on NSX if logical Tier1 router is not
			// attached to Tier0 or Centralized Service Port. Consider dummy port attribute here
			// to enforce this dependency.
			"logical_router_id": {
				Type:        schema.TypeString,
				Description: "Logical Tier1 Router to which the Load Balancer is to be attached",
				Required:    true,
			},
			"virtual_server_ids": {
				Type:        schema.TypeSet,
				Description: "Virtual servers associated with this Load Balancer",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceNsxtLbServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	enabled := d.Get("enabled").(bool)
	errorLogLevel := d.Get("error_log_level").(string)
	size := d.Get("size").(string)
	virtualServerIds := getStringListFromSchemaSet(d, "virtual_server_ids")

	lbService := loadbalancer.LbService{
		Description:      description,
		DisplayName:      displayName,
		Tags:             tags,
		Attachment:       makeResourceReference("LogicalRouter", logicalRouterID),
		Enabled:          enabled,
		ErrorLogLevel:    errorLogLevel,
		Size:             size,
		VirtualServerIds: virtualServerIds,
	}

	lbService, resp, err := nsxClient.ServicesApi.CreateLoadBalancerService(nsxClient.Context, lbService)

	if err != nil {
		return fmt.Errorf("Error during LbService create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbService create: %v", resp.StatusCode)
	}
	d.SetId(lbService.Id)

	return resourceNsxtLbServiceRead(d, m)
}

func resourceNsxtLbServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbService, resp, err := nsxClient.ServicesApi.ReadLoadBalancerService(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbService %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbService read: %v", err)
	}

	d.Set("revision", lbService.Revision)
	d.Set("description", lbService.Description)
	d.Set("display_name", lbService.DisplayName)
	setTagsInSchema(d, lbService.Tags)
	if lbService.Attachment != nil {
		if lbService.Attachment.TargetType != "LogicalRouter" {
			return fmt.Errorf("Error during LbService attachment read: attachment type %s is not supported", lbService.Attachment.TargetType)
		}
		d.Set("logical_router_id", lbService.Attachment.TargetId)
	}
	d.Set("enabled", lbService.Enabled)
	d.Set("error_log_level", lbService.ErrorLogLevel)
	d.Set("size", lbService.Size)
	d.Set("virtual_server_ids", lbService.VirtualServerIds)

	return nil
}

func resourceNsxtLbServiceUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	revision := int32(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	enabled := d.Get("enabled").(bool)
	errorLogLevel := d.Get("error_log_level").(string)
	size := d.Get("size").(string)
	virtualServerIds := getStringListFromSchemaSet(d, "virtual_server_ids")
	lbService := loadbalancer.LbService{
		Revision:         revision,
		Description:      description,
		DisplayName:      displayName,
		Tags:             tags,
		Attachment:       makeResourceReference("LogicalRouter", logicalRouterID),
		Enabled:          enabled,
		ErrorLogLevel:    errorLogLevel,
		Size:             size,
		VirtualServerIds: virtualServerIds,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerService(nsxClient.Context, id, lbService)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbService update: %v", err)
	}

	return resourceNsxtLbServiceRead(d, m)
}

func resourceNsxtLbServiceDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerService(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbService delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbService %s not found", id)
		d.SetId("")
	}
	return nil
}
