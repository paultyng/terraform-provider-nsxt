// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceNsxtPolicyEdgeCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyEdgeClusterRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path of the site this Edge cluster belongs to",
				Optional:     true,
				ValidateFunc: validatePolicyPath(),
			},
		},
	}
}

func dataSourceNsxtPolicyEdgeClusterRead(d *schema.ResourceData, m interface{}) error {
	// Read an edge cluster by name or id
	objSitePath := d.Get("site_path").(string)

	if !isPolicyGlobalManager(m) && objSitePath != "" {
		return globalManagerOnlyError()
	}
	if isPolicyGlobalManager(m) {
		if objSitePath == "" {
			return attributeRequiredGlobalManagerError("site_path", "nsxt_policy_edge_cluster")
		}

		query := make(map[string]string)
		globalPolicyEnforcementPointPath := getGlobalPolicyEnforcementPointPath(m, &objSitePath)
		query["parent_path"] = globalPolicyEnforcementPointPath
		_, err := policyDataSourceResourceReadWithValidation(d, getPolicyConnector(m), getSessionContext(d, m), "PolicyEdgeCluster", query, false)
		if err != nil {
			return err
		}
		return nil
	}

	// Local manager
	connector := getPolicyConnector(m)
	_, err := policyDataSourceResourceRead(d, connector, getSessionContext(d, m), "PolicyEdgeCluster", nil)
	if err != nil {
		return err
	}
	return nil
}
