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

func dataSourceNsxtNsGroup() *schema.Resource {
	return &schema.Resource{
		Read:               dataSourceNsxtNsGroupRead,
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

func dataSourceNsxtNsGroupRead(d *schema.ResourceData, m interface{}) error {
	// Read NS Group by name or id
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return dataSourceNotSupportedError()
	}

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj manager.NsGroup
	if objID != "" {
		// Get by id
		localVarOptionals := make(map[string]interface{})
		localVarOptionals["populateReferences"] = true
		objGet, resp, err := nsxClient.GroupingObjectsApi.ReadNSGroup(nsxClient.Context, objID, localVarOptionals)

		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("NS group %s was not found", objID)
		}
		if err != nil {
			return fmt.Errorf("Error while reading NS group %s: %v", objID, err)
		}
		obj = objGet
	} else if objName != "" {
		// Get by full name
		found := false
		lister := func(info *paginationInfo) error {
			objList, _, err := nsxClient.GroupingObjectsApi.ListNSGroups(nsxClient.Context, info.LocalVarOptionals)
			if err != nil {
				return fmt.Errorf("Error while reading NS groups: %v", err)
			}
			info.PageCount = int64(len(objList.Results))
			info.TotalCount = objList.ResultCount
			info.Cursor = objList.Cursor

			// go over the list to find the correct one
			for _, objInList := range objList.Results {
				if objInList.DisplayName == objName {
					if found {
						return fmt.Errorf("Found multiple NS groups with name '%s'", objName)
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
			return fmt.Errorf("NS group with name '%s' was not found among %d groups", objName, total)
		}
	} else {
		return fmt.Errorf("Error obtaining NS group ID or name during read")
	}

	d.SetId(obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
