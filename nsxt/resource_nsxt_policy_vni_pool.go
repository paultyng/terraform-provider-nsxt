// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyVniPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyVniPoolCreate,
		Read:   resourceNsxtPolicyVniPoolRead,
		Update: resourceNsxtPolicyVniPoolUpdate,
		Delete: resourceNsxtPolicyVniPoolDelete,
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
			"start": {
				Type:         schema.TypeInt,
				Description:  "Start value of VNI Pool range",
				Required:     true,
				ValidateFunc: validation.IntBetween(75001, 16777215),
			},
			"end": {
				Type:         schema.TypeInt,
				Description:  "End value of VNI Pool range",
				Required:     true,
				ValidateFunc: validation.IntBetween(75001, 16777215),
			},
		},
	}
}

func resourceNsxtPolicyVniPoolExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	client := infra.NewVniPoolsClient(connector)
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func policyVniPoolPatch(id string, d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	start := int64(d.Get("start").(int))
	end := int64(d.Get("end").(int))

	obj := model.VniPoolConfig{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Start:       &start,
		End:         &end,
	}

	// Create the resource using PATCH
	client := infra.NewVniPoolsClient(connector)
	return client.Patch(id, obj)
}

func resourceNsxtPolicyVniPoolCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyVniPoolExists)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating VNI Pool with ID %s", id)
	err = policyVniPoolPatch(id, d, m)
	if err != nil {
		return handleCreateError("VNI Pool", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyVniPoolRead(d, m)
}

func resourceNsxtPolicyVniPoolRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VNI Pool ID")
	}

	client := infra.NewVniPoolsClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "VNI Pool", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)

	d.Set("start", obj.Start)
	d.Set("end", obj.End)

	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	return nil
}

func resourceNsxtPolicyVniPoolUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VNI Pool ID")
	}

	log.Printf("[INFO] Creating VNI POOL with ID %s", id)
	err := policyVniPoolPatch(id, d, m)
	if err != nil {
		return handleUpdateError("VNI pool", id, err)
	}

	return resourceNsxtPolicyVniPoolRead(d, m)
}

func resourceNsxtPolicyVniPoolDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VNI Pool ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewVniPoolsClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("VNI Pool", id, err)
	}

	return nil
}
