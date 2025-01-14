---
subcategory: "Firewall"
layout: "nsxt"
page_title: "NSXT: nsxt_policy_gateway_policy"
description: A resource to Gateway security policies.
---

# nsxt_policy_gateway_policy

This resource provides a method for the management of a Gateway Policy and its Rules.

This resource is applicable to NSX Global Manager and NSX Policy Manager.

## Example Usage

```hcl
resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "tf-gw-policy"
  description     = "Terraform provisioned Gateway Policy"
  category        = "LocalGatewayRules"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name       = "rule1"
    destination_groups = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    disabled           = true
    action             = "DROP"
    logged             = true
    scope              = [nsxt_policy_tier1_gateway.policygateway.path]
  }

  lifecycle {
    create_before_destroy = true
  }
}
```

## Example Usage with Global Manager

```hcl
resource "nsxt_policy_domain" "france" {
  display_name = "France"
  sites        = ["Paris"]
}

resource "nsxt_policy_gateway_policy" "test" {
  display_name    = "tf-gw-policy"
  description     = "Terraform provisioned Gateway Policy"
  domain          = nsxt_policy_domain.france.id
  category        = "LocalGatewayRules"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name       = "rule1"
    destination_groups = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    disabled           = true
    action             = "DROP"
    logged             = true
    scope              = [nsxt_policy_tier1_gateway.policygateway.path]
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

resource "nsxt_policy_gateway_policy" "test" {
  context {
    project_id = data.nsxt_policy_project.demoproj.id
  }
  display_name    = "tf-gw-policy"
  description     = "Terraform provisioned Gateway Policy"
  category        = "LocalGatewayRules"
  locked          = false
  sequence_number = 3
  stateful        = true
  tcp_strict      = false

  tag {
    scope = "color"
    tag   = "orange"
  }

  rule {
    display_name       = "rule1"
    destination_groups = [nsxt_policy_group.group1.path, nsxt_policy_group.group2.path]
    disabled           = true
    action             = "DROP"
    logged             = true
    scope              = [nsxt_policy_tier1_gateway.policygateway.path]
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
* `category` - (Required) The category to use for priority of this Gateway Policy. For local manager must be one of: `Emergency`, `SystemRules`, `SharedPreRules`, `LocalGatewayRules`, `AutoServiceRules` and `Default`. For global manager must be `SharedPreRules` or `LocalGatewayRules`.
* `description` - (Optional) Description of the resource.
* `domain` - (Optional) The domain to use for the Gateway Policy. This domain must already exist. For VMware Cloud on AWS use `cgw`.
* `tag` - (Optional) A list of scope + tag pairs to associate with this Gateway Policy.
* `nsx_id` - (Optional) The NSX ID of this resource. If set, this ID will be used to create the Gateway Policy resource.
* `context` - (Optional) The context which the object belongs to
  * `project_id` - (Required) The ID of the project which the object belongs to
* `comments` - (Optional) Comments for this Gateway Policy including lock/unlock comments.
* `locked` - (Optional) A boolean value indicating if the policy is locked. If locked, no other users can update the resource.
* `sequence_number` - (Optional) An int value used to resolve conflicts between security policies across domains
* `stateful` - (Optional) A boolean value to indicate if this Policy is stateful. When it is stateful, the state of the network connects are tracked and a stateful packet inspection is performed.
* `tcp_strict` - (Optional) A boolean value to enable/disable a 3 way TCP handshake is done before the data packets are sent.
* `rule` (Optional) A repeatable block to specify rules for the Gateway Policy. Each rule includes the following fields:
  * `display_name` - (Required) Display name of the resource.
  * `description` - (Optional) Description of the resource.
  * `destination_groups` - (Optional) Set of group paths that serve as the destination for this rule. IPs, IP ranges, or CIDRs may also be used starting in NSX-T 3.0. An empty set can be used to specify "Any".
  * `destinations_excluded` - (Optional) A boolean value indicating negation of destination groups.
  * `direction` - (Optional) The traffic direction for the policy. Must be one of: `IN`, `OUT` or `IN_OUT`. Defaults to `IN_OUT`.
  * `disabled` - (Optional) A boolean value to indicate the rule is disabled. Defaults to `false`.
  * `ip_version` - (Optional) The IP Protocol for the rule. Must be one of: `IPV4`, `IPV6` or `IPV4_IPV6`. Defaults to `IPV4_IPV6`.
  * `logged` - (Optional) A boolean flag to enable packet logging.
  * `notes` - (Optional) Text for additional notes on changes for the rule.
  * `profiles` - (Optional) A list of context profiles for the rule. Note: due to platform issue, this setting is only supported with NSX 3.2 onwards.
  * `scope` - (Required) List of policy paths where the rule is applied.
  * `services` - (Optional) List of services to match.
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
      * `algorithm` - (Required) Algorithm: `FTP` or `TFTP`
  * `source_groups` - (Optional) Set of group paths that serve as the source for this rule. IPs, IP ranges, or CIDRs may also be used starting in NSX-T 3.0. An empty set can be used to specify "Any".
  * `source_excluded` - (Optional) A boolean value indicating negation of source groups.
  * `log_label` - (Optional) Additional information (string) which will be propagated to the rule syslog.
  * `tag` - (Optional) A list of scope + tag pairs to associate with this Rule.
  * `action` - (Optional) The action for the Rule. Must be one of: `ALLOW`, `DROP` or `REJECT`. Defaults to `ALLOW`.
  * `sequence_number` - (Optional) It is recommended not to specify sequence number for rules, but rather rely on provider to auto-assign them. If you choose to specify sequence numbers, you must make sure the numbers are consistent with order of the rules in configuration. Please note that sequence numbers should start with 1, not 0. To avoid confusion, either specify sequence numbers in all rules, or none at all.

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

An existing Gateway Policy can be [imported][docs-import] into this resource, via the following command:

[docs-import]: https://www.terraform.io/cli/import

```
terraform import nsxt_policy_gateway_policy.gwpolicy1 ID
```

The above command imports the policy Gateway Policy named `gwpolicy1` with the NSX Policy id `ID`.

If the Policy to import isn't in the `default` domain, the domain name can be added to the `ID` before a slash.

For example to import a Group with `ID` in the `MyDomain` domain:

```
terraform import nsxt_policy_gateway_policy.gwpolicy1 MyDomain/ID
```
