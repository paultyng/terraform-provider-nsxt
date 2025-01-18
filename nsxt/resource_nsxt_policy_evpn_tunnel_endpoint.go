// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyEvpnTunnelEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyEvpnTunnelEndpointCreate,
		Read:   resourceNsxtPolicyEvpnTunnelEndpointRead,
		Update: resourceNsxtPolicyEvpnTunnelEndpointUpdate,
		Delete: resourceNsxtPolicyEvpnTunnelEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyEvpnTunnelEndpointImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":                  getNsxIDSchema(),
			"path":                    getPathSchema(),
			"display_name":            getDisplayNameSchema(),
			"description":             getDescriptionSchema(),
			"revision":                getRevisionSchema(),
			"tag":                     getTagsSchema(),
			"external_interface_path": getPolicyPathSchema(true, true, "Path External Interfaceon Tier0 Gateway"),
			"edge_node_path":          getPolicyPathSchema(true, false, "Edge Node Path"),
			"local_address": {
				Type:         schema.TypeString,
				Description:  "Local IPv4 IP address",
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"mtu": getMtuSchema(),
			"locale_service_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Gateway Locale Service on NSX",
				Computed:    true,
			},
			"gateway_id": {
				Type:        schema.TypeString,
				Description: "Id of associated Tier0 Gateway on NSX",
				Computed:    true,
			},
		},
	}
}

func policyEvpnTunnelEndpointPatch(d *schema.ResourceData, m interface{}, gwID string, localeServiceID string, id string) error {
	connector := getPolicyConnector(m)

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	edgePath := d.Get("edge_node_path").(string)
	mtu := int64(d.Get("mtu").(int))
	localAddress := d.Get("local_address").(string)
	localAddressList := []string{localAddress}

	obj := model.EvpnTunnelEndpointConfig{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		EdgePath:       &edgePath,
		LocalAddresses: localAddressList,
	}

	if mtu > 0 {
		obj.Mtu = &mtu
	}

	client := locale_services.NewEvpnTunnelEndpointsClient(connector)
	return client.Patch(gwID, localeServiceID, id, obj)
}

func resourceNsxtPolicyEvpnTunnelEndpointExists(connector client.Connector, gwID string, localeServiceID string, id string) (bool, error) {
	client := locale_services.NewEvpnTunnelEndpointsClient(connector)
	_, err := client.Get(gwID, localeServiceID, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return true, err
}

func resourceNsxtPolicyEvpnTunnelEndpointCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	interfacePath := d.Get("external_interface_path").(string)
	isT0, gwID, localeServiceID, _ := parseGatewayInterfacePolicyPath(interfacePath)
	if !isT0 || gwID == "" {
		return fmt.Errorf("Tier0 External Interface path expected, got %s", interfacePath)
	}

	if id == "" {
		id = newUUID()
	} else {
		exists, err := resourceNsxtPolicyEvpnTunnelEndpointExists(connector, gwID, localeServiceID, id)
		if exists {
			return fmt.Errorf("EVPN Tunnel Endpoint with nsx_id '%s' already exists on Gateway %s and Locale Service %s", id, gwID, localeServiceID)
		} else if err != nil {
			return err
		}

	}

	err := policyEvpnTunnelEndpointPatch(d, m, gwID, localeServiceID, id)
	if err != nil {
		return handleCreateError("EVPN Tunnel Endpoint", id, err)
	}

	d.Set("nsx_id", id)
	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)
	d.Set("external_interface_path", interfacePath)
	d.SetId(id)

	return resourceNsxtPolicyEvpnTunnelEndpointRead(d, m)
}

func resourceNsxtPolicyEvpnTunnelEndpointRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)

	client := locale_services.NewEvpnTunnelEndpointsClient(connector)
	obj, err := client.Get(gwID, localeServiceID, id)
	if err != nil {
		return handleReadError(d, "EVPN Tunnel Endpoint", id, err)
	}

	d.Set("edge_node_path", obj.EdgePath)
	if len(obj.LocalAddresses) > 0 {
		d.Set("local_address", obj.LocalAddresses[0])
	} else {
		d.Set("local_address", "")
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("mtu", obj.Mtu)

	return nil
}

func resourceNsxtPolicyEvpnTunnelEndpointUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || gwID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	err := policyEvpnTunnelEndpointPatch(d, m, gwID, localeServiceID, id)
	if err != nil {
		return handleCreateError("EVPN Tunnel Endpoint", id, err)
	}

	return resourceNsxtPolicyEvpnTunnelEndpointRead(d, m)
}

func resourceNsxtPolicyEvpnTunnelEndpointDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	gwID := d.Get("gateway_id").(string)
	localeServiceID := d.Get("locale_service_id").(string)
	if id == "" || gwID == "" || localeServiceID == "" {
		return fmt.Errorf("Error obtaining Tier0 id or Locale Service id")
	}

	client := locale_services.NewEvpnTunnelEndpointsClient(connector)
	err := client.Delete(gwID, localeServiceID, id)

	if err != nil {
		return handleDeleteError("EVPN Tunnel Endpoint", id, err)
	}

	return nil
}

func resourceNsxtPolicyEvpnTunnelEndpointImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 4 {
		return nil, fmt.Errorf("Please provide gateway-id/locale-service-id/interface-id/endpoint-id as an input")
	}

	gwID := s[0]
	localeServiceID := s[1]
	interfaceID := s[2]
	id := s[3]

	d.Set("gateway_id", gwID)
	d.Set("locale_service_id", localeServiceID)
	d.Set("external_interface_path", fmt.Sprintf("/infra/tier-0s/%s/locale-services/%s/interfaces/%s", gwID, localeServiceID, interfaceID))
	d.SetId(id)

	return []*schema.ResourceData{d}, nil
}
