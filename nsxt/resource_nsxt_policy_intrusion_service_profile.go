// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	services "github.com/vmware/terraform-provider-nsxt/api/infra/settings/firewall/security/intrusion_services"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

var idsProfileSeverityValues = []string{
	model.IdsProfile_PROFILE_SEVERITY_MEDIUM,
	model.IdsProfile_PROFILE_SEVERITY_HIGH,
	model.IdsProfile_PROFILE_SEVERITY_LOW,
	model.IdsProfile_PROFILE_SEVERITY_CRITICAL,
}

var idsProfileCvssValues = []string{
	model.IdsSignature_CVSS_NONE,
	model.IdsSignature_CVSS_LOW,
	model.IdsSignature_CVSS_MEDIUM,
	model.IdsSignature_CVSS_HIGH,
	model.IdsSignature_CVSS_CRITICAL,
}

var idsProfileSignatureActionValues = []string{
	model.IdsProfileLocalSignature_ACTION_ALERT,
	model.IdsProfileLocalSignature_ACTION_DROP,
	model.IdsProfileLocalSignature_ACTION_REJECT,
}

func resourceNsxtPolicyIntrusionServiceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIntrusionServiceProfileCreate,
		Read:   resourceNsxtPolicyIntrusionServiceProfileRead,
		Update: resourceNsxtPolicyIntrusionServiceProfileUpdate,
		Delete: resourceNsxtPolicyIntrusionServiceProfileDelete,
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
			"context":      getContextSchema(false, false, false),
			"criteria": {
				Type:        schema.TypeList,
				Description: "Filtering criteria for the IDS Profile",
				Optional:    true,
				Elem:        getIdsProfileCriteriaSchema(),
				MaxItems:    1,
			},
			"overridden_signature": {
				Type:        schema.TypeSet,
				Description: "Signatures that has been overridden for this Profile",
				Optional:    true,
				Elem:        getIdsProfileSignatureSchema(),
			},
			"severities": {
				Type:        schema.TypeSet,
				Description: "Severities of signatures which are part of this profile",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(idsProfileSeverityValues, false),
				},
				Required: true,
			},
		},
	}
}

func getIdsProfileCriteriaSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"attack_types": {
				Type:        schema.TypeSet,
				Description: "List of attack type criteria",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"attack_targets": {
				Type:        schema.TypeSet,
				Description: "List of attack target criteria",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cvss": {
				Type:        schema.TypeSet,
				Description: "Common Vulnerability Scoring System Ranges",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(idsProfileCvssValues, false),
				},
			},
			"products_affected": {
				Type:        schema.TypeSet,
				Description: "List of products affected",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func getIdsProfileSignatureSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"signature_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"action": {
				Type:         schema.TypeString,
				Description:  "This will take precedence over IDS signature action",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(idsProfileSignatureActionValues, false),
				Default:      model.IdsProfileLocalSignature_ACTION_ALERT,
			},
		},
	}
}

func buildIdsProfileCriteriaFilter(name string, values []string) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()
	item := model.IdsProfileFilterCriteria{
		FilterName:   &name,
		FilterValue:  values,
		ResourceType: model.IdsProfileCriteria_RESOURCE_TYPE_IDSPROFILEFILTERCRITERIA,
	}

	dataValue, errs := converter.ConvertToVapi(item, model.IdsProfileFilterCriteriaBindingType())
	if errs != nil {
		return nil, errs[0]
	}
	return dataValue.(*data.StructValue), nil
}

func buildIdsProfileCriteriaOperator() (*data.StructValue, error) {
	op := "AND"
	operator := model.IdsProfileConjunctionOperator{
		Operator:     &op,
		ResourceType: model.IdsProfileCriteria_RESOURCE_TYPE_IDSPROFILECONJUNCTIONOPERATOR,
	}

	converter := bindings.NewTypeConverter()

	dataValue, errs := converter.ConvertToVapi(operator, model.IdsProfileConjunctionOperatorBindingType())
	if errs != nil {
		return nil, errs[0]
	}

	return dataValue.(*data.StructValue), nil
}

