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
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	mpmodel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/transport_nodes"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	"golang.org/x/exp/maps"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"
)

var ipAssignmentTypes = []string{
	"assigned_by_dhcp",
	"no_ipv4",
	"static_ip",
	"static_ip_pool",
	"static_ip_mac",
}

var ipv6AssignmentTypes = []string{
	"assigned_by_dhcpv6",
	"assigned_by_autoconf",
	"no_ipv6",
	"static_ip",
	"static_ip_pool",
	"static_ip_mac",
}

var mpHostSwitchProfileTypeFromPolicyType = map[string]string{
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYUPLINKHOSTSWITCHPROFILE:          mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_UPLINKHOSTSWITCHPROFILE,
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYLLDPHOSTSWITCHPROFILE:            mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_LLDPHOSTSWITCHPROFILE,
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYNIOCPROFILE:                      mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_NIOCPROFILE,
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYEXTRACONFIGHOSTSWITCHPROFILE:     mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_EXTRACONFIGHOSTSWITCHPROFILE,
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYVTEPHAHOSTSWITCHPROFILE:          mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_VTEPHAHOSTSWITCHPROFILE,
	model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYHIGHPERFORMANCEHOSTSWITCHPROFILE: mpmodel.BaseHostSwitchProfile_RESOURCE_TYPE_HIGHPERFORMANCEHOSTSWITCHPROFILE,
}

const nodeTypeEdge = "EdgeNode"
const nodeTypeHost = "HostNode"

var hostSwitchModeValues = []string{
	mpmodel.StandardHostSwitch_HOST_SWITCH_MODE_STANDARD,
	mpmodel.StandardHostSwitch_HOST_SWITCH_MODE_ENS,
	mpmodel.StandardHostSwitch_HOST_SWITCH_MODE_ENS_INTERRUPT,
	mpmodel.StandardHostSwitch_HOST_SWITCH_MODE_LEGACY,
}

var edgeNodeFormFactorValues = []string{
	mpmodel.EdgeNodeDeploymentConfig_FORM_FACTOR_SMALL,
	mpmodel.EdgeNodeDeploymentConfig_FORM_FACTOR_MEDIUM,
	mpmodel.EdgeNodeDeploymentConfig_FORM_FACTOR_LARGE,
	mpmodel.EdgeNodeDeploymentConfig_FORM_FACTOR_XLARGE,
}

var cpuReservationValues = []string{
	mpmodel.CPUReservation_RESERVATION_IN_SHARES_EXTRA_HIGH_PRIORITY,
	mpmodel.CPUReservation_RESERVATION_IN_SHARES_HIGH_PRIORITY,
	mpmodel.CPUReservation_RESERVATION_IN_SHARES_NORMAL_PRIORITY,
	mpmodel.CPUReservation_RESERVATION_IN_SHARES_LOW_PRIORITY,
}

var syslogLogLevelValues = []string{
	mpmodel.SyslogConfiguration_LOG_LEVEL_EMERGENCY,
	mpmodel.SyslogConfiguration_LOG_LEVEL_ALERT,
	mpmodel.SyslogConfiguration_LOG_LEVEL_CRITICAL,
	mpmodel.SyslogConfiguration_LOG_LEVEL_ERROR,
	mpmodel.SyslogConfiguration_LOG_LEVEL_WARNING,
	mpmodel.SyslogConfiguration_LOG_LEVEL_NOTICE,
	mpmodel.SyslogConfiguration_LOG_LEVEL_INFO,
	mpmodel.SyslogConfiguration_LOG_LEVEL_DEBUG,
}

var syslogProtocolValues = []string{
	mpmodel.SyslogConfiguration_PROTOCOL_TCP,
	mpmodel.SyslogConfiguration_PROTOCOL_UDP,
	mpmodel.SyslogConfiguration_PROTOCOL_TLS,
	mpmodel.SyslogConfiguration_PROTOCOL_LI,
	mpmodel.SyslogConfiguration_PROTOCOL_LI_TLS,
}

func resourceNsxtEdgeTransportNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtEdgeTransportNodeCreate,
		Read:   resourceNsxtEdgeTransportNodeRead,
		Update: resourceNsxtEdgeTransportNodeUpdate,
		Delete: resourceNsxtEdgeTransportNodeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"revision":     getRevisionSchema(),
			"description":  getDescriptionSchema(),
			"display_name": getDisplayNameSchema(),
			"tag":          getTagsSchema(),
			"node_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Unique Id of the fabric node",
				ConflictsWith: []string{"ip_addresses", "fqdn", "deployment_config", "external_id"},
			},
			"failure_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Id of the failure domain",
			},
			// host_switch_spec
			"standard_host_switch": getStandardHostSwitchSchema(nodeTypeEdge),

			// node_deployment_info
			"deployment_config": getEdgeNodeDeploymentConfigSchema(),
			"node_settings":     getEdgeNodeSettingsSchema(),
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "ID of the Node",
			},
			"fqdn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Fully qualified domain name of the fabric node",
			},
			"ip_addresses": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "IP Addresses of the Node, version 4 or 6",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getEdgeNodeDeploymentConfigSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "Config for automatic deployment of edge node virtual machine",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"form_factor": {
					Type:         schema.TypeString,
					Default:      mpmodel.EdgeNodeDeploymentConfig_FORM_FACTOR_MEDIUM,
					Optional:     true,
					ValidateFunc: validation.StringInSlice(edgeNodeFormFactorValues, false),
				},
				"node_user_settings": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Required:    true,
					Description: "Node user settings",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"audit_password": {
								Type:        schema.TypeString,
								Optional:    true,
								Sensitive:   true,
								Description: "Node audit user password",
							},
							"audit_username": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "CLI \"audit\" username",
							},
							"cli_password": {
								Type:        schema.TypeString,
								Required:    true,
								Sensitive:   true,
								Description: "Node cli password",
							},
							"cli_username": {
								Type:        schema.TypeString,
								Optional:    true,
								Default:     "admin",
								Description: "CLI \"admin\" username",
							},
							"root_password": {
								Type:        schema.TypeString,
								Required:    true,
								Sensitive:   true,
								Description: "Node root user password",
							},
						},
					},
				},
				"vm_deployment_config": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Required:    true,
					Description: "The vSphere deployment configuration determines where to deploy the edge node",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"compute_folder_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Compute folder identifier in the specified vcenter server",
							},
							"compute_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Cluster identifier or resourcepool identifier for specified vcenter server",
							},
							"data_network_ids": {
								Type:        schema.TypeList,
								MinItems:    1,
								MaxItems:    4,
								Required:    true,
								Description: "List of portgroups, logical switch identifiers or segment paths for datapath connectivity",
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							"default_gateway_address": {
								Type:        schema.TypeList,
								MaxItems:    2,
								Optional:    true,
								Computed:    true,
								Description: "Default gateway for the node",
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateSingleIP(),
								},
							},
							"host_id": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Host identifier in the specified vcenter server",
							},
							"management_network_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Portgroup, logical switch identifier or segment path for management network connectivity",
							},
							"management_port_subnet": {
								Type:        schema.TypeList,
								Optional:    true,
								Description: "Port subnets for management port. IPv4, IPv6 and Dual Stack Address is supported",
								MinItems:    1,
								MaxItems:    2,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"ip_addresses": {
											Type:        schema.TypeList,
											Required:    true,
											Description: "IP Addresses",
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validateSingleIP(),
											},
										},
										"prefix_length": {
											Type:         schema.TypeInt,
											Required:     true,
											Description:  "Subnet Prefix Length",
											ValidateFunc: validation.IntBetween(1, 128),
										},
									},
								},
							},
							"reservation_info": {
								Type:        schema.TypeList,
								MaxItems:    1,
								Description: "Resource reservation settings",
								Optional:    true,
								Computed:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cpu_reservation_in_mhz": {
											Type:        schema.TypeInt,
											Description: "CPU reservation in MHz",
											Optional:    true,
										},
										"cpu_reservation_in_shares": {
											Type:         schema.TypeString,
											Description:  "CPU reservation in shares",
											Optional:     true,
											Default:      mpmodel.CPUReservation_RESERVATION_IN_SHARES_HIGH_PRIORITY,
											ValidateFunc: validation.StringInSlice(cpuReservationValues, false),
										},
										"memory_reservation_percentage": {
											Type:         schema.TypeInt,
											Optional:     true,
											Description:  "Memory reservation percentage",
											ValidateFunc: validation.IntBetween(0, 100),
											Default:      100,
										},
									},
								},
							},
							"storage_id": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Storage/datastore identifier in the specified vcenter server",
							},
							"vc_id": {
								Type:        schema.TypeString,
								Description: "Vsphere compute identifier for identifying the vcenter server",
								Required:    true,
							},
						},
					},
				},
			},
		},
	}
}

func getEdgeNodeSettingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Description: "Current configuration on edge node",
		Required:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"advanced_configuration": getKeyValuePairListSchema(),
				"allow_ssh_root_login": {
					Type:        schema.TypeBool,
					Default:     false,
					Description: "Allow root SSH logins",
					Optional:    true,
				},
				"dns_servers": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "DNS servers",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateSingleIP(),
					},
				},
				"enable_ssh": {
					Type:        schema.TypeBool,
					Default:     false,
					Description: "Enable SSH",
					Optional:    true,
				},
				"enable_upt_mode": {
					Type:        schema.TypeBool,
					Default:     false,
					Description: "Enable Uniform Passthrough mode",
					Optional:    true,
				},
				"hostname": {
					Type:        schema.TypeString,
					Description: "Host name or FQDN for edge node",
					Required:    true,
				},
				"ntp_servers": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "NTP servers",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"search_domains": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Search domain names",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"syslog_server": {
					Type:        schema.TypeList,
					MaxItems:    5,
					Optional:    true,
					Description: "Syslog servers",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"log_level": {
								Type:         schema.TypeString,
								Optional:     true,
								Description:  "Log level to be redirected",
								Default:      mpmodel.SyslogConfiguration_LOG_LEVEL_INFO,
								ValidateFunc: validation.StringInSlice(syslogLogLevelValues, false),
							},
							"port": {
								Type:     schema.TypeString,
								Optional: true,
								Default:  "514",
								// Spec defines port-or-range format here, however range
								// is not accepted by NSX
								ValidateFunc: validateSinglePort(),
								Description:  "Syslog server port",
							},
							"protocol": {
								Type:         schema.TypeString,
								Optional:     true,
								Description:  "Syslog protocol",
								Default:      mpmodel.SyslogConfiguration_PROTOCOL_UDP,
								ValidateFunc: validation.StringInSlice(syslogProtocolValues, false),
							},
							"server": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Server IP or fqdn",
							},
						},
					},
				},
			},
		},
	}
}

func getStandardHostSwitchSchema(nodeType string) *schema.Schema {
	s := schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		Description: "Standard host switch specification",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host_switch_id": {
					Type:        schema.TypeString,
					Description: "The host switch id. This ID will be used to reference a host switch",
					Optional:    true,
					Computed:    true,
				},
				"host_switch_name": {
					Type:        schema.TypeString,
					Description: "Host switch name. This name will be used to reference a host switch",
					Optional:    true,
					Computed:    true,
				},
				"host_switch_profile": getHostSwitchProfileIDsSchema(),
				"ip_assignment":       getIPAssignmentSchema(false),
				"ipv6_assignment":     getIPv6AssignmentSchema(),
				"pnic": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Physical NICs connected to the host switch",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"device_name": {
								Type:        schema.TypeString,
								Description: "Device name or key",
								Required:    true,
							},
							"uplink_name": {
								Type:        schema.TypeString,
								Description: "Uplink name for this Pnic",
								Required:    true,
							},
						},
					},
				},
				"transport_zone_endpoint": getTransportZoneEndpointSchema(),
			},
		},
	}
	if nodeType == nodeTypeHost {
		elemSchema := s.Elem.(*schema.Resource).Schema
		maps.Copy(elemSchema, map[string]*schema.Schema{
			"cpu_config": {
				Type:        schema.TypeList,
				Description: "Enhanced Networking Stack enabled HostSwitch CPU configuration",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"num_lcores": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of Logical cpu cores (Lcores) to be placed on a specified NUMA node",
						},
						"numa_node_index": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Unique index of the Non Uniform Memory Access (NUMA) node",
						},
					},
				},
			},
			"host_switch_mode": {
				Type:         schema.TypeString,
				Description:  "Operational mode of a HostSwitch",
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(hostSwitchModeValues, false),
			},
			"is_migrate_pnics": {
				Type:        schema.TypeBool,
				Description: "Migrate any pnics which are in use",
				Optional:    true,
			},
			"transport_node_profile_sub_config": {
				Type:        schema.TypeList,
				MaxItems:    16,
				Optional:    true,
				Description: "Transport Node Profile sub-configuration Options",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_switch_config_option": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Required:    true,
							Description: "Subset of the host switch configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_switch_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The host switch id. This ID will be used to reference a host switch",
									},
									"host_switch_profile": getHostSwitchProfileIDsSchema(),
									"ip_assignment":       getIPAssignmentSchema(false),
									"ipv6_assignment":     getIPv6AssignmentSchema(),
									"uplink":              getUplinksSchema(),
								},
							},
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the transport node profile config option",
						},
					},
				},
			},
			"uplink": getUplinksSchema(),
			"vmk_install_migration": {
				Type:        schema.TypeList,
				Description: "The vmknic and logical switch mappings",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_network": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The network id to which the ESX vmk interface will be migrated",
						},
						"device_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ESX vmk interface name",
						},
					},
				},
			},
		})
	}
	return &s
}

func getTransportZoneEndpointSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Transport zone endpoints",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"transport_zone": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Unique ID identifying the transport zone for this endpoint",
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				"transport_zone_profiles": {
					Type:        schema.TypeList,
					Optional:    true,
					Computed:    true,
					Description: "Identifiers of the transport zone profiles associated with this transport zone endpoint on this transport node",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}

func getUplinksSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Uplink/LAG of VMware vSphere Distributed Switch connected to the HostSwitch",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uplink_name": {
					Type:         schema.TypeString,
					Required:     true,
					Description:  "Uplink name from UplinkHostSwitch profile",
					ValidateFunc: validation.StringIsNotWhiteSpace,
				},
				"vds_lag_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Link Aggregation Group (LAG) name of Virtual Distributed Switch",
				},
				"vds_uplink_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Uplink name of VMware vSphere Distributed Switch (VDS)",
				},
			},
		},
	}
}

func getHostSwitchProfileIDsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Identifiers of host switch profiles to be associated with this host switch",
		Optional:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

func getIpMacPairSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of IP MAC pairs",
		MinItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ip_address": {
					Type:         schema.TypeString,
					Description:  "A single IPv6 address",
					Required:     true,
					ValidateFunc: validation.IsIPv6Address,
				},
				"mac_address": {
					Type:        schema.TypeString,
					Description: "A single MAC address",
					Required:    true,
				},
			},
		},
		Required: true,
	}
}

func getIPAssignmentSchema(required bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Specification for IPs to be used with host switch virtual tunnel endpoints",
		MaxItems:    1,
		Required:    required,
		Optional:    !required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"assigned_by_dhcp": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Enables DHCP assignment",
				},
				"static_ip": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "IP assignment specification for Static IP List.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_gateway": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Gateway IP",
								ValidateFunc: validateSingleIP(),
							},
							"ip_addresses": {
								Type:        schema.TypeList,
								Description: "List of IPs for transport node host switch virtual tunnel endpoints",
								MinItems:    1,
								Required:    true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateSingleIP(),
								},
							},
							"subnet_mask": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Subnet mask",
								ValidateFunc: validateSingleIP(),
							},
						},
					},
				},
				"static_ip_pool": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "IP assignment specification for Static IP Pool",
				},
				"static_ip_mac": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "IP assignment specification for Static MAC List.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_gateway": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Gateway IP",
								ValidateFunc: validateSingleIP(),
							},
							"ip_mac_pair": getIpMacPairSchema(),
							"subnet_mask": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Subnet mask",
							},
						},
					},
				},
				"no_ipv4": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "No IPv4 assignment",
				},
			},
		},
	}
}

