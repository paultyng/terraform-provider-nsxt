// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package metadata

import (
	"fmt"
	"log"
	"os"
	"reflect"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vapiBindings_ "github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
)

// package level logger to include log.Lshortfile context
var logger = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)

const (
	PolymorphicTypeFlatten = "flatten"
	PolymorphicTypeNested  = "nested"
)

type Metadata struct {
	// we need a separate schema type, in addition to terraform SDK type,
	// in order to distinguish between single subclause and a list of entries
	SchemaType   string
	ReadOnly     bool
	SdkFieldName string
	// This attribute is parent path for the object
	IsParentPath bool
	// The type of polymorphic relation between SDK and TF schema
	// Empty for non-polymorphic fields
	PolymorphicType string
	// SDK vapi binding type for converting polymorphic structs
	BindingType vapiBindings_.BindingType
	// SDK resource type to match and filter for a schema key
	// Only applicable to PolymorphicTypeFlatten schema
	ResourceType string
	// Map from schema key of polymorphic attr to this SDK resource type
	// Only applicable to PolymorphicTypeNested schema
	ResourceTypeMap map[string]string
	// Type identifier name for both SDK and JSON (StructValue field name)
	TypeIdentifier      TypeIdentifier
	IntroducedInVersion string
	// skip handling of this attribute - it will be done manually
	Skip        bool
	ReflectType reflect.Type
	OmitIfEmpty bool
	TestData    Testdata
}

type ExtendedSchema struct {
	Schema   schema.Schema
	Metadata Metadata
}

type ExtendedResource struct {
	Schema map[string]*ExtendedSchema
}

type Testdata struct {
	CreateValue interface{}
	UpdateValue interface{}
}

type TypeIdentifier struct {
	SdkName      string
	APIFieldName string
}

// GetSdkName returns the SDK field type identifier
// Defaults to ResourceType
func (t TypeIdentifier) GetSdkName() string {
	if t.SdkName != "" {
		return t.SdkName
	}
	return "ResourceType"
}

// GetAPIFieldName returns the API JSON type identifier
// Defaults to resource_type
func (t TypeIdentifier) GetAPIFieldName() string {
	if t.APIFieldName != "" {
		return t.APIFieldName
	}
	return "resource_type"
}

// GetExtendedSchema is a helper to convert terraform sdk schema to extended schema
func GetExtendedSchema(sch *schema.Schema) *ExtendedSchema {
	shallowCopy := *sch
	return &ExtendedSchema{
		Schema: shallowCopy,
		Metadata: Metadata{
			Skip: true,
		},
	}
}

// GetSchemaFromExtendedSchema gets terraform sdk schema from extended schema definition
func GetSchemaFromExtendedSchema(ext map[string]*ExtendedSchema) map[string]*schema.Schema {
	result := make(map[string]*schema.Schema)

	for key, value := range ext {
		logger.Printf("[TRACE] inspecting schema key %s, value %v", key, value)
		shallowCopy := value.Schema
		if (value.Schema.Type == schema.TypeList) || (value.Schema.Type == schema.TypeSet) {
			elem, ok := shallowCopy.Elem.(*ExtendedSchema)
			if ok {
				shallowCopy.Elem = &elem.Schema
			} else {
				elem, ok := shallowCopy.Elem.(*ExtendedResource)
				if ok {
					shallowCopy.Elem = &schema.Resource{
						Schema: GetSchemaFromExtendedSchema(elem.Schema),
					}
				}
			}
		}
		// TODO: deepcopy needed?
		result[key] = &shallowCopy
	}

	return result
}

func getContextString(prefix, parent string, elemType reflect.Type) string {
	ctx := elemType.String()
	if len(parent) > 0 {
		ctx = fmt.Sprintf("%s->%s", parent, elemType.Name())
	}
	return fmt.Sprintf("[%s %s]", prefix, ctx)
}

