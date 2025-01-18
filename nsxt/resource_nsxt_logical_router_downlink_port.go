// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var logicalRouterPortUrpfModeValues = []string{"NONE", "STRICT"}

func resourceNsxtLogicalRouterDownLinkPort() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalRouterDownLinkPortCreate,
		Read:   resourceNsxtLogicalRouterDownLinkPortRead,
		Update: resourceNsxtLogicalRouterDownLinkPortUpdate,
		Delete: resourceNsxtLogicalRouterDownLinkPortDelete,
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
			"tag": getTagsSchema(),
			"logical_router_id": {
				Type:        schema.TypeString,
				Description: "Identifier for logical router on which this port is created",
				Required:    true,
				ForceNew:    true,
			},
			"linked_logical_switch_port_id": {
				Type:        schema.TypeString,
				Description: "Identifier for port on logical switch to connect to",
				Required:    true,
				ForceNew:    true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "Logical router port subnet (ipAddress / prefix length)",
				Required:     true,
				ValidateFunc: validatePortAddress(),
			},
			"mac_address": {
				Type:        schema.TypeString,
				Description: "MAC address",
				Computed:    true,
			},
			"urpf_mode": {
				Type:         schema.TypeString,
				Description:  "Unicast Reverse Path Forwarding mode",
				Optional:     true,
				Default:      "STRICT",
				ValidateFunc: validation.StringInSlice(logicalRouterPortUrpfModeValues, false),
			},
			"service_binding": getResourceReferencesSchema(false, false, []string{"LogicalService", "DhcpRelayService"}, "Service Bindings"),
		},
	}
}

func getIPSubnetsFromCidr(cidr string) []manager.IpSubnet {
	s := strings.Split(cidr, "/")
	ipAddress := s[0]
	prefix, _ := strconv.ParseUint(s[1], 10, 32)
	var subnetList []manager.IpSubnet
	elem := manager.IpSubnet{
		IpAddresses:  []string{ipAddress},
		PrefixLength: int64(prefix),
	}
	subnetList = append(subnetList, elem)
	return subnetList
}

func setIPSubnetsInSchema(d *schema.ResourceData, subnets []manager.IpSubnet) {
	for _, subnet := range subnets {
		// only 1 subnet is expected
		cidr := fmt.Sprintf("%s/%d", subnet.IpAddresses[0], subnet.PrefixLength)
		d.Set("ip_address", cidr)
	}
}

func resourceNsxtLogicalRouterDownLinkPortCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	macAddress := d.Get("mac_address").(string)
	linkedLogicalSwitchPortID := d.Get("linked_logical_switch_port_id").(string)
	subnets := getIPSubnetsFromCidr(d.Get("ip_address").(string))
	urpfMode := d.Get("urpf_mode").(string)
	serviceBinding := getServiceBindingsFromSchema(d, "service_binding")
	logicalRouterDownLinkPort := manager.LogicalRouterDownLinkPort{
		Description:               description,
		DisplayName:               displayName,
		Tags:                      tags,
		LogicalRouterId:           logicalRouterID,
		MacAddress:                macAddress,
		LinkedLogicalSwitchPortId: makeResourceReference("LogicalPort", linkedLogicalSwitchPortID),
		Subnets:                   subnets,
		UrpfMode:                  urpfMode,
		ServiceBindings:           serviceBinding,
	}

	logicalRouterDownLinkPort, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouterDownLinkPort(nsxClient.Context, logicalRouterDownLinkPort)

	if err != nil {
		return fmt.Errorf("Error during LogicalRouterDownLinkPort create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during LogicalRouterDownLinkPort create: %v", resp.StatusCode)
	}
	d.SetId(logicalRouterDownLinkPort.Id)

	return resourceNsxtLogicalRouterDownLinkPortRead(d, m)
}

func resourceNsxtLogicalRouterDownLinkPortRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router downlink port id while reading")
	}

	logicalRouterDownLinkPort, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouterDownLinkPort(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalRouterDownLinkPort %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterDownLinkPort read: %v", err)
	}

	d.Set("revision", logicalRouterDownLinkPort.Revision)
	d.Set("description", logicalRouterDownLinkPort.Description)
	d.Set("display_name", logicalRouterDownLinkPort.DisplayName)
	setTagsInSchema(d, logicalRouterDownLinkPort.Tags)
	d.Set("logical_router_id", logicalRouterDownLinkPort.LogicalRouterId)
	d.Set("mac_address", logicalRouterDownLinkPort.MacAddress)
	d.Set("linked_logical_switch_port_id", logicalRouterDownLinkPort.LinkedLogicalSwitchPortId.TargetId)
	setIPSubnetsInSchema(d, logicalRouterDownLinkPort.Subnets)
	d.Set("urpf_mode", logicalRouterDownLinkPort.UrpfMode)
	err = setServiceBindingsInSchema(d, logicalRouterDownLinkPort.ServiceBindings, "service_binding")
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterDownLinkPort service_binding set in schema: %v", err)
	}

	return nil
}

func resourceNsxtLogicalRouterDownLinkPortUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router downlink port id while updating")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	logicalRouterID := d.Get("logical_router_id").(string)
	linkedLogicalSwitchPortID := d.Get("linked_logical_switch_port_id").(string)
	subnets := getIPSubnetsFromCidr(d.Get("ip_address").(string))
	macAddress := d.Get("mac_address").(string)
	urpfMode := d.Get("urpf_mode").(string)
	serviceBinding := getServiceBindingsFromSchema(d, "service_binding")
	logicalRouterDownLinkPort := manager.LogicalRouterDownLinkPort{
		Revision:                  revision,
		Description:               description,
		DisplayName:               displayName,
		Tags:                      tags,
		LogicalRouterId:           logicalRouterID,
		MacAddress:                macAddress,
		LinkedLogicalSwitchPortId: makeResourceReference("LogicalPort", linkedLogicalSwitchPortID),
		Subnets:                   subnets,
		UrpfMode:                  urpfMode,
		ServiceBindings:           serviceBinding,
		ResourceType:              "LogicalRouterDownLinkPort",
	}

	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateLogicalRouterDownLinkPort(nsxClient.Context, id, logicalRouterDownLinkPort)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalRouterDownLinkPort update: %v", err)
	}

	return resourceNsxtLogicalRouterDownLinkPortRead(d, m)
}

func resourceNsxtLogicalRouterDownLinkPortDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical router downlink port id while deleting")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouterPort(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LogicalRouterDownLinkPort delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalRouterDownLinkPort %s not found", id)
		d.SetId("")
	}

	return nil
}
