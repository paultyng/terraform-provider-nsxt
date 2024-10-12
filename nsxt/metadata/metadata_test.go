// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package metadata

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	StringField      *string
	BoolField        *bool
	IntField         *int64
	StringFieldNil   *string
	BoolFieldNil     *bool
	IntFieldNil      *int64
	StringFieldOIE   *string
	BoolFieldOIE     *bool
	IntFieldOIE      *int64
	StructField      *testNestedStruct
	StringListField  []string
	BoolListField    []bool
	IntListField     []int64
	StructList       []testNestedStruct
	StringSetField   []string
	BoolSetField     []bool
	IntSetField      []int64
	StructSet        []testNestedStruct
	DeepNestedStruct *testDeepNestedStruct
}

type testNestedStruct struct {
	StringField *string
	BoolField   *bool
	IntField    *int64
}

type testDeepNestedStruct struct {
	StringField *string
	BoolList    []bool
	IntSet      []int64
	StructList  []testNestedStruct
}

var testSchema = map[string]*schema.Schema{
	"string_field": {
		Type: schema.TypeString,
	},
	"bool_field": {
		Type: schema.TypeBool,
	},
	"int_field": {
		Type: schema.TypeInt,
	},
	"string_field_nil": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"bool_field_nil": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"int_field_nil": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"string_field_oie": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"bool_field_oie": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"int_field_oie": {
		Type:     schema.TypeInt,
		Optional: true,
	},
	"struct_field": {
		Type:     schema.TypeList,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"string_field": {
					Type: schema.TypeString,
				},
				"bool_field": {
					Type: schema.TypeBool,
				},
				"int_field": {
					Type: schema.TypeInt,
				},
			},
		},
	},
	"string_list": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"bool_list": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeBool,
		},
	},
	"int_list": {
		Type: schema.TypeList,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
	},
	"struct_list": {
		Type: schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"string_field": {
					Type: schema.TypeString,
				},
				"bool_field": {
					Type: schema.TypeBool,
				},
				"int_field": {
					Type: schema.TypeInt,
				},
			},
		},
	},
	"string_set": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"bool_set": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeBool,
		},
	},
	"int_set": {
		Type: schema.TypeSet,
		Elem: &schema.Schema{
			Type: schema.TypeInt,
		},
	},
	"struct_set": {
		Type: schema.TypeSet,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"string_field": {
					Type: schema.TypeString,
				},
				"bool_field": {
					Type: schema.TypeBool,
				},
				"int_field": {
					Type: schema.TypeInt,
				},
			},
		},
	},
	"deep_nested_struct": {
		Type:     schema.TypeList,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"string_field": {
					Type: schema.TypeString,
				},
				"bool_list": {
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeBool,
					},
				},
				"int_set": {
					Type: schema.TypeSet,
					Elem: &schema.Schema{
						Type: schema.TypeInt,
					},
				},
				"struct_list": {
					Type: schema.TypeList,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"string_field": {
								Type: schema.TypeString,
							},
							"bool_field": {
								Type: schema.TypeBool,
							},
							"int_field": {
								Type: schema.TypeInt,
							},
						},
					},
				},
			},
		},
	},
}

func basicStringSchema(sdkName string, optional, omitIfEmpty bool) *ExtendedSchema {
	return &ExtendedSchema{
		Schema: schema.Schema{
			Type:     schema.TypeString,
			Optional: optional,
		},
		Metadata: Metadata{
			SchemaType:   "string",
			SdkFieldName: sdkName,
			OmitIfEmpty:  omitIfEmpty,
		},
	}
}

func basicBoolSchema(sdkName string, optional, omitIfEmpty bool) *ExtendedSchema {
	return &ExtendedSchema{
		Schema: schema.Schema{
			Type:     schema.TypeBool,
			Optional: optional,
		},
		Metadata: Metadata{
			SchemaType:   "bool",
			SdkFieldName: sdkName,
		},
	}
}

func basicIntSchema(sdkName string, optional, omitIfEmpty bool) *ExtendedSchema {
	return &ExtendedSchema{
		Schema: schema.Schema{
			Type:     schema.TypeInt,
			Optional: optional,
		},
		Metadata: Metadata{
			SchemaType:   "int",
			SdkFieldName: sdkName,
		},
	}
}

