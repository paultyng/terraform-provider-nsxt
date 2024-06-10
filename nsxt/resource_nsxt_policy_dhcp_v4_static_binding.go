/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/segments"
	t1_segments "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s/segments"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyDhcpV4StaticBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDhcpV4StaticBindingCreate,
		Read:   resourceNsxtPolicyDhcpV4StaticBindingRead,
		Update: resourceNsxtPolicyDhcpV4StaticBindingUpdate,
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
			"gateway_address": {
				Type:         schema.TypeString,
				Description:  "When not specified, gateway address is auto-assigned from segment configuration",
				ValidateFunc: validation.IsIPv4Address,
				Optional:     true,
			},
			"hostname": {
				Type:        schema.TypeString,
				Description: "Hostname to assign to the host",
				Optional:    true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "IP assigned to host. The IP address must belong to the subnetconfigured on segment",
				ValidateFunc: validation.IsIPv4Address,
				Required:     true,
			},
			"lease_time": getDhcpLeaseTimeSchema(),
			"mac_address": {
				Type:         schema.TypeString,
				Description:  "MAC address of the host",
				Required:     true,
				ValidateFunc: validation.IsMACAddress,
			},
			"dhcp_option_121":     getDhcpOptions121Schema(),
			"dhcp_generic_option": getDhcpGenericOptionsSchema(),
		},
	}
}

func getPolicyDchpStaticBindingOnSegment(context utl.SessionContext, id string, segmentPath string, connector client.Connector) (*data.StructValue, error) {
	_, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if segmentID == "" {
		return nil, fmt.Errorf("Invalid Segment Path %s", segmentPath)
	}
	if context.ClientType == utl.Global || gwID == "" {
		client := segments.NewDhcpStaticBindingConfigsClient(context, connector)
		return client.Get(segmentID, id)
	}

	client := t1_segments.NewDhcpStaticBindingConfigsClient(context, connector)
	return client.Get(gwID, segmentID, id)
}

func resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(context utl.SessionContext, id string, segmentPath string, connector client.Connector) (bool, error) {
	_, err := getPolicyDchpStaticBindingOnSegment(context, id, segmentPath, connector)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyDhcpStaticBindingExists(segmentPath string) func(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
	return func(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
		return resourceNsxtPolicyDhcpStaticBindingExistsOnSegment(context, id, segmentPath, connector)
	}
}

func policyDhcpV4StaticBindingConvertAndPatch(d *schema.ResourceData, segmentPath string, id string, m interface{}) error {

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	gatewayAddress := d.Get("gateway_address").(string)
	hostName := d.Get("hostname").(string)
	ipAddress := d.Get("ip_address").(string)
	leaseTime := int64(d.Get("lease_time").(int))
	macAddress := d.Get("mac_address").(string)
	dhcpOptions := getDhcpOptsFromSchema(d)

	obj := model.DhcpV4StaticBindingConfig{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		HostName:     &hostName,
		IpAddress:    &ipAddress,
		LeaseTime:    &leaseTime,
		MacAddress:   &macAddress,
		Options:      dhcpOptions,
		ResourceType: "DhcpV4StaticBindingConfig",
	}

	if len(gatewayAddress) > 0 {
		obj.GatewayAddress = &gatewayAddress
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	converter := bindings.NewTypeConverter()

	isT0, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if isT0 {
		return fmt.Errorf("This resource is not applicable to segment %s", segmentPath)
	}

	if context.ClientType == utl.Global && gwID != "" {
		return fmt.Errorf("This resource is not applicable to segment on Global Manager %s", segmentPath)
	}

	convObj, convErrs := converter.ConvertToVapi(obj, model.DhcpV4StaticBindingConfigBindingType())
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

func getDhcpOptsFromSchema(d *schema.ResourceData) *model.DhcpV4Options {
	dhcpOpts := model.DhcpV4Options{}

	dhcp121Opts := d.Get("dhcp_option_121").([]interface{})
	if len(dhcp121Opts) > 0 {
		dhcp121OptStruct := getPolicyDhcpOptions121(dhcp121Opts)
		dhcpOpts.Option121 = &dhcp121OptStruct
	}

	otherDhcpOpts := d.Get("dhcp_generic_option").([]interface{})
	if len(otherDhcpOpts) > 0 {
		otherOptStructs := getPolicyDhcpGenericOptions(otherDhcpOpts)
		dhcpOpts.Others = otherOptStructs
	}

	if len(dhcp121Opts)+len(otherDhcpOpts) > 0 {
		return &dhcpOpts
	}

	return nil

}

func resourceNsxtPolicyDhcpV4StaticBindingCreate(d *schema.ResourceData, m interface{}) error {

	segmentPath := d.Get("segment_path").(string)
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyDhcpStaticBindingExists(segmentPath))
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating DhcpV4 Static Binding Config with ID %s on segment %s", id, segmentPath)
	err = policyDhcpV4StaticBindingConvertAndPatch(d, segmentPath, id, m)

	if err != nil {
		return handleCreateError("DhcpV4 Static Binding Config", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDhcpV4StaticBindingRead(d, m)
}

func resourceNsxtPolicyDhcpV4StaticBindingRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpV4 Static Binding Config ID")
	}

	segmentPath := d.Get("segment_path").(string)

	var obj model.DhcpV4StaticBindingConfig
	converter := bindings.NewTypeConverter()
	var err error
	var dhcpObj *data.StructValue
	isT0, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if isT0 {
		return fmt.Errorf("This resource is not applicable to segment %s", segmentPath)
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if context.ClientType == utl.Global && gwID != "" {
		return fmt.Errorf("This resource is not applicable to segment on Global Manager %s", segmentPath)
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
		return handleReadError(d, "DhcpV4 Static Binding Config", id, err)
	}

	convObj, errs := converter.ConvertToGolang(dhcpObj, model.DhcpV4StaticBindingConfigBindingType())
	if errs != nil {
		return errs[0]
	}
	obj = convObj.(model.DhcpV4StaticBindingConfig)

	if obj.ResourceType != "DhcpV4StaticBindingConfig" {
		return handleReadError(d, "DhcpV4 Static Binding Config", id, fmt.Errorf("Unexpected ResourceType"))
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("gateway_address", obj.GatewayAddress)
	d.Set("hostname", obj.HostName)
	d.Set("ip_address", obj.IpAddress)
	d.Set("lease_time", obj.LeaseTime)
	d.Set("mac_address", obj.MacAddress)

	if obj.Options != nil {
		if obj.Options.Option121 != nil {
			d.Set("dhcp_option_121", getPolicyDhcpOptions121FromStruct(obj.Options.Option121))
		}

		if len(obj.Options.Others) > 0 {
			d.Set("dhcp_generic_option", getPolicyDhcpGenericOptionsFromStruct(obj.Options.Others))
		}
	}

	return nil
}

func resourceNsxtPolicyDhcpV4StaticBindingUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpV4 Static Binding Config ID")
	}
	segmentPath := d.Get("segment_path").(string)

	log.Printf("[INFO] Updating DhcpV4 Static Binding Config with ID %s", id)
	err := policyDhcpV4StaticBindingConvertAndPatch(d, segmentPath, id, m)
	if err != nil {
		return handleUpdateError("DhcpV4 Static Binding Config", id, err)
	}

	return resourceNsxtPolicyDhcpV4StaticBindingRead(d, m)
}

func resourceNsxtPolicyDhcpStaticBindingDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Dhcp Static Binding Config ID")
	}
	segmentPath := d.Get("segment_path").(string)
	_, gwID, segmentID := parseSegmentPolicyPath(segmentPath)

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if gwID == "" {
		// infra segment
		client := segments.NewDhcpStaticBindingConfigsClient(context, connector)
		err = client.Delete(segmentID, id)
	} else {
		// fixed segment
		client := t1_segments.NewDhcpStaticBindingConfigsClient(context, connector)
		err = client.Delete(gwID, segmentID, id)
	}

	if err != nil {
		return handleDeleteError("Dhcp Static Binding Config", id, err)
	}

	return nil
}

func nsxtSegmentResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	importSegment := ""
	importGW := ""
	s := strings.Split(importID, "/")
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		segmentPath, err := getParameterFromPolicyPath("", "/dhcp-static-binding-configs/", importID)
		if err != nil {
			return nil, err
		}
		d.Set("segment_path", segmentPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	if len(s) < 2 {
		return []*schema.ResourceData{d}, fmt.Errorf("Import format [gatewayID]/segmentID/bindingID expected, got %s", importID)
	}
	if len(s) == 3 {
		importGW = s[0]
		importSegment = s[1]
		d.SetId(s[2])
	} else {
		importSegment = s[0]
		d.SetId(s[1])
	}

	infra := "infra"
	if isPolicyGlobalManager(m) {
		infra = "global-infra"
	}

	parentPath := fmt.Sprintf("/%s/segments", infra)
	if len(importGW) > 0 {
		parentPath = fmt.Sprintf("/%s/tier-1s/%s/segments", infra, importGW)
	}
	segmentPath := fmt.Sprintf("%s/%s", parentPath, importSegment)
	d.Set("segment_path", segmentPath)

	return []*schema.ResourceData{d}, nil
}
