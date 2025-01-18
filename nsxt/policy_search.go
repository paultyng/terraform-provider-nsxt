// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/search"
	lm_search "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/search"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

type policySearchDataValue struct {
	StructValue *data.StructValue
	Resource    model.PolicyResource
}

func policyDataSourceResourceFilterAndSet(d *schema.ResourceData, resultValues []*data.StructValue, resourceType string) (*data.StructValue, error) {
	var perfectMatch, prefixMatch []policySearchDataValue
	var obj policySearchDataValue
	objName := d.Get("display_name").(string)
	objID := d.Get("id").(string)
	converter := bindings.NewTypeConverter()

	for _, result := range resultValues {
		dataValue, errors := converter.ConvertToGolang(result, model.PolicyResourceBindingType())
		if len(errors) > 0 {
			return nil, errors[0]
		}
		policyResource := dataValue.(model.PolicyResource)
		if resourceType != *policyResource.ResourceType {
			continue
		}

		if objID != "" {
			perfectMatch = append(perfectMatch, policySearchDataValue{StructValue: result, Resource: policyResource})
			break
		} else {
			if *policyResource.DisplayName == objName {
				perfectMatch = append(perfectMatch, policySearchDataValue{StructValue: result, Resource: policyResource})
			}
			if strings.HasPrefix(*policyResource.DisplayName, objName) {
				prefixMatch = append(prefixMatch, policySearchDataValue{StructValue: result, Resource: policyResource})
			}
		}
	}

	if len(perfectMatch) > 0 {
		if len(perfectMatch) > 1 {
			if objID != "" {
				return nil, fmt.Errorf("Found multiple %s with ID '%s'", resourceType, objID)
			}
			return nil, fmt.Errorf("Found multiple %s with name '%s'", resourceType, objName)
		}
		obj = perfectMatch[0]
	} else if len(prefixMatch) > 0 {
		if len(prefixMatch) > 1 {
			return nil, fmt.Errorf("Found multiple %s with name starting with '%s'", resourceType, objName)
		}
		obj = prefixMatch[0]
	} else {
		if objID != "" {
			return nil, fmt.Errorf("%s with ID '%s' was not found", resourceType, objID)
		}
		return nil, fmt.Errorf("%s with name '%s' was not found", resourceType, objName)
	}

	d.SetId(*obj.Resource.Id)
	d.Set("display_name", obj.Resource.DisplayName)
	d.Set("description", obj.Resource.Description)
	d.Set("path", obj.Resource.Path)

	return obj.StructValue, nil
}

func policyDataSourceResourceRead(d *schema.ResourceData, connector client.Connector, context utl.SessionContext, resourceType string, additionalQuery map[string]string) (*data.StructValue, error) {
	return policyDataSourceResourceReadWithValidation(d, connector, context, resourceType, additionalQuery, true)
}

func policyDataSourceResourceReadWithValidation(d *schema.ResourceData, connector client.Connector, context utl.SessionContext, resourceType string, additionalQuery map[string]string, paramsValidation bool) (*data.StructValue, error) {
	objName := d.Get("display_name").(string)
	objID := d.Get("id").(string)
	var err error
	var resultValues []*data.StructValue
	additionalQueryString := buildQueryStringFromMap(additionalQuery)
	if paramsValidation && objID == "" && objName == "" {
		return nil, fmt.Errorf("No 'id' or 'display_name' specified for %s", resourceType)
	}
	if objID != "" {
		if resourceType == "PolicyEdgeNode" {
			// Edge Node is a special case where id != nsx_id
			// TODO: consider switching all searches to nsx id
			resultValues, err = listPolicyResourcesByNsxID(connector, context, &objID, &additionalQueryString)
		} else {
			resultValues, err = listPolicyResourcesByID(connector, context, &objID, &additionalQueryString)
		}
	} else {
		resultValues, err = listPolicyResourcesByNameAndType(connector, context, objName, resourceType, &additionalQueryString)
	}
	if err != nil {
		return nil, err
	}

	return policyDataSourceResourceFilterAndSet(d, resultValues, resourceType)
}

func listPolicyResourcesByNameAndType(connector client.Connector, context utl.SessionContext, displayName string, resourceType string, additionalQuery *string) ([]*data.StructValue, error) {
	query := fmt.Sprintf("resource_type:%s AND display_name:%s* AND marked_for_delete:false", resourceType, escapeSpecialCharacters(displayName))
	switch context.ClientType {
	case utl.Local:
		return searchLMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Global:
		return searchGMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Multitenancy, utl.VPC:
		return searchMultitenancyResources(connector, context, *buildPolicyResourcesQuery(&query, additionalQuery))
	}

	return nil, errors.New("invalid ClientType")
}

