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

func resourceNsxtVlanLogicalSwitch() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVlanLogicalSwitchCreate,
		Read:   resourceNsxtVlanLogicalSwitchRead,
		Update: resourceNsxtVlanLogicalSwitchUpdate,
		Delete: resourceNsxtVlanLogicalSwitchDelete,
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
			"tag":             getTagsSchema(),
			"address_binding": getAddressBindingsSchema(),
			"admin_state":     getAdminStateSchema(),
			"ip_pool_id": {
				Type:        schema.TypeString,
				Description: "IP pool id that associated with a LogicalSwitch",
				Optional:    true,
			},
			"mac_pool_id": {
				Type:        schema.TypeString,
				Description: "Mac pool id that associated with a LogicalSwitch",
				Optional:    true,
			},
			"switching_profile_id": getSwitchingProfileIdsSchema(),
			"transport_zone_id": {
				Type:        schema.TypeString,
				Description: "Id of the TransportZone to which this LogicalSwitch is associated",
				Required:    true,
				ForceNew:    true,
			},
			"vlan": {
				Type:        schema.TypeInt,
				Description: "VLAN Id",
				Required:    true,
			},
			// TODO - add option for vlan trunk spec
		},
	}
}

func resourceNsxtVlanLogicalSwitchCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	addressBindings := getAddressBindingsFromSchema(d)
	adminState := d.Get("admin_state").(string)
	ipPoolID := d.Get("ip_pool_id").(string)
	macPoolID := d.Get("mac_pool_id").(string)
	switchingProfileID := getSwitchingProfileIdsFromSchema(d)
	transportZoneID := d.Get("transport_zone_id").(string)
	vlan := int64(d.Get("vlan").(int))

	logicalSwitch := manager.LogicalSwitch{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		AddressBindings:     addressBindings,
		AdminState:          adminState,
		IpPoolId:            ipPoolID,
		MacPoolId:           macPoolID,
		SwitchingProfileIds: switchingProfileID,
		TransportZoneId:     transportZoneID,
		Vlan:                vlan,
	}

	logicalSwitch, resp, err := nsxClient.LogicalSwitchingApi.CreateLogicalSwitch(nsxClient.Context, logicalSwitch)

	if err != nil {
		return fmt.Errorf("Error during LogicalSwitch create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during LogicalSwitch create: %v", resp.StatusCode)
	}

	toleratePartialSuccess := getCommonProviderConfig(m.(nsxtClients)).ToleratePartialSuccess
	err = resourceNsxtLogicalSwitchVerifyRealization(d, nsxClient, &logicalSwitch, toleratePartialSuccess)

	if err != nil {
		return err
	}

	d.SetId(logicalSwitch.Id)

	return resourceNsxtVlanLogicalSwitchRead(d, m)
}

func resourceNsxtVlanLogicalSwitchRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical switch id")
	}

	logicalSwitch, resp, err := nsxClient.LogicalSwitchingApi.GetLogicalSwitch(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalSwitch %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalSwitch read: %v", err)
	}

	d.Set("revision", logicalSwitch.Revision)
	d.Set("description", logicalSwitch.Description)
	d.Set("display_name", logicalSwitch.DisplayName)
	setTagsInSchema(d, logicalSwitch.Tags)
	err = setAddressBindingsInSchema(d, logicalSwitch.AddressBindings)
	if err != nil {
		return fmt.Errorf("Error during logical switch address bindings set in schema: %v", err)
	}
	d.Set("admin_state", logicalSwitch.AdminState)
	d.Set("ip_pool_id", logicalSwitch.IpPoolId)
	d.Set("mac_pool_id", logicalSwitch.MacPoolId)
	err = setSwitchingProfileIdsInSchema(d, nsxClient, logicalSwitch.SwitchingProfileIds)
	if err != nil {
		return fmt.Errorf("Error during logical switch profiles set in schema: %v", err)
	}
	d.Set("transport_zone_id", logicalSwitch.TransportZoneId)
	d.Set("vlan", logicalSwitch.Vlan)

	return nil
}

func resourceNsxtVlanLogicalSwitchUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical switch id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	addressBindings := getAddressBindingsFromSchema(d)
	adminState := d.Get("admin_state").(string)
	ipPoolID := d.Get("ip_pool_id").(string)
	macPoolID := d.Get("mac_pool_id").(string)
	switchingProfileID := getSwitchingProfileIdsFromSchema(d)
	transportZoneID := d.Get("transport_zone_id").(string)
	vlan := int64(d.Get("vlan").(int))
	revision := int64(d.Get("revision").(int))
	logicalSwitch := manager.LogicalSwitch{
		Description:         description,
		DisplayName:         displayName,
		Tags:                tags,
		AddressBindings:     addressBindings,
		AdminState:          adminState,
		IpPoolId:            ipPoolID,
		MacPoolId:           macPoolID,
		SwitchingProfileIds: switchingProfileID,
		TransportZoneId:     transportZoneID,
		Vlan:                vlan,
		Revision:            revision,
	}

	_, resp, err := nsxClient.LogicalSwitchingApi.UpdateLogicalSwitch(nsxClient.Context, id, logicalSwitch)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalSwitch update: %v", err)
	}

	return resourceNsxtVlanLogicalSwitchRead(d, m)
}

func resourceNsxtVlanLogicalSwitchDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical switch id")
	}

	localVarOptionals := make(map[string]interface{})
	localVarOptionals["cascade"] = true
	resp, err := nsxClient.LogicalSwitchingApi.DeleteLogicalSwitch(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LogicalSwitch delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalSwitch %s not found", id)
		d.SetId("")
	}
	return nil
}
