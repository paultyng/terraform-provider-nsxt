// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/aaa"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var roleBindingUserTypes = [](string){
	nsxModel.RoleBinding_TYPE_LOCAL_USER,
	nsxModel.RoleBinding_TYPE_REMOTE_USER,
	nsxModel.RoleBinding_TYPE_REMOTE_GROUP,
}

var roleBindingIdentitySourceTypes = [](string){
	nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_VIDM,
	nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_LDAP,
	nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_OIDC,
	nsxModel.RoleBinding_IDENTITY_SOURCE_TYPE_CSP,
}

var defaultRolePath = "/"
var defaultRoleName = "auditor"

func resourceNsxtPolicyUserManagementRoleBinding() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyUserManagementRoleBindingCreate,
		Read:   resourceNsxtPolicyUserManagementRoleBindingRead,
		Update: resourceNsxtPolicyUserManagementRoleBindingUpdate,
		Delete: resourceNsxtPolicyUserManagementRoleBindingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"display_name": getDataSourceDisplayNameSchema(),
			"description":  getDataSourceDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"name": {
				Type:        schema.TypeString,
				Description: "User/Group's name",
				Required:    true,
			},
			"type": {
				Type:         schema.TypeString,
				Description:  "Indicates the type of the user",
				Required:     true,
				ValidateFunc: validation.StringInSlice(roleBindingUserTypes, false),
			},
			"user_id": {
				Type:        schema.TypeString,
				Description: "Local user's numeric id",
				Computed:    true,
			},
			"identity_source_id": {
				Type:        schema.TypeString,
				Description: "ID of the external identity source",
				Optional:    true,
			},
			"identity_source_type": {
				Type:         schema.TypeString,
				Description:  "ID of the external identity source",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(roleBindingIdentitySourceTypes, false),
			},
			"roles_for_path": getRolesForPathSchema(false),
			"overwrite_local_user": {
				Type:        schema.TypeBool,
				Description: "Allow overwriting auto-created role binding on NSX for local users",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

// getRolesForPathSchema return schema for RolesForPath, which is shared between role bindings and PI
func getRolesForPathSchema(forceNew bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "List of roles that are associated with the user, limiting them to a path",
		Required:    true,
		ForceNew:    forceNew,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"path": {
					Type:        schema.TypeString,
					Description: "Path of the entity in parent hierarchy.",
					Required:    true,
					ForceNew:    forceNew,
				},
				"roles": {
					Type:        schema.TypeSet,
					Description: "Applicable roles",
					Required:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					ForceNew: forceNew,
				},
			},
		},
	}
}

type rolesForPath map[string]rolesPerPath
type rolesPerPath map[string]bool

func (r rolesPerPath) getAnyRole() *string {
	for k, v := range r {
		if v {
			return &k
		}
	}
	return nil
}

func getRolesForPathFromSchema(d *schema.ResourceData) rolesForPath {
	rolesForPathMap := make(rolesForPath)
	rolesForPathInput := d.Get("roles_for_path").(*schema.Set).List()
	for _, rolesPerPathInput := range rolesForPathInput {
		data := rolesPerPathInput.(map[string]interface{})
		path := data["path"].(string)
		roles := interface2StringList(data["roles"].(*schema.Set).List())
		rolesPerPathMap := make(rolesPerPath)
		for _, role := range roles {
			rolesPerPathMap[role] = true
		}
		rolesForPathMap[path] = rolesPerPathMap
	}

	return rolesForPathMap
}

func setRolesForPathInSchema(d *schema.ResourceData, nsxRolesForPathList []nsxModel.RolesForPath) {
	var rolesForPathList []map[string]interface{}
	for _, nsxRolesForPath := range nsxRolesForPathList {
		elem := make(map[string]interface{})
		elem["path"] = *nsxRolesForPath.Path
		roles := make([]string, 0, len(nsxRolesForPath.Roles))
		for _, nsxRole := range nsxRolesForPath.Roles {
			roles = append(roles, *nsxRole.Role)
		}
		elem["roles"] = roles
		rolesForPathList = append(rolesForPathList, elem)
	}
	err := d.Set("roles_for_path", rolesForPathList)
	if err != nil {
		log.Printf("[WARNING] Failed to set roles_for_path in schema: %v", err)
	}
}

