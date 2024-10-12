// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/fabric"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func dataSourceNsxtDiscoveredNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtDiscoveredNodeRead,
		Schema: map[string]*schema.Schema{
			"compute_manager_state": {
				Type:        schema.TypeString,
				Description: "State of compute manager realization",
				Optional:    true,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "External id of the discovered node, ex. a mo-ref from VC",
				Optional:    true,
				Computed:    true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "IP Address of the the discovered node.",
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateCidrOrIPOrRange(),
			},
		},
	}
}

func dataSourceNsxtDiscoveredNodeRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	discoveredNodeClient := fabric.NewDiscoveredNodesClient(connector)

	objID := d.Get("id").(string)
	ipAddress := d.Get("ip_address").(string)

	var obj model.DiscoveredNode
	if objID != "" {
		// Get by ID
		objGet, err := discoveredNodeClient.Get(objID)
		if isNotFoundError(err) {
			return fmt.Errorf("Discovered Node %s was not found", objID)
		}
		if err != nil {
			return fmt.Errorf("Error while reading Discovered Node %s: %v", objID, err)
		}
		obj = objGet
	} else if ipAddress == "" {
		return fmt.Errorf("Error obtaining Discovered Node external ID or IP address during read")
	} else {
		// Get by IP address
		objList, err := discoveredNodeClient.List(nil, nil, nil, nil, nil, nil, &ipAddress, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("Error while reading Discovered Node: %v", err)
		}
		if len(objList.Results) == 0 {
			return fmt.Errorf("Discovered Node with IP %s was not found", ipAddress)
		}
		obj = objList.Results[0]
	}

	d.SetId(*obj.ExternalId)
	if len(obj.IpAddresses) > 0 {
		d.Set("ip_address", obj.IpAddresses[0])
	}
	return nil
}