func getIPv6AssignmentSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Specification for IPv6 to be used with host switch virtual tunnel endpoints",
		MaxItems:    1,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"assigned_by_dhcpv6": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Enables DHCPv6 assignment",
				},
				"assigned_by_autoconf": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "Enables autoconf assignment",
				},
				"no_ipv6": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
					Description: "No IPv6 assignment",
				},
				"static_ip": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "IP assignment specification for Static IP List.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_gateway": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Gateway IP",
								ValidateFunc: validateSingleIP(),
							},
							"ip_addresses": {
								Type:        schema.TypeList,
								Description: "List of IPs for transport node host switch virtual tunnel endpoints",
								MinItems:    1,
								Required:    true,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validateSingleIP(),
								},
							},
							"prefix_length": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Prefix length",
							},
						},
					},
				},
				"static_ip_mac": {
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Description: "IP assignment specification for Static MAC List.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_gateway": {
								Type:         schema.TypeString,
								Required:     true,
								Description:  "Gateway IP",
								ValidateFunc: validateSingleIP(),
							},
							"ip_mac_pair": getIpMacPairSchema(),
							"prefix_length": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Prefix length",
							},
						},
					},
				},
				"static_ip_pool": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "IP assignment specification for Static IP Pool",
				},
			},
		},
	}
}

func getTransportNodeFromSchema(d *schema.ResourceData, m interface{}) (*mpmodel.TransportNode, error) {
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getMPTagsFromSchema(d)
	nodeID := d.Get("node_id").(string)
	failureDomain := d.Get("failure_domain").(string)
	hostSwitchSpec, err := getHostSwitchSpecFromSchema(d, m, nodeTypeEdge)
	if err != nil {
		return nil, fmt.Errorf("failed to create Transport Node: %v", err)
	}

	var nodeDeploymentInfo *data.StructValue
	if nodeID == "" {
		/*
			node_id attribute conflicts with node_deployment_info. As node_deployment_info is a complex object which has
			attributes with default values, this schema property will always have values - therefore there is no simple way
			to enforce this conflict within the provider (e.g check if node_id and node_deployment_info have values, then
			fail.
			So the provider will ignore node_deployment_info properties when node_id has a value - which would mean that
			this edge appliance was created externally.
		*/
		log.Printf("node_id not specified, will deploy edge using values in deploymentConfig")
		converter := bindings.NewTypeConverter()
		var dataValue data.DataValue
		var errs []error

		externalID := d.Get("external_id").(string)
		fqdn := d.Get("fqdn").(string)
		ipAddresses := interfaceListToStringList(d.Get("ip_addresses").([]interface{}))

		deploymentConfig, err := getEdgeNodeDeploymentConfigFromSchema(d.Get("deployment_config"))
		if err != nil {
			return nil, err
		}
		nodeSettings, err := getEdgeNodeSettingsFromSchema(d.Get("node_settings"))
		if err != nil {
			return nil, err
		}
		node := mpmodel.EdgeNode{
			ExternalId:       &externalID,
			Fqdn:             &fqdn,
			IpAddresses:      ipAddresses,
			DeploymentConfig: deploymentConfig,
			NodeSettings:     nodeSettings,
			ResourceType:     mpmodel.EdgeNode__TYPE_IDENTIFIER,
		}
		dataValue, errs = converter.ConvertToVapi(node, mpmodel.EdgeNodeBindingType())

		if errs != nil {
			log.Printf("Failed to convert node object, errors are %v", errs)
			return nil, errs[0]
		}
		nodeDeploymentInfo = dataValue.(*data.StructValue)
	}

	obj := mpmodel.TransportNode{
		Description:        &description,
		DisplayName:        &displayName,
		Tags:               tags,
		FailureDomainId:    &failureDomain,
		HostSwitchSpec:     hostSwitchSpec,
		NodeDeploymentInfo: nodeDeploymentInfo,
	}

	return &obj, nil
}

func resourceNsxtEdgeTransportNodeCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewTransportNodesClient(connector)

	nodeID := d.Get("node_id").(string)
	if nodeID != "" {
		obj, err := client.Get(nodeID)
		if err != nil {
			return handleCreateError("TransportNode", nodeID, err)
		}
		// Set node_id, revision and computed values in schema
		d.Set("failure_domain", obj.FailureDomainId)

		converter := bindings.NewTypeConverter()
		base, errs := converter.ConvertToGolang(obj.NodeDeploymentInfo, mpmodel.EdgeNodeBindingType())
		if errs != nil {
			return handleCreateError("TransportNode", nodeID, errs[0])
		}
		node := base.(mpmodel.EdgeNode)
		d.Set("external_id", node.ExternalId)
		d.Set("fqdn", node.Fqdn)
		d.Set("ip_addresses", node.IpAddresses)
		if obj.Revision != nil {
			d.Set("revision", obj.Revision)
		}

		d.SetId(nodeID)
		err = resourceNsxtEdgeTransportNodeUpdate(d, m)
		if err != nil {
			// There is a failure in update, let's discard this so state will remain clean
			d.SetId("")
		}
		return err
	}

	obj, err := getTransportNodeFromSchema(d, m)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating Transport Node with name %s", *obj.DisplayName)

	obj1, err := client.Create(*obj)
	if err != nil {
		return handleCreateError("TransportNode", *obj.DisplayName, err)
	}

	d.SetId(*obj1.Id)
	return resourceNsxtEdgeTransportNodeRead(d, m)
}

func getEdgeNodeDeploymentConfigFromSchema(cfg interface{}) (*mpmodel.EdgeNodeDeploymentConfig, error) {
	converter := bindings.NewTypeConverter()

	if cfg == nil {
		return nil, nil
	}
	for _, ci := range cfg.([]interface{}) {
		c := ci.(map[string]interface{})
		formFactor := c["form_factor"].(string)
		var nodeUserSettings *mpmodel.NodeUserSettings
		if c["node_user_settings"] != nil {
			for _, nusi := range c["node_user_settings"].([]interface{}) {
				nus := nusi.(map[string]interface{})
				auditPassword := nus["audit_password"].(string)
				auditUsername := nus["audit_username"].(string)
				cliPassword := nus["cli_password"].(string)
				cliUsername := nus["cli_username"].(string)
				rootPassword := nus["root_password"].(string)

				nodeUserSettings = &mpmodel.NodeUserSettings{
					CliPassword:  &cliPassword,
					CliUsername:  &cliUsername,
					RootPassword: &rootPassword,
				}
				if auditUsername != "" {
					nodeUserSettings.AuditUsername = &auditUsername
				}
				if auditPassword != "" {
					nodeUserSettings.AuditPassword = &auditPassword
				}
			}
		}
		var vmDeploymentConfig *data.StructValue
		for _, vdci := range c["vm_deployment_config"].([]interface{}) {
			vdc := vdci.(map[string]interface{})
			computeFolderID := vdc["compute_folder_id"].(string)
			computeID := vdc["compute_id"].(string)
			dataNetworkIds := interface2StringList(vdc["data_network_ids"].([]interface{}))
			defaultGatewayAddresses := interface2StringList(vdc["default_gateway_address"].([]interface{}))
			hostID := vdc["host_id"].(string)
			managementNetworkID := vdc["management_network_id"].(string)
			var managemenPortSubnets []mpmodel.IPSubnet
			for _, ipsi := range vdc["management_port_subnet"].([]interface{}) {
				ips := ipsi.(map[string]interface{})
				ipAddresses := interface2StringList(ips["ip_addresses"].([]interface{}))
				prefixLength := int64(ips["prefix_length"].(int))
				subnet := mpmodel.IPSubnet{
					IpAddresses:  ipAddresses,
					PrefixLength: &prefixLength,
				}
				managemenPortSubnets = append(managemenPortSubnets, subnet)
			}
			var reservationInfo *mpmodel.ReservationInfo
			for _, ri := range vdc["reservation_info"].([]interface{}) {
				rInfo := ri.(map[string]interface{})
				cpuReservationInMhz := int64(rInfo["cpu_reservation_in_mhz"].(int))
				cpuReservationInShares := rInfo["cpu_reservation_in_shares"].(string)
				memoryReservationPercentage := int64(rInfo["memory_reservation_percentage"].(int))

				reservationInfo = &mpmodel.ReservationInfo{
					CpuReservation: &mpmodel.CPUReservation{
						ReservationInMhz:    &cpuReservationInMhz,
						ReservationInShares: &cpuReservationInShares,
					},
					MemoryReservation: &mpmodel.MemoryReservation{
						ReservationPercentage: &memoryReservationPercentage,
					},
				}
			}
			storageID := vdc["storage_id"].(string)
			vcID := vdc["vc_id"].(string)
			cfg := mpmodel.VsphereDeploymentConfig{
				ComputeId:             &computeID,
				DataNetworkIds:        dataNetworkIds,
				ManagementNetworkId:   &managementNetworkID,
				ManagementPortSubnets: managemenPortSubnets,
				ReservationInfo:       reservationInfo,
				StorageId:             &storageID,
				VcId:                  &vcID,
				PlacementType:         mpmodel.DeploymentConfig_PLACEMENT_TYPE_VSPHEREDEPLOYMENTCONFIG,
			}
			if len(defaultGatewayAddresses) > 0 {
				cfg.DefaultGatewayAddresses = defaultGatewayAddresses
			}
			if hostID != "" {
				cfg.HostId = &hostID
			}
			// Passing an empty folder here confuses vSphere while creating the Edge VM
			if computeFolderID != "" {
				cfg.ComputeFolderId = &computeFolderID
			}
			dataValue, errs := converter.ConvertToVapi(cfg, mpmodel.VsphereDeploymentConfigBindingType())
			if errs != nil {
				return nil, errs[0]
			} else if dataValue != nil {
				vmDeploymentConfig = dataValue.(*data.StructValue)
			}
		}
		return &mpmodel.EdgeNodeDeploymentConfig{
			FormFactor:         &formFactor,
			NodeUserSettings:   nodeUserSettings,
			VmDeploymentConfig: vmDeploymentConfig,
		}, nil
	}
	return nil, nil
}

