// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points/edge_clusters"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyEdgeNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyEdgeNodeRead,

		Schema: map[string]*schema.Schema{
			"edge_cluster_path": getPolicyPathSchema(true, false, "Edge cluster Path"),
			"member_index": {
				Type:         schema.TypeInt,
				Description:  "Index of this node within edge cluster",
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(0),
			},
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func dataSourceNsxtPolicyEdgeNodeRead(d *schema.ResourceData, m interface{}) error {
	// Read an edge node by name or id
	edgeClusterPath := d.Get("edge_cluster_path").(string)
	// Note - according to the documentation GetOkExists should be used
	// for bool types, but in this case it works and GetOk doesn't
	memberIndex, memberIndexSet := d.GetOkExists("member_index")

	if isPolicyGlobalManager(m) || util.NsxVersionHigherOrEqual("3.2.0") {
		query := make(map[string]string)
		query["parent_path"] = edgeClusterPath
		if memberIndexSet {
			query["member_index"] = strconv.Itoa(memberIndex.(int))
		}
		obj, err := policyDataSourceResourceReadWithValidation(d, getPolicyConnector(m), getSessionContext(d, m), "PolicyEdgeNode", query, false)
		if err != nil {
			return err
		}
		converter := bindings.NewTypeConverter()
		dataValue, errors := converter.ConvertToGolang(obj, model.PolicyEdgeNodeBindingType())
		if len(errors) > 0 {
			return errors[0]
		}
		policyEdgeNode := dataValue.(model.PolicyEdgeNode)
		d.Set("member_index", policyEdgeNode.MemberIndex)
		return nil
	}

	// Local manager
	connector := getPolicyConnector(m)
	client := edge_clusters.NewEdgeNodesClient(connector)
	var obj model.PolicyEdgeNode
	edgeClusterID := getPolicyIDFromPath(edgeClusterPath)
	objID := d.Get("id").(string)

	if objID != "" {
		// Get by id
		objGet, err := client.Get(defaultSite, getPolicyEnforcementPoint(m), edgeClusterID, objID)

		if err != nil {
			return handleDataSourceReadError(d, "Edge Node", objID, err)
		}
		obj = objGet
	} else {
		// Get by full name/prefix
		name, nameSet := d.GetOk("display_name")
		objName := name.(string)
		objMemberIndex := int64(memberIndex.(int))
		includeMarkForDeleteObjectsParam := false
		objList, err := client.List(defaultSite, getPolicyEnforcementPoint(m), edgeClusterID, nil, &includeMarkForDeleteObjectsParam, nil, nil, nil, nil)
		if err != nil {
			return handleListError("Edge Node", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.PolicyEdgeNode
		var prefixMatch []model.PolicyEdgeNode
		for _, objInList := range objList.Results {
			indexMatch := true
			if memberIndexSet && objMemberIndex != *objInList.MemberIndex {
				indexMatch = false
			}
			if nameSet && strings.HasPrefix(*objInList.DisplayName, objName) && indexMatch {
				prefixMatch = append(prefixMatch, objInList)
			}
			if nameSet && *objInList.DisplayName == objName && indexMatch {
				perfectMatch = append(perfectMatch, objInList)
			}
			if !nameSet && indexMatch {
				perfectMatch = append(perfectMatch, objInList)
			}
		}

		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("Found multiple edge nodes with name '%s' and index %d", objName, memberIndex)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple edge nodes with name starting with '%s' and index %d", objName, memberIndex)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("edge node '%s' was not found and %d", objName, memberIndex)
		}
	}

	d.SetId(*obj.Id)
	d.Set("member_index", obj.MemberIndex)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	return nil
}
