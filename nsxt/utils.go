// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"log"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/common"
	"github.com/vmware/go-vmware-nsxt/manager"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	mp_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/node"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/search"
)

var adminStateValues = []string{"UP", "DOWN"}

func interface2StringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, val)
		}
	}
	return vs
}

func stringList2Interface(list []string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, v)
	}
	return vs
}

func interface2Int32List(configured []interface{}) []int32 {
	vs := make([]int32, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(int)
		if ok {
			vs = append(vs, int32(val))
		}
	}
	return vs
}

func interface2Int64List(configured []interface{}) []int64 {
	vs := make([]int64, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(int)
		if ok {
			vs = append(vs, int64(val))
		}
	}
	return vs
}

func int32List2Interface(list []int32) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, int(v))
	}
	return vs
}

func int64List2Interface(list []int64) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, int(v))
	}
	return vs
}

func getStringListFromSchemaSet(d *schema.ResourceData, schemaAttrName string) []string {
	return interface2StringList(d.Get(schemaAttrName).(*schema.Set).List())
}

func getStringListFromSchemaList(d *schema.ResourceData, schemaAttrName string) []string {
	return interface2StringList(d.Get(schemaAttrName).([]interface{}))
}

// helper to construct a map based on curtain attribute in schema set
// this helper is only relevant for Sets of nested objects (not scalars), and attrName
// is the object attribute value of which would appear as key in the returned map object.
// this is useful when Read function needs to make a decision based on intent provided
// by user in a nested schema
func getAttrKeyMapFromSchemaSet(schemaSet interface{}, attrName string) map[string]bool {

	keyMap := make(map[string]bool)
	for _, item := range schemaSet.(*schema.Set).List() {
		mapItem := item.(map[string]interface{})
		if value, ok := mapItem[attrName]; ok {
			keyMap[value.(string)] = true
		}
	}

	return keyMap
}

func intList2int64List(configured []interface{}) []int64 {
	vs := make([]int64, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(int)
		if ok {
			vs = append(vs, int64(val))
		}
	}
	return vs
}

func getRevisionSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Description: "The _revision property describes the current revision of the resource. To prevent clients from overwriting each other's changes, PUT operations must include the current _revision of the resource, which clients should obtain by issuing a GET operation. If the _revision provided in a PUT request is missing or stale, the operation will be rejected",
		Computed:    true,
	}
}

// utilities to define & handle tags
func getTagsSchemaInternal(required bool, forceNew bool) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Set of opaque identifiers meaningful to the user",
		Optional:    !required,
		Required:    required,
		ForceNew:    forceNew,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"scope": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: forceNew,
				},
				"tag": {
					Type:     schema.TypeString,
					Optional: true,
					ForceNew: forceNew,
				},
			},
		},
	}
}

// utilities to define & handle tags
func getTagsSchema() *schema.Schema {
	return getTagsSchemaInternal(false, false)
}

func getTagsSchemaForceNew() *schema.Schema {
	return getTagsSchemaInternal(false, true)
}

func getCustomizedTagsFromSchema(d *schema.ResourceData, schemaName string) []common.Tag {
	tags := d.Get(schemaName).(*schema.Set).List()
	tagList := make([]common.Tag, 0)
	for _, tag := range tags {
		data := tag.(map[string]interface{})
		elem := common.Tag{
			Scope: data["scope"].(string),
			Tag:   data["tag"].(string)}

		tagList = append(tagList, elem)
	}
	return tagList
}

func setCustomizedTagsInSchema(d *schema.ResourceData, tags []common.Tag, schemaName string) {
	var tagList []map[string]string
	for _, tag := range tags {
		elem := make(map[string]string)
		elem["scope"] = tag.Scope
		elem["tag"] = tag.Tag
		tagList = append(tagList, elem)
	}
	err := d.Set(schemaName, tagList)
	if err != nil {
		log.Printf("[WARNING] Failed to set tag in schema: %v", err)
	}
}

func getTagsFromSchema(d *schema.ResourceData) []common.Tag {
	return getCustomizedTagsFromSchema(d, "tag")
}

func setTagsInSchema(d *schema.ResourceData, tags []common.Tag) {
	setCustomizedTagsInSchema(d, tags, "tag")
}

// utilities to define & handle switching profiles
func getSwitchingProfileIdsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "List of IDs of switching profiles (of various types) to be associated with this object. Default switching profiles will be used if not specified",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:        schema.TypeString,
					Description: "The resource type of this profile",
					Required:    true,
				},
				"value": {
					Type:        schema.TypeString,
					Description: "The ID of this profile",
					Required:    true,
				},
			},
		},
	}
}

