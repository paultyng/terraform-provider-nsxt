// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-nsxt/api/infra/shares"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicySharedResource() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySharedResourceCreate,
		Read:   resourceNsxtPolicySharedResourceRead,
		Update: resourceNsxtPolicySharedResourceUpdate,
		Delete: resourceNsxtPolicySharedResourceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicySharedResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"share_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Share ID to associate the resource to",
				ForceNew:    true,
			},
			"resource_object": {
				Type:        schema.TypeList,
				Description: "List of resources to be shared",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_children": {
							Type:        schema.TypeBool,
							Description: "Denotes if the children of the shared path are also shared",
							Optional:    true,
							Default:     false,
						},
						"resource_path": {
							Type:        schema.TypeString,
							Description: "Path of the resource to be shared",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicySharedResourceCreate(d *schema.ResourceData, m interface{}) error {
	id := newUUID()

	connector := getPolicyConnector(m)
	sharePath := d.Get("share_path").(string)
	shareID := getPolicyIDFromPath(sharePath)
	context := getParentContext(d, m, sharePath)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	var resourceObjects []model.ResourceObject
	for _, o := range d.Get("resource_object").([]interface{}) {
		ro := o.(map[string]interface{})
		includeChildren := ro["include_children"].(bool)
		resourcePath := ro["resource_path"].(string)
		resourceObjects = append(resourceObjects, model.ResourceObject{
			IncludeChildren: &includeChildren,
			ResourcePath:    &resourcePath,
		})
	}

	obj := model.SharedResource{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		ResourceObjects: resourceObjects,
	}
	client := shares.NewResourcesClient(context, connector)
	err := client.Patch(shareID, id, obj)
	if err != nil {
		return handleCreateError("Shared Resource", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySharedResourceRead(d, m)
}

func resourceNsxtPolicySharedResourceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Shared Resource ID")
	}

	sharePath := d.Get("share_path").(string)
	shareID := getPolicyIDFromPath(sharePath)
	context := getParentContext(d, m, sharePath)
	client := shares.NewResourcesClient(context, connector)
	obj, err := client.Get(shareID, id)
	if err != nil {
		return handleReadError(d, "Shared Resource", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	var resourceObjects []map[string]interface{}
	for _, ro := range obj.ResourceObjects {
		elem := make(map[string]interface{})
		elem["include_children"] = ro.IncludeChildren
		elem["resource_path"] = ro.ResourcePath
		resourceObjects = append(resourceObjects, elem)
	}
	d.Set("resource_object", resourceObjects)

	return nil
}

func resourceNsxtPolicySharedResourceUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()

	connector := getPolicyConnector(m)
	sharePath := d.Get("share_path").(string)
	shareID := getPolicyIDFromPath(sharePath)
	context := getParentContext(d, m, sharePath)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	var resourceObjects []model.ResourceObject
	for _, o := range d.Get("resource_object").([]interface{}) {
		ro := o.(map[string]interface{})
		includeChildren := ro["include_children"].(bool)
		resourcePath := ro["resource_path"].(string)
		resourceObjects = append(resourceObjects, model.ResourceObject{
			IncludeChildren: &includeChildren,
			ResourcePath:    &resourcePath,
		})
	}

	obj := model.SharedResource{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		ResourceObjects: resourceObjects,
	}
	client := shares.NewResourcesClient(context, connector)
	err := client.Patch(shareID, id, obj)
	if err != nil {
		return handleCreateError("Shared Resource", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySharedResourceRead(d, m)
}

func resourceNsxtPolicySharedResourceDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Shared Resource ID")
	}
	sharePath := d.Get("share_path").(string)
	shareID := getPolicyIDFromPath(sharePath)

	connector := getPolicyConnector(m)
	context := getParentContext(d, m, sharePath)
	client := shares.NewResourcesClient(context, connector)
	err := client.Delete(shareID, id)

	if err != nil {
		return handleDeleteError("Shared Resource", id, err)
	}

	return nil

}

func resourceNsxtPolicySharedResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}

	sharePath, err := getParameterFromPolicyPath("", "/resources/", importID)
	if err != nil {
		return nil, err
	}
	d.Set("share_path", sharePath)
	return rd, nil
}
