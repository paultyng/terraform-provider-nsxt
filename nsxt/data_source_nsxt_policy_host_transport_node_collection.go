// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyHostTransportNodeCollection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyHostTransportNodeCollectionRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path of the site this Transport Node Collection belongs to",
				Optional:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"transport_node_profile_id": {
				Type:        schema.TypeString,
				Description: "Transport Node Profile Path",
				Optional:    true,
				Computed:    true,
			},
			"unique_id": {
				Type:        schema.TypeString,
				Description: "A unique identifier assigned by the system",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyHostTransportNodeCollectionRead(d *schema.ResourceData, m interface{}) error {
	if isPolicyGlobalManager(m) {
		return localManagerOnlyError()
	}

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	objSitePath := d.Get("site_path").(string)
	// For local manager, if site path is not provided, use default site
	if objSitePath == "" {
		objSitePath = defaultSite
	}
	connector := getPolicyConnector(m)
	client := enforcement_points.NewTransportNodeCollectionsClient(connector)

	var obj model.HostTransportNodeCollection
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objSitePath, getPolicyEnforcementPoint(m), objID)

		if err != nil {
			return handleDataSourceReadError(d, "HostTransportNodeCollection", objID, err)
		}
		obj = objGet
	} else {
		// Get by full name/prefix, or the existing transport node collection if name is not provided
		objList, err := client.List(objSitePath, getPolicyEnforcementPoint(m), nil, nil, nil)
		if err != nil {
			return handleListError("HostTransportNodeCollection", err)
		}
		for _, objInList := range objList.Results {
			if *objInList.DisplayName == objName || strings.HasPrefix(*objInList.DisplayName, objName) || objName == "" {
				obj = objInList
			}
		}
		if obj.Id == nil {
			return fmt.Errorf("HostTransportNodeCollection '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	d.Set("site_path", objSitePath)
	d.Set("transport_node_profile_id", obj.TransportNodeProfileId)
	d.Set("unique_id", obj.UniqueId)
	return nil
}
