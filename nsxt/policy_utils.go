// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkerrors "github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/realized_state"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

// ErrNotAPolicyPath - Define an ignorable error for  policy path importer - to indicate that the given path is not a
// policy path and may be processed as an id - which is handy for legacy import method
var ErrNotAPolicyPath = errors.New("specified import identifier is not a policy path")
var ErrEmptyImportID = errors.New("import identifier cannot be empty")

func isSpaceString(s string) bool {
	return (strings.TrimSpace(s) == "")
}

func getOrGenerateID2(d *schema.ResourceData, m interface{}, presenceChecker func(utl.SessionContext, string, client.Connector) (bool, error)) (string, error) {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		return newUUID(), nil
	}

	exists, err := presenceChecker(getSessionContext(d, m), id, connector)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("Resource with id %s already exists", id)
	}

	return id, nil
}

func getOrGenerateID(d *schema.ResourceData, m interface{}, presenceChecker func(string, client.Connector, bool) (bool, error)) (string, error) {
	connector := getPolicyConnector(m)
	isGlobalManager := isPolicyGlobalManager(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		return newUUID(), nil
	}

	exists, err := presenceChecker(id, connector, isGlobalManager)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("Resource with id %s already exists", id)
	}

	return id, nil
}

func getOrGenerateIDWithParent(d *schema.ResourceData, m interface{}, presenceChecker func(utl.SessionContext, string, string, client.Connector) (bool, error)) (string, error) {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		return newUUID(), nil
	}

	parentPath := d.Get("parent_path").(string)

	exists, err := presenceChecker(getSessionContext(d, m), parentPath, id, connector)
	if err != nil {
		return "", err
	}

	if exists {
		return "", fmt.Errorf("Resource with id %s already exists", id)
	}

	return id, nil
}

func newUUID() string {
	uuid, _ := uuid.NewRandom()
	return uuid.String()
}

func getPolicyTagsFromSet(tagSet *schema.Set) []model.Tag {
	tags := tagSet.List()
	var tagList []model.Tag
	for _, tag := range tags {
		data := tag.(map[string]interface{})
		tagScope := data["scope"].(string)
		tagTag := data["tag"].(string)
		elem := model.Tag{
			Scope: &tagScope,
			Tag:   &tagTag}

		tagList = append(tagList, elem)
	}
	return tagList
}

func initPolicyTagsSet(tags []model.Tag) []map[string]interface{} {
	var tagList []map[string]interface{}
	for _, tag := range tags {
		elem := make(map[string]interface{})
		elem["scope"] = tag.Scope
		elem["tag"] = tag.Tag
		tagList = append(tagList, elem)
	}
	return tagList
}

func getIgnoredTagsFromSchema(d *schema.ResourceData) []model.Tag {
	tags, defined := d.GetOk("ignore_tags")
	if !defined {
		return nil
	}

	tagMaps := tags.([]interface{})
	if len(tagMaps) == 0 {
		return nil
	}
	tagMap := tagMaps[0].(map[string]interface{})
	discoveredTags := tagMap["detected"].(*schema.Set)
	return getPolicyTagsFromSet(discoveredTags)
}

func getCustomizedPolicyTagsFromSchema(d *schema.ResourceData, schemaName string) ([]model.Tag, error) {
	tags := d.Get(schemaName).(*schema.Set).List()
	ignoredTags := getIgnoredTagsFromSchema(d)
	tagList := make([]model.Tag, 0)
	for _, tag := range tags {
		data := tag.(map[string]interface{})
		tagScope := data["scope"].(string)
		tagTag := data["tag"].(string)

		if len(tagScope)+len(tagTag) == 0 {
			return tagList, fmt.Errorf("tag value or scope value needs to be specified")

		}
		elem := model.Tag{
			Scope: &tagScope,
			Tag:   &tagTag}

		tagList = append(tagList, elem)
	}
	if len(ignoredTags) > 0 {
		tagList = append(tagList, ignoredTags...)
	}
	return tagList, nil
}

