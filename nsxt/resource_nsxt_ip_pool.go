// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtIPPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIPPoolCreate,
		Read:   resourceNsxtIPPoolRead,
		Update: resourceNsxtIPPoolUpdate,
		Delete: resourceNsxtIPPoolDelete,
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
			"tag":    getTagsSchema(),
			"subnet": getSubnetSchema(),
		},
	}
}

func getSubnetSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of IPv4 subnets",
		Optional:    true,
		MaxItems:    5,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"allocation_ranges": {
					Type:        schema.TypeList,
					Description: "A collection of IPv4 Pool Ranges",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateIPRange(),
					},
					Required: true,
				},
				"cidr": {
					Type:         schema.TypeString,
					Description:  "Network address and the prefix length which will be associated with a layer-2 broadcast domain",
					Required:     true,
					ValidateFunc: validateCidr(),
				},
				"dns_nameservers": {
					Type:        schema.TypeList,
					Description: "A collection of DNS servers for the subnet",
					Optional:    true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateSingleIP(),
					},
					MaxItems: 3,
				},
				"dns_suffix": {
					Type:        schema.TypeString,
					Description: "The DNS suffix for the DNS server",
					Optional:    true,
				},
				"gateway_ip": {
					Type:         schema.TypeString,
					Description:  "The default gateway address on a layer-3 router",
					Optional:     true,
					ValidateFunc: validateSingleIP(),
				},
			},
		},
	}
}

func getAllocationRangesFromRanges(ranges []interface{}) []manager.IpPoolRange {
	var allocationRanges []manager.IpPoolRange
	for _, rng := range ranges {
		r := rng.(string)
		s := strings.Split(r, "-")
		start := s[0]
		end := s[1]
		elem := manager.IpPoolRange{
			Start: start,
			End:   end,
		}
		allocationRanges = append(allocationRanges, elem)
	}
	return allocationRanges
}

func getSubnetsFromSchema(d *schema.ResourceData) []manager.IpPoolSubnet {
	subnets := d.Get("subnet").([]interface{})
	var subnetsList []manager.IpPoolSubnet
	for _, subnet := range subnets {
		data := subnet.(map[string]interface{})
		elem := manager.IpPoolSubnet{
			Cidr:             data["cidr"].(string),
			DnsSuffix:        data["dns_suffix"].(string),
			GatewayIp:        data["gateway_ip"].(string),
			DnsNameservers:   interface2StringList(data["dns_nameservers"].([]interface{})),
			AllocationRanges: getAllocationRangesFromRanges(data["allocation_ranges"].([]interface{})),
		}

		subnetsList = append(subnetsList, elem)
	}
	return subnetsList
}

func getRangesFromAllocationRanges(ranges []manager.IpPoolRange) []string {
	var rangesList []string
	for _, r := range ranges {
		elem := fmt.Sprintf("%s-%s", r.Start, r.End)
		rangesList = append(rangesList, elem)
	}
	return rangesList
}

func setSubnetsInSchema(d *schema.ResourceData, subnets []manager.IpPoolSubnet) error {
	subnetsList := make([]map[string]interface{}, 0, len(subnets))
	for _, subnet := range subnets {
		elem := make(map[string]interface{})
		elem["cidr"] = subnet.Cidr
		elem["gateway_ip"] = subnet.GatewayIp
		elem["dns_suffix"] = subnet.DnsSuffix
		elem["dns_nameservers"] = stringList2Interface(subnet.DnsNameservers)
		elem["allocation_ranges"] = getRangesFromAllocationRanges(subnet.AllocationRanges)
		subnetsList = append(subnetsList, elem)
	}
	err := d.Set("subnet", subnetsList)
	return err
}

func resourceNsxtIPPoolCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	displayName := d.Get("display_name").(string)
	subnets := getSubnetsFromSchema(d)
	description := d.Get("description").(string)
	tags := getTagsFromSchema(d)
	ipPool := manager.IpPool{
		DisplayName: displayName,
		Description: description,
		Subnets:     subnets,
		Tags:        tags,
	}

	ipPool, resp, err := nsxClient.PoolManagementApi.CreateIpPool(nsxClient.Context, ipPool)

	if err != nil {
		return fmt.Errorf("Error during IpPool create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during IpPool create: %v", resp.StatusCode)
	}
	d.SetId(ipPool.Id)

	return resourceNsxtIPPoolRead(d, m)
}

func resourceNsxtIPPoolRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	ipPool, resp, err := nsxClient.PoolManagementApi.ReadIpPool(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpPool %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during IpPool read: %v", err)
	}

	d.Set("display_name", ipPool.DisplayName)
	d.Set("description", ipPool.Description)
	d.Set("revision", ipPool.Revision)
	setTagsInSchema(d, ipPool.Tags)
	err = setSubnetsInSchema(d, ipPool.Subnets)
	if err != nil {
		return fmt.Errorf("Error during IpPool set in schema: %v", err)
	}

	return nil
}

func resourceNsxtIPPoolUpdate(d *schema.ResourceData, m interface{}) error {
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
	subnets := getSubnetsFromSchema(d)
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	ipPool := manager.IpPool{
		DisplayName: displayName,
		Description: description,
		Subnets:     subnets,
		Tags:        tags,
		Revision:    revision,
	}

	_, resp, err := nsxClient.PoolManagementApi.UpdateIpPool(nsxClient.Context, id, ipPool)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during IpPool update: %v", err)
	}

	return resourceNsxtIPPoolRead(d, m)
}

func resourceNsxtIPPoolDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.PoolManagementApi.DeleteIpPool(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during IpPool delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpPool %s not found", id)
		d.SetId("")
	}
	return nil
}
