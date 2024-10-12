// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs"
)

const defaultOrgID = "default"

func dataSourceNsxtPolicyProject() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyProjectRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"short_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_info": getSiteInfoSchema(),
			"tier0_gateway_paths": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
			"external_ipv4_blocks": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Computed: true,
			},
		},
	}
}

func getSiteInfoSchema() *schema.Schema {
	return &schema.Schema{
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"edge_cluster_paths": {
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				},
				"site_path": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
		Optional: true,
	}
}
func dataSourceNsxtPolicyProjectRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewProjectsClient(connector)

	// As Project resource type paths reside under project and not under /infra or /global_infra or such, and since
	// this data source fetches extra attributes, e.g site_info and tier0_gateway_paths, it's simpler to implement using .List()
	// instead of using search API.

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj model.Project
	if objID != "" {
		// Get by id
		objGet, err := client.Get(defaultOrgID, objID, nil)
		if err != nil {
			return handleDataSourceReadError(d, "Project", objID, err)
		}
		obj = objGet
	} else if objName == "" {
		return fmt.Errorf("Error obtaining Project ID or name during read")
	} else {
		// Get by full name/prefix
		objList, err := client.List(defaultOrgID, nil, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return handleListError("Project", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.Project
		var prefixMatch []model.Project
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
				return fmt.Errorf("Found multiple Project with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple Projects with name starting with '%s'", objName)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("Project with name '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)

	var siteInfosList []map[string]interface{}
	for _, item := range obj.SiteInfos {
		data := make(map[string]interface{})
		data["edge_cluster_paths"] = item.EdgeClusterPaths
		data["site_path"] = item.SitePath
		siteInfosList = append(siteInfosList, data)
	}
	d.Set("site_info", siteInfosList)
	d.Set("tier0_gateway_paths", obj.Tier0s)
	d.Set("external_ipv4_blocks", obj.ExternalIpv4Blocks)

	return nil
}