func setIgnoredTagsInSchema(d *schema.ResourceData, scopesToIgnore []string, tags []map[string]interface{}) {

	elem := make(map[string]interface{})
	elem["scopes"] = scopesToIgnore
	elem["detected"] = tags

	d.Set("ignore_tags", []map[string]interface{}{elem})
}

func getTagScopesToIgnore(d *schema.ResourceData) []string {

	tags, defined := d.GetOk("ignore_tags")
	if !defined {
		return nil
	}
	tagMaps := tags.([]interface{})
	if len(tagMaps) == 0 {
		return nil
	}
	tagMap := tagMaps[0].(map[string]interface{})
	return interface2StringList(tagMap["scopes"].([]interface{}))
}

// TODO - replace with slices.Contains when go is upgraded everywhere
func shouldIgnoreScope(scope string, scopesToIgnore []string) bool {
	for _, scopeToIgnore := range scopesToIgnore {
		if scope == scopeToIgnore {
			return true
		}
	}
	return false
}

func setCustomizedPolicyTagsInSchema(d *schema.ResourceData, tags []model.Tag, schemaName string) {
	var tagList []map[string]interface{}
	var ignoredTagList []map[string]interface{}
	scopesToIgnore := getTagScopesToIgnore(d)
	for _, tag := range tags {
		elem := make(map[string]interface{})
		elem["scope"] = tag.Scope
		elem["tag"] = tag.Tag
		if tag.Scope != nil && shouldIgnoreScope(*tag.Scope, scopesToIgnore) {
			ignoredTagList = append(ignoredTagList, elem)
		} else {
			tagList = append(tagList, elem)
		}
	}
	err := d.Set(schemaName, tagList)
	if err != nil {
		log.Printf("[WARNING] Failed to set tag in schema: %v", err)
	}

	if len(scopesToIgnore) > 0 {
		setIgnoredTagsInSchema(d, scopesToIgnore, ignoredTagList)
	}
}

func getPolicyTagsFromSchema(d *schema.ResourceData) []model.Tag {
	tags, _ := getCustomizedPolicyTagsFromSchema(d, "tag")
	return tags
}

func getValidatedTagsFromSchema(d *schema.ResourceData) ([]model.Tag, error) {
	return getCustomizedPolicyTagsFromSchema(d, "tag")
}

func setPolicyTagsInSchema(d *schema.ResourceData, tags []model.Tag) {
	setCustomizedPolicyTagsInSchema(d, tags, "tag")
}

func getPathListFromMap(data map[string]interface{}, attrName string) []string {
	pathList := interface2StringList(data[attrName].(*schema.Set).List())
	if len(pathList) == 0 {
		// Convert empty value to "ANY"
		pathList = append(pathList, "ANY")
	}

	return pathList
}

func setPathListInMap(data map[string]interface{}, attrName string, pathList []string) {
	if len(pathList) == 1 && pathList[0] == "ANY" {
		data[attrName] = nil
	} else {
		data[attrName] = pathList
	}
}

func getPathListFromSchema(d *schema.ResourceData, schemaAttrName string) []string {
	pathList := interface2StringList(d.Get(schemaAttrName).(*schema.Set).List())
	if len(pathList) == 0 {
		// Convert empty value to "ANY"
		pathList = append(pathList, "ANY")
	}
	return pathList
}

func setPathListInSchema(d *schema.ResourceData, attrName string, pathList []string) {
	if len(pathList) == 1 && pathList[0] == "ANY" {
		d.Set(attrName, nil)
		return
	}
	d.Set(attrName, pathList)
}

func getDomainFromResourcePath(rPath string) string {
	return getResourceIDFromResourcePath(rPath, "domains")
}

func getProjectIDFromResourcePath(rPath string) string {
	return getResourceIDFromResourcePath(rPath, "projects")
}

