---
subcategory: "Fabric"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_host_transport_node_collection"
description: A resource to configure Policy Host Transport Node Collection.
---

# nsxt_policy_host_transport_node_collection

This resource provides a method for the management of a Host Transport Node Collection.

# Example Usage

```hcl
data "vsphere_compute_cluster" "compute_cluster" {
  provider      = vsphere.vsphere-nsxcfg-data
  name          = "Compute-Cluster"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "nsxt_compute_collection" "compute_cluster_collection" {
  display_name = data.vsphere_compute_cluster.compute_cluster.name
}

resource "nsxt_policy_host_transport_node_collection" "htnc1" {
  display_name                = "HostTransportNodeCollection1"
  compute_collection_id       = data.nsxt_compute_collection.compute_cluster_collection.id
  transport_node_profile_path = nsxt_policy_host_transport_node_profile.tnp.path
}
```

## Argument Reference

The following arguments are supported:

* `display_name` - (Optional) The Display Name of the Transport Zone.
* `description` - (Optional) Description of the Transport Zone.
* `tag` - (Optional) A list of scope + tag pairs to associate with this resource.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the policy resource.
* `compute_collection_id` - (Required) Compute collection id.
* `sub_cluster_config` - (Optional) List of sub-cluster configuration.
  * `host_switch_config_source` - (Required) List of overridden HostSwitch configuration.
    * `host_switch_id` - (Required) HostSwitch ID.
    * `transport_node_profile_sub_config_name` - (Required) Name of the Transport Node Profile sub configuration to be used.
  * `sub_cluster_id` - (Required) sub-cluster ID.
* `transport_node_profile_path` - (Optional) Transport Node Profile Path.
* `remove_nsx_on_destroy` - (Optional) Upon deletion, uninstall NSX from Transport Node Collection member hosts. Default is true. 

## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `id` - ID of the resource.
* `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
* `path` - The NSX path of the policy resource.

## Importing

An existing policy Host Transport Node Collection can be [imported][docs-import] into this resource, via the following command:

```
terraform import nsxt_policy_host_transport_node_collection.test POLICY_PATH
```
The above command imports Policy Host Transport Node Collection named test with NSX policy path POLICY_PATH.
Note: `remove_nsx_on_destroy` will be set to default value upon import. To enforce the intent value, reapply the plan.