func getEdgeNodeSettingsFromSchema(s interface{}) (*mpmodel.EdgeNodeSettings, error) {
	if s == nil {
		return nil, nil
	}
	settings := s.([]interface{})
	for _, settingIf := range settings {
		setting := settingIf.(map[string]interface{})
		advCfg := getKeyValuePairListFromSchema(setting["advanced_configuration"])
		allowSSHRootLogin := setting["allow_ssh_root_login"].(bool)
		dnsServers := interface2StringList(setting["dns_servers"].([]interface{}))
		enableSSH := setting["enable_ssh"].(bool)
		enableUptMode := setting["enable_upt_mode"].(bool)
		hostName := setting["hostname"].(string)
		ntpServers := interface2StringList(setting["ntp_servers"].([]interface{}))
		searchDomains := interface2StringList(setting["search_domains"].([]interface{}))
		var syslogServers []mpmodel.SyslogConfiguration
		for _, sli := range setting["syslog_server"].([]interface{}) {
			syslogServer := sli.(map[string]interface{})
			logLevel := syslogServer["log_level"].(string)
			port := syslogServer["port"].(string)
			protocol := syslogServer["protocol"].(string)
			server := syslogServer["server"].(string)
			syslogServers = append(syslogServers, mpmodel.SyslogConfiguration{
				LogLevel: &logLevel,
				Port:     &port,
				Protocol: &protocol,
				Server:   &server,
			})
		}
		obj := &mpmodel.EdgeNodeSettings{
			AdvancedConfiguration: advCfg,
			AllowSshRootLogin:     &allowSSHRootLogin,
			DnsServers:            dnsServers,
			EnableSsh:             &enableSSH,
			Hostname:              &hostName,
			NtpServers:            ntpServers,
			SearchDomains:         searchDomains,
			SyslogServers:         syslogServers,
		}
		if util.NsxVersionHigherOrEqual("4.0.0") {
			obj.EnableUptMode = &enableUptMode
		}
		return obj, nil
	}

	return nil, nil
}

func getCPUConfigFromSchema(cpuConfigList []interface{}) []mpmodel.CpuCoreConfigForEnhancedNetworkingStackSwitch {
	var cpuConfig []mpmodel.CpuCoreConfigForEnhancedNetworkingStackSwitch
	for _, cc := range cpuConfigList {
		data := cc.(map[string]interface{})
		numLCores := int64(data["num_lcores"].(int))
		numaNodeIndex := int64(data["numa_node_index"].(int))
		elem := mpmodel.CpuCoreConfigForEnhancedNetworkingStackSwitch{
			NumLcores:     &numLCores,
			NumaNodeIndex: &numaNodeIndex,
		}
		cpuConfig = append(cpuConfig, elem)
	}
	return cpuConfig
}

func getHostSwitchProfileResourceType(m interface{}, id string) (string, error) {
	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	t := true
	// we retrieve a list of profiles instead of using Get(), as the id could be either MP id or Policy id
	list, err := client.List(nil, nil, nil, nil, &t, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return "", err
	}
	converter := bindings.NewTypeConverter()

	for _, structValue := range list.Results {
		baseInterface, errs := converter.ConvertToGolang(structValue, model.PolicyBaseHostSwitchProfileBindingType())
		if errs != nil {
			return "", errs[0]
		}
		base := baseInterface.(model.PolicyBaseHostSwitchProfile)

		if *base.Id == id || *base.RealizationId == id {
			resourceType, ok := mpHostSwitchProfileTypeFromPolicyType[base.ResourceType]
			if !ok {
				return "", fmt.Errorf("MP resource type not found for %s", base.ResourceType)
			}
			return resourceType, nil
		}
	}

	return "", fmt.Errorf("Host Switch Profile type not found for %s", id)
}

func getHostSwitchProfileIDsFromSchema(m interface{}, hswProfileList []interface{}) ([]mpmodel.HostSwitchProfileTypeIdEntry, error) {
	var hswProfiles []mpmodel.HostSwitchProfileTypeIdEntry
	for _, hswp := range hswProfileList {
		val := hswp.(string)
		key, err := getHostSwitchProfileResourceType(m, getPolicyIDFromPath(val))
		if err != nil {
			return nil, err
		}
		elem := mpmodel.HostSwitchProfileTypeIdEntry{
			Key:   &key,
			Value: &val,
		}
		hswProfiles = append(hswProfiles, elem)
	}
	return hswProfiles, nil
}

func getIPAssignmentFromSchema(ipAssignmentList interface{}) (*data.StructValue, error) {
	if ipAssignmentList == nil {
		return nil, nil
	}
	converter := bindings.NewTypeConverter()

	for _, ia := range ipAssignmentList.([]interface{}) {
		iaType, iaData, err := getIPAssignmentData(ia.(map[string]interface{}), ipAssignmentTypes)
		if err != nil {
			return nil, err
		}

		var dataValue data.DataValue
		var errs []error
		switch iaType {
		case "assigned_by_dhcp":
			dhcpEnabled := iaData.(bool)
			if dhcpEnabled {
				elem := mpmodel.AssignedByDhcp{
					ResourceType: mpmodel.IpAssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYDHCP,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.AssignedByDhcpBindingType())
			} else {
				return nil, fmt.Errorf("no valid IP assignment found")
			}
		case "no_ipv4":
			nope := iaData.(bool)
			if nope {
				elem := mpmodel.NoIpv4{
					ResourceType: mpmodel.IpAssignmentSpec_RESOURCE_TYPE_NOIPV4,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.NoIpv4BindingType())
			} else {
				return nil, fmt.Errorf("no valid IP assignment found")
			}
		case "static_ip":
			for _, iad := range iaData.([]interface{}) {
				data := iad.(map[string]interface{})
				defaultGateway := data["default_gateway"].(string)
				ipList := interfaceListToStringList(data["ip_addresses"].([]interface{}))
				subnetMask := data["subnet_mask"].(string)
				elem := mpmodel.StaticIpListSpec{
					DefaultGateway: &defaultGateway,
					IpList:         ipList,
					SubnetMask:     &subnetMask,
					ResourceType:   mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPLISTSPEC,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpListSpecBindingType())
				break
			}

		case "static_ip_mac":
			for _, iad := range iaData.([]interface{}) {
				data := iad.(map[string]interface{})
				defaultGateway := data["default_gateway"].(string)
				var ipMacList []mpmodel.IpMacPair
				schemaList := data["ip_mac_pair"].([]interface{})
				for _, schemaEntry := range schemaList {
					ipmac := schemaEntry.(map[string]interface{})
					ip := ipmac["ip_address"].(string)
					mac := ipmac["mac_address"].(string)
					entry := mpmodel.IpMacPair{
						Ip:  &ip,
						Mac: &mac,
					}
					ipMacList = append(ipMacList, entry)
				}
				subnetMask := data["subnet_mask"].(string)
				elem := mpmodel.StaticIpMacListSpec{
					DefaultGateway: &defaultGateway,
					IpMacList:      ipMacList,
					SubnetMask:     &subnetMask,
					ResourceType:   mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPMACLISTSPEC,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpMacListSpecBindingType())
				break
			}

		case "static_ip_pool":
			staticIPPoolID := iaData.(string)
			elem := mpmodel.StaticIpPoolSpec{
				IpPoolId:     &staticIPPoolID,
				ResourceType: mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPPOOLSPEC,
			}
			dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpPoolSpecBindingType())

		default:
			return nil, fmt.Errorf("no valid IP assignment found")
		}
		if errs != nil {
			return nil, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		return entryStruct, nil
	}
	return nil, nil
}

