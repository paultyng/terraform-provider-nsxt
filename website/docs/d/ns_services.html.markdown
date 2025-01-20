---
subcategory: "Deprecated"
layout: "nsxt"
page_title: "NSXT: ns_services"
description: A networking and security services data source. This data source builds "display name to id" map representation of the whole table.
---

# nsxt_ns_services

This data source builds a "name to uuid" map of the whole NS Services table. Such map can be referenced in configuration to obtain object uuids by display name at a cost of single roundtrip to NSX, which improves apply and refresh
time at scale, compared to multiple instances of `nsxt_ns_service` data source.

## Example Usage

```hcl
data "nsxt_ns_services" "map" {
}

resource "nsxt_firewall_section" "s1" {
  display_name = "section1"

  rule {
    display_name = "in_rule"
    action       = "DROP"
    direction    = "IN"

    service {
      target_type = "NSService"
      target_id   = data.nsxt_ns_services.map.items["service1"]
    }
  }

  section_type = "LAYER3"
}

```

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `items` - Map of ns service uuids keyed by display name.
