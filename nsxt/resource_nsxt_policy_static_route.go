// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	tier0s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_0s"
	tier1s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyStaticRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyStaticRouteCreate,
		Read:   resourceNsxtPolicyStaticRouteRead,
		Update: resourceNsxtPolicyStaticRouteUpdate,
		Delete: resourceNsxtPolicyStaticRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyStaticRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"gateway_path": getPolicyGatewayPathSchema(),
			"network": {
				Type:         schema.TypeString,
				Description:  "Network address in CIDR format",
				Required:     true,
				ValidateFunc: validateCidr(),
			},
			"next_hop": {
				Type:        schema.TypeList,
				Description: "Next hop routes for network",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"admin_distance": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Cost associated with next hop route",
							ValidateFunc: validation.IntBetween(1, 255),
							Default:      1,
						},
						"ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Next hop gateway IP address",
							ValidateFunc: validateSingleIP(),
						},
						"interface": {
							// NOTE: this is called 'scope' in the golang vapi struct
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Interface path associated with current route",
							ValidateFunc: validatePolicyPath(),
						},
					},
				},
			},
		},
	}
}

func patchNsxtPolicyStaticRoute(sessionContext utl.SessionContext, connector client.Connector, gwID string, route model.StaticRoutes, isT0 bool) error {
	if isT0 {
		routeClient := tier0s.NewStaticRoutesClient(sessionContext, connector)
		if routeClient == nil {
			return policyResourceNotSupportedError()
		}
		return routeClient.Patch(gwID, *route.Id, route)
	}
	routeClient := tier1s.NewStaticRoutesClient(sessionContext, connector)
	if routeClient == nil {
		return policyResourceNotSupportedError()
	}
	return routeClient.Patch(gwID, *route.Id, route)
}

func deleteNsxtPolicyStaticRoute(sessionContext utl.SessionContext, connector client.Connector, gwID string, isT0 bool, routeID string) error {
	if isT0 {
		routeClient := tier0s.NewStaticRoutesClient(sessionContext, connector)
		if routeClient == nil {
			return policyResourceNotSupportedError()
		}
		return routeClient.Delete(gwID, routeID)
	}
	routeClient := tier1s.NewStaticRoutesClient(sessionContext, connector)
	if routeClient == nil {
		return policyResourceNotSupportedError()
	}
	return routeClient.Delete(gwID, routeID)
}

func getNsxtPolicyStaticRouteByID(sessionContext utl.SessionContext, connector client.Connector, gwID string, isT0 bool, routeID string) (model.StaticRoutes, error) {
	if isT0 {
		routeClient := tier0s.NewStaticRoutesClient(sessionContext, connector)
		if routeClient == nil {
			return model.StaticRoutes{}, policyResourceNotSupportedError()
		}
		return routeClient.Get(gwID, routeID)
	}
	routeClient := tier1s.NewStaticRoutesClient(sessionContext, connector)
	if routeClient == nil {
		return model.StaticRoutes{}, policyResourceNotSupportedError()
	}
	return routeClient.Get(gwID, routeID)
}

func resourceNsxtPolicyStaticRouteCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not a valid")
	}
	context := getSessionContext(d, m)
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := getNsxtPolicyStaticRouteByID(context, connector, gwID, isT0, id)
		if err == nil {
			return fmt.Errorf("Static Route with nsx_id '%s' already exists", id)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	network := d.Get("network").(string)

	var nextHopsStructs []model.RouterNexthop
	nextHops := d.Get("next_hop").([]interface{})
	for _, nextHop := range nextHops {
		nextHopMap := nextHop.(map[string]interface{})
		distance := int64(nextHopMap["admin_distance"].(int))
		ip := nextHopMap["ip_address"].(string)
		scope := nextHopMap["interface"].(string)
		var scopeList []string
		if scope != "" {
			scopeList = append(scopeList, scope)
		}
		hopStruct := model.RouterNexthop{
			AdminDistance: &distance,
			Scope:         scopeList,
		}

		if len(ip) > 0 {
			hopStruct.IpAddress = &ip
		}
		nextHopsStructs = append(nextHopsStructs, hopStruct)
	}

	routeStruct := model.StaticRoutes{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Network:     &network,
	}

	if len(nextHopsStructs) > 0 {
		routeStruct.NextHops = nextHopsStructs
	}

	log.Printf("[INFO] Creating Static Route with ID %s", id)
	err := patchNsxtPolicyStaticRoute(getSessionContext(d, m), connector, gwID, routeStruct, isT0)
	if err != nil {
		return handleCreateError("Static Route", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyStaticRouteRead(d, m)
}

func resourceNsxtPolicyStaticRouteRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Static Route ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not a valid")
	}
	context := getSessionContext(d, m)
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	obj, err := getNsxtPolicyStaticRouteByID(context, connector, gwID, isT0, id)
	if err != nil {
		return handleReadError(d, "Static Route", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("network", obj.Network)

	var nextHopMaps []map[string]interface{}
	for _, nextHop := range obj.NextHops {
		nextHopMap := make(map[string]interface{})

		iface := ""
		scope := nextHop.Scope
		if len(scope) > 0 {
			iface = scope[0]
		}
		nextHopMap["interface"] = iface
		if nextHop.IpAddress != nil {
			nextHopMap["ip_address"] = *nextHop.IpAddress
		}
		if nextHop.AdminDistance != nil {
			nextHopMap["admin_distance"] = nextHop.AdminDistance
		}

		nextHopMaps = append(nextHopMaps, nextHopMap)
	}

	d.Set("next_hop", nextHopMaps)

	d.SetId(id)

	return nil
}

func resourceNsxtPolicyStaticRouteUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Static Route ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	context := getSessionContext(d, m)
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	network := d.Get("network").(string)

	var nextHopsStructs []model.RouterNexthop
	nextHops := d.Get("next_hop").([]interface{})
	for _, nextHop := range nextHops {
		nextHopMap := nextHop.(map[string]interface{})
		distance := int64(nextHopMap["admin_distance"].(int))
		ip := nextHopMap["ip_address"].(string)
		scope := nextHopMap["interface"].(string)
		var scopeList []string
		if scope != "" {
			scopeList = append(scopeList, scope)
		}
		hopStruct := model.RouterNexthop{
			AdminDistance: &distance,
			Scope:         scopeList,
		}
		if len(ip) > 0 {
			hopStruct.IpAddress = &ip
		}
		nextHopsStructs = append(nextHopsStructs, hopStruct)
	}

	routeStruct := model.StaticRoutes{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Network:     &network,
	}

	if len(nextHopsStructs) > 0 {
		routeStruct.NextHops = nextHopsStructs
	}

	log.Printf("[INFO] Updating Static Route with ID %s", id)
	err := patchNsxtPolicyStaticRoute(context, connector, gwID, routeStruct, isT0)
	if err != nil {
		return handleUpdateError("Static Route", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyStaticRouteRead(d, m)
}

func resourceNsxtPolicyStaticRouteDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Static Route ID")
	}

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	context := getSessionContext(d, m)
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	err := deleteNsxtPolicyStaticRoute(context, getPolicyConnector(m), gwID, isT0, id)
	if err != nil {
		return handleDeleteError("Static Route", id, err)
	}

	return nil
}

func resourceNsxtPolicyStaticRouteImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		gwPath, err := getParameterFromPolicyPath("", "/static-routes", importID)
		if err != nil {
			return nil, err
		}
		d.Set("gateway_path", gwPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <gateway-id>/<static-route-id> as an input")
	}

	gwID := s[0]
	connector := getPolicyConnector(m)
	t0Client := infra.NewTier0sClient(getSessionContext(d, m), connector)
	if t0Client == nil {
		return nil, policyResourceNotSupportedError()
	}
	t0gw, err := t0Client.Get(gwID)
	if err != nil {
		if !isNotFoundError(err) {
			return nil, err
		}
		t1Client := infra.NewTier1sClient(getSessionContext(d, m), connector)
		if t1Client == nil {
			return nil, policyResourceNotSupportedError()
		}
		t1gw, err := t1Client.Get(gwID)
		if err != nil {
			return nil, err
		}
		d.Set("gateway_path", t1gw.Path)
	} else {
		d.Set("gateway_path", t0gw.Path)
	}

	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
