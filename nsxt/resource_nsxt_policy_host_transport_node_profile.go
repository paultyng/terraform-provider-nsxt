// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyHostTransportNodeProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyHostTransportNodeProfileCreate,
		Read:   resourceNsxtPolicyHostTransportNodeProfileRead,
		Update: resourceNsxtPolicyHostTransportNodeProfileUpdate,
		Delete: resourceNsxtPolicyHostTransportNodeProfileDelete,
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
			// host_switch_spec
			"standard_host_switch": getStandardHostSwitchSchema(nodeTypeHost),
			"ignore_overridden_hosts": {
				Type:        schema.TypeBool,
				Default:     false,
				Description: "Determines if cluster-level configuration should be applied on overridden hosts",
				Optional:    true,
			},
		},
	}
}

func resourceNsxtPolicyHostTransportNodeProfileExists(id string, connector client.Connector, isGlobal bool) (bool, error) {
	client := infra.NewHostTransportNodeProfilesClient(connector)

	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving IP Block", err)
}

func resourceNsxtPolicyHostTransportNodeProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewHostTransportNodeProfilesClient(connector)

	id, err := getOrGenerateID(d, m, resourceNsxtPolicyHostTransportNodeProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	ignoreOverridenHosts := d.Get("ignore_overridden_hosts").(bool)

	hostSwitchSpec, err := getHostSwitchSpecFromSchema(d, m, nodeTypeHost)
	if err != nil {
		return err
	}

	obj := model.PolicyHostTransportNodeProfile{
		DisplayName:           &displayName,
		Description:           &description,
		Tags:                  tags,
		HostSwitchSpec:        hostSwitchSpec,
		IgnoreOverriddenHosts: &ignoreOverridenHosts,
	}

	_, err = client.Update(id, obj, nil)
	if err != nil {
		return handleCreateError("Policy Host Transport Node Profile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyHostTransportNodeProfileRead(d, m)
}

func resourceNsxtPolicyHostTransportNodeProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewHostTransportNodeProfilesClient(connector)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Policy Host Transport Node Profile ID")
	}

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Policy Host Transport Node Profile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", obj.Id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	err = setHostSwitchSpecInSchema(d, obj.HostSwitchSpec, nodeTypeHost)
	if err != nil {
		return err
	}

	return nil
}

func resourceNsxtPolicyHostTransportNodeProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewHostTransportNodeProfilesClient(connector)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Policy Host Transport Node Profile ID")
	}

	// Read the rest of the configured parameters
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	revision := int64(d.Get("revision").(int))
	tags := getPolicyTagsFromSchema(d)
	ignoreOverridenHosts := d.Get("ignore_overridden_hosts").(bool)

	hostSwitchSpec, err := getHostSwitchSpecFromSchema(d, m, nodeTypeHost)
	if err != nil {
		return err
	}

	obj := model.PolicyHostTransportNodeProfile{
		DisplayName:           &displayName,
		Description:           &description,
		Tags:                  tags,
		HostSwitchSpec:        hostSwitchSpec,
		IgnoreOverriddenHosts: &ignoreOverridenHosts,
		Revision:              &revision,
	}

	_, err = client.Update(id, obj, nil)
	if err != nil {
		return handleUpdateError("Policy Host Transport Node Profile", id, err)
	}

	return resourceNsxtPolicyHostTransportNodeProfileRead(d, m)
}

func resourceNsxtPolicyHostTransportNodeProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Policy Host Transport Node Profile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewHostTransportNodeProfilesClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("Policy Host Transport Node Profile", id, err)
	}

	return nil
}
