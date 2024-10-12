// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	t0_l2vpn_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/l2vpn_services"
	t0_l2vpn_nested_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services/l2vpn_services"
	t1_l2vpn_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/l2vpn_services"
	t1_l2vpn_nested_services "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/locale_services/l2vpn_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var L2VpnSessionTCPSegmentClampingDirection = []string{
	model.L2TcpMaxSegmentSizeClamping_DIRECTION_NONE,
	model.L2TcpMaxSegmentSizeClamping_DIRECTION_BOTH,
}

var L2VpnTunnelEncapsulationProtocal = []string{
	model.L2VPNTunnelEncapsulation_PROTOCOL_GRE,
}

func resourceNsxtPolicyL2VPNSession() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyL2VPNSessionCreate,
		Read:   resourceNsxtPolicyL2VPNSessionRead,
		Update: resourceNsxtPolicyL2VPNSessionUpdate,
		Delete: resourceNsxtPolicyL2VPNSessionDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVpnSessionImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"service_path": getPolicyPathSchema(true, true, "Policy path for L2 VPN service"),
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable/Disable IPSec VPN session.",
				Optional:    true,
				Default:     true,
			},
			"transport_tunnels": {
				Type:        schema.TypeList,
				Description: "List of transport tunnels(vpn sessions path) for redundancy",
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"direction": {
				Type:         schema.TypeString,
				Description:  "The traffic direction apply to the MSS clamping",
				Optional:     true,
				Default:      model.L2TcpMaxSegmentSizeClamping_DIRECTION_BOTH,
				ValidateFunc: validation.StringInSlice(L2VpnSessionTCPSegmentClampingDirection, false),
			},
			"max_segment_size": {
				Type:        schema.TypeInt,
				Description: "Maximum amount of data the host will accept in a Tcp segment.",
				Optional:    true,
				Computed:    true,
			},
			"local_address": {
				Type:         schema.TypeString,
				Description:  "IP Address of the local tunnel port. This property only applies in CLIENT mode",
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"peer_address": {
				Type:         schema.TypeString,
				Description:  "IP Address of the peer tunnel port. This property only applies in CLIENT mode",
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"protocol": {
				Type:         schema.TypeString,
				Description:  "Encapsulation protocol used by the tunnel.",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(L2VpnTunnelEncapsulationProtocal, false),
			},
		},
	}
}

