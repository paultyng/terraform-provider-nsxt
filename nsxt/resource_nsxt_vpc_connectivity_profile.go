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
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var vpcConnectivityProfileSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(true, false, false)),
	"transit_gateway_path": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validatePolicyPath(),
			Required:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "TransitGatewayPath",
		},
	},
	"private_tgw_ip_blocks": {
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
			SdkFieldName: "PrivateTgwIpBlocks",
		},
	},
	"external_ip_blocks": {
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
			SdkFieldName: "ExternalIpBlocks",
		},
	},
	"service_gateway": {
		Schema: schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"nat_config": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"enable_default_snat": {
										Schema: schema.Schema{
											Type:     schema.TypeBool,
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "bool",
											SdkFieldName: "EnableDefaultSnat",
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "NatConfig",
							ReflectType:  reflect.TypeOf(model.VpcNatConfig{}),
						},
					},
					"qos_config": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"ingress_qos_profile_path": {
										Schema: schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validatePolicyPath(),
											Optional:     true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "string",
											SdkFieldName: "IngressQosProfilePath",
										},
									},
									"egress_qos_profile_path": {
										Schema: schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validatePolicyPath(),
											Optional:     true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "string",
											SdkFieldName: "EgressQosProfilePath",
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "QosConfig",
							ReflectType:  reflect.TypeOf(model.GatewayQosProfileConfig{}),
						},
					},
					"enable": {
						Schema: schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "bool",
							SdkFieldName: "Enable",
						},
					},
					"edge_cluster_paths": {
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
							SdkFieldName: "EdgeClusterPaths",
						},
					},
				},
			},
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "struct",
			SdkFieldName: "ServiceGateway",
			ReflectType:  reflect.TypeOf(model.VpcServiceGatewayConfig{}),
		},
	},
}

func resourceNsxtVpcConnectivityProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVpcConnectivityProfileCreate,
		Read:   resourceNsxtVpcConnectivityProfileRead,
		Update: resourceNsxtVpcConnectivityProfileUpdate,
		Delete: resourceNsxtVpcConnectivityProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathOnlyResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(vpcConnectivityProfileSchema),
	}
}

func resourceNsxtVpcConnectivityProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	parents := getVpcParentsFromContext(sessionContext)
	client := clientLayer.NewVpcConnectivityProfilesClient(connector)
	_, err = client.Get(parents[0], parents[1], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtVpcConnectivityProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID2(d, m, resourceNsxtVpcConnectivityProfileExists)
	if err != nil {
		return err
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags, tagErr := getValidatedTagsFromSchema(d)
	if tagErr != nil {
		return tagErr
	}

	obj := model.VpcConnectivityProfile{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcConnectivityProfileSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating VpcConnectivityProfile with ID %s", id)

	client := clientLayer.NewVpcConnectivityProfilesClient(connector)
	err = client.Patch(parents[0], parents[1], id, obj)
	if err != nil {
		return handleCreateError("VpcConnectivityProfile", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVpcConnectivityProfileRead(d, m)
}

func resourceNsxtVpcConnectivityProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcConnectivityProfile ID")
	}

	client := clientLayer.NewVpcConnectivityProfilesClient(connector)
	parents := getVpcParentsFromContext(getSessionContext(d, m))
	obj, err := client.Get(parents[0], parents[1], id)
	if err != nil {
		return handleReadError(d, "VpcConnectivityProfile", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, vpcConnectivityProfileSchema, "", nil)
}

func resourceNsxtVpcConnectivityProfileUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcConnectivityProfile ID")
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags, tagErr := getValidatedTagsFromSchema(d)
	if tagErr != nil {
		return tagErr
	}

	revision := int64(d.Get("revision").(int))

	obj := model.VpcConnectivityProfile{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcConnectivityProfileSchema, "", nil); err != nil {
		return err
	}
	client := clientLayer.NewVpcConnectivityProfilesClient(connector)
	_, err := client.Update(parents[0], parents[1], id, obj)
	if err != nil {
		// Trigger partial update to avoid terraform updating state based on failed intent
		// TODO - move this into handleUpdateError
		d.Partial(true)
		return handleUpdateError("VpcConnectivityProfile", id, err)
	}

	return resourceNsxtVpcConnectivityProfileRead(d, m)
}

func resourceNsxtVpcConnectivityProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcConnectivityProfile ID")
	}

	connector := getPolicyConnector(m)
	parents := getVpcParentsFromContext(getSessionContext(d, m))

	client := clientLayer.NewVpcConnectivityProfilesClient(connector)
	err := client.Delete(parents[0], parents[1], id)

	if err != nil {
		return handleDeleteError("VpcConnectivityProfile", id, err)
	}

	return nil
}
