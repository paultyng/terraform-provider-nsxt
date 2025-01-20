// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/terraform-provider-nsxt/api/infra/domains"
)

func dataSourceNsxtPolicyGroups() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyGroupsRead,

		Schema: map[string]*schema.Schema{
			"context": getContextSchema(false, false, false),
			"domain":  getDomainNameSchema(),
			"items": {
				Type:        schema.TypeMap,
				Description: "Mapping of service policy path by display name",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceNsxtPolicyGroupsRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	domainName := d.Get("domain").(string)

	client := domains.NewGroupsClient(getSessionContext(d, m), connector)

	groupsMap := make(map[string]string)
	results, err := client.List(domainName, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	for _, r := range results.Results {
		groupsMap[*r.DisplayName] = *r.Path
	}

	d.SetId(newUUID())
	return nil
}
