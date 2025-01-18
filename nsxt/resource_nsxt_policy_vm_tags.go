// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	realizedstate "github.com/vmware/terraform-provider-nsxt/api/infra/realized_state"
	"github.com/vmware/terraform-provider-nsxt/api/infra/segments"
	t1_segments "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s/segments"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/realized_state/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/realized_state/virtual_machines"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var (
	nsxtPolicyBiosUUIDKey     = "biosUuid"
	nsxtPolicyInstanceUUIDKey = "instanceUuid"
)

func resourceNsxtPolicyVMTags() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyVMTagsCreate,
		Read:   resourceNsxtPolicyVMTagsRead,
		Update: resourceNsxtPolicyVMTagsUpdate,
		Delete: resourceNsxtPolicyVMTagsDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Description: "Instance id",
				Required:    true,
			},
			"tag": getTagsSchema(),
			"port": {
				Type:        schema.TypeList,
				Description: "Tag specificiation for corresponding segment port",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"segment_path": getPolicyPathSchema(true, true, "Segment path where VM port should be tagged"),
						"tag":          getTagsSchema(),
					},
				},
			},
			"context": getContextSchema(false, false, false),
		},
	}
}

func listAllPolicyVirtualMachines(context utl.SessionContext, connector client.Connector, m interface{}) ([]model.VirtualMachine, error) {
	client := realizedstate.NewVirtualMachinesClient(context, connector)
	if client == nil {
		return nil, policyResourceNotSupportedError()
	}
	var results []model.VirtualMachine
	boolFalse := false
	var cursor *string
	total := 0

	for {
		// NOTE: Search API doesn't filter by realized state resources
		// NOTE: Contrary to the spec, this API does not populate cursor and result count
		// parameters, respects cursor input. Therefore we determine end of VM list by
		// looking for empty result.
		sortBy := "external_id"
		var efPtr *string
		if getPolicyEnforcementPoint(m) != "default" {
			// To minimize changes, avoid passing enforcement point unless its specified in provider
			enforcementPointPath := getPolicyEnforcementPointPath(m)
			efPtr = &enforcementPointPath
		}
		vms, err := client.List(cursor, efPtr, &boolFalse, nil, nil, &boolFalse, &sortBy)
		if err != nil {
			return results, err
		}
		results = append(results, vms.Results...)
		if len(vms.Results) == 0 {
			// no more results
			return results, nil
		}
		if total == 0 && vms.ResultCount != nil {
			// first response
			total = int(*vms.ResultCount)
		}
		cursor = vms.Cursor
		if cursor == nil {
			resultCount := strconv.Itoa(len(results))
			cursor = &resultCount
		}

		if (total > 0) && (len(results) >= total) {
			return results, nil
		}
	}
}

func listAllPolicySegmentPorts(context utl.SessionContext, connector client.Connector, segmentPath string) ([]model.SegmentPort, error) {

	var results []model.SegmentPort
	isT0, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
	if isT0 || len(segmentID) == 0 {
		return results, fmt.Errorf("invalid segment path %s", segmentPath)
	}
	boolFalse := false
	var cursor *string
	total := 0
	var err error
	var ports model.SegmentPortListResult

	for {
		if len(gwID) == 0 {
			client := segments.NewPortsClient(context, connector)
			if client == nil {
				return nil, policyResourceNotSupportedError()
			}
			ports, err = client.List(segmentID, cursor, &boolFalse, nil, nil, &boolFalse, nil)
		} else {
			// fixed segments
			client := t1_segments.NewPortsClient(context, connector)
			if client == nil {
				return nil, policyResourceNotSupportedError()
			}
			ports, err = client.List(gwID, segmentID, cursor, &boolFalse, nil, nil, &boolFalse, nil)

		}
		if err != nil {
			return results, err
		}
		results = append(results, ports.Results...)
		if total == 0 && ports.ResultCount != nil {
			// first response
			total = int(*ports.ResultCount)
		}
		cursor = ports.Cursor
		if len(results) >= total {
			log.Printf("[DEBUG] Found %d ports for segment %s", len(results), segmentID)
			return results, nil
		}
	}
}

