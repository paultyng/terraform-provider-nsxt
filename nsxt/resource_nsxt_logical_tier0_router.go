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

var highAvailabilityValues = []string{"ACTIVE_ACTIVE", "ACTIVE_STANDBY"}

// TODO: add advanced config
func resourceNsxtLogicalTier0Router() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalTier0RouterCreate,
		Read:   resourceNsxtLogicalTier0RouterRead,
		Update: resourceNsxtLogicalTier0RouterUpdate,
		Delete: resourceNsxtLogicalTier0RouterDelete,
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
			"high_availability_mode": {
				Type:         schema.TypeString,
				Description:  "High availability mode",
				Default:      "ACTIVE_STANDBY",
				Optional:     true,
				ForceNew:     true, // Cannot change the HA mode of a router
				ValidateFunc: validation.StringInSlice(highAvailabilityValues, false),
			},
			"failover_mode": {
				Type:         schema.TypeString,
				Description:  "Failover mode which determines whether the preferred service router instance for given logical router will preempt the peer",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"PREEMPTIVE", "NON_PREEMPTIVE"}, false),
			},
			"firewall_sections": getResourceReferencesSchema(false, true, []string{}, "List of Firewall sections related to the logical router"),
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Description: "Edge Cluster Id",
				Required:    true,
				ForceNew:    true, // Cannot change the edge cluster of existing router
			},
			// TODO - add PreferredEdgeClusterMemberIndex when appropriate data source
			// becomes available
		},
	}
}

func resourceNsxtLogicalTier0RouterCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	highAvailabilityMode := d.Get("high_availability_mode").(string)
	failoverMode := d.Get("failover_mode").(string)
	routerType := "TIER0"
	edgeClusterID := d.Get("edge_cluster_id").(string)
	logicalRouter := manager.LogicalRouter{
		Description:          description,
		DisplayName:          displayName,
		Tags:                 tags,
		RouterType:           routerType,
		EdgeClusterId:        edgeClusterID,
		HighAvailabilityMode: highAvailabilityMode,
		FailoverMode:         failoverMode,
	}
	logicalRouter, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouter(nsxClient.Context, logicalRouter)

	if err != nil {
		return fmt.Errorf("Error during LogicalTier0Router create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned: %d", resp.StatusCode)
	}

	d.SetId(logicalRouter.Id)

	return resourceNsxtLogicalTier0RouterRead(d, m)
}

func resourceNsxtLogicalTier0RouterRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical Tier0 logical router id")
	}

	logicalRouter, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalTier0Router %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalTier0Router read: %v", err)
	}

	d.Set("revision", logicalRouter.Revision)
	d.Set("description", logicalRouter.Description)
	d.Set("display_name", logicalRouter.DisplayName)
	setTagsInSchema(d, logicalRouter.Tags)
	d.Set("edge_cluster_id", logicalRouter.EdgeClusterId)
	d.Set("high_availability_mode", logicalRouter.HighAvailabilityMode)
	d.Set("failover_mode", logicalRouter.FailoverMode)
	err = setResourceReferencesInSchema(d, logicalRouter.FirewallSections, "firewall_sections")
	if err != nil {
		return fmt.Errorf("Error during LogicalTier0Router firewall sections set in schema: %v", err)
	}

	return nil
}

func resourceNsxtLogicalTier0RouterUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical tier0 router id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	highAvailabilityMode := d.Get("high_availability_mode").(string)
	failoverMode := d.Get("failover_mode").(string)
	routerType := "TIER0"
	edgeClusterID := d.Get("edge_cluster_id").(string)
	logicalRouter := manager.LogicalRouter{
		Revision:             revision,
		Description:          description,
		DisplayName:          displayName,
		Tags:                 tags,
		RouterType:           routerType,
		EdgeClusterId:        edgeClusterID,
		HighAvailabilityMode: highAvailabilityMode,
		FailoverMode:         failoverMode,
	}
	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateLogicalRouter(nsxClient.Context, id, logicalRouter)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalTier0Router update error: %v, resp %+v", err, resp)
	}

	return resourceNsxtLogicalTier0RouterRead(d, m)
}

func resourceNsxtLogicalTier0RouterDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical tier0 router id")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouter(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LogicalTier0Router delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalTier0Router %s not found", id)
		d.SetId("")
	}
	return nil
}
