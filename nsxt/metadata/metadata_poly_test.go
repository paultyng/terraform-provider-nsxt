// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package metadata

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	vapiBindings_ "github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
)

type testCatStruct struct {
	Age  *int64
	Name *string
	Type *string
}

func testCatStructBindingType() vapiBindings_.BindingType {
	fields := make(map[string]vapiBindings_.BindingType)
	fieldNameMap := make(map[string]string)
	fields["age"] = vapiBindings_.NewOptionalType(vapiBindings_.NewIntegerType())
	fieldNameMap["age"] = "Age"
	fields["name"] = vapiBindings_.NewOptionalType(vapiBindings_.NewStringType())
	fieldNameMap["name"] = "Name"
	fields["type"] = vapiBindings_.NewOptionalType(vapiBindings_.NewStringType())
	fieldNameMap["type"] = "Type"
	var validators = []vapiBindings_.Validator{}
	return vapiBindings_.NewStructType("com.vmware.nsx.fake.cat", fields,
		reflect.TypeOf(testCatStruct{}), fieldNameMap, validators)
}

func TestCatStructBinding(t *testing.T) {
	age := int64(123)
	name := "John"
	rType := "FakeCat"
	obj := testCatStruct{
		Age:  &age,
		Name: &name,
		Type: &rType,
	}
	converter := vapiBindings_.NewTypeConverter()
	dv, err := converter.ConvertToVapi(obj, testCatStructBindingType())
	assert.Nil(t, err)
	obs, err := converter.ConvertToGolang(dv, testCatStructBindingType())
	assert.Nil(t, err)
	assert.Equal(t, obj, obs.(testCatStruct))
}

type testCoffeeStruct struct {
	IsDecaf *bool
	Name    *string
	Type    *string
}

func testCoffeeStructBindingType() vapiBindings_.BindingType {
	fields := make(map[string]vapiBindings_.BindingType)
	fieldNameMap := make(map[string]string)
	fields["is_decaf"] = vapiBindings_.NewOptionalType(vapiBindings_.NewBooleanType())
	fieldNameMap["is_decaf"] = "IsDecaf"
	fields["name"] = vapiBindings_.NewOptionalType(vapiBindings_.NewStringType())
	fieldNameMap["name"] = "Name"
	fields["type"] = vapiBindings_.NewOptionalType(vapiBindings_.NewStringType())
	fieldNameMap["type"] = "Type"
	var validators = []vapiBindings_.Validator{}
	return vapiBindings_.NewStructType("com.vmware.nsx.fake.coffee", fields,
		reflect.TypeOf(testCoffeeStruct{}), fieldNameMap, validators)
}

func TestCoffeeStructBinding(t *testing.T) {
	decaf := false
	name := "Latte"
	rType := "FakeCoffee"
	obj := testCoffeeStruct{
		IsDecaf: &decaf,
		Name:    &name,
		Type:    &rType,
	}
	converter := vapiBindings_.NewTypeConverter()
	dv, err := converter.ConvertToVapi(obj, testCoffeeStructBindingType())
	assert.Nil(t, err)
	obs, err := converter.ConvertToGolang(dv, testCoffeeStructBindingType())
	assert.Nil(t, err)
	assert.Equal(t, obj, obs.(testCoffeeStruct))
}

type testPolyStruct struct {
	PolyStruct *data.StructValue
}

type testPolyListStruct struct {
	PolyList []*data.StructValue
}

func testPolyStructNestedSchema(t string) map[string]*schema.Schema {
	schemaType := schema.TypeList
	maxItems := 0
	if t == "set" {
		schemaType = schema.TypeSet
	} else if t == "struct" {
		maxItems = 1
	}

	return map[string]*schema.Schema{
		"poly_struct": {
			Type:     schemaType,
			MaxItems: maxItems,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"cat": {
						Type:          schema.TypeList,
						MaxItems:      1,
						ConflictsWith: []string{"coffee"},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type: schema.TypeString,
								},
								"age": {
									Type: schema.TypeInt,
								},
							},
						},
					},
					"coffee": {
						Type:          schema.TypeList,
						MaxItems:      1,
						ConflictsWith: []string{"cat"},
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"name": {
									Type: schema.TypeString,
								},
								"is_decaf": {
									Type: schema.TypeBool,
								},
							},
						},
					},
				},
			},
		},
	}
}

