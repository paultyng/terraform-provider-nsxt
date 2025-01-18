// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtDhcpServerProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtDhcpServerProfileCreate,
		Read:   resourceNsxtDhcpServerProfileRead,
		Update: resourceNsxtDhcpServerProfileUpdate,
		Delete: resourceNsxtDhcpServerProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Description: "Edge cluster uuid",
				Required:    true,
			},
			"edge_cluster_member_indexes": {
				Type:        schema.TypeList,
				Description: "Edge nodes from the given cluster",
				Optional:    true,
				MaxItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func resourceNsxtDhcpServerProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	edgeClusterID := d.Get("edge_cluster_id").(string)
	edgeClusterMemberIndexes := intList2int64List(d.Get("edge_cluster_member_indexes").([]interface{}))
	dhcpProfile := manager.DhcpProfile{
		DisplayName:              displayName,
		Description:              description,
		EdgeClusterId:            edgeClusterID,
		EdgeClusterMemberIndexes: edgeClusterMemberIndexes,
		Tags:                     tags,
	}

	dhcpProfile, resp, err := nsxClient.ServicesApi.CreateDhcpProfile(nsxClient.Context, dhcpProfile)

	if err != nil {
		return fmt.Errorf("Error during DhcpProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during DhcpProfile create: %v", resp.StatusCode)
	}
	d.SetId(dhcpProfile.Id)

	return resourceNsxtDhcpServerProfileRead(d, m)
}

func resourceNsxtDhcpServerProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	dhcpProfile, resp, err := nsxClient.ServicesApi.ReadDhcpProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during DhcpProfile read: %v", err)
	}

	d.Set("revision", dhcpProfile.Revision)
	d.Set("description", dhcpProfile.Description)
	d.Set("display_name", dhcpProfile.DisplayName)
	setTagsInSchema(d, dhcpProfile.Tags)
	d.Set("edge_cluster_id", dhcpProfile.EdgeClusterId)
	d.Set("edge_cluster_member_indexes", dhcpProfile.EdgeClusterMemberIndexes)

	return nil
}

func resourceNsxtDhcpServerProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	edgeClusterID := d.Get("edge_cluster_id").(string)
	edgeClusterMemberIndexes := intList2int64List(d.Get("edge_cluster_member_indexes").([]interface{}))
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	dhcpProfile := manager.DhcpProfile{
		DisplayName:              displayName,
		Description:              description,
		EdgeClusterId:            edgeClusterID,
		EdgeClusterMemberIndexes: edgeClusterMemberIndexes,
		Tags:                     tags,
		Revision:                 revision,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateDhcpProfile(nsxClient.Context, id, dhcpProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during DhcpProfile update: %v", err)
	}

	return resourceNsxtDhcpServerProfileRead(d, m)
}

func resourceNsxtDhcpServerProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteDhcpProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during DhcpProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] DhcpProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
