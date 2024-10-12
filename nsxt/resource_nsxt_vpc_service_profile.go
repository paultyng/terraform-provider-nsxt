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

var vpcServiceProfileSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(true, false, false)),
	"mac_discovery_profile": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "MacDiscoveryProfile",
		},
	},
	"spoof_guard_profile": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "SpoofGuardProfile",
		},
	},
	"dhcp_config": {
		Schema: schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"dhcp_relay_config": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"server_addresses": {
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
											SdkFieldName: "ServerAddresses",
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "DhcpRelayConfig",
							ReflectType:  reflect.TypeOf(model.VpcDhcpRelayConfig{}),
						},
					},
					"dhcp_server_config": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"dns_client_config": {
										Schema: schema.Schema{
											Type:     schema.TypeList,
											MaxItems: 1,
											Elem: &metadata.ExtendedResource{
												Schema: map[string]*metadata.ExtendedSchema{
													"dns_server_ips": {
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
															SdkFieldName: "DnsServerIps",
														},
													},
												},
											},
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "struct",
											SdkFieldName: "DnsClientConfig",
											ReflectType:  reflect.TypeOf(model.DnsClientConfig{}),
										},
									},
									"lease_time": {
										Schema: schema.Schema{
											Type:     schema.TypeInt,
											Optional: true,
											Default:  86400,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "int",
											SdkFieldName: "LeaseTime",
										},
									},
									"ntp_servers": {
										Schema: schema.Schema{
											Type: schema.TypeList,
											Elem: &metadata.ExtendedSchema{
												Schema: schema.Schema{
													Type:         schema.TypeString,
													ValidateFunc: validateSingleIPOrHostName(),
												},
												Metadata: metadata.Metadata{
													SchemaType: "string",
												},
											},
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "list",
											SdkFieldName: "NtpServers",
										},
									},
									"advanced_config": {
										Schema: schema.Schema{
											Type:     schema.TypeList,
											MaxItems: 1,
											Elem: &metadata.ExtendedResource{
												Schema: map[string]*metadata.ExtendedSchema{
													"is_distributed_dhcp": {
														Schema: schema.Schema{
															Type:     schema.TypeBool,
															Optional: true,
															Computed: true,
														},
														Metadata: metadata.Metadata{
															SchemaType:   "bool",
															SdkFieldName: "IsDistributedDhcp",
														},
													},
												},
											},
											Optional: true,
											Computed: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "struct",
											SdkFieldName: "AdvancedConfig",
											ReflectType:  reflect.TypeOf(model.VpcDhcpAdvancedConfig{}),
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "DhcpServerConfig",
							ReflectType:  reflect.TypeOf(model.VpcDhcpServerConfig{}),
						},
					},
				},
			},
		},
		Metadata: metadata.Metadata{
			SchemaType:   "struct",
			SdkFieldName: "DhcpConfig",
			ReflectType:  reflect.TypeOf(model.VpcProfileDhcpConfig{}),
		},
	},
	"ip_discovery_profile": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "IpDiscoveryProfile",
		},
	},
	"security_profile": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "SecurityProfile",
		},
	},
	"qos_profile": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validatePolicyPath(),
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "QosProfile",
		},
	},
}

func resourceNsxtVpcServiceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVpcServiceProfileCreate,
		Read:   resourceNsxtVpcServiceProfileRead,
		Update: resourceNsxtVpcServiceProfileUpdate,
		Delete: resourceNsxtVpcServiceProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathOnlyResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(vpcServiceProfileSchema),
	}
}

func resourceNsxtVpcServiceProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	parents := getVpcParentsFromContext(sessionContext)
	client := clientLayer.NewVpcServiceProfilesClient(connector)
	_, err = client.Get(parents[0], parents[1], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtVpcServiceProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID2(d, m, resourceNsxtVpcServiceProfileExists)
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

	obj := model.VpcServiceProfile{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcServiceProfileSchema, "", nil); err != nil {
		return err
	}

	if obj.DhcpConfig == nil {
		// NSX requires to send this struct even if empty
		obj.DhcpConfig = &model.VpcProfileDhcpConfig{}
	}

	log.Printf("[INFO] Creating VpcServiceProfile with ID %s", id)

	client := clientLayer.NewVpcServiceProfilesClient(connector)
	err = client.Patch(parents[0], parents[1], id, obj)
	if err != nil {
		return handleCreateError("VpcServiceProfile", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVpcServiceProfileRead(d, m)
}

func resourceNsxtVpcServiceProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcServiceProfile ID")
	}

	client := clientLayer.NewVpcServiceProfilesClient(connector)
	parents := getVpcParentsFromContext(getSessionContext(d, m))
	obj, err := client.Get(parents[0], parents[1], id)
	if err != nil {
		return handleReadError(d, "VpcServiceProfile", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, vpcServiceProfileSchema, "", nil)
}

func resourceNsxtVpcServiceProfileUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcServiceProfile ID")
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags, tagErr := getValidatedTagsFromSchema(d)
	if tagErr != nil {
		return tagErr
	}

	revision := int64(d.Get("revision").(int))

	obj := model.VpcServiceProfile{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcServiceProfileSchema, "", nil); err != nil {
		return err
	}

	if obj.DhcpConfig == nil {
		// NSX requires to send this struct even if empty
		obj.DhcpConfig = &model.VpcProfileDhcpConfig{}
	}

	client := clientLayer.NewVpcServiceProfilesClient(connector)
	_, err := client.Update(parents[0], parents[1], id, obj)
	if err != nil {
		// Trigger partial update to avoid terraform updating state based on failed intent
		// TODO - move this into handleUpdateError
		d.Partial(true)
		return handleUpdateError("VpcServiceProfile", id, err)
	}

	return resourceNsxtVpcServiceProfileRead(d, m)
}

func resourceNsxtVpcServiceProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcServiceProfile ID")
	}

	connector := getPolicyConnector(m)
	parents := getVpcParentsFromContext(getSessionContext(d, m))

	client := clientLayer.NewVpcServiceProfilesClient(connector)
	err := client.Delete(parents[0], parents[1], id)

	if err != nil {
		return handleDeleteError("VpcServiceProfile", id, err)
	}

	return nil
}
