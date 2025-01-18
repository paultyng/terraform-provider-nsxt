// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var sharingStrategyVals = []string{
	model.Share_SHARING_STRATEGY_ALL_DESCENDANTS,
	model.Share_SHARING_STRATEGY_NONE_DESCENDANTS,
}

func resourceNsxtPolicyShare() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyShareCreate,
		Read:   resourceNsxtPolicyShareRead,
		Update: resourceNsxtPolicyShareUpdate,
		Delete: resourceNsxtPolicyShareDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathOnlyResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"shared_with": {
				Type:        schema.TypeList,
				Description: "Path of the context",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sharing_strategy": {
				Type:         schema.TypeString,
				Description:  "Sharing Strategy",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(sharingStrategyVals, false),
				Default:      model.Share_SHARING_STRATEGY_NONE_DESCENDANTS,
			},
		},
	}
}

func resourceNsxtPolicyShareExists(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	client := infra.NewSharesClient(context, connector)
	_, err = client.Get(id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyShareCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyShareExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	sharingStrategy := d.Get("sharing_strategy").(string)
	sharedWith := interface2StringList(d.Get("shared_with").([]interface{}))

	obj := model.Share{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		SharingStrategy: &sharingStrategy,
		SharedWith:      sharedWith,
	}
	context := getSessionContext(d, m)
	client := infra.NewSharesClient(context, connector)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("Share", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyShareRead(d, m)
}

func resourceNsxtPolicyShareRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Share ID")
	}

	context := getSessionContext(d, m)
	client := infra.NewSharesClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Share", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("shared_with", stringList2Interface(obj.SharedWith))
	d.Set("sharing_strategy", obj.SharingStrategy)

	return nil
}

func resourceNsxtPolicyShareUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Share ID")
	}

	connector := getPolicyConnector(m)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	sharingStrategy := d.Get("sharing_strategy").(string)
	sharedWith := interface2StringList(d.Get("shared_with").([]interface{}))

	obj := model.Share{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		SharingStrategy: &sharingStrategy,
		SharedWith:      sharedWith,
	}
	context := getSessionContext(d, m)
	client := infra.NewSharesClient(context, connector)
	err := client.Patch(id, obj)
	if err != nil {
		return handleCreateError("Share", id, err)
	}

	return resourceNsxtPolicyShareRead(d, m)
}

func resourceNsxtPolicyShareDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Share ID")
	}

	connector := getPolicyConnector(m)
	context := getSessionContext(d, m)
	client := infra.NewSharesClient(context, connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("Share", id, err)
	}

	return nil
}