func basicStructSchema(t string) schema.Schema {
	schemaType := schema.TypeList
	maxItems := 0
	if t == "set" {
		schemaType = schema.TypeSet
	} else if t == "struct" {
		maxItems = 1
	}
	return schema.Schema{
		Type:     schemaType,
		MaxItems: maxItems,
		Elem: &ExtendedResource{
			Schema: map[string]*ExtendedSchema{
				"string_field": basicStringSchema("StringField", false, false),
				"bool_field":   basicBoolSchema("BoolField", false, false),
				"int_field":    basicIntSchema("IntField", false, false),
			},
		},
	}
}

func mixedStructSchema() schema.Schema {
	return schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Elem: &ExtendedResource{
			Schema: map[string]*ExtendedSchema{
				"string_field": basicStringSchema("StringField", false, false),
				"bool_list": {
					Schema: schema.Schema{
						Type: schema.TypeList,
						Elem: basicBoolSchema("BoolList", false, false),
					},
					Metadata: Metadata{
						SchemaType:   "list",
						SdkFieldName: "BoolList",
					},
				},
				"int_set": {
					Schema: schema.Schema{
						Type: schema.TypeSet,
						Elem: basicIntSchema("IntSet", false, false),
					},
					Metadata: Metadata{
						SchemaType:   "set",
						SdkFieldName: "IntSet",
					},
				},
				"struct_list": {
					Schema: basicStructSchema("list"),
					Metadata: Metadata{
						SchemaType:   "list",
						SdkFieldName: "StructList",
						ReflectType:  reflect.TypeOf(testNestedStruct{}),
					},
				},
			},
		},
	}
}

