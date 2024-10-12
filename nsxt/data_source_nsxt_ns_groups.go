// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtNsGroups() *schema.Resource {
	return &schema.Resource{
		Read:               dataSourceNsxtNsGroupsRead,
		DeprecationMessage: mpObjectDataSourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"items": {
				Type:        schema.TypeMap,
				Description: "Mapping of group UUID by display name",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceNsxtNsGroupsRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return dataSourceNotSupportedError()
	}

	// Get by full name
	groupMap := make(map[string]string)
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
			groupMap[objInList.DisplayName] = objInList.Id
		}
		return nil
	}

	_, err := handlePagination(lister)
	if err != nil {
		return err
	}

	d.SetId(newUUID())
	d.Set("items", groupMap)

	return nil
}
