// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
)

// Note - this resource is time consuming to configure since
// there is no API to search VMs by host ID

func resourceNsxtVMTags() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVMTagsCreate,
		Read:   resourceNsxtVMTagsRead,
		Update: resourceNsxtVMTagsUpdate,
		Delete: resourceNsxtVMTagsDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:        schema.TypeString,
				Description: "Instance id",
				Required:    true,
			},
			"tag":              getTagsSchema(),
			"logical_port_tag": getTagsSchema(),
		},
	}
}

func getVMList(nsxClient *api.APIClient) (*manager.VirtualMachineListResult, error) {

	localVarOptionals := make(map[string]interface{})
	vmList, resp, err := nsxClient.FabricApi.ListVirtualMachines(nsxClient.Context, localVarOptionals)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status returned during operation on VM Tags: %v", resp.StatusCode)
	}

	return &vmList, nil

}

func getVIFList(nsxClient *api.APIClient) (*manager.VirtualNetworkInterfaceListResult, error) {

	localVarOptionals := make(map[string]interface{})
	vifList, resp, err := nsxClient.FabricApi.ListVifs(nsxClient.Context, localVarOptionals)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status returned during operation on Logical Port Tags: %v", resp.StatusCode)
	}

	return &vifList, nil

}

func getPortList(nsxClient *api.APIClient) (*manager.LogicalPortListResult, error) {
	localVarOptionals := make(map[string]interface{})
	portList, resp, err := nsxClient.LogicalSwitchingApi.ListLogicalPorts(nsxClient.Context, localVarOptionals)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status returned during operation on Logical Port Tags: %v", resp.StatusCode)
	}

	return &portList, nil
}

func findVMIDByLocalID(nsxClient *api.APIClient, localID string) (*manager.VirtualMachine, error) {
	// VM provider (for example, vsphere provider) can supply host id(BIOS id) for the VM
	// NSX API needs NSX id for the update tags API call
	vmList, err := getVMList(nsxClient)
	if err != nil {
		return nil, err
	}

	for _, elem := range vmList.Results {
		for _, computeID := range elem.ComputeIds {
			log.Printf("[DEBUG] Inspecting compute id %s for VM %s", computeID, elem.LocalIdOnHost)
			slices := strings.Split(computeID, ":")
			if len(slices) == 2 && (slices[0] == "biosUuid") && (slices[1] == localID) {
				return &elem, nil
			}
		}
	}

	return nil, fmt.Errorf("Failed to find Virtual Machine with local id %s in inventory", localID)
}

func findVMByExternalID(nsxClient *api.APIClient, instanceID string) (*manager.VirtualMachine, error) {
	vmList, err := getVMList(nsxClient)
	if err != nil {
		return nil, err
	}

	for _, elem := range vmList.Results {
		if elem.ExternalId == instanceID {
			return &elem, nil
		}
	}

	return nil, fmt.Errorf("Failed to find Virtual Machine with id %s in inventory", instanceID)
}

func findVIFsByExternalID(nsxClient *api.APIClient, instanceID string) ([]manager.VirtualNetworkInterface, error) {

	vifList, err := getVIFList(nsxClient)
	var resultList []manager.VirtualNetworkInterface
	if err != nil {
		return nil, err
	}

	for _, vif := range vifList.Results {
		if vif.OwnerVmId == instanceID {
			resultList = append(resultList, vif)
		}
	}

	return resultList, nil
}

func findPortsByExternalID(nsxClient *api.APIClient, instanceID string) ([]manager.LogicalPort, error) {

	var resultList []manager.LogicalPort
	vifs, err := findVIFsByExternalID(nsxClient, instanceID)
	if err != nil {
		return resultList, err
	}

	portList, err := getPortList(nsxClient)
	if err != nil {
		return resultList, err
	}
	for _, elem := range portList.Results {
		for _, vif := range vifs {
			if elem.Attachment != nil && elem.Attachment.Id == vif.LportAttachmentId {
				resultList = append(resultList, elem)
				continue
			}
		}
	}

	return resultList, nil
}