func getIPv6AssignmentFromSchema(ipAssignmentList interface{}) (*data.StructValue, error) {
	if ipAssignmentList == nil {
		return nil, nil
	}
	converter := bindings.NewTypeConverter()

	for _, ia := range ipAssignmentList.([]interface{}) {
		iaType, iaData, err := getIPAssignmentData(ia.(map[string]interface{}), ipv6AssignmentTypes)
		if err != nil {
			return nil, err
		}

		var dataValue data.DataValue
		var errs []error
		switch iaType {
		case "assigned_by_dhcpv6":
			dhcpEnabled := iaData.(bool)
			if dhcpEnabled {
				elem := mpmodel.AssignedByDhcpv6{
					ResourceType: mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYDHCPV6,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.AssignedByDhcpv6BindingType())
			} else {
				return nil, fmt.Errorf("no valid IP assignment found")
			}
		case "assigned_by_autoconf":
			autoConf := iaData.(bool)
			if autoConf {
				elem := mpmodel.AssignedByAutoConf{
					ResourceType: mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYAUTOCONF,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.AssignedByAutoConfBindingType())
			} else {
				return nil, fmt.Errorf("no valid IP assignment found")
			}
		case "no_ipv6":
			nope := iaData.(bool)
			if nope {
				elem := mpmodel.NoIpv6{
					ResourceType: mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_NOIPV6,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.NoIpv6BindingType())
			} else {
				return nil, fmt.Errorf("no valid IP assignment found")
			}
		case "static_ip":
			for _, iad := range iaData.([]interface{}) {
				data := iad.(map[string]interface{})
				defaultGateway := data["default_gateway"].(string)
				ipList := interfaceListToStringList(data["ip_addresses"].([]interface{}))
				prefixLength := data["prefix_length"].(string)
				elem := mpmodel.StaticIpv6ListSpec{
					DefaultGateway: &defaultGateway,
					Ipv6List:       ipList,
					PrefixLength:   &prefixLength,
					ResourceType:   mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6LISTSPEC,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpv6ListSpecBindingType())
				break
			}

		case "static_ip_mac":
			for _, iad := range iaData.([]interface{}) {
				data := iad.(map[string]interface{})
				defaultGateway := data["default_gateway"].(string)
				var ipMacList []mpmodel.Ipv6MacPair
				schemaList := data["ip_mac_pair"].([]interface{})
				for _, schemaEntry := range schemaList {
					ipmac := schemaEntry.(map[string]interface{})
					ipv6 := ipmac["ip_address"].(string)
					mac := ipmac["mac_address"].(string)
					entry := mpmodel.Ipv6MacPair{
						Ipv6: &ipv6,
						Mac:  &mac,
					}
					ipMacList = append(ipMacList, entry)
				}
				prefixLength := data["prefix_length"].(string)
				elem := mpmodel.StaticIpv6MacListSpec{
					DefaultGateway: &defaultGateway,
					Ipv6MacList:    ipMacList,
					PrefixLength:   &prefixLength,
					ResourceType:   mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6MACLISTSPEC,
				}
				dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpv6MacListSpecBindingType())
				break
			}

		case "static_ip_pool":
			staticIPPoolID := iaData.(string)
			elem := mpmodel.StaticIpv6PoolSpec{
				Ipv6PoolId:   &staticIPPoolID,
				ResourceType: mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6POOLSPEC,
			}
			dataValue, errs = converter.ConvertToVapi(elem, mpmodel.StaticIpv6PoolSpecBindingType())

		default:
			return nil, fmt.Errorf("no valid IP assignment found")
		}
		if errs != nil {
			return nil, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		return entryStruct, nil
	}
	return nil, nil
}

func isSlice(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func isBool(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Bool
}

func isString(v interface{}) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.String
}

func getIPAssignmentData(data map[string]interface{}, valueList []string) (string, interface{}, error) {
	var t string
	var d interface{}

	n := 0
	for _, iaType := range valueList {
		d1 := data[iaType]
		if (isString(d1) && d1 != "") || (isSlice(d1) && len(d1.([]interface{})) > 0) || (isBool(d1) && d1.(bool)) {
			t, d = iaType, data[iaType]
			n++
		}
	}
	if n > 1 {
		return "", nil, fmt.Errorf("exactly one IP assignment is allowed for ip_assignment object")
	}
	return t, d, nil
}

func getHostSwitchSpecFromSchema(d *schema.ResourceData, m interface{}, nodeType string) (*data.StructValue, error) {
	var dataValue data.DataValue
	var errs []error
	var hsList []mpmodel.StandardHostSwitch

	converter := bindings.NewTypeConverter()
	standardSwitchList := d.Get("standard_host_switch").([]interface{})
	for _, swEntry := range standardSwitchList {
		swData := swEntry.(map[string]interface{})

		hostSwitchID := swData["host_switch_id"].(string)
		hostSwitchName := swData["host_switch_name"].(string)
		var hostSwitchMode, hostSwitchType string
		var isMigratePNics bool
		var uplinks []mpmodel.VdsUplink
		var cpuConfig []mpmodel.CpuCoreConfigForEnhancedNetworkingStackSwitch
		if nodeType == nodeTypeHost {
			cpuConfig = getCPUConfigFromSchema(swData["cpu_config"].([]interface{}))
			hostSwitchMode = swData["host_switch_mode"].(string)
			hostSwitchType = mpmodel.StandardHostSwitch_HOST_SWITCH_TYPE_VDS
			isMigratePNics = swData["is_migrate_pnics"].(bool)
			uplinks = getUplinksFromSchema(swData["uplink"].([]interface{}))
		} else if nodeType == nodeTypeEdge {
			hostSwitchMode = mpmodel.StandardHostSwitch_HOST_SWITCH_MODE_STANDARD
			hostSwitchType = mpmodel.StandardHostSwitch_HOST_SWITCH_TYPE_NVDS
		}
		hostSwitchProfileIDs, err := getHostSwitchProfileIDsFromSchema(m, swData["host_switch_profile"].([]interface{}))
		if err != nil {
			return nil, err
		}
		iPAssignmentSpec, err := getIPAssignmentFromSchema(swData["ip_assignment"])
		if err != nil {
			return nil, fmt.Errorf("error parsing HostSwitchSpec schema %v", err)
		}
		iPv6AssignmentSpec, err := getIPv6AssignmentFromSchema(swData["ipv6_assignment"])
		if err != nil {
			return nil, fmt.Errorf("error parsing HostSwitchSpec schema %v", err)
		}
		var pNics []mpmodel.Pnic
		for _, p := range swData["pnic"].([]interface{}) {
			data := p.(map[string]interface{})
			deviceName := data["device_name"].(string)
			uplinkName := data["uplink_name"].(string)
			elem := mpmodel.Pnic{
				DeviceName: &deviceName,
				UplinkName: &uplinkName,
			}
			pNics = append(pNics, elem)
		}
		var transportNodeSubProfileCfg []mpmodel.TransportNodeProfileSubConfig
		if nodeType == nodeTypeHost {
			transportNodeSubProfileCfg, err = getTransportNodeSubProfileCfg(m, swData["transport_node_profile_sub_config"])
			if err != nil {
				return nil, err
			}
		}
		transportZoneEndpoints := getTransportZoneEndpointsFromSchema(swData["transport_zone_endpoint"].([]interface{}))

		hsw := mpmodel.StandardHostSwitch{
			HostSwitchId:           &hostSwitchID,
			HostSwitchMode:         &hostSwitchMode,
			HostSwitchProfileIds:   hostSwitchProfileIDs,
			HostSwitchType:         &hostSwitchType,
			IpAssignmentSpec:       iPAssignmentSpec,
			Ipv6AssignmentSpec:     iPv6AssignmentSpec,
			Pnics:                  pNics,
			TransportZoneEndpoints: transportZoneEndpoints,
		}
		if hostSwitchName != "" {
			hsw.HostSwitchName = &hostSwitchName
		}
		if nodeType == nodeTypeHost {
			hsw.CpuConfig = cpuConfig
			hsw.IsMigratePnics = &isMigratePNics
			hsw.TransportNodeProfileSubConfigs = transportNodeSubProfileCfg
			hsw.Uplinks = uplinks
		}

		hsList = append(hsList, hsw)
	}
	hostSwitchSpec := mpmodel.StandardHostSwitchSpec{
		HostSwitches: hsList,
		ResourceType: mpmodel.HostSwitchSpec_RESOURCE_TYPE_STANDARDHOSTSWITCHSPEC,
	}
	dataValue, errs = converter.ConvertToVapi(hostSwitchSpec, mpmodel.StandardHostSwitchSpecBindingType())

	if errs != nil {
		return nil, errs[0]
	} else if dataValue != nil {
		return dataValue.(*data.StructValue), nil
	}
	return nil, nil
}

