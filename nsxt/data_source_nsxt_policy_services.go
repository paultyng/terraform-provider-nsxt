// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
)

func dataSourceNsxtPolicyServices() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyServicesRead,

		Schema: map[string]*schema.Schema{
			"context": getContextSchema(false, false, false),
			"items": {
				Type:        schema.TypeMap,
				Description: "Mapping of services policy path by display name",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceNsxtPolicyServicesRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewServicesClient(getSessionContext(d, m), connector)

	servicesMap := make(map[string]string)
	results, err := client.List(nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	for _, r := range results.Results {
		servicesMap[*r.DisplayName] = *r.Path
	}

	d.SetId(newUUID())
	return nil
}
