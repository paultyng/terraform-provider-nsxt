// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/aaa"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var featurePermissionTypes = [](string){
	nsxModel.FeaturePermission_PERMISSION_CRUD,
	nsxModel.FeaturePermission_PERMISSION_READ,
	nsxModel.FeaturePermission_PERMISSION_EXECUTE,
}

func resourceNsxtPolicyUserManagementRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyUserManagementRoleCreate,
		Read:   resourceNsxtPolicyUserManagementRoleRead,
		Update: resourceNsxtPolicyUserManagementRoleUpdate,
		Delete: resourceNsxtPolicyUserManagementRoleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"role": {
				Type:        schema.TypeString,
				Description: "Short identifier for the role. Must be all lower case with no spaces",
				Required:    true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile(
						"^[_a-z0-9-]+$"),
					"Must be a valid role identifier matching: ^[_a-z0-9-]+$"),
			},
			"feature": {
				Type:        schema.TypeList,
				Description: "List of permissions for features",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"feature": {
							Type:        schema.TypeString,
							Description: "Feature Id",
							Required:    true,
						},
						"feature_description": {
							Type:        schema.TypeString,
							Description: "Feature Description",
							Computed:    true,
						},
						"feature_name": {
							Type:        schema.TypeString,
							Description: "Feature Name",
							Computed:    true,
						},
						"permission": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(featurePermissionTypes, false),
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyUserManagementRoleExists(id string, connector client.Connector) (bool, error) {
	var err error
	roleClient := aaa.NewRolesClient(connector)
	_, err = roleClient.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func getFeaturePermissionFromSchema(d *schema.ResourceData) []nsxModel.FeaturePermission {
	features := d.Get("feature").([]interface{})
	featurePermissions := make([]nsxModel.FeaturePermission, 0, len(features))
	for _, feature := range features {
		data := feature.(map[string]interface{})
		featureID := data["feature"].(string)
		permission := data["permission"].(string)
		featurePermissions = append(featurePermissions, nsxModel.FeaturePermission{
			Feature:    &featureID,
			Permission: &permission,
		})
	}

	return featurePermissions
}

func setFeaturePermissionInSchema(d *schema.ResourceData, permissions []nsxModel.FeaturePermission) {
	var permissionList []map[string]interface{}
	for _, permission := range permissions {
		// NSX returns the complete feature set on GET
		// Bypass ineffective items from the list
		if permission.IsInternal != nil && *permission.IsInternal {
			// Internal permissions is not effective for roles
			continue
		}
		if permission.Permission != nil && *permission.Permission == nsxModel.FeaturePermission_PERMISSION_NONE {
			// Permissions are none by default. None-valued permissions are not allowed either in schema.
			continue
		}

		elem := make(map[string]interface{})
		elem["feature"] = permission.Feature
		elem["feature_description"] = permission.FeatureDescription
		elem["feature_name"] = permission.FeatureName
		elem["permission"] = permission.Permission
		permissionList = append(permissionList, elem)
	}
	err := d.Set("feature", permissionList)
	if err != nil {
		log.Printf("[WARNING] Failed to set feature in schema: %v", err)
	}
}

func policyUserManagementRoleUpdate(id string, d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	role := d.Get("role").(string)
	features := getFeaturePermissionFromSchema(d)
	revision := int64(d.Get("revision").(int))

	// Validate feature set first
	validateObj := nsxModel.FeaturePermissionArray{
		FeaturePermissions: features,
	}
	roleClient := aaa.NewRolesClient(connector)
	reco, err := roleClient.Validate(validateObj)
	if err != nil {
		log.Printf("[ERROR] Error validating RoleWithFeatures with ID %s", id)
		return err
	}
	if len(reco.Results) > 0 {
		missingPerm := make([]string, 0, len(reco.Results))
		for _, permission := range reco.Results {
			missingPerm = append(missingPerm, *permission.TargetFeature)
			log.Printf("[ERROR] RoleWithFeatures ID %s requires %s set to %v for %v",
				id, *permission.TargetFeature, permission.RecommendedPermissions, permission.SrcFeatures)
		}
		return fmt.Errorf("RoleWithFeatures ID %s missing dependent permissions %v",
			id, missingPerm)
	}

	obj := nsxModel.RoleWithFeatures{
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		Role:        &role,
		Features:    features,
		Revision:    &revision,
	}
	_, err = roleClient.Update(id, obj)
	return err
}

func resourceNsxtPolicyUserManagementRoleCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("role").(string)
	exists, err := resourceNsxtPolicyUserManagementRoleExists(id, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("resource with ID %s already exists", id)
	}

	// Create the resource using PUT
	log.Printf("[INFO] Creating RoleWithFeatures with ID %s", id)
	err = policyUserManagementRoleUpdate(id, d, m)
	if err != nil {
		return handleCreateError("RoleWithFeatures", id, err)
	}

	d.SetId(id)

	return resourceNsxtPolicyUserManagementRoleRead(d, m)
}

func resourceNsxtPolicyUserManagementRoleRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	roleClient := aaa.NewRolesClient(connector)

	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleWithFeatures ID")
		return err
	}

	obj, err := roleClient.Get(id)
	if err != nil {
		return handleReadError(d, "RoleWithFeatures", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("revision", obj.Revision)
	d.Set("role", obj.Role)
	setFeaturePermissionInSchema(d, obj.Features)

	return nil
}

func resourceNsxtPolicyUserManagementRoleUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleWithFeatures ID")
		return err
	}

	log.Printf("[INFO] Updateing RoleWithFeatures with ID %s", id)
	err := policyUserManagementRoleUpdate(id, d, m)
	if err != nil {
		return handleUpdateError("RoleWithFeatures", id, err)
	}

	return resourceNsxtPolicyUserManagementRoleRead(d, m)
}

func resourceNsxtPolicyUserManagementRoleDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	roleClient := aaa.NewRolesClient(connector)

	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleWithFeatures ID")
		return err
	}

	log.Printf("[INFO] Deleting RoleWithFeatures with ID %s", id)
	err := roleClient.Delete(id)
	if err != nil {
		return handleDeleteError("RoleWithFeatures", id, err)
	}

	return nil
}
