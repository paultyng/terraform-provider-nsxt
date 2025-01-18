// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var algTypeValues = []string{"ORACLE_TNS", "FTP", "SUN_RPC_TCP", "SUN_RPC_UDP", "MS_RPC_TCP", "MS_RPC_UDP", "NBNS_BROADCAST", "NBDG_BROADCAST", "TFTP"}

func resourceNsxtAlgorithmTypeNsService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtAlgorithmTypeNsServiceCreate,
		Read:   resourceNsxtAlgorithmTypeNsServiceRead,
		Update: resourceNsxtAlgorithmTypeNsServiceUpdate,
		Delete: resourceNsxtAlgorithmTypeNsServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
			"default_service": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether this is a default NSServices which can't be modified/deleted",
				Computed:    true,
			},
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
	}
}

func resourceNsxtAlgorithmTypeNsServiceCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	alg := d.Get("algorithm").(string)
	sourcePorts := getStringListFromSchemaSet(d, "source_ports")
	destinationPorts := make([]string, 0, 1)
	destinationPorts = append(destinationPorts, d.Get("destination_port").(string))

	nsService := manager.AlgTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
		},
		NsserviceElement: manager.AlgTypeNsServiceEntry{
			ResourceType:     "ALGTypeNSService",
			Alg:              alg,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
		},
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.CreateAlgTypeNSService(nsxClient.Context, nsService)

	if err != nil {
		return fmt.Errorf("Error during NsService create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during NsService create: %v", resp.StatusCode)
	}
	d.SetId(nsService.Id)
	return resourceNsxtAlgorithmTypeNsServiceRead(d, m)
}

func resourceNsxtAlgorithmTypeNsServiceRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	nsService, resp, err := nsxClient.GroupingObjectsApi.ReadAlgTypeNSService(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsService %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during NsService read: %v", err)
	}

	nsserviceElement := nsService.NsserviceElement

	d.Set("revision", nsService.Revision)
	d.Set("description", nsService.Description)
	d.Set("display_name", nsService.DisplayName)
	setTagsInSchema(d, nsService.Tags)
	d.Set("default_service", nsService.DefaultService)
	d.Set("algorithm", nsserviceElement.Alg)
	d.Set("destination_port", nsserviceElement.DestinationPorts[0])
	d.Set("source_ports", nsserviceElement.SourcePorts)

	return nil
}

func resourceNsxtAlgorithmTypeNsServiceUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	alg := d.Get("algorithm").(string)
	sourcePorts := getStringListFromSchemaSet(d, "source_ports")
	destinationPorts := make([]string, 0, 1)
	destinationPorts = append(destinationPorts, d.Get("destination_port").(string))
	revision := int64(d.Get("revision").(int))

	nsService := manager.AlgTypeNsService{
		NsService: manager.NsService{
			Description: description,
			DisplayName: displayName,
			Tags:        tags,
			Revision:    revision,
		},
		NsserviceElement: manager.AlgTypeNsServiceEntry{
			ResourceType:     "ALGTypeNSService",
			Alg:              alg,
			DestinationPorts: destinationPorts,
			SourcePorts:      sourcePorts,
		},
	}

	_, resp, err := nsxClient.GroupingObjectsApi.UpdateAlgTypeNSService(nsxClient.Context, id, nsService)
	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during NsService update: %v %v", err, resp)
	}

	return resourceNsxtAlgorithmTypeNsServiceRead(d, m)
}

func resourceNsxtAlgorithmTypeNsServiceDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining ns service id")
	}

	localVarOptionals := make(map[string]interface{})
	localVarOptionals["force"] = true
	resp, err := nsxClient.GroupingObjectsApi.DeleteNSService(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during NsService delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] NsService %s not found", id)
		d.SetId("")
	}
	return nil
}