func getResourceIDFromResourcePath(rPath string, rType string) string {
	segments := strings.Split(rPath, "/")
	for i, seg := range segments {
		if seg == rType && i+1 < len(segments) {
			return segments[i+1]
		}
	}
	return ""
}

func parseStandardPolicyPath(path string) ([]string, error) {
	var parents []string
	segments := strings.Split(path, "/")
	if len(segments) < 3 {
		return nil, fmt.Errorf("invalid policy path %s", path)
	}
	if segments[0] != "" {
		return nil, fmt.Errorf("policy path is expected to start with /")
	}
	// starting with *infra index
	idx := 1
	infraPath := true
	if segments[1] == "orgs" {
		if len(segments) < 5 {
			return nil, fmt.Errorf("invalid multitenant policy path %s", path)
		}

		// append org and project
		parents = append(parents, segments[2])
		parents = append(parents, segments[4])
		if len(segments) == 5 { // This is a project path, no further parsing is required
			return parents, nil
		}
		idx = 5

		if len(segments) > 6 && segments[5] == "vpcs" {
			parents = append(parents, segments[6])
			idx = 7
			// vpc paths do not contain infra
			infraPath = false
		}
	}
	if len(segments) < idx {
		return nil, fmt.Errorf("unexpected policy path %s", path)
	}
	if infraPath {
		// continue after infra marker
		if segments[idx] == "infra" || segments[idx] == "global-infra" {
			idx++
		}
	}

	for i, seg := range segments[idx:] {
		// in standard policy path, odd segments are object ids
		if i%2 == 1 {
			parents = append(parents, seg)
		}
	}
	return parents, nil
}

func parseStandardPolicyPathVerifySize(path string, expectedSize int) ([]string, error) {
	parents, err := parseStandardPolicyPath(path)
	if err != nil {
		return parents, err
	}
	if len(parents) != expectedSize {
		return parents, fmt.Errorf("Unexpected parent path %s (expected %d parent ids, got %d)", path, expectedSize, len(parents))
	}

	return parents, nil
}

func nsxtDomainResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importDomain := defaultDomain
	importID := d.Id()
	if isSpaceString(importID) {
		return []*schema.ResourceData{d}, ErrEmptyImportID
	}
	s := strings.Split(importID, "/")
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		for i, seg := range s {
			if seg == "domains" {
				d.Set("domain", s[i+1])
			}
		}
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	if len(s) == 2 {
		importDomain = s[0]
		d.SetId(s[1])
	} else {
		d.SetId(s[0])
	}

	d.Set("domain", importDomain)

	return []*schema.ResourceData{d}, nil
}

func getParameterFromPolicyPath(startDelimiter, endDelimiter, policyPath string) (string, error) {
	startIndex := strings.Index(policyPath, startDelimiter)
	endIndex := strings.Index(policyPath, endDelimiter)
	if startIndex < 0 || endIndex < 0 || (startIndex+len(startDelimiter)) > endIndex {
		return "", fmt.Errorf("failed to parse policy path %s, delimited by '%s' and '%s'", policyPath, startDelimiter, endDelimiter)
	}
	return policyPath[startIndex+len(startDelimiter) : endIndex], nil
}

func validateImportPolicyPath(policyPath string) error {
	if isSpaceString(policyPath) {
		return ErrEmptyImportID
	}
	_, err := url.ParseRequestURI(policyPath)
	if err != nil {
		return ErrNotAPolicyPath
	}
	if strings.Contains(policyPath, "//") {
		return ErrNotAPolicyPath
	}
	if !isPolicyPath(policyPath) {
		return ErrNotAPolicyPath
	}
	return nil
}

// This importer function accepts policy path and resource ID
func nsxtPolicyPathResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if errors.Is(err, ErrNotAPolicyPath) {
		return rd, nil
	} else if err != nil {
		return rd, err
	}
	return rd, nil
}

// This importer function accepts policy path only as import ID
func nsxtPolicyPathOnlyResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	return nsxtPolicyPathResourceImporterHelper(d, m)
}

func nsxtVPCPathResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}
	projectID, vpcID := getContextDataFromSchema(d)
	if projectID == "" || vpcID == "" {
		return rd, fmt.Errorf("imported resource policy path should have both project_id and vpc_id fields")
	}
	return rd, nil
}

func nsxtPolicyPathResourceImporterHelper(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	err := validateImportPolicyPath(importID)
	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	pathSegs := strings.Split(importID, "/")
	if strings.Contains(pathSegs[1], "infra") {
		d.SetId(pathSegs[len(pathSegs)-1])
	} else if pathSegs[1] == "orgs" && pathSegs[3] == "projects" {
		if len(pathSegs) < 5 {
			return nil, fmt.Errorf("invalid policy multitenancy path %s", importID)
		}
		// pathSegs[2] should contain the organization. Once we support multiple organization, it should be
		// assigned into the context as well
		ctxMap := make(map[string]interface{})
		ctxMap["project_id"] = pathSegs[4]
		if pathSegs[5] == "vpcs" {
			ctxMap["vpc_id"] = pathSegs[6]
		}
		d.Set("context", []interface{}{ctxMap})
		d.SetId(pathSegs[len(pathSegs)-1])
	}
	return []*schema.ResourceData{d}, nil
}

func nsxtParentPathResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	err := validateImportPolicyPath(importID)
	if err != nil {
		return []*schema.ResourceData{}, err
	}

	pathSegs := strings.Split(importID, "/")
	segCount := len(pathSegs)
	d.SetId(pathSegs[segCount-1])
	// to get parent path size, remove last two tokens (id and separator) plus two slashes
	truncateSize := len(pathSegs[segCount-1]) + len(pathSegs[segCount-2]) + 2
	d.Set("parent_path", importID[:len(importID)-truncateSize])
	return []*schema.ResourceData{d}, nil
}

func isPolicyPath(policyPath string) bool {
	pathSegs := strings.Split(policyPath, "/")
	if len(pathSegs) < 4 {
		return false
	} else if pathSegs[0] != "" || pathSegs[len(pathSegs)-1] == "" {
		return false
	} else if !strings.Contains(pathSegs[1], "infra") && pathSegs[1] != "orgs" {
		// must be infra, global-infra or orgs as of now
		return false
	}
	return true
}

func isValidID(id string) bool {
	v, err := regexp.MatchString("^[0-9a-zA-Z_\\-]+$", id)
	if err != nil {
		return false
	}
	return v
}

func getPolicyIDFromPath(path string) string {
	tokens := strings.Split(path, "/")
	return tokens[len(tokens)-1]
}

func interfaceListToStringList(interfaces []interface{}) []string {
	var strList []string
	for _, elem := range interfaces {
		strList = append(strList, elem.(string))
	}
	return strList
}

func policyResourceNotSupportedError() error {
	return fmt.Errorf("This NSX policy resource is not supported with given provider settings")
}

func collectSeparatedStringListToMap(stringList []string, separator string) map[string]string {
	strMap := make(map[string]string)
	for _, elem := range stringList {
		segs := strings.Split(elem, separator)
		if len(segs) > 1 {
			strMap[segs[0]] = segs[1]
		}

	}
	return strMap
}

func stringListToCommaSeparatedString(stringList []string) *string {
	if len(stringList) > 0 {
		var str string
		for i, seg := range stringList {
			str += seg
			if i < len(stringList)-1 {
				str += ","
			}
		}
		return &str
	}
	return nil
}

func commaSeparatedStringToStringList(commaString string) []string {
	var strList []string
	for _, seg := range strings.Split(commaString, ",") {
		if seg != "" {
			strList = append(strList, seg)
		}
	}
	return strList
}