func listAllPolicyVifs(m interface{}) ([]model.VirtualNetworkInterface, error) {

	client := enforcement_points.NewVifsClient(getPolicyConnector(m))
	var results []model.VirtualNetworkInterface
	var cursor *string
	total := 0

	enforcementPointPath := getPolicyEnforcementPoint(m)
	for {
		// NOTE: Search API doesn't filter by realized state resources
		vifs, err := client.List(enforcementPointPath, cursor, nil, nil, nil, nil, nil)
		if err != nil {
			return results, err
		}
		results = append(results, vifs.Results...)
		if total == 0 && vifs.ResultCount != nil {
			// first response
			total = int(*vifs.ResultCount)
		}
		cursor = vifs.Cursor
		if len(results) >= total {
			return results, nil
		}
	}
}

func findNsxtPolicyVMByNamePrefix(context utl.SessionContext, connector client.Connector, namePrefix string, m interface{}) ([]model.VirtualMachine, []model.VirtualMachine, error) {
	var perfectMatch, prefixMatch, allVMs []model.VirtualMachine
	var err error

	if util.NsxVersionHigherOrEqual("4.1.2") {
		// Search API works for inventory objects for 4.1.2 and above
		resourceType := "VirtualMachine"
		resultValues, err1 := listInventoryResourcesByNameAndType(connector, context, namePrefix, resourceType, nil)
		if err1 != nil {
			log.Printf("[INFO] Error searching for Virtual Machine with name: %s", namePrefix)
			return perfectMatch, prefixMatch, err1
		}
		allVMs, err = convertSearchResultToVMList(resultValues)

	} else {
		allVMs, err = listAllPolicyVirtualMachines(context, connector, m)
	}
	if err != nil {
		log.Printf("[INFO] Error reading Virtual Machine with name: %s", namePrefix)
		return perfectMatch, prefixMatch, err
	}

	for _, vmResult := range allVMs {
		if *vmResult.DisplayName == namePrefix {
			perfectMatch = append(perfectMatch, vmResult)
		} else if strings.HasPrefix(*vmResult.DisplayName, namePrefix) {
			prefixMatch = append(prefixMatch, vmResult)
		}
	}
	return perfectMatch, prefixMatch, nil
}

func convertSearchResultToVMList(searchResults []*data.StructValue) ([]model.VirtualMachine, error) {
	var vms []model.VirtualMachine
	converter := bindings.NewTypeConverter()

	for _, item := range searchResults {
		dataValue, errors := converter.ConvertToGolang(item, model.VirtualMachineBindingType())
		if len(errors) > 0 {
			return vms, errors[0]
		}
		vm := dataValue.(model.VirtualMachine)
		vms = append(vms, vm)
	}
	return vms, nil
}

func findNsxtPolicyVMByID(context utl.SessionContext, connector client.Connector, vmID string, m interface{}) (model.VirtualMachine, error) {
	var virtualMachineStruct model.VirtualMachine
	var vms []model.VirtualMachine
	var err error

	if util.NsxVersionHigherOrEqual("4.1.2") {
		// Search API works for inventory objects for 4.1.2 and above
		resourceType := "VirtualMachine"
		resultValues, err1 := listInventoryResourcesByAnyFieldAndType(connector, context, vmID, resourceType, nil)
		if err1 != nil {
			log.Printf("[INFO] Error searching for Virtual Machine with id: %s", vmID)
			return virtualMachineStruct, err1
		}

		vms, err = convertSearchResultToVMList(resultValues)
	} else {
		vms, err = listAllPolicyVirtualMachines(context, connector, m)
	}
	if err != nil {
		log.Printf("[INFO] Error reading Virtual Machines when looking for ID: %s", vmID)
		return virtualMachineStruct, err
	}

	for _, vmResult := range vms {
		if (vmResult.ExternalId != nil) && *vmResult.ExternalId == vmID {
			return vmResult, nil
		}
		computeIDMap := collectSeparatedStringListToMap(vmResult.ComputeIds, ":")
		if computeIDMap[nsxtPolicyBiosUUIDKey] == vmID {
			return vmResult, nil
		} else if computeIDMap[nsxtPolicyInstanceUUIDKey] == vmID {
			return vmResult, nil
		}
	}
	return virtualMachineStruct, fmt.Errorf("Could not find Virtual Machine with ID: %s", vmID)
}

