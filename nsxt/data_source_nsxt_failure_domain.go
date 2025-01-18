// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func dataSourceNsxtFailureDomain() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtFailureDomainRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
		},
	}
}

func dataSourceNsxtFailureDomainRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewFailureDomainsClient(connector)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj model.FailureDomain
	if objID != "" {
		// Get by id
		objGet, err := client.Get(objID)
		if isNotFoundError(err) {
			return fmt.Errorf("FailureDomain with ID %s was not found", objID)
		}

		if err != nil {
			return fmt.Errorf("error while reading FailureDomain %s: %v", objID, err)
		}
		obj = objGet
	} else if objName == "" {
		return fmt.Errorf("error obtaining FailureDomain ID or name during read")
	} else {
		// Get by full name/prefix
		objList, err := client.List()
		if err != nil {
			return fmt.Errorf("error while reading FailureDomains: %v", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.FailureDomain
		var prefixMatch []model.FailureDomain
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
				return fmt.Errorf("found multiple FailureDomains with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("found multiple FailureDomains with name starting with '%s'", objName)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("FailureDomain with name '%s' was not found", objName)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