var testExtendedSchema = map[string]*ExtendedSchema{
	"string_field":     basicStringSchema("StringField", false, false),
	"bool_field":       basicBoolSchema("BoolField", false, false),
	"int_field":        basicIntSchema("IntField", false, false),
	"string_field_nil": basicStringSchema("StringFieldNil", true, false),
	"bool_field_nil":   basicBoolSchema("BoolFieldNil", true, false),
	"int_field_nil":    basicIntSchema("IntFieldNil", true, false),
	"string_field_oie": basicStringSchema("StringFieldNil", true, true),
	"bool_field_oie":   basicBoolSchema("BoolFieldNil", true, true),
	"int_field_oie":    basicIntSchema("IntFieldNil", true, true),
	"struct_field": {
		Schema: basicStructSchema("struct"),
		Metadata: Metadata{
			SchemaType:   "struct",
			SdkFieldName: "StructField",
			ReflectType:  reflect.TypeOf(testNestedStruct{}),
		},
	},
	"string_list": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: basicStringSchema("StringListField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "list",
			SdkFieldName: "StringListField",
		},
	},
	"bool_list": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: basicBoolSchema("BoolListField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "list",
			SdkFieldName: "BoolListField",
		},
	},
	"int_list": {
		Schema: schema.Schema{
			Type: schema.TypeList,
			Elem: basicIntSchema("IntListField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "list",
			SdkFieldName: "IntListField",
		},
	},
	"struct_list": {
		Schema: basicStructSchema("list"),
		Metadata: Metadata{
			SchemaType:   "list",
			SdkFieldName: "StructList",
			ReflectType:  reflect.TypeOf(testNestedStruct{}),
		},
	},
	"string_set": {
		Schema: schema.Schema{
			Type: schema.TypeSet,
			Elem: basicStringSchema("StringSetField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "set",
			SdkFieldName: "StringSetField",
		},
	},
	"bool_set": {
		Schema: schema.Schema{
			Type: schema.TypeSet,
			Elem: basicBoolSchema("BoolSetField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "set",
			SdkFieldName: "BoolSetField",
		},
	},
	"int_set": {
		Schema: schema.Schema{
			Type: schema.TypeSet,
			Elem: basicIntSchema("IntSetField", false, false),
		},
		Metadata: Metadata{
			SchemaType:   "set",
			SdkFieldName: "IntSetField",
		},
	},
	"struct_set": {
		Schema: basicStructSchema("set"),
		Metadata: Metadata{
			SchemaType:   "set",
			SdkFieldName: "StructSet",
			ReflectType:  reflect.TypeOf(testNestedStruct{}),
		},
	},
	"deep_nested_struct": {
		Schema: mixedStructSchema(),
		Metadata: Metadata{
			SchemaType:   "struct",
			SdkFieldName: "DeepNestedStruct",
			ReflectType:  reflect.TypeOf(testDeepNestedStruct{}),
		},
	},
}

func TestGetSchemaFromExtendedSchema(t *testing.T) {
	obs := GetSchemaFromExtendedSchema(testExtendedSchema)
	assertSchemaEqual(t, testSchema, obs)
}

func assertSchemaEqual(t *testing.T, expected, actual map[string]*schema.Schema) {
	for k, v := range actual {
		assert.Contains(t, expected, k, "unexpected schema key %s", k)
		expectedSchema := expected[k]
		assert.Equal(t, expectedSchema.Type, v.Type,
			"expected type %s, actual schema %s", expectedSchema.Type, v.Type)
		if v.Type == schema.TypeList || v.Type == schema.TypeSet {
			if _, ok := v.Elem.(*schema.Resource); ok {
				assertSchemaEqual(t,
					expectedSchema.Elem.(*schema.Resource).Schema,
					v.Elem.(*schema.Resource).Schema)
			} else if _, ok := v.Elem.(*schema.Schema); ok {
				assert.Equal(t, expectedSchema, v)
			} else {
				assert.Fail(t, fmt.Sprintf("unexpected schema element type %s", reflect.TypeOf(v.Elem)))
			}
		} else {
			assert.Equal(t, expectedSchema, v, "schema for key %s not equal", k)
		}
	}
}

func TestStructToSchema(t *testing.T) {
	testStr := "test_string"
	nestStr1, nestStr2, nestStr3 := "nest_str1", "nest_str2", "nest_str3"
	testBool := true
	nestBool1, nestBool2 := true, false
	testInt := int64(123)
	nestInt1, nestInt2, nestInt3 := int64(123), int64(456), int64(789)
	obj := testStruct{
		StringField: &testStr,
		BoolField:   &testBool,
		IntField:    &testInt,
		StructField: &testNestedStruct{
			StringField: &testStr,
			BoolField:   &testBool,
			IntField:    &testInt,
		},
		StringListField: []string{"string_1", "string_2"},
		BoolListField:   []bool{true, false},
		IntListField:    []int64{123, 456},
		StringSetField:  []string{"string_3"},
		BoolSetField:    []bool{true},
		IntSetField:     []int64{789},
		StructList: []testNestedStruct{
			{
				StringField: &nestStr1,
				BoolField:   &nestBool1,
				IntField:    &nestInt1,
			},
			{
				StringField: &nestStr2,
				BoolField:   &nestBool2,
				IntField:    &nestInt2,
			},
		},
		StructSet: []testNestedStruct{
			{
				StringField: &nestStr3,
				BoolField:   &nestBool2,
				IntField:    &nestInt3,
			},
		},
		DeepNestedStruct: &testDeepNestedStruct{
			StringField: &nestStr2,
			BoolList:    []bool{false, true},
			IntSet:      []int64{135},
			StructList: []testNestedStruct{
				{
					StringField: &nestStr1,
					BoolField:   &nestBool2,
					IntField:    &nestInt2,
				},
			},
		},
	}
	d := schema.TestResourceDataRaw(
		t, testSchema, map[string]interface{}{})

	elem := reflect.ValueOf(&obj).Elem()
	err := StructToSchema(elem, d, testExtendedSchema, "", nil)
	assert.NoError(t, err, "unexpected error calling StructToSchema")

	t.Run("Base types", func(t *testing.T) {
		assert.Equal(t, "test_string", d.Get("string_field").(string))
		assert.Equal(t, true, d.Get("bool_field").(bool))
		assert.Equal(t, 123, d.Get("int_field").(int))
	})

	t.Run("Zero values", func(t *testing.T) {
		_, ok := d.GetOk("string_field_nil")
		assert.False(t, ok)
		_, ok = d.GetOk("bool_field_nil")
		assert.False(t, ok)
		_, ok = d.GetOk("int_field_nil")
		assert.False(t, ok)
	})

	t.Run("Zero values OIE", func(t *testing.T) {
		_, ok := d.GetOk("string_field_oie")
		assert.False(t, ok)
		_, ok = d.GetOk("bool_field_oie")
		assert.False(t, ok)
		_, ok = d.GetOk("int_field_oie")
		assert.False(t, ok)
	})

	t.Run("Nested struct", func(t *testing.T) {
		nestedObj := d.Get("struct_field").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "test_string", nestedObj["string_field"].(string))
		assert.Equal(t, true, nestedObj["bool_field"].(bool))
		assert.Equal(t, 123, nestedObj["int_field"].(int))
	})

	t.Run("Base list", func(t *testing.T) {
		stringSlice := d.Get("string_list").([]interface{})
		assert.Equal(t, []interface{}{"string_1", "string_2"}, stringSlice)
		boolSlice := d.Get("bool_list").([]interface{})
		assert.Equal(t, []interface{}{true, false}, boolSlice)
		intSlice := d.Get("int_list").([]interface{})
		assert.Equal(t, []interface{}{123, 456}, intSlice)
	})

	t.Run("Base set", func(t *testing.T) {
		stringSlice := d.Get("string_set").(*schema.Set).List()
		assert.Equal(t, []interface{}{"string_3"}, stringSlice)
		boolSlice := d.Get("bool_set").(*schema.Set).List()
		assert.Equal(t, []interface{}{true}, boolSlice)
		intSlice := d.Get("int_set").(*schema.Set).List()
		assert.Equal(t, []interface{}{789}, intSlice)
	})

	t.Run("Struct list", func(t *testing.T) {
		structSlice := d.Get("struct_list").([]interface{})
		assert.Equal(t, 2, len(structSlice))
		assert.Equal(t, map[string]interface{}{
			"string_field": nestStr1,
			"bool_field":   nestBool1,
			"int_field":    123,
		}, structSlice[0].(map[string]interface{}))
		assert.Equal(t, map[string]interface{}{
			"string_field": nestStr2,
			"bool_field":   nestBool2,
			"int_field":    456,
		}, structSlice[1].(map[string]interface{}))
	})

	t.Run("Struct set", func(t *testing.T) {
		structSlice := d.Get("struct_set").(*schema.Set).List()
		assert.Equal(t, 1, len(structSlice))
		assert.Equal(t, map[string]interface{}{
			"string_field": nestStr3,
			"bool_field":   nestBool2,
			"int_field":    789,
		}, structSlice[0].(map[string]interface{}))
	})

	t.Run("Deep nested struct", func(t *testing.T) {
		deepNestedObj := d.Get("deep_nested_struct").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, nestStr2, deepNestedObj["string_field"].(string))
		assert.Equal(t, []interface{}{false, true}, deepNestedObj["bool_list"].([]interface{}))
		assert.Equal(t, []interface{}{135}, deepNestedObj["int_set"].(*schema.Set).List())
		nestedStructList := deepNestedObj["struct_list"].([]interface{})
		assert.Equal(t, 1, len(nestedStructList))
		assert.Equal(t, map[string]interface{}{
			"string_field": nestStr1,
			"bool_field":   nestBool2,
			"int_field":    456,
		}, nestedStructList[0].(map[string]interface{}))
	})
}