func getTransportZoneEndpointsFromSchema(endpointList []interface{}) []mpmodel.TransportZoneEndPoint {
	var tzEPList []mpmodel.TransportZoneEndPoint
	for _, endpoint := range endpointList {
		data := endpoint.(map[string]interface{})
		transportZoneID := data["transport_zone"].(string)
		var transportZoneProfileIDs []mpmodel.TransportZoneProfileTypeIdEntry
		if data["transport_zone_profiles"] != nil {
			for _, tzpID := range data["transport_zone_profiles"].([]interface{}) {
				profileID := tzpID.(string)
				resourceType := mpmodel.TransportZoneProfileTypeIdEntry_RESOURCE_TYPE_BFDHEALTHMONITORINGPROFILE
				elem := mpmodel.TransportZoneProfileTypeIdEntry{
					ProfileId:    &profileID,
					ResourceType: &resourceType,
				}
				transportZoneProfileIDs = append(transportZoneProfileIDs, elem)
			}
		}

		elem := mpmodel.TransportZoneEndPoint{
			TransportZoneId:         &transportZoneID,
			TransportZoneProfileIds: transportZoneProfileIDs,
		}
		tzEPList = append(tzEPList, elem)
	}
	return tzEPList
}

func getUplinksFromSchema(uplinksList []interface{}) []mpmodel.VdsUplink {
	var uplinks []mpmodel.VdsUplink
	for _, ul := range uplinksList {
		data := ul.(map[string]interface{})
		uplinkName := data["uplink_name"].(string)
		vdsLagName := data["vds_lag_name"].(string)
		vdsUplinkName := data["vds_uplink_name"].(string)
		elem := mpmodel.VdsUplink{
			UplinkName:    &uplinkName,
			VdsLagName:    &vdsLagName,
			VdsUplinkName: &vdsUplinkName,
		}
		uplinks = append(uplinks, elem)
	}
	return uplinks
}

func getTransportNodeSubProfileCfg(m interface{}, iface interface{}) ([]mpmodel.TransportNodeProfileSubConfig, error) {
	var cfgList []mpmodel.TransportNodeProfileSubConfig
	if iface == nil {
		return cfgList, nil
	}
	profileCfgList := iface.([]interface{})
	for _, pCfg := range profileCfgList {
		data := pCfg.(map[string]interface{})
		name := data["name"].(string)
		var swCfgOpt *mpmodel.HostSwitchConfigOption
		for _, cfgOpt := range data["host_switch_config_option"].([]interface{}) {
			opt := cfgOpt.(map[string]interface{})
			swID := opt["host_switch_id"].(string)
			profileIDs, err := getHostSwitchProfileIDsFromSchema(m, opt["host_switch_profile"].([]interface{}))
			if err != nil {
				return nil, err
			}
			iPAssignmentSpec, _ := getIPAssignmentFromSchema(opt["ip_assignment"].([]interface{}))
			uplinks := getUplinksFromSchema(opt["uplink"].([]interface{}))
			swCfgOpt = &mpmodel.HostSwitchConfigOption{
				HostSwitchId:         &swID,
				HostSwitchProfileIds: profileIDs,
				IpAssignmentSpec:     iPAssignmentSpec,
				Uplinks:              uplinks,
			}
		}

		elem := mpmodel.TransportNodeProfileSubConfig{
			Name:                   &name,
			HostSwitchConfigOption: swCfgOpt,
		}
		cfgList = append(cfgList, elem)
	}
	return cfgList, nil
}

func resourceNsxtEdgeTransportNodeRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := nsx.NewTransportNodesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "TransportNode", id, err)
	}

	d.Set("revision", obj.Revision)
	d.Set("description", obj.Description)
	d.Set("display_name", obj.DisplayName)
	setMPTagsInSchema(d, obj.Tags)
	d.Set("node_id", obj.NodeId)
	d.Set("failure_domain", obj.FailureDomainId)

	if obj.HostSwitchSpec != nil {
		err = setHostSwitchSpecInSchema(d, obj.HostSwitchSpec, nodeTypeEdge)
		if err != nil {
			return handleReadError(d, "TransportNode", id, err)
		}
	}

	converter := bindings.NewTypeConverter()
	base, errs := converter.ConvertToGolang(obj.NodeDeploymentInfo, mpmodel.EdgeNodeBindingType())
	if errs != nil {
		return handleReadError(d, "TransportNode", id, errs[0])
	}
	node := base.(mpmodel.EdgeNode)

	if node.DeploymentConfig != nil {
		err = setEdgeDeploymentConfigInSchema(d, node.DeploymentConfig)
		if err != nil {
			return handleReadError(d, "TransportNode", id, err)
		}
	}

	if node.NodeSettings != nil {
		err = setEdgeNodeSettingsInSchema(d, node.NodeSettings)
		if err != nil {
			return handleReadError(d, "TransportNode", id, err)
		}
	}

	// Set base object attributes
	d.Set("external_id", node.ExternalId)
	d.Set("fqdn", node.Fqdn)
	d.Set("ip_addresses", node.IpAddresses)

	if err != nil {
		return handleReadError(d, "TransportNode", id, err)
	}

	return nil
}

func setEdgeNodeSettingsInSchema(d *schema.ResourceData, nodeSettings *mpmodel.EdgeNodeSettings) error {
	elem := getElemOrEmptyMapFromSchema(d, "node_settings")
	elem["advanced_configuration"] = setKeyValueListForSchema(nodeSettings.AdvancedConfiguration)
	elem["allow_ssh_root_login"] = nodeSettings.AllowSshRootLogin
	elem["dns_servers"] = nodeSettings.DnsServers
	elem["enable_ssh"] = nodeSettings.EnableSsh
	elem["enable_upt_mode"] = nodeSettings.EnableUptMode
	elem["hostname"] = nodeSettings.Hostname
	elem["ntp_servers"] = nodeSettings.NtpServers
	elem["search_domains"] = nodeSettings.SearchDomains
	var syslogServers []map[string]interface{}
	for _, syslogServer := range nodeSettings.SyslogServers {
		e := make(map[string]interface{})
		e["log_level"] = syslogServer.LogLevel
		e["port"] = syslogServer.Port
		e["protocol"] = syslogServer.Protocol
		e["server"] = syslogServer.Server
		syslogServers = append(syslogServers, e)
	}
	elem["syslog_server"] = syslogServers
	d.Set("node_settings", []map[string]interface{}{elem})
	return nil
}