func listInventoryResourcesByNameAndType(connector client.Connector, context utl.SessionContext, displayName string, resourceType string, additionalQuery *string) ([]*data.StructValue, error) {
	query := fmt.Sprintf("resource_type:%s AND display_name:%s*", resourceType, escapeSpecialCharacters(displayName))
	return searchLM(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
}

func listInventoryResourcesByAnyFieldAndType(connector client.Connector, context utl.SessionContext, anyField string, resourceType string, additionalQuery *string) ([]*data.StructValue, error) {
	query := fmt.Sprintf("resource_type:%s AND %s", resourceType, escapeSpecialCharacters(anyField))
	return searchLM(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
}

func escapeSpecialCharacters(str string) string {
	// we replace special characters that can be encountered in object IDs
	specials := "()[]+-=&|><!{}^~*?:/"
	if !strings.ContainsAny(str, specials) {
		return str
	}
	for _, chr := range specials {
		strchr := string(chr)
		str = strings.Replace(str, strchr, "\\"+strchr, -1)
	}

	return str
}

func listPolicyResourcesByID(connector client.Connector, context utl.SessionContext, resourceID *string, additionalQuery *string) ([]*data.StructValue, error) {
	query := fmt.Sprintf("id:%s AND marked_for_delete:false", escapeSpecialCharacters(*resourceID))
	switch context.ClientType {
	case utl.Local:
		return searchLMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Global:
		return searchGMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Multitenancy, utl.VPC:
		return searchMultitenancyResources(connector, context, *buildPolicyResourcesQuery(&query, additionalQuery))
	}

	return nil, errors.New("invalid ClientType")
}

func listPolicyResourcesByNsxID(connector client.Connector, context utl.SessionContext, resourceID *string, additionalQuery *string) ([]*data.StructValue, error) {
	query := fmt.Sprintf("nsx_id:%s AND marked_for_delete:false", escapeSpecialCharacters(*resourceID))
	switch context.ClientType {
	case utl.Local:
		return searchLMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Global:
		return searchGMPolicyResources(connector, *buildPolicyResourcesQuery(&query, additionalQuery))
	case utl.Multitenancy, utl.VPC:
		return searchMultitenancyResources(connector, context, *buildPolicyResourcesQuery(&query, additionalQuery))
	}
	return nil, errors.New("invalid ClientType")
}

func buildPolicyResourcesQuery(query *string, additionalQuery *string) *string {
	if additionalQuery != nil && *additionalQuery != "" {
		*query = *query + " AND " + *additionalQuery
	}
	return query
}

func searchGMPolicyResources(connector client.Connector, query string) ([]*data.StructValue, error) {
	client := search.NewQueryClient(connector)
	var results []*data.StructValue
	var cursor *string
	total := 0

	// Make sure local objects are not found (path needs to start with global-infra)
	query = query + " AND path:\\/global-infra*"

	for {
		searchResponse, err := client.List(query, cursor, nil, nil, nil, nil)
		if err != nil {
			return results, err
		}
		results = append(results, searchResponse.Results...)
		if total == 0 {
			// first response
			total = int(*searchResponse.ResultCount)
		}
		cursor = searchResponse.Cursor
		if len(results) >= total {
			return results, nil
		}
	}
}

func searchLM(connector client.Connector, query string) ([]*data.StructValue, error) {
	client := lm_search.NewQueryClient(connector)
	var results []*data.StructValue
	var cursor *string
	total := 0

	for {
		searchResponse, err := client.List(query, cursor, nil, nil, nil, nil)
		if err != nil {
			return results, err
		}
		results = append(results, searchResponse.Results...)
		if total == 0 {
			// first response
			total = int(*searchResponse.ResultCount)
		}
		cursor = searchResponse.Cursor
		if len(results) >= total {
			return results, nil
		}
	}
}

func searchLMPolicyResources(connector client.Connector, query string) ([]*data.StructValue, error) {
	// Make sure global objects are not found (path needs to start with infra)
	query = query + " AND path:\\/infra*"

	return searchLM(connector, query)
}

func searchMultitenancyResources(connector client.Connector, context utl.SessionContext, query string) ([]*data.StructValue, error) {
	if len(context.VPCID) > 0 {
		query = query + fmt.Sprintf(" AND path:\\/orgs\\/%s\\/projects\\/%s\\/vpcs\\/%s*", utl.DefaultOrgID, context.ProjectID, context.VPCID)
	} else {
		query = query + fmt.Sprintf(" AND path:\\/orgs\\/%s\\/projects\\/%s*", utl.DefaultOrgID, context.ProjectID)
	}
	return searchLM(connector, query)
}
