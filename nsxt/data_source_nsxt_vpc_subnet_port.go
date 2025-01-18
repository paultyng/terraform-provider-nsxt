// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/vpcs/subnets"
)

func dataSourceNsxtVpcSubnetPort() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtVpcSubnetPortRead,

		Schema: map[string]*schema.Schema{
			"id": getDataSourceIDSchema(),
			"subnet_path": {
				Type:         schema.TypeString,
				Description:  "Path of parent subnet",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"vm_id": {
				Type:        schema.TypeString,
				Description: "external ID of VM",
				Required:    true,
			},
			"display_name": getDataSourceExtendedDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"path":         getPathSchema(),
		},
	}
}

func listVpcSubnetPorts(connector client.Connector, subnetPath string) ([]model.VpcSubnetPort, error) {

	var results []model.VpcSubnetPort
	parents, pathErr := parseStandardPolicyPathVerifySize(subnetPath, 4)
	if pathErr != nil {
		return results, pathErr
	}
	boolFalse := false
	var cursor *string
	total := 0
	var err error
	var ports model.VpcSubnetPortListResult

	for {
		portClient := subnets.NewPortsClient(connector)
		if portClient == nil {
			return results, policyResourceNotSupportedError()
		}
		ports, err = portClient.List(parents[0], parents[1], parents[2], parents[3], cursor, &boolFalse, nil, nil, &boolFalse, nil)

		if err != nil {
			return results, err
		}
		results = append(results, ports.Results...)
		if total == 0 && ports.ResultCount != nil {
			// first response
			total = int(*ports.ResultCount)
		}
		cursor = ports.Cursor
		if len(results) >= total {
			log.Printf("[DEBUG] Found %d ports for subnet %s", len(results), subnetPath)
			return results, nil
		}
	}
}

func dataSourceNsxtVpcSubnetPortRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	externalID := d.Get("vm_id").(string)
	subnetPath := d.Get("subnet_path").(string)
	vifAttachmentIds, err := listPolicyVifAttachmentsForVM(m, externalID)
	if err != nil {
		return fmt.Errorf("failed to list port attachments for VM id %s", externalID)
	}

	ports, portsErr := listVpcSubnetPorts(connector, subnetPath)
	if portsErr != nil {
		return portsErr
	}
	for _, port := range ports {
		if port.Attachment == nil || port.Attachment.Id == nil {
			continue
		}

		for _, attachment := range vifAttachmentIds {
			if attachment == *port.Attachment.Id {
				log.Printf("[DEBUG] Matching port %s found", *port.Path)
				d.SetId(*port.Id)
				d.Set("path", port.Path)
				d.Set("display_name", port.DisplayName)
				d.Set("description", port.Description)
				return nil
			}
		}
	}

	return fmt.Errorf("failed to find port for vm %s on subnet %s", externalID, subnetPath)
}
