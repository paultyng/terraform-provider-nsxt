// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/cluster"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

var ClusterVirtualIPForceType = []string{
	nsxModel.ClusterVirtualIpProperties_FORCE_TRUE,
	nsxModel.ClusterVirtualIpProperties_FORCE_FALSE,
}

var DefaultIPv4VirtualAddress = "0.0.0.0"

var DefaultIPv6VirtualAddress = "::"

func resourceNsxtClusterVirualIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtClusterVirualIPCreate,
		Read:   resourceNsxtClusterVirualIPRead,
		Update: resourceNsxtClusterVirualIPUpdate,
		Delete: resourceNsxtClusterVirualIPDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"force": {
				Type:        schema.TypeBool,
				Description: "On enable it ignores duplicate address detection and DNS lookup validation check",
				Optional:    true,
				Default:     true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Description:  "Virtual IPv4 address",
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
				Default:      DefaultIPv4VirtualAddress,
			},
			"ipv6_address": {
				Type:         schema.TypeString,
				Description:  "Virtual IPv6 address",
				Optional:     true,
				ValidateFunc: validation.IsIPv6Address,
				Default:      DefaultIPv6VirtualAddress,
			},
		},
	}
}

func resourceNsxtClusterVirualIPCreate(d *schema.ResourceData, m interface{}) error {
	// Create and update workflow are mostly the same for virtual IP resource
	// except that create workflow sets the ID of this resource
	id := d.Id()
	if id == "" {
		id = newUUID()
	}
	err := setClusterVirtualIP(d, m)
	if err != nil {
		return handleCreateError("ClusterVirtualIP", id, err)
	}
	d.SetId(id)
	return resourceNsxtClusterVirualIPRead(d, m)
}

func resourceNsxtClusterVirualIPRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ClusterVirtualIP ID")
	}
	connector := getPolicyConnector(m)
	client := cluster.NewApiVirtualIpClient(connector)

	obj, err := client.Get()
	if err != nil {
		return handleReadError(d, "ClusterVirtualIP", id, err)
	}

	d.Set("ip_address", obj.IpAddress)
	d.Set("ipv6_address", obj.Ip6Address)

	return nil
}

func setClusterVirtualIP(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := cluster.NewApiVirtualIpClient(connector)
	force := d.Get("force").(bool)
	ipAddress := d.Get("ip_address").(string)
	ipv6Address := d.Get("ipv6_address").(string)
	var forceStr string
	if force {
		forceStr = nsxModel.ClusterVirtualIpProperties_FORCE_TRUE
	} else {
		forceStr = nsxModel.ClusterVirtualIpProperties_FORCE_FALSE
	}
	var err error
	if util.NsxVersionHigherOrEqual("4.0.0") {
		_, err = client.Setvirtualip(&forceStr, &ipv6Address, &ipAddress)
	} else {
		// IPv6 not supported
		_, err = client.Setvirtualip(nil, nil, &ipAddress)
	}
	if err != nil {
		log.Printf("[WARNING] Failed to set virtual ip: %v", err)
		return err
	}
	return nil
}

func resourceNsxtClusterVirualIPUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	err := setClusterVirtualIP(d, m)
	if err != nil {
		return handleUpdateError("ClusterVirtualIP", id, err)
	}
	return resourceNsxtClusterVirualIPRead(d, m)
}

func resourceNsxtClusterVirualIPDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ClusterVirtualIP ID")
	}
	connector := getPolicyConnector(m)
	client := cluster.NewApiVirtualIpClient(connector)
	_, err := client.Clearvirtualip()
	if err != nil {
		log.Printf("[WARNING] Failed to clear virtual ip: %v", err)
		return handleDeleteError("ClusterVirtualIP", id, err)
	}
	if util.NsxVersionHigherOrEqual("4.0.0") {
		_, err = client.Clearvirtualip6()
		if err != nil {
			log.Printf("[WARNING] Failed to clear virtual ipv6 ip: %v", err)
			return handleDeleteError("ClusterVirtualIP", id, err)
		}
	}
	return nil
}