func updateTags(nsxClient *api.APIClient, id string, tags []common.Tag) error {
	log.Printf("[DEBUG] Updating tags for %s", id)

	tagsUpdate := manager.VirtualMachineTagUpdate{
		ExternalId: id,
		Tags:       tags,
	}

	resp, err := nsxClient.FabricApi.UpdateVirtualMachineTagsUpdateTags(nsxClient.Context, tagsUpdate)

	if err != nil {
		return fmt.Errorf("Error during operation on Tags for VM %s: %v", id, err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Unexpected status returned during operation on Tags for VM %d: %v", resp.StatusCode, id)
	}

	return nil
}

func updatePortTags(nsxClient *api.APIClient, id string, tags []common.Tag) error {
	log.Printf("[DEBUG] Updating logical port tags for %s", id)

	ports, err := findPortsByExternalID(nsxClient, id)
	if err != nil {
		return err
	}

	portsUpdated := 0

	for _, port := range ports {
		port.Tags = tags
		log.Printf("[DEBUG] Applying %d tags on logical port %s", len(tags), port.Id)
		_, resp, err := nsxClient.LogicalSwitchingApi.UpdateLogicalPort(nsxClient.Context, port.Id, port)

		if err != nil || (resp != nil && resp.StatusCode == http.StatusNotFound) {
			return fmt.Errorf("Error while updating tags on logical port %s: %v", port.Id, err)
		}
		portsUpdated++
	}

	log.Printf("[INFO] Applied %d tags on %d logical ports for VM %s", len(tags), portsUpdated, id)

	return nil
}

func resourceNsxtVMTagsCreate(d *schema.ResourceData, m interface{}) error {
	instanceID := d.Get("instance_id").(string)

	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	vm, err := findVMIDByLocalID(nsxClient, instanceID)
	if err != nil {
		return fmt.Errorf("Error during VM retrieval: %v", err)
	}

	tags := getTagsFromSchema(d)
	if len(tags) > 0 || d.HasChange("tag") {
		err = updateTags(nsxClient, vm.ExternalId, tags)
		if err != nil {
			return err
		}
	}

	portTags := getCustomizedTagsFromSchema(d, "logical_port_tag")
	if len(portTags) > 0 || d.HasChange("logical_port_tag") {
		err = updatePortTags(nsxClient, vm.ExternalId, portTags)
		if err != nil {
			return err
		}
	}

	d.SetId(vm.ExternalId)

	return resourceNsxtVMTagsRead(d, m)
}

func resourceNsxtVMTagsRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	vm, err := findVMByExternalID(nsxClient, id)
	if err != nil {
		return fmt.Errorf("Error during VM retrieval: %v", err)
	}

	ports, err := findPortsByExternalID(nsxClient, id)
	if err != nil {
		return fmt.Errorf("Error during logical port retrieval: %v", err)
	}

	setTagsInSchema(d, vm.Tags)
	// assuming all ports have same tags
	// note - more flexible implementation will be provided with policy resource
	if len(ports) > 0 {
		setCustomizedTagsInSchema(d, ports[0].Tags, "logical_port_tag")
	} else {
		// assign empty list
		var tagList []map[string]string
		d.Set("logical_port_tag", tagList)
	}

	return nil
}

func resourceNsxtVMTagsUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceNsxtVMTagsCreate(d, m)
}

func resourceNsxtVMTagsDelete(d *schema.ResourceData, m interface{}) error {
	instanceID := d.Get("instance_id").(string)

	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	vm, err := findVMIDByLocalID(nsxClient, instanceID)
	if err != nil {
		return fmt.Errorf("Error during VM retrieval: %v", err)
	}

	noTags := make([]common.Tag, 0)
	vmTags := getTagsFromSchema(d)
	if len(vmTags) > 0 {
		// Update tags only if they were configured by the provider
		err = updateTags(nsxClient, vm.ExternalId, noTags)
		if err != nil {
			return err
		}
	}

	portTags := getCustomizedTagsFromSchema(d, "logical_port_tag")
	if len(portTags) > 0 {
		// Update port tags only if they were configured by the provider
		err := updatePortTags(nsxClient, vm.ExternalId, noTags)
		if err != nil {
			return err
		}
	}

	return nil
}