func testPolyStructNestedExtSchema(t, sdkName string) map[string]*ExtendedSchema {
	typeIdentifier := TypeIdentifier{
		SdkName:      "Type",
		APIFieldName: "type",
	}

	schemaType := schema.TypeList
	maxItems := 0
	if t == "set" {
		schemaType = schema.TypeSet
	} else if t == "struct" {
		maxItems = 1
	}
	return map[string]*ExtendedSchema{
		"poly_struct": {
			Schema: schema.Schema{
				Type:     schemaType,
				MaxItems: maxItems,
				Elem: &ExtendedResource{
					Schema: map[string]*ExtendedSchema{
						"cat": {
							Schema: schema.Schema{
								Type:          schema.TypeList,
								MaxItems:      1,
								ConflictsWith: []string{"coffee"},
								Elem: &ExtendedResource{
									Schema: map[string]*ExtendedSchema{
										"name": basicStringSchema("Name", false, false),
										"age":  basicIntSchema("Age", false, false),
									},
								},
							},
							Metadata: Metadata{
								SchemaType:     "struct",
								ReflectType:    reflect.TypeOf(testCatStruct{}),
								BindingType:    testCatStructBindingType(),
								TypeIdentifier: typeIdentifier,
							},
						},
						"coffee": {
							Schema: schema.Schema{
								Type:          schema.TypeList,
								MaxItems:      1,
								ConflictsWith: []string{"cat"},
								Elem: &ExtendedResource{
									Schema: map[string]*ExtendedSchema{
										"name":     basicStringSchema("Name", false, false),
										"is_decaf": basicBoolSchema("IsDecaf", false, false),
									},
								},
							},
							Metadata: Metadata{
								SchemaType:     "struct",
								ReflectType:    reflect.TypeOf(testCoffeeStruct{}),
								BindingType:    testCoffeeStructBindingType(),
								TypeIdentifier: typeIdentifier,
							},
						},
					},
				},
			},
			Metadata: Metadata{
				SchemaType:      t,
				SdkFieldName:    sdkName,
				PolymorphicType: PolymorphicTypeNested,
				ResourceTypeMap: map[string]string{
					"cat":    "FakeCat",
					"coffee": "FakeCoffee",
				},
				TypeIdentifier: typeIdentifier,
			},
		},
	}
}