// StructToSchema converts NSX model struct to terraform schema
// currently supports nested subtype and trivial types
func StructToSchema(elem reflect.Value, d *schema.ResourceData, metadata map[string]*ExtendedSchema, parent string, parentMap map[string]interface{}) (err error) {
	ctx := getContextString("from", parent, elem.Type())
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s recovered from panic: %v", ctx, r)
			logger.Printf("[ERROR] %v", err)
		}
	}()

	for key, item := range metadata {
		if item.Metadata.Skip {
			continue
		}

		logger.Printf("[TRACE] %s inspecting key %s", ctx, key)
		if len(parent) > 0 {
			logger.Printf("[TRACE] %s parent %s key %s", ctx, parent, key)
		}
		if !elem.FieldByName(item.Metadata.SdkFieldName).IsValid() {
			// FieldByName can't find the field by name
			logger.Printf("[ERROR] %s skip key %s as %s not found in struct",
				ctx, key, elem.Type())
			err = fmt.Errorf("%s key %s not found in %s",
				ctx, key, elem.Type())
			return
		}
		if elem.FieldByName(item.Metadata.SdkFieldName).IsNil() {
			logger.Printf("[TRACE] %s skip key %s with nil value", ctx, key)
			continue
		}
		if len(item.Metadata.PolymorphicType) > 0 {
			childElem := elem.FieldByName(item.Metadata.SdkFieldName)
			var nestedVal interface{}
			switch item.Metadata.PolymorphicType {
			case PolymorphicTypeNested:
				nestedVal, err = polyStructToNestedSchema(ctx, childElem, item)
			case PolymorphicTypeFlatten:
				nestedVal, err = polyStructToFlattenSchema(ctx, childElem, key, item)
			default:
				err = fmt.Errorf("%s unknown polymorphic type %s", ctx,
					item.Metadata.PolymorphicType)
			}
			if err != nil {
				return
			}

			if len(parent) > 0 {
				parentMap[key] = nestedVal
			} else {
				d.Set(key, nestedVal)
			}
			logger.Printf("[TRACE] %s adding polymorphic slice %+v to key %s", ctx, nestedVal, key)
			continue
		}
		if item.Metadata.SchemaType == "struct" {
			nestedObj := elem.FieldByName(item.Metadata.SdkFieldName)
			nestedSchema := make(map[string]interface{})
			childElem := item.Schema.Elem.(*ExtendedResource)
			if err = StructToSchema(nestedObj.Elem(), d, childElem.Schema, key, nestedSchema); err != nil {
				return
			}
			logger.Printf("[TRACE] %s assigning struct %+v to %s", ctx, nestedObj, key)
			var nestedSlice []map[string]interface{}
			nestedSlice = append(nestedSlice, nestedSchema)
			if len(parent) > 0 {
				parentMap[key] = nestedSlice
			} else {
				d.Set(key, nestedSlice)
			}
		} else if item.Metadata.SchemaType == "list" || item.Metadata.SchemaType == "set" {
			if _, ok := item.Schema.Elem.(*ExtendedSchema); ok {
				// List of string, bool, int
				nestedSlice := elem.FieldByName(item.Metadata.SdkFieldName)
				logger.Printf("[TRACE] %s assigning slice %v to %s", ctx, nestedSlice.Interface(), key)
				if len(parent) > 0 {
					parentMap[key] = nestedSlice.Interface()
				} else {
					d.Set(key, nestedSlice.Interface())
				}
			} else if childElem, ok := item.Schema.Elem.(*ExtendedResource); ok {
				// List of struct
				sliceElem := elem.FieldByName(item.Metadata.SdkFieldName)
				var nestedSlice []map[string]interface{}
				for i := 0; i < sliceElem.Len(); i++ {
					nestedSchema := make(map[string]interface{})
					if err = StructToSchema(sliceElem.Index(i), d, childElem.Schema, key, nestedSchema); err != nil {
						return
					}
					nestedSlice = append(nestedSlice, nestedSchema)
					logger.Printf("[TRACE] %s appending slice item %+v to %s", ctx, nestedSchema, key)
				}
				if len(parent) > 0 {
					parentMap[key] = nestedSlice
				} else {
					d.Set(key, nestedSlice)
				}
			}
		} else {
			if len(parent) > 0 {
				logger.Printf("[TRACE] %s assigning nested value %+v to %s",
					ctx, elem.FieldByName(item.Metadata.SdkFieldName).Interface(), key)
				parentMap[key] = elem.FieldByName(item.Metadata.SdkFieldName).Interface()
			} else {
				logger.Printf("[TRACE] %s assigning value %+v to %s",
					ctx, elem.FieldByName(item.Metadata.SdkFieldName).Interface(), key)
				d.Set(key, elem.FieldByName(item.Metadata.SdkFieldName).Interface())
			}
		}
	}

	return
}