func nsxtPolicyWaitForRealizationStateConf(connector client.Connector, d *schema.ResourceData, realizedEntityPath string, timeout int) *resource.StateChangeConf {
	client := realized_state.NewRealizedEntitiesClient(connector)
	pendingStates := []string{"UNKNOWN", "UNREALIZED"}
	targetStates := []string{"REALIZED", "ERROR"}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {

			realizationResult, realizationError := client.List(realizedEntityPath, nil)
			if realizationError == nil {
				// Find the right entry
				for _, objInList := range realizationResult.Results {
					if objInList.State != nil {
						return objInList, *objInList.State, nil
					}
				}
				// Realization info not found yet
				return nil, "UNKNOWN", nil
			}
			return nil, "", realizationError
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}

	return stateConf
}

func getPolicyEnforcementPointPath(m interface{}) string {
	return "/infra/sites/default/enforcement-points/" + getPolicyEnforcementPoint(m)
}

func getGlobalPolicyEnforcementPointPathWithLocation(m interface{}, location string) string {
	return "/global-infra/sites/" + location + "/enforcement-points/" + getPolicyEnforcementPoint(m)
}

func convertModelBindingType(obj interface{}, sourceType bindings.BindingType, destType bindings.BindingType) (interface{}, error) {
	converter := bindings.NewTypeConverter()
	dataValue, err := converter.ConvertToVapi(obj, sourceType)
	if err != nil {
		return nil, err[0]
	}

	gmObj, err := converter.ConvertToGolang(dataValue, destType)
	if err != nil {
		return nil, err[0]
	}

	return gmObj, nil
}

func retryUponPreconditionFailed(readAndUpdate func() error, maxRetryAttempts int) error {
	// This retry specific to Precondition Error, and solution
	// here required refreshing the object, and updating revision
	// in request body. This can not be solved with SDK-based retry
	// functionality since it always retries with same request.
	var err error
	for i := 0; i <= maxRetryAttempts; i++ {
		err = readAndUpdate()
		if err == nil {
			return nil
		}

		if _, ok := err.(sdkerrors.InvalidRequest); !ok {
			// other type of error
			return err
		}

		log.Printf("[INFO] Refreshing object and repeating operation, attempt %d", i+1)
	}

	return err
}

func getElemOrEmptyMapFromSchema(d *schema.ResourceData, key string) map[string]interface{} {
	e := d.Get(key)
	if e != nil {
		elems := e.([]interface{})
		if len(elems) > 0 {
			if elemMap, ok := elems[0].(map[string]interface{}); ok {
				return elemMap
			}
		}
	}
	return make(map[string]interface{})
}

func getElemOrEmptyMapFromMap(d map[string]interface{}, key string) map[string]interface{} {
	e := d[key]
	if e != nil {
		elems := e.([]interface{})
		if len(elems) > 0 {
			if elemMap, ok := elems[0].(map[string]interface{}); ok {
				return elemMap
			}
		}
	}
	return make(map[string]interface{})
}

func setPolicyLbHTTPHeaderInSchema(d *schema.ResourceData, attrName string, headers []model.LbHttpRequestHeader) {
	var headerList []map[string]string
	for _, header := range headers {
		elem := make(map[string]string)
		elem["name"] = *header.HeaderName
		elem["value"] = *header.HeaderValue
		headerList = append(headerList, elem)
	}
	d.Set(attrName, headerList)
}

func getPolicyLbHTTPHeaderFromSchema(d *schema.ResourceData, attrName string) []model.LbHttpRequestHeader {
	headers := d.Get(attrName).(*schema.Set).List()
	var headerList []model.LbHttpRequestHeader
	for _, header := range headers {
		data := header.(map[string]interface{})
		name := data["name"].(string)
		value := data["value"].(string)
		elem := model.LbHttpRequestHeader{
			HeaderName:  &name,
			HeaderValue: &value}

		headerList = append(headerList, elem)
	}
	return headerList
}

func getPolicyLbMonitorPortSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Description: "If the monitor port is specified, it would override pool member port setting for healthcheck. A port range is not supported",
		Optional:    true,
	}
}

func getVpcParentsFromContext(context utl.SessionContext) []string {
	return []string{utl.DefaultOrgID, context.ProjectID, context.VPCID}
}