func TestPolyStructToNestedSchema(t *testing.T) {
	t.Run("cat struct", func(t *testing.T) {
		name := "matcha"
		rType := "FakeCat"
		age := int64(1)
		catObj := testCatStruct{
			Age:  &age,
			Name: &name,
			Type: &rType,
		}
		obj := testPolyStruct{}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(catObj, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyStruct = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructNestedExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("poly_struct"), 1)
		polyData := d.Get("poly_struct").([]interface{})[0].(map[string]interface{})
		assert.Len(t, polyData["cat"], 1)
		assert.Len(t, polyData["coffee"], 0)
		assert.Equal(t, map[string]interface{}{
			"name": name,
			"age":  1,
		}, polyData["cat"].([]interface{})[0].(map[string]interface{}))
	})

	t.Run("coffee struct", func(t *testing.T) {
		name := "latte"
		rType := "FakeCoffee"
		isDecaf := true
		coffeeObj := testCoffeeStruct{
			IsDecaf: &isDecaf,
			Name:    &name,
			Type:    &rType,
		}
		obj := testPolyStruct{}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(coffeeObj, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyStruct = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructNestedExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("poly_struct"), 1)
		polyData := d.Get("poly_struct").([]interface{})[0].(map[string]interface{})
		assert.Len(t, polyData["cat"], 0)
		assert.Len(t, polyData["coffee"], 1)
		assert.Equal(t, map[string]interface{}{
			"name":     name,
			"is_decaf": true,
		}, polyData["coffee"].([]interface{})[0].(map[string]interface{}))
	})

	t.Run("mixed list", func(t *testing.T) {
		catName := "oolong"
		coffeeName := "mocha"
		catResType := "FakeCat"
		coffeeResType := "FakeCoffee"
		isDecaf := false
		age := int64(2)
		catObj := testCatStruct{
			Age:  &age,
			Name: &catName,
			Type: &catResType,
		}
		coffeeObj := testCoffeeStruct{
			IsDecaf: &isDecaf,
			Name:    &coffeeName,
			Type:    &coffeeResType,
		}
		obj := testPolyListStruct{
			PolyList: make([]*data.StructValue, 2),
		}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(coffeeObj, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[0] = dv.(*data.StructValue)
		dv, errors = converter.ConvertToVapi(catObj, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[1] = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructNestedExtSchema("list", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("poly_struct"), 2)
		// idx 0: coffee
		coffeeData := d.Get("poly_struct").([]interface{})[0].(map[string]interface{})
		assert.Len(t, coffeeData["cat"], 0)
		assert.Len(t, coffeeData["coffee"], 1)
		assert.Equal(t, map[string]interface{}{
			"name":     coffeeName,
			"is_decaf": false,
		}, coffeeData["coffee"].([]interface{})[0].(map[string]interface{}))
		// idx 1: cat
		catData := d.Get("poly_struct").([]interface{})[1].(map[string]interface{})
		assert.Len(t, catData["cat"], 1)
		assert.Len(t, catData["coffee"], 0)
		assert.Equal(t, map[string]interface{}{
			"name": catName,
			"age":  2,
		}, catData["cat"].([]interface{})[0].(map[string]interface{}))
	})

	t.Run("mixed set", func(t *testing.T) {
		catName := "oolong"
		coffeeName := "mocha"
		catResType := "FakeCat"
		coffeeResType := "FakeCoffee"
		isDecaf := false
		age := int64(2)
		catObj := testCatStruct{
			Age:  &age,
			Name: &catName,
			Type: &catResType,
		}
		coffeeObj := testCoffeeStruct{
			IsDecaf: &isDecaf,
			Name:    &coffeeName,
			Type:    &coffeeResType,
		}
		obj := testPolyListStruct{
			PolyList: make([]*data.StructValue, 2),
		}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(coffeeObj, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[1] = dv.(*data.StructValue)
		dv, errors = converter.ConvertToVapi(catObj, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[0] = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructNestedExtSchema("set", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("poly_struct"), 2)
		for _, v := range d.Get("poly_struct").([]interface{}) {
			dataVal := v.(map[string]interface{})
			if len(dataVal["cat"].([]interface{})) > 0 {
				assert.Len(t, dataVal["cat"], 1)
				assert.Len(t, dataVal["coffee"], 0)
				assert.Equal(t, map[string]interface{}{
					"name": catName,
					"age":  2,
				}, dataVal["cat"].([]interface{})[0].(map[string]interface{}))
			} else {
				assert.Len(t, dataVal["cat"], 0)
				assert.Len(t, dataVal["coffee"], 1)
				assert.Equal(t, map[string]interface{}{
					"name":     coffeeName,
					"is_decaf": false,
				}, dataVal["coffee"].([]interface{})[0].(map[string]interface{}))
			}
		}
	})
}

func TestNestedSchemaToPolyStruct(t *testing.T) {
	t.Run("cat struct", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{
				"poly_struct": []interface{}{
					map[string]interface{}{
						"cat": []interface{}{
							map[string]interface{}{
								"name": "matcha",
								"age":  1,
							},
						},
					},
				},
			})

		obj := testPolyStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructNestedExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		obs, errors := converter.ConvertToGolang(obj.PolyStruct, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		assert.Equal(t, "matcha", *obs.(testCatStruct).Name)
		assert.Equal(t, int64(1), *obs.(testCatStruct).Age)
		assert.Equal(t, "FakeCat", *obs.(testCatStruct).Type)
	})

	t.Run("coffee struct", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("struct"), map[string]interface{}{
				"poly_struct": []interface{}{
					map[string]interface{}{
						"coffee": []interface{}{
							map[string]interface{}{
								"name":     "latte",
								"is_decaf": true,
							},
						},
					},
				},
			})

		obj := testPolyStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructNestedExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		obs, errors := converter.ConvertToGolang(obj.PolyStruct, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		assert.Equal(t, "latte", *obs.(testCoffeeStruct).Name)
		assert.Equal(t, true, *obs.(testCoffeeStruct).IsDecaf)
		assert.Equal(t, "FakeCoffee", *obs.(testCoffeeStruct).Type)
	})

	t.Run("mixed list", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("list"), map[string]interface{}{
				"poly_struct": []interface{}{
					map[string]interface{}{
						"coffee": []interface{}{
							map[string]interface{}{
								"name":     "latte",
								"is_decaf": true,
							},
						},
					},
					map[string]interface{}{
						"cat": []interface{}{
							map[string]interface{}{
								"name": "matcha",
								"age":  1,
							},
						},
					},
				},
			})

		obj := testPolyListStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructNestedExtSchema("list", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		assert.Len(t, obj.PolyList, 2)
		coffee, errors := converter.ConvertToGolang(obj.PolyList[0], testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		assert.Equal(t, "latte", *coffee.(testCoffeeStruct).Name)
		assert.Equal(t, true, *coffee.(testCoffeeStruct).IsDecaf)
		assert.Equal(t, "FakeCoffee", *coffee.(testCoffeeStruct).Type)
		cat, errors := converter.ConvertToGolang(obj.PolyList[1], testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		assert.Equal(t, "matcha", *cat.(testCatStruct).Name)
		assert.Equal(t, int64(1), *cat.(testCatStruct).Age)
		assert.Equal(t, "FakeCat", *cat.(testCatStruct).Type)
	})

	t.Run("mixed set", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructNestedSchema("set"), map[string]interface{}{
				"poly_struct": []interface{}{
					map[string]interface{}{
						"cat": []interface{}{
							map[string]interface{}{
								"name": "oolong",
								"age":  2,
							},
						},
					},
					map[string]interface{}{
						"coffee": []interface{}{
							map[string]interface{}{
								"name":     "mocha",
								"is_decaf": false,
							},
						},
					},
				},
			})

		obj := testPolyListStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructNestedExtSchema("set", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		assert.Len(t, obj.PolyList, 2)
		for _, item := range obj.PolyList {
			if item.HasField("age") {
				cat, errors := converter.ConvertToGolang(item, testCatStructBindingType())
				assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
				assert.Equal(t, "oolong", *cat.(testCatStruct).Name)
				assert.Equal(t, int64(2), *cat.(testCatStruct).Age)
				assert.Equal(t, "FakeCat", *cat.(testCatStruct).Type)
			} else {
				coffee, errors := converter.ConvertToGolang(item, testCoffeeStructBindingType())
				assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
				assert.Equal(t, "mocha", *coffee.(testCoffeeStruct).Name)
				assert.Equal(t, false, *coffee.(testCoffeeStruct).IsDecaf)
				assert.Equal(t, "FakeCoffee", *coffee.(testCoffeeStruct).Type)
			}
		}
	})
}

func testPolyStructFlattenSchema(t string) map[string]*schema.Schema {
	schemaType := schema.TypeList
	maxItems := 0
	if t == "set" {
		schemaType = schema.TypeSet
	} else if t == "struct" {
		maxItems = 1
	}
	return map[string]*schema.Schema{
		"cat": {
			Type:     schemaType,
			MaxItems: maxItems,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type: schema.TypeString,
					},
					"age": {
						Type: schema.TypeInt,
					},
				},
			},
		},
		"coffee": {
			Type:     schemaType,
			MaxItems: maxItems,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type: schema.TypeString,
					},
					"is_decaf": {
						Type: schema.TypeBool,
					},
				},
			},
		},
	}
}

