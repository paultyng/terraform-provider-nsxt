#!/usr/bin/python
# -*- coding: latin-1 -*-

# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: MPL-2.0s

# Generating Resources:
# 1. Run the following command: python policygen.py <resources struct name>
# 2. Format generated go files with: go fmt
# 3. The generated files end up in your current directory; move them to the proper directory.
# 4. Add the resource to provider.go.
# 5. Make sure go is happy with the format by running golint and go vet on the source files.
# 6. Run the generated test(s) to ensure they pass.

import sys
import re
import os
import json

PACKAGE_NAME = "nsxt"

# TODO: clone when opensourced
# TODO: support non-infra resources/data sources
try:
    GO_PATH = os.environ['GOPATH']
except Exception:
    print("Please set GOPATH in your environment")
    exit(2)

BASE_SDK_PATH = "%s/src/github.com/vmware/vsphere-automation-sdk-go/services/nsxt" % GO_PATH
NSX_SPEC_PATH = "nsx_policy_vapi.json"
STRUCTS_FILE = "%s/model/ModelPackageTypes.go" % BASE_SDK_PATH
TEMPLATE_RESOURCE_FILE = "resource_nsxt_policy_template"
TEMPLATE_RESOURCE_TEST_FILE = "resource_nsxt_policy_test_template"
TEMPLATE_RESOURCE_DOC_FILE = "resource_nsxt_policy_doc_template"
TEMPLATE_DATA_SOURCE_FILE = "data_source_nsxt_policy_template"
TEMPLATE_DATA_SOURCE_TEST_FILE = "data_source_nsxt_policy_test_template"
TEMPLATE_DATA_SOURCE_DOC_FILE = "data_source_nsxt_policy_doc_template"

DONT_SPLIT_US = ["IP", "LB", "SSL", "TCP", "UDP"]

# Resource-specific attributes are either in the beginning or the end
FIRST_COMMON_ATTR = "Links []ResourceLink"
LAST_COMMON_ATTR = "Overridden *bool"

TYPE_MAP = {"string": "schema.TypeString",
            "integer": "schema.TypeInt",
            "boolean": "schema.TypeBool"}

DEFAULT_VALUES_MAP = {"string": '"test"',
                      "int32": 2,
                      "int64": 2,
                      "bool": 'true'}

TEST_CREATE_VALUES_MAP = {"string": '"test-create"',
                          "int32": '"2"',
                          "int64": '"2"',
                          "bool": '"true"'}

TEST_UPDATE_VALUES_MAP = {"string": '"test-update"',
                          "int32": '"5"',
                          "int64": '"5"',
                          "bool": '"false"'}

TYPECAST_MAP = {"int64": "int", "int32": "int"}

# Metadata will contain all attributes of the given object, including nested ones
# Parent attribute will have `object_type` key set to indicate model type of the child object
# Child attributes will have `parent` key set
metadata = {}
api_spec = {}

def to_lower_separated(name):
    tmp = re.sub(r'([A-Z])', r'_\1', name).lower()
    for abbr in DONT_SPLIT_US:
        split_abbr = re.sub(r'([A-Z])', r'_\1', abbr).lower()
        tmp = tmp.replace(split_abbr[1:], abbr.lower())

    return tmp[1:]

def lowercase_first(name):
    return name[:1].lower() + name[1:]

def name_to_upper(name):
    return name.title().replace('_', '')

def to_upper_separated(name):
    tmp = re.sub(r'([A-Z])', r'_\1', name).upper()[1:]
    for abbr in DONT_SPLIT_US:
        split_abbr = re.sub(r'([A-Z])', r'_\1', abbr)
        tmp = tmp.replace(split_abbr[1:], abbr)

    return tmp

def get_const_values_name(attr_name):
    return "%s%sValues" % (lowercase_first(metadata['resource']), attr_name)

def build_enums():

    result = ""
    for attr in metadata['attrs']:
        if not attr['const_needed']:
            continue

        result += "\nvar %s = []string{\n" % get_const_values_name(attr['name'])
        const_candidate = "%s_%s" % (metadata['resource'], to_upper_separated(attr['name']))
        for const in metadata['constants'].keys():
            if const.startswith(const_candidate):
                result += "model.%s,\n" % const

        result += "}\n"

    return result

def get_value_for_const(attr, num=1):
    const_candidate = "%s_%s" % (metadata['resource'], to_upper_separated(attr['name']))
    for name, value in metadata['constants'].items():
        if name.startswith(const_candidate):
            num -= 1
            if num == 0:
                return value