func getRolesForPathList(d *schema.ResourceData, rolesToRemove rolesForPath) []nsxModel.RolesForPath {
	boolTrue := true
	rolesPerPathMap := getRolesForPathFromSchema(d)
	nsxRolesForPaths := make([]nsxModel.RolesForPath, 0)

	for k, v := range rolesPerPathMap {
		path := k
		nsxRoles := make([]nsxModel.Role, 0, len(v))
		for k := range v {
			roleID := k
			nsxRoles = append(nsxRoles, nsxModel.Role{
				Role: &roleID,
			})
		}
		nsxRolesForPaths = append(nsxRolesForPaths, nsxModel.RolesForPath{
			Path:  &path,
			Roles: nsxRoles,
		})
	}

	// Remove roles explicitly marked for removal.
	// This is only expected for overwriting local users' role bindings
	pathRemoved := make(map[string]bool)
	for path, rolesMap := range rolesToRemove {
		if _, ok := rolesPerPathMap[path]; ok {
			// This path is also in TF definition
			// Roles in it will be overwritten
			continue
		}
		nsxRolesForPaths = append(nsxRolesForPaths, nsxModel.RolesForPath{
			Path:       &path,
			DeletePath: &boolTrue,
			Roles:      []nsxModel.Role{{Role: rolesMap.getAnyRole()}},
		})
		pathRemoved[path] = true
	}

	// Handle deletion of entire paths on change
	if d.HasChange("roles_for_path") {
		o, _ := d.GetChange("roles_for_path")
		oldRoles := o.(*schema.Set).List()
		for _, oldRole := range oldRoles {
			data := oldRole.(map[string]interface{})
			path := data["path"].(string)
			if _, ok := rolesPerPathMap[path]; ok {
				continue
			}
			if _, ok := pathRemoved[path]; ok {
				continue
			}
			// Add one role in the list to make NSX happy
			roles := interface2StringList(data["roles"].(*schema.Set).List())
			if len(roles) == 0 {
				continue
			}
			nsxRolesForPaths = append(nsxRolesForPaths, nsxModel.RolesForPath{
				Path:       &path,
				DeletePath: &boolTrue,
				Roles:      []nsxModel.Role{{Role: &roles[0]}},
			})
			pathRemoved[path] = true
		}
	}
	return nsxRolesForPaths
}

func getRoleBindingObject(d *schema.ResourceData, removeRoles rolesForPath) *nsxModel.RoleBinding {
	boolTrue := true
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	name := d.Get("name").(string)
	identitySrcID := d.Get("identity_source_id").(string)
	identitySrcType := d.Get("identity_source_type").(string)
	roleBindingType := d.Get("type").(string)
	nsxRolesForPaths := getRolesForPathList(d, removeRoles)

	obj := nsxModel.RoleBinding{
		DisplayName:       &displayName,
		Description:       &description,
		Tags:              tags,
		Name:              &name,
		Type_:             &roleBindingType,
		ReadRolesForPaths: &boolTrue,
		RolesForPaths:     nsxRolesForPaths,
	}
	if len(identitySrcID) > 0 {
		obj.IdentitySourceId = &identitySrcID
	}
	if len(identitySrcType) > 0 {
		obj.IdentitySourceType = &identitySrcType
	}
	return &obj
}

func getExistingRoleBinding(rbClient aaa.RoleBindingsClient, username, userType string) (*nsxModel.RoleBinding, error) {
	// Pagination should be unnecessary since filtered with name
	rbList, err := rbClient.List(nil, nil, nil, nil, &username, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list RoleBinding for %s", username)
	}

	var rbObj *nsxModel.RoleBinding
	for i, roleBinding := range rbList.Results {
		if *roleBinding.Name == username && *roleBinding.Type_ == userType {
			rbObj = &rbList.Results[i]
			break
		}
	}
	if rbObj == nil {
		return nil, fmt.Errorf("failed to match RoleBinding id for %s", username)
	}
	return rbObj, nil
}

func overwriteRoleBinding(d *schema.ResourceData, m interface{}, existingRoleBinding *nsxModel.RoleBinding) error {
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)
	id := d.Id()

	existingRoles := make(rolesForPath)
	for _, roles := range existingRoleBinding.RolesForPaths {
		if len(roles.Roles) == 0 {
			continue
		}
		existingRoles[*roles.Path] = make(rolesPerPath)
		existingRoles[*roles.Path][*roles.Roles[0].Role] = true
	}

	log.Printf("[INFO] Overwriting RoleBinding with ID %s", id)
	obj := getRoleBindingObject(d, existingRoles)
	_, err := rbClient.Update(id, *obj)
	if err != nil {
		return handleUpdateError("RoleBinding", id, err)
	}

	return resourceNsxtPolicyUserManagementRoleBindingRead(d, m)
}

