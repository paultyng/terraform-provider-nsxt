/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyDistributedFloodProtectionProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDistributedFloodProtectionProfileCreate,
		Read:   resourceNsxtPolicyDistributedFloodProtectionProfileRead,
		Update: resourceNsxtPolicyDistributedFloodProtectionProfileUpdate,
		Delete: resourceNsxtPolicyFloodProtectionProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},
		Schema: getDistributedFloodProtectionProfile(),
	}
}

func getFloodProtectionProfile() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"nsx_id":       getNsxIDSchema(),
		"path":         getPathSchema(),
		"display_name": getDisplayNameSchema(),
		"description":  getDescriptionSchema(),
		"revision":     getRevisionSchema(),
		"tag":          getTagsSchema(),
		"context":      getContextSchema(false, false, false),
		"icmp_active_flow_limit": {
			Type:         schema.TypeInt,
			Description:  "Active ICMP connections limit",
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 1000000),
		},
		"other_active_conn_limit": {
			Type:         schema.TypeInt,
			Description:  "Timeout after first TN",
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 1000000),
		},
		"tcp_half_open_conn_limit": {
			Type:         schema.TypeInt,
			Description:  "Active half open TCP connections limit",
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 1000000),
		},
		"udp_active_flow_limit": {
			Type:         schema.TypeInt,
			Description:  "Active UDP connections limit",
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 1000000),
		},
	}
}

func getDistributedFloodProtectionProfile() map[string]*schema.Schema {
	baseProfile := getFloodProtectionProfile()
	baseProfile["enable_rst_spoofing"] = &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Flag to indicate rst spoofing is enabled",
		Optional:    true,
		Default:     false,
	}
	baseProfile["enable_syncache"] = &schema.Schema{
		Type:        schema.TypeBool,
		Description: "Flag to indicate syncache is enabled",
		Optional:    true,
		Default:     false,
	}
	return baseProfile
}

func resourceNsxtPolicyFloodProtectionProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {

	client := infra.NewFloodProtectionProfilesClient(sessionContext, connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyDistributedFloodProtectionProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	icmpActiveFlowLimit := int64(d.Get("icmp_active_flow_limit").(int))
	otherActiveConnLimit := int64(d.Get("other_active_conn_limit").(int))
	tcpHalfOpenConnLimit := int64(d.Get("tcp_half_open_conn_limit").(int))
	udpActiveFlowLimit := int64(d.Get("udp_active_flow_limit").(int))
	enableRstSpoofing := d.Get("enable_rst_spoofing").(bool)
	enableSyncache := d.Get("enable_syncache").(bool)

	obj := model.DistributedFloodProtectionProfile{
		DisplayName:       &displayName,
		Description:       &description,
		Tags:              tags,
		ResourceType:      model.FloodProtectionProfile_RESOURCE_TYPE_DISTRIBUTEDFLOODPROTECTIONPROFILE,
		EnableRstSpoofing: &enableRstSpoofing,
		EnableSyncache:    &enableSyncache,
	}
	if icmpActiveFlowLimit != 0 {
		obj.IcmpActiveFlowLimit = &icmpActiveFlowLimit
	}
	if otherActiveConnLimit != 0 {
		obj.OtherActiveConnLimit = &otherActiveConnLimit
	}
	if tcpHalfOpenConnLimit != 0 {
		obj.TcpHalfOpenConnLimit = &tcpHalfOpenConnLimit
	}
	if udpActiveFlowLimit != 0 {
		obj.UdpActiveFlowLimit = &udpActiveFlowLimit
	}

	converter := bindings.NewTypeConverter()
	profileValue, errs := converter.ConvertToVapi(obj, model.DistributedFloodProtectionProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)

	log.Printf("[INFO] Patching DistributedFloodProtectionProfile with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewFloodProtectionProfilesClient(context, connector)
	return client.Patch(id, profileStruct, nil)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyFloodProtectionProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyDistributedFloodProtectionProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("FloodProtectionProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDistributedFloodProtectionProfileRead(d, m)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining FloodProtectionProfile ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewFloodProtectionProfilesClient(context, connector)
	dpffData, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "FloodProtectionProfile", id, err)
	}

	dfppInterface, errs := converter.ConvertToGolang(dpffData, model.DistributedFloodProtectionProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	obj := dfppInterface.(model.DistributedFloodProtectionProfile)

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("icmp_active_flow_limit", obj.IcmpActiveFlowLimit)
	d.Set("other_active_conn_limit", obj.OtherActiveConnLimit)
	d.Set("tcp_half_open_conn_limit", obj.TcpHalfOpenConnLimit)
	d.Set("udp_active_flow_limit", obj.UdpActiveFlowLimit)
	d.Set("enable_rst_spoofing", obj.EnableRstSpoofing)
	d.Set("enable_syncache", obj.EnableSyncache)

	return nil
}

func resourceNsxtPolicyDistributedFloodProtectionProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining FloodProtectionProfile ID")
	}

	err := resourceNsxtPolicyDistributedFloodProtectionProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("FloodProtectionProfile", id, err)
	}

	return resourceNsxtPolicyDistributedFloodProtectionProfileRead(d, m)
}

func resourceNsxtPolicyFloodProtectionProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining FloodProtectionProfile ID")
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewFloodProtectionProfilesClient(context, connector)
	err = client.Delete(id, nil)
	if err != nil {
		return handleDeleteError("FloodProtectionProfile", id, err)
	}
	return nil
}
