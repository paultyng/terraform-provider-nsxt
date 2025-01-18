// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/trust"
)

func dataSourceNsxtCertificate() *schema.Resource {
	return &schema.Resource{
		Read:               dataSourceNsxtCertificateRead,
		DeprecationMessage: mpObjectDataSourceDeprecationMessage,
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
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return dataSourceNotSupportedError()
	}

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
		found := false
		lister := func(info *paginationInfo) error {
			objList, _, err := nsxClient.NsxComponentAdministrationApi.GetCertificates(nsxClient.Context, info.LocalVarOptionals)
			if err != nil {
				return fmt.Errorf("Error while reading certificates: %v", err)
			}

			info.PageCount = int64(len(objList.Results))
			info.TotalCount = objList.ResultCount
			info.Cursor = objList.Cursor

			// go over the list to find the correct one
			for _, objInList := range objList.Results {
				if objInList.DisplayName == objName {
					if found {
						return fmt.Errorf("Found multiple certificates with name '%s'", objName)
					}
					obj = objInList
					found = true
				}
			}
			return nil
		}
		total, err := handlePagination(lister)
		if err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("Certificate with name '%s' was not found among %d certs", objName, total)
		}
	} else {
		return fmt.Errorf("Error obtaining certificate ID or name during read")
	}

	d.SetId(obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)

	return nil
}