func getSwitchingProfileIdsFromSchema(d *schema.ResourceData) []manager.SwitchingProfileTypeIdEntry {
	profiles := d.Get("switching_profile_id").(*schema.Set).List()
	var profileList []manager.SwitchingProfileTypeIdEntry
	for _, profile := range profiles {
		data := profile.(map[string]interface{})
		elem := manager.SwitchingProfileTypeIdEntry{
			Key:   data["key"].(string),
			Value: data["value"].(string)}

		profileList = append(profileList, elem)
	}
	return profileList
}

func setSwitchingProfileIdsInSchema(d *schema.ResourceData, nsxClient *api.APIClient, profiles []manager.SwitchingProfileTypeIdEntry) error {
	var profileList []map[string]string
	for _, profile := range profiles {
		// ignore system owned profiles
		obj, _, _ := nsxClient.LogicalSwitchingApi.GetSwitchingProfile(nsxClient.Context, profile.Value)
		if obj.SystemOwned {
			continue
		}

		elem := make(map[string]string)
		elem["key"] = profile.Key
		elem["value"] = profile.Value
		profileList = append(profileList, elem)
	}
	err := d.Set("switching_profile_id", profileList)
	return err
}

// utilities to define & handle address bindings
func getAddressBindingsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Address bindings for the Logical switch",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"ip_address": {
					Type:         schema.TypeString,
					Description:  "A single IP address or a subnet cidr",
					Optional:     true,
					ValidateFunc: validateSingleIP(),
				},
				"mac_address": {
					Type:        schema.TypeString,
					Description: "A single MAC address",
					Optional:    true,
				},
				"vlan": {
					Type:        schema.TypeInt,
					Description: "A single vlan tag value",
					Optional:    true,
				},
			},
		},
	}
}

func getAddressBindingsFromSchema(d *schema.ResourceData) []manager.PacketAddressClassifier {
	bindings := d.Get("address_binding").(*schema.Set).List()
	var bindingList []manager.PacketAddressClassifier
	for _, binding := range bindings {
		data := binding.(map[string]interface{})
		elem := manager.PacketAddressClassifier{
			IpAddress:  data["ip_address"].(string),
			MacAddress: data["mac_address"].(string),
			Vlan:       int64(data["vlan"].(int)),
		}

		bindingList = append(bindingList, elem)
	}
	return bindingList
}

func setAddressBindingsInSchema(d *schema.ResourceData, bindings []manager.PacketAddressClassifier) error {
	var bindingList []map[string]interface{}
	for _, binding := range bindings {
		elem := make(map[string]interface{})
		elem["ip_address"] = binding.IpAddress
		elem["mac_address"] = binding.MacAddress
		elem["vlan"] = binding.Vlan
		bindingList = append(bindingList, elem)
	}
	err := d.Set("address_binding", bindingList)
	return err
}

func getResourceReferencesSchema(required bool, computed bool, validTargetTypes []string, description string) *schema.Schema {
	return getResourceReferencesSchemaByType(required, computed, validTargetTypes, true, description, 0)
}

func getSingleResourceReferencesSchema(required bool, computed bool, validTargetTypes []string, description string) *schema.Schema {
	return getResourceReferencesSchemaByType(required, computed, validTargetTypes, true, description, 1)
}

func getResourceReferencesSetSchema(required bool, computed bool, validTargetTypes []string, description string) *schema.Schema {
	return getResourceReferencesSchemaByType(required, computed, validTargetTypes, false, description, 0)
}

func getResourceReferencesSchemaByType(required bool, computed bool, validTargetTypes []string, isList bool, description string, maxItems int) *schema.Schema {
	schType := schema.TypeSet
	if isList {
		schType = schema.TypeList
	}

	return &schema.Schema{
		Type:        schType,
		Required:    required,
		Optional:    !required,
		Computed:    computed,
		MaxItems:    maxItems,
		Description: description,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"is_valid": {
					Type:        schema.TypeBool,
					Description: "A boolean flag which will be set to false if the referenced NSX resource has been deleted",
					Computed:    true,
				},
				"target_display_name": {
					Type:        schema.TypeString,
					Description: "Display name of the NSX resource",
					Computed:    true,
				},
				"target_id": {
					Type:        schema.TypeString,
					Description: "Identifier of the NSX resource",
					Optional:    true,
				},
				"target_type": {
					Type:         schema.TypeString,
					Description:  "Type of the NSX resource",
					Optional:     true,
					ValidateFunc: validation.StringInSlice(validTargetTypes, false),
				},
			},
		},
	}
}