func resourceNsxtPolicyL2VPNSessionCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	servicePath := d.Get("service_path").(string)
	isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
	if err != nil {
		return err
	}
	transportTunnel := getStringListFromSchemaList(d, "transport_tunnels")
	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}
	_, err = resourceNsxtPolicyL2VpnSessionExists(isT0, gwID, localeServiceID, serviceID, id, connector)
	if err == nil {
		return fmt.Errorf("L2VpnSession with nsx_id '%s' already exists.'", id)
	} else if !isNotFoundError(err) {
		return err
	}
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	enabled := d.Get("enabled").(bool)
	tags := getPolicyTagsFromSchema(d)

	obj := model.L2VPNSession{
		DisplayName:      &displayName,
		Description:      &description,
		Enabled:          &enabled,
		Tags:             tags,
		TransportTunnels: transportTunnel,
	}

	if util.NsxVersionHigherOrEqual("3.2.0") {
		direction := d.Get("direction").(string)
		maxSegmentSize := int64(d.Get("max_segment_size").(int))
		if direction != "" {
			l2TcpMSSClamping := model.L2TcpMaxSegmentSizeClamping{
				Direction: &direction,
			}
			if maxSegmentSize > 0 {
				l2TcpMSSClamping.MaxSegmentSize = &maxSegmentSize
			}
			obj.TcpMssClamping = &l2TcpMSSClamping
		}
		localAddress := d.Get("local_address").(string)
		peerAddress := d.Get("peer_address").(string)
		protocol := d.Get("protocol").(string)
		if localAddress != "" && peerAddress != "" {
			l2VpnTunnelEncapsulation := model.L2VPNTunnelEncapsulation{
				LocalEndpointAddress: &localAddress,
				PeerEndpointAddress:  &peerAddress,
				Protocol:             &protocol,
			}
			obj.TunnelEncapsulation = &l2VpnTunnelEncapsulation
		}
	}

	if isT0 {
		if localeServiceID == "" {
			client := t0_l2vpn_services.NewSessionsClient(connector)
			err = client.Patch(gwID, serviceID, id, obj)
		} else {
			client := t0_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Patch(gwID, localeServiceID, serviceID, id, obj)
		}
	} else {
		if localeServiceID == "" {
			client := t1_l2vpn_services.NewSessionsClient(connector)
			err = client.Patch(gwID, serviceID, id, obj)
		} else {
			client := t1_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Patch(gwID, localeServiceID, serviceID, id, obj)
		}
	}
	if err != nil {
		return handleCreateError("L2VPNSession", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyL2VPNSessionRead(d, m)
}

func parseL2VPNServicePolicyPath(path string) (bool, string, string, string, error) {
	segs := strings.Split(path, "/")
	// Path should be like /infra/tier-1s/aaa/locale-services/bbb/l2vpn-services/ccc
	// or /infra/tier-0s/aaa/l2vpn-services/bbb
	segCount := len(segs)
	if segCount < 6 || segCount > 8 || (segs[segCount-2] != "l2vpn-services") {
		// error - this is not a segment path
		return false, "", "", "", fmt.Errorf("Invalid L2 VPN service path %s", path)
	}

	serviceID := segs[segCount-1]
	gwPath := strings.Join(segs[:4], "/")

	localeServiceID := ""
	if segCount == 8 {
		localeServiceID = segs[5]
	}

	isT0, gwID := parseGatewayPolicyPath(gwPath)
	return isT0, gwID, localeServiceID, serviceID, nil
}

func resourceNsxtPolicyL2VPNSessionRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	servicePath := d.Get("service_path").(string)
	isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
	if err != nil {
		return err
	}
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2VPNSession ID")
	}

	var obj model.L2VPNSession
	if isT0 {
		if localeServiceID == "" {
			client := t0_l2vpn_services.NewSessionsClient(connector)
			obj, err = client.Get(gwID, serviceID, id)
		} else {
			client := t0_l2vpn_nested_services.NewSessionsClient(connector)
			obj, err = client.Get(gwID, localeServiceID, serviceID, id)
		}
	} else {
		if localeServiceID == "" {
			client := t1_l2vpn_services.NewSessionsClient(connector)
			obj, err = client.Get(gwID, serviceID, id)
		} else {
			client := t1_l2vpn_nested_services.NewSessionsClient(connector)
			obj, err = client.Get(gwID, localeServiceID, serviceID, id)
		}
	}
	if err != nil {
		return handleReadError(d, "L2VPNSession", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("enabled", obj.Enabled)

	if len(obj.TransportTunnels) > 0 {
		d.Set("transport_tunnels", obj.TransportTunnels)
	}
	if util.NsxVersionHigherOrEqual("3.2.0") {
		if obj.TcpMssClamping != nil {
			direction := obj.TcpMssClamping.Direction
			mss := obj.TcpMssClamping.MaxSegmentSize
			d.Set("direction", direction)
			d.Set("max_segment_size", mss)
		}
		if obj.TunnelEncapsulation != nil {
			localAddress := obj.TunnelEncapsulation.LocalEndpointAddress
			peerAddress := obj.TunnelEncapsulation.PeerEndpointAddress
			protocol := obj.TunnelEncapsulation.Protocol
			d.Set("local_address", localAddress)
			d.Set("peer_address", peerAddress)
			d.Set("protocol", protocol)
		}
	}
	d.SetId(id)

	return nil
}

func resourceNsxtPolicyL2VpnSessionExists(isT0 bool, gwID string, localeServiceID string, serviceID string, sessionID string, connector client.Connector) (bool, error) {
	var err error
	if isT0 {
		if localeServiceID == "" {
			client := t0_l2vpn_services.NewSessionsClient(connector)
			_, err = client.Get(gwID, serviceID, sessionID)
		} else {
			client := t0_l2vpn_nested_services.NewSessionsClient(connector)
			_, err = client.Get(gwID, localeServiceID, serviceID, sessionID)
		}
	} else {
		if localeServiceID == "" {
			client := t1_l2vpn_services.NewSessionsClient(connector)
			_, err = client.Get(gwID, serviceID, sessionID)
		} else {
			client := t1_l2vpn_nested_services.NewSessionsClient(connector)
			_, err = client.Get(gwID, localeServiceID, serviceID, sessionID)
		}
	}

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, err
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyL2VPNSessionUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	servicePath := d.Get("service_path").(string)
	isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
	if err != nil {
		return err
	}
	transportTunnel := getStringListFromSchemaList(d, "transport_tunnels")

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2VPNSession ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	revision := int64(d.Get("revision").(int))
	enabled := d.Get("enabled").(bool)
	tags := getPolicyTagsFromSchema(d)
	obj := model.L2VPNSession{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		TransportTunnels: transportTunnel,
		Revision:         &revision,
		Enabled:          &enabled,
	}
	if util.NsxVersionHigherOrEqual("3.2.0") {
		direction := d.Get("direction").(string)
		maxSegmentSize := int64(d.Get("max_segment_size").(int))
		if direction != "" {
			l2TcpMSSClamping := model.L2TcpMaxSegmentSizeClamping{
				Direction: &direction,
			}
			if maxSegmentSize > 0 {
				l2TcpMSSClamping.MaxSegmentSize = &maxSegmentSize
			}
			obj.TcpMssClamping = &l2TcpMSSClamping
		}
		localAddress := d.Get("local_address").(string)
		peerAddress := d.Get("peer_address").(string)
		protocol := d.Get("protocol").(string)
		if localAddress != "" && peerAddress != "" {
			l2VpnTunnelEncapsulation := model.L2VPNTunnelEncapsulation{
				LocalEndpointAddress: &localAddress,
				PeerEndpointAddress:  &peerAddress,
				Protocol:             &protocol,
			}
			obj.TunnelEncapsulation = &l2VpnTunnelEncapsulation
		}
	}
	if isT0 {
		if localeServiceID == "" {
			client := t0_l2vpn_services.NewSessionsClient(connector)
			err = client.Patch(gwID, serviceID, id, obj)
		} else {
			client := t0_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Patch(gwID, localeServiceID, serviceID, id, obj)
		}
	} else {
		if localeServiceID == "" {
			client := t1_l2vpn_services.NewSessionsClient(connector)
			err = client.Patch(gwID, serviceID, id, obj)
		} else {
			client := t1_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Patch(gwID, localeServiceID, serviceID, id, obj)
		}
	}

	if err != nil {
		return handleUpdateError("L2VPNSession", id, err)
	}

	return resourceNsxtPolicyL2VPNSessionRead(d, m)

}

func resourceNsxtPolicyL2VPNSessionDelete(d *schema.ResourceData, m interface{}) error {

	servicePath := d.Get("service_path").(string)
	isT0, gwID, localeServiceID, serviceID, err := parseL2VPNServicePolicyPath(servicePath)
	if err != nil {
		return err
	}
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2VPNSession ID")
	}

	connector := getPolicyConnector(m)
	if isT0 {
		if localeServiceID == "" {
			client := t0_l2vpn_services.NewSessionsClient(connector)
			err = client.Delete(gwID, serviceID, id)
		} else {
			client := t0_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Delete(gwID, localeServiceID, serviceID, id)
		}
	} else {
		if localeServiceID == "" {
			client := t1_l2vpn_services.NewSessionsClient(connector)
			err = client.Delete(gwID, serviceID, id)
		} else {
			client := t1_l2vpn_nested_services.NewSessionsClient(connector)
			err = client.Delete(gwID, localeServiceID, serviceID, id)
		}
	}

	if err != nil {
		return handleDeleteError("L2VPNSession", id, err)
	}

	return nil
}
