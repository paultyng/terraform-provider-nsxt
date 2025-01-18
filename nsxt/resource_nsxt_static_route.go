// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtStaticRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtStaticRouteCreate,
		Read:   resourceNsxtStaticRouteRead,
		Update: resourceNsxtStaticRouteUpdate,
		Delete: resourceNsxtStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtStaticRouteImport,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"logical_router_id": {
				Type:        schema.TypeString,
				Description: "Logical router id",
				Required:    true,
			},
			"network": {
				Type:         schema.TypeString,
				Description:  "CIDR",
				Required:     true,
				ValidateFunc: validation.IsCIDRNetwork(0, 32),
			},
			"next_hop": getNextHopsSchema(),
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
		},
	}
}

func getNextHopsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"administrative_distance": {
					Type:        schema.TypeInt,
					Description: "Administrative Distance for the next hop IP",
					Optional:    true,
				},
				"bfd_enabled": {
					Type:        schema.TypeBool,
					Description: "Status of bfd for this next hop where bfdEnabled = true indicate bfd is enabled for this next hop and bfdEnabled = false indicate bfd peer is disabled or not configured for this next hop.",
					Computed:    true,
				},
				"blackhole_action": {
					Type:        schema.TypeString,
					Description: "Action to be taken on matching packets for NULL routes",
					Computed:    true,
				},
				"ip_address": {
					Type:         schema.TypeString,
					Description:  "Next Hop IP",
					Optional:     true,
					ValidateFunc: validateSingleIP(),
				},
				"logical_router_port_id": {
					Type:        schema.TypeString,
					Description: "Logical router port id",
					Optional:    true,
				},
			},
		},
	}
}

func getNextHopsFromSchema(d *schema.ResourceData) []manager.StaticRouteNextHop {
	hops := d.Get("next_hop").([]interface{})
	var nextHopsList []manager.StaticRouteNextHop
	for _, hop := range hops {
		data := hop.(map[string]interface{})
		administrativeDistance := int64(data["administrative_distance"].(int))
		bfdEnabled := data["bfd_enabled"].(bool)
		blackholeAction := data["blackhole_action"].(string)
		ipAddress := data["ip_address"].(string)
		logicalRouterPortID := &common.ResourceReference{
			TargetType: "LogicalPort",
			TargetId:   data["logical_router_port_id"].(string),
		}
		elem := manager.StaticRouteNextHop{
			AdministrativeDistance: administrativeDistance,
			BfdEnabled:             bfdEnabled,
			BlackholeAction:        blackholeAction,
			IpAddress:              ipAddress,
			LogicalRouterPortId:    logicalRouterPortID,
		}
		nextHopsList = append(nextHopsList, elem)
	}
	return nextHopsList
}

func setNextHopsInSchema(d *schema.ResourceData, nextHops []manager.StaticRouteNextHop) error {
	var nextHopsList []map[string]interface{}
	for _, staticRouteNextHop := range nextHops {
		elem := make(map[string]interface{})
		elem["administrative_distance"] = staticRouteNextHop.AdministrativeDistance
		elem["bfd_enabled"] = staticRouteNextHop.BfdEnabled
		elem["blackhole_action"] = staticRouteNextHop.BlackholeAction
		elem["ip_address"] = staticRouteNextHop.IpAddress
		if staticRouteNextHop.LogicalRouterPortId != nil {
			elem["logical_router_port_id"] = staticRouteNextHop.LogicalRouterPortId.TargetId
		}
		nextHopsList = append(nextHopsList, elem)
	}
	err := d.Set("next_hop", nextHopsList)
	return err
}

func resourceNsxtStaticRouteCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route creation")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	network := d.Get("network").(string)
	nextHops := getNextHopsFromSchema(d)
	staticRoute := manager.StaticRoute{
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		LogicalRouterId: logicalRouterID,
		Network:         network,
		NextHops:        nextHops,
	}

	staticRoute, resp, err := nsxClient.LogicalRoutingAndServicesApi.AddStaticRoute(nsxClient.Context, logicalRouterID, staticRoute)

	if err != nil {
		return fmt.Errorf("Error during StaticRoute create on router %s: %v", logicalRouterID, err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during StaticRoute create on router %s: %v", logicalRouterID, resp.StatusCode)
	}
	d.SetId(staticRoute.Id)

	return resourceNsxtStaticRouteRead(d, m)
}

func resourceNsxtStaticRouteRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route read")
	}

	staticRoute, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadStaticRoute(nsxClient.Context, logicalRouterID, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] StaticRoute %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during StaticRoute read: %v", err)
	}

	d.Set("revision", staticRoute.Revision)
	d.Set("description", staticRoute.Description)
	d.Set("display_name", staticRoute.DisplayName)
	setTagsInSchema(d, staticRoute.Tags)
	d.Set("logical_router_id", staticRoute.LogicalRouterId)
	d.Set("network", staticRoute.Network)
	err = setNextHopsInSchema(d, staticRoute.NextHops)
	if err != nil {
		return fmt.Errorf("Error during StaticRoute set in schema: %v", err)
	}

	return nil
}

func resourceNsxtStaticRouteUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route update")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	network := d.Get("network").(string)
	nextHops := getNextHopsFromSchema(d)
	staticRoute := manager.StaticRoute{
		Revision:        revision,
		Description:     description,
		DisplayName:     displayName,
		Tags:            tags,
		LogicalRouterId: logicalRouterID,
		Network:         network,
		NextHops:        nextHops,
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateStaticRoute(nsxClient.Context, logicalRouterID, id, staticRoute)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during StaticRoute update: %v", err)
	}

	return resourceNsxtStaticRouteRead(d, m)
}

func resourceNsxtStaticRouteDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalRouterID := d.Get("logical_router_id").(string)
	if logicalRouterID == "" {
		return fmt.Errorf("Error obtaining logical router id during static route deletion")
	}

	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteStaticRoute(nsxClient.Context, logicalRouterID, id)
	if err != nil {
		return fmt.Errorf("Error during StaticRoute delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] StaticRoute %s for router %s not found", id, logicalRouterID)
		d.SetId("")
	}
	return nil
}

func resourceNsxtStaticRouteImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <router-id>/<static-route-id> as an input")
	}
	d.SetId(s[1])
	d.Set("logical_router_id", s[0])
	return []*schema.ResourceData{d}, nil
}