// SchemaToStruct converts terraform schema to NSX model struct
// currently supports nested subtype and trivial types
func SchemaToStruct(elem reflect.Value, d *schema.ResourceData, metadata map[string]*ExtendedSchema, parent string, parentMap map[string]interface{}) (err error) {
	ctx := getContextString("to", parent, elem.Type())
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s recovered from panic: %v", ctx, r)
			logger.Printf("[ERROR] %v", err)
		}
	}()

	for key, item := range metadata {
		if item.Metadata.ReadOnly {
			logger.Printf("[TRACE] %s skip key %s as read only", ctx, key)
			continue
		}
		if item.Metadata.Skip {
			logger.Printf("[TRACE] %s skip key %s", ctx, key)
			continue
		}
		if item.Metadata.IntroducedInVersion != "" && util.NsxVersionLower(item.Metadata.IntroducedInVersion) {
			logger.Printf("[TRACE] %s skip key %s as NSX version is lower than %v", ctx, key, item.Metadata.IntroducedInVersion)
			continue
		}
		if !elem.FieldByName(item.Metadata.SdkFieldName).IsValid() {
			// FieldByName can't find the field by name
			logger.Printf("[WARN] %s skip key %s as %s not found in struct",
				ctx, key, elem.Type())
			err = fmt.Errorf("%s key %s not found in %s",
				ctx, key, elem.Type())
			return
		}

		logger.Printf("[TRACE] %s inspecting key %s with type %s", ctx, key, item.Metadata.SchemaType)
		if len(parent) > 0 {
			logger.Printf("[TRACE] %s parent %s key %s", ctx, parent, key)
		}
		if len(item.Metadata.PolymorphicType) > 0 {
			logger.Printf("[TRACE] %s inspecting polymorphic key %s", ctx, key)
			itemList := getItemListForSchemaToStruct(d, item.Metadata.SchemaType, key, parent, parentMap)
			switch item.Metadata.PolymorphicType {
			case PolymorphicTypeNested:
				err = polyNestedSchemaToStruct(ctx, elem, itemList, item)
			case PolymorphicTypeFlatten:
				err = polyFlattenSchemaToStruct(ctx, elem, key, itemList, item)
			default:
				err = fmt.Errorf("%s unknown polymorphic type %s", ctx,
					item.Metadata.PolymorphicType)
			}
			if err != nil {
				return
			}
			continue
		}
		if item.Metadata.SchemaType == "string" {
			var value string
			exists := false
			if len(parent) > 0 && parentMap[key] != nil {
				value = parentMap[key].(string)
				// For nested maps, value is initialized to zero string even if not
				// specified by user explicitly
				if len(value) > 0 {
					exists = true
				}
			} else {
				var v interface{}
				v, exists = d.GetOk(key)
				if exists {
					value = v.(string)
				}
			}
			if item.Metadata.OmitIfEmpty && !exists {
				logger.Printf("[TRACE] %s skip key %s since its empty and OmitIfEmpty is true", ctx, key)
				continue
			}
			logger.Printf("[TRACE] %s assigning string %v to %s", ctx, value, key)
			elem.FieldByName(item.Metadata.SdkFieldName).Set(reflect.ValueOf(&value))
		}
		if item.Metadata.SchemaType == "bool" {
			var value bool
			exists := false
			if len(parent) > 0 && parentMap[key] != nil {
				value = parentMap[key].(bool)
				exists = true
			} else {
				var v interface{}
				v, exists = d.GetOk(key)
				if exists {
					value = v.(bool)
				}
			}
			if item.Metadata.OmitIfEmpty && !exists {
				logger.Printf("[TRACE] %s skip key %s since its empty and OmitIfEmpty is true", ctx, key)
				continue
			}
			logger.Printf("[TRACE] %s assigning bool %v to %s", ctx, value, key)
			elem.FieldByName(item.Metadata.SdkFieldName).Set(reflect.ValueOf(&value))
		}
		if item.Metadata.SchemaType == "int" {
			var value int64
			exists := false
			if len(parent) > 0 && parentMap[key] != nil {
				value = int64(parentMap[key].(int))
				exists = true
			} else {
				var v interface{}
				v, exists = d.GetOk(key)
				if exists {
					value = int64(v.(int))
				}
			}
			if item.Metadata.OmitIfEmpty && value == 0 && !exists {
				logger.Printf("[TRACE] %s skip key %s since its empty and OmitIfEmpty is true", ctx, key)
				continue
			}
			logger.Printf("[TRACE] %s assigning int %v to %s", ctx, value, key)
			elem.FieldByName(item.Metadata.SdkFieldName).Set(reflect.ValueOf(&value))
		}
		if item.Metadata.SchemaType == "struct" {
			nestedObj := reflect.New(item.Metadata.ReflectType)
			itemList := getItemListForSchemaToStruct(d, item.Metadata.SchemaType, key, parent, parentMap)
			if len(itemList) == 0 {
				logger.Printf("[TRACE] Item list empty")
				continue
			}
			if itemList[0] == nil {
				// empty clause is specified
				logger.Printf("[TRACE] Item list contains empty value")
				continue
			}
			nestedSchema := itemList[0].(map[string]interface{})

			childElem := item.Schema.Elem.(*ExtendedResource)
			if err = SchemaToStruct(nestedObj.Elem(), d, childElem.Schema, key, nestedSchema); err != nil {
				return
			}
			logger.Printf("[TRACE] %s assigning struct %v to %s", ctx, nestedObj, key)
			elem.FieldByName(item.Metadata.SdkFieldName).Set(nestedObj)
		}
		if item.Metadata.SchemaType == "list" || item.Metadata.SchemaType == "set" {
			itemList := getItemListForSchemaToStruct(d, item.Metadata.SchemaType, key, parent, parentMap)
			// List of string, bool, int
			if childElem, ok := item.Schema.Elem.(*ExtendedSchema); ok {
				sliceElem := elem.FieldByName(item.Metadata.SdkFieldName)
				switch childElem.Metadata.SchemaType {
				case "string":
					sliceElem.Set(
						reflect.MakeSlice(reflect.TypeOf([]string{}), len(itemList), len(itemList)))
				case "bool":
					sliceElem.Set(
						reflect.MakeSlice(reflect.TypeOf([]bool{}), len(itemList), len(itemList)))
				case "int":
					sliceElem.Set(
						reflect.MakeSlice(reflect.TypeOf([]int64{}), len(itemList), len(itemList)))
				}

				for i, v := range itemList {
					if v == nil {
						continue
					}
					if childElem.Metadata.SchemaType == "int" {
						sliceElem.Index(i).Set(reflect.ValueOf(v).Convert(reflect.TypeOf(int64(0))))
					} else {
						sliceElem.Index(i).Set(reflect.ValueOf(v))
					}
					logger.Printf("[TRACE] %s appending %v to %s", ctx, v, key)
				}
			}

			// List of struct
			if childElem, ok := item.Schema.Elem.(*ExtendedResource); ok {
				sliceElem := elem.FieldByName(item.Metadata.SdkFieldName)
				sliceElem.Set(
					reflect.MakeSlice(reflect.SliceOf(item.Metadata.ReflectType), len(itemList), len(itemList)))
				for i, childItem := range itemList {
					if childItem == nil {
						continue
					}
					nestedObj := reflect.New(item.Metadata.ReflectType)
					nestedSchema := childItem.(map[string]interface{})
					if err = SchemaToStruct(nestedObj.Elem(), d, childElem.Schema, key, nestedSchema); err != nil {
						return
					}
					sliceElem.Index(i).Set(nestedObj.Elem())
					logger.Printf("[TRACE] %s appending %+v to %s", ctx, nestedObj.Elem(), key)
				}
			}
		}
	}

	return
}

