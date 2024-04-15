---
subcategory: "Beta"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_ods_runbook_invocation"
description: A resource to configure ODS runbook invocation in NSX Policy manager.
---

# nsxt_policy_ods_runbook_invocation

This resource provides a method for the management of a ODS runbook invocation.

## Example Usage

```hcl
data "nsxt_policy_host_transport_node" "tn1" {
  display_name = "some_node_name.org"
}

data "nsxt_policy_ods_pre_defined_runbook" "ot_runbook" {
  display_name = "OverlayTunnel"
}

resource "nsxt_policy_ods_runbook_invocation" "test" {
  display_name = "overlay_tunnel_invocation"
  runbook_path = data.nsxt_policy_ods_pre_defined_runbook.ot_runbook.path
  argument {
    key   = "src"
    value = "192.168.0.11"
  }
  argument {
    key   = "dst"
    value = "192.168.0.10"
  }
  target_node = data.nsxt_policy_host_transport_node.tn1.id
}
```

## Argument Reference

The following arguments are supported:

* `display_name` - (Required) Display name of the resource.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the policy resource.
* `argument` - (Optional) Arguments for runbook invocation.
  * `key` - (Required) Key
  * `value` - (Required) Value
* `runbook_path` - (Required) Path of runbook object.
* `target_node` - (Optional) Identifier of an appliance node or transport node.

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
* `path` - The NSX path of the policy resource.

## Importing

An existing policy ODS runbook invocation can be [imported][docs-import] into this resource, via the following command:

[docs-import]: https://www.terraform.io/cli/import

```
terraform import nsxt_policy_ods_runbook_invocation.test POLICY_PATH
```
The above command imports the policy ODS runbook invocation named `test` for policy path `POLICY_PATH`.
