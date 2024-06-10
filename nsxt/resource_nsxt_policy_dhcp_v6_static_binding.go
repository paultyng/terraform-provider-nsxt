/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/segments"
	t1_segments "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s/segments"
)

func resourceNsxtPolicyDhcpV6StaticBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDhcpV6StaticBindingCreate,
		Read:   resourceNsxtPolicyDhcpV6StaticBindingRead,
		Update: resourceNsxtPolicyDhcpV6StaticBindingUpdate,
		Delete: resourceNsxtPolicyDhcpStaticBindingDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtSegmentResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"segment_path": getPolicyPathSchema(true, true, "segment path"),
			"dns_nameservers": {
				Type:        schema.TypeList,
				Description: "DNS nameservers",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sntp_servers": {
				Type:        schema.TypeList,
				Description: "SNTP server IP addresses",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
			},
			"domain_names": getDomainNamesSchema(),
			"ip_addresses": {
				Type:        schema.TypeList,
				Description: "IP addresses",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv6Address,
				},
			},
			"lease_time":     getDhcpLeaseTimeSchema(),
			"preferred_time": getDhcpPreferredTimeSchema(),
			"mac_address": {
				Type:         schema.TypeString,
				Description:  "MAC address of the host",
				Required:     true,
				ValidateFunc: validation.IsMACAddress,
			},
		},
	}
}

func policyDhcpV6StaticBindingConvertAndPatch(d *schema.ResourceData, segmentPath string, id string, m interface{}) error {

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	ipAddresses := getStringListFromSchemaList(d, "ip_addresses")
	domainNames := getStringListFromSchemaList(d, "domain_names")
	dnsNameservers := getStringListFromSchemaList(d, "dns_nameservers")
	sntpServers := getStringListFromSchemaList(d, "sntp_servers")
	leaseTime := int64(d.Get("lease_time").(int))
	preferredTime := int64(d.Get("preferred_time").(int))
	macAddress := d.Get("mac_address").(string)

	obj := model.DhcpV6StaticBindingConfig{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		LeaseTime:    &leaseTime,
		MacAddress:   &macAddress,
		ResourceType: "DhcpV6StaticBindingConfig",
	}

	if len(ipAddresses) > 0 {
		obj.IpAddresses = ipAddresses
	}

	if len(domainNames) > 0 {
		obj.DomainNames = domainNames
	}

	if len(dnsNameservers) > 0 {
		obj.DnsNameservers = dnsNameservers
	}

	if len(sntpServers) > 0 {
		obj.SntpServers = sntpServers
	}

	if preferredTime > 0 {
		obj.PreferredTime = &preferredTime
	}

	connector := getPolicyConnector(m)

	converter := bindings.NewTypeConverter()

	isT0, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if isT0 {
		return fmt.Errorf("This resource is not applicable to segment %s", segmentPath)
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	if isPolicyGlobalManager(m) && gwID != "" {
		return fmt.Errorf("This resource is not applicable to segment on Global Manager %s", segmentPath)
	}
	convObj, convErrs := converter.ConvertToVapi(obj, model.DhcpV6StaticBindingConfigBindingType())
	if convErrs != nil {
		return convErrs[0]
	}
	if gwID == "" {
		// infra segment
		client := segments.NewDhcpStaticBindingConfigsClient(context, connector)
		return client.Patch(segmentID, id, convObj.(*data.StructValue))
	}

	// fixed segment
	client := t1_segments.NewDhcpStaticBindingConfigsClient(context, connector)
	return client.Patch(gwID, segmentID, id, convObj.(*data.StructValue))
}

func resourceNsxtPolicyDhcpV6StaticBindingCreate(d *schema.ResourceData, m interface{}) error {

	segmentPath := d.Get("segment_path").(string)
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyDhcpStaticBindingExists(segmentPath))
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating DhcpV6 Static Binding Config with ID %s", id)
	err = policyDhcpV6StaticBindingConvertAndPatch(d, segmentPath, id, m)

	if err != nil {
		return handleCreateError("DhcpV6 Static Binding Config", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDhcpV6StaticBindingRead(d, m)
}

func resourceNsxtPolicyDhcpV6StaticBindingRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpV6 Static Binding Config ID")
	}

	segmentPath := d.Get("segment_path").(string)
	isT0, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if isT0 {
		return fmt.Errorf("This resource is not applicable to segment %s", segmentPath)
	}

	var obj model.DhcpV6StaticBindingConfig
	converter := bindings.NewTypeConverter()
	var err error
	var dhcpObj *data.StructValue
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	if gwID == "" {
		// infra segment
		client := segments.NewDhcpStaticBindingConfigsClient(context, connector)
		dhcpObj, err = client.Get(segmentID, id)
	} else {
		// fixed segment
		client := t1_segments.NewDhcpStaticBindingConfigsClient(context, connector)
		dhcpObj, err = client.Get(gwID, segmentID, id)
	}
	if err != nil {
		return handleReadError(d, "DhcpV6 Static Binding Config", id, err)
	}

	convObj, errs := converter.ConvertToGolang(dhcpObj, model.DhcpV6StaticBindingConfigBindingType())
	if errs != nil {
		return errs[0]
	}
	obj = convObj.(model.DhcpV6StaticBindingConfig)

	if obj.ResourceType != "DhcpV6StaticBindingConfig" {
		return handleReadError(d, "DhcpV6 Static Binding Config", id, fmt.Errorf("Unexpected ResourceType"))
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("ip_addresses", obj.IpAddresses)
	d.Set("lease_time", obj.LeaseTime)
	d.Set("preferred_time", obj.PreferredTime)
	d.Set("mac_address", obj.MacAddress)
	d.Set("dns_nameservers", obj.DnsNameservers)
	d.Set("domain_names", obj.DomainNames)
	d.Set("sntp_servers", obj.SntpServers)

	return nil
}

func resourceNsxtPolicyDhcpV6StaticBindingUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpV6 Static Binding Config ID")
	}
	segmentPath := d.Get("segment_path").(string)

	log.Printf("[INFO] Updating DhcpV6StaticBindingConfig with ID %s", id)
	err := policyDhcpV6StaticBindingConvertAndPatch(d, segmentPath, id, m)
	if err != nil {
		return handleUpdateError("DhcpV6 Static Binding Config", id, err)
	}

	return resourceNsxtPolicyDhcpV6StaticBindingRead(d, m)
}
