// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/transit_gateways"
)

var transitGatewayNatTypes = []string{
	model.PolicyNat_NAT_TYPE_USER,
	model.PolicyNat_NAT_TYPE_DEFAULT,
}

func dataSourceNsxtPolicyTransitGatewayNat() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyTransitGatewayNatRead,

		Schema: map[string]*schema.Schema{
			"id":                   getDataSourceIDSchema(),
			"transit_gateway_path": getPolicyPathSchema(true, false, "Transit Gateway Path"),
			"nat_type": {
				Type:         schema.TypeString,
				Description:  "Nat Type",
				Optional:     true,
				Default:      model.PolicyNat_NAT_TYPE_USER,
				ValidateFunc: validation.StringInSlice(transitGatewayNatTypes, false),
			},
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func dataSourceNsxtPolicyTransitGatewayNatRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	parentPath := d.Get("transit_gateway_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return fmt.Errorf("invalid transit_gateway_path: %v", pathErr)
	}

	natType := d.Get("nat_type").(string)

	client := clientLayer.NewNatClient(connector)

	// Nat type is the ID
	obj, err := client.Get(parents[0], parents[1], parents[2], natType)
	if err != nil {
		return fmt.Errorf("NAT with type %s was not found for Transit Gateway %s", natType, parentPath)
	}

	d.SetId(*obj.Id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("path", obj.Path)

	return nil
}