def convert_value_to_hcl(attr_type, attr_value):
    if attr_type == "string":
        return "\"%s\"" % attr_value

    if attr_type == "bool":
        if attr_value:
            return "true"
        return "false"

    # int and other
    return attr_value

def build_schema_attr(attr):

    result = ""
    is_array = False
    is_object = False
    attr['const_needed'] = False
    attr['object_type'] = ""
    attr_type = attr['type']
    print(attr_type)
    # Computed and required attributes are not ref. At the moment we can't
    # distinguish between them and assume required, since this will necessarily
    # error out in test.
    # TODO: parse spec json for this purpose

    optional = not attr['required']
    if attr['list'] and attr_type != 'object':
        # Handle arrays. By default, arrays are translated to sets
        # assuming in most cases order is not significant. When order is
        # significant, these should be changed to lists
        attr['helper'] = "%sListFromSchemaSet" % name_to_upper(attr_type)
        is_array = True

    if attr_type == 'object':
        load_resource_metadata(attr_type, attr['name'])
        attr['object_type'] = attr_type
        is_object = True
        is_array = True

    result += "\"%s\": {\n" % attr['schema_name']

    if is_array:
        result += "Type:        schema.TypeList,\n"
    else:
        result += "Type:        %s,\n" % attr_type

    #TODO: concatenate multi-string comment, and remove enum specs
    #if attr['description']:
    #    result += "Description: \"%s\"," % attr['description']

    const_candidate = "%s_%s" % (metadata['resource'], to_upper_separated(attr['name']))
    if attr['type'] == "string":
        for const in metadata['constants'].keys():
            if const.startswith(const_candidate):
                attr['const_needed'] = True

    validation = None
    if attr['const_needed']:
        validation = "ValidateFunc: validation.StringInSlice(%s, false),\n" % get_const_values_name(attr['name'])

    if is_array:
        if is_object:
            result += "Elem: &schema.Resource{\n"
            result += "Schema: map[string]*schema.Schema{\n"
            result += build_schema_attrs(attr['name'])
            result +=  "},\n"
        else:
            result += "Elem: &schema.Schema{\n"
            result += "Type:        %s,\n" % attr_type
            if validation:
                result += validation
        result +=  "},\n"
    elif validation:
        result += validation

    if optional:
        result += "Optional:    true,\n"
    else:
        result += "Required:    true,\n"

    if attr['default']:
        result += "Default:     %s,\n" % convert_value_to_hcl(attr['type'], attr['default'])

    result += "},\n"

    return result

def build_get_attr_from_schema(attr):
    if attr['object_type'] != "":
        # sub-clause
        nameBase = lowercase_first(attr['name'])
        result = '%sList := d.Get("%s").(%s)\n' % (
                nameBase,
                attr['schema_name'],
                '[]interface{}')
        if attr["list"]:
            result += 'var %s []model.%s\n' % (nameBase, attr['object_type'])
        else:
            result += 'var %s *model.%s\n' % (nameBase, attr['object_type'])

        result += 'for _, item := range %sList {\n' % nameBase
        result += 'data := item.(map[string]interface{})\n'
        for childAttr in metadata['attrs']:
            if childAttr['parent'] == attr['name']:
                result += build_get_attr_from_schema(childAttr)

        result += 'obj := model.%s{\n' % attr['object_type']
        result += build_set_attrs_in_obj(attr['name'])
        result += '}\n'
        if attr["list"]:
            result += '%s = append(%s, obj)\n}\n' % (nameBase, nameBase)
        else:
            result += '%s = &obj\nbreak}\n' % nameBase

        return result

    if 'helper' in attr:
        # helper function name already computed - this is the case for arrays
        return '%s := get%s(d, "%s")\n' % (
            lowercase_first(attr['name']),
            attr['helper'],
            attr['schema_name'])

    if attr['parent'] != "":
        if attr['type'] in TYPECAST_MAP:
            # type casting is needed
            return '%s := %s(data["%s"].(%s))\n' % (
                     lowercase_first(attr['name']),
                     attr['type'],
                     attr['schema_name'],
                     TYPECAST_MAP[attr['type']])
        return '%s := data["%s"].(%s)\n' % (
                     lowercase_first(attr['name']),
                     attr['schema_name'],
                     attr['type'])

    if attr['type'] in TYPECAST_MAP:
        # type casting is needed
        return '%s := %s(d.Get("%s").(%s))\n' % (
                 lowercase_first(attr['name']),
                 attr['type'],
                 attr['schema_name'],
                 TYPECAST_MAP[attr['type']])
    return '%s := d.Get("%s").(%s)\n' % (
                 lowercase_first(attr['name']),
                 attr['schema_name'],
                 attr['type'])