func getItemListForSchemaToStruct(d *schema.ResourceData, schemaType, key, parent string, parentMap map[string]interface{}) []interface{} {
	var itemList []interface{}
	if schemaType == "list" || schemaType == "struct" {
		if len(parent) > 0 {
			itemList = parentMap[key].([]interface{})
		} else {
			itemList = d.Get(key).([]interface{})
		}
	} else if schemaType == "set" {
		if len(parent) > 0 {
			itemList = parentMap[key].(*schema.Set).List()
		} else {
			itemList = d.Get(key).(*schema.Set).List()
		}
	}
	return itemList
}

// getResourceTypeFromStructValue returns resource type from SDK object based on identifier
func getResourceTypeFromStructValue(ctx string, value data.StructValue, identifier TypeIdentifier) (string, error) {
	if !value.HasField(identifier.GetAPIFieldName()) {
		err := fmt.Errorf(
			"%s failed to get resource type", ctx)
		logger.Printf("[ERROR] %v", err)
		return "", err
	}
	var rTypeData data.DataValue
	rTypeData, err := value.Field(identifier.GetAPIFieldName())
	if err != nil {
		return "", err
	}
	if strVal, ok := rTypeData.(*data.StringValue); ok {
		return strVal.Value(), nil
	}
	if opVal, ok := rTypeData.(*data.OptionalValue); ok {
		return opVal.String()
	}

	err = fmt.Errorf("%s failed to convert resource type %s",
		ctx, identifier.GetAPIFieldName())
	logger.Printf("[ERROR] %v", err)
	return "", err
}

