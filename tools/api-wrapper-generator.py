#!/usr/bin/python3

# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: MPL-2.0

import argparse
import atexit
import git
import os
import re
import shutil
import string
import tempfile
import yaml

SDK_REPO = 'github.com/vmware/vsphere-automation-sdk-go'
SDK_BASE_PATH = 'services/nsxt'
TF_PROVIDER_REPO = 'github.com/vmware/terraform-provider-nsxt'


def cleanup():
    try:
        # TODO: uncomment
        shutil.rmtree(get_func_definition.repo_path)
        pass
    except (FileNotFoundError, AttributeError):
        pass


def mkpath(path):
    if not os.path.isdir(path):
        os.makedirs(path)


def get_pkg_name_from_path(path):
    return path.split('/')[-1].replace("_", "")


def camel_to_snake(name):
    name = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', name)
    return re.sub('([a-z0-9])([A-Z])', r'\1_\2', name).lower()


def get_yaml_file(fname):
    with open(fname) as f:
        return yaml.safe_load(f)


def var_name(var):
    # Variable names within function defs are the obj_name but with 1st char lowercase
    return var[0].lower() + var[1:]


def get_arglist(arg_str):
    # Retrieve function argument list from function definition
    arg_list = arg_str.replace(', ', ',').split(',')
    return [arg[0] for arg in [a.split(' ') for a in arg_list]]


def parse_new_call(func_def):
    # Parse New* definition
    try:
        r = parse_new_call.regex
    except AttributeError:
        r = re.compile('func\s+(.*)\((.*)\)\s+(.*)\s+\{.*')
        parse_new_call.regex = r

    m = r.match(func_def)
    return m.groups()


def new_func_call_setup(api, subs_dict):
    g = parse_new_call(subs_dict['func_def'])
    arg_list = get_arglist(g[1])
    return '%s(%s)' % (g[0], ', '.join(arg_list))


def find_api_package_attributes(api, type):
    for a in api['api_packages']:
        if type == a['type']:
            return a


def api_func_call_setup(api, subs_dict):
    g = parse_api_call(subs_dict['func_def'])
    arg_list = get_arglist(g[2])
    if subs_dict['type'] == "Multitenancy":
        arg_list = ['utl.DefaultOrgID', 'c.ProjectID'] + arg_list
    elif subs_dict['type'] == "VPC":
        arg_list = ['utl.DefaultOrgID', 'c.ProjectID', 'c.VPCID'] + arg_list
        arg_list.remove('domainIdParam')
        attrs = find_api_package_attributes(api, subs_dict['type'])
        x = attrs.get('ignore_params', {})
        for n in attrs.get('ignore_params', {}).get(g[1], []):
            arg_list.remove(n)

    return '%s(%s)' % (g[1], ', '.join(arg_list))


def patch_func_call_setup(api, subs_dict):
    g = parse_api_call(subs_dict['func_def'])
    arg_list = get_arglist(g[2])
    if api['template_type'] == 'Convert':
        for n in range(0, len(arg_list)):
            if arg_list[n] == subs_dict['var_name']:
                arg_list[n] = 'gmObj.(%s.%s)' % (subs_dict['model_import'], subs_dict['model_name'])
    elif subs_dict['type'] == "Multitenancy":
        arg_list = ['utl.DefaultOrgID', 'c.ProjectID'] + arg_list
    elif subs_dict['type'] == "VPC":
        arg_list = ['utl.DefaultOrgID', 'c.ProjectID', 'c.VPCID'] + arg_list
        arg_list.remove('domainIdParam')
    return '%s(%s)' % (g[1], ', '.join(arg_list))


def parse_api_call(func_def):
    # Parse API definition
    try:
        r = parse_api_call.regex
    except AttributeError:
        r = re.compile('func\s+(.*)\s+(.*)\((.*)\)\s+(.*)\s+\{')
        parse_api_call.regex = r

    m = r.match(func_def)
    return m.groups()


def new_func_def_setup(subs_dict):
    # Prepare New* definition from function definition
    g = parse_new_call(subs_dict['func_def'])
    return '%s(sessionContext utl.SessionContext, %s) *%sClientContext' % (g[0], g[1], subs_dict['model_name'])