func getSingleResourceReference(references []interface{}) *common.ResourceReference {
	for _, reference := range references {
		data := reference.(map[string]interface{})
		// only 1 ref is allowed so return the first 1
		elem := common.ResourceReference{
			IsValid:           data["is_valid"].(bool),
			TargetDisplayName: data["target_display_name"].(string),
			TargetId:          data["target_id"].(string),
			TargetType:        data["target_type"].(string),
		}
		return &elem
	}
	return nil
}

func getResourceReferences(references []interface{}) []common.ResourceReference {
	var referenceList []common.ResourceReference
	for _, reference := range references {
		data := reference.(map[string]interface{})
		elem := common.ResourceReference{
			IsValid:           data["is_valid"].(bool),
			TargetDisplayName: data["target_display_name"].(string),
			TargetId:          data["target_id"].(string),
			TargetType:        data["target_type"].(string),
		}

		referenceList = append(referenceList, elem)
	}
	return referenceList
}

func getResourceReferencesFromSchemaSet(d *schema.ResourceData, schemaAttrName string) []common.ResourceReference {
	references := d.Get(schemaAttrName).(*schema.Set).List()
	return getResourceReferences(references)
}

func returnResourceReferences(references []common.ResourceReference) []map[string]interface{} {
	var referenceList []map[string]interface{}
	for _, reference := range references {
		elem := make(map[string]interface{})
		elem["is_valid"] = reference.IsValid
		elem["target_display_name"] = reference.TargetDisplayName
		elem["target_id"] = reference.TargetId
		elem["target_type"] = reference.TargetType
		referenceList = append(referenceList, elem)
	}
	return referenceList
}

func resourceReferenceHash(v interface{}) int {
	var buf bytes.Buffer

	if v != nil {
		m := v.(map[string]interface{})
		buf.WriteString(fmt.Sprintf("%s-%s", m["target_type"], m["target_id"]))
	}
	result := int(crc32.ChecksumIEEE(buf.Bytes()))
	if result < 0 {
		return -result
	}
	return result
}

func resourceKeyValueHash(v interface{}) int {
	var buf bytes.Buffer

	if v != nil {
		m := v.(map[string]interface{})
		for k, v := range m {
			buf.WriteString(fmt.Sprintf("%s-%s", k, v))
		}
	}
	result := int(crc32.ChecksumIEEE(buf.Bytes()))
	if result < 0 {
		return -result
	}
	return result
}

func returnResourceReferencesSet(references []common.ResourceReference) *schema.Set {
	var referenceList []interface{}
	for _, reference := range references {
		elem := make(map[string]interface{})
		elem["is_valid"] = reference.IsValid
		elem["target_display_name"] = reference.TargetDisplayName
		elem["target_id"] = reference.TargetId
		elem["target_type"] = reference.TargetType
		referenceList = append(referenceList, elem)
	}

	s := schema.NewSet(resourceReferenceHash, referenceList)
	return s
}

func setResourceReferencesInSchema(d *schema.ResourceData, references []common.ResourceReference, schemaAttrName string) error {
	referenceList := returnResourceReferences(references)
	err := d.Set(schemaAttrName, referenceList)
	return err
}

func getServiceBindingsFromSchema(d *schema.ResourceData, schemaAttrName string) []manager.ServiceBinding {
	references := d.Get(schemaAttrName).([]interface{})
	var bindingList []manager.ServiceBinding
	for _, reference := range references {
		data := reference.(map[string]interface{})
		ref := common.ResourceReference{
			IsValid:           data["is_valid"].(bool),
			TargetDisplayName: data["target_display_name"].(string),
			TargetId:          data["target_id"].(string),
			TargetType:        data["target_type"].(string),
		}
		elem := manager.ServiceBinding{ServiceId: &ref}
		bindingList = append(bindingList, elem)
	}
	return bindingList
}

func setServiceBindingsInSchema(d *schema.ResourceData, serviceBindings []manager.ServiceBinding, schemaAttrName string) error {
	var referenceList []map[string]interface{}
	for _, binding := range serviceBindings {
		elem := make(map[string]interface{})
		elem["is_valid"] = binding.ServiceId.IsValid
		elem["target_display_name"] = binding.ServiceId.TargetDisplayName
		elem["target_id"] = binding.ServiceId.TargetId
		elem["target_type"] = binding.ServiceId.TargetType
		referenceList = append(referenceList, elem)
	}
	err := d.Set(schemaAttrName, referenceList)
	return err
}

