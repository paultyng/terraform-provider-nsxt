// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyServiceCreate,
		Read:   resourceNsxtPolicyServiceRead,
		Update: resourceNsxtPolicyServiceUpdate,
		Delete: resourceNsxtPolicyServiceDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),

			"icmp_entry": {
				Type:        schema.TypeSet,
				Description: "ICMP type service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
						"protocol": {
							Type:         schema.TypeString,
							Description:  "Version of ICMP protocol (ICMPv4/ICMPv6)",
							Required:     true,
							ValidateFunc: validation.StringInSlice(icmpProtocolValues, false),
						},
						"icmp_type": {
							// NOTE: icmp_type is required if icmp_code is set
							Type:         schema.TypeString,
							Description:  "ICMP message type",
							Optional:     true,
							ValidateFunc: validateStringIntBetween(0, 255),
						},
						"icmp_code": {
							Type:         schema.TypeString,
							Description:  "ICMP message code",
							Optional:     true,
							ValidateFunc: validateStringIntBetween(0, 255),
						},
					},
				},
			},

			"l4_port_set_entry": {
				Type:        schema.TypeSet,
				Description: "L4 port set type service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
						"destination_ports": {
							Type:        schema.TypeSet,
							Description: "Set of destination ports",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validatePortRange(),
							},
							Optional: true,
						},
						"source_ports": {
							Type:        schema.TypeSet,
							Description: "Set of source ports",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validatePortRange(),
							},
							Optional: true,
						},
						"protocol": {
							Type:         schema.TypeString,
							Description:  "L4 Protocol",
							Required:     true,
							ValidateFunc: validation.StringInSlice(protocolValues, false),
						},
					},
				},
			},

			"igmp_entry": {
				Type:        schema.TypeSet,
				Description: "IGMP type service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
					},
				},
			},

			"ether_type_entry": {
				Type:          schema.TypeSet,
				Description:   "Ether type service entry",
				Optional:      true,
				ConflictsWith: []string{"algorithm_entry", "igmp_entry", "icmp_entry", "l4_port_set_entry", "ip_protocol_entry"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
						"ether_type": {
							Type:        schema.TypeInt,
							Description: "Type of the encapsulated protocol",
							Required:    true,
						},
					},
				},
			},

			"ip_protocol_entry": {
				Type:        schema.TypeSet,
				Description: "IP Protocol type service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
						"protocol": {
							Type:         schema.TypeInt,
							Description:  "IP protocol number",
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 255),
						},
					},
				},
			},

			"algorithm_entry": {
				Type:        schema.TypeSet,
				Description: "Algorithm type service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name": getOptionalDisplayNameSchema(false),
						"description":  getDescriptionSchema(),
						"destination_port": {
							Type:         schema.TypeString,
							Description:  "A single destination port",
							Required:     true,
							ValidateFunc: validateSinglePort(),
						},
						"source_ports": {
							Type:        schema.TypeSet,
							Description: "Set of source ports or ranges",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validatePortRange(),
							},
							Optional: true,
						},
						"algorithm": {
							Type:         schema.TypeString,
							Description:  "Algorithm",
							Required:     true,
							ValidateFunc: validation.StringInSlice(algTypeValues, false),
						},
					},
				},
			},

			"nested_service_entry": {
				Type:        schema.TypeSet,
				Description: "Nested service service entry",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"display_name":        getOptionalDisplayNameSchema(false),
						"description":         getDescriptionSchema(),
						"nested_service_path": getPolicyPathSchema(true, false, "Nested Service Path"),
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyServiceGetEntriesFromSchema(d *schema.ResourceData) ([]*data.StructValue, error) {
	converter := bindings.NewTypeConverter()
	serviceEntries := []*data.StructValue{}

	// ICMP Type service entries
	icmpEntries := d.Get("icmp_entry").(*schema.Set).List()
	for _, icmpEntry := range icmpEntries {
		entryData := icmpEntry.(map[string]interface{})
		// Type and code can be unset
		var typePtr *int64
		var codePtr *int64
		if entryData["icmp_type"] != "" {
			icmpType, err := strconv.Atoi(entryData["icmp_type"].(string))
			if err != nil {
				return serviceEntries, err
			}
			icmpType64 := int64(icmpType)
			typePtr = &icmpType64
		}
		if entryData["icmp_code"] != "" {
			icmpCode, err := strconv.Atoi(entryData["icmp_code"].(string))
			if err != nil {
				return serviceEntries, err
			}
			icmpCode64 := int64(icmpCode)
			codePtr = &icmpCode64
		}
		protocol := entryData["protocol"].(string)
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.ICMPTypeServiceEntry{
			Id:           &id,
			DisplayName:  &displayName,
			Description:  &description,
			IcmpType:     typePtr,
			IcmpCode:     codePtr,
			Protocol:     &protocol,
			ResourceType: model.ServiceEntry_RESOURCE_TYPE_ICMPTYPESERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.ICMPTypeServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// L4 port set Type service entries
	l4Entries := d.Get("l4_port_set_entry").(*schema.Set).List()
	for _, l4Entry := range l4Entries {
		entryData := l4Entry.(map[string]interface{})
		l4Protocol := entryData["protocol"].(string)
		sourcePorts := interface2StringList(entryData["source_ports"].(*schema.Set).List())
		destinationPorts := interface2StringList(entryData["destination_ports"].(*schema.Set).List())
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.L4PortSetServiceEntry{
			Id:               &id,
			DisplayName:      &displayName,
			Description:      &description,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
			L4Protocol:       &l4Protocol,
			ResourceType:     model.ServiceEntry_RESOURCE_TYPE_L4PORTSETSERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.L4PortSetServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// IGMP Type service entries
	igmpEntries := d.Get("igmp_entry").(*schema.Set).List()
	for _, igmpEntry := range igmpEntries {
		entryData := igmpEntry.(map[string]interface{})
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.IGMPTypeServiceEntry{
			Id:           &id,
			DisplayName:  &displayName,
			Description:  &description,
			ResourceType: model.ServiceEntry_RESOURCE_TYPE_IGMPTYPESERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.IGMPTypeServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// Ether Type service entries
	etherEntries := d.Get("ether_type_entry").(*schema.Set).List()
	for _, etherEntry := range etherEntries {
		entryData := etherEntry.(map[string]interface{})
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)
		etherType := int64(entryData["ether_type"].(int))

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.EtherTypeServiceEntry{
			Id:           &id,
			DisplayName:  &displayName,
			Description:  &description,
			EtherType:    &etherType,
			ResourceType: model.ServiceEntry_RESOURCE_TYPE_ETHERTYPESERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.EtherTypeServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// IP Protocol Type service entries
	ipProtEntries := d.Get("ip_protocol_entry").(*schema.Set).List()
	for _, ipProtEntry := range ipProtEntries {
		entryData := ipProtEntry.(map[string]interface{})
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)
		protocolNumber := int64(entryData["protocol"].(int))

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.IPProtocolServiceEntry{
			Id:             &id,
			DisplayName:    &displayName,
			Description:    &description,
			ProtocolNumber: &protocolNumber,
			ResourceType:   model.ServiceEntry_RESOURCE_TYPE_IPPROTOCOLSERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.IPProtocolServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// Algorithm Type service entries
	algEntries := d.Get("algorithm_entry").(*schema.Set).List()
	for _, algEntry := range algEntries {
		entryData := algEntry.(map[string]interface{})
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)
		alg := entryData["algorithm"].(string)
		sourcePorts := interface2StringList(entryData["source_ports"].(*schema.Set).List())
		destinationPorts := make([]string, 0, 1)
		destinationPorts = append(destinationPorts, entryData["destination_port"].(string))

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.ALGTypeServiceEntry{
			Id:               &id,
			DisplayName:      &displayName,
			Description:      &description,
			Alg:              &alg,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
			ResourceType:     model.ServiceEntry_RESOURCE_TYPE_ALGTYPESERVICEENTRY,
		}
		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.ALGTypeServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)
	}

	// Nested Service service entries
	nestedEntries := d.Get("nested_service_entry").(*schema.Set).List()
	for _, nestedEntry := range nestedEntries {
		entryData := nestedEntry.(map[string]interface{})
		displayName := entryData["display_name"].(string)
		description := entryData["description"].(string)
		nestedServicePath := entryData["nested_service_path"].(string)

		// Use a different random Id each time
		id := newUUID()

		serviceEntry := model.NestedServiceServiceEntry{
			Id:                &id,
			DisplayName:       &displayName,
			Description:       &description,
			NestedServicePath: &nestedServicePath,
			ResourceType:      model.ServiceEntry_RESOURCE_TYPE_NESTEDSERVICESERVICEENTRY,
		}

		dataValue, errs := converter.ConvertToVapi(serviceEntry, model.NestedServiceServiceEntryBindingType())
		if errs != nil {
			return serviceEntries, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		serviceEntries = append(serviceEntries, entryStruct)

	}

	return serviceEntries, nil
}

func resourceNsxtPolicyServiceExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewServicesClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}
	_, err := client.Get(id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving service", err)
}

func filterServiceEntryDisplayName(entryDisplayName string, entryID string) string {
	if entryDisplayName == entryID {
		return ""
	}
	return entryDisplayName
}

func resourceNsxtPolicyServiceCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyServiceExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	serviceEntries, errc := resourceNsxtPolicyServiceGetEntriesFromSchema(d)
	if errc != nil {
		return fmt.Errorf("Error during Service entries conversion: %v", errc)
	}

	obj := model.Service{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		ServiceEntries: serviceEntries,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating service with ID %s", id)

	client := infra.NewServicesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("Service", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyServiceRead(d, m)
}

func resourceNsxtPolicyServiceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining service id")
	}

	client := infra.NewServicesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Service", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	// Translate the returned service entries
	converter := bindings.NewTypeConverter()
	var icmpEntriesList []map[string]interface{}
	var l4EntriesList []map[string]interface{}
	var igmpEntriesList []map[string]interface{}
	var etherEntriesList []map[string]interface{}
	var ipProtEntriesList []map[string]interface{}
	var algEntriesList []map[string]interface{}
	var nestedServiceEntriesList []map[string]interface{}

	for _, entry := range obj.ServiceEntries {
		elem := make(map[string]interface{})
		base, errs := converter.ConvertToGolang(entry, model.ServiceEntryBindingType())
		resourceType := base.(model.ServiceEntry).ResourceType
		if errs != nil {
			return errs[0]
		}

		if resourceType == model.ServiceEntry_RESOURCE_TYPE_ICMPTYPESERVICEENTRY {
			icmpEntry, errs := converter.ConvertToGolang(entry, model.ICMPTypeServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := icmpEntry.(model.ICMPTypeServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			if serviceEntry.IcmpType != nil {
				elem["icmp_type"] = strconv.Itoa(int(*serviceEntry.IcmpType))
			} else {
				elem["icmp_type"] = ""
			}
			if serviceEntry.IcmpCode != nil {
				elem["icmp_code"] = strconv.Itoa(int(*serviceEntry.IcmpCode))
			} else {
				elem["icmp_code"] = ""
			}
			elem["protocol"] = serviceEntry.Protocol
			icmpEntriesList = append(icmpEntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_L4PORTSETSERVICEENTRY {
			l4Entry, errs := converter.ConvertToGolang(entry, model.L4PortSetServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := l4Entry.(model.L4PortSetServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			elem["destination_ports"] = serviceEntry.DestinationPorts
			elem["source_ports"] = serviceEntry.SourcePorts
			elem["protocol"] = serviceEntry.L4Protocol
			l4EntriesList = append(l4EntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_ETHERTYPESERVICEENTRY {
			etherEntry, errs := converter.ConvertToGolang(entry, model.EtherTypeServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := etherEntry.(model.EtherTypeServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			elem["ether_type"] = serviceEntry.EtherType
			etherEntriesList = append(etherEntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_IPPROTOCOLSERVICEENTRY {
			ipProtEntry, errs := converter.ConvertToGolang(entry, model.IPProtocolServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := ipProtEntry.(model.IPProtocolServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			elem["protocol"] = serviceEntry.ProtocolNumber
			ipProtEntriesList = append(ipProtEntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_ALGTYPESERVICEENTRY {
			algEntry, errs := converter.ConvertToGolang(entry, model.ALGTypeServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := algEntry.(model.ALGTypeServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			elem["algorithm"] = serviceEntry.Alg
			elem["destination_port"] = serviceEntry.DestinationPorts[0]
			elem["source_ports"] = serviceEntry.SourcePorts
			algEntriesList = append(algEntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_IGMPTYPESERVICEENTRY {
			igmpEntry, errs := converter.ConvertToGolang(entry, model.IGMPTypeServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := igmpEntry.(model.IGMPTypeServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			igmpEntriesList = append(igmpEntriesList, elem)
		} else if resourceType == model.ServiceEntry_RESOURCE_TYPE_NESTEDSERVICESERVICEENTRY {
			nestedEntry, errs := converter.ConvertToGolang(entry, model.NestedServiceServiceEntryBindingType())
			if errs != nil {
				return errs[0]
			}

			serviceEntry := nestedEntry.(model.NestedServiceServiceEntry)
			elem["display_name"] = filterServiceEntryDisplayName(*serviceEntry.DisplayName, *serviceEntry.Id)
			elem["description"] = serviceEntry.Description
			elem["nested_service_path"] = serviceEntry.NestedServicePath
			nestedServiceEntriesList = append(nestedServiceEntriesList, elem)

		} else {
			return fmt.Errorf("Unrecognized Service Entry Type %s", resourceType)
		}
	}

	err = d.Set("icmp_entry", icmpEntriesList)
	if err != nil {
		return err
	}

	err = d.Set("l4_port_set_entry", l4EntriesList)
	if err != nil {
		return err
	}

	err = d.Set("igmp_entry", igmpEntriesList)
	if err != nil {
		return err
	}

	err = d.Set("ether_type_entry", etherEntriesList)
	if err != nil {
		return err
	}

	err = d.Set("ip_protocol_entry", ipProtEntriesList)
	if err != nil {
		return err
	}

	err = d.Set("algorithm_entry", algEntriesList)
	if err != nil {
		return err
	}

	err = d.Set("nested_service_entry", nestedServiceEntriesList)
	if err != nil {
		return err
	}

	return nil
}

func resourceNsxtPolicyServiceUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining service id")
	}

	// Read the rest of the configured parameters
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	revision := int64(d.Get("revision").(int))
	tags := getPolicyTagsFromSchema(d)
	serviceEntries, errc := resourceNsxtPolicyServiceGetEntriesFromSchema(d)
	if errc != nil {
		return fmt.Errorf("Error during Service entries conversion: %v", errc)
	}
	obj := model.Service{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		ServiceEntries: serviceEntries,
		Revision:       &revision,
	}

	// Update the resource using Update to totally replace the list of entries
	client := infra.NewServicesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	_, err := client.Update(id, obj)

	if err != nil {
		return handleUpdateError("Service", id, err)
	}
	return resourceNsxtPolicyServiceRead(d, m)
}

func resourceNsxtPolicyServiceDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining service id")
	}

	connector := getPolicyConnector(m)

	doDelete := func() error {
		client := infra.NewServicesClient(getSessionContext(d, m), connector)
		if client == nil {
			return policyResourceNotSupportedError()
		}
		return client.Delete(id)
	}

	err := doDelete()

	if err != nil {
		return handleDeleteError("Service", id, err)
	}

	return nil
}
