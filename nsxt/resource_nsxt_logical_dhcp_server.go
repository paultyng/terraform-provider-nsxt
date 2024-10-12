// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtLogicalDhcpServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalDhcpServerCreate,
		Read:   resourceNsxtLogicalDhcpServerRead,
		Update: resourceNsxtLogicalDhcpServerUpdate,
		Delete: resourceNsxtLogicalDhcpServerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"attached_logical_port_id": {
				Type:        schema.TypeString,
				Description: "Id of attached logical port",
				Computed:    true,
			},
			"dhcp_profile_id": {
				Type:        schema.TypeString,
				Description: "DHCP profile uuid",
				Required:    true,
			},
			"dhcp_server_ip": {
				Type:         schema.TypeString,
				Description:  "DHCP server ip in cidr format",
				Required:     true,
				ValidateFunc: validatePortAddress(),
			},
			"gateway_ip": {
				Type:         schema.TypeString,
				Description:  "Gateway IP",
				Optional:     true,
				ValidateFunc: validateSingleIP(),
			},
			"domain_name": {
				Type:        schema.TypeString,
				Description: "Domain name",
				Optional:    true,
			},
			"dns_name_servers": {
				Type:        schema.TypeList,
				Description: "DNS IPs",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
			},
			"dhcp_option_121":     getDhcpOptions121Schema(),
			"dhcp_generic_option": getDhcpGenericOptionsSchema(),
			"tag":                 getTagsSchema(),
			"revision":            getRevisionSchema(),
		},
	}
}

func getDhcpOptions121Schema() *schema.Schema {

	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "DHCP classless static routes",
		Optional:    true,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"network": {
					Type:         schema.TypeString,
					Description:  "Destination in cidr",
					Required:     true,
					ValidateFunc: validateIPCidr(),
				},
				"next_hop": {
					Type:         schema.TypeString,
					Description:  "Next hop IP",
					Required:     true,
					ValidateFunc: validateSingleIP(),
				},
			},
		},
	}
}

func getDhcpGenericOptionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Generic DHCP options",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"code": {
					Type:         schema.TypeInt,
					Description:  "DHCP option code, [0-255]",
					Required:     true,
					ValidateFunc: validation.IntBetween(0, 255),
				},
				"values": {
					Type:        schema.TypeList,
					Description: "DHCP option values",
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func getDhcpOptions121(d *schema.ResourceData) []manager.ClasslessStaticRoute {
	var routes []manager.ClasslessStaticRoute
	configured121 := d.Get("dhcp_option_121").([]interface{})
	for _, opt121 := range configured121 {
		data := opt121.(map[string]interface{})
		elem := manager.ClasslessStaticRoute{
			Network: data["network"].(string),
			NextHop: data["next_hop"].(string),
		}
		routes = append(routes, elem)
	}
	return routes
}

func setDhcpOptions121InSchema(d *schema.ResourceData, routes []manager.ClasslessStaticRoute) error {
	var dhcpOpt121 []map[string]interface{}
	for _, route := range routes {
		elem := make(map[string]interface{})
		elem["network"] = route.Network
		elem["next_hop"] = route.NextHop
		dhcpOpt121 = append(dhcpOpt121, elem)
	}
	err := d.Set("dhcp_option_121", dhcpOpt121)
	return err
}

func getDhcpGenericOptions(d *schema.ResourceData) []manager.GenericDhcpOption {
	var options []manager.GenericDhcpOption
	configuredOptions := d.Get("dhcp_generic_option").([]interface{})
	for _, opt := range configuredOptions {
		data := opt.(map[string]interface{})
		elem := manager.GenericDhcpOption{
			Code:   int64(data["code"].(int)),
			Values: interface2StringList(data["values"].([]interface{})),
		}
		options = append(options, elem)
	}
	return options
}

func setDhcpGenericOptionsInSchema(d *schema.ResourceData, opts []manager.GenericDhcpOption) error {
	var dhcpOptions []map[string]interface{}
	for _, opt := range opts {
		elem := make(map[string]interface{})
		elem["code"] = opt.Code
		elem["values"] = opt.Values
		dhcpOptions = append(dhcpOptions, elem)
	}
	err := d.Set("dhcp_generic_option", dhcpOptions)
	return err
}

func resourceNsxtLogicalDhcpServerCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	dhcpProfileID := d.Get("dhcp_profile_id").(string)
	opt121Routes := getDhcpOptions121(d)
	var opt121 *manager.DhcpOption121
	if opt121Routes != nil {
		opt121 = &manager.DhcpOption121{
			StaticRoutes: opt121Routes,
		}
	}
	ipv4DhcpServer := manager.IPv4DhcpServer{
		DhcpServerIp:   d.Get("dhcp_server_ip").(string),
		DnsNameservers: interface2StringList(d.Get("dns_name_servers").([]interface{})),
		DomainName:     d.Get("domain_name").(string),
		GatewayIp:      d.Get("gateway_ip").(string),
		Options: &manager.DhcpOptions{
			Option121: opt121,
			Others:    getDhcpGenericOptions(d),
		},
	}
	tags := getTagsFromSchema(d)
	logicalDhcpServer := manager.LogicalDhcpServer{
		DisplayName:    displayName,
		Description:    description,
		DhcpProfileId:  dhcpProfileID,
		Ipv4DhcpServer: &ipv4DhcpServer,
		Tags:           tags,
	}

	logicalDhcpServer, resp, err := nsxClient.ServicesApi.CreateDhcpServer(nsxClient.Context, logicalDhcpServer)

	if err != nil {
		return fmt.Errorf("Error during LogicalDhcpServer create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during LogicalDhcpServer create: %v", resp.StatusCode)
	}
	d.SetId(logicalDhcpServer.Id)

	return resourceNsxtLogicalDhcpServerRead(d, m)
}

func resourceNsxtLogicalDhcpServerRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	logicalDhcpServer, resp, err := nsxClient.ServicesApi.ReadDhcpServer(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalDhcpServer %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalDhcpServer read: %v", err)
	}

	d.Set("revision", logicalDhcpServer.Revision)
	d.Set("description", logicalDhcpServer.Description)
	d.Set("display_name", logicalDhcpServer.DisplayName)
	setTagsInSchema(d, logicalDhcpServer.Tags)
	d.Set("attached_logical_port_id", logicalDhcpServer.AttachedLogicalPortId)
	d.Set("dhcp_profile_id", logicalDhcpServer.DhcpProfileId)
	d.Set("dhcp_server_ip", logicalDhcpServer.Ipv4DhcpServer.DhcpServerIp)
	d.Set("domain_name", logicalDhcpServer.Ipv4DhcpServer.DomainName)
	d.Set("gateway_ip", logicalDhcpServer.Ipv4DhcpServer.GatewayIp)
	d.Set("dns_name_servers", logicalDhcpServer.Ipv4DhcpServer.DnsNameservers)
	if logicalDhcpServer.Ipv4DhcpServer.Options != nil && logicalDhcpServer.Ipv4DhcpServer.Options.Option121 != nil {
		err = setDhcpOptions121InSchema(d, logicalDhcpServer.Ipv4DhcpServer.Options.Option121.StaticRoutes)
		if err != nil {
			return fmt.Errorf("Error during LogicalDhcpServer read option 121: %v", err)
		}
	} else {
		var emptyDhcpOpt121 []map[string]interface{}
		d.Set("dhcp_option_121", emptyDhcpOpt121)
	}
	if logicalDhcpServer.Ipv4DhcpServer.Options != nil {
		err = setDhcpGenericOptionsInSchema(d, logicalDhcpServer.Ipv4DhcpServer.Options.Others)
		if err != nil {
			return fmt.Errorf("Error during LogicalDhcpServer read generic options: %v", err)
		}
	} else {
		var emptyDhcpGenOpt []map[string]interface{}
		d.Set("dhcp_generic_option", emptyDhcpGenOpt)
	}

	return nil
}

func resourceNsxtLogicalDhcpServerUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getTagsFromSchema(d)
	dhcpProfileID := d.Get("dhcp_profile_id").(string)
	revision := int64(d.Get("revision").(int))
	opt121Routes := getDhcpOptions121(d)
	var opt121 *manager.DhcpOption121
	if opt121Routes != nil {
		opt121 = &manager.DhcpOption121{
			StaticRoutes: opt121Routes,
		}
	}
	ipv4DhcpServer := manager.IPv4DhcpServer{
		DhcpServerIp:   d.Get("dhcp_server_ip").(string),
		DnsNameservers: interface2StringList(d.Get("dns_name_servers").([]interface{})),
		DomainName:     d.Get("domain_name").(string),
		GatewayIp:      d.Get("gateway_ip").(string),
		Options: &manager.DhcpOptions{
			Option121: opt121,
			Others:    getDhcpGenericOptions(d),
		},
	}
	logicalDhcpServer := manager.LogicalDhcpServer{
		DisplayName:    displayName,
		Description:    description,
		DhcpProfileId:  dhcpProfileID,
		Ipv4DhcpServer: &ipv4DhcpServer,
		Tags:           tags,
		Revision:       revision,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateDhcpServer(nsxClient.Context, id, logicalDhcpServer)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalDhcpServer update: %v", err)
	}

	return resourceNsxtLogicalDhcpServerRead(d, m)
}

func resourceNsxtLogicalDhcpServerDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteDhcpServer(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LogicalDhcpServer delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalDhcpServer %s not found", id)
		d.SetId("")
	}
	return nil
}
