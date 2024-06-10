/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var macDiscoveryProfileMacLimitPolicyValues = []string{
	model.MacDiscoveryProfile_MAC_LIMIT_POLICY_ALLOW,
	model.MacDiscoveryProfile_MAC_LIMIT_POLICY_DROP,
}

var macDiscoveryProfileSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(false, false, false)),
	"mac_change_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "MacChangeEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"mac_learning_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "MacLearningEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"mac_limit": {
		Schema: schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(0, 4096),
			Default:      4096,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "int",
			SdkFieldName: "MacLimit",
			TestData: metadata.Testdata{
				CreateValue: "20",
				UpdateValue: "50",
			},
		},
	},
	"mac_limit_policy": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(macDiscoveryProfileMacLimitPolicyValues, false),
			Optional:     true,
			Default:      model.MacDiscoveryProfile_MAC_LIMIT_POLICY_ALLOW,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "MacLimitPolicy",
			TestData: metadata.Testdata{
				CreateValue: model.MacDiscoveryProfile_MAC_LIMIT_POLICY_ALLOW,
				UpdateValue: model.MacDiscoveryProfile_MAC_LIMIT_POLICY_DROP,
			},
		},
	},
	"remote_overlay_mac_limit": {
		Schema: schema.Schema{
			Type:         schema.TypeInt,
			Optional:     true,
			ValidateFunc: validation.IntBetween(2048, 8192),
			Default:      2048,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "int",
			SdkFieldName: "RemoteOverlayMacLimit",
			TestData: metadata.Testdata{
				CreateValue: "2048",
				UpdateValue: "4096",
			},
		},
	},
	"unknown_unicast_flooding_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "UnknownUnicastFloodingEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
}

func resourceNsxtPolicyMacDiscoveryProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyMacDiscoveryProfileCreate,
		Read:   resourceNsxtPolicyMacDiscoveryProfileRead,
		Update: resourceNsxtPolicyMacDiscoveryProfileUpdate,
		Delete: resourceNsxtPolicyMacDiscoveryProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: metadata.GetSchemaFromExtendedSchema(macDiscoveryProfileSchema),
	}
}

func resourceNsxtPolicyMacDiscoveryProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	client := infra.NewMacDiscoveryProfilesClient(sessionContext, connector)
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyMacDiscoveryProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyMacDiscoveryProfileExists)
	if err != nil {
		return err
	}

	// TODO - consider including standard object attributes in the schema
	tags := getPolicyTagsFromSchema(d)
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)

	obj := model.MacDiscoveryProfile{
		Tags:        tags,
		DisplayName: &displayName,
		Description: &description,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, macDiscoveryProfileSchema, "", nil); err != nil {
		return err
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating MacDiscoveryProfile with ID %s", id)
	boolFalse := false
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewMacDiscoveryProfilesClient(context, connector)
	err = client.Patch(id, obj, &boolFalse)
	if err != nil {
		return handleCreateError("MacDiscoveryProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyMacDiscoveryProfileRead(d, m)
}

func resourceNsxtPolicyMacDiscoveryProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining MacDiscoveryProfile ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewMacDiscoveryProfilesClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "MacDiscoveryProfile", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, macDiscoveryProfileSchema, "", nil)
}

func resourceNsxtPolicyMacDiscoveryProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining MacDiscoveryProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.MacDiscoveryProfile{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, macDiscoveryProfileSchema, "", nil); err != nil {
		return err
	}

	// Update the resource using PATCH
	boolFalse := false
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewMacDiscoveryProfilesClient(context, connector)
	_, err = client.Update(id, obj, &boolFalse)
	if err != nil {
		return handleUpdateError("MacDiscoveryProfile", id, err)
	}

	return resourceNsxtPolicyMacDiscoveryProfileRead(d, m)
}

func resourceNsxtPolicyMacDiscoveryProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining MacDiscoveryProfile ID")
	}

	connector := getPolicyConnector(m)
	boolFalse := false
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewMacDiscoveryProfilesClient(context, connector)
	err = client.Delete(id, &boolFalse)

	if err != nil {
		return handleDeleteError("MacDiscoveryProfile", id, err)
	}

	return nil
}
