/* Copyright © 2024 Broadcom, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var projectIpAddressAllocationSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(true, false, false)),
	"allocation_ips": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateCidrOrIPOrRange(),
			Optional:     true,
			Computed:     true,
			ForceNew:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "AllocationIps",
			OmitIfEmpty:  true,
		},
	},
	"allocation_size": {
		Schema: schema.Schema{
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			ForceNew: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "int",
			SdkFieldName: "AllocationSize",
			OmitIfEmpty:  true,
		},
	},
	"ip_block": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
			ForceNew:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "IpBlock",
		},
	},
}

func resourceNsxtProjectIpAddressAllocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtProjectIpAddressAllocationCreate,
		Read:   resourceNsxtProjectIpAddressAllocationRead,
		Update: resourceNsxtProjectIpAddressAllocationUpdate,
		Delete: resourceNsxtProjectIpAddressAllocationDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(projectIpAddressAllocationSchema),
	}
}

func resourceNsxtProjectIpAddressAllocationExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	parents := getVpcParentsFromContext(sessionContext)
	client := clientLayer.NewIpAddressAllocationsClient(connector)
	_, err = client.Get(parents[0], parents[1], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtProjectIpAddressAllocationCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID2(d, m, resourceNsxtProjectIpAddressAllocationExists)
	if err != nil {
		return err
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.ProjectIpAddressAllocation{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, projectIpAddressAllocationSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating ProjectIpAddressAllocation with ID %s", id)

	client := clientLayer.NewIpAddressAllocationsClient(connector)
	err = client.Patch(parents[0], parents[1], id, obj)
	if err != nil {
		return handleCreateError("ProjectIpAddressAllocation", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtProjectIpAddressAllocationRead(d, m)
}

func resourceNsxtProjectIpAddressAllocationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ProjectIpAddressAllocation ID")
	}

	client := clientLayer.NewIpAddressAllocationsClient(connector)
	parents := getVpcParentsFromContext(getSessionContext(d, m))
	obj, err := client.Get(parents[0], parents[1], id)
	if err != nil {
		return handleReadError(d, "ProjectIpAddressAllocation", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, projectIpAddressAllocationSchema, "", nil)
}

func resourceNsxtProjectIpAddressAllocationUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ProjectIpAddressAllocation ID")
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.ProjectIpAddressAllocation{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	// Only the above attributes can be updated, others force recreation
	client := clientLayer.NewIpAddressAllocationsClient(connector)
	_, err := client.Update(parents[0], parents[1], id, obj)
	if err != nil {
		return handleUpdateError("ProjectIpAddressAllocation", id, err)
	}

	return resourceNsxtProjectIpAddressAllocationRead(d, m)
}

func resourceNsxtProjectIpAddressAllocationDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ProjectIpAddressAllocation ID")
	}

	connector := getPolicyConnector(m)
	parents := getVpcParentsFromContext(getSessionContext(d, m))

	client := clientLayer.NewIpAddressAllocationsClient(connector)
	err := client.Delete(parents[0], parents[1], id)

	if err != nil {
		return handleDeleteError("ProjectIpAddressAllocation", id, err)
	}

	return nil
}