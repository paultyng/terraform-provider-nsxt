// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var lbMonitorTypeValues = []string{"HTTP", "HTTPS", "TCP", "UDP", "ICMP", "PASSIVE", "ANY"}
var lbMonitorTypeMap = map[string]string{
	model.LBMonitorProfile_RESOURCE_TYPE_LBHTTPSMONITORPROFILE:   "HTTPS",
	model.LBMonitorProfile_RESOURCE_TYPE_LBHTTPMONITORPROFILE:    "HTTP",
	model.LBMonitorProfile_RESOURCE_TYPE_LBTCPMONITORPROFILE:     "TCP",
	model.LBMonitorProfile_RESOURCE_TYPE_LBUDPMONITORPROFILE:     "UDP",
	model.LBMonitorProfile_RESOURCE_TYPE_LBICMPMONITORPROFILE:    "ICMP",
	model.LBMonitorProfile_RESOURCE_TYPE_LBPASSIVEMONITORPROFILE: "PASSIVE",
}

func dataSourceNsxtPolicyLBMonitor() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyLBMonitorRead,

		Schema: map[string]*schema.Schema{
			"id": getDataSourceIDSchema(),
			"type": {
				Type:         schema.TypeString,
				Description:  "Load Balancer Monitor Type",
				Optional:     true,
				Default:      "ANY",
				ValidateFunc: validation.StringInSlice(lbMonitorTypeValues, false),
			},
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func policyLbMonitorConvert(obj *data.StructValue, requestedType string) (*model.LBMonitorProfile, error) {
	converter := bindings.NewTypeConverter()

	data, errs := converter.ConvertToGolang(obj, model.LBMonitorProfileBindingType())
	if errs != nil {
		return nil, errs[0]
	}

	profile := data.(model.LBMonitorProfile)
	profileType, ok := lbMonitorTypeMap[profile.ResourceType]
	if !ok {
		return nil, fmt.Errorf("Unknown LB Monitor type %s", profile.ResourceType)
	}
	if (requestedType != "ANY") && (requestedType != profileType) {
		return nil, nil
	}
	return &profile, nil
}

func dataSourceNsxtPolicyLBMonitorRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbMonitorProfilesClient(connector)

	objID := d.Get("id").(string)
	objTypeValue, typeSet := d.GetOk("type")
	objType := objTypeValue.(string)
	objName := d.Get("display_name").(string)
	var result *model.LBMonitorProfile
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)

		if err != nil {
			return handleDataSourceReadError(d, "LBMonitor", objID, err)
		}
		result, err = policyLbMonitorConvert(objGet, objType)
		if err != nil {
			return fmt.Errorf("Error while converting LBMonitor %s: %v", objID, err)
		}
		if result == nil {
			return fmt.Errorf("LBMonitor with ID '%s' and type %s was not found", objID, objType)
		}
	} else if objName == "" && !typeSet {
		return fmt.Errorf("Error obtaining LBMonitor ID or name or type during read")
	} else {
		// Get by full name/prefix
		includeMarkForDeleteObjectsParam := false
		objList, err := client.List(nil, &includeMarkForDeleteObjectsParam, nil, nil, nil, nil)
		if err != nil {
			return handleListError("LBMonitor", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.LBMonitorProfile
		var prefixMatch []model.LBMonitorProfile
		for _, objInList := range objList.Results {
			obj, err := policyLbMonitorConvert(objInList, objType)
			if err != nil {
				return fmt.Errorf("Error while converting LBMonitor %s: %v", objID, err)
			}
			if obj == nil {
				continue
			}
			if objName != "" && obj.DisplayName != nil && strings.HasPrefix(*obj.DisplayName, objName) {
				prefixMatch = append(prefixMatch, *obj)
			}
			if obj.DisplayName != nil && *obj.DisplayName == objName {
				perfectMatch = append(perfectMatch, *obj)
			}
			if objName == "" && typeSet {
				// match only by type
				perfectMatch = append(perfectMatch, *obj)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("Found multiple LBMonitors with name '%s'", objName)
			}
			result = &perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple LBMonitors with name starting with '%s'", objName)
			}
			result = &prefixMatch[0]
		} else {
			return fmt.Errorf("LBMonitor with name '%s' and type %s was not found", objName, objType)
		}
	}

	d.SetId(*result.Id)
	d.Set("display_name", result.DisplayName)
	d.Set("type", lbMonitorTypeMap[result.ResourceType])
	d.Set("description", result.Description)
	d.Set("path", result.Path)
	return nil
}
