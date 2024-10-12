// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/cluster"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func dataSourceNsxtManagerClusterNode() *schema.Resource {
	return &schema.Resource{
		Read:               dataSourceNsxtManagerClusterNodeRead,
		DeprecationMessage: mpObjectDataSourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Unique ID of this resource",
				Optional:    true,
				Computed:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource",
				Optional:    true,
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
				Computed:    true,
			},
			"appliance_mgmt_listen_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The IP and port for the appliance management API service on this node",
			},
		},
	}
}

func dataSourceNsxtManagerClusterNodeRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := cluster.NewNodesClient(connector)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj model.ClusterNodeConfig
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)
		if err != nil {
			return handleDataSourceReadError(d, "ClusterNode", objID, err)
		}
		obj = objGet
	} else if objName == "" {
		return fmt.Errorf("error obtaining ClusterNode ID or name during read")
	} else {
		// Get by full name/prefix
		objList, err := client.List(nil, nil, nil, nil, nil)
		if err != nil {
			return handleListError("ClusterNode", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.ClusterNodeConfig
		var prefixMatch []model.ClusterNodeConfig
		for _, objInList := range objList.Results {
			if strings.HasPrefix(*objInList.DisplayName, objName) {
				prefixMatch = append(prefixMatch, objInList)
			}
			if *objInList.DisplayName == objName {
				perfectMatch = append(perfectMatch, objInList)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("found multiple ClusterNode with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("found multiple ClusterNodes with name starting with '%s'", objName)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("ClusterNode with name '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("appliance_mgmt_listen_address", obj.ApplianceMgmtListenAddr)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