func testPolyStructFlattenExtSchema(t, sdkName string) map[string]*ExtendedSchema {
	typeIdentifier := TypeIdentifier{
		SdkName:      "Type",
		APIFieldName: "type",
	}
	schemaType := schema.TypeList
	maxItems := 0
	if t == "set" {
		schemaType = schema.TypeSet
	} else if t == "struct" {
		maxItems = 1
	}
	return map[string]*ExtendedSchema{
		"cat": {
			Schema: schema.Schema{
				Type:     schemaType,
				MaxItems: maxItems,
				Elem: &ExtendedResource{
					Schema: map[string]*ExtendedSchema{
						"name": basicStringSchema("Name", false, false),
						"age":  basicIntSchema("Age", false, false),
					},
				},
			},
			Metadata: Metadata{
				SchemaType:      t,
				ReflectType:     reflect.TypeOf(testCatStruct{}),
				BindingType:     testCatStructBindingType(),
				TypeIdentifier:  typeIdentifier,
				PolymorphicType: PolymorphicTypeFlatten,
				ResourceType:    "FakeCat",
				SdkFieldName:    sdkName,
			},
		},
		"coffee": {
			Schema: schema.Schema{
				Type:     schemaType,
				MaxItems: maxItems,
				Elem: &ExtendedResource{
					Schema: map[string]*ExtendedSchema{
						"name":     basicStringSchema("Name", false, false),
						"is_decaf": basicBoolSchema("IsDecaf", false, false),
					},
				},
			},
			Metadata: Metadata{
				SchemaType:      t,
				ReflectType:     reflect.TypeOf(testCoffeeStruct{}),
				BindingType:     testCoffeeStructBindingType(),
				TypeIdentifier:  typeIdentifier,
				PolymorphicType: PolymorphicTypeFlatten,
				ResourceType:    "FakeCoffee",
				SdkFieldName:    sdkName,
			},
		},
	}
}

