#!/usr/bin/python
# -*- coding: latin-1 -*-

# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: MPL-2.0

import sys
import re

PACKAGE_NAME = "nsxt"
SDK_PACKAGE_NAME = "api"
MANAGER_PACKAGE_NAME = "manager"

IGNORE_ATTRS = ["Links", "Schema", "Self", "Id", "ResourceType", "CreateTime", "CreateUser", "LastModifiedTime", "LastModifiedUser", "SystemOwned", "Protection"]
COMPUTED_ATTRS = ["CreateTime", "CreateUser", "LastModifiedTime", "LastModifiedUser", "SystemOwned"]
COMPUTED_AND_OPTIONAL_ATTRS = ["DisplayName"]
FORCENEW_ATTRS = ["TransportZoneId"]
# TODO: ServiceBindings
VIP_SCHEMA_ATTRS = ["Tags", "SwitchingProfileIds", "Revision", "AddressBindings"]
VIP_GETTER_ATTRS = ["Tags", "SwitchingProfileIds", "AddressBindings"]
VIP_SETTER_ATTRS = VIP_GETTER_ATTRS

TYPE_MAP = {"string": "schema.TypeString",
            "int32": "schema.TypeInt",
            "int64": "schema.TypeInt",
            "bool": "schema.TypeBool"}

TYPECAST_MAP = {"int64": "int", "int32": "int"}

indent = 0


def name_to_lower(name):
    tmp = re.sub(r'([A-Z])', r'_\1', name).lower()
    return tmp[1:]

def lowercase_first(name):
    return name[:1].lower() + name[1:]

def name_to_upper(name):
    return name.title().replace('_', '')

def is_list_complex_attr(attr):
    if attr['type'].startswith('[]'):
        # this is a list.
        if attr['type'][2:] not in TYPE_MAP:
            # complex type: needs to be in a single form
            return True
    return False


def get_attr_fixed_name(attr):
    fixed_name = attr['name']
    if is_list_complex_attr(attr) and fixed_name.endswith('s'):
        # remove last s
        fixed_name = fixed_name[:-1]
    fixed_name = name_to_lower(fixed_name)
    return fixed_name


def shift():
    global indent
    indent += 1


def unshift():
    global indent
    indent -= 1


def pretty_writeln(f, line):
    for i in range(indent):
        f.write("\t")
    f.write(line)
    f.write("\n")


def write_header(f):
    pretty_writeln(f, "// © Broadcom. All Rights Reserved.")
    pretty_writeln(f, "// The term \"Broadcom\" refers to Broadcom Inc. and/or its subsidiaries.")
    pretty_writeln(f, "// SPDX-License-Identifier: MPL-2.0\n")

    pretty_writeln(f, "package %s\n" % PACKAGE_NAME)
    pretty_writeln(f, "import (")
    shift()
    pretty_writeln(f, "\"fmt\"")
    pretty_writeln(f, "\"github.com/hashicorp/terraform/helper/schema\"")
    pretty_writeln(f, "%s \"github.com/vmware/go-vmware-nsxt\"" % SDK_PACKAGE_NAME)
    pretty_writeln(f, "\"github.com/vmware/go-vmware-nsxt/%s\"" % MANAGER_PACKAGE_NAME)
    pretty_writeln(f, "\"log\"")
    pretty_writeln(f, "\"net/http\"")
    unshift()
    pretty_writeln(f, ")\n")