func TestSchemaToStruct(t *testing.T) {
	d := schema.TestResourceDataRaw(
		t, testSchema, map[string]interface{}{
			"string_field": "test_string",
			"bool_field":   true,
			"int_field":    100,
			"struct_field": []interface{}{
				map[string]interface{}{
					"string_field": "nested_string",
					"bool_field":   true,
					"int_field":    1,
				},
			},
			"string_list": []interface{}{"string_1", "string_2"},
			"bool_list":   []interface{}{true, false},
			"int_list":    []interface{}{123, 456},
			"string_set":  []interface{}{"string_3"},
			"bool_set":    []interface{}{true},
			"int_set":     []interface{}{789},
			"struct_list": []interface{}{
				map[string]interface{}{
					"string_field": "nested_string_1",
					"bool_field":   true,
					"int_field":    1,
				},
				map[string]interface{}{
					"string_field": "nested_string_2",
					"bool_field":   true,
					"int_field":    2,
				},
			},
			"struct_set": []interface{}{
				map[string]interface{}{
					"string_field": "nested_string_3",
					"bool_field":   false,
					"int_field":    3,
				},
			},
			"deep_nested_struct": []interface{}{
				map[string]interface{}{
					"string_field": "nested_string_1",
					"bool_list":    []interface{}{true, false},
					"int_set":      []interface{}{135},
					"struct_list": []interface{}{
						map[string]interface{}{
							"string_field": "nested_string_2",
							"bool_field":   true,
							"int_field":    1,
						},
					},
				},
			},
		})

	obj := testStruct{}
	elem := reflect.ValueOf(&obj).Elem()
	err := SchemaToStruct(elem, d, testExtendedSchema, "", nil)
	assert.NoError(t, err, "unexpected error calling SchemaToStruct")

	nestStr1, nestStr2, nestStr3 := "nested_string_1", "nested_string_2", "nested_string_3"
	trueVal, falseVal := true, false
	intVal1, intVal2, intVal3 := int64(1), int64(2), int64(3)

	t.Run("Base types", func(t *testing.T) {
		assert.Equal(t, "test_string", *obj.StringField)
		assert.Equal(t, true, *obj.BoolField)
		assert.Equal(t, int64(100), *obj.IntField)
	})

	t.Run("Zero values", func(t *testing.T) {
		assert.Equal(t, "", *obj.StringFieldNil)
		assert.Equal(t, false, *obj.BoolFieldNil)
		assert.Equal(t, int64(0), *obj.IntFieldNil)
	})

	t.Run("Zero values OIE", func(t *testing.T) {
		assert.Equal(t, (*string)(nil), obj.StringFieldOIE)
		assert.Equal(t, (*bool)(nil), obj.BoolFieldOIE)
		assert.Equal(t, (*int64)(nil), obj.IntFieldOIE)
	})

	t.Run("Nested struct", func(t *testing.T) {
		assert.Equal(t, "nested_string", *obj.StructField.StringField)
		assert.Equal(t, true, *obj.StructField.BoolField)
		assert.Equal(t, int64(1), *obj.StructField.IntField)
	})

	t.Run("Base list", func(t *testing.T) {
		assert.Equal(t, []string{"string_1", "string_2"}, obj.StringListField)
		assert.Equal(t, []bool{true, false}, obj.BoolListField)
		assert.Equal(t, []int64{123, 456}, obj.IntListField)
	})

	t.Run("Base set", func(t *testing.T) {
		assert.Equal(t, []string{"string_3"}, obj.StringSetField)
		assert.Equal(t, []bool{true}, obj.BoolSetField)
		assert.Equal(t, []int64{789}, obj.IntSetField)
	})

	t.Run("Struct list", func(t *testing.T) {
		assert.Equal(t, 2, len(obj.StructList))
		assert.EqualValues(t, testNestedStruct{
			StringField: &nestStr1,
			BoolField:   &trueVal,
			IntField:    &intVal1,
		}, obj.StructList[0])
		assert.EqualValues(t, testNestedStruct{
			StringField: &nestStr2,
			BoolField:   &trueVal,
			IntField:    &intVal2,
		}, obj.StructList[1])
	})

	t.Run("Struct set", func(t *testing.T) {
		assert.Equal(t, 1, len(obj.StringSetField))
		assert.EqualValues(t, testNestedStruct{
			StringField: &nestStr3,
			BoolField:   &falseVal,
			IntField:    &intVal3,
		}, obj.StructSet[0])
	})

	t.Run("Deep nested struct", func(t *testing.T) {
		assert.EqualValues(t, &testDeepNestedStruct{
			StringField: &nestStr1,
			BoolList:    []bool{true, false},
			IntSet:      []int64{135},
			StructList: []testNestedStruct{{
				StringField: &nestStr2,
				BoolField:   &trueVal,
				IntField:    &intVal1,
			}},
		}, obj.DeepNestedStruct)
	})
}

