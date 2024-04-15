---
subcategory: "ODS Runbook"
layout: "nsxt"
page_title: "NSXT: policy_ods_pre_defined_runbook"
description: Policy ODS pre-defined runbook data source.
---

# nsxt_policy_ods_pre_defined_runbook

This data source provides information about policy ODS pre-defined runbook configured on NSX.
This data source is applicable to NSX Policy Manager.

## Example Usage

```hcl
data "nsxt_policy_ods_pre_defined_runbook" "test" {
  display_name = "OverlayTunnel"
}
```

## Argument Reference

* `id` - (Optional) The ID of ODS pre-defined runbook to retrieve. If ID is specified, no additional argument should be configured.
* `display_name` - (Optional) The Display Name prefix of the ODS pre-defined runbook to retrieve.

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `description` - The description of the resource.
* `path` - The NSX path of the policy resource.