func getAdminStateSchema() *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "Represents Desired state of the object",
		Default:      "UP",
		ValidateFunc: validation.StringInSlice(adminStateValues, false),
	}
}

func getIDSetSchema(description string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: description,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	}
}

func getIPRangesSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "List of IP Ranges",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"start": {
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
					Required:     true,
				},
				"end": {
					Type:         schema.TypeString,
					ValidateFunc: validateSingleIP(),
					Required:     true,
				},
			},
		},
	}
}

func getIPRangesFromSchema(d *schema.ResourceData) []manager.IpPoolRange {
	ranges := d.Get("ip_range").([]interface{})
	var rangeList []manager.IpPoolRange
	for _, r := range ranges {
		data := r.(map[string]interface{})
		elem := manager.IpPoolRange{
			Start: data["start"].(string),
			End:   data["end"].(string)}

		rangeList = append(rangeList, elem)
	}
	return rangeList
}

func setIPRangesInSchema(d *schema.ResourceData, ranges []manager.IpPoolRange) {
	var rangeList []map[string]string
	for _, r := range ranges {
		elem := make(map[string]string)
		elem["start"] = r.Start
		elem["end"] = r.End
		rangeList = append(rangeList, elem)
	}
	err := d.Set("ip_range", rangeList)
	if err != nil {
		log.Printf("[WARNING]: Failed to set ip range in schema: %v", err)
	}
}

func makeResourceReference(resourceType string, resourceID string) *common.ResourceReference {
	return &common.ResourceReference{
		TargetType: resourceType,
		TargetId:   resourceID,
	}
}

func getNSXVersion(connector client.Connector) (string, error) {
	client := node.NewVersionClient(connector)
	version, err := client.Get()
	if err != nil {
		return "", logAPIError("Failed to retrieve NSX version, please check connectivity and authentication settings of the provider", err)

	}
	log.Printf("[DEBUG] NSX version is %s", *version.NodeVersion)
	return *version.NodeVersion, nil
}

func initNSXVersion(connector client.Connector) error {
	var err error
	util.NsxVersion, err = getNSXVersion(connector)
	return err
}

func initNSXVersionVMC(clients interface{}) {
	// TODO: find a ireliable way to retrieve NSX version on VMC
	// For now, we need to determine whether the deployment is 3.0.0 and up, or below
	// For this purpose, we fire indicator search API (introduced in 3.0.0)
	util.NsxVersion = "3.0.0"

	connector := getPolicyConnector(clients)
	client := search.NewQueryClient(connector)
	var cursor *string
	query := "resource_type:dummy"
	_, err := client.List(query, cursor, nil, nil, nil, nil)
	if err == nil {
		// we are 3.0.0 and above
		log.Printf("[INFO] Assuming NSX version >= 3.0.0 in VMC environment")
		return
	}

	if isNotFoundError(err) {
		// search API not supported
		log.Printf("[INFO] Assuming NSX version < 3.0.0 in VMC environment")
		util.NsxVersion = "2.5.0"
		return
	}

	// Connectivity error - alert the user
	log.Printf("[ERROR] Failed to determine NSX version in VMC environment: %s", err)
}

func resourceNotSupportedError() error {
	return fmt.Errorf("This resource is not supported with given provider settings")
}

func dataSourceNotSupportedError() error {
	return fmt.Errorf("This data source is not supported with given provider settings")
}

func mpResourceRemovedError(resourceName string) error {
	return fmt.Errorf("MP resource %s was deprecated and has been removed in NSX 9.0.0", resourceName)
}

func mpDataSourceRemovedError(dataSourceName string) error {
	return fmt.Errorf("MP data source %s was deprecated and has been removed in NSX 9.0.0", dataSourceName)
}

func stringInList(target string, list []string) bool {
	// util to check if target string is in list
	for _, value := range list {
		if target == value {
			return true
		}
	}
	return false
}

func containsElements(target []string, list []string) bool {
	// util to check if all strings in target []string is in list. NOT ordered
	for _, value := range target {
		if !stringInList(value, list) {
			return false
		}
	}
	return true
}

type paginationInfo struct {
	TotalCount        int64
	PageCount         int64
	Cursor            string
	LocalVarOptionals map[string]interface{}
}

func handlePagination(lister func(*paginationInfo) error) (int64, error) {
	info := paginationInfo{}
	info.LocalVarOptionals = make(map[string]interface{})

	total := int64(0)
	count := int64(0)

	for total == 0 || (count < total) {
		err := lister(&info)
		if err != nil {
			return total, err
		}

		if total == 0 {
			// first response
			total = info.TotalCount
			if total == 0 {
				// empty list
				return total, nil
			}
		}
		count += info.PageCount
		log.Printf("[DEBUG] Fetching next page after %d/%d inspected", count, total)

		info.LocalVarOptionals["cursor"] = info.Cursor
	}

	return total, nil
}

