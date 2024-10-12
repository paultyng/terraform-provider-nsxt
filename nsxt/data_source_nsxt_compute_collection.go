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

func dataSourceNsxtComputeCollection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtComputeCollectionRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDisplayNameSchema(),
			"origin_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ComputeCollection type like VC_Cluster. Here the Compute Manager type prefix would help in differentiating similar named Compute Collection types from different Compute Managers",
			},
			"origin_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Id of the compute manager from where this Compute Collection was discovered",
			},
			"cm_local_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Local Id of the compute collection in the Compute Manager",
			},
		},
	}
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func dataSourceNsxtComputeCollectionRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := fabric.NewComputeCollectionsClient(connector)
	objID := d.Get("id").(string)

	objName := nullIfEmpty(d.Get("display_name").(string))
	objOrigin := nullIfEmpty(d.Get("origin_id").(string))
	objType := nullIfEmpty(d.Get("origin_type").(string))
	objLocalID := nullIfEmpty(d.Get("cm_local_id").(string))

	var obj model.ComputeCollection

	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)
		if err != nil {
			return fmt.Errorf("failed to read ComputeCollection %s: %v", objID, err)
		}
		obj = objGet
	} else {
		objList, err := client.List(objLocalID, nil, nil, objName, nil, nil, nil, objOrigin, objType, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to read Compute Collections: %v", err)
		} else if *objList.ResultCount == 0 {
			return fmt.Errorf("no Compute Collections that matched the specified parameters found")
		} else if *objList.ResultCount > 1 {
			return fmt.Errorf("found multiple Compute Collections that matched specified parameters")
		}
		obj = objList.Results[0]
	}

	d.SetId(*obj.ExternalId)
	d.Set("display_name", obj.DisplayName)
	d.Set("origin_type", obj.OriginType)
	d.Set("origin_id", obj.OriginId)
	d.Set("cm_local_id", obj.CmLocalId)

	return nil
}
