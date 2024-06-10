/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_infra "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_domain "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/domains"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

// This resource is supported only for Policy Global Manager
func resourceNsxtPolicyDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyDomainCreate,
		Read:   resourceNsxtPolicyDomainRead,
		Update: resourceNsxtPolicyDomainUpdate,
		Delete: resourceNsxtPolicyDomainDelete,
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
			"sites": {
				Type:        schema.TypeSet,
				Description: "Sites where this domain is deployed",
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceNsxtPolicyDomainExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	if isGlobalManager {
		client := gm_infra.NewDomainsClient(connector)
		_, err = client.Get(id)
	} else {
		return false, fmt.Errorf("Domain resource is not supported for local manager")
	}

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving domain", err)
}

func createChildDomainDeploymentMap(m interface{}, domainID string, location string) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	mapID := domainID + "-" + location
	path := getGlobalPolicyEnforcementPointPathWithLocation(m, location)

	Type := "DomainDeploymentMap"
	mapObj := model.DomainDeploymentMap{
		Id:                   &mapID,
		DisplayName:          &location,
		EnforcementPointPath: &path,
		ResourceType:         &Type,
	}

	childMap := model.ChildDomainDeploymentMap{
		Id:                  &mapID,
		ResourceType:        "ChildDomainDeploymentMap",
		DomainDeploymentMap: &mapObj,
	}

	dataValue, errors := converter.ConvertToVapi(childMap, model.ChildDomainDeploymentMapBindingType())
	if len(errors) > 0 {
		return nil, errors[0]
	}
	return dataValue.(*data.StructValue), nil
}

func setDomainStructWithChildren(m interface{}, domain *model.Domain, locations []string, isUpdate bool) error {
	var children []*data.StructValue

	// Loop over locations
	for _, location := range locations {
		child, err := createChildDomainDeploymentMap(m, *domain.Id, location)
		if err != nil {
			return err
		}
		if child != nil {
			children = append(children, child)
		}
	}

	if isUpdate {
		// Delete old existing locations
		connector := getPolicyConnector(m)
		converter := bindings.NewTypeConverter()
		dmClient := gm_domain.NewDomainDeploymentMapsClient(connector)
		// Get all current locations
		objList, err := dmClient.List(*domain.Id, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return handleListError("Domain", err)
		}
		for _, objInList := range objList.Results {
			found := false
			for _, newLocation := range locations {
				if newLocation == *objInList.DisplayName {
					found = true
					break
				}
			}
			if !found {
				// Remove the unused deployment map
				toDelete := true
				lmObj, err2 := convertModelBindingType(objInList, gm_model.DomainDeploymentMapBindingType(), model.DomainDeploymentMapBindingType())
				if err2 != nil {
					return err2
				}
				deleteObj := lmObj.(model.DomainDeploymentMap)
				childMap := model.ChildDomainDeploymentMap{
					Id:                  objInList.Id,
					ResourceType:        "ChildDomainDeploymentMap",
					MarkedForDelete:     &toDelete,
					DomainDeploymentMap: &deleteObj,
				}

				dataValue, errors := converter.ConvertToVapi(childMap, model.ChildDomainDeploymentMapBindingType())
				if len(errors) > 0 {
					return errors[0]
				}
				children = append(children, dataValue.(*data.StructValue))
			}
		}
	}
	domain.Children = children
	return nil
}

func resourceNsxtPolicyDomainCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyDomainExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	Type := "Domain"
	obj := model.Domain{
		Id:           &id,
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		ResourceType: &Type,
	}
	locations := getStringListFromSchemaSet(d, "sites")
	err = setDomainStructWithChildren(m, &obj, locations, false)
	if err != nil {
		return err
	}

	childDomain := model.ChildDomain{
		Domain:       &obj,
		ResourceType: "ChildDomain",
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childDomain, model.ChildDomainBindingType())
	if errors != nil {
		return fmt.Errorf("Error converting Domain Child: %v", errors[0])
	}

	infraType := "Infra"
	var infraChildren []*data.StructValue
	infraChildren = append(infraChildren, dataValue.(*data.StructValue))
	infraStruct := model.Infra{
		Children:     infraChildren,
		ResourceType: &infraType,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating Domain with ID %s", id)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	err = policyInfraPatch(context, infraStruct, getPolicyConnector(m), false)
	if err != nil {
		return handleCreateError("Domain", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyDomainRead(d, m)
}

func resourceNsxtPolicyDomainRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Domain ID")
	}

	if !isPolicyGlobalManager(m) {
		return handleCreateError("Domain", id, fmt.Errorf("Domain resource is not supported for local manager"))
	}

	client := gm_infra.NewDomainsClient(connector)
	gmObj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Domain", id, err)
	}

	lmObj, err := convertModelBindingType(gmObj, gm_model.DomainBindingType(), model.DomainBindingType())
	if err != nil {
		return err
	}
	obj := lmObj.(model.Domain)

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	// Also read deployment maps
	dmClient := gm_domain.NewDomainDeploymentMapsClient(connector)
	objList, err := dmClient.List(id, nil, nil, nil, nil, nil, nil)
	if err != nil {
		return handleListError("Domain", err)
	}
	var locations []string
	for _, objInList := range objList.Results {
		locations = append(locations, *objInList.DisplayName)
	}
	d.Set("sites", stringList2Interface(locations))
	return nil
}

func resourceNsxtPolicyDomainUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Domain ID")
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	Type := "Domain"
	obj := model.Domain{
		Id:           &id,
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		ResourceType: &Type,
	}
	locations := getStringListFromSchemaSet(d, "sites")
	err := setDomainStructWithChildren(m, &obj, locations, true)
	if err != nil {
		return err
	}

	childDomain := model.ChildDomain{
		Domain:       &obj,
		ResourceType: "ChildDomain",
	}

	converter := bindings.NewTypeConverter()
	dataValue, errors := converter.ConvertToVapi(childDomain, model.ChildDomainBindingType())
	if errors != nil {
		return fmt.Errorf("Error converting Domain Child: %v", errors[0])
	}

	infraType := "Infra"
	var infraChildren []*data.StructValue
	infraChildren = append(infraChildren, dataValue.(*data.StructValue))
	infraStruct := model.Infra{
		Children:     infraChildren,
		ResourceType: &infraType,
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	err = policyInfraPatch(context, infraStruct, getPolicyConnector(m), false)
	if err != nil {
		return handleUpdateError("Domain", id, err)
	}
	return resourceNsxtPolicyDomainRead(d, m)
}

func resourceNsxtPolicyDomainDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Domain ID")
	}

	connector := getPolicyConnector(m)
	client := gm_infra.NewDomainsClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("Domain", id, err)
	}

	return nil
}
