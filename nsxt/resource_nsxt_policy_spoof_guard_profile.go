/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

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

func resourceNsxtPolicySpoofGuardProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySpoofGuardProfileCreate,
		Read:   resourceNsxtPolicySpoofGuardProfileRead,
		Update: resourceNsxtPolicySpoofGuardProfileUpdate,
		Delete: resourceNsxtPolicySpoofGuardProfileDelete,
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
			"address_binding_allowlist": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func resourceNsxtPolicySpoofGuardProfileExists(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewSpoofguardProfilesClient(context, connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicySpoofGuardProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	addressBindingAllowlist := d.Get("address_binding_allowlist").(bool)

	obj := model.SpoofGuardProfile{
		DisplayName:             &displayName,
		Description:             &description,
		Tags:                    tags,
		AddressBindingAllowlist: &addressBindingAllowlist,
	}

	log.Printf("[INFO] Patching SpoofGuardProfile with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSpoofguardProfilesClient(context, connector)
	return client.Patch(id, obj, nil)
}

func resourceNsxtPolicySpoofGuardProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicySpoofGuardProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicySpoofGuardProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("SpoofGuardProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySpoofGuardProfileRead(d, m)
}

func resourceNsxtPolicySpoofGuardProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SpoofGuardProfile ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSpoofguardProfilesClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "SpoofGuardProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("address_binding_allowlist", obj.AddressBindingAllowlist)

	return nil
}

func resourceNsxtPolicySpoofGuardProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SpoofGuardProfile ID")
	}

	err := resourceNsxtPolicySpoofGuardProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("SpoofGuardProfile", id, err)
	}

	return resourceNsxtPolicySpoofGuardProfileRead(d, m)
}

func resourceNsxtPolicySpoofGuardProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SpoofGuardProfile ID")
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSpoofguardProfilesClient(context, connector)
	err = client.Delete(id, nil)

	if err != nil {
		return handleDeleteError("SpoofGuardProfile", id, err)
	}

	return nil
}
