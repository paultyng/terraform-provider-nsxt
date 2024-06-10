/* Copyright © 2021 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyGatewayLocaleService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGatewayLocaleServiceRead,

		Schema: map[string]*schema.Schema{
			"gateway_path": getPolicyPathSchema(true, true, "Gateway path"),
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"bgp_path":     getComputedPolicyPathSchema("Path for BGP config"),
			"edge_cluster_path": {
				Type:        schema.TypeString,
				Description: "The path of the edge cluster connected to this Tier0 gateway",
				Optional:    true,
				Computed:    true,
			},
			"context": getContextSchema(false, false, false),
		},
	}
}

func dataSourceNsxtPolicyGatewayLocaleServiceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPath := d.Get("gateway_path").(string)
	query := make(map[string]string)
	query["parent_path"] = gwPath
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	obj, err := policyDataSourceResourceReadWithValidation(d, connector, context, "LocaleServices", query, false)

	if err != nil {
		return err
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(obj, model.LocaleServicesBindingType())
	if len(errors) > 0 {
		return errors[0]
	}
	localeService := dataValue.(model.LocaleServices)

	if localeService.EdgeClusterPath != nil {
		d.Set("edge_cluster_path", localeService.EdgeClusterPath)
	}
	d.SetId(*localeService.Id)
	d.Set("display_name", localeService.DisplayName)
	d.Set("description", localeService.Description)
	d.Set("path", localeService.Path)
	if localeService.Path != nil {
		bgpPath := fmt.Sprintf("%s/bgp", *localeService.Path)
		d.Set("bgp_path", bgpPath)
	}

	return nil
}