def build_enum_var_name(resource, attr, value):
    return "model.%s_%s_%s" % (resource, to_upper_separated(attr), to_upper_separated(value))

def standartize(v):
    if isinstance(v, unicode):
        return v.encode('ascii','ignore')
    return v

def load_resource_metadata(resource, parent=""):

    stage = ""
    description = ""
    api_spec = load_json_spec(resource)
    for attr, attr_spec in api_spec.items():
        schema_name = standartize(attr)
        sdk_name = name_to_upper(schema_name)
        attr_type = standartize(attr_spec.get('type'))
        if 'enum' in attr_spec:
            for enum_value in attr_spec['enum']:
                enum_value = standartize(enum_value)
                var_name = build_enum_var_name(resource, schema_name, enum_value)
                metadata['constants'][var_name] = enum_value

        print(schema_name)
        print(attr_type)
        if attr_type not in TYPE_MAP:
                if schema_name.endswith('es'):
                    schema_name = schema_name[:-2]
                if schema_name.endswith('s'):
                    schema_name = schema_name[:-1]

        default = standartize(attr_spec.get('default'))
        if attr_spec.get('x-deprecated'):
            print("Skipping deprecated attribute %s" % schema_name)
            continue

        full_type = TYPE_MAP.get(attr_type)
        is_ref = False
        is_list = False
        if attr_type == 'None':
            if '$ref' in attr_spec:
                print("REF")
                is_ref = True
                full_type = attr_spec['$ref'].split('/')[:1]

        if attr_type == 'array':
            is_list = True
            items = attr_spec.get('items')
            if '$ref' in items:
                full_type = standartize(items['$ref'].split('/')[:1])
            if 'type' in items:
                full_type = standartize(items['type'])

        metadata['attrs'].append({'name': sdk_name,
                                  'parent': parent,
                                  'description': standartize(attr_spec.get('description')),
                                  'type': full_type,
                                  'list': is_list,
                                  'ref': is_ref,
                                  'schema_name': schema_name,
                                  'default': default,
                                  'required': False})
        metadata["name_map"][sdk_name] = schema_name
        print(metadata)


def build_schema_attrs(parent=""):
    result = ""
    for attr in metadata['attrs']:
        if attr['parent'] == parent:
            result += build_schema_attr(attr)

    return result


def build_get_attrs_from_schema(parent=""):
    result = ""
    for attr in metadata['attrs']:
        if parent != attr['parent']:
            continue
        result += build_get_attr_from_schema(attr)

    return result


def build_set_attrs_in_obj(parent=""):
    result = ""
    for attr in metadata['attrs']:
        if attr['parent'] != parent:
            continue
        if attr['list'] or attr['object_type'] != "":
            result += "%s: %s,\n" % (attr['name'], lowercase_first(attr['name']))
        else:
            result += "%s: &%s,\n" % (attr['name'], lowercase_first(attr['name']))

    return result


def build_set_obj_attrs_in_schema(parent=""):
    result = ""
    for attr in metadata['attrs']:
        if attr['parent'] != parent:
            continue
        if attr['object_type'] != "":
            nameBase = lowercase_first(attr['name'])
            result += 'var %sList []map[string]interface{}\n' % nameBase
            if attr['list']:
                result += 'for _, item := range obj.%s {\n' % attr['name']
            else:
                result += 'item := obj.%s\n' % attr['name']
            result += 'data := make(map[string]interface{})\n'
            result += build_set_obj_attrs_in_schema(attr['name'])
            result += '%sList = append(%sList, data)\n' % (nameBase, nameBase)
            if attr['list']:
                result += '}\n'
            result += 'd.Set("%s", %sList)\n'% (attr['schema_name'], nameBase)
            return result

        if attr['parent']:
            result += 'data["%s"] = item.%s\n'% (attr['schema_name'], attr['name'])
        else:
            result += 'd.Set("%s", obj.%s)\n'% (attr['schema_name'], attr['name'])

    return result


