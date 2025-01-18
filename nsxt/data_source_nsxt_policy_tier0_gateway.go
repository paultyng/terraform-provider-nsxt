// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func dataSourceNsxtPolicyTier0Gateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyTier0GatewayRead,

		Schema: map[string]*schema.Schema{
			"id":           getDataSourceIDSchema(),
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
			"edge_cluster_path": {
				Type:        schema.TypeString,
				Description: "The path of the edge cluster connected to this Tier0 gateway",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtPolicyTier0GatewayRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	obj, err := policyDataSourceResourceRead(d, connector, getSessionContext(d, m), "Tier0", nil)
	if err != nil {
		return err
	}

	// Single edge cluster is not informative for global manager
	if isPolicyGlobalManager(m) {
		d.Set("edge_cluster_path", "")
	} else {
		converter := bindings.NewTypeConverter()
		dataValue, errors := converter.ConvertToGolang(obj, model.Tier0BindingType())
		if len(errors) > 0 {
			return errors[0]
		}
		gw := dataValue.(model.Tier0)
		err := resourceNsxtPolicyTier0GatewayReadEdgeCluster(getSessionContext(d, m), d, connector)
		if err != nil {
			return fmt.Errorf("failed to get Tier0 %s locale-services: %v", *gw.Id, err)
		}
	}
	return nil
}

func getPolicyTier0GatewayLocaleServiceEntry(context utl.SessionContext, gwID string, connector client.Connector) (*model.LocaleServices, error) {
	// Get the locale services of this Tier1 for the edge-cluster id
	client := tier_0s.NewLocaleServicesClient(connector)
	obj, err := client.Get(gwID, defaultPolicyLocaleServiceID)
	if err == nil {
		return &obj, nil
	}

	// No locale-service with the default ID
	// List all the locale services
	objList, errList := listPolicyTier0GatewayLocaleServices(context, connector, gwID)
	if errList != nil {
		return nil, fmt.Errorf("Error while reading Tier1 %v locale-services: %v", gwID, err)
	}
	for _, objInList := range objList {
		// Find the one with the edge cluster path
		if objInList.EdgeClusterPath != nil {
			return &objInList, nil
		}
	}
	// No locale service with edge cluster path found.
	// Return any of the locale services (To avoid creating a new one)
	for _, objInList := range objList {
		return &objInList, nil
	}

	// No locale service with edge cluster path found
	return nil, nil
}

func resourceNsxtPolicyTier0GatewayReadEdgeCluster(context utl.SessionContext, d *schema.ResourceData, connector client.Connector) error {
	// Get the locale services of this Tier1 for the edge-cluster id
	obj, err := getPolicyTier0GatewayLocaleServiceEntry(context, d.Id(), connector)
	if err != nil || obj == nil {
		// No locale-service found
		return nil
	}
	if obj.EdgeClusterPath != nil {
		d.Set("edge_cluster_path", obj.EdgeClusterPath)
		return nil
	}
	// No edge cluster found
	return nil
}