def write_attr(f, attr):
    fixed_name = get_attr_fixed_name(attr)
    if attr['name'] in VIP_SCHEMA_ATTRS:
        pretty_writeln(f, "\"%s\": get%sSchema()," % (fixed_name, attr['name']))
        return

    is_array = False
    attr_type = attr['type']
    if attr_type.startswith("[]") and attr_type[2:] in TYPE_MAP:
        # Handle arrays. By default, arrays are translated to sets
        # assuming in most cases order is not significant. When order is
        # significant, these should be changed to lists
        # TODO: add both with choice comment
        is_array = True
        attr_type = attr_type[2:]
        attr['helper'] = "%sListFromSchemaSet" % name_to_upper(attr_type)

    if attr_type not in TYPE_MAP:
        print("Skipping attribute %s due to mysterious type %s" % (attr['name'], attr_type))
        return

    pretty_writeln(f, "\"%s\": &schema.Schema{" % fixed_name)
    shift()
    if is_array:
        pretty_writeln(f, "Type:        schema.TypeSet,")
    else:
        pretty_writeln(f, "Type:        %s," % TYPE_MAP[attr_type])

    comment = ' '
    if attr['comment']:
        comment = attr['comment']
    if attr['name'] == 'DisplayName' and comment == 'Defaults to ID if not set':
        comment = "The display name of this resource. " + comment
    pretty_writeln(f, "Description: \"%s\"," % comment)

    if is_array:
        pretty_writeln(f, "Elem: &schema.Schema{")
        shift()
        pretty_writeln(f, "Type:        %s," % TYPE_MAP[attr_type])
        pretty_writeln(f, "ValidateFunc: validate%s," % name_to_upper(fixed_name))
        unshift()
        pretty_writeln(f, "},")

    if attr['optional']:
        pretty_writeln(f, "Optional:    true,")
    else:
        pretty_writeln(f, "Required:    true,")
    if attr['name'] in FORCENEW_ATTRS:
        pretty_writeln(f, "ForceNew:    true,")
    if attr['name'] in COMPUTED_ATTRS or attr['name'] in COMPUTED_AND_OPTIONAL_ATTRS:
        pretty_writeln(f, "Computed:    true,")

    unshift()
    pretty_writeln(f, "},")

def write_func_header(f, resource, operation):
    f.write("\n")
    pretty_writeln(f, "func resourceNsxt%s%s(d *schema.ResourceData, m interface{}) error {" %
            (resource, operation))
    shift()

def write_nsxclient(f):
    pretty_writeln(f, "nsxClient := m.(*%s.APIClient)" % SDK_PACKAGE_NAME)


def write_get_id(f):
    pretty_writeln(f, "id := d.Id()")
    pretty_writeln(f, "if id == \"\" {")
    shift()
    pretty_writeln(f, "return fmt.Errorf(\"Error obtaining logical object id\")")
    unshift()
    pretty_writeln(f, "}\n")


def write_error_check(f, resource, operation):
    if operation == "update":
        pretty_writeln(f, "if err != nil || resp.StatusCode == http.StatusNotFound {")
    else:
        pretty_writeln(f, "if err != nil {")

    shift()
    pretty_writeln(f, "return fmt.Errorf(\"Error during %s %s: " % (resource, operation) + '%v", err)')
    unshift()
    pretty_writeln(f, "}\n")

def write_object(f, resource, attrs, is_create=True):
    used_attrs = []
    for attr in attrs:
        if (is_create and attr['name'] == 'Revision') or attr['name'] in COMPUTED_ATTRS:
            # Revision is irrelevant in create
            continue

        used_attrs.append(attr['name'])
        fixed_name = get_attr_fixed_name(attr)
        if 'helper' in attr:
            # helper function name already computed - this is the case for arrays
            pretty_writeln(f, '%s := get%s(d, "%s")' % (
                lowercase_first(attr['name']),
                attr['helper'],
                fixed_name))
            continue

        if attr['name'] in VIP_GETTER_ATTRS:
            pretty_writeln(f, "%s := get%sFromSchema(d)" % (
                lowercase_first(attr['name']), attr['name']))
            continue

        if attr['type'] in TYPECAST_MAP:
            # type casting is needed
            pretty_writeln(f, "%s := %s(d.Get(\"%s\").(%s))" %
                    (lowercase_first(attr['name']),
                     attr['type'],
                     fixed_name,
                     TYPECAST_MAP[attr['type']]))
        else:
            pretty_writeln(f, "%s := d.Get(\"%s\").(%s)" %
                        (lowercase_first(attr['name']),
                         fixed_name,
                         attr['type']))

    pretty_writeln(f, "%s := %s.%s{" % (lowercase_first(resource), MANAGER_PACKAGE_NAME, resource))
    shift()
    for attr in used_attrs:
        pretty_writeln(f, "%s: %s," % (attr, lowercase_first(attr)))

    unshift()

    pretty_writeln(f, "}\n")