def build_test_attrs(required_only=False, parent="", indent="  "):

    result = ""
    for attr in metadata['attrs']:
        if required_only and not attr['required']:
            continue
        if parent != attr['parent']:
            continue
        if attr['object_type'] != "":
            result += "\n%s%s {\n" % (indent, attr['schema_name'])
            result += build_test_attrs(required_only, attr['name'], indent + "  ")
            result += "%s}\n\n" % indent
            continue

        result += "%s%s = " % (indent, attr['schema_name'])
        if attr['list']:
            result += '['
        if attr['type'] == 'string':
            result += '"%s"'
        else:
            result += '%s'
        if attr['list']:
            result += ']'

        result += "\n"

    return result


def build_test_attrs_map(is_create=True):

    result = ""
    values_map = TEST_CREATE_VALUES_MAP if is_create else TEST_UPDATE_VALUES_MAP
    for attr in metadata['attrs']:
        value = ""
        if attr['const_needed']:
            num = 1 if is_create else 2
            value = get_value_for_const(attr, num)
        elif attr['type'] in values_map:
            value = values_map[attr['type']]
        if value:
            result += '"%s": %s,\n' % (attr["schema_name"], value)

    return result


def build_check_test_attrs(is_create=True):

    result = ""
    for attr in metadata['attrs']:
        attrStr = attr['schema_name']
        if attr['parent']:
            parentSchemaName = metadata['name_map'][attr['parent']]
            attrStr = "%s.0.%s" % (parentSchemaName, attr['schema_name'])
        if attr['object_type']:
            result += 'resource.TestCheckResourceAttr(testResourceName, "%s.#", "1"),\n' % attrStr
            continue
        if attr['list']:
            result += 'resource.TestCheckResourceAttr(testResourceName, "%s.0", ' % attrStr
        else:
            result += 'resource.TestCheckResourceAttr(testResourceName, "%s", ' % attrStr
        result += 'accTestPolicy%s%sAttributes["%s"]),\n' % (metadata['resource'], "Create" if is_create else "Update", attr['schema_name'])

    return result


def build_doc_attrs():
    doc_attrs = ""
    for attr in metadata['attrs']:
        if attr['object_type'] or attr['parent']:
            continue
        optional = attr['ref'] or attr['list']
        if attr['const_needed']:
            value = get_value_for_const(attr)
        else:
            if attr['type'] in DEFAULT_VALUES_MAP:
                value = DEFAULT_VALUES_MAP[attr['type']]
            else:
                value = "FILL VALUE FOR %s" % attr['type']

        if not value:
            value = "FILL ENUM VALUE"
        doc_attrs += "%s = %s\n" % (attr['schema_name'], value)

    return doc_attrs


def build_doc_attrs_reference(parent="", indent=""):
    result = ""

    for attr in metadata['attrs']:
        if attr['parent'] != parent:
            continue
        result += "%s* `%s` - (%s) %s\n" % (indent, attr['schema_name'],
                                   "Required" if attr['required'] else "Optional", attr['description'])
        if attr['object_type']:
            result += build_doc_attrs_reference(attr['name'], indent + "  ")
    return result


def build_test_attrs_sprintf(required_only=False):
    result = ""

    for attr in metadata['attrs']:
        if required_only and not attr['required']:
            continue
        if attr['object_type']:
            continue
        result += ', attrMap["%s"]' % attr['schema_name']
    return result


def generate_replace_targets():
    metadata['schema_attrs'] = build_schema_attrs()
    metadata['enums'] = build_enums()
    metadata['set_attrs_in_obj'] = build_set_attrs_in_obj()
    metadata['get_attrs_from_schema'] = build_get_attrs_from_schema()
    metadata['set_obj_attrs_in_schema'] = build_set_obj_attrs_in_schema()
    metadata['doc_attrs'] = build_doc_attrs()
    metadata['doc_attrs_reference'] = build_doc_attrs_reference()
    metadata['test_attrs'] = build_test_attrs()
    metadata['test_required_attrs'] = build_test_attrs(True)
    metadata['test_attrs_create'] = build_test_attrs_map()
    metadata['test_attrs_update'] = build_test_attrs_map(False)
    metadata['check_attrs_create'] = build_check_test_attrs()
    metadata['check_attrs_update'] = build_check_test_attrs(False)
    metadata['test_attrs_sprintf'] = build_test_attrs_sprintf()
    metadata['test_required_attrs_sprintf'] = build_test_attrs_sprintf(True)


