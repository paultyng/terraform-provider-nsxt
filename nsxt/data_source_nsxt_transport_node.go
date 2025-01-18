// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func dataSourceNsxtTransportNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtTransportNodeRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDescriptionSchema(),
		},
	}
}

func dataSourceNsxtTransportNodeRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)

	var obj model.TransportNode

	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)
		if err != nil {
			return fmt.Errorf("failed to read TransportNode %s: %v", objID, err)
		}

		// Make sure that found obj is an EdgeNode
		converter := bindings.NewTypeConverter()
		base, errs := converter.ConvertToGolang(obj.NodeDeploymentInfo, model.NodeBindingType())
		if errs != nil {
			return fmt.Errorf("failed to convert NodeDeploymentInfo for node %s %v", objID, errs[0])
		}
		node := base.(model.Node)
		if node.ResourceType != model.EdgeNode__TYPE_IDENTIFIER {
			return fmt.Errorf("no Transport Node matches the criteria")
		}
		obj = objGet
	} else {
		// Get by full name/prefix - filter out anything except EdgeNodes
		objType := model.EdgeNode__TYPE_IDENTIFIER
		objList, err := client.List(nil, nil, nil, nil, nil, &objType, nil, nil, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to read Transport Nodes: %v", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.TransportNode
		var prefixMatch []model.TransportNode
		for _, objInList := range objList.Results {
			if len(objName) == 0 {
				// We want to grab single Transport Node
				perfectMatch = append(perfectMatch, objInList)
				continue
			}
			if strings.HasPrefix(*objInList.DisplayName, objName) {
				prefixMatch = append(prefixMatch, objInList)
			}
			if *objInList.DisplayName == objName {
				perfectMatch = append(perfectMatch, objInList)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("found multiple Transport Nodes matching the criteria")
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("found multiple Transport Nodes matching the criteria")
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("no Transport Node matches the criteria")
		}
	}
	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