def write_create_func(f, resource, attrs, api_section):

    lower_resource = lowercase_first(resource)

    write_func_header(f, resource, "Create")

    write_nsxclient(f)

    write_object(f, resource, attrs)

    pretty_writeln(f, "%s, resp, err := nsxClient.%s.Create%s(nsxClient.Context, %s)" % (
        lower_resource, api_section, resource, lower_resource))

    f.write("\n")
    write_error_check(f, resource, "create")

    pretty_writeln(f, "if resp.StatusCode != http.StatusCreated {")
    shift()
    pretty_writeln(f, "return fmt.Errorf(\"Unexpected status returned during %s create: " % resource + '%v", resp.StatusCode)')
    unshift()
    pretty_writeln(f, "}")


    pretty_writeln(f, "d.SetId(%s.Id)\n" % lower_resource)

    pretty_writeln(f, "return resourceNsxt%sRead(d, m)" % resource)
    unshift()
    pretty_writeln(f, "}")


def write_read_func(f, resource, attrs, api_section):

    lower_resource = lowercase_first(resource)
    write_func_header(f, resource, "Read")

    write_nsxclient(f)
    write_get_id(f)

    # For some resources this is GET and for other it is read
    pretty_writeln(f, "//TerraGen TODO - select the right command for this resource, and delete this comment")
    pretty_writeln(f, "%s, resp, err := nsxClient.%s.Get%s(nsxClient.Context, id)" %
            (lower_resource, api_section, resource))
    pretty_writeln(f, "%s, resp, err := nsxClient.%s.Read%s(nsxClient.Context, id)" %
            (lower_resource, api_section, resource))

    write_error_check(f, resource, "read")

    pretty_writeln(f, "if resp.StatusCode == http.StatusNotFound {")
    shift()
    pretty_writeln(f, "log.Printf(\"[DEBUG] %s " % resource + '%s not found\", id)')
    pretty_writeln(f, 'd.SetId("")')
    pretty_writeln(f, "return nil")
    unshift()
    pretty_writeln(f, "}")


    for attr in attrs:
        if attr['name'] in IGNORE_ATTRS:
            continue

        if attr['name'] in VIP_SETTER_ATTRS:
            pretty_writeln(f, "set%sInSchema(d, %s.%s)" % (
                attr['name'], lower_resource, attr['name']))
            continue

        fixed_name = get_attr_fixed_name(attr)
        pretty_writeln(f, "d.Set(\"%s\", %s.%s)" %
                (fixed_name, lower_resource, attr['name']))

    f.write("\n")
    pretty_writeln(f, "return nil")
    unshift()
    pretty_writeln(f, "}")


def write_update_func(f, resource, attrs, api_section):

    lower_resource = lowercase_first(resource)
    write_func_header(f, resource, "Update")

    write_nsxclient(f)
    write_get_id(f)

    write_object(f, resource, attrs, is_create=False)
    pretty_writeln(f, "%s, resp, err := nsxClient.%s.Update%s(nsxClient.Context, id, %s)" % (
        lower_resource, api_section, resource, lower_resource))

    f.write("\n")
    write_error_check(f, resource, "update")

    pretty_writeln(f, "return resourceNsxt%sRead(d, m)" % resource)
    unshift()
    pretty_writeln(f, "}")


