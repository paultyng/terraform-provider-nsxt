// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyTier1Gateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyTier1GatewayRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"edge_cluster_path": {
				Type:        schema.TypeString,
				Description: "The path of the edge cluster connected to this Tier1 gateway",
				Optional:    true,
				Computed:    true,
			},
			"context": getContextSchema(false, false, false),
		},
	}
}

func dataSourceNsxtPolicyTier1GatewayRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	obj, err := policyDataSourceResourceRead(d, connector, getSessionContext(d, m), "Tier1", nil)
	if err != nil {
		return err
	}

	// Single edge cluster is not informative for global manager
	if isPolicyGlobalManager(m) {
		d.Set("edge_cluster_path", "")
	} else {
		converter := bindings.NewTypeConverter()
		dataValue, errors := converter.ConvertToGolang(obj, model.Tier1BindingType())
		if len(errors) > 0 {
			return errors[0]
		}
		tier1 := dataValue.(model.Tier1)
		err := resourceNsxtPolicyTier1GatewayReadEdgeCluster(getSessionContext(d, m), d, connector)
		if err != nil {
			return fmt.Errorf("failed to get Tier1 %s locale-services: %v", *tier1.Id, err)
		}
	}
	return nil
}