def api_func_def_setup(subs_dict):
    # Prepare API definition from function definition
    func_def = subs_dict['func_def'].replace('%s.' % subs_dict['model_prefix'], '%s.' % subs_dict['main_model_import'])
    g = parse_api_call(func_def)
    return '(c %sClientContext) %s(%s) %s' % (subs_dict['model_name'], g[1], g[2], g[3])


def list_func_def_setup(subs_dict):
    # Prepare API definition from function definition
    func_def = subs_dict['func_def'].replace('nsx_policyModel.', '%s.' % subs_dict['list_main_model_import'])
    g = parse_api_call(func_def)
    return '(c %sClientContext) %s(%s) %s' % (subs_dict['model_name'], g[1], g[2], g[3])


def get_func_definition(api, func, path):
    # Retrieve func definition from SDK repo
    try:
        repo_path = get_func_definition.repo_path

    except AttributeError:
        # Get SDK repo from GitHub
        repo_path = tempfile.mkdtemp()
        git.Repo.clone_from('https://%s' % SDK_REPO, repo_path)
        get_func_definition.repo_path = repo_path

    fname = api.get('client_name', '%ssClient' % api['obj_name'])
    with open('%s/%s/%s.go' % (repo_path, path.replace(SDK_REPO, ''), fname), 'r') as f:
        for line in f:
            if line.startswith('func ') and ' ' + func + '(' in line:
                return line
    raise Exception('No function definition found for %s for model %s' % (func, api['model_name']))


FUNC_DEF_CALLBACK = {
    'New': new_func_def_setup,
    'Create': api_func_def_setup,
    'Get': api_func_def_setup,
    'Patch': api_func_def_setup,
    'Update': api_func_def_setup,
    'Delete': api_func_def_setup,
    'List': list_func_def_setup
}

FUNC_CALL_CALLBACK = {
    'New': new_func_call_setup,
    'Get': api_func_call_setup,
    'Create': patch_func_call_setup,
    'Patch': patch_func_call_setup,
    'Update': patch_func_call_setup,
    'Delete': api_func_call_setup,
    'List': api_func_call_setup
}

atexit.register(cleanup)


def get_args():
    parser = argparse.ArgumentParser(
        prog='api-wrapper-generator.py',
        description='Generate API wrappers for NSX Golang SDK')
    parser.add_argument('--api_list', required=True)
    parser.add_argument('--api_templates', required=True)
    parser.add_argument('--api_file_template', required=True)
    parser.add_argument('--utl_file_template', required=True)
    parser.add_argument('--out_dir', required=True)
    return parser.parse_args()


def get_dicts_from_yaml(api_list_file, api_template_file, api_file_template_file, utl_file_template_file):
    api_list = get_yaml_file(api_list_file)
    api_templates = get_yaml_file(api_template_file)
    api_file_template = get_yaml_file(api_file_template_file)
    utl_file_template = get_yaml_file(utl_file_template_file)

    return api_list, api_templates, api_file_template, utl_file_template


def format_file(f):
    os.system('gofmt -s -w %s' % f)


def write_utl_file(out_dir, utl_file_template):
    # Create type definitions
    type_set = set()
    for api in api_list:
        for pkg in api['api_packages']:
            type_set.add(pkg['type'])

    # Build type list
    n = 0
    type_list = []
    for t in sorted(type_set):
        type_list.append("    %s = %d" % (t, n))
        n += 1

    template = string.Template(utl_file_template)
    output = template.substitute({
        "const": "\n".join(type_list),
        "pkg_name": get_pkg_name_from_path(out_dir)})

    utl_dir = out_dir + '/utl'
    mkpath(utl_dir)
    outfile = '%s/api_util.go' % utl_dir
    with open(outfile, 'w') as f:
        f.write(output)

    format_file(outfile)


