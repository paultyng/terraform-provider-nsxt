// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	clientLayer "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/vpcs"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/vpcs/subnets"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/metadata"
)

var vpcSubnetAccessModeValues = []string{
	model.VpcSubnet_ACCESS_MODE_PRIVATE,
	model.VpcSubnet_ACCESS_MODE_PUBLIC,
	model.VpcSubnet_ACCESS_MODE_ISOLATED,
	model.VpcSubnet_ACCESS_MODE_PRIVATE_TGW,
}

var vpcSubnetConnectivityStateValues = []string{
	model.SubnetAdvancedConfig_CONNECTIVITY_STATE_CONNECTED,
	model.SubnetAdvancedConfig_CONNECTIVITY_STATE_DISCONNECTED,
}

var vpcSubnetModeValues = []string{
	model.SubnetDhcpConfig_MODE_SERVER,
	model.SubnetDhcpConfig_MODE_RELAY,
	model.SubnetDhcpConfig_MODE_DEACTIVATED,
}

var vpcSubnetSchema = map[string]*metadata.ExtendedSchema{
	"nsx_id":       metadata.GetExtendedSchema(getNsxIDSchema()),
	"path":         metadata.GetExtendedSchema(getPathSchema()),
	"display_name": metadata.GetExtendedSchema(getDisplayNameSchema()),
	"description":  metadata.GetExtendedSchema(getDescriptionSchema()),
	"revision":     metadata.GetExtendedSchema(getRevisionSchema()),
	"tag":          metadata.GetExtendedSchema(getTagsSchema()),
	"context":      metadata.GetExtendedSchema(getContextSchema(true, false, true)),
	"ip_blocks": {
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
			ForceNew: true,
			Computed: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "list",
			SdkFieldName: "IpBlocks",
		},
	},
	"advanced_config": {
		Schema: schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"gateway_addresses": {
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
							SdkFieldName: "GatewayAddresses",
						},
					},
					"extra_config": {
						Schema: schema.Schema{
							Type: schema.TypeList,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"config_pair": {
										Schema: schema.Schema{
											Type:     schema.TypeList,
											MaxItems: 1,
											Elem: &metadata.ExtendedResource{
												Schema: map[string]*metadata.ExtendedSchema{
													"value": {
														Schema: schema.Schema{
															Type:     schema.TypeString,
															Required: true,
														},
														Metadata: metadata.Metadata{
															SchemaType:   "string",
															SdkFieldName: "Value",
														},
													},
													"key": {
														Schema: schema.Schema{
															Type:     schema.TypeString,
															Required: true,
														},
														Metadata: metadata.Metadata{
															SchemaType:   "string",
															SdkFieldName: "Key",
														},
													},
												},
											},
											Required: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "struct",
											SdkFieldName: "ConfigPair",
											ReflectType:  reflect.TypeOf(model.UnboundedKeyValuePair{}),
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "list",
							SdkFieldName: "ExtraConfigs",
							ReflectType:  reflect.TypeOf(model.SubnetExtraConfig{}),
							OmitIfEmpty:  true,
						},
					},
					"dhcp_server_addresses": {
						Schema: schema.Schema{
							Type: schema.TypeList,
							Elem: &metadata.ExtendedSchema{
								Schema: schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateSingleIP(),
								},
								Metadata: metadata.Metadata{
									SchemaType: "string",
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "list",
							SdkFieldName: "DhcpServerAddresses",
						},
					},
					"connectivity_state": {
						Schema: schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(vpcSubnetConnectivityStateValues, false),
							Optional:     true,
							Default:      model.SubnetAdvancedConfig_CONNECTIVITY_STATE_CONNECTED,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "string",
							SdkFieldName: "ConnectivityState",
						},
					},
					"static_ip_allocation": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"enabled": {
										Schema: schema.Schema{
											Type:     schema.TypeBool,
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "bool",
											SdkFieldName: "Enabled",
										},
									},
								},
							},
							Optional: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "StaticIpAllocation",
							ReflectType:  reflect.TypeOf(model.StaticIpAllocation{}),
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
			ReflectType:  reflect.TypeOf(model.SubnetAdvancedConfig{}),
		},
	},
	"ipv4_subnet_size": {
		Schema: schema.Schema{
			Type:          schema.TypeInt,
			Optional:      true,
			Computed:      true,
			ForceNew:      true,
			ConflictsWith: []string{"ip_addresses"},
		},
		Metadata: metadata.Metadata{
			SchemaType:   "int",
			SdkFieldName: "Ipv4SubnetSize",
			OmitIfEmpty:  true,
		},
	},
	"ip_addresses": {
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
			Computed: true,
			ForceNew: true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "list",
			SdkFieldName: "IpAddresses",
		},
	},
	"access_mode": {
		Schema: schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validation.StringInSlice(vpcSubnetAccessModeValues, false),
			Optional:     true,
			Default:      model.VpcSubnet_ACCESS_MODE_PRIVATE,
			ForceNew:     true,
		},
		Metadata: metadata.Metadata{
			SchemaType:   "string",
			SdkFieldName: "AccessMode",
		},
	},
	"dhcp_config": {
		Schema: schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Computed: true,
			Elem: &metadata.ExtendedResource{
				Schema: map[string]*metadata.ExtendedSchema{
					"mode": {
						Schema: schema.Schema{
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(vpcSubnetModeValues, false),
							Optional:     true,
							Default:      model.SubnetDhcpConfig_MODE_DEACTIVATED,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "string",
							SdkFieldName: "Mode",
						},
					},
					"dhcp_server_additional_config": {
						Schema: schema.Schema{
							Type:     schema.TypeList,
							MaxItems: 1,
							Elem: &metadata.ExtendedResource{
								Schema: map[string]*metadata.ExtendedSchema{
									"options": {
										Schema: schema.Schema{
											Type:     schema.TypeList,
											MaxItems: 1,
											Elem: &metadata.ExtendedResource{
												Schema: map[string]*metadata.ExtendedSchema{
													"option121": {
														Schema: schema.Schema{
															Type:     schema.TypeList,
															MaxItems: 1,
															Elem: &metadata.ExtendedResource{
																Schema: map[string]*metadata.ExtendedSchema{
																	"static_route": {
																		Schema: schema.Schema{
																			Type: schema.TypeList,
																			Elem: &metadata.ExtendedResource{
																				Schema: map[string]*metadata.ExtendedSchema{
																					"next_hop": {
																						Schema: schema.Schema{
																							Type:         schema.TypeString,
																							ValidateFunc: validateSingleIP(),
																							Optional:     true,
																						},
																						Metadata: metadata.Metadata{
																							SchemaType:   "string",
																							SdkFieldName: "NextHop",
																						},
																					},
																					"network": {
																						Schema: schema.Schema{
																							Type:         schema.TypeString,
																							ValidateFunc: validateCidrOrIPOrRange(),
																							Optional:     true,
																						},
																						Metadata: metadata.Metadata{
																							SchemaType:   "string",
																							SdkFieldName: "Network",
																						},
																					},
																				},
																			},
																			Optional: true,
																		},
																		Metadata: metadata.Metadata{
																			SchemaType:   "list",
																			SdkFieldName: "StaticRoutes",
																			ReflectType:  reflect.TypeOf(model.ClasslessStaticRoute{}),
																		},
																	},
																},
															},
															Optional: true,
														},
														Metadata: metadata.Metadata{
															SchemaType:   "struct",
															SdkFieldName: "Option121",
															ReflectType:  reflect.TypeOf(model.DhcpOption121{}),
														},
													},
													"other": {
														Schema: schema.Schema{
															Type: schema.TypeList,
															Elem: &metadata.ExtendedResource{
																Schema: map[string]*metadata.ExtendedSchema{
																	"code": {
																		Schema: schema.Schema{
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		Metadata: metadata.Metadata{
																			SchemaType:   "int",
																			SdkFieldName: "Code",
																		},
																	},
																	"values": {
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
																			SdkFieldName: "Values",
																		},
																	},
																},
															},
															Optional: true,
														},
														Metadata: metadata.Metadata{
															SchemaType:   "list",
															SdkFieldName: "Others",
															ReflectType:  reflect.TypeOf(model.GenericDhcpOption{}),
														},
													},
												},
											},
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "struct",
											SdkFieldName: "Options",
											ReflectType:  reflect.TypeOf(model.DhcpV4Options{}),
										},
									},
									"reserved_ip_ranges": {
										Schema: schema.Schema{
											Type: schema.TypeList,
											Elem: &metadata.ExtendedSchema{
												Schema: schema.Schema{
													Type:         schema.TypeString,
													ValidateFunc: validateCidrOrIPOrRange(),
												},
												Metadata: metadata.Metadata{
													SchemaType: "string",
												},
											},
											Optional: true,
										},
										Metadata: metadata.Metadata{
											SchemaType:   "list",
											SdkFieldName: "ReservedIpRanges",
										},
									},
								},
							},
							Optional: true,
							Computed: true,
						},
						Metadata: metadata.Metadata{
							SchemaType:   "struct",
							SdkFieldName: "DhcpServerAdditionalConfig",
							ReflectType:  reflect.TypeOf(model.DhcpServerAdditionalConfig{}),
						},
					},
				},
			},
			Optional: true,
		},
		Metadata: metadata.Metadata{
			SchemaType: "struct",
			// Note that SubnetDhcpConfig is populated here (rather than the alernative VpcSubnetDhcpConfig)
			// This is the recommended way to configure DHCP
			SdkFieldName: "SubnetDhcpConfig",
			ReflectType:  reflect.TypeOf(model.SubnetDhcpConfig{}),
		},
	},
}

func resourceNsxtVpcSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVpcSubnetCreate,
		Read:   resourceNsxtVpcSubnetRead,
		Update: resourceNsxtVpcSubnetUpdate,
		Delete: resourceNsxtVpcSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPCPathResourceImporter,
		},
		Schema: metadata.GetSchemaFromExtendedSchema(vpcSubnetSchema),
	}
}

func resourceNsxtVpcSubnetExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	parents := getVpcParentsFromContext(sessionContext)
	client := clientLayer.NewSubnetsClient(connector)
	_, err = client.Get(parents[0], parents[1], parents[2], id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtVpcSubnetCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id, err := getOrGenerateID2(d, m, resourceNsxtVpcSubnetExists)
	if err != nil {
		return err
	}

	if err = validateDhcpConfig(d); err != nil {
		return err
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	obj := model.VpcSubnet{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcSubnetSchema, "", nil); err != nil {
		return err
	}

	log.Printf("[INFO] Creating VpcSubnet with ID %s", id)

	client := clientLayer.NewSubnetsClient(connector)
	err = client.Patch(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		return handleCreateError("VpcSubnet", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVpcSubnetRead(d, m)
}

func validateDhcpConfig(d *schema.ResourceData) error {
	dhcpConfig, ok := d.GetOk("dhcp_config")
	if !ok {
		return nil
	}

	configList := dhcpConfig.([]interface{})
	if len(configList) == 0 {
		return nil
	}

	config := configList[0].(map[string]interface{})

	if config["mode"].(string) == model.SubnetDhcpConfig_MODE_RELAY {
		if ac, ok := config["dhcp_server_additional_config"]; ok {
			advancedConfig := ac.([]interface{})
			if len(advancedConfig) > 0 {
				return fmt.Errorf("dhcp_server_additional_config can not be specified with DHCP_RELAY mode")
			}
		}
	}

	return nil
}

func resourceNsxtVpcSubnetRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcSubnet ID")
	}

	client := clientLayer.NewSubnetsClient(connector)
	parents := getVpcParentsFromContext(getSessionContext(d, m))
	obj, err := client.Get(parents[0], parents[1], parents[2], id)
	if err != nil {
		return handleReadError(d, "VpcSubnet", id, err)
	}

	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	d.Set("path", obj.Path)
	// Depending on subnet type, this attribute might not be sent back by NSX
	// If not provided by NSX, the next line will explicitly assign empty list to ip_blocks
	d.Set("ip_blocks", obj.IpBlocks)

	elem := reflect.ValueOf(&obj).Elem()
	return metadata.StructToSchema(elem, d, vpcSubnetSchema, "", nil)
}

func resourceNsxtVpcSubnetUpdate(d *schema.ResourceData, m interface{}) error {

	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcSubnet ID")
	}

	if err := validateDhcpConfig(d); err != nil {
		return err
	}

	parents := getVpcParentsFromContext(getSessionContext(d, m))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	revision := int64(d.Get("revision").(int))

	obj := model.VpcSubnet{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Revision:    &revision,
	}

	elem := reflect.ValueOf(&obj).Elem()
	if err := metadata.SchemaToStruct(elem, d, vpcSubnetSchema, "", nil); err != nil {
		return err
	}

	// Since dhcp block is Computed (sent back by NSX even if not specified), we need to
	// explicitly clear out additional DHCP config in case of DHCP RELAY mode, otherwise
	// NSX throws an error
	if (obj.SubnetDhcpConfig != nil) && (obj.SubnetDhcpConfig.Mode != nil && *obj.SubnetDhcpConfig.Mode == model.SubnetDhcpConfig_MODE_RELAY) {
		obj.SubnetDhcpConfig.DhcpServerAdditionalConfig = nil
	}

	client := clientLayer.NewSubnetsClient(connector)
	_, err := client.Update(parents[0], parents[1], parents[2], id, obj)
	if err != nil {
		// Trigger partial update to avoid terraform updating state based on failed intent
		// TODO - move this into handleUpdateError
		d.Partial(true)
		return handleUpdateError("VpcSubnet", id, err)
	}

	return resourceNsxtVpcSubnetRead(d, m)
}

func resourceNsxtVpcSubnetDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VpcSubnet ID")
	}

	connector := getPolicyConnector(m)
	parents := getVpcParentsFromContext(getSessionContext(d, m))

	// Wait until potential VM ports are deleted
	pendingStates := []string{"pending"}
	targetStates := []string{"ok", "error"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			portsClient := subnets.NewPortsClient(connector)
			ports, err := portsClient.List(parents[0], parents[1], parents[2], id, nil, nil, nil, nil, nil, nil)
			if err != nil {
				return ports, "error", logAPIError("Error listing VPC subnet ports", err)
			}
			numOfPorts := len(ports.Results)
			log.Printf("[DEBUG] Current number of ports on subnet %s is %d", id, numOfPorts)

			if numOfPorts > 0 {
				return ports, "pending", nil
			}
			return ports, "ok", nil

		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to get port information for subnet %s: %v", id, err)
	}

	client := clientLayer.NewSubnetsClient(connector)
	err = client.Delete(parents[0], parents[1], parents[2], id)

	if err != nil {
		return handleDeleteError("VpcSubnet", id, err)
	}

	return nil
}