func polyMetadataSanityCheck(ctx string, item *ExtendedSchema, polyType string) error {
	if item.Metadata.PolymorphicType != polyType {
		err := fmt.Errorf("%s unexpected polymorphic attr", ctx)
		logger.Printf("[ERROR] %v", err)
		return err
	}

	_, ok := item.Schema.Elem.(*ExtendedResource)
	if !ok {
		err := fmt.Errorf("%s polymorphic attr has non-ExtendedResource element", ctx)
		logger.Printf("[ERROR] %v", err)
		return err
	}

	return nil
}

func polyStructToNestedSchema(ctx string, elem reflect.Value, item *ExtendedSchema) (ret []map[string]interface{}, err error) {
	if err = polyMetadataSanityCheck(ctx, item, PolymorphicTypeNested); err != nil {
		return
	}

	elemSlice := elem
	if item.Metadata.SchemaType == "struct" {
		if elem.IsNil() {
			return
		}
		// convert to slice to share logic with list and set
		elemSlice = reflect.MakeSlice(reflect.SliceOf(elem.Type()), 1, 1)
		elemSlice.Index(0).Set(elem)
	} else if item.Metadata.SchemaType != "list" && item.Metadata.SchemaType != "set" {
		err = fmt.Errorf("%s unsupported polymorphic schema type %s", ctx, item.Metadata.SchemaType)
		logger.Printf("[ERROR] %v", err)
		return
	}
	if elemSlice.Len() == 0 {
		return
	}

	converter := vapiBindings_.NewTypeConverter()
	sliceValue := make([]map[string]interface{}, elemSlice.Len())
	for i := 0; i < elemSlice.Len(); i++ {
		childElem := elemSlice.Index(i).Elem().Interface().(data.StructValue)
		nestedSchema := make(map[string]interface{})
		var key, rType string
		var childExtSch *ExtendedSchema
		rType, err = getResourceTypeFromStructValue(ctx, childElem, item.Metadata.TypeIdentifier)
		if err != nil {
			return
		}

		// Get metadata for the corresponding type
		for k, v := range item.Metadata.ResourceTypeMap {
			if v == rType {
				key = k
				childExtSch = item.Schema.Elem.(*ExtendedResource).Schema[k]
				break
			}
		}
		if len(key) == 0 || childExtSch == nil {
			err = fmt.Errorf("%s polyStructToSchema failed to get schema meta of type %s for %s",
				ctx, rType, item.Metadata.SdkFieldName)
			logger.Printf("[ERROR] %v", err)
			return
		}

		// Convert to concrete struct
		dv, errors := converter.ConvertToGolang(&childElem, childExtSch.Metadata.BindingType)
		if errors != nil {
			err = errors[0]
			logger.Printf("[ERROR] %v", err)
			return
		}

		if err = StructToSchema(reflect.ValueOf(dv), nil, childExtSch.Schema.Elem.(*ExtendedResource).Schema, key, nestedSchema); err != nil {
			return
		}
		sliceValue[i] = make(map[string]interface{})
		sliceValue[i][key] = []interface{}{nestedSchema}
		logger.Printf("[TRACE] %s adding %+v of key %s", ctx, dv, key)
	}

	ret = sliceValue
	return
}