func getIdsProfileCriteriaFromSchema(d *schema.ResourceData) ([]*data.StructValue, error) {
	var result []*data.StructValue

	criteria := d.Get("criteria").([]interface{})
	if len(criteria) == 0 {
		return result, nil
	}

	data := criteria[0].(map[string]interface{})
	attackTypes := interfaceListToStringList(data["attack_types"].(*schema.Set).List())
	attackTargets := interfaceListToStringList(data["attack_targets"].(*schema.Set).List())
	cvss := interfaceListToStringList(data["cvss"].(*schema.Set).List())
	productsAffected := interfaceListToStringList(data["products_affected"].(*schema.Set).List())

	if len(attackTypes) > 0 {
		item, err := buildIdsProfileCriteriaFilter(model.IdsProfileFilterCriteria_FILTER_NAME_ATTACK_TYPE, attackTypes)
		if err != nil {
			return result, err
		}

		result = append(result, item)

		operator, err1 := buildIdsProfileCriteriaOperator()
		if err1 != nil {
			return result, err1
		}
		result = append(result, operator)
	}

	if len(attackTargets) > 0 {
		item, err := buildIdsProfileCriteriaFilter(model.IdsProfileFilterCriteria_FILTER_NAME_ATTACK_TARGET, attackTargets)
		if err != nil {
			return result, err
		}

		result = append(result, item)

		operator, err1 := buildIdsProfileCriteriaOperator()
		if err1 != nil {
			return result, err1
		}
		result = append(result, operator)
	}

	if len(cvss) > 0 {
		item, err := buildIdsProfileCriteriaFilter(model.IdsProfileFilterCriteria_FILTER_NAME_CVSS, cvss)
		if err != nil {
			return result, err
		}

		result = append(result, item)

		operator, err1 := buildIdsProfileCriteriaOperator()
		if err1 != nil {
			return result, err1
		}
		result = append(result, operator)
	}

	if len(productsAffected) > 0 {
		item, err := buildIdsProfileCriteriaFilter(model.IdsProfileFilterCriteria_FILTER_NAME_PRODUCT_AFFECTED, productsAffected)
		if err != nil {
			return result, err
		}

		result = append(result, item)

		operator, err1 := buildIdsProfileCriteriaOperator()
		if err1 != nil {
			return result, err1
		}
		result = append(result, operator)
	}

	// Remove last operator
	return result[:len(result)-1], nil
}

func setIdsProfileCriteriaInSchema(criteriaList []*data.StructValue, d *schema.ResourceData) error {
	var schemaList []map[string]interface{}
	converter := bindings.NewTypeConverter()
	criteriaMap := make(map[string]interface{})

	for i, item := range criteriaList {
		// Odd elements are AND operators - ignoring them
		if i%2 == 1 {
			continue
		}

		dataValue, errs := converter.ConvertToGolang(item, model.IdsProfileFilterCriteriaBindingType())
		if errs != nil {
			return errs[0]
		}

		criteria := dataValue.(model.IdsProfileFilterCriteria)
		if *criteria.FilterName == model.IdsProfileFilterCriteria_FILTER_NAME_ATTACK_TYPE {
			criteriaMap["attack_types"] = criteria.FilterValue
		} else if *criteria.FilterName == model.IdsProfileFilterCriteria_FILTER_NAME_ATTACK_TARGET {
			criteriaMap["attack_targets"] = criteria.FilterValue
		} else if *criteria.FilterName == model.IdsProfileFilterCriteria_FILTER_NAME_CVSS {
			criteriaMap["cvss"] = criteria.FilterValue
		} else if *criteria.FilterName == model.IdsProfileFilterCriteria_FILTER_NAME_PRODUCT_AFFECTED {
			criteriaMap["products_affected"] = criteria.FilterValue
		}
	}
	schemaList = append(schemaList, criteriaMap)

	return d.Set("criteria", schemaList)
}