func TestPolyStructToFlattenSchema(t *testing.T) {
	t.Run("cat struct", func(t *testing.T) {
		name := "matcha"
		rType := "FakeCat"
		age := int64(1)
		catObj := testCatStruct{
			Age:  &age,
			Name: &name,
			Type: &rType,
		}
		obj := testPolyStruct{}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(catObj, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyStruct = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructFlattenExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("cat"), 1)
		assert.Len(t, d.Get("coffee"), 0)
		assert.Equal(t, map[string]interface{}{
			"name": name,
			"age":  1,
		}, d.Get("cat").([]interface{})[0].(map[string]interface{}))
	})

	t.Run("coffee struct", func(t *testing.T) {
		name := "latte"
		rType := "FakeCoffee"
		isDecaf := true
		coffeeObj := testCoffeeStruct{
			IsDecaf: &isDecaf,
			Name:    &name,
			Type:    &rType,
		}
		obj := testPolyStruct{}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(coffeeObj, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyStruct = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("struct"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructFlattenExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")
		assert.Len(t, d.Get("coffee"), 1)
		assert.Len(t, d.Get("cat"), 0)
		assert.Equal(t, map[string]interface{}{
			"name":     name,
			"is_decaf": true,
		}, d.Get("coffee").([]interface{})[0].(map[string]interface{}))
	})

	t.Run("mixed list", func(t *testing.T) {
		catName := "oolong"
		coffeeName := "mocha"
		catResType := "FakeCat"
		coffeeResType := "FakeCoffee"
		isDecaf := false
		age := int64(2)
		catObj := testCatStruct{
			Age:  &age,
			Name: &catName,
			Type: &catResType,
		}
		coffeeObj := testCoffeeStruct{
			IsDecaf: &isDecaf,
			Name:    &coffeeName,
			Type:    &coffeeResType,
		}
		obj := testPolyListStruct{
			PolyList: make([]*data.StructValue, 2),
		}
		converter := vapiBindings_.NewTypeConverter()
		dv, errors := converter.ConvertToVapi(coffeeObj, testCoffeeStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[0] = dv.(*data.StructValue)
		dv, errors = converter.ConvertToVapi(catObj, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		obj.PolyList[1] = dv.(*data.StructValue)
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("list"), map[string]interface{}{})

		elem := reflect.ValueOf(&obj).Elem()
		err := StructToSchema(elem, d, testPolyStructFlattenExtSchema("list", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling StructToSchema")

		assert.Len(t, d.Get("coffee"), 1)
		assert.Equal(t, map[string]interface{}{
			"name":     coffeeName,
			"is_decaf": false,
		}, d.Get("coffee").([]interface{})[0].(map[string]interface{}))

		assert.Len(t, d.Get("cat"), 1)
		assert.Equal(t, map[string]interface{}{
			"name": catName,
			"age":  2,
		}, d.Get("cat").([]interface{})[0].(map[string]interface{}))
	})
}

func TestFlattenSchemaToPolyStruct(t *testing.T) {
	t.Run("cat struct", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("struct"), map[string]interface{}{
				"cat": []interface{}{
					map[string]interface{}{
						"name": "matcha",
						"age":  1,
					},
				},
			})

		obj := testPolyStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructFlattenExtSchema("struct", "PolyStruct"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		obs, errors := converter.ConvertToGolang(obj.PolyStruct, testCatStructBindingType())
		assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
		assert.Equal(t, "matcha", *obs.(testCatStruct).Name)
		assert.Equal(t, int64(1), *obs.(testCatStruct).Age)
		assert.Equal(t, "FakeCat", *obs.(testCatStruct).Type)
	})

	t.Run("error on assigning struct multiple times", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("struct"), map[string]interface{}{
				"cat": []interface{}{
					map[string]interface{}{
						"name": "matcha",
						"age":  1,
					},
				},
				"coffee": []interface{}{
					map[string]interface{}{
						"name":     "mocha",
						"is_decaf": true,
					},
				},
			})

		obj := testPolyStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructFlattenExtSchema("struct", "PolyStruct"), "", nil)
		assert.ErrorContainsf(t, err, "is already set", "expected error raised if same sdk is set twice")
	})

	t.Run("mixed list", func(t *testing.T) {
		d := schema.TestResourceDataRaw(
			t, testPolyStructFlattenSchema("list"), map[string]interface{}{
				"coffee": []interface{}{
					map[string]interface{}{
						"name":     "mocha",
						"is_decaf": true,
					},
				},
				"cat": []interface{}{
					map[string]interface{}{
						"name": "oolong",
						"age":  2,
					},
				},
			})

		obj := testPolyListStruct{}
		elem := reflect.ValueOf(&obj).Elem()
		err := SchemaToStruct(elem, d, testPolyStructFlattenExtSchema("list", "PolyList"), "", nil)
		assert.NoError(t, err, "unexpected error calling SchemaToStruct")

		converter := vapiBindings_.NewTypeConverter()
		assert.Len(t, obj.PolyList, 2)
		for _, item := range obj.PolyList {
			if item.HasField("age") {
				cat, errors := converter.ConvertToGolang(item, testCatStructBindingType())
				assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
				assert.Equal(t, "oolong", *cat.(testCatStruct).Name)
				assert.Equal(t, int64(2), *cat.(testCatStruct).Age)
				assert.Equal(t, "FakeCat", *cat.(testCatStruct).Type)
			} else {
				coffee, errors := converter.ConvertToGolang(item, testCoffeeStructBindingType())
				assert.Nil(t, errors, "unexpected error calling ConvertToGolang")
				assert.Equal(t, "mocha", *coffee.(testCoffeeStruct).Name)
				assert.Equal(t, true, *coffee.(testCoffeeStruct).IsDecaf)
				assert.Equal(t, "FakeCoffee", *coffee.(testCoffeeStruct).Type)
			}
		}
	})
}
