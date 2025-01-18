// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
)

var globalManagerModeValues = []string{
	gm_model.GlobalManager_MODE_ACTIVE,
	gm_model.GlobalManager_MODE_STANDBY,
}

func resourceNsxtPolicyGlobalManager() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGlobalManagerCreate,
		Read:   resourceNsxtPolicyGlobalManagerRead,
		Update: resourceNsxtPolicyGlobalManagerUpdate,
		Delete: resourceNsxtPolicyGlobalManagerDelete,
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
			"fail_if_rtt_exceeded": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Fail onboarding if maximum RTT exceeded",
				Default:     true,
			},
			"maximum_rtt": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Maximum acceptable packet round trip time (RTT)",
				Default:      250,
				ValidateFunc: validation.IntBetween(0, 1000),
			},
			"mode": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Mode of the global manager",
				ValidateFunc: validation.StringInSlice(globalManagerModeValues, false),
			},
			"connection_info": getConnectionInfoSchema(),
		},
	}
}

func getGlobalManagerFromSchema(d *schema.ResourceData) gm_model.GlobalManager {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getGMTagsFromSchema(d)
	failIfRttExceeded := d.Get("fail_if_rtt_exceeded").(bool)
	maximumRtt := int64(d.Get("maximum_rtt").(int))
	connectionInfos := getConnectionInfosFromSchema(d, "connection_info")
	mode := d.Get("mode").(string)

	return gm_model.GlobalManager{
		DisplayName:       &displayName,
		Description:       &description,
		Tags:              tags,
		FailIfRttExceeded: &failIfRttExceeded,
		MaximumRtt:        &maximumRtt,
		ConnectionInfo:    connectionInfos,
		Mode:              &mode,
	}
}

func resourceNsxtPolicyGlobalManagerExists(id string, connector client.Connector, isGlobal bool) (bool, error) {
	client := global_infra.NewGlobalManagersClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyGlobalManagerCreate(d *schema.ResourceData, m interface{}) error {
	if !isPolicyGlobalManager(m) {
		return globalManagerOnlyError()
	}
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyGlobalManagerExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := global_infra.NewGlobalManagersClient(connector)
	gm := getGlobalManagerFromSchema(d)

	err = client.Patch(id, gm, nil)
	if err != nil {
		return handleCreateError("GlobalManager", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGlobalManagerRead(d, m)
}

func resourceNsxtPolicyGlobalManagerRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining GlobalManager ID")
	}
	client := global_infra.NewGlobalManagersClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "GlobalManager", id, err)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setGMTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("fail_if_rtt_exceeded", obj.FailIfRttExceeded)
	d.Set("maximum_rtt", obj.MaximumRtt)

	setConnectionInfosInSchema(d, obj.ConnectionInfo, "connection_info")
	d.Set("mode", obj.Mode)

	return nil
}

func resourceNsxtPolicyGlobalManagerUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining GlobalManager ID")
	}

	connector := getPolicyConnector(m)
	client := global_infra.NewGlobalManagersClient(connector)

	obj := getGlobalManagerFromSchema(d)
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision

	_, err := client.Update(id, obj, nil)
	if err != nil {
		return handleUpdateError("GlobalManager", id, err)
	}

	return nil
}

func resourceNsxtPolicyGlobalManagerDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining GlobalManager ID")
	}
	connector := getPolicyConnector(m)
	client := global_infra.NewGlobalManagersClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("GlobalManager", id, err)
	}

	return nil
}
