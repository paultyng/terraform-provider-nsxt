/* Copyright © 2024 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-nsxt/api/infra/domains/groups"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyDistributedFloodProtectionProfileBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDistributedFloodProtectionProfileBindingCreate,
		Read:   resourceNsxtPolicyDistributedFloodProtectionProfileBindingRead,
		Update: resourceNsxtPolicyDistributedFloodProtectionProfileBindingUpdate,
		Delete: resourceNsxtPolicyDistributedFloodProtectionProfileBindingDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtDistributedFloodProtectionProfileBindingImporter,
		},
		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"profile_path": {
				Type:         schema.TypeString,
				Description:  "The path of the flood protection profile",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"group_path": {
				Type:         schema.TypeString,
				Description:  "The path of the group to bind with the flood protection profile",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePolicyPath(),
			},
			"sequence_number": {
				Type:        schema.TypeInt,
				Description: "Sequence number of this profile binding",
				Required:    true,
			},
		},
	}
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingPatch(d *schema.ResourceData, m interface{}, id string, isCreate bool) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	bindingClient := groups.NewFirewallFloodProtectionProfileBindingMapsClient(context, connector)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	profilePath := d.Get("profile_path").(string)
	seqNum := int64(d.Get("sequence_number").(int))
	obj := model.PolicyFirewallFloodProtectionProfileBindingMap{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		ProfilePath:    &profilePath,
		SequenceNumber: &seqNum,
	}

	groupPath := d.Get("group_path").(string)
	groupID := getPolicyIDFromPath(groupPath)
	domain := getDomainFromResourcePath(groupPath)

	if !isCreate {
		revision := int64(d.Get("revision").(int))
		obj.Revision = &revision
	}
	return bindingClient.Patch(domain, groupID, id, obj)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingExists(sessionContext utl.SessionContext, connector client.Connector, groupPath, id string) (bool, error) {
	bindingClient := groups.NewFirewallFloodProtectionProfileBindingMapsClient(sessionContext, connector)
	domain := getDomainFromResourcePath(groupPath)
	groupID := getPolicyIDFromPath(groupPath)
	_, err := bindingClient.Get(domain, groupID, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingCreate(d *schema.ResourceData, m interface{}) error {
	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}

	groupPath := d.Get("group_path").(string)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	exist, err := resourceNsxtPolicyDistributedFloodProtectionProfileBindingExists(context, getPolicyConnector(m), groupPath, id)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("Resource with id %s already exists", id)
	}

	err = resourceNsxtPolicyDistributedFloodProtectionProfileBindingPatch(d, m, id, true)
	if err != nil {
		return handleCreateError("DistributedFloodProtectionProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDistributedFloodProtectionProfileBindingRead(d, m)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining FloodProtectionProfile ID")
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	bindingClient := groups.NewFirewallFloodProtectionProfileBindingMapsClient(context, connector)

	groupPath := d.Get("group_path").(string)
	domain := getDomainFromResourcePath(groupPath)
	groupID := getPolicyIDFromPath(groupPath)

	binding, err := bindingClient.Get(domain, groupID, id)
	if err != nil {
		return handleReadError(d, "FloodProtectionProfileBinding", id, err)
	}

	floodProtectionProfileBindingModelToSchema(d, *binding.DisplayName, *binding.Description, id, *binding.Path, *binding.ProfilePath, *binding.SequenceNumber, binding.Tags, *binding.Revision)

	return nil
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DistributedFloodProtectionProfileBinding ID")
	}

	err := resourceNsxtPolicyDistributedFloodProtectionProfileBindingPatch(d, m, id, false)
	if err != nil {
		return handleUpdateError("DistributedFloodProtectionProfileBinding", id, err)
	}

	return resourceNsxtPolicyDistributedFloodProtectionProfileBindingRead(d, m)
}

func resourceNsxtPolicyDistributedFloodProtectionProfileBindingDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining DistributedFloodProtectionProfileBinding ID")
	}

	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	bindingClient := groups.NewFirewallFloodProtectionProfileBindingMapsClient(context, connector)

	groupPath := d.Get("group_path").(string)
	domain := getDomainFromResourcePath(groupPath)
	groupID := getPolicyIDFromPath(groupPath)

	err = bindingClient.Delete(domain, groupID, id)
	if err != nil {
		return handleDeleteError("FloodProtectionProfileBinding", id, err)
	}
	return nil
}

func nsxtDistributedFloodProtectionProfileBindingImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	_, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return nil, err
	}
	targetSection := "/firewall-flood-protection-profile-binding-maps/"
	splitIdx := strings.LastIndex(importID, targetSection)
	if splitIdx == -1 {
		return nil, fmt.Errorf("invalid importID for DistributedFloodProtectionProfileBinding: %s", importID)
	}
	parentPath := importID[:splitIdx]
	id := importID[splitIdx+len(targetSection):]
	d.Set("group_path", parentPath)
	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}
