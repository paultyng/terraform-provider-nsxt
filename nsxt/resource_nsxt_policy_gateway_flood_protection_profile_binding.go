// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tier0s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_0s"
	t0localeservices "github.com/vmware/terraform-provider-nsxt/api/infra/tier_0s/locale_services"
	tier1s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s"
	t1localeservices "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s/locale_services"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyGatewayFloodProtectionProfileBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayFloodProtectionProfileBindingCreate,
		Read:   resourceNsxtPolicyGatewayFloodProtectionProfileBindingRead,
		Update: resourceNsxtPolicyGatewayFloodProtectionProfileBindingUpdate,
		Delete: resourceNsxtPolicyGatewayFloodProtectionProfileBindingDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtGatewayFloodProtectionProfileBindingImporter,
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
			"parent_path": {
				Type:         schema.TypeString,
				Description:  "The path of the parent to be bind with the profile. It could be either Tier0 path, Tier1 path, or locale service path",
				Required:     true,
				ValidateFunc: validatePolicyPath(),
			},
		},
	}
}

func extractGatewayIDLocaleServiceID(parentPath string) (string, string, string, error) {
	var tier0ID, tier1ID, localeServiceID string
	// Example:
	// 1: /infra/tier-0s/{tier0-id}
	// 2: /infra/tier-0s/{tier0-id}/locale-services/{locale-services-id}
	parentPathList := strings.Split(parentPath, "/")
	for i := 0; i < len(parentPathList)-1; i++ {
		if parentPathList[i] == "tier-0s" {
			tier0ID = parentPathList[i+1]
		} else if parentPathList[i] == "tier-1s" {
			tier1ID = parentPathList[i+1]
		} else if parentPathList[i] == "locale-services" {
			localeServiceID = parentPathList[i+1]
		}
	}

	if tier0ID == "" && tier1ID == "" {
		return "", "", "", fmt.Errorf("invalid parent_path: %s", parentPath)
	}
	return tier0ID, tier1ID, localeServiceID, nil
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingPatch(d *schema.ResourceData, m interface{}, parentPath string, id string, isCreate bool) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	profilePath := d.Get("profile_path").(string)
	obj := model.FloodProtectionProfileBindingMap{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		ProfilePath: &profilePath,
	}

	if !isCreate {
		revision := int64(d.Get("revision").(int))
		obj.Revision = &revision
	}

	tier0ID, tier1ID, localeServiceID, err := extractGatewayIDLocaleServiceID(parentPath)
	if err != nil {
		return err
	}
	if tier0ID != "" {
		if localeServiceID == "" {
			bindingClient := tier0s.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Patch(tier0ID, id, obj)
		} else {
			bindingClient := t0localeservices.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Patch(tier0ID, localeServiceID, id, obj)
		}
		if err != nil {
			return err
		}
	} else if tier1ID != "" {
		if localeServiceID == "" {
			bindingClient := tier1s.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Patch(tier1ID, id, obj)
		} else {
			bindingClient := t1localeservices.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Patch(tier1ID, localeServiceID, id, obj)
		}
	}
	return err
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingGet(sessionContext utl.SessionContext, connector client.Connector, parentPath, id string) (model.FloodProtectionProfileBindingMap, error) {
	var binding model.FloodProtectionProfileBindingMap
	tier0ID, tier1ID, localeServiceID, err := extractGatewayIDLocaleServiceID(parentPath)
	if err != nil {
		return binding, err
	}

	if tier0ID != "" {
		if localeServiceID == "" {
			bindingClient := tier0s.NewFloodProtectionProfileBindingsClient(sessionContext, connector)
			if bindingClient == nil {
				return binding, policyResourceNotSupportedError()
			}
			binding, err = bindingClient.Get(tier0ID, id)
		} else {
			bindingClient := t0localeservices.NewFloodProtectionProfileBindingsClient(sessionContext, connector)
			if bindingClient == nil {
				return binding, policyResourceNotSupportedError()
			}
			binding, err = bindingClient.Get(tier0ID, localeServiceID, id)
		}
	} else if tier1ID != "" {
		if localeServiceID == "" {
			bindingClient := tier1s.NewFloodProtectionProfileBindingsClient(sessionContext, connector)
			if bindingClient == nil {
				return binding, policyResourceNotSupportedError()
			}
			binding, err = bindingClient.Get(tier1ID, id)
		} else {
			bindingClient := t1localeservices.NewFloodProtectionProfileBindingsClient(sessionContext, connector)
			if bindingClient == nil {
				return binding, policyResourceNotSupportedError()
			}
			binding, err = bindingClient.Get(tier1ID, localeServiceID, id)
		}
	}
	return binding, err
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingExists(sessionContext utl.SessionContext, connector client.Connector, parentPath, id string) (bool, error) {
	_, err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingGet(sessionContext, connector, parentPath, id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingCreate(d *schema.ResourceData, m interface{}) error {

	var err error
	// While binding flood protection profile with gateway, the only supported profile binding id value is 'default'.
	// API response: Invalid profile binding id c1cf7a71-7549-45d5-ba9f-19b7979e2620, only supported value is 'default'. (code 523203)
	id := "default"
	parentPath := d.Get("parent_path").(string)

	exist, err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingExists(getSessionContext(d, m), getPolicyConnector(m), parentPath, id)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("Resource with id %s already exists", id)
	}

	err = resourceNsxtPolicyGatewayFloodProtectionProfileBindingPatch(d, m, parentPath, id, true)
	if err != nil {
		return handleCreateError("GatewayFloodProtectionProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyGatewayFloodProtectionProfileBindingRead(d, m)
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayFloodProtectionProfile ID")
	}

	parentPath := d.Get("parent_path").(string)
	binding, err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingGet(getSessionContext(d, m), connector, parentPath, id)
	if err != nil {
		return handleReadError(d, "GatewayFloodProtectionProfileBinding", id, err)
	}

	floodProtectionProfileBindingModelToSchema(d, *binding.DisplayName, *binding.Description, id, *binding.Path, *binding.ProfilePath, -1, binding.Tags, *binding.Revision)
	return nil
}

func floodProtectionProfileBindingModelToSchema(d *schema.ResourceData, displayName, description, nsxID, path, profilePath string, seqNum int64, tags []model.Tag, revision int64) {
	d.Set("display_name", displayName)
	d.Set("description", description)
	setPolicyTagsInSchema(d, tags)
	d.Set("nsx_id", nsxID)
	d.Set("path", path)
	d.Set("revision", revision)

	d.Set("profile_path", profilePath)
	if seqNum != -1 {
		d.Set("sequence_number", seqNum)
	}
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayFloodProtectionProfileBinding ID")
	}

	parentPath := d.Get("parent_path").(string)

	err := resourceNsxtPolicyGatewayFloodProtectionProfileBindingPatch(d, m, parentPath, id, false)
	if err != nil {
		return handleUpdateError("GatewayFloodProtectionProfileBinding", id, err)
	}

	return resourceNsxtPolicyGatewayFloodProtectionProfileBindingRead(d, m)
}

func resourceNsxtPolicyGatewayFloodProtectionProfileBindingDelete(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining GatewayFloodProtectionProfileBinding ID")
	}

	connector := getPolicyConnector(m)
	parentPath := d.Get("parent_path").(string)
	tier0ID, tier1ID, localeServiceID, err := extractGatewayIDLocaleServiceID(parentPath)
	if err != nil {
		return err
	}

	if tier0ID != "" {
		if localeServiceID == "" {
			bindingClient := tier0s.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Delete(tier0ID, id)
		} else {
			bindingClient := t0localeservices.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Delete(tier0ID, localeServiceID, id)
		}
		if err != nil {
			return err
		}
	} else if tier1ID != "" {
		if localeServiceID == "" {
			bindingClient := tier1s.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Delete(tier1ID, id)
		} else {
			bindingClient := t1localeservices.NewFloodProtectionProfileBindingsClient(getSessionContext(d, m), connector)
			if bindingClient == nil {
				return policyResourceNotSupportedError()
			}
			err = bindingClient.Delete(tier1ID, localeServiceID, id)
		}
	}

	if err != nil {
		return handleDeleteError("GatewayFloodProtectionProfileBinding", id, err)
	}
	return nil
}

func nsxtGatewayFloodProtectionProfileBindingImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	_, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return nil, err
	}

	targetSection := "/flood-protection-profile-bindings/"
	splitIdx := strings.LastIndex(importID, targetSection)
	if splitIdx == -1 {
		return nil, fmt.Errorf("invalid importID for GatewayFloodProtectionProfileBinding: %s", importID)
	}
	parentPath := importID[:splitIdx]
	id := importID[splitIdx+len(targetSection):]
	d.Set("parent_path", parentPath)
	d.SetId(id)
	return []*schema.ResourceData{d}, nil
}
