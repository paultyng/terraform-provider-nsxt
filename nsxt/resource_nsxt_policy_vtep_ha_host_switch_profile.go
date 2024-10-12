// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtVtepHAHostSwitchProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtVtepHAHostSwitchProfileCreate,
		Read:   resourceNsxtVtepHAHostSwitchProfileRead,
		Update: resourceNsxtVtepHAHostSwitchProfileUpdate,
		Delete: resourceNsxtVtepHAHostSwitchProfileDelete,
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
			"auto_recovery": {
				Type:        schema.TypeBool,
				Description: "Enabled status of autonomous recovery option",
				Optional:    true,
				Default:     true,
			},
			"auto_recovery_initial_wait": {
				Type:         schema.TypeInt,
				Description:  "Start time of autonomous recovery (in seconds)",
				Optional:     true,
				Default:      300,
				ValidateFunc: validation.IntBetween(300, 3600),
			},
			"auto_recovery_max_backoff": {
				Type:         schema.TypeInt,
				Description:  "Maximum backoff time for autonomous recovery (in seconds)",
				Optional:     true,
				Default:      86400,
				ValidateFunc: validation.IntBetween(3600, 86400),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enabled status of VTEP High Availability feature",
				Optional:    true,
				Default:     false,
			},
			"failover_timeout": {
				Type:         schema.TypeInt,
				Description:  "VTEP High Availability failover timeout (in seconds)",
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntBetween(0, 60),
			},
			"realized_id": {
				Type:        schema.TypeString,
				Description: "Computed ID of the realized object",
				Computed:    true,
			},
		},
	}
}

func resourceNsxtVtepHAHostSwitchProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewHostSwitchProfilesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func vtepHAHostSwitchProfileSchemaToModel(d *schema.ResourceData) model.PolicyVtepHAHostSwitchProfile {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	autoRecovery := d.Get("auto_recovery").(bool)
	autoRecoveryInitialWait := int64(d.Get("auto_recovery_initial_wait").(int))
	autoRecoveryMaxBackoff := int64(d.Get("auto_recovery_max_backoff").(int))
	enabled := d.Get("enabled").(bool)
	failoverTimeout := int64(d.Get("failover_timeout").(int))

	vtepHAHostSwitchProfile := model.PolicyVtepHAHostSwitchProfile{
		DisplayName:             &displayName,
		Description:             &description,
		Tags:                    tags,
		ResourceType:            model.PolicyBaseHostSwitchProfile_RESOURCE_TYPE_POLICYVTEPHAHOSTSWITCHPROFILE,
		AutoRecovery:            &autoRecovery,
		AutoRecoveryInitialWait: &autoRecoveryInitialWait,
		AutoRecoveryMaxBackoff:  &autoRecoveryMaxBackoff,
		Enabled:                 &enabled,
		FailoverTimeout:         &failoverTimeout,
	}
	return vtepHAHostSwitchProfile
}

func resourceNsxtVtepHAHostSwitchProfileCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtVtepHAHostSwitchProfileExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	vtepHAHostSwitchProfile := vtepHAHostSwitchProfileSchemaToModel(d)
	profileValue, errs := converter.ConvertToVapi(vtepHAHostSwitchProfile, model.PolicyVtepHAHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)
	_, err = client.Patch(id, profileStruct)
	if err != nil {
		return handleCreateError("VtepHAHostSwitchProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtVtepHAHostSwitchProfileRead(d, m)
}

func resourceNsxtVtepHAHostSwitchProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VtepHAHostSwitchProfile ID")
	}
	client := infra.NewHostSwitchProfilesClient(connector)
	structValue, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "VtepHAHostSwitchProfile", id, err)
	}

	converter := bindings.NewTypeConverter()
	baseInterface, errs := converter.ConvertToGolang(structValue, model.PolicyBaseHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	base := baseInterface.(model.PolicyBaseHostSwitchProfile)

	resourceType := base.ResourceType
	if resourceType != infra.HostSwitchProfiles_LIST_HOSTSWITCH_PROFILE_TYPE_POLICYVTEPHAHOSTSWITCHPROFILE {
		return handleReadError(d, "VtepHAHostSwitchProfile", id, err)
	}

	vtepHAInterface, errs := converter.ConvertToGolang(structValue, model.PolicyVtepHAHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	vtepHAHostSwitchProfile := vtepHAInterface.(model.PolicyVtepHAHostSwitchProfile)
	d.Set("display_name", vtepHAHostSwitchProfile.DisplayName)
	d.Set("description", vtepHAHostSwitchProfile.Description)
	setPolicyTagsInSchema(d, vtepHAHostSwitchProfile.Tags)
	d.Set("nsx_id", id)
	d.Set("path", vtepHAHostSwitchProfile.Path)
	d.Set("revision", vtepHAHostSwitchProfile.Revision)
	d.Set("realized_id", vtepHAHostSwitchProfile.RealizationId)

	d.Set("auto_recovery", vtepHAHostSwitchProfile.AutoRecovery)
	d.Set("auto_recovery_initial_wait", vtepHAHostSwitchProfile.AutoRecoveryInitialWait)
	d.Set("auto_recovery_max_backoff", vtepHAHostSwitchProfile.AutoRecoveryMaxBackoff)
	d.Set("enabled", vtepHAHostSwitchProfile.Enabled)
	d.Set("failover_timeout", vtepHAHostSwitchProfile.FailoverTimeout)

	return nil
}

func resourceNsxtVtepHAHostSwitchProfileUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining VtepHAHostSwitchProfile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	converter := bindings.NewTypeConverter()

	vtepHAHostSwitchProfile := vtepHAHostSwitchProfileSchemaToModel(d)
	revision := int64(d.Get("revision").(int))
	vtepHAHostSwitchProfile.Revision = &revision
	profileValue, errs := converter.ConvertToVapi(vtepHAHostSwitchProfile, model.PolicyVtepHAHostSwitchProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	profileStruct := profileValue.(*data.StructValue)
	_, err := client.Update(id, profileStruct)
	if err != nil {
		return handleUpdateError("VtepHAHostSwitchProfile", id, err)
	}

	return resourceNsxtVtepHAHostSwitchProfileRead(d, m)
}

func resourceNsxtVtepHAHostSwitchProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining VtepHAHostSwitchProfile ID")
	}
	connector := getPolicyConnector(m)
	client := infra.NewHostSwitchProfilesClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("VtepHAHostSwitchProfile", id, err)
	}
	return nil
}
