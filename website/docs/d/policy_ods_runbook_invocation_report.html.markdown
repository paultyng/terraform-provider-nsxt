---
subcategory: "ODS Runbook"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_ods_runbook_invocation_report"
description: Policy ODS runbook invocation report data source.
---

# nsxt_policy_ods_runbook_invocation_report

This data source provides information about policy ODS runbook invocation report on NSX.
This data source is applicable to NSX Policy Manager.

## Example Usage

```hcl
data "nsxt_policy_ods_runbook_invocation_report" "test" {
  invocation_id = nsxt_policy_ods_runbook_invocation.test.id
}
```

## Argument Reference

* `invocation_id` - (Required) UUID of runbook invocation.

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `target_node` - Identifier of an appliance node or transport node.
* `error_detail` - The report error detail.
* `invalid_reason` - Invalid report reason.
* `recommendation_code` - Online Diagnostic System recommendation code.
* `recommendation_message` - Online Diagnostic System recommendation message.
* `result_code` - Online Diagnostic System result code.
* `result_message` - Online Diagnostic System result message.
* `request_status` - Request status of a runbook invocation.
* `operation_state` - Operation state of a runbook invocation on the target node.
