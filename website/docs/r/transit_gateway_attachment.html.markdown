---
subcategory: "Beta"
layout: "nsxt"
page_title: "NSXT: nsxt_transit_gateway_attachment"
description: A resource to configure a Transit Gateway Attachment.
---

# nsxt_transit_gateway_attachment

This resource provides a method for the management of a Transit Gateway Attachment.

This resource is applicable to NSX Policy Manager.

## Example Usage

```hcl
resource "nsxt_transit_gateway_attachment" "test" {
  parent_path     = nsxt_transit_gateway.test.path
  display_name    = "test"
  description     = "Terraform provisioned Transit Gateway Attachment"
  connection_path = nsxt_gateway_connection.test.path
}
```

## Argument Reference

The following arguments are supported:

* `parent_path` - (Required) The path of the parent to bind with the profile. This is a policy path of a transit gateway.
* `display_name` - (Required) Display name of the resource.
* `description` - (Optional) Description of the resource.
* `tag` - (Optional) A list of scope + tag pairs to associate with this resource.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the resource.
* `connection_path` - (Optional) Policy path of the desidered transit gateway external connection.

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `id` - ID of the resource.
* `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
* `path` - The NSX path of the policy resource.

## Importing

An existing object can be [imported][docs-import] into this resource, via the following command:

[docs-import]: https://www.terraform.io/cli/import

```
terraform import nsxt_transit_gateway_attachment.test PATH
```

The above command imports TransitGatewayAttachment named `test` with the policy path `PATH`.
