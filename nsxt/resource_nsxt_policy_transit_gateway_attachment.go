// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/transit_gateways"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var transitGatewayAttachmentSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"parent_path":  metadata.GetExtendedSchema(getPolicyPathSchema(true, true, "Policy path of the parent")),
	"connection_path": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validatePolicyPath(),
			Required:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "ConnectionPath",
			OmitIfEmpty:  true,
		},
	},
}

func resourceNsxtPolicyTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyTransitGatewayAttachmentCreate,
		Read:   resourceNsxtPolicyTransitGatewayAttachmentRead,
		Update: resourceNsxtPolicyTransitGatewayAttachmentUpdate,
		Delete: resourceNsxtPolicyTransitGatewayAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtParentPathResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(transitGatewayAttachmentSchema),
	}
}

func resourceNsxtPolicyTransitGatewayAttachmentExists(sessionContext utl.SessionContext, parentPath string, id string, connector client.Connector) (bool, error) {
	var err error
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return false, pathErr
	}
	client := clientLayer.NewAttachmentsClient(connector)
	_, err = client.Get(parents[0], parents[1], parents[2], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyTransitGatewayAttachmentCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateIDWithParent(d, m, resourceNsxtPolicyTransitGatewayAttachmentExists)
	if err != nil {
		return err
	}

	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return pathErr
	}
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.TransitGatewayAttachment{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, transitGatewayAttachmentSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating TransitGatewayAttachment with ID %s", id)

	client := clientLayer.NewAttachmentsClient(connector)
	err = client.Patch(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		return handleCreateError("TransitGatewayAttachment", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyTransitGatewayAttachmentRead(d, m)
}

func resourceNsxtPolicyTransitGatewayAttachmentRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining TransitGatewayAttachment ID")
	}

	client := clientLayer.NewAttachmentsClient(connector)
	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return pathErr
	}
	obj, err := client.Get(parents[0], parents[1], parents[2], id)
	if err != nil {
		return handleReadError(d, "TransitGatewayAttachment", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, transitGatewayAttachmentSchema, "", nil)
}

func resourceNsxtPolicyTransitGatewayAttachmentUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining TransitGatewayAttachment ID")
	}

	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return pathErr
	}
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.TransitGatewayAttachment{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, transitGatewayAttachmentSchema, "", nil); err != nil {
		return err
	}
	client := clientLayer.NewAttachmentsClient(connector)
	_, err := client.Update(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		return handleUpdateError("TransitGatewayAttachment", id, err)
	}

	return resourceNsxtPolicyTransitGatewayAttachmentRead(d, m)
}

func resourceNsxtPolicyTransitGatewayAttachmentDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining TransitGatewayAttachment ID")
	}

	connector := getPolicyConnector(m)
	parentPath := d.Get("parent_path").(string)
	parents, pathErr := parseStandardPolicyPathVerifySize(parentPath, 3)
	if pathErr != nil {
		return pathErr
	}

	client := clientLayer.NewAttachmentsClient(connector)
	err := client.Delete(parents[0], parents[1], parents[2], id)

	if err != nil {
		return handleDeleteError("TransitGatewayAttachment", id, err)
	}

	return nil
}
