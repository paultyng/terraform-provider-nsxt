#!/usr/bin/python
# -*- coding: latin-1 -*-

# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: MPL-2.0

# This script detects difference in model between different versions of NSX spec. List of relevant objects is for now hard-coded.

import sys
import re
import os
import json

OBJECTS = [
"ContextProfile",
"DhcpRelayConfig",
"DhcpServerConfig",
"Domain",
"EvpnConfig",
"EvpnTenantConfig",
"GatewayPolicy",
"Group",
"IPSecVpnDpdProfile",
"IPSecVpnIkeProfile",
"IPSecVpnService",
"IPSecVpnTunnelProfile",
"IPSecVpnLocalEndpoint",
"IdsPolicy",
"IdsProfile",
"IpAddressBlock",
"IpAddressPool",
"L2VPNService",
"L2VPNSession",
"LBPool",
"LBService",
"LBVirtualServer",
"MacDiscoveryProfile",
"OspfConfig",
"PolicyDnsForwarder",
"PolicyNatRule",
"PrefixList",
"QoSProfile",
"SecurityPolicy",
"Segment",
"Service",
"StaticRoutes",
"Tier0",
"Tier0Interface",
"Tier1",
"Tier1Interface"]

def load_spec(path):

    with open(path, 'r') as f:
        spec = json.load(f)

        obj_map = {}
        defs = spec["definitions"]
        for key in defs:
            if "allOf" not in defs[key]:
                if "properties" in defs[key]:
                    # no inheritance
                    obj_map[key] = defs[key]["properties"]
                continue
            if len(defs[key]["allOf"]) == 2:
                # object inheritance
                if "properties" not in defs[key]["allOf"][1]:
                    continue
                obj_map[key] = defs[key]["allOf"][1]["properties"]

        return obj_map

def ref_to_def(ref):
    return ref[len("#definitions/ "):]

def print_ident(text, level):
    ident = "  "*level
    print("%s%s" % (ident, text))

class color:
    PURPLE = '\033[95m'
    BLUE = '\033[94m'
    GREEN = '\033[92m'
    END = '\033[0m'

def main():

    if len(sys.argv) < 3:
        print("Usage: %s <baseline.json> <target.json>" % sys.argv[0])
        sys.exit()

    baseline_file_path = sys.argv[1]
    target_file_path = sys.argv[2]
    baseline_map = load_spec(baseline_file_path)
    target_map = load_spec(target_file_path)
    level = 0

    def analyze_obj(obj, level):
        print_ident("analyzing %s.." % obj, level)
        if obj not in baseline_map:
            return
        for attr in target_map[obj]:
            if attr not in baseline_map[obj]:
                print_ident(color.BLUE + "new attribute %s" % attr + color.END, level + 1)

            if "$ref" in target_map[obj][attr]:
                analyze_obj(ref_to_def(target_map[obj][attr]["$ref"]), level + 1)
            if "items" in target_map[obj][attr] and "$ref" in target_map[obj][attr]["items"]:
                analyze_obj(ref_to_def(target_map[obj][attr]["items"]["$ref"]), level + 1)
            if "enum" in target_map[obj][attr] and attr in baseline_map[obj]:
                target_set = set(target_map[obj][attr]["enum"])
                baseline_set = set(baseline_map[obj][attr]["enum"])
                diff = target_set - baseline_set
                if diff:
                    print_ident(color.GREEN + "new enum values for attribute %s: %s" % (attr, diff) + color.END, level + 1)

        for attr in baseline_map[obj]:
            if obj not in target_map:
                continue
            if attr not in target_map[obj]:
                print_ident(color.PURPLE + "deleted attribute %s" % attr + color.END, level + 1)

    for obj in OBJECTS:
        analyze_obj(obj, level + 1)


    print("Done.")

main()
