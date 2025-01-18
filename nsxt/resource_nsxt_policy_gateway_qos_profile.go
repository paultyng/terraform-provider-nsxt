// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var gatewayQosProfileExcessActionValues = []string{
	model.GatewayQosProfile_EXCESS_ACTION_DROP,
}

func resourceNsxtPolicyGatewayQosProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayQosProfileCreate,
		Read:   resourceNsxtPolicyGatewayQosProfileRead,
		Update: resourceNsxtPolicyGatewayQosProfileUpdate,
		Delete: resourceNsxtPolicyGatewayQosProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"burst_size": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
			},
			"committed_bandwidth": {
				Type:     schema.TypeInt,
				Default:  1,
				Optional: true,
			},
			"excess_action": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(gatewayQosProfileExcessActionValues, false),
				Optional:     true,
				Default:      model.GatewayQosProfile_EXCESS_ACTION_DROP,
			},
		},
	}
}

func resourceNsxtPolicyGatewayQosProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	if isGlobalManager {
		client := gm_infra.NewGatewayQosProfilesClient(connector)
		_, err = client.Get(id)
	} else {
		client := infra.NewGatewayQosProfilesClient(connector)
		_, err = client.Get(id)
	}
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyGatewayQosProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	burstSize := int64(d.Get("burst_size").(int))
	committedBandwidth := int64(d.Get("committed_bandwidth").(int))
	excessAction := d.Get("excess_action").(string)

	// Note - we also need to specify deprecated property due to NSX bug
	obj := model.GatewayQosProfile{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		BurstSize:           &burstSize,
		CommittedBandwitdth: &committedBandwidth,
		CommittedBandwidth:  &committedBandwidth,
		ExcessAction:        &excessAction,
	}

	log.Printf("[INFO] Patching GatewayQosProfile with ID %s", id)
	if isPolicyGlobalManager(m) {
		gmObj, convErr := convertModelBindingType(obj, model.GatewayQosProfileBindingType(), gm_model.GatewayQosProfileBindingType())
		if convErr != nil {
			return convErr
		}
		client := gm_infra.NewGatewayQosProfilesClient(connector)
		return client.Patch(id, gmObj.(gm_model.GatewayQosProfile), nil)
	}

	client := infra.NewGatewayQosProfilesClient(connector)
	return client.Patch(id, obj, nil)
}

func resourceNsxtPolicyGatewayQosProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyGatewayQosProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyGatewayQosProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("GatewayQosProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayQosProfileRead(d, m)
}

func resourceNsxtPolicyGatewayQosProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayQosProfile ID")
	}

	var obj model.GatewayQosProfile
	if isPolicyGlobalManager(m) {
		client := gm_infra.NewGatewayQosProfilesClient(connector)
		gmObj, err := client.Get(id)
		if err != nil {
			return handleReadError(d, "GatewayQosProfile", id, err)
		}

		lmObj, err := convertModelBindingType(gmObj, gm_model.GatewayQosProfileBindingType(), model.GatewayQosProfileBindingType())
		if err != nil {
			return err
		}
		obj = lmObj.(model.GatewayQosProfile)
	} else {
		client := infra.NewGatewayQosProfilesClient(connector)
		var err error
		obj, err = client.Get(id)
		if err != nil {
			return handleReadError(d, "GatewayQosProfile", id, err)
		}
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("burst_size", obj.BurstSize)
	d.Set("committed_bandwidth", obj.CommittedBandwidth)
	d.Set("excess_action", obj.ExcessAction)

	return nil
}

func resourceNsxtPolicyGatewayQosProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayQosProfile ID")
	}

	err := resourceNsxtPolicyGatewayQosProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("GatewayQosProfile", id, err)
	}

	return resourceNsxtPolicyGatewayQosProfileRead(d, m)
}

func resourceNsxtPolicyGatewayQosProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayQosProfile ID")
	}

	connector := getPolicyConnector(m)
	var err error
	if isPolicyGlobalManager(m) {
		client := gm_infra.NewGatewayQosProfilesClient(connector)
		err = client.Delete(id, nil)
	} else {
		client := infra.NewGatewayQosProfilesClient(connector)
		err = client.Delete(id, nil)
	}

	if err != nil {
		return handleDeleteError("GatewayQosProfile", id, err)
	}

	return nil
}
