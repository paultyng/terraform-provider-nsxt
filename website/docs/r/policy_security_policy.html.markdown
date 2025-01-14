---
subcategory: "Firewall"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_security_policy"
description: A resource to configure a Security Group and its rules.
---

# nsxt_policy_security_policy

This resource provides a method for the management of Security Policy and rules under it.

This resource is applicable to NSX Global Manager, NSX Policy Manager and VMC.

## Example Usage

```hcl
resource "nsxt_policy_security_policy" "policy1" {
  display_name = "policy1"
  description  = "Terraform provisioned Security Policy"
  category     = "Application"
  locked       = false
  stateful     = true
  tcp_strict   = false
  scope        = [nsxt_policy_group.pets.path]

  rule {
    display_name       = "block_icmp"
    destination_groups = [nsxt_policy_group.cats.path, nsxt_policy_group.dogs.path]
    action             = "DROP"
    services           = [nsxt_policy_service.icmp.path]
    logged             = true
  }

  rule {
    display_name     = "allow_udp"
    source_groups    = [nsxt_policy_group.fish.path]
    sources_excluded = true
    scope            = [nsxt_policy_group.aquarium.path]
    action           = "ALLOW"
    services         = [nsxt_policy_service.udp.path]
    logged           = true
    disabled         = true
    notes            = "Disabled by starfish for debugging"
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Global Manager example

```hcl
data "nsxt_policy_site" "paris" {
  display_name = "Paris"
}
resource "nsxt_policy_security_policy" "policy1" {
  display_name = "policy1"
  description  = "Terraform provisioned Security Policy"
  category     = "Application"
  locked       = false
  stateful     = true
  tcp_strict   = false
  scope        = [nsxt_policy_group.pets.path]
  domain       = data.nsxt_policy_site.paris.id

  rule {
    display_name       = "block_icmp"
    destination_groups = [nsxt_policy_group.cats.path, nsxt_policy_group.dogs.path]
    action             = "DROP"
    services           = [nsxt_policy_service.icmp.path]
    logged             = true
  }

  rule {
    display_name     = "allow_udp"
    source_groups    = [nsxt_policy_group.fish.path]
    sources_excluded = true
    scope            = [nsxt_policy_group.aquarium.path]
    action           = "ALLOW"
    services         = [nsxt_policy_service.udp.path]
    logged           = true
    disabled         = true
    notes            = "Disabled by starfish for debugging"
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Example Usage - Multi-Tenancy

```hcl
data "nsxt_policy_project" "demoproj" {
  display_name = "demoproj"
}

resource "nsxt_policy_security_policy" "policy1" {
  context {
    project_id = data.nsxt_policy_project.demoproj.id
  }
  display_name = "policy1"
  description  = "Terraform provisioned Security Policy"
  category     = "Application"
  locked       = false
  stateful     = true
  tcp_strict   = false
  scope        = [nsxt_policy_group.pets.path]

  rule {
    display_name       = "block_icmp"
    destination_groups = [nsxt_policy_group.cats.path, nsxt_policy_group.dogs.path]
    action             = "DROP"
    services           = [nsxt_policy_service.icmp.path]
    logged             = true
  }

  rule {
    display_name     = "allow_udp"
    source_groups    = [nsxt_policy_group.fish.path]
    sources_excluded = true
    scope            = [nsxt_policy_group.aquarium.path]
    action           = "ALLOW"
    service_entries {
      l4_port_set_entry {
        protocol          = "UDP"
        destination_ports = ["9080"]
      }
    }
    logged   = true
    disabled = true
    notes    = "Disabled by starfish for debugging"
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

-> We recommend using `lifecycle` directive as in samples above, in order to avoid dependency issues when updating groups/services simultaneously with the rule.

## Argument Reference

The following arguments are supported:

* `display_name` - (Required) Display name of the resource.
* `description` - (Optional) Description of the resource.
* `domain` - (Optional) The domain to use for the resource. This domain must already exist. For VMware Cloud on AWS use `cgw`. For Global Manager, please use site id for this field. If not specified, this field is default to `default`.
* `tag` - (Optional) A list of scope + tag pairs to associate with this policy.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the resource.
* `context` - (Optional) The context which the object belongs to
  * `project_id` - (Required) The ID of the project which the object belongs to
* `category` - (Required) Category of this policy. For local manager must be one of `Ethernet`, `Emergency`, `Infrastructure`, `Environment`, `Application`. For global manager must be one of: `Infrastructure`, `Environment`, `Application`.
* `comments` - (Optional) Comments for security policy lock/unlock.
* `locked` - (Optional) Indicates whether a security policy should be locked. If locked by a user, no other user would be able to modify this policy.
* `scope` - (Optional) The list of policy object paths where the rules in this policy will get applied.
* `sequence_number` - (Optional) This field is used to resolve conflicts between security policies across domains.
* `stateful` - (Optional) If true, state of the network connects are tracked and a stateful packet inspection is performed. Default is true.
* `tcp_strict` - (Optional) Ensures that a 3 way TCP handshake is done before the data packets are sent. Default is false.
* `rule` - (Optional) A repeatable block to specify rules for the Security Policy. Each rule includes the following fields:
  * `display_name` - (Required) Display name of the resource.
  * `description` - (Optional) Description of the resource.
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
  * `tag` - (Optional) A list of scope + tag pairs to associate with this Rule.
  * `sequence_number` - (Optional) It is recommended not to specify sequence number for rules, and rely on provider to auto-assign them. If you choose to specify sequence numbers, you must make sure the numbers are consistent with order of the rules in configuration. Please note that sequence numbers should start with 1 and not 0. To avoid confusion, either specify sequence numbers in all rules, or none at all.


## Attributes Reference

In addition to arguments listed above, the following attributes are exported:

* `id` - ID of the Security Policy.
* `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
* `path` - The NSX path of the policy resource.
* `rule`:
  * `revision` - Indicates current revision number of the object as seen by NSX-T API server. This attribute can be useful for debugging.
  * `path` - The NSX path of the policy resource.
  * `sequence_number` - Sequence number for the rule.
  * `rule_id` - Unique positive number that is assigned by the system and is useful for debugging.

## Importing

An existing security policy can be [imported][docs-import] into this resource, via the following command:

[docs-import]: https://www.terraform.io/cli/import

```
terraform import nsxt_policy_security_policy.policy1 domain/ID
```

The above command imports the security policy named `policy1` under NSX domain `domain` with the NSX Policy ID `ID`.