func polyNestedSchemaToStruct(ctx string, elem reflect.Value, dataList []interface{}, item *ExtendedSchema) (err error) {
	if err = polyMetadataSanityCheck(ctx, item, PolymorphicTypeNested); err != nil {
		return
	}

	childElem := item.Schema.Elem.(*ExtendedResource)
	converter := vapiBindings_.NewTypeConverter()
	dv := make([]*data.StructValue, len(dataList))
	for i, dataElem := range dataList {
		dataMap := dataElem.(map[string]interface{})
		for k, v := range dataMap {
			if len(v.([]interface{})) == 0 {
				continue
			}
			rType, ok := item.Metadata.ResourceTypeMap[k]
			if !ok {
				err = fmt.Errorf("%s polymorphic attr has key %s not found in resource type map", ctx, k)
				logger.Printf("[ERROR] %v", err)
				return
			}
			if _, ok := childElem.Schema[k]; !ok {
				err = fmt.Errorf("%s polymorphic attr has key %s not found in metadata", ctx, k)
				logger.Printf("[ERROR] %v", err)
				return
			}

			childMeta := childElem.Schema[k]
			nestedObj := reflect.New(childMeta.Metadata.ReflectType)
			nestedSchema := v.([]interface{})[0].(map[string]interface{})
			if err = SchemaToStruct(nestedObj.Elem(), nil, childMeta.Schema.Elem.(*ExtendedResource).Schema, k, nestedSchema); err != nil {
				return
			}
			// set resource type based on mapping
			nestedObj.Elem().FieldByName(item.Metadata.TypeIdentifier.GetSdkName()).Set(reflect.ValueOf(&rType))

			dataValue, errors := converter.ConvertToVapi(nestedObj.Interface(), childMeta.Metadata.BindingType)
			if errors != nil {
				err = errors[0]
				logger.Printf("[ERROR] %v", err)
				return
			}
			dv[i] = dataValue.(*data.StructValue)
			logger.Printf("[TRACE] %s adding polymorphic value %+v to %s",
				ctx, nestedObj.Interface(), item.Metadata.SdkFieldName)

			// there should be only one non-empty entry in the map
			break
		}
	}

	if item.Metadata.SchemaType == "struct" {
		elem.FieldByName(item.Metadata.SdkFieldName).Set(reflect.ValueOf(dv[0]))
	} else if item.Metadata.SchemaType == "list" || item.Metadata.SchemaType == "set" {
		sliceElem := elem.FieldByName(item.Metadata.SdkFieldName)
		sliceElem.Set(
			reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(&data.StructValue{})), len(dv), len(dv)))
		for i, v := range dv {
			sliceElem.Index(i).Set(reflect.ValueOf(v))
		}
	} else {
		err = fmt.Errorf("%s unsupported polymorphic schema type %s", ctx, item.Metadata.SchemaType)
		logger.Printf("[ERROR] %v", err)
		return
	}

	return
}

