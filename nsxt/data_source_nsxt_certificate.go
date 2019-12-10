/* Copyright © 2017 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/trust"
	"net/http"
)

func dataSourceNsxtCertificate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtCertificateRead,

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

func dataSourceNsxtCertificateRead(d *schema.ResourceData, m interface{}) error {
	// Read cerificate by name or id
	nsxClient := m.(*api.APIClient)
	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)
	var obj trust.Certificate
	if objID != "" {
		// Get by id
		objGet, resp, err := nsxClient.NsxComponentAdministrationApi.GetCertificate(nsxClient.Context, objID, nil)

		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("certificate %s was not found", objID)
		}
		if err != nil {
			return fmt.Errorf("Error while reading certificate %s: %v", objID, err)
		}
		obj = objGet

	} else if objName != "" {
		// Get by name
		// TODO use 2nd parameter localVarOptionals for paging
		objList, _, err := nsxClient.NsxComponentAdministrationApi.GetCertificates(nsxClient.Context, nil)
		if err != nil {
			return fmt.Errorf("Error while reading certificates: %v", err)
		}
		// go over the list to find the correct one
		found := false
		for _, objInList := range objList.Results {
			if objInList.DisplayName == objName {
				if found {
					return fmt.Errorf("Found multiple certificates with name '%s'", objName)
				}
				obj = objInList
				found = true
			}
		}
		if !found {
			return fmt.Errorf("Certificate with name '%s' was not found", objName)
		}
	} else {
		return fmt.Errorf("Error obtaining certificate ID or name during read")
	}

	d.SetId(obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
