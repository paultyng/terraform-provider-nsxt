/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyIPSecVpnLocalEndpoint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyIPSecVpnLocalEndpointRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"service_path": getPolicyPathSchema(false, false, "Policy path for IPSec VPN service"),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"local_address": {
				Type:        schema.TypeString,
				Description: "Local IPv4 IP address",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyIPSecVpnLocalEndpointRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	servicePath := d.Get("service_path").(string)
	query := make(map[string]string)
	if len(servicePath) > 0 {
		s := strings.Split(servicePath, "/")
		if len(s) != 8 && len(s) != 6 {
			// The policy path of IPSec VPN Service should be like /infra/tier-0s/aaa/locale-services/bbb/ipsec-vpn-services/ccc
			// or /infra/tier-0s/aaa/ipsec-vpn-services/bbb
			return fmt.Errorf("Invalid IPSec Vpn Service path: %s", servicePath)
		}
		if len(s) == 8 {
			// search API does not recognized the locale-services part in the VPN service path
			servicePath = strings.Join(append(s[:4], s[6:]...), "/")
		}
		query["parent_path"] = fmt.Sprintf("%s*", servicePath)
	}
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	objInt, err := policyDataSourceResourceReadWithValidation(d, connector, context, "IPSecVpnLocalEndpoint", query, false)
	if err != nil {
		return err
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToGolang(objInt, model.IPSecVpnLocalEndpointBindingType())
	if len(errors) > 0 {
		return fmt.Errorf("Failed to convert type for Local Endpoint: %v", errors[0])
	}
	obj := dataValue.(model.IPSecVpnLocalEndpoint)
	d.Set("local_address", obj.LocalAddress)
	return nil
}
