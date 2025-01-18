// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	ippools "github.com/vmware/terraform-provider-nsxt/api/infra/ip_pools"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

var addressRealizationTimeoutDefault = int(1200)

func resourceNsxtPolicyIPAddressAllocation() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPAddressAllocationCreate,
		Read:   resourceNsxtPolicyIPAddressAllocationRead,
		Update: resourceNsxtPolicyIPAddressAllocationUpdate,
		Delete: resourceNsxtPolicyIPAddressAllocationDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyIPAddressAllocationImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"pool_path":    getPolicyPathSchema(true, true, "The path of the IP Pool for this allocation"),
			"allocation_ip": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The IP allocated. If unspecified any free IP will be allocated.",
				ValidateFunc: validateSingleIP(),
				Computed:     true,
				ForceNew:     true,
			},
			"timeout": {
				Type:         schema.TypeInt,
				Description:  "Realization timeout in seconds",
				Optional:     true,
				Default:      addressRealizationTimeoutDefault,
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func resourceNsxtPolicyIPAddressAllocationExists(sessionContext utl.SessionContext, poolID string, allocationID string, connector client.Connector) (bool, error) {
	client := ippools.NewIpAllocationsClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}

	_, err := client.Get(poolID, allocationID)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPAddressAllocationCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	sessionContext := getSessionContext(d, m)
	client := ippools.NewIpAllocationsClient(sessionContext, connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	poolID := getPolicyIDFromPath(d.Get("pool_path").(string))

	id := d.Get("nsx_id").(string)
	if id == "" {
		uuid, _ := uuid.NewRandom()
		id = uuid.String()
	}

	exists, err := resourceNsxtPolicyIPAddressAllocationExists(sessionContext, poolID, id, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Resource with ID %s already exists", id)
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	allocationIP := d.Get("allocation_ip").(string)

	obj := model.IpAddressAllocation{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	if allocationIP != "" {
		obj.AllocationIp = &allocationIP
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating IPAddressAllocation with ID %s", id)
	err = client.Patch(poolID, id, obj)
	if err != nil {
		return handleCreateError("IPAddressAllocation", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPAddressAllocationRead(d, m)
}

func resourceNsxtPolicyIPAddressAllocationRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := ippools.NewIpAllocationsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPAddressAllocation ID")
	}

	poolID := getPolicyIDFromPath(d.Get("pool_path").(string))

	obj, err := client.Get(poolID, id)
	if err != nil {
		return handleReadError(d, "IPAddressAllocation", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("pool_path", obj.ParentPath)

	d.Set("allocation_ip", obj.AllocationIp)

	if d.Get("allocation_ip").(string) == "" {
		timeout := d.Get("timeout").(int)
		log.Printf("[DEBUG] Waiting for realization of IP Address for IP Allocation with ID %s", id)

		stateConf := nsxtPolicyWaitForRealizationStateConf(connector, d, d.Get("path").(string), timeout)
		entity, err := stateConf.WaitForState()
		if err != nil {
			return err
		}
		realizedResource := entity.(model.GenericPolicyRealizedResource)
		for _, attr := range realizedResource.ExtendedAttributes {
			if *attr.Key == "allocation_ip" {
				d.Set("allocation_ip", attr.Values[0])
				return nil
			}
		}
		return fmt.Errorf("Failed to get realized IP for path %s", d.Get("path"))
	}

	return nil
}

func resourceNsxtPolicyIPAddressAllocationUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := ippools.NewIpAllocationsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Group ID")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	poolID := getPolicyIDFromPath(d.Get("pool_path").(string))

	obj := model.IpAddressAllocation{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
	}

	// Update the resource using PATCH
	log.Printf("[INFO] Updating IPAddressAllocation with ID %s", id)
	err := client.Patch(poolID, id, obj)
	if err != nil {
		return handleUpdateError("IPAddressAllocation", id, err)
	}

	return resourceNsxtPolicyIPAddressAllocationRead(d, m)
}

func resourceNsxtPolicyIPAddressAllocationDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := ippools.NewIpAllocationsClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPAddressAllocation ID")
	}

	poolID := getPolicyIDFromPath(d.Get("pool_path").(string))

	err := client.Delete(poolID, id)
	if err != nil {
		return handleDeleteError("IPAddressAllocation", id, err)
	}

	return nil
}

func resourceNsxtPolicyIPAddressAllocationImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")

	d.Set("timeout", addressRealizationTimeoutDefault)

	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		poolPath, err := getParameterFromPolicyPath("", "/ip-allocations/", importID)
		if err != nil {
			return nil, err
		}
		d.Set("pool_path", poolPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}

	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <ip-pool-id>/<allocation-id> as an input")
	}

	poolID := s[0]
	connector := getPolicyConnector(m)
	client := infra.NewIpPoolsClient(getSessionContext(d, m), connector)
	if client == nil {
		return nil, policyResourceNotSupportedError()
	}

	pool, err := client.Get(poolID)
	if err != nil {
		return nil, err
	}
	d.Set("pool_path", pool.Path)
	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