func getIdsProfileSignaturesFromSchema(d *schema.ResourceData) []model.IdsProfileLocalSignature {
	var result []model.IdsProfileLocalSignature

	signatures := d.Get("overridden_signature").(*schema.Set).List()
	if len(signatures) == 0 {
		return result
	}

	for _, item := range signatures {
		dataMap := item.(map[string]interface{})
		action := dataMap["action"].(string)
		enabled := dataMap["enabled"].(bool)
		id := dataMap["signature_id"].(string)
		signature := model.IdsProfileLocalSignature{
			Action:      &action,
			Enable:      &enabled,
			SignatureId: &id,
		}

		result = append(result, signature)
	}

	return result
}

func setIdsProfileSignaturesInSchema(profileList []model.IdsProfileLocalSignature, d *schema.ResourceData) error {
	var schemaList []map[string]interface{}

	for _, profile := range profileList {
		signatureMap := make(map[string]interface{})
		signatureMap["action"] = profile.Action
		signatureMap["enabled"] = profile.Enable
		signatureMap["signature_id"] = profile.SignatureId

		schemaList = append(schemaList, signatureMap)
	}

	return d.Set("overridden_signature", schemaList)
}

func resourceNsxtPolicyIntrusionServiceProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	var err error
	client := services.NewProfilesClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIntrusionServiceProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyIntrusionServiceProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	criteria, err := getIdsProfileCriteriaFromSchema(d)
	if err != nil {
		return fmt.Errorf("Failed to read criteria from Ids Profile: %v", err)
	}
	signatures := getIdsProfileSignaturesFromSchema(d)
	profileSeverity := getStringListFromSchemaSet(d, "severities")

	obj := model.IdsProfile{
		DisplayName:          &displayName,
		Description:          &description,
		Tags:                 tags,
		Criteria:             criteria,
		OverriddenSignatures: signatures,
		ProfileSeverity:      profileSeverity,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating Intrusion Service Profile with ID %s", id)
	client := services.NewProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("Ids Profile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIntrusionServiceProfileRead(d, m)
}

func resourceNsxtPolicyIntrusionServiceProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Ids Profile ID")
	}

	client := services.NewProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Ids Profile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	err = setIdsProfileCriteriaInSchema(obj.Criteria, d)
	if err != nil {
		return handleReadError(d, "Ids Profile", id, err)
	}
	err = setIdsProfileSignaturesInSchema(obj.OverriddenSignatures, d)
	if err != nil {
		return handleReadError(d, "Ids Profile", id, err)
	}
	d.Set("severities", obj.ProfileSeverity)

	return nil
}

func resourceNsxtPolicyIntrusionServiceProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IdsProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	criteria, err := getIdsProfileCriteriaFromSchema(d)
	if err != nil {
		return fmt.Errorf("Failed to read criteria from Ids Profile: %v", err)
	}
	signatures := getIdsProfileSignaturesFromSchema(d)
	profileSeverity := getStringListFromSchemaSet(d, "severities")

	obj := model.IdsProfile{
		DisplayName:          &displayName,
		Description:          &description,
		Tags:                 tags,
		Criteria:             criteria,
		OverriddenSignatures: signatures,
		ProfileSeverity:      profileSeverity,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Update Intrusion Service Profile with ID %s", id)
	client := services.NewProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("Ids Profile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIntrusionServiceProfileRead(d, m)
}

func resourceNsxtPolicyIntrusionServiceProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Ids Profile ID")
	}

	connector := getPolicyConnector(m)
	var err error
	client := services.NewProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Delete(id)

	if err != nil {
		return handleDeleteError("Ids Profile", id, err)
	}

	return nil
}
