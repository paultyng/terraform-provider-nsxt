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
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/vpcs"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var staticRoutesSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(true, false, true)),
	"next_hop": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"ip_address": {
						Schema: schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validateSingleIP(),
							Optional:     true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "string",
							SdkFieldName: "IpAddress",
						},
					},
					"admin_distance": {
						Schema: schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "int",
							SdkFieldName: "AdminDistance",
						},
					},
				},
			},
			Required: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "list",
			SdkFieldName: "NextHops",
			ReflectType:  reflect.TypeOf(model.RouterNexthop{}),
		},
	},
	"network": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateCidrOrIPOrRange(),
			Required:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "Network",
		},
	},
}

func resourceNsxtVpcStaticRoutes() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVpcStaticRoutesCreate,
		Read:   resourceNsxtVpcStaticRoutesRead,
		Update: resourceNsxtVpcStaticRoutesUpdate,
		Delete: resourceNsxtVpcStaticRoutesDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPCPathResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(staticRoutesSchema),
	}
}

func resourceNsxtVpcStaticRoutesExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	parents := getVpcParentsFromContext(sessionContext)
	client := clientLayer.NewStaticRoutesClient(connector)
	_, err = client.Get(parents[0], parents[1], parents[2], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtVpcStaticRoutesCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID2(d, m, resourceNsxtVpcStaticRoutesExists)
	if err != nil {
		return err
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.StaticRoutes{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, staticRoutesSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating StaticRoutes with ID %s", id)

	client := clientLayer.NewStaticRoutesClient(connector)
	err = client.Patch(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		return handleCreateError("StaticRoutes", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVpcStaticRoutesRead(d, m)
}

func resourceNsxtVpcStaticRoutesRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining StaticRoutes ID")
	}

	client := clientLayer.NewStaticRoutesClient(connector)
	parents := getVpcParentsFromContext(getSessionContext(d, m))
	obj, err := client.Get(parents[0], parents[1], parents[2], id)
	if err != nil {
		return handleReadError(d, "StaticRoutes", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, staticRoutesSchema, "", nil)
}

func resourceNsxtVpcStaticRoutesUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining StaticRoutes ID")
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.StaticRoutes{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, staticRoutesSchema, "", nil); err != nil {
		return err
	}
	client := clientLayer.NewStaticRoutesClient(connector)
	_, err := client.Update(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		// Trigger partial update to avoid terraform updating state based on failed intent
		// TODO - move this into handleUpdateError
		d.Partial(true)
		return handleUpdateError("StaticRoutes", id, err)
	}

	return resourceNsxtVpcStaticRoutesRead(d, m)
}

func resourceNsxtVpcStaticRoutesDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining StaticRoutes ID")
	}

	connector := getPolicyConnector(m)
	parents := getVpcParentsFromContext(getSessionContext(d, m))

	client := clientLayer.NewStaticRoutesClient(connector)
	err := client.Delete(parents[0], parents[1], parents[2], id)

	if err != nil {
		return handleDeleteError("StaticRoutes", id, err)
	}

	return nil
}
