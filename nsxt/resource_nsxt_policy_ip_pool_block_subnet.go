/* Copyright © 2019 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	ippools "github.com/vmware/terraform-provider-nsxt/api/infra/ip_pools"
	realizedstate "github.com/vmware/terraform-provider-nsxt/api/infra/realized_state"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyIPPoolBlockSubnet() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPPoolBlockSubnetCreate,
		Read:   resourceNsxtPolicyIPPoolBlockSubnetRead,
		Update: resourceNsxtPolicyIPPoolBlockSubnetUpdate,
		Delete: resourceNsxtPolicyIPPoolBlockSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyIPPoolSubnetImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"auto_assign_gateway": {
				Type:        schema.TypeBool,
				Description: "If true, the first IP in the range will be reserved for gateway",
				Optional:    true,
				Default:     true,
			},
			"size": {
				Type:         schema.TypeInt,
				Description:  "Number of addresses",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePowerOf2(false, 0),
			},
			"pool_path":  getPolicyPathSchema(true, true, "Policy path to the IP Pool for this Subnet"),
			"block_path": getPolicyPathSchema(true, true, "Policy path to the IP Block"),
		},
	}
}

func resourceNsxtPolicyIPPoolBlockSubnetSchemaToStructValue(d *schema.ResourceData, id string) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	blockPath := d.Get("block_path").(string)
	autoAssignGateway := d.Get("auto_assign_gateway").(bool)
	size := d.Get("size").(int)
	size64 := int64(size)
	tags := getPolicyTagsFromSchema(d)

	obj := model.IpAddressPoolBlockSubnet{
		DisplayName:       &displayName,
		Description:       &description,
		Tags:              tags,
		Size:              &size64,
		AutoAssignGateway: &autoAssignGateway,
		ResourceType:      "IpAddressPoolBlockSubnet",
		IpBlockPath:       &blockPath,
		Id:                &id,
	}

	dataValue, errors := converter.ConvertToVapi(obj, model.IpAddressPoolBlockSubnetBindingType())
	if errors != nil {
		return nil, fmt.Errorf("Error converting Block Subnet: %v", errors[0])
	}

	return dataValue.(*data.StructValue), nil
}

func resourceNsxtPolicyIPPoolBlockSubnetRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)
	converter := bindings.NewTypeConverter()

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Block Subnet ID")
	}

	subnetData, err := client.Get(poolID, id)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			log.Printf("[DEBUG] Block Subnet %s not found", id)
			return nil
		}
		return handleReadError(d, "Block Subnet", id, err)
	}

	snet, errs := converter.ConvertToGolang(subnetData, model.IpAddressPoolBlockSubnetBindingType())
	if len(errs) > 0 {
		return fmt.Errorf("Error converting Block Subnet %s", errs[0])
	}
	blockSubnet := snet.(model.IpAddressPoolBlockSubnet)

	d.Set("display_name", blockSubnet.DisplayName)
	d.Set("description", blockSubnet.Description)
	setPolicyTagsInSchema(d, blockSubnet.Tags)
	d.Set("nsx_id", blockSubnet.Id)
	d.Set("path", blockSubnet.Path)
	d.Set("revision", blockSubnet.Revision)
	d.Set("auto_assign_gateway", blockSubnet.AutoAssignGateway)
	d.Set("size", blockSubnet.Size)
	d.Set("pool_path", poolPath)
	d.Set("block_path", blockSubnet.IpBlockPath)

	return nil
}

func resourceNsxtPolicyIPPoolBlockSubnetCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := client.Get(poolID, id)
		if err == nil {
			return fmt.Errorf("Block Subnet with ID '%s' already exists on Pool %s", id, poolID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	dataValue, err := resourceNsxtPolicyIPPoolBlockSubnetSchemaToStructValue(d, id)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating IP Pool Block Subnet with ID %s", id)
	err = client.Patch(poolID, id, dataValue)
	if err != nil {
		return handleCreateError("Block Subnet", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyIPPoolBlockSubnetRead(d, m)
}

func resourceNsxtPolicyIPPoolBlockSubnetUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Block Subnet ID")
	}

	dataValue, err := resourceNsxtPolicyIPPoolBlockSubnetSchemaToStructValue(d, id)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating IP Pool Block Subnet with ID %s", id)
	err = client.Patch(poolID, id, dataValue)
	if err != nil {
		return handleUpdateError("Block Subnet", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyIPPoolBlockSubnetRead(d, m)
}

func resourceNsxtPolicyIPPoolBlockSubnetDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := ippools.NewIpSubnetsClient(context, connector)

	poolPath := d.Get("pool_path").(string)
	poolID := getPolicyIDFromPath(poolPath)

	id := d.Id()
	if id == "" || poolID == "" {
		return fmt.Errorf("Error obtaining Block Subnet ID")
	}

	log.Printf("[INFO] Deleting Block Subnet with ID %s", id)
	err = client.Delete(poolID, id, nil)
	if err != nil {
		return handleDeleteError("Block Subnet", id, err)
	}

	return resourceNsxtPolicyIPPoolBlockSubnetVerifyDelete(context, d, connector)
}

// NOTE: This will not be needed when IPAM is handled by NSXT Policy
func resourceNsxtPolicyIPPoolBlockSubnetVerifyDelete(sessionContext utl.SessionContext, d *schema.ResourceData, connector client.Connector) error {

	client := realizedstate.NewRealizedEntitiesClient(sessionContext, connector)

	path := d.Get("path").(string)
	// Wait for realization state to disappear (not_found) - this means
	// block subnet deletion is realized
	pendingStates := []string{"PENDING"}
	targetStates := []string{"DELETED", "ERROR"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {

			_, realizationError := client.List(path, nil)
			if realizationError != nil {
				if isNotFoundError(realizationError) {
					return 0, "DELETED", nil
				}
				return 0, "ERROR", realizationError
			}
			// realization info found
			log.Printf("[INFO] IP Block realization still present")
			return 0, "PENDING", nil
		},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to confirm delete realization for %s: %v", path, err)
	}

	return nil
}

func resourceNsxtPolicyIPPoolSubnetImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		poolPath, err := getParameterFromPolicyPath("", "/ip-subnets/", importID)
		if err != nil {
			return nil, err
		}
		d.Set("pool_path", poolPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	if len(s) != 2 {
		return nil, fmt.Errorf("Please provide <ip-pool-id>/<subnet-id> as an input")
	}

	poolID := s[0]
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return nil, err
	}
	client := infra.NewIpPoolsClient(context, connector)

	pool, err := client.Get(poolID)
	if err != nil {
		return nil, err
	}
	d.Set("pool_path", pool.Path)

	d.SetId(s[1])

	return []*schema.ResourceData{d}, nil
}