def write_delete_func(f, resource, attrs, api_section):

    write_func_header(f, resource, "Delete")

    write_nsxclient(f)
    write_get_id(f)

    pretty_writeln(f, "//TerraGen TODO - select the right command for this resource, and delete this comment")
    pretty_writeln(f, "localVarOptionals := make(map[string]interface{})")
    pretty_writeln(f, "resp, err := nsxClient.%s.Delete%s(nsxClient.Context, id, localVarOptionals)" % (
        api_section, resource))
    pretty_writeln(f, "resp, err := nsxClient.%s.Delete%s(nsxClient.Context, id)" % (
        api_section, resource))

    write_error_check(f, resource, "delete")


    pretty_writeln(f, "if resp.StatusCode == http.StatusNotFound {")
    shift()
    pretty_writeln(f, "log.Printf(\"[DEBUG] %s " % resource + '%s not found\", id)')
    pretty_writeln(f, 'd.SetId("")')
    unshift()
    pretty_writeln(f, "}")

    pretty_writeln(f, "return nil")
    unshift()
    pretty_writeln(f, "}")


def write_doc_header(f, resource):
    description = "Provides a resource to configure %s on NSX-T manager" % re.sub('_', ' ', resource)
    pretty_writeln(f, "---")
    pretty_writeln(f, "layout: \"nsxt\"")
    pretty_writeln(f, "page_title: \"NSXT: nsxt_%s\"" % resource)
    pretty_writeln(f, "sidebar_current: \"docs-nsxt-resource-%s\"" % re.sub('_', '-', resource))
    pretty_writeln(f, "description: |-")
    pretty_writeln(f, "  %s" % description)
    pretty_writeln(f, "---\n")
    pretty_writeln(f, "# nsxt_%s\n" % resource)
    pretty_writeln(f, "%s\n" % description)


def write_doc_example(f, resource, attrs):
    obj_name = resource
    pretty_writeln(f, "## Example Usage\n")
    pretty_writeln(f, "```hcl")

    pretty_writeln(f, "resource \"nsxt_%s\" \"%s\" {" % (resource, obj_name))
    for attr in attrs:
        if attr['name'] == 'Revision' or attr['name'] in COMPUTED_ATTRS or attr['name'] in IGNORE_ATTRS:
            continue
        name = get_attr_fixed_name(attr)
        val = "..."
        eq = " = "
        if name == 'display_name':
            val = "\"%s\"" % obj_name
        elif name == 'description':
            val = "\"%s provisioned by Terraform\"" % obj_name
        elif name == 'tag':
            pretty_writeln(f, "")
            val = "{\n    scope = \"color\"\n    tag   = \"red\"\n  }\n"
        pretty_writeln(f, "  %s%s%s" % (name, eq, val))

    pretty_writeln(f, "}")
    pretty_writeln(f, "```\n")


def write_arguments_reference(f, resource, attrs):
    pretty_writeln(f, "## Argument Reference\n")
    pretty_writeln(f, "The following arguments are supported:\n")
    for attr in attrs:
        if attr['name'] == 'Revision' or attr['name'] in COMPUTED_ATTRS or attr['name'] in IGNORE_ATTRS:
            continue
        name = get_attr_fixed_name(attr)
        desc = attr['comment']
        optional = 'Optional' if attr['optional'] else 'Required'
        if name == 'display_name':
            desc = "The display name of this resource. " + desc
        if name == 'tag':
            desc = "A list of scope + tag pairs to associate with this %s" % re.sub('_', ' ', resource)
        pretty_writeln(f, "* `%s` - (%s) %s." % (name, optional, desc))
    pretty_writeln(f, "\n")

def write_attributes_reference(f, resource, attrs):
    res = re.sub('_', ' ', resource)
    pretty_writeln(f, "## Attributes Reference\n")
    pretty_writeln(f, "In addition to arguments listed above, the following attributes are exported:\n")
    pretty_writeln(f, "* `id` - ID of the %s." % res)
    for attr in attrs:
        if attr['name'] == 'Revision' or attr['name'] in COMPUTED_ATTRS or attr['name'] in IGNORE_ATTRS:
            name = name_to_lower(attr['name'])
            desc = attr['comment']
            if name == 'revision':
                desc = 'Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging'

            pretty_writeln(f, "* `%s` - %s." % (name, desc))
    pretty_writeln(f, "\n")


