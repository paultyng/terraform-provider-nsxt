// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
)

func dataSourceNsxtManagementCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtManagementClusterRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Unique identifier of this cluster.",
				Computed:    true,
			},
			"node_sha256_thumbprint": {
				Type:        schema.TypeString,
				Description: "SHA256 of certificate thumbprint of this manager node.",
				Computed:    true,
			},
		},
	}
}

func dataSourceNsxtManagementClusterRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewClusterClient(connector)

	hostIP, err := net.LookupIP(m.(nsxtClients).Host[len("https://"):])
	if err != nil {
		return fmt.Errorf("error while resolving client hostname %s: %v", m.(nsxtClients).Host[len("https://"):], err)
	}

	clusterObj, err := client.Get()
	if err != nil {
		return fmt.Errorf("error while reading cluster configuration: %v", err)
	}
	for _, nodeConfig := range clusterObj.Nodes {
		if nodeConfig.ApiListenAddr != nil && *nodeConfig.ApiListenAddr.IpAddress == hostIP[0].String() {
			if *nodeConfig.ApiListenAddr.CertificateSha256Thumbprint == "" {
				return fmt.Errorf("manager node thumbprint not found while reading cluster node configuration for node %s", *nodeConfig.ApiListenAddr.IpAddress)
			}
			d.Set("node_sha256_thumbprint", nodeConfig.ApiListenAddr.CertificateSha256Thumbprint)
		}
	}

	if d.Get("node_sha256_thumbprint").(string) == "" {
		return fmt.Errorf("cluster node sha256 thumbprint not found for node %s", m.(nsxtClients).Host[len("https://"):])
	}
	d.SetId(*clusterObj.ClusterId)

	return nil
}
