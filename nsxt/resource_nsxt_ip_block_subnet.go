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

func resourceNsxtIPBlockSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtIPBlockSubnetCreate,
		Read:   resourceNsxtIPBlockSubnetRead,
		// Update IP block subnet is not supported by the NSX
		Delete: resourceNsxtIPBlockSubnetDelete,
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
				ForceNew:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
			"tag": getTagsSchemaForceNew(),
			"block_id": {
				Type:        schema.TypeString,
				Description: "Block id for which the subnet is created",
				Required:    true,
				ForceNew:    true,
			},
			"size": {
				Type:         schema.TypeInt,
				Description:  "Represents the size or number of ip addresses in the subnet",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePowerOf2(false, 0),
			},
			"cidr": {
				Type:        schema.TypeString,
				Description: "Represents network address and the prefix length which will be associated with a layer-2 broadcast domain",
				Computed:    true,
			},
			"allocation_ranges": {
				Type:        schema.TypeList,
				Description: "A collection of IPv4 Pool Ranges",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start": {
							Type:        schema.TypeString,
							Description: "The start IP Address of the IP Range",
							Computed:    true,
						},
						"end": {
							Type:        schema.TypeString,
							Description: "The end IP Address of the IP Range",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func resourceNsxtIPBlockSubnetCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	blockID := d.Get("block_id").(string)
	size := int64(d.Get("size").(int))
	tags := getTagsFromSchema(d)
	ipBlockSubnet := manager.IpBlockSubnet{
		DisplayName: displayName,
		Description: description,
		BlockId:     blockID,
		Size:        size,
		Tags:        tags,
	}

	ipBlockSubnet, resp, err := nsxClient.PoolManagementApi.CreateIpBlockSubnet(nsxClient.Context, ipBlockSubnet)

	if err != nil {
		return fmt.Errorf("Error during IpBlockSubnet create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during IpBlockSubnet create: %v", resp.StatusCode)
	}
	d.SetId(ipBlockSubnet.Id)

	return resourceNsxtIPBlockSubnetRead(d, m)
}

func setAllocationRangesInSchema(d *schema.ResourceData, ranges []manager.IpPoolRange) error {
	allocationRanges := make([]map[string]interface{}, 0, len(ranges))
	for _, rng := range ranges {
		elem := make(map[string]interface{})
		elem["start"] = rng.Start
		elem["end"] = rng.End
		allocationRanges = append(allocationRanges, elem)
	}
	err := d.Set("allocation_ranges", allocationRanges)
	return err
}

func resourceNsxtIPBlockSubnetRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	ipBlockSubnet, resp, err := nsxClient.PoolManagementApi.ReadIpBlockSubnet(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpBlockSubnet %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during IpBlockSubnet read: %v", err)
	}

	d.Set("display_name", ipBlockSubnet.DisplayName)
	d.Set("description", ipBlockSubnet.Description)
	d.Set("block_id", ipBlockSubnet.BlockId)
	d.Set("size", ipBlockSubnet.Size)
	setTagsInSchema(d, ipBlockSubnet.Tags)
	err = setAllocationRangesInSchema(d, ipBlockSubnet.AllocationRanges)
	if err != nil {
		return fmt.Errorf("Error during IpBlockSubnet allocation ranges set in schema: %v", err)
	}
	d.Set("cidr", ipBlockSubnet.Cidr)
	d.Set("revision", ipBlockSubnet.Revision)

	return nil
}

func resourceNsxtIPBlockSubnetDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.PoolManagementApi.DeleteIpBlockSubnet(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during IpBlockSubnet delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] IpBlockSubnet %s not found", id)
		d.SetId("")
	}
	return nil
}