func polyStructToFlattenSchema(ctx string, elem reflect.Value, key string, item *ExtendedSchema) (ret []map[string]interface{}, err error) {
	if err = polyMetadataSanityCheck(ctx, item, PolymorphicTypeFlatten); err != nil {
		return
	}

	elemSlice := elem
	if item.Metadata.SchemaType == "struct" {
		if elem.IsNil() {
			return
		}
		// convert to slice to share logic with list and set
		elemSlice = reflect.MakeSlice(reflect.SliceOf(elem.Type()), 1, 1)
		elemSlice.Index(0).Set(elem)
	} else if item.Metadata.SchemaType != "list" && item.Metadata.SchemaType != "set" {
		err = fmt.Errorf("%s unsupported polymorphic schema type %s", ctx, item.Metadata.SchemaType)
		logger.Printf("[ERROR] %v", err)
		return
	}
	if elemSlice.Len() == 0 {
		return
	}

	converter := vapiBindings_.NewTypeConverter()
	sliceValue := make([]map[string]interface{}, 0)
	for i := 0; i < elemSlice.Len(); i++ {
		childValue := elemSlice.Index(i).Elem().Interface().(data.StructValue)
		nestedSchema := make(map[string]interface{})
		var rType string

		rType, err = getResourceTypeFromStructValue(ctx, childValue, item.Metadata.TypeIdentifier)
		if err != nil {
			return
		}
		if rType != item.Metadata.ResourceType {
			continue
		}

		dv, errors := converter.ConvertToGolang(&childValue, item.Metadata.BindingType)
		if errors != nil {
			err = errors[0]
			logger.Printf("[ERROR] %v", err)
			return
		}
		if err = StructToSchema(reflect.ValueOf(dv), nil, item.Schema.Elem.(*ExtendedResource).Schema, key, nestedSchema); err != nil {
			return
		}
		sliceValue = append(sliceValue, nestedSchema)
	}

	ret = sliceValue
	return
}

func polyFlattenSchemaToStruct(ctx string, elem reflect.Value, key string, dataList []interface{}, item *ExtendedSchema) (err error) {
	if err = polyMetadataSanityCheck(ctx, item, PolymorphicTypeFlatten); err != nil {
		return
	}

	childElem := item.Schema.Elem.(*ExtendedResource)
	converter := vapiBindings_.NewTypeConverter()
	rSlice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(&data.StructValue{})), len(dataList), len(dataList))
	for i, dataElem := range dataList {
		nestedObj := reflect.New(item.Metadata.ReflectType)
		nestedSchema := dataElem.(map[string]interface{})
		if err = SchemaToStruct(nestedObj.Elem(), nil, childElem.Schema, key, nestedSchema); err != nil {
			return
		}

		// set resource type
		nestedObj.Elem().FieldByName(item.Metadata.TypeIdentifier.GetSdkName()).Set(reflect.ValueOf(
			&item.Metadata.ResourceType))

		dataValue, errors := converter.ConvertToVapi(nestedObj.Interface(), item.Metadata.BindingType)
		if errors != nil {
			err = errors[0]
			logger.Printf("[ERROR] %v", err)
			return
		}
		rSlice.Index(i).Set(reflect.ValueOf(dataValue.(*data.StructValue)))
		logger.Printf("[TRACE] %s adding polymorphic value %+v to %s",
			ctx, nestedObj.Interface(), item.Metadata.SdkFieldName)
	}

	if item.Metadata.SchemaType == "struct" {
		if rSlice.Len() == 0 {
			return
		}
		structElem := elem.FieldByName(item.Metadata.SdkFieldName)
		if !structElem.IsZero() {
			err = fmt.Errorf("%s %s is already set", ctx, item.Metadata.SdkFieldName)
			logger.Printf("[ERROR] %v", err)
			return
		}
		structElem.Set(rSlice.Index(0))
	} else if item.Metadata.SchemaType == "list" || item.Metadata.SchemaType == "set" {
		sliceElem := elem.FieldByName(item.Metadata.SdkFieldName)
		if sliceElem.IsZero() {
			sliceElem.Set(rSlice)
		} else {
			sliceElem.Set(reflect.AppendSlice(sliceElem, rSlice))
		}
	} else {
		err = fmt.Errorf("%s unsupported polymorphic schema type %s", ctx, item.Metadata.SchemaType)
		logger.Printf("[ERROR] %v", err)
		return
	}

	return
}
