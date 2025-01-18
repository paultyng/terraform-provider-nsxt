// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func resourceNsxtEdgeTransportNodeRTEP() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtEdgeTransportNodeRTEPCreate,
		Read:   resourceNsxtEdgeTransportNodeRTEPRead,
		Update: resourceNsxtEdgeTransportNodeRTEPUpdate,
		Delete: resourceNsxtEdgeTransportNodeRTEPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"edge_id": {
				Type:        schema.TypeString,
				Description: "Edge ID to associate with remote tunnel endpoint.",
				Required:    true,
				ForceNew:    true,
			},
			"host_switch_name": {
				Type:        schema.TypeString,
				Description: "The host switch name to be used for the remote tunnel endpoint",
				Required:    true,
			},
			"ip_assignment": getIPAssignmentSchema(true),
			"named_teaming_policy": {
				Type:        schema.TypeString,
				Description: "The named teaming policy to be used by the remote tunnel endpoint",
				Optional:    true,
			},
			"rtep_vlan": {
				Type:         schema.TypeInt,
				Description:  "VLAN id for remote tunnel endpoint",
				Required:     true,
				ValidateFunc: validation.IntBetween(0, 4094),
			},
		},
	}
}

func resourceNsxtEdgeTransportNodeRTEPCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	id := d.Get("edge_id").(string)

	obj, err := client.Get(id)
	if err != nil {
		return handleCreateError("TransportNodeRTEP", id, err)
	}

	if obj.RemoteTunnelEndpoint != nil {
		return fmt.Errorf("remote tunnel endpoint for Edge appliance %s already exists", id)
	}
	hostSwitchName := d.Get("host_switch_name").(string)
	namedTeamingPolicy := d.Get("named_teaming_policy").(string)
	rtepVlan := int64(d.Get("rtep_vlan").(int))

	ipAssignment, err := getIPAssignmentFromSchema(d.Get("ip_assignment").([]interface{}))
	if err != nil {
		return err
	}

	rtep := model.TransportNodeRemoteTunnelEndpointConfig{
		HostSwitchName:   &hostSwitchName,
		IpAssignmentSpec: ipAssignment,
		RtepVlan:         &rtepVlan,
	}
	if namedTeamingPolicy != "" {
		rtep.NamedTeamingPolicy = &namedTeamingPolicy
	}
	obj.RemoteTunnelEndpoint = &rtep
	_, err = client.Update(id, obj, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleCreateError("TransportNodeRTEP", id, err)
	}

	d.SetId(id)

	return resourceNsxtEdgeTransportNodeRTEPRead(d, m)
}

func resourceNsxtEdgeTransportNodeRTEPRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	id := d.Get("edge_id").(string)

	obj, err := client.Get(id)
	if err != nil {
		return err
	}

	if obj.RemoteTunnelEndpoint == nil {
		return errors.NotFound{}
	}

	d.Set("host_switch_name", obj.RemoteTunnelEndpoint.HostSwitchName)
	ipAssignment, err := setIPAssignmentInSchema(obj.RemoteTunnelEndpoint.IpAssignmentSpec)
	if err != nil {
		return handleReadError(d, "TransportNodeRTEP", id, err)
	}
	d.Set("ip_assignment", ipAssignment)
	d.Set("named_teaming_policy", obj.RemoteTunnelEndpoint.NamedTeamingPolicy)
	d.Set("rtep_vlan", obj.RemoteTunnelEndpoint.RtepVlan)

	return nil
}

func resourceNsxtEdgeTransportNodeRTEPUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	id := d.Get("edge_id").(string)

	obj, err := client.Get(id)
	if err != nil {
		return err
	}

	if obj.RemoteTunnelEndpoint == nil {
		return errors.NotFound{}
	}

	hostSwitchName := d.Get("host_switch_name").(string)
	namedTeamingPolicy := d.Get("named_teaming_policy").(string)
	rtepVlan := int64(d.Get("rtep_vlan").(int))

	ipAssignment, err := getIPAssignmentFromSchema(d.Get("ip_assignment").([]interface{}))
	if err != nil {
		return err
	}

	rtep := model.TransportNodeRemoteTunnelEndpointConfig{
		HostSwitchName:     &hostSwitchName,
		IpAssignmentSpec:   ipAssignment,
		NamedTeamingPolicy: &namedTeamingPolicy,
		RtepVlan:           &rtepVlan,
	}
	obj.RemoteTunnelEndpoint = &rtep
	_, err = client.Update(id, obj, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleUpdateError("TransportNodeRTEP", id, err)
	}

	return resourceNsxtEdgeTransportNodeRTEPRead(d, m)
}

func resourceNsxtEdgeTransportNodeRTEPDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	id := d.Get("edge_id").(string)

	obj, err := client.Get(id)
	if err != nil {
		return err
	}

	obj.RemoteTunnelEndpoint = nil

	_, err = client.Update(id, obj, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleDeleteError("TransportNodeRTEP", id, err)
	}

	return nil
}
