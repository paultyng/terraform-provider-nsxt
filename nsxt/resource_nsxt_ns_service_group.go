// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
)

func resourceNsxtNsServiceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtNsServiceGroupCreate,
		Read:   resourceNsxtNsServiceGroupRead,
		Update: resourceNsxtNsServiceGroupUpdate,
		Delete: resourceNsxtNsServiceGroupDelete,
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
			"members": {
				Type:        schema.TypeSet,
				Description: "List of NSService or NSServiceGroup resources that can be added as members to an NSServiceGroup",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
		},
	}
}

func getResourceReferencesFromStringsSet(d *schema.ResourceData, schemaAttrName string) []common.ResourceReference {
	targetIds := interface2StringList(d.Get(schemaAttrName).(*schema.Set).List())
	var referenceList []common.ResourceReference
	for _, targetID := range targetIds {
		elem := common.ResourceReference{
			IsValid:    true,
			TargetId:   targetID,
			TargetType: "NSService",
		}

		referenceList = append(referenceList, elem)
	}
	return referenceList
}

func returnResourceReferencesTargetIDs(references []common.ResourceReference) []string {
	var TargetIds []string
	for _, reference := range references {
		TargetIds = append(TargetIds, reference.TargetId)
	}
	return TargetIds
}

func resourceNsxtNsServiceGroupCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	members := getResourceReferencesFromStringsSet(d, "members")
	nsServiceGroup := manager.NsServiceGroup{
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		Members:     members,
	}

	nsServiceGroup, resp, err := nsxClient.GroupingObjectsApi.CreateNSServiceGroup(nsxClient.Context, nsServiceGroup)

	if err != nil {
		return fmt.Errorf("Error during NsServiceGroup create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during NsServiceGroup create: %v", resp.StatusCode)
	}
	d.SetId(nsServiceGroup.Id)

	return resourceNsxtNsServiceGroupRead(d, m)
}

func resourceNsxtNsServiceGroupRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	nsServiceGroup, resp, err := nsxClient.GroupingObjectsApi.ReadNSServiceGroup(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsServiceGroup %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during NsServiceGroup read: %v", err)
	}

	d.Set("revision", nsServiceGroup.Revision)
	d.Set("description", nsServiceGroup.Description)
	d.Set("display_name", nsServiceGroup.DisplayName)
	setTagsInSchema(d, nsServiceGroup.Tags)
	d.Set("members", returnResourceReferencesTargetIDs(nsServiceGroup.Members))

	return nil
}

func resourceNsxtNsServiceGroupUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	members := getResourceReferencesFromStringsSet(d, "members")
	nsServiceGroup := manager.NsServiceGroup{
		Revision:    revision,
		Description: description,
		DisplayName: displayName,
		Tags:        tags,
		Members:     members,
	}

	_, resp, err := nsxClient.GroupingObjectsApi.UpdateNSServiceGroup(nsxClient.Context, id, nsServiceGroup)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during NsServiceGroup update: %v", err)
	}

	return resourceNsxtNsServiceGroupRead(d, m)
}

func resourceNsxtNsServiceGroupDelete(d *schema.ResourceData, m interface{}) error {

	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	localVarOptionals := make(map[string]interface{})
	localVarOptionals["force"] = true
	resp, err := nsxClient.GroupingObjectsApi.DeleteNSServiceGroup(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during NsServiceGroup delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsServiceGroup %s not found", id)
		d.SetId("")
	}

	return nil
}