func revertRoleBinding(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)

	id := d.Id()
	boolTrue := true
	currRoles := getRolesForPathFromSchema(d)
	nsxRolesForPaths := make([]nsxModel.RolesForPath, 0)
	for path, roleMap := range currRoles {
		if path == defaultRolePath {
			continue
		}
		rPath := strings.Clone(path)
		nsxRolesForPaths = append(nsxRolesForPaths, nsxModel.RolesForPath{
			Path:       &rPath,
			DeletePath: &boolTrue,
			Roles:      []nsxModel.Role{{Role: roleMap.getAnyRole()}},
		})
	}

	// Add the auditor role
	nsxRolesForPaths = append(nsxRolesForPaths, nsxModel.RolesForPath{
		Path:  &defaultRolePath,
		Roles: []nsxModel.Role{{Role: &defaultRoleName}},
	})
	name := d.Get("name").(string)
	roleBindingType := d.Get("type").(string)
	obj := nsxModel.RoleBinding{
		Name:              &name,
		Type_:             &roleBindingType,
		RolesForPaths:     nsxRolesForPaths,
		ReadRolesForPaths: &boolTrue,
	}

	log.Printf("[INFO] Reverting RoleBinding ID %s to %s", id, defaultRoleName)
	if _, err := rbClient.Update(id, obj); err != nil {
		return handleUpdateError("RoleBinding", id, err)
	}
	return nil
}

func resourceNsxtPolicyUserManagementRoleBindingCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)

	roleBindingType := d.Get("type").(string)
	overwriteLocaluser := d.Get("overwrite_local_user").(bool)
	username := d.Get("name").(string)
	if roleBindingType == nsxModel.RoleBinding_TYPE_LOCAL_USER {
		if !overwriteLocaluser {
			return fmt.Errorf(
				"RoleBinding for %s owned by NSX. Import the binding, or set overwrite_local_user",
				roleBindingType)
		}
		rbObj, err := getExistingRoleBinding(rbClient, username, nsxModel.RoleBinding_TYPE_LOCAL_USER)
		if err != nil {
			return err
		}
		d.SetId(*rbObj.Id)
		return overwriteRoleBinding(d, m, rbObj)
	}

	// Create the resource using POST
	log.Printf("[INFO] Creating RoleBinding for %s %s", roleBindingType, username)
	obj := getRoleBindingObject(d, rolesForPath{})

	rbObj, err := rbClient.Create(*obj)
	if err != nil {
		return handleCreateError("RoleBinding", username, err)
	}

	id := *rbObj.Id
	d.SetId(id)

	return resourceNsxtPolicyUserManagementRoleBindingRead(d, m)
}

func resourceNsxtPolicyUserManagementRoleBindingRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)

	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleBinding ID")
		return err
	}

	obj, err := rbClient.Get(id, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleReadError(d, "RoleBinding", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("revision", obj.Revision)
	d.Set("name", obj.Name)
	d.Set("type", obj.Type_)
	if obj.IdentitySourceId != nil {
		d.Set("identity_source_id", obj.IdentitySourceId)
	}
	if obj.IdentitySourceType != nil {
		d.Set("identity_source_type", obj.IdentitySourceType)
	}
	if obj.UserId != nil {
		d.Set("user_id", obj.UserId)
	}
	setRolesForPathInSchema(d, obj.RolesForPaths)

	return nil
}

func resourceNsxtPolicyUserManagementRoleBindingUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleBinding ID")
		return err
	}

	log.Printf("[INFO] Updateing RoleBinding with ID %s", id)
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)
	obj := getRoleBindingObject(d, rolesForPath{})
	_, err := rbClient.Update(id, *obj)
	if err != nil {
		return handleCreateError("RoleBinding", id, err)
	}

	return resourceNsxtPolicyUserManagementRoleBindingRead(d, m)
}

func resourceNsxtPolicyUserManagementRoleBindingDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	rbClient := aaa.NewRoleBindingsClient(connector)

	id := d.Id()
	if id == "" {
		err := fmt.Errorf("error obtaining RoleBinding ID")
		return err
	}
	roleBindingType := d.Get("type").(string)
	if roleBindingType == nsxModel.RoleBinding_TYPE_LOCAL_USER {
		if d.Get("overwrite_local_user").(bool) {
			return revertRoleBinding(d, m)
		}
		return fmt.Errorf("role binding for %s %s can not be deleted", roleBindingType, id)
	}

	log.Printf("[INFO] Deleting RoleBinding with ID %s", id)
	err := rbClient.Delete(id, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleDeleteError("RoleBinding", id, err)
	}

	return nil
}