def write_api_file(api, out_dir, api_file_template):
    client_dict = {}
    model_dict = {}
    list_model_dict = {}

    for pkg in api['api_packages']:
        if not client_dict.get(pkg['client']):
            client_dict[pkg['client']] = 'client%d' % len(client_dict)
        if not model_dict.get(pkg['model']):
            model_dict[pkg['model']] = 'model%d' % len(model_dict)
        if pkg.get('list_result_model') and not list_model_dict.get(pkg['list_result_model']):
            list_model_dict[pkg['list_result_model']] = 'lrmodel%d' % len(list_model_dict)
    # Build import list
    import_dict = client_dict.copy()
    import_dict.update(model_dict.copy())
    import_dict.update(list_model_dict.copy())
    import_list = ['    %s "%s"' % (import_dict[s], s) for s in sorted(import_dict.keys())]

    template = string.Template(api_file_template)
    context_type = '%sClientContext utl.ClientContext' % api['model_name']

    pkg_name = get_pkg_name_from_path(api['api_packages'][0]['client'])
    output = template.substitute({
        "imports": "\n".join(import_list),
        "context_type": context_type,
        "utl_pkg_import": 'utl "%s/%s"' % (TF_PROVIDER_REPO, os.path.relpath(out_dir + '/utl')),
        "pkg_name": pkg_name})

    subs_dict = {
        "client_name": api.get('client_name', '%ssClient' % api['obj_name']),
        "obj_name": api['obj_name'],
        "model_name": api['model_name'],
        "var_name": api.get('var_name', '%sParam' % var_name(api['obj_name'])),
        "list_result_name": api.get('list_result_name', '%sListResult' % api['model_name']),
        "model_prefix": api.get('model_prefix', 'nsx_policyModel'),
        "ptr_prefix": '*' if api.get('model_pass_ptr') else ''
    }
    main_model_import = None
    list_main_model_import = None
    out_path = None
    api_list = api.get('supported_method', api_templates.keys())
    for api_name in api_list:
        tpl = api_templates[api_name]
        case_items = ""
        for pkg in api['api_packages']:
            model_import = model_dict[pkg["model"]]
            client_import = client_dict[pkg["client"]]
            list_model_import = list_model_dict.get(pkg.get("list_result_model"), model_import)
            if main_model_import is None:
                main_model_import = model_import
                list_main_model_import = list_model_import
                out_path = pkg["client"].replace(SDK_REPO, '').replace(SDK_BASE_PATH, '')
            if api_name == 'New':
                if api.get('client_name'):
                    func_name = 'New%s' % api['client_name']
                else:
                    func_name = 'New%ssClient' % api['obj_name']
            else:
                func_name = api_name
            func_def = get_func_definition(api, func_name, api['api_packages'][0]['client'])
            subs_dict.update({
                "client_import": client_import,
                "model_import": model_import,
                "main_model_import": main_model_import,
                "list_model_import": list_model_import,
                "list_main_model_import": list_main_model_import,
                "func_def": func_def,
                "type": pkg['type']
            })
            if api_name != 'List' and model_import == main_model_import:
                api['template_type'] = "NoConvert"
            elif api_name == 'List' and list_model_import == list_main_model_import:
                api['template_type'] = "NoConvert"
            else:
                api['template_type'] = "Convert"
            if FUNC_DEF_CALLBACK.get(api_name):
                subs_dict['api_func_def'] = FUNC_DEF_CALLBACK[api_name](subs_dict)
            if FUNC_CALL_CALLBACK.get(api_name):
                subs_dict['api_func_call'] = FUNC_CALL_CALLBACK[api_name](api, subs_dict)
            template = string.Template(tpl[api['template_type']])
            case_items += template.substitute(subs_dict)

        template = string.Template(tpl['main'])
        subs_dict['case_items'] = case_items
        output += template.substitute(subs_dict)

    mkpath('%s/%s' % (out_dir, out_path))
    outfile = '%s/%s/%s.go' % (out_dir, out_path, camel_to_snake(api.get('file_name', api['model_name'])))
    with open(outfile, 'w') as f:
        f.write(output)

    format_file(outfile)


args = get_args()
api_list, api_templates, api_file_template, utl_file_template = get_dicts_from_yaml(args.api_list, args.api_templates,
                                                                                    args.api_file_template,
                                                                                    args.utl_file_template)
out_dir = args.out_dir

write_utl_file(out_dir, utl_file_template)

for api in api_list:
    write_api_file(api, out_dir, api_file_template)

exit(0)