func setEdgeDeploymentConfigInSchema(d *schema.ResourceData, deploymentConfig *mpmodel.EdgeNodeDeploymentConfig) error {
	var err error
	elem := getElemOrEmptyMapFromSchema(d, "deployment_config")

	elem["form_factor"] = deploymentConfig.FormFactor
	var nodeUserSettings map[string]interface{}
	if elem["node_user_settings"] != nil {
		nodeUserSettings = elem["node_user_settings"].([]interface{})[0].(map[string]interface{})
	} else {
		nodeUserSettings = make(map[string]interface{})
	}
	// Note: password attributes is sensitive and is not returned by NSX
	if deploymentConfig.NodeUserSettings.AuditUsername != nil {
		nodeUserSettings["audit_username"] = deploymentConfig.NodeUserSettings.AuditUsername
	}
	nodeUserSettings["cli_username"] = deploymentConfig.NodeUserSettings.CliUsername
	elem["node_user_settings"] = []interface{}{nodeUserSettings}

	elem["vm_deployment_config"], err = setVMDeploymentConfigInSchema(deploymentConfig.VmDeploymentConfig)
	if err != nil {
		return err
	}
	d.Set("deployment_config", []map[string]interface{}{elem})

	return nil
}

func setVMDeploymentConfigInSchema(config *data.StructValue) (interface{}, error) {
	converter := bindings.NewTypeConverter()
	base, errs := converter.ConvertToGolang(config, mpmodel.DeploymentConfigBindingType())
	if errs != nil {
		return nil, errs[0]
	}
	cfgType := base.(mpmodel.DeploymentConfig).PlacementType

	// Only VsphereDeploymentConfig is supported
	if cfgType != mpmodel.DeploymentConfig_PLACEMENT_TYPE_VSPHEREDEPLOYMENTCONFIG {
		return nil, fmt.Errorf("unsupported PlacementType %s", cfgType)
	}

	vCfg, errs := converter.ConvertToGolang(config, mpmodel.VsphereDeploymentConfigBindingType())
	if errs != nil {
		return nil, errs[0]
	}
	vSphereCfg := vCfg.(mpmodel.VsphereDeploymentConfig)
	elem := make(map[string]interface{})
	elem["compute_folder_id"] = vSphereCfg.ComputeFolderId
	elem["compute_id"] = vSphereCfg.ComputeId
	elem["data_network_ids"] = vSphereCfg.DataNetworkIds
	elem["default_gateway_address"] = vSphereCfg.DefaultGatewayAddresses
	elem["host_id"] = vSphereCfg.HostId
	elem["management_network_id"] = vSphereCfg.ManagementNetworkId

	var mpSubnets []map[string]interface{}
	for _, mps := range vSphereCfg.ManagementPortSubnets {
		e := make(map[string]interface{})
		e["ip_addresses"] = mps.IpAddresses
		e["prefix_length"] = mps.PrefixLength
		mpSubnets = append(mpSubnets, e)
	}
	elem["management_port_subnet"] = mpSubnets

	reservationInfo := make(map[string]interface{})
	if vSphereCfg.ReservationInfo.CpuReservation.ReservationInMhz != nil {
		reservationInfo["cpu_reservation_in_mhz"] = vSphereCfg.ReservationInfo.CpuReservation.ReservationInMhz
	}
	if vSphereCfg.ReservationInfo.CpuReservation.ReservationInShares != nil {
		reservationInfo["cpu_reservation_in_shares"] = vSphereCfg.ReservationInfo.CpuReservation.ReservationInShares
	}
	if vSphereCfg.ReservationInfo.MemoryReservation.ReservationPercentage != nil {
		reservationInfo["memory_reservation_percentage"] = vSphereCfg.ReservationInfo.MemoryReservation.ReservationPercentage
	}
	elem["reservation_info"] = []interface{}{reservationInfo}

	elem["storage_id"] = vSphereCfg.StorageId
	elem["vc_id"] = vSphereCfg.VcId

	return []interface{}{elem}, nil
}

func setHostSwitchSpecInSchema(d *schema.ResourceData, spec *data.StructValue, nodeType string) error {
	converter := bindings.NewTypeConverter()

	base, errs := converter.ConvertToGolang(spec, mpmodel.HostSwitchSpecBindingType())
	if errs != nil {
		return errs[0]
	}
	swType := base.(mpmodel.HostSwitchSpec).ResourceType

	switch swType {
	case mpmodel.HostSwitchSpec_RESOURCE_TYPE_STANDARDHOSTSWITCHSPEC:
		var swList []map[string]interface{}
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StandardHostSwitchSpecBindingType())
		if errs != nil {
			return errs[0]
		}
		swEntry := entry.(mpmodel.StandardHostSwitchSpec)
		for _, sw := range swEntry.HostSwitches {

			elem := make(map[string]interface{})
			elem["host_switch_id"] = sw.HostSwitchId
			elem["host_switch_name"] = sw.HostSwitchName
			profiles := setHostSwitchProfileIDsInSchema(sw.HostSwitchProfileIds)
			if len(profiles) > 0 {
				elem["host_switch_profile"] = profiles
			}
			var err error
			if sw.IpAssignmentSpec != nil {
				elem["ip_assignment"], err = setIPAssignmentInSchema(sw.IpAssignmentSpec)
			}
			if err != nil {
				return err
			}
			if sw.Ipv6AssignmentSpec != nil {
				elem["ipv6_assignment"], err = setIPv6AssignmentInSchema(sw.Ipv6AssignmentSpec)
			}
			if err != nil {
				return err
			}
			var pnics []map[string]interface{}
			for _, pnic := range sw.Pnics {
				e := make(map[string]interface{})
				e["device_name"] = pnic.DeviceName
				e["uplink_name"] = pnic.UplinkName
				pnics = append(pnics, e)
			}
			elem["pnic"] = pnics
			if nodeType == nodeTypeHost {
				var cpuConfig []map[string]interface{}
				for _, c := range sw.CpuConfig {
					e := make(map[string]interface{})
					e["num_lcores"] = c.NumLcores
					e["numa_node_index"] = c.NumaNodeIndex
					cpuConfig = append(cpuConfig, e)
				}
				elem["cpu_config"] = cpuConfig
				elem["is_migrate_pnics"] = sw.IsMigratePnics
				elem["host_switch_mode"] = sw.HostSwitchMode
				elem["uplink"] = setUplinksFromSchema(sw.Uplinks)
				var tnpSubConfig []map[string]interface{}
				for _, tnpsc := range sw.TransportNodeProfileSubConfigs {
					e := make(map[string]interface{})
					hsCfgOpt := make(map[string]interface{})
					hsCfgOpt["host_switch_id"] = tnpsc.HostSwitchConfigOption.HostSwitchId
					profiles := setHostSwitchProfileIDsInSchema(tnpsc.HostSwitchConfigOption.HostSwitchProfileIds)
					if len(profiles) > 0 {
						hsCfgOpt["host_switch_profile"] = profiles
					}
					if tnpsc.HostSwitchConfigOption.IpAssignmentSpec != nil {
						hsCfgOpt["ip_assignment"], err = setIPAssignmentInSchema(tnpsc.HostSwitchConfigOption.IpAssignmentSpec)
						if err != nil {
							return err
						}
					}
					if tnpsc.HostSwitchConfigOption.IpAssignmentSpec != nil {
						hsCfgOpt["ipv6_assignment"], err = setIPv6AssignmentInSchema(tnpsc.HostSwitchConfigOption.Ipv6AssignmentSpec)
						if err != nil {
							return err
						}
					}
					hsCfgOpt["uplink"] = setUplinksFromSchema(tnpsc.HostSwitchConfigOption.Uplinks)
					e["host_switch_config_option"] = []interface{}{hsCfgOpt}
					e["name"] = tnpsc.Name
					tnpSubConfig = append(tnpSubConfig, e)
				}
				elem["transport_node_profile_sub_config"] = tnpSubConfig

				var vmkIMList []map[string]interface{}
				for _, vmkIM := range sw.VmkInstallMigration {
					e := make(map[string]interface{})
					e["destination_network"] = vmkIM.DestinationNetwork
					e["device_name"] = vmkIM.DeviceName
					vmkIMList = append(vmkIMList, e)
				}
				elem["vmk_install_migration"] = vmkIMList
			}
			elem["transport_zone_endpoint"] = setTransportZoneEndpointInSchema(sw.TransportZoneEndpoints)

			swList = append(swList, elem)
		}
		d.Set("standard_host_switch", swList)

	case mpmodel.HostSwitchSpec_RESOURCE_TYPE_PRECONFIGUREDHOSTSWITCHSPEC:
		return fmt.Errorf("preconfigured host switch is not supported")
	}
	return nil
}

