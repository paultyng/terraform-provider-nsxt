// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func dataSourceNsxtPolicySecurityPolicy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicySecurityPolicyRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"context":      getContextSchema(false, false, false),
			"domain":       getDataSourceDomainNameSchema(),
			"is_default": {
				Type:        schema.TypeBool,
				Description: "Is this a default policy",
				Optional:    true,
				Default:     false,
			},
			"category": {
				Type:         schema.TypeString,
				Description:  "Category",
				ValidateFunc: validation.StringInSlice(securityPolicyCategoryValues, false),
				Optional:     true,
				Computed:     true,
			},
		},
	}
}

// Local Manager Only
func listSecurityPolicies(context utl.SessionContext, domain string, connector client.Connector) ([]model.SecurityPolicy, error) {
	client := domains.NewSecurityPoliciesClient(context, connector)
	if client == nil {
		return nil, policyResourceNotSupportedError()
	}

	var results []model.SecurityPolicy
	boolFalse := false
	var cursor *string
	total := 0

	for {
		policies, err := client.List(domain, cursor, nil, nil, nil, nil, &boolFalse, nil)
		if err != nil {
			return results, err
		}
		results = append(results, policies.Results...)
		if total == 0 && policies.ResultCount != nil {
			// first response
			total = int(*policies.ResultCount)
		}

		cursor = policies.Cursor
		if len(results) >= total {
			return results, nil
		}
	}
}

func dataSourceNsxtPolicySecurityPolicyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	category := d.Get("category").(string)
	domain := d.Get("domain").(string)
	isDefault := d.Get("is_default").(bool)

	objID := d.Get("id").(string)
	objName := d.Get("display_name").(string)

	var obj model.SecurityPolicy
	context := getSessionContext(d, m)
	if objID != "" {
		// Get by id
		client := domains.NewSecurityPoliciesClient(context, connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		objGet, err := client.Get(domain, objID)
		if isNotFoundError(err) {
			return fmt.Errorf("Security Policy with ID %s was not found", objID)
		}

		if err != nil {
			return fmt.Errorf("Error while reading Security Policy %s: %v", objID, err)
		}
		obj = objGet
	} else if objName == "" && category == "" {
		return fmt.Errorf("Security Policy id, display name or category must be specified")
	} else {
		objList, err := listSecurityPolicies(context, domain, connector)
		if err != nil {
			return fmt.Errorf("Error while reading Security Policies: %v", err)
		}
		// go over the list to find the correct one (prefer a perfect match. If not - prefix match)
		var perfectMatch []model.SecurityPolicy
		var prefixMatch []model.SecurityPolicy
		for _, objInList := range objList {
			if category != "" && objInList.Category != nil && category != *objInList.Category {
				continue
			}
			if objInList.IsDefault != nil && *objInList.IsDefault != isDefault {
				continue
			}
			if objName != "" && objInList.DisplayName != nil {
				if strings.HasPrefix(*objInList.DisplayName, objName) {
					prefixMatch = append(prefixMatch, objInList)
				}
				if *objInList.DisplayName == objName {
					perfectMatch = append(perfectMatch, objInList)
				}
			} else {
				prefixMatch = append(prefixMatch, objInList)
			}
		}
		if len(perfectMatch) > 0 {
			if len(perfectMatch) > 1 {
				return fmt.Errorf("Found multiple Security Policies with name '%s'", objName)
			}
			obj = perfectMatch[0]
		} else if len(prefixMatch) > 0 {
			if len(prefixMatch) > 1 {
				return fmt.Errorf("Found multiple Security Policies with name starting with '%s' and category '%s'", objName, category)
			}
			obj = prefixMatch[0]
		} else {
			return fmt.Errorf("Security Policy with name '%s' and category '%s' was not found", objName, category)
		}
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)
	d.Set("category", obj.Category)
	d.Set("is_default", obj.IsDefault)

	return nil
}