def replace_templates(line):
    result = line.replace("<!RESOURCE!>", metadata['resource'])
    result = result.replace("<!RESOURCES!>", "%ss" % metadata['resource'])
    result = result.replace("<!MODULE!>", metadata['module'])
    result = result.replace("<!SCHEMA_ATTRS!>", metadata['schema_attrs'])
    result = result.replace("<!GET_ATTRS_FROM_SCHEMA!>", metadata['get_attrs_from_schema'])
    result = result.replace("<!SET_ATTRS_IN_OBJ!>", metadata['set_attrs_in_obj'])
    result = result.replace("<!SET_OBJ_ATTRS_IN_SCHEMA!>", metadata['set_obj_attrs_in_schema'])
    result = result.replace("<!ENUMS!>", metadata['enums'])
    result = result.replace("<!resource_lower!>", metadata['resource_lower'])
    result = result.replace("<!resource-lower!>", metadata['resource-lower'])
    result = result.replace("<!TEST_ATTRS!>", metadata['test_attrs'])
    result = result.replace("<!TEST_REQUIRED_ATTRS!>", metadata['test_required_attrs'])
    result = result.replace("<!TEST_ATTRS_CREATE!>", metadata['test_attrs_create'])
    result = result.replace("<!TEST_ATTRS_UPDATE!>", metadata['test_attrs_update'])
    result = result.replace("<!TEST_ATTRS_SPRINTF!>", metadata['test_attrs_sprintf'])
    result = result.replace("<!TEST_REQUIRED_ATTRS_SPRINTF!>", metadata['test_required_attrs_sprintf'])
    result = result.replace("<!CHECK_ATTRS_CREATE!>", metadata['check_attrs_create'])
    result = result.replace("<!CHECK_ATTRS_UPDATE!>", metadata['check_attrs_update'])
    result = result.replace("<!DOC_ATTRS!>", metadata['doc_attrs'])
    result = result.replace("<!DOC_ATTRS_REFERENCE!>", metadata['doc_attrs_reference'])
    result = result.replace("PolicyPolicy", "Policy")
    return result

def load_json_spec(resource):

    with open(NSX_SPEC_PATH, 'r') as f:
        spec = json.load(f)

        defs = spec["definitions"]
        if resource not in defs:
            return {}

        if "allOf" not in defs[resource]:
            if "properties" in defs[resource]:
                # no inheritance
                return defs[resource]["properties"]
            return {}
        if len(defs[resource]["allOf"]) == 2:
            # object inheritance
            if "properties" not in defs[resource]["allOf"][1]:
                return {}
            return defs[resource]["allOf"][1]["properties"]

        return {}

def main():

    if len(sys.argv) < 2:
        print("Usage: %s <resource name> [data]" % sys.argv[0])
        sys.exit()

    generate_data_source = False
    if len(sys.argv) == 3 and sys.argv[2] == 'data':
        generate_data_source = True

    resource = sys.argv[1]
    metadata['resource'] = resource
    metadata['module'] = resource[0].lower() + resource[1:] + 's'
    resource_lower = to_lower_separated(resource)
    if resource_lower.startswith("policy_"):
        # remove double "policy" indication
        resource_lower = resource_lower[len("policy_"):]
    metadata['resource_lower'] = resource_lower
    metadata['resource-lower'] = resource_lower.replace('_','-')
    metadata["constants"] = {}
    metadata["attrs"] = []
    metadata["name_map"] = {}
    print("Building resource from %s" % resource)

    print("Loading metadata..")
    load_resource_metadata(resource)
    generate_replace_targets()

    spec = {}
    if generate_data_source:
        spec = {TEMPLATE_DATA_SOURCE_FILE: "data_source_nsxt_policy_%s.go" % resource_lower,
                TEMPLATE_DATA_SOURCE_TEST_FILE: "data_source_nsxt_policy_%s_test.go" % resource_lower,
                TEMPLATE_DATA_SOURCE_DOC_FILE: "policy_%s.html.markdown" % resource_lower}

    else:
        spec = {TEMPLATE_RESOURCE_FILE: "resource_nsxt_policy_%s.go" % resource_lower,
                TEMPLATE_RESOURCE_TEST_FILE: "resource_nsxt_policy_%s_test.go" % resource_lower,
                TEMPLATE_RESOURCE_DOC_FILE: "policy_%s.html.markdown" % resource_lower}

    for source, target in spec.items():
        print("Generating %s.." % target)
        with open(source, 'r') as f:
            lines = f.readlines()

        with open(target, 'w') as f:
            for line in lines:
                f.write(replace_templates(line))

    print("Formatting..")
    os.system("go fmt")
    print("Done.")

main()
