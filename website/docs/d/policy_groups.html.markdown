---
subcategory: "Firewall"
layout: "nsxt"
page_title: "NSXT: policy_groups"
description: A policy groups data source. This data source builds "display name to policy paths" map representation of the whole table.
---

# nsxt_policy_groups

This data source builds a "name to paths" map of the whole policy Groups table. Such map can be referenced in configuration to obtain object identifier attributes by display name at a cost of single roundtrip to NSX, which improves apply and refresh
time at scale, compared to multiple instances of `nsxt_policy_group` data source.

## Example Usage

```hcl
data "nsxt_policy_groups" "map" {
}

resource "nsxt_policy_predefined_security_policy" "test" {
  path = data.nsxt_policy_security_policy.default_l3.path

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name       = "allow_icmp"
    destination_groups = [data.nsxt_policy_groups.items["Cats"], data.nsxt_policy_groups.items["Dogs"]]
    action             = "ALLOW"
    services           = [nsxt_policy_service.icmp.path]
    logged             = true
  }

  rule {
    display_name     = "allow_udp"
    source_groups    = [data.nsxt_policy_groups.items["Fish"]]
    sources_excluded = true
    scope            = [data.nsxt_policy_groups.items["Aquarium"]]
    action           = "ALLOW"
    services         = [nsxt_policy_service.udp.path]
    logged           = true
    disabled         = true
  }

  default_rule {
    action = "DROP"
  }

}
```

## Argument Reference

* `domain` - (Optional) The domain this Group belongs to. For VMware Cloud on AWS use `cgw`. For Global Manager, please use site id for this field. If not specified, this field is default to `default`.
* `context` - (Optional) The context which the object belongs to
  * `project_id` - (Required) The ID of the project which the object belongs to

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `items` - Map of policy service policy paths keyed by display name.