func updateNsxtPolicyVMTags(connector client.Connector, externalID string, tags []model.Tag, m interface{}) error {
	tagUpdate := model.VirtualMachineTagsUpdate{
		Tags:             tags,
		VirtualMachineId: &externalID,
	}
	if util.NsxVersionHigherOrEqual("4.1.1") {
		client := virtual_machines.NewTagsClient(connector)
		return client.Create(externalID, tagUpdate, nil, nil, nil, nil, nil, nil, nil)
	}
	client := enforcement_points.NewVirtualMachinesClient(connector)

	return client.Updatetags(getPolicyEnforcementPoint(m), tagUpdate)
}

func listPolicyVifAttachmentsForVM(m interface{}, externalID string) ([]string, error) {
	var vifAttachmentIds []string
	vifs, err := listAllPolicyVifs(m)
	if err != nil {
		return vifAttachmentIds, err
	}

	for _, vif := range vifs {
		if (vif.LportAttachmentId != nil) && (vif.OwnerVmId != nil) && *vif.OwnerVmId == externalID {
			vifAttachmentIds = append(vifAttachmentIds, *vif.LportAttachmentId)
		}
	}

	return vifAttachmentIds, nil
}

func updateNsxtPolicyVMPortTags(context utl.SessionContext, connector client.Connector, externalID string, portTags []interface{}, m interface{}, isDelete bool) error {

	vifAttachmentIds, err := listPolicyVifAttachmentsForVM(m, externalID)
	if err != nil {
		return err
	}

	for _, portTag := range portTags {
		data := portTag.(map[string]interface{})
		var tags []model.Tag
		segmentPath := data["segment_path"].(string)
		if !isDelete {
			tags = getPolicyTagsFromSet(data["tag"].(*schema.Set))
		}

		ports, portsErr := listAllPolicySegmentPorts(context, connector, segmentPath)
		if portsErr != nil {
			return portsErr
		}
		_, gwID, segmentID := parseSegmentPolicyPath(segmentPath)
		for _, port := range ports {
			if port.Attachment == nil || port.Attachment.Id == nil {
				continue
			}

			for _, attachment := range vifAttachmentIds {
				if attachment == *port.Attachment.Id {
					port.Tags = tags
					log.Printf("[DEBUG] Updating port %s with %d tags", *port.Path, len(tags))
					if len(gwID) == 0 {
						client := segments.NewPortsClient(context, connector)
						if client == nil {
							return policyResourceNotSupportedError()
						}
						_, err = client.Update(segmentID, *port.Id, port)
					} else {
						// fixed segment
						client := t1_segments.NewPortsClient(context, connector)
						if client == nil {
							return policyResourceNotSupportedError()
						}
						_, err = client.Update(gwID, segmentID, *port.Id, port)
					}
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}

func setPolicyVMPortTagsInSchema(d *schema.ResourceData, m interface{}, externalID string) error {

	connector := getPolicyConnector(m)
	vifAttachmentIds, err := listPolicyVifAttachmentsForVM(m, externalID)
	if err != nil {
		return err
	}
	context := getSessionContext(d, m)

	portTags := d.Get("port").([]interface{})
	var actualPortTags []map[string]interface{}
	for _, portTag := range portTags {
		data := portTag.(map[string]interface{})
		segmentPath := data["segment_path"].(string)

		ports, portsErr := listAllPolicySegmentPorts(context, connector, segmentPath)
		if portsErr != nil {
			return portsErr
		}
		for _, port := range ports {
			if port.Attachment == nil || port.Attachment.Id == nil {
				continue
			}

			for _, attachment := range vifAttachmentIds {
				if attachment == *port.Attachment.Id {
					tags := make(map[string]interface{})
					tags["segment_path"] = segmentPath
					tags["tag"] = initPolicyTagsSet(port.Tags)
					actualPortTags = append(actualPortTags, tags)
				}
			}
		}
	}

	d.Set("port", actualPortTags)

	return nil
}

func resourceNsxtPolicyVMTagsRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	vmID := d.Id()
	if vmID == "" {
		return fmt.Errorf("Error obtaining Virtual Machine ID")
	}

	vm, err := findNsxtPolicyVMByID(getSessionContext(d, m), connector, vmID, m)
	if err != nil {
		d.SetId("")
		log.Printf("[ERROR] Cannot find VM with ID %s, skip reading VM tag", vmID)
		return nil
	}

	setPolicyTagsInSchema(d, vm.Tags)

	if d.Get("instance_id") == "" {
		// for import
		d.Set("instance_id", vm.ExternalId)
	}

	return setPolicyVMPortTagsInSchema(d, m, *vm.ExternalId)
}

func resourceNsxtPolicyVMTagsCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	instanceID := d.Get("instance_id").(string)
	context := getSessionContext(d, m)

	vm, err := findNsxtPolicyVMByID(context, connector, instanceID, m)
	if err != nil {
		if len(d.Id()) == 0 {
			// This is create flow
			return fmt.Errorf("Cannot find VM with ID %s", instanceID)
		}
		// Update flow - when VM is deleted, we don't want to error out,
		// Customers prefer to swallow the error
		log.Printf("[ERROR] Cannot find VM with ID %s, swallowing error", instanceID)
		return nil
	}

	tags := getPolicyTagsFromSchema(d)
	if tags == nil {
		tags = make([]model.Tag, 0)
	}

	err = updateNsxtPolicyVMTags(connector, *vm.ExternalId, tags, m)
	if err != nil {
		return handleCreateError("Virtual Machine Tag", *vm.ExternalId, err)
	}

	portTags := d.Get("port").([]interface{})
	err = updateNsxtPolicyVMPortTags(context, connector, *vm.ExternalId, portTags, m, false)
	if err != nil {
		return handleCreateError("Segment Port Tag", *vm.ExternalId, err)
	}

	d.SetId(*vm.ExternalId)
	d.Set("port", portTags)

	return resourceNsxtPolicyVMTagsRead(d, m)
}

func resourceNsxtPolicyVMTagsUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtPolicyVMTagsCreate(d, m)
}

func resourceNsxtPolicyVMTagsDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	instanceID := d.Get("instance_id").(string)
	context := getSessionContext(d, m)

	vm, err := findNsxtPolicyVMByID(context, connector, instanceID, m)

	if err != nil {
		d.SetId("")
		log.Printf("[ERROR] Cannot find VM with ID %s, deleting stale VM Tag on provider", instanceID)
		return nil
	}

	tags := make([]model.Tag, 0)
	err = updateNsxtPolicyVMTags(connector, *vm.ExternalId, tags, m)

	if err != nil {
		return handleDeleteError("Virtual Machine Tag", *vm.ExternalId, err)
	}

	portTags := d.Get("port").([]interface{})
	err = updateNsxtPolicyVMPortTags(context, connector, *vm.ExternalId, portTags, m, true)
	if err != nil {
		return handleDeleteError("Segment Port Tag", *vm.ExternalId, err)
	}

	return err
}
