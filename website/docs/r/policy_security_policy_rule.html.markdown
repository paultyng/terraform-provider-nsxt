---
subcategory: "Firewall"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_security_policy_rule"
description: A resource to configure a Security Policy Rule.
---

# nsxt_policy_security_policy_rule

This resource provides a method for the management of Security Policy Rule.

Note: to avoid unexpected behavior, don't use this resource and resource `nsxt_policy_security_policy` to manage rules under a security policy at the same time. 
Instead, please use this resource with resource `nsxt_policy_parent_security_policy` to manage a security policy and its rules separately, and use `nsxt_policy_security_policy` to manage a security policy and its rules in one single resource.

This resource is applicable to NSX Global Manager, NSX Policy Manager and VMC.

## Example Usage

```hcl
resource "nsxt_policy_security_policy_rule" "rule1" {
  display_name       = "rule1"
  description        = "Terraform provisioned Security Policy Rule"
  policy_path        = nsxt_policy_parent_security_policy.policy1.path
  sequence_number    = 1
  destination_groups = [nsxt_policy_group.cats.path, nsxt_policy_group.dogs.path]
  action             = "DROP"
  services           = [nsxt_policy_service.icmp.path]
  logged             = true
}
```

## Example Usage - Multi-Tenancy

```hcl
resource "nsxt_policy_security_policy_rule" "rule1" {
  display_name       = "rule1"
  description        = "Terraform provisioned Security Policy Rule"
  policy_path        = nsxt_policy_parent_security_policy.policy1.path # Path of a multi-tenancy policy
  sequence_number    = 1
  destination_groups = [nsxt_policy_group.cats.path, nsxt_policy_group.dogs.path]
  action             = "DROP"
  services           = [nsxt_policy_service.icmp.path]
  logged             = true
}
```

## Argument Reference

The following arguments are supported:

* `display_name` - (Required) Display name of the resource.
* `description` - (Optional) Description of the resource.
* `tag` - (Optional) A list of scope + tag pairs to associate with this policy.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the resource.
* `policy_path` - (Required) The path of the Security Policy which the object belongs to
* `context` - (Optional) The context which the object belongs to. If it's not provided, it will be derived from `policy_path`.
  * `project_id` - (Required) The ID of the project which the object belongs to
* `sequence_number` - (Required) This field is used to resolve conflicts between multiple Rules under Security or Gateway Policy for a Domain. Please note that sequence numbers should start with 1 and not 0 to avoid confusion.
* `action` - (Optional) Rule action, one of `ALLOW`, `DROP`, `REJECT` and `JUMP_TO_APPLICATION`. Default is `ALLOW`. `JUMP_TO_APPLICATION` is only applicable in `Environment` category.
* `destination_groups` - (Optional) Set of group paths that serve as the destination for this rule. IPs, IP ranges, or CIDRs may also be used starting in NSX-T 3.0. An empty set can be used to specify "Any".
* `source_groups` - (Optional) Set of group paths that serve as the source for this rule. IPs, IP ranges, or CIDRs may also be used starting in NSX-T 3.0. An empty set can be used to specify "Any".
* `destinations_excluded` - (Optional) A boolean value indicating negation of destination groups.
* `sources_excluded` - (Optional) A boolean value indicating negation of source groups.
* `direction` - (Optional) Traffic direction, one of `IN`, `OUT` or `IN_OUT`. Default is `IN_OUT`.
* `disabled` - (Optional) Flag to disable this rule. Default is false.
* `ip_version` - (Optional) Version of IP protocol, one of `NONE`, `IPV4`, `IPV6`, `IPV4_IPV6`. Default is `IPV4_IPV6`. For `Ethernet` category rules, use `NONE` value.
* `logged` - (Optional) Flag to enable packet logging. Default is false.
* `notes` - (Optional) Additional notes on changes.
* `profiles` - (Optional) Set of profile paths relevant for this rule.
* `scope` - (Optional) Set of policy object paths where the rule is applied.
* `services` - (Optional) Set of service paths to match.
* `service_entries` - (Optional) Set of explicit protocol/port service definition
  * `icmp_entry` - (Optional) Set of ICMP type service entries
    * `display_name` - (Optional) Display name of the service entry
    * `protocol` - (Required) Version of ICMP protocol: `ICMPv4` or `ICMPv6`
    * `icmp_code` - (Optional) ICMP message code
    * `icmp_type` - (Optional) ICMP message type
  * `l4_port_set_entry` - (Optional) Set of L4 ports set service entries
    * `display_name` - (Optional) Display name of the service entry
    * `protocol` - (Required) L4 protocol: `TCP` or `UDP`
    * `destination_ports` - (Optional) Set of destination ports
    * `source_ports` - (Optional) Set of source ports
  * `igmp_entry` - (Optional) Set of IGMP type service entries
    * `display_name` - (Optional) Display name of the service entry
  * `ether_type_entry` - (Optional) Set of Ether type service entries
    * `display_name` - (Optional) Display name of the service entry
    * `ether_type` - (Required) Type of the encapsulated protocol
  * `ip_protocol_entry` - (Optional) Set of IP Protocol type service entries
    * `display_name` - (Optional) Display name of the service entry
    * `protocol` - (Required) IP protocol number
  * `algorithm_entry` - (Optional) Set of Algorithm type service entries
    * `display_name` - (Optional) Display name of the service entry
    * `destination_port` - (Required) a single destination port
    * `source_ports` - (Optional) Set of source ports/ranges
    * `algorithm` - (Required) Algorithm: one of `ORACLE_TNS`, `FTP`, `SUN_RPC_TCP`, `SUN_RPC_UDP`, `MS_RPC_TCP`, `MS_RPC_UDP`, `NBNS_BROADCAST`(Deprecated), `NBDG_BROADCAST`(Deprecated), `TFTP`
* `log_label` - (Optional) Additional information (string) which will be propagated to the rule syslog.


## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
* `path` - The NSX path of the policy resource.
* `rule_id` - Unique positive number that is assigned by the system and is useful for debugging.

## Importing

An existing security policy can be [imported][docs-import] into this resource, via the following command:

[docs-import]: https://www.terraform.io/cli/import

```
terraform import nsxt_policy_security_policy_rule.rule1 POLICY_PATH
```

The above command imports the security policy rule named `rule1` with policy path `POLICY_PATH`.