func setTransportZoneEndpointInSchema(endpoints []mpmodel.TransportZoneEndPoint) interface{} {
	var endpointList []map[string]interface{}
	for _, endpoint := range endpoints {
		e := make(map[string]interface{})
		e["transport_zone"] = endpoint.TransportZoneId
		var tzpIDs []string
		for _, tzpID := range endpoint.TransportZoneProfileIds {
			tzpIDs = append(tzpIDs, *tzpID.ProfileId)
		}
		e["transport_zone_profiles"] = tzpIDs
		endpointList = append(endpointList, e)
	}
	return endpointList
}

func setUplinksFromSchema(uplinks []mpmodel.VdsUplink) interface{} {
	var uplinkList []map[string]interface{}
	for _, uplink := range uplinks {
		e := make(map[string]interface{})
		e["uplink_name"] = uplink.UplinkName
		e["vds_lag_name"] = uplink.VdsLagName
		e["vds_uplink_name"] = uplink.VdsUplinkName
		uplinkList = append(uplinkList, e)
	}
	return uplinkList
}

func setIPAssignmentInSchema(spec *data.StructValue) (interface{}, error) {
	elem := make(map[string]interface{})

	converter := bindings.NewTypeConverter()
	base, errs := converter.ConvertToGolang(spec, mpmodel.IpAssignmentSpecBindingType())
	if errs != nil {
		return nil, errs[0]
	}
	assignmentType := base.(mpmodel.IpAssignmentSpec).ResourceType

	switch assignmentType {
	case mpmodel.IpAssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYDHCP:
		elem["assigned_by_dhcp"] = true

	case mpmodel.IpAssignmentSpec_RESOURCE_TYPE_NOIPV4:
		elem["no_ipv4"] = true

	case mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPLISTSPEC:
		e := make(map[string]interface{})
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpListSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpListSpec)
		e["default_gateway"] = ipAsEntry.DefaultGateway
		e["ip_addresses"] = ipAsEntry.IpList
		e["subnet_mask"] = ipAsEntry.SubnetMask
		elem["static_ip"] = []map[string]interface{}{e}

	case mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPPOOLSPEC:
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpPoolSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpPoolSpec)
		elem["static_ip_pool"] = ipAsEntry.IpPoolId

	case mpmodel.IpAssignmentSpec_RESOURCE_TYPE_STATICIPMACLISTSPEC:
		e := make(map[string]interface{})
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpMacListSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpMacListSpec)
		e["default_gateway"] = ipAsEntry.DefaultGateway
		e["subnet_mask"] = ipAsEntry.SubnetMask

		var ipMacPairList []map[string]interface{}
		for _, elem := range ipAsEntry.IpMacList {
			pair := make(map[string]interface{})
			pair["ip_address"] = elem.Ip
			pair["mac_address"] = elem.Mac
			ipMacPairList = append(ipMacPairList, pair)
		}
		e["ip_mac_pair"] = ipMacPairList
		elem["static_ip_mac"] = []map[string]interface{}{e}
	}

	return []interface{}{elem}, nil
}

func setIPv6AssignmentInSchema(spec *data.StructValue) (interface{}, error) {
	elem := make(map[string]interface{})

	converter := bindings.NewTypeConverter()
	base, errs := converter.ConvertToGolang(spec, mpmodel.Ipv6AssignmentSpecBindingType())
	if errs != nil {
		return nil, errs[0]
	}
	assignmentType := base.(mpmodel.Ipv6AssignmentSpec).ResourceType

	switch assignmentType {
	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYDHCPV6:
		elem["assigned_by_dhcpv6"] = true

	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_ASSIGNEDBYAUTOCONF:
		elem["assigned_by_autoconf"] = true

	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_NOIPV6:
		elem["no_ipv6"] = true

	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6LISTSPEC:
		e := make(map[string]interface{})
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpv6ListSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpv6ListSpec)
		e["default_gateway"] = ipAsEntry.DefaultGateway
		e["ip_addresses"] = ipAsEntry.Ipv6List
		e["prefix_length"] = ipAsEntry.PrefixLength
		elem["static_ip"] = []map[string]interface{}{e}

	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6POOLSPEC:
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpv6PoolSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpv6PoolSpec)
		elem["static_ip_pool"] = ipAsEntry.Ipv6PoolId

	case mpmodel.Ipv6AssignmentSpec_RESOURCE_TYPE_STATICIPV6MACLISTSPEC:
		e := make(map[string]interface{})
		entry, errs := converter.ConvertToGolang(spec, mpmodel.StaticIpv6MacListSpecBindingType())
		if errs != nil {
			return nil, errs[0]
		}
		ipAsEntry := entry.(mpmodel.StaticIpv6MacListSpec)
		e["default_gateway"] = ipAsEntry.DefaultGateway
		e["prefix_length"] = ipAsEntry.PrefixLength

		var ipMacPairList []map[string]interface{}
		for _, elem := range ipAsEntry.Ipv6MacList {
			pair := make(map[string]interface{})
			pair["ip_address"] = elem.Ipv6
			pair["mac_address"] = elem.Mac
			ipMacPairList = append(ipMacPairList, pair)
		}
		e["ip_mac_pair"] = ipMacPairList
		elem["static_ip_mac"] = []map[string]interface{}{e}
	}

	return []interface{}{elem}, nil
}

func setHostSwitchProfileIDsInSchema(hspIDs []mpmodel.HostSwitchProfileTypeIdEntry) []interface{} {
	var hostSwitchProfileIDs []interface{}
	for _, hspID := range hspIDs {
		hostSwitchProfileIDs = append(hostSwitchProfileIDs, hspID.Value)
	}
	return hostSwitchProfileIDs
}

func resourceNsxtEdgeTransportNodeUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := nsx.NewTransportNodesClient(connector)

	obj, err := getTransportNodeFromSchema(d, m)
	if err != nil {
		return handleUpdateError("TransportNode", id, err)
	}
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision

	_, err = client.Update(id, *obj, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleUpdateError("TransportNode", id, err)
	}

	return resourceNsxtEdgeTransportNodeRead(d, m)
}

func getTransportNodeStateConf(connector client.Connector, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{"notyet"},
		Target:  []string{"success", "failed"},
		Refresh: func() (interface{}, string, error) {
			client := transport_nodes.NewStateClient(connector)

			_, err := client.Get(id)

			if isNotFoundError(err) {
				return "success", "success", nil
			}

			if err != nil {
				log.Printf("[DEBUG]: NSX Failed to retrieve TransportNode state: %v", err)
				return nil, "failed", err
			}

			return "notyet", "notyet", nil
		},
		Delay:        time.Duration(5) * time.Second,
		Timeout:      time.Duration(1200) * time.Second,
		PollInterval: time.Duration(5) * time.Second,
	}
}

func resourceNsxtEdgeTransportNodeDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := nsx.NewTransportNodesClient(connector)

	err := client.Delete(id, nil, nil)
	if err != nil {
		return handleDeleteError("TransportNode", id, err)
	}

	stateConf := getTransportNodeStateConf(connector, id)
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("failed to get deletion status for %s: %v", id, err)
	}

	return nil
}
