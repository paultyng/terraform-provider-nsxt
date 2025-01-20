---
subcategory: "Firewall"
layout: "nsxt"
page_title: "NSXT: policy_services"
description: A policy services data source. This data source builds "display name to policy path" map representation of the whole table.
---

# nsxt_policy_services

This data source builds a "name to policy path" map of the whole policy Services table. Such map can be referenced in configuration to obtain object identifier attributes by display name at a cost of single roundtrip to NSX, which improves apply and refresh
time at scale, compared to multiple instances of `nsxt_policy_service` data source.

## Example Usage

```hcl
data "nsxt_policy_services" "map" {
}

resource "nsxt_policy_nat_rule" "dnat1" {
  display_name          = "dnat_rule1"
  action                = "DNAT"
  source_networks       = ["9.1.1.1", "9.2.1.1"]
  destination_networks  = ["11.1.1.1"]
  translated_networks   = ["10.1.1.1"]
  gateway_path          = nsxt_policy_tier1_gateway.t1gateway.path
  logging               = false
  firewall_match        = "MATCH_INTERNAL_ADDRESS"
  policy_based_vpn_mode = "BYPASS"
  service               = data.nsxt_policy_services.map.items["DNS-UDP"]

  tag {
    scope = "color"
    tag   = "blue"
  }
}
```

## Argument Reference

* `context` - (Optional) The context which the object belongs to
    * `project_id` - (Required) The ID of the project which the object belongs to

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `items` - Map of policy service policy paths keyed by display name.
