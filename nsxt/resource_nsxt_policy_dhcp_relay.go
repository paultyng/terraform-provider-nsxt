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

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyDhcpRelayConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDhcpRelayConfigCreate,
		Read:   resourceNsxtPolicyDhcpRelayConfigRead,
		Update: resourceNsxtPolicyDhcpRelayConfigUpdate,
		Delete: resourceNsxtPolicyDhcpRelayConfigDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"server_addresses": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
				},
			},
		},
	}
}

func resourceNsxtPolicyDhcpRelayConfigExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewDhcpRelayConfigsClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}

	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyDhcpRelayConfigCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDhcpRelayConfigsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyDhcpRelayConfigExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	serverAddresses := getStringListFromSchemaList(d, "server_addresses")

	obj := model.DhcpRelayConfig{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		ServerAddresses: serverAddresses,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating DhcpRelayConfig with ID %s", id)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("DhcpRelayConfig", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDhcpRelayConfigRead(d, m)
}

func resourceNsxtPolicyDhcpRelayConfigRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDhcpRelayConfigsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpRelayConfig ID")
	}

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "DhcpRelayConfig", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("server_addresses", obj.ServerAddresses)

	return nil
}

func resourceNsxtPolicyDhcpRelayConfigUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewDhcpRelayConfigsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpRelayConfig ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))

	serverAddresses := getStringListFromSchemaList(d, "server_addresses")

	obj := model.DhcpRelayConfig{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		ServerAddresses: serverAddresses,
		Revision:        &revision,
	}

	_, err := client.Update(id, obj)
	if err != nil {
		return handleUpdateError("DhcpRelayConfig", id, err)
	}

	return resourceNsxtPolicyDhcpRelayConfigRead(d, m)
}

func resourceNsxtPolicyDhcpRelayConfigDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DhcpRelayConfig ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewDhcpRelayConfigsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("DhcpRelayConfig", id, err)
	}

	return nil
}
