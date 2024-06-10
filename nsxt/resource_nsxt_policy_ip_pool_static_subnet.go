/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ippools "github.com/vmware/terraform-provider-nsxt/api/infra/ip_pools"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyIPPoolStaticSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPPoolStaticSubnetCreate,
		Read:   resourceNsxtPolicyIPPoolStaticSubnetRead,
		Update: resourceNsxtPolicyIPPoolStaticSubnetUpdate,
		Delete: resourceNsxtPolicyIPPoolStaticSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyIPPoolSubnetImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":           getNsxIDSchema(),
			"path":             getPathSchema(),
			"display_name":     getDisplayNameSchema(),
			"description":      getDescriptionSchema(),
			"revision":         getRevisionSchema(),
			"tag":              getTagsSchema(),
			"context":          getContextSchema(false, false, false),
			"pool_path":        getPolicyPathSchema(true, true, "Policy path to the IP Pool for this Subnet"),
			"allocation_range": getAllocationRangeListSchema(true, "A collection of IPv4 or IPv6 IP ranges"),
			"cidr": {
				Type:         schema.TypeString,
				Description:  "Network address and prefix length",
				Required:     true,
				ValidateFunc: validateCidr(),
			},
			"dns_nameservers": {
				Type:        schema.TypeList,
				Description: "The collection of up to 3 DNS servers for the subnet",
				Optional:    true,
				MaxItems:    3,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Description: "DNS suffix for the nameserver",
				Optional:    true,
				// TODO: validate hostname
			},
			"gateway": {
				Type:         schema.TypeString,
				Description:  "The default gateway address",
				Optional:     true,
				ValidateFunc: validateSingleIP(),
			},
		},
	}
}

func resourceNsxtPolicyIPPoolStaticSubnetSchemaToStructValue(d *schema.ResourceData, id string) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	cidr := d.Get("cidr").(string)
	dnsNameservers := interfaceListToStringList(d.Get("dns_nameservers").([]interface{}))
	dnsSuffix := d.Get("dns_suffix").(string)
	gateway := d.Get("gateway").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.IpAddressPoolStaticSubnet{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		ResourceType: "IpAddressPoolStaticSubnet",
		Cidr:         &cidr,
		Id:           &id,
	}

	// attributes that should only be set if they have a value specified
	if dnsSuffix != "" {
		obj.DnsSuffix = &dnsSuffix
	}
	if gateway != "" {
		obj.GatewayIp = &gateway
	}
	if len(dnsNameservers) > 0 {
		obj.DnsNameservers = dnsNameservers
	}

	var poolRanges []model.IpPoolRange
	for _, allocRange := range d.Get("allocation_range").([]interface{}) {
		allocMap := allocRange.(map[string]interface{})
		start := allocMap["start"].(string)
		end := allocMap["end"].(string)
		ipRange := model.IpPoolRange{
			Start: &start,
			End:   &end,
		}
		poolRanges = append(poolRanges, ipRange)
	}
	obj.AllocationRanges = poolRanges

	dataValue, errors := converter.ConvertToVapi(obj, model.IpAddressPoolStaticSubnetBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting Static Subnet: %v", errors[0])
	}

	return dataValue.(*data.StructValue), nil
}

func resourceNsxtPolicyIPPoolStaticSubnetRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)
	converter := bindings.NewTypeConverter()

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Static Subnet ID")
	}

	subnetData, err := client.Get(poolID, id)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			log.Printf("[DEBUG] Static Subnet %s not found", id)
			return nil
		}
		return handleReadError(d, "Static Subnet", id, err)
	}

	snet, errs := converter.ConvertToGolang(subnetData, model.IpAddressPoolStaticSubnetBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting Static Subnet %s", errs[0])
	}
	staticSubnet := snet.(model.IpAddressPoolStaticSubnet)

	d.Set("display_name", staticSubnet.DisplayName)
	d.Set("description", staticSubnet.Description)
	setPolicyTagsInSchema(d, staticSubnet.Tags)
	d.Set("nsx_id", staticSubnet.Id)
	d.Set("path", staticSubnet.Path)
	d.Set("revision", staticSubnet.Revision)
	d.Set("pool_path", poolPath)
	d.Set("cidr", staticSubnet.Cidr)
	d.Set("dns_nameservers", staticSubnet.DnsNameservers)
	d.Set("dns_suffix", staticSubnet.DnsSuffix)
	d.Set("gateway", staticSubnet.GatewayIp)

	var allocations []map[string]interface{}
	for _, allocRange := range staticSubnet.AllocationRanges {
		allocMap := make(map[string]interface{})
		allocMap["start"] = allocRange.Start
		allocMap["end"] = allocRange.End
		allocations = append(allocations, allocMap)
	}
	d.Set("allocation_range", allocations)

	return nil
}

func resourceNsxtPolicyIPPoolStaticSubnetCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := client.Get(poolID, id)
		if err == nil {
			return fmt.Errorf("Static Subnet with ID '%s' already exists on Pool %s", id, poolID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	dataValue, err := resourceNsxtPolicyIPPoolStaticSubnetSchemaToStructValue(d, id)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating IP Pool Static Subnet with ID %s", id)
	err = client.Patch(poolID, id, dataValue)
	if err != nil {
		return handleCreateError("Static Subnet", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyIPPoolStaticSubnetRead(d, m)
}

func resourceNsxtPolicyIPPoolStaticSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Static Subnet ID")
	}

	dataValue, err := resourceNsxtPolicyIPPoolStaticSubnetSchemaToStructValue(d, id)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating IP Pool Static Subnet with ID %s", id)
	err = client.Patch(poolID, id, dataValue)
	if err != nil {
		return handleUpdateError("Static Subnet", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyIPPoolStaticSubnetRead(d, m)
}

func resourceNsxtPolicyIPPoolStaticSubnetDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Static Subnet ID")
	}

	log.Printf("[INFO] Deleting Static Subnet with ID %s", id)
	err = client.Delete(poolID, id, nil)
	if err != nil {
		return handleDeleteError("Static Subnet", id, err)
	}

	return nil
}