func getContextSchema(isRequired, isComputed, isVPC bool) *schema.Schema {
	elemSchema := map[string]*schema.Schema{
		"project_id": {
			Type:         schema.TypeString,
			Description:  "Id of the project which the resource belongs to.",
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validateID(),
		},
	}
	if isVPC {
		elemSchema["vpc_id"] = &schema.Schema{
			Type:         schema.TypeString,
			Description:  "Id of the VPC which the resource belongs to.",
			Required:     true,
			ForceNew:     true,
			ValidateFunc: validateID(),
		}
	}
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Resource context",
		Optional:    !isRequired,
		Required:    isRequired,
		Computed:    isComputed,
		MaxItems:    1,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: elemSchema,
		},
	}
}

func getCustomizedMPTagsFromSchema(d *schema.ResourceData, schemaName string) []mp_model.Tag {
	tags := d.Get(schemaName).(*schema.Set).List()
	tagList := make([]mp_model.Tag, 0)
	for _, tag := range tags {
		data := tag.(map[string]interface{})
		scope := data["scope"].(string)
		tag := data["tag"].(string)
		elem := mp_model.Tag{
			Scope: &scope,
			Tag:   &tag}

		tagList = append(tagList, elem)
	}
	return tagList
}

func setCustomizedMPTagsInSchema(d *schema.ResourceData, tags []mp_model.Tag, schemaName string) {
	var tagList []map[string]string
	for _, tag := range tags {
		elem := make(map[string]string)
		elem["scope"] = *tag.Scope
		elem["tag"] = *tag.Tag
		tagList = append(tagList, elem)
	}
	err := d.Set(schemaName, tagList)
	if err != nil {
		log.Printf("[WARNING] Failed to set tag in schema: %v", err)
	}
}

func getMPTagsFromSchema(d *schema.ResourceData) []mp_model.Tag {
	return getCustomizedMPTagsFromSchema(d, "tag")
}

func setMPTagsInSchema(d *schema.ResourceData, tags []mp_model.Tag) {
	setCustomizedMPTagsInSchema(d, tags, "tag")
}

func getCustomizedGMTagsFromSchema(d *schema.ResourceData, schemaName string) []gm_model.Tag {
	tags := d.Get(schemaName).(*schema.Set).List()
	tagList := make([]gm_model.Tag, 0)
	for _, tag := range tags {
		data := tag.(map[string]interface{})
		scope := data["scope"].(string)
		tag := data["tag"].(string)
		elem := gm_model.Tag{
			Scope: &scope,
			Tag:   &tag}

		tagList = append(tagList, elem)
	}
	return tagList
}

func setCustomizedGMTagsInSchema(d *schema.ResourceData, tags []gm_model.Tag, schemaName string) {
	var tagList []map[string]string
	for _, tag := range tags {
		elem := make(map[string]string)
		elem["scope"] = *tag.Scope
		elem["tag"] = *tag.Tag
		tagList = append(tagList, elem)
	}
	err := d.Set(schemaName, tagList)
	if err != nil {
		log.Printf("[WARNING] Failed to set tag in schema: %v", err)
	}
}

func getGMTagsFromSchema(d *schema.ResourceData) []gm_model.Tag {
	return getCustomizedGMTagsFromSchema(d, "tag")
}

func setGMTagsInSchema(d *schema.ResourceData, tags []gm_model.Tag) {
	setCustomizedGMTagsInSchema(d, tags, "tag")
}

func getKeyValuePairListSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Advanced configuration",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func getKeyValuePairListFromSchema(kvIList interface{}) []mp_model.KeyValuePair {
	var kvList []mp_model.KeyValuePair
	if kvIList != nil {
		for _, kv := range kvIList.([]interface{}) {
			kvMap := kv.(map[string]interface{})
			key := kvMap["key"].(string)
			val := kvMap["value"].(string)
			kvList = append(kvList, mp_model.KeyValuePair{Key: &key, Value: &val})
		}
	}
	return kvList
}

func setKeyValueListForSchema(kvList []mp_model.KeyValuePair) interface{} {
	var kvIList []interface{}
	for _, ec := range kvList {
		kvMap := make(map[string]interface{})
		kvMap["key"] = ec.Key
		kvMap["value"] = ec.Value
		kvIList = append(kvIList, kvMap)
	}
	return kvIList
}
