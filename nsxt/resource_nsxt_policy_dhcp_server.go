/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyDhcpServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDhcpServerCreate,
		Read:   resourceNsxtPolicyDhcpServerRead,
		Update: resourceNsxtPolicyDhcpServerUpdate,
		Delete: resourceNsxtPolicyDhcpServerDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":            getNsxIDSchema(),
			"path":              getPathSchema(),
			"display_name":      getDisplayNameSchema(),
			"description":       getDescriptionSchema(),
			"revision":          getRevisionSchema(),
			"tag":               getTagsSchema(),
			"context":           getContextSchema(false, false, false),
			"edge_cluster_path": getPolicyPathSchema(false, false, "Edge Cluster path"),
			"lease_time": {
				Type:         schema.TypeInt,
				Description:  "IP Address lease time in seconds",
				Optional:     true,
				Default:      86400,
				ValidateFunc: validation.IntAtLeast(60),
			},
			"preferred_edge_paths": {
				Type:        schema.TypeList,
				Description: "The first edge node is assigned as active edge, and second one as standby edge",
				Optional:    true,
				MaxItems:    2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePolicyPath(),
				},
			},
			"server_addresses": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "DHCP server address in CIDR format",
				MaxItems:    2,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIPCidr(),
				},
			},
		},
	}
}

func resourceNsxtPolicyDhcpServerExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {

	client := infra.NewDhcpServerConfigsClient(sessionContext, connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyDhcpServerSchemaToModel(d *schema.ResourceData) model.DhcpServerConfig {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	edgeClusterPath := d.Get("edge_cluster_path").(string)
	leaseTime := int64(d.Get("lease_time").(int))
	preferredEdgePaths := interface2StringList(d.Get("preferred_edge_paths").([]interface{}))
	serverAddresses := interface2StringList(d.Get("server_addresses").([]interface{}))

	obj := model.DhcpServerConfig{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		LeaseTime:   &leaseTime,
	}

	if edgeClusterPath != "" {
		obj.EdgeClusterPath = &edgeClusterPath
	}
	if len(preferredEdgePaths) > 0 {
		obj.PreferredEdgePaths = preferredEdgePaths
	}
	if len(serverAddresses) > 0 {
		obj.ServerAddresses = serverAddresses
		// TODO: remove setting of server_address once this deprecated property is removed from NSX API
		obj.ServerAddress = &serverAddresses[0]
	}

	return obj
}

func resourceNsxtPolicyDhcpServerCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyDhcpServerExists)
	if err != nil {
		return err
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating DhcpServer with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewDhcpServerConfigsClient(context, connector)
	err = client.Patch(id, resourceNsxtPolicyDhcpServerSchemaToModel(d))
	if err != nil {
		return handleCreateError("DhcpServer", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDhcpServerRead(d, m)
}

func resourceNsxtPolicyDhcpServerRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpServer ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewDhcpServerConfigsClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "DhcpServer", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("edge_cluster_path", obj.EdgeClusterPath)
	d.Set("lease_time", obj.LeaseTime)
	d.Set("preferred_edge_paths", obj.PreferredEdgePaths)
	d.Set("server_addresses", obj.ServerAddresses)

	return nil
}

func resourceNsxtPolicyDhcpServerUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewDhcpServerConfigsClient(context, connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpServer ID")
	}

	// Update the resource using PATCH
	err = client.Patch(id, resourceNsxtPolicyDhcpServerSchemaToModel(d))
	if err != nil {
		return handleUpdateError("DhcpServer", id, err)
	}

	return resourceNsxtPolicyDhcpServerRead(d, m)
}

func resourceNsxtPolicyDhcpServerDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpServer ID")
	}

	var err error
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewDhcpServerConfigsClient(context, connector)
	err = client.Delete(id)

	if err != nil {
		return handleDeleteError("DhcpServer", id, err)
	}

	return nil
}
