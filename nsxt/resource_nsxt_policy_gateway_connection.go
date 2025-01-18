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
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var gatewayConnectionSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"advertise_outbound_route_filters": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: &metadata.ExtendedSchema{
				Schema: schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePolicyPath(),
				},
				Metadata: metadata.Metadata{
					SchemaType: "string",
				},
			},
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "list",
			SdkFieldName: "AdvertiseOutboundRouteFilters",
		},
	},
	"aggregate_routes": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: &metadata.ExtendedSchema{
				Schema: schema.Schema{
					Type: schema.TypeString,
				},
				Metadata: metadata.Metadata{
					SchemaType: "string",
				},
			},
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "list",
			SdkFieldName: "AggregateRoutes",
		},
	},
	"tier0_path": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validatePolicyPath(),
			Required:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "Tier0Path",
		},
	},
}

func resourceNsxtPolicyGatewayConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayConnectionCreate,
		Read:   resourceNsxtPolicyGatewayConnectionRead,
		Update: resourceNsxtPolicyGatewayConnectionUpdate,
		Delete: resourceNsxtPolicyGatewayConnectionDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathOnlyResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(gatewayConnectionSchema),
	}
}

func resourceNsxtPolicyGatewayConnectionExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error

	client := clientLayer.NewGatewayConnectionsClient(connector)
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyGatewayConnectionCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID(d, m, resourceNsxtPolicyGatewayConnectionExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.GatewayConnection{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, gatewayConnectionSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating GatewayConnection with ID %s", id)

	client := clientLayer.NewGatewayConnectionsClient(connector)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("GatewayConnection", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayConnectionRead(d, m)
}

func resourceNsxtPolicyGatewayConnectionRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayConnection ID")
	}

	client := clientLayer.NewGatewayConnectionsClient(connector)

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "GatewayConnection", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, gatewayConnectionSchema, "", nil)
}

func resourceNsxtPolicyGatewayConnectionUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayConnection ID")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.GatewayConnection{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, gatewayConnectionSchema, "", nil); err != nil {
		return err
	}
	client := clientLayer.NewGatewayConnectionsClient(connector)
	_, err := client.Update(id, obj)
	if err != nil {
		return handleUpdateError("GatewayConnection", id, err)
	}

	return resourceNsxtPolicyGatewayConnectionRead(d, m)
}

func resourceNsxtPolicyGatewayConnectionDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayConnection ID")
	}

	connector := getPolicyConnector(m)

	client := clientLayer.NewGatewayConnectionsClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("GatewayConnection", id, err)
	}

	return nil
}