def write_import_doc(f, resource):
    name = re.sub('_', ' ', resource)
    pretty_writeln(f, "## Importing\n")
    pretty_writeln(f, "An existing %s can be [imported][docs-import] into this resource, via the following command:\n" % name)
    pretty_writeln(f, "[docs-import]: /docs/import/index.html\n");
    pretty_writeln(f, "```")
    pretty_writeln(f, "terraform import nsxt_%s.%s UUID" % (resource, resource))
    pretty_writeln(f, "```\n")
    pretty_writeln(f, "The above would import the %s named `%s` with the nsx id `UUID`" % (name, resource))


def main():

    if len(sys.argv) != 3:
        print("Usage: %s <sdk resource file> <api section>" % sys.argv[0])
        sys.exit()

    print("Building resource from %s" % sys.argv[1])
    api_section = sys.argv[2]

    with open(sys.argv[1], 'r') as f:
        lines = f.readlines()

    resource = None
    resource_started = False
    attr_comment = None
    attrs = []
    for line in lines:
        line = line.strip()
        match = re.match("type (.+?) struct", line)
        if match:
            resource = match.group(1)
            resource_started = True
            continue

        if not resource_started:
            continue

        if line.startswith('//'):
            attr_comment = line[3:]
            # remove dot if exists
            if attr_comment.endswith('.'):
                attr_comment = attr_comment[:-1]
            continue

        match = re.match("(.+?) (.+?) `json:\"(.+?)\"`", line)
        if match:
            attr_name = match.group(1).strip()
            attr_type = match.group(2).strip()
            attr_meta = match.group(3).strip()
            attr_optional = 'omitempty' in attr_meta
            if attr_name not in IGNORE_ATTRS:
                attrs.append({'name': attr_name,
                              'type': attr_type,
                              'meta': attr_meta,
                              'comment': attr_comment,
                              'optional': attr_optional})
            attr_comment = None


    print("Resource: %s" % resource)
    resource_lower = name_to_lower(resource)
    print(resource_lower)

    # write the resource file
    with open("resource_nsxt_%s.go" % resource_lower, 'w') as f:
        write_header(f)

        pretty_writeln(f, "func resourceNsxt%s() *schema.Resource {" % resource)
        shift()
        pretty_writeln(f, "return &schema.Resource{")
        shift()
        for op in ("Create", "Read", "Update", "Delete"):
            spaces = "  " if op == "Read" else ""
            pretty_writeln(f, "%s: %sresourceNsxt%s%s," % (op, spaces, resource, op))
        # Add the importer line
        pretty_writeln(f, "Importer: &schema.ResourceImporter{")
        shift()
        pretty_writeln(f, "State: schema.ImportStatePassthrough,")
        unshift()
        pretty_writeln(f, "},")

        f.write("\n")
        pretty_writeln(f, "Schema: map[string]*schema.Schema{")
        shift()

        for attr in attrs:
            write_attr(f, attr)

        unshift()
        pretty_writeln(f, "},")
        unshift()
        pretty_writeln(f, "}")
        unshift()
        pretty_writeln(f, "}")

        write_create_func(f, resource, attrs, api_section)
        write_read_func(f, resource, attrs, api_section)
        write_update_func(f, resource, attrs, api_section)
        write_delete_func(f, resource, attrs, api_section)

    # write the documentation file
    with open("%s.html.markdown" % resource_lower, 'w') as f:
        write_doc_header(f, resource_lower)
        write_doc_example(f, resource_lower, attrs)
        write_arguments_reference(f, resource_lower, attrs)
        write_attributes_reference(f, resource_lower, attrs)
        write_import_doc(f, resource_lower)


main()
