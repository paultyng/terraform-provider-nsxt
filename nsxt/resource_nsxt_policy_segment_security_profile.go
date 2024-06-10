/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"
	"reflect"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var segmentSecurityProfileSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(false, false, false)),
	"bpdu_filter_allow": {
		Schema: schema.Schema{
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsMACAddress,
			},
			Optional: true,
		},
		Metadata: metadata.Metadata{
			Skip: true,
		},
	},
	"bpdu_filter_enable": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "BpduFilterEnable",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"dhcp_client_block_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "DhcpClientBlockEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"dhcp_server_block_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "DhcpServerBlockEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"dhcp_client_block_v6_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "DhcpClientBlockV6Enabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"dhcp_server_block_v6_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "DhcpServerBlockV6Enabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"non_ip_traffic_block_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "NonIpTrafficBlockEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"ra_guard_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "RaGuardEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"rate_limits_enabled": {
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "bool",
			SdkFieldName: "RateLimitsEnabled",
			TestData: metadata.Testdata{
				CreateValue: "true",
				UpdateValue: "false",
			},
		},
	},
	"rate_limit": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"rx_broadcast": {
						Schema: schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "int",
							SdkFieldName: "RxBroadcast",
							TestData: metadata.Testdata{
								CreateValue: "100",
								UpdateValue: "1000",
							},
						},
					},
					"rx_multicast": {
						Schema: schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "int",
							SdkFieldName: "RxMulticast",
							TestData: metadata.Testdata{
								CreateValue: "100",
								UpdateValue: "1000",
							},
						},
					},
					"tx_broadcast": {
						Schema: schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "int",
							SdkFieldName: "TxBroadcast",
							TestData: metadata.Testdata{
								CreateValue: "100",
								UpdateValue: "1000",
							},
						},
					},
					"tx_multicast": {
						Schema: schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "int",
							SdkFieldName: "TxMulticast",
							TestData: metadata.Testdata{
								CreateValue: "100",
								UpdateValue: "1000",
							},
						},
					},
				},
			},
			Optional: true,
			Computed: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "struct",
			SdkFieldName: "RateLimits",
			ReflectType:  reflect.TypeOf(model.TrafficRateLimits{}),
		},
	},
}

func resourceNsxtPolicySegmentSecurityProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySegmentSecurityProfileCreate,
		Read:   resourceNsxtPolicySegmentSecurityProfileRead,
		Update: resourceNsxtPolicySegmentSecurityProfileUpdate,
		Delete: resourceNsxtPolicySegmentSecurityProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: metadata.GetSchemaFromExtendedSchema(segmentSecurityProfileSchema),
	}
}

func resourceNsxtPolicySegmentSecurityProfileExists(context utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewSegmentSecurityProfilesClient(context, connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicySegmentSecurityProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	// example of attribute that we prefer to handle manually in the code
	// it is marked with `skip` in metadata
	bpduFilterAllow := getStringListFromSchemaSet(d, "bpdu_filter_allow")

	obj := model.SegmentSecurityProfile{
		DisplayName:     &displayName,
		Description:     &description,
		Tags:            tags,
		BpduFilterAllow: bpduFilterAllow,
	}
	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, segmentSecurityProfileSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Sending SegmentSecurityProfile with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSegmentSecurityProfilesClient(context, connector)
	return client.Patch(id, obj, nil)
}

func resourceNsxtPolicySegmentSecurityProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicySegmentSecurityProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicySegmentSecurityProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("SegmentSecurityProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySegmentSecurityProfileRead(d, m)
}

func resourceNsxtPolicySegmentSecurityProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SegmentSecurityProfile ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSegmentSecurityProfilesClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "SegmentSecurityProfile", id, err)
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.StructToSchema(elem, d, segmentSecurityProfileSchema, "", nil); err != nil {
		return err
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	return nil
}

func resourceNsxtPolicySegmentSecurityProfileUpdate(d *schema.ResourceData, m interface{}) error {
	log.Printf("[INFO] updating SegmentSecurityProfile")
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SegmentSecurityProfile ID")
	}

	err := resourceNsxtPolicySegmentSecurityProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("SegmentSecurityProfile", id, err)
	}

	return resourceNsxtPolicySegmentSecurityProfileRead(d, m)
}

func resourceNsxtPolicySegmentSecurityProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining SegmentSecurityProfile ID")
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewSegmentSecurityProfilesClient(context, connector)
	err = client.Delete(id, nil)

	if err != nil {
		return handleDeleteError("SegmentSecurityProfile", id, err)
	}

	return nil
}
