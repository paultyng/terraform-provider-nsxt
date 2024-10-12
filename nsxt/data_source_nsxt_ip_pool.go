// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func dataSourceNsxtIPPool() *schema.Resource {
	return &schema.Resource{
		Read:               dataSourceNsxtIPPoolRead,
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
		},
	}
}

func dataSourceNsxtIPPoolRead(d *schema.ResourceData, m interface{}) error {
	// Read IP Pool by name or id
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return dataSourceNotSupportedError()
	}
	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj manager.IpPool
	if objID != "" {
		// Get by id
		objGet, resp, err := nsxClient.PoolManagementApi.ReadIpPool(nsxClient.Context, objID)

		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("IP pool %s was not found", objID)
		}
		if err != nil {
			return fmt.Errorf("Error while reading ns service %s: %v", objID, err)
		}
		obj = objGet
	} else if objName != "" {
		// Get by full name
		found := false
		lister := func(info *paginationInfo) error {
			objList, _, err := nsxClient.PoolManagementApi.ListIpPools(nsxClient.Context, info.LocalVarOptionals)
			if err != nil {
				return fmt.Errorf("Error while reading IP pool: %v", err)
			}

			info.PageCount = int64(len(objList.Results))
			info.TotalCount = objList.ResultCount
			info.Cursor = objList.Cursor

			// go over the list to find the correct one
			for _, objInList := range objList.Results {
				if objInList.DisplayName == objName {
					if found {
						return fmt.Errorf("Found multiple IP pool with name '%s'", objName)
					}
					obj = objInList
					found = true
				}
			}
			return nil
		}

		total, err := handlePagination(lister)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("IP pool '%s' was not found out of %d objects", objName, total)
		}
	} else {
		return fmt.Errorf("Error obtaining IP pool ID or name during read")
	}

	d.SetId(obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
