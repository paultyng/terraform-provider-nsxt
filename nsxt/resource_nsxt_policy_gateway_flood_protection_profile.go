/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func resourceNsxtPolicyGatewayFloodProtectionProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayFloodProtectionProfileCreate,
		Read:   resourceNsxtPolicyGatewayFloodProtectionProfileRead,
		Update: resourceNsxtPolicyGatewayFloodProtectionProfileUpdate,
		Delete: resourceNsxtPolicyFloodProtectionProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},
		Schema: getGatewayFloodProtectionProfile(),
	}
}

func getGatewayFloodProtectionProfile() map[string]*schema.Schema {
	baseProfile := getFloodProtectionProfile()
	baseProfile["nat_active_conn_limit"] = &schema.Schema{
		Type:        schema.TypeInt,
		Description: "Maximum limit of active NAT connections",
		Optional:    true,
		Computed:    true,
	}
	return baseProfile
}

func resourceNsxtPolicyGatewayFloodProtectionProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	icmpActiveFlowLimit := int64(d.Get("icmp_active_flow_limit").(int))
	otherActiveConnLimit := int64(d.Get("other_active_conn_limit").(int))
	tcpHalfOpenConnLimit := int64(d.Get("tcp_half_open_conn_limit").(int))
	udpActiveFlowLimit := int64(d.Get("udp_active_flow_limit").(int))
	natActiveConnLimit := int64(d.Get("nat_active_conn_limit").(int))

	obj := model.GatewayFloodProtectionProfile{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		ResourceType: model.FloodProtectionProfile_RESOURCE_TYPE_GATEWAYFLOODPROTECTIONPROFILE,
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
	if natActiveConnLimit != 0 {
		obj.NatActiveConnLimit = &natActiveConnLimit
	}

	converter := bindings.NewTypeConverter()
	profileValue, errs := converter.ConvertToVapi(obj, model.GatewayFloodProtectionProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)

	log.Printf("[INFO] Patching GatewayFloodProtectionProfile with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewFloodProtectionProfilesClient(context, connector)
	return client.Patch(id, profileStruct, nil)
}

func resourceNsxtPolicyGatewayFloodProtectionProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyFloodProtectionProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyGatewayFloodProtectionProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("GatewayFloodProtectionProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayFloodProtectionProfileRead(d, m)
}

func resourceNsxtPolicyGatewayFloodProtectionProfileRead(d *schema.ResourceData, m interface{}) error {
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
	gpffData, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "GatewayFloodProtectionProfile", id, err)
	}

	gfppInterface, errs := converter.ConvertToGolang(gpffData, model.GatewayFloodProtectionProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	obj := gfppInterface.(model.GatewayFloodProtectionProfile)

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
	d.Set("nat_active_conn_limit", obj.NatActiveConnLimit)

	return nil
}

func resourceNsxtPolicyGatewayFloodProtectionProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayFloodProtectionProfile ID")
	}

	err := resourceNsxtPolicyGatewayFloodProtectionProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("GatewayFloodProtectionProfile", id, err)
	}

	return resourceNsxtPolicyGatewayFloodProtectionProfileRead(d, m)
}
