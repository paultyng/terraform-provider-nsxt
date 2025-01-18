// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_static_routes "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/tier_0s/static_routes"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/static_routes"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyStaticRouteBfdPeer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyStaticRouteBfdPeerCreate,
		Read:   resourceNsxtPolicyStaticRouteBfdPeerRead,
		Update: resourceNsxtPolicyStaticRouteBfdPeerUpdate,
		Delete: resourceNsxtPolicyStaticRouteBfdPeerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0GatewayImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":           getNsxIDSchema(),
			"path":             getPathSchema(),
			"display_name":     getDisplayNameSchema(),
			"description":      getDescriptionSchema(),
			"revision":         getRevisionSchema(),
			"tag":              getTagsSchema(),
			"gateway_path":     getPolicyPathSchema(true, true, "Policy path for Tier0 gateway"),
			"bfd_profile_path": getPolicyPathSchema(true, false, "Policy path for BFD Profile"),
			"enabled": {
				Type:        schema.TypeBool,
				Default:     true,
				Description: "Flag to enable/disable this peer",
				Optional:    true,
			},
			"peer_address": {
				Type:         schema.TypeString,
				Description:  "IPv4 Address of the peer",
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"source_addresses": {
				Type:        schema.TypeList,
				Description: "Array of Tier0 external interface IP addresses",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
				Optional: true,
			},
		},
	}
}

func resourceNsxtPolicyStaticRouteBfdPeerExists(gwID string, id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	if isGlobalManager {
		client := gm_static_routes.NewBfdPeersClient(connector)
		_, err = client.Get(gwID, id)
	} else {
		client := static_routes.NewBfdPeersClient(connector)
		_, err = client.Get(gwID, id)
	}
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func policyStaticRouteBfdPeerPatch(d *schema.ResourceData, m interface{}, gwID string, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	enabled := d.Get("enabled").(bool)
	bfdProfilePath := d.Get("bfd_profile_path").(string)
	peerAddress := d.Get("peer_address").(string)
	sourceAddresses := getStringListFromSchemaList(d, "source_addresses")

	obj := model.StaticRouteBfdPeer{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		Enabled:        &enabled,
		BfdProfilePath: &bfdProfilePath,
		PeerAddress:    &peerAddress,
	}

	if len(sourceAddresses) > 0 {
		obj.SourceAddresses = sourceAddresses
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating Gateway BFD Peer with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.StaticRouteBfdPeerBindingType(), gm_model.StaticRouteBfdPeerBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_static_routes.NewBfdPeersClient(connector)
		return client.Patch(gwID, id, gmObj.(gm_model.StaticRouteBfdPeer))
	}

	client := static_routes.NewBfdPeersClient(connector)
	return client.Patch(gwID, id, obj)
}

func resourceNsxtPolicyStaticRouteBfdPeerExistsOnGateway(gwID string) func(id string, connector client.Connector, isGlobalManager bool) (bool, error) {

	return func(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
		return resourceNsxtPolicyStaticRouteBfdPeerExists(id, gwID, connector, isGlobalManager)
	}
}

func resourceNsxtPolicyStaticRouteBfdPeerCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	id, err := getOrGenerateID(d, m, resourceNsxtPolicyStaticRouteBfdPeerExistsOnGateway(gwID))
	if err != nil {
		return err
	}

	err = policyStaticRouteBfdPeerPatch(d, m, gwID, id)
	if err != nil {
		return handleCreateError("BFD Peer", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyStaticRouteBfdPeerRead(d, m)
}

func resourceNsxtPolicyStaticRouteBfdPeerRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Gateway BFD Peer ID")
	}
	gwPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPath)
	if !isT0 {
		return fmt.Errorf("Tier0 Gateway path expected, got %s", gwPath)
	}

	var obj model.StaticRouteBfdPeer
	if isPolicyGlobalManager(m) {
		client := gm_static_routes.NewBfdPeersClient(connector)
		gmObj, err := client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway BFD Peer", id, err)
		}

		lmObj, err := convertModelBindingType(gmObj, gm_model.StaticRouteBfdPeerBindingType(), model.StaticRouteBfdPeerBindingType())
		if err != nil {
			return err
		}
		obj = lmObj.(model.StaticRouteBfdPeer)
	} else {
		client := static_routes.NewBfdPeersClient(connector)
		var err error
		obj, err = client.Get(gwID, id)
		if err != nil {
			return handleReadError(d, "Gateway BFD Peer", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("bfd_profile_path", obj.BfdProfilePath)
	d.Set("enabled", obj.Enabled)
	d.Set("peer_address", obj.PeerAddress)
	d.Set("source_addresses", obj.SourceAddresses)

	return nil
}

func resourceNsxtPolicyStaticRouteBfdPeerUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Static Route Bfd Peer ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	err := policyStaticRouteBfdPeerPatch(d, m, gwID, id)
	if err != nil {
		return handleUpdateError("BFD Peer", id, err)
	}

	return resourceNsxtPolicyStaticRouteBfdPeerRead(d, m)
}

func resourceNsxtPolicyStaticRouteBfdPeerDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Static Route Bfd Peer ID")
	}
	gwPath := d.Get("gateway_path").(string)
	_, gwID := parseGatewayPolicyPath(gwPath)

	connector := getPolicyConnector(m)
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_static_routes.NewBfdPeersClient(connector)
		err = client.Delete(gwID, id)
	} else {
		client := static_routes.NewBfdPeersClient(connector)
		err = client.Delete(gwID, id)
	}

	if err != nil {
		return handleDeleteError("Static Route Bfd Peer", id, err)
	}

	return nil
}