func TestSchemaToStructEmptySlice(t *testing.T) {
	d := schema.TestResourceDataRaw(
		t, testSchema, map[string]interface{}{
			"bool_list":   []interface{}{},
			"int_set":     []interface{}{},
			"struct_list": []interface{}{},
		})

	obj := testDeepNestedStruct{}
	elem := reflect.ValueOf(&obj).Elem()
	testSch := mixedStructSchema().Elem.(*ExtendedResource).Schema
	err := SchemaToStruct(elem, d, testSch, "", nil)
	assert.NoError(t, err, "unexpected error calling SchemaToStruct")

	t.Run("Empty bool list", func(t *testing.T) {
		assert.NotNil(t, obj.BoolList)
		assert.Equal(t, 0, len(obj.BoolList))
	})

	t.Run("Empty int set", func(t *testing.T) {
		assert.NotNil(t, obj.IntSet)
		assert.Equal(t, 0, len(obj.IntSet))
	})

	t.Run("Struct list", func(t *testing.T) {
		assert.NotNil(t, obj.StructList)
		assert.Equal(t, 0, len(obj.StructList))
	})
}

func TestSchemaToStructNilStruct(t *testing.T) {
	d := schema.TestResourceDataRaw(
		t, testSchema, map[string]interface{}{
			"struct_field": []interface{}{nil},
		})

	obj := testStruct{}
	elem := reflect.ValueOf(&obj).Elem()
	err := SchemaToStruct(elem, d, testExtendedSchema, "", nil)
	assert.NoError(t, err, "unexpected error calling SchemaToStruct")

	t.Run("Struct field with Nil", func(t *testing.T) {
		assert.Nil(t, obj.StructField)
	})
}
