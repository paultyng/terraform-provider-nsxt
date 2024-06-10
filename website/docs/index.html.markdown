---
layout: "nsxt"
page_title: "Provider: NSXT"
sidebar_current: "docs-nsxt-index"
description: |-
  The VMware NSX-T Terraform Provider
---

# The NSX Terraform Provider

The NSX Terraform provider gives the NSX administrator a way to automate NSX to
provide virtualized networking and security services using both ESXi and KVM
based hypervisor hosts as well as container networking and security.

More information on NSX can be found on the [NSX Product
Page](https://www.vmware.com/products/nsx.html)

Documentation on the NSX platform can be found on the [NSX Documentation
Page](https://docs.vmware.com/en/VMware-NSX-T/index.html)

Please use the navigation to the left to read about available data sources and
resources.

## Basic Configuration of the NSX Terraform Provider

In order to use the NSX Terraform provider you must first configure the
provider to communicate with the VMware NSX manager. The NSX manager is the
system which serves the NSX REST API and provides a way to configure the
desired state of the NSX system. The configuration of the NSX provider requires
the IP address, hostname, or FQDN of the NSX manager.

The NSX provider offers several ways to authenticate to the NSX manager.
Credentials can be provided statically or provided as environment variables. In
addition, client certificates can be used for authentication. For
authentication with certificates Terraform will require a certificate file and
private key file in PEM format. To use client certificates the client
certificate needs to be registered with NSX-T manager prior to invoking
Terraform.

The provider also can accept both signed and self-signed server certificates.
It is recommended that in production environments you only use certificates
signed by a certificate authority. NSX ships by default with a self-signed
server certificates as the hostname of the NSX manager is not known until the
NSX administrator determines what name or IP to use.

Setting the `allow_unverified_ssl` parameter to `true` will direct the
Terraform client to skip server certificate verification. This is not
recommended in production deployments as it is recommended that you use trusted
connection using certificates signed by a certificate authority.

With the `ca_file` parameter you can also specify a file that contains your
certificate authority certificate in PEM format to verify certificates with a
certificate authority.

There are also a number of other parameters that can be set to tune how the
provider connects to the NSX REST API. It is recommended you leave these to the
defaults unless you experience issues in which case they can be tuned to
optimize the system in your environment.

Note that with terraform 0.14 onwards, `terraform` block should be added to your
configuration:

```hcl
terraform {
  required_providers {
    nsxt = {
      source = "vmware/nsxt"
    }
  }
}
```

Note that in all of the examples you will need to update attributes such as
`host`, `username`, `password`, `vmc_token` to match your NSX deployment.

### Example of Configuration with Credentials

```hcl
provider "nsxt" {
  host                 = "192.168.110.41"
  username             = "admin"
  password             = "default"
  allow_unverified_ssl = true
  max_retries          = 2
}

```


### Example of Setting Environment Variables

```
export NSXT_MANAGER_HOST="192.168.110.41"
export NSXT_USERNAME="admin"
export NSXT_PASSWORD="default"
```

### Example using a Client Certificate

```hcl
provider "nsxt" {
  host                  = "192.168.110.41"
  client_auth_cert_file = "mycert.pem"
  client_auth_key_file  = "mykey.pem"
  allow_unverified_ssl  = true
}

```

### Example with Certificate Authority Certificate

```hcl
provider "nsxt" {
  host     = "10.160.94.11"
  username = "admin"
  password = "qwerty"
  ca_file  = "myca.pem"
}

```

### VMC Environment Example

Note that only a limited subset of policy resources are supported with VMC.

```hcl
provider "nsxt" {
  host                 = "x-54-200-54-5.rp.vmwarevmc.com/vmc/reverse-proxy/api/orgs/b003c3a5-3f68-4a8c-a74f-f79a0625da17/sddcs/d2f43050-f4e2-4989-ab52-2eb0b89d8487/sks-nsxt-manager"
  vmc_token            = "5aVZEj6dJN1bQ6ZheakMyV0Qbj7P65sa2pYuhgx7Mp5glvgCkFKHcGxy3KmslllT"
  allow_unverified_ssl = true
  enforcement_point    = "vmc-enforcementpoint"
}

```

### VMC PCI Compliant Environment Example

```hcl
provider "nsxt" {
  host                 = "10.4.14.23"
  username             = "admin"
  password             = "qwerty"
  vmc_auth_mode        = "Basic"
  allow_unverified_ssl = true
  enforcement_point    = "vmc-enforcementpoint"
}

```

### Policy Global Manager Example

```hcl
provider "nsxt" {
  host            = "192.168.110.41"
  username        = "admin"
  password        = "default"
  global_manager  = true
  max_retries     = 10
  retry_min_delay = 500
  retry_max_delay = 1000
}

```


## Argument Reference

The following arguments are used to configure the VMware NSX-T Provider:

* `host` - (Required) The host name or IP address of the NSX-T manager. Can also
  be specified with the `NSXT_MANAGER_HOST` environment variable. Do not include
  `http://` or `https://` in the host.
* `username` - (Required) The user name to connect to the NSX-T manager as. Can
  also be specified with the `NSXT_USERNAME` environment variable.
* `password` - (Required) The password for the NSX-T manager user. Can also be
  specified with the `NSXT_PASSWORD` environment variable.
* `client_auth_cert_file` - (Optional) The path to a certificate file for client
  certificate authorization. Can also be specified with the
  `NSXT_CLIENT_AUTH_CERT_FILE` environment variable.
* `client_auth_key_file` - (Optional) The path to a private key file for the
  certificate supplied to `client_auth_cert_file`. Can also be specified with
  the `NSXT_CLIENT_AUTH_KEY_FILE` environment variable.
* `client_auth_cert` - (Optional) Client certificate string.
  Can also be specified with the `NSXT_CLIENT_AUTH_CERT` environment variable.
* `client_auth_key` - (Optional) Client certificate private key string.
  Can also be specified with the `NSXT_CLIENT_AUTH_KEY` environment variable.
* `allow_unverified_ssl` - (Optional) Boolean that can be set to true to disable
  SSL certificate verification. This should be used with care as it could allow
  an attacker to intercept your auth token. If omitted, default value is
  `false`. Can also be specified with the `NSXT_ALLOW_UNVERIFIED_SSL`
  environment variable.
* `ca_file` - (Optional) The path to an optional CA certificate file for SSL
  validation. Can also be specified with the `NSXT_CA_FILE` environment
  variable.
* `ca` - (Optional) CA certificate string for SSL validation.
  Can also be specified with the `NSXT_CA` environment variable.
* `max_retries` - (Optional) The maximum number of retires before failing an API
  request. Default: `4` Can also be specified with the `NSXT_MAX_RETRIES`
  environment variable. For Global Manager, it is recommended to increase this value
  since slower realization times tend to delay resolution of some errors.
* `retry_min_delay` - (Optional) The minimum delay, in milliseconds, between
  retries. Default: `0`. For Global Manager, it is recommended to increase this value
  since slower realization times tend to delay resolution of some errors.
  Can also be specified with the `NSXT_RETRY_MIN_DELAY` environment variable.
* `retry_max_delay` - (Optional) The maximum delay, in milliseconds, between
  retries. Default: `500`. For Global Manager, it is recommended to increase this
  value since slower realization times tend to delay resolution of some errors.
  Can also be specified with the `NSXT_RETRY_MAX_DELAY` environment variable.
* `retry_on_status_codes` - (Optional) A list of HTTP status codes to retry on.
  By default, the provider supplies a set of status codes recommended for retry with
  policy resources: `409, 429, 500, 503, 504`. Can also be specified with the
  `NSXT_RETRY_ON_STATUS_CODES` environment variable.
* `remote_auth` - (Optional) Would trigger remote authorization instead of basic
  authorization. This is required for users based on vIDM authentication for early
  NSX versions.
* `session_auth` - (Optional) Creates session to avoid re-authentication for every
  request. Speeds up terraform execution for vIDM based environments. Defaults to `true`
  The default for this flag is false. Can also be specified with the
  `NSXT_REMOTE_AUTH` environment variable.
* `tolerate_partial_success` - (Optional) Setting this flag to true would treat
  partially successful realization as valid state and not fail apply.
* `vmc_token` - (Optional) Long-lived API token for authenticating with VMware
  Cloud Services APIs. This token will be used to short-lived token that is
  needed to communicate with NSX Manager in VMC environment. Can not be specified 
  together with `vmc_client_id` and `vmc_client_secret`.
  Note that only subset of policy resources are supported with VMC environment.
* `vmc_client_id`- (Optional) ID of OAuth App associated with the VMC organization. 
  The combination with `vmc_client_secret` is used to authenticate when calling 
  VMware Cloud Services APIs. Can not be specified together with `vmc_token`.
* `vmc_client_secret` - (Optional) Secret of OAuth App associated with the VMC 
  organization. The combination with `vmc_client_id` is used to authenticate when 
  calling VMware Cloud Services APIs. Can not be specified together with `vmc_token`.
  Note that only subset of policy resources are supported with VMC environment.
* `vmc_auth_host` - (Optional) URL for VMC authorization service that is used
  to obtain short-lived token for NSX manager access. Defaults to VMC
  console authorization URL.
* `vmc_auth_mode` - (Optional) VMC authorization mode, that determines what HTTP
  header is used for authorization. Accepted values are `Default`, `Bearer`, `Basic`.
  For direct VMC connections with a token, use `Bearer` mode. For PCI mode with basic
  authentication, use `Basic`. Otherwise no need to specify this setting.
* `enforcement_point` - (Optional) Enforcement point, mostly relevant for policy
  data sources. For VMC environment, this should be set to `vmc-enforcementpoint`.
  For on-prem deployments, this setting should not be specified.
* `global_manager` - (Optional) True if this is a global manager endpoint.
  False by default.
* `license_keys` - (Optional) List of NSX-T license keys. License keys are applied
  during plan or apply commands. Note that the provider will not remove license keys if
  those are removed from provider config - please clean up licenses manually.
* `on_demand_connection` - (Optional) Avoid verification on NSX connectivity on provider
  startup. Instead, initialize the connection on demand. This setting can not be turned on
  for VMC environments, and is not supported with deprecated NSX manager resources and
  data sources. Note - this setting is useful when NSX manager is not yet available at 
  time of provider evaluation, and not recommended to be turned on otherwise.
* `project_id` - (Optional) ID of the project which the plan executes in its context.
* `vpc_id` - (Optional) ID of the VPC which the plan executes in its context.

## NSX Logical Networking

This release of the NSX-T Terraform Provider extends to cover NSX-T declarative
API called Policy. This API aims to simplify the consumption of logical objects
and offer additional capabilities.The NSX-T Terraform Provider covers most of NSX
functionality.
While you can still build topologies from the imperative API and existing config files
will continue to work, the recommendation
is to build logical topologies from the declarative API(Policy Objects).The resources
and data sources using the policy API have _policy_ in their name.
For more details on the NSX-T Policy API usage, please refer to NSX-T documentation.

The existing data sources and resources are still available to consume but using
the new Policy based data sources and resources are recommended.

### Logical Networking and Security Example Usage

The following example demonstrates using the NSX Terraform provider to create
Tier-1 Gateways, Segments, DHCP Service, Static and Dynamic Groups, Firewall
rules and tags the VMs

#### Example variables.tf File

This file allows you to define some variables that can be reused in multiple
.tf files.

```hcl
variable "nsx_manager" {}
variable "nsx_username" {}
variable "nsx_password" {}
```
#### Example terraform.tfvars File

This file allows you to set some variables that can be reused in multiple .tf
files.

```hcl
nsx_manager = "192.168.110.41"

nsx_username = "admin"

nsx_password = "default"
```


#### Example nsx.tf file

```hcl

################################################################################
#
# This configuration file is an example of creating a full-fledged 3-Tier App
# using Terraform.
#
# It creates the following objects:
#   - Tier-1 Gateway (that gets attached to an existing Tier-0 Gateway)
#   - A DHCP Server providing DHCP Addresses to all 3 Segments
#   - 3 Segments (Web, App, DB)
#   - Dynamic Groups based on VM Tags
#   - Static Group based on IP Addresses
#   - Distributed Firewall Rules
#   - Services
#   - NAT Rules
#   - VM tags
#
# The config has been validated against:
#    NSX-T 3.0 using NSX-T Terraform Provider v2.0
#
# The config below requires the following to be pre-created
#   - Edge Cluster
#   - Overlay Transport Zone
#   - Tier-0 Gateway
#
# It also uses these 3 Services available by default on NSX-T
#   - HTTPS
#   - MySQL
#   - SSH
#
# The configuration also assumes the following Virtual Machines (VMs)
# are available through the vCenter (Compute Manager) Inventory. Assignment
# of the Virtual Machine network to the Segments provided is done outside
# the scope of this example
#   - web-VM
#   - app-VM
#   - db-VM
#
################################################################################


#
# The first step is to configure the VMware NSX provider to connect to the NSX
# REST API running on the NSX manager.
#
provider "nsxt" {
  host                  = var.nsx_manager
  username              = var.nsx_username
  password              = var.nsx_password
  allow_unverified_ssl  = true
  max_retries           = 10
  retry_min_delay       = 500
  retry_max_delay       = 5000
  retry_on_status_codes = [429]
}

#
# Here we show that you define a NSX tag which can be used later to easily to
# search for the created objects in NSX.
#
variable "nsx_tag_scope" {
  default = "project"
}

variable "nsx_tag" {
  default = "terraform-demo"
}


#
# This part of the example shows some data sources we will need to refer to
# later in the .tf file. They include the transport zone, tier 0 router and
# edge cluster.
# There Tier-0 (T0) Gateway is considered a "provider" router that is pre-created
# by the NSX Admin. A T0 Gateway is used for north/south connectivity between
# the logical networking space and the physical networking space. Many Tier1
# Gateways will be connected to the T0 Gateway
#
data "nsxt_policy_edge_cluster" "demo" {
  display_name = "Edge-Cluster"
}

data "nsxt_policy_transport_zone" "overlay_tz" {
  display_name = "Overlay-TZ"
}

data "nsxt_policy_tier0_gateway" "t0_gateway" {
  display_name = "TF-T0-Gateway"
}

#
# Create a DHCP Profile that is used later
# Note, this resource is only in NSX 3.0.0+
resource "nsxt_policy_dhcp_server" "tier_dhcp" {
  display_name     = "tier_dhcp"
  description      = "DHCP server servicing all 3 Segments"
  server_addresses = ["12.12.99.2/24"]
}

#
# In this part of the example, the settings required to create a Tier1 Gateway
# are defined. In NSX a Tier1 Gateway is often used on a per user, tenant,
# department or application basis. Each application may have it's own Tier1
# Gateway. The Tier1 Gateway provides the default gateway for virtual machines
# connected to the Segments on the Tier1 Gateway
#
resource "nsxt_policy_tier1_gateway" "t1_gateway" {
  display_name              = "TF_T1"
  description               = "Tier1 provisioned by Terraform"
  edge_cluster_path         = data.nsxt_policy_edge_cluster.demo.path
  dhcp_config_path          = nsxt_policy_dhcp_server.tier_dhcp.path
  failover_mode             = "PREEMPTIVE"
  default_rule_logging      = "false"
  enable_firewall           = "true"
  enable_standby_relocation = "false"
  force_whitelisting        = "false"
  tier0_path                = data.nsxt_policy_tier0_gateway.t0_gateway.path
  route_advertisement_types = ["TIER1_STATIC_ROUTES", "TIER1_CONNECTED"]
  pool_allocation           = "ROUTING"

  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }

  route_advertisement_rule {
    name                      = "rule1"
    action                    = "DENY"
    subnets                   = ["20.0.0.0/24", "21.0.0.0/24"]
    prefix_operator           = "GE"
    route_advertisement_types = ["TIER1_CONNECTED"]
  }
}

#
# This shows the settings required to create NSX Segment (Logical Switch) to
# which you can attach Virtual Machines (VMs)
#
resource "nsxt_policy_segment" "web" {
  display_name        = "web-tier"
  description         = "Terraform provisioned Web Segment"
  connectivity_path   = nsxt_policy_tier1_gateway.t1_gateway.path
  transport_zone_path = data.nsxt_policy_transport_zone.overlay_tz.path

  subnet {
    cidr        = "12.12.1.1/24"
    dhcp_ranges = ["12.12.1.100-12.12.1.160"]

    dhcp_v4_config {
      server_address = "12.12.1.2/24"
      lease_time     = 36000

      dhcp_option_121 {
        network  = "6.6.6.0/24"
        next_hop = "1.1.1.21"
      }
    }
  }

  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
  tag {
    scope = "tier"
    tag   = "web"
  }
}

resource "nsxt_policy_segment" "app" {
  display_name        = "app-tier"
  description         = "Terraform provisioned App Segment"
  connectivity_path   = nsxt_policy_tier1_gateway.t1_gateway.path
  transport_zone_path = data.nsxt_policy_transport_zone.overlay_tz.path

  subnet {
    cidr        = "12.12.2.1/24"
    dhcp_ranges = ["12.12.2.100-12.12.2.160"]

    dhcp_v4_config {
      server_address = "12.12.2.2/24"
      lease_time     = 36000

      dhcp_option_121 {
        network  = "6.6.6.0/24"
        next_hop = "1.1.1.21"
      }
    }
  }

  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
  tag {
    scope = "tier"
    tag   = "app"
  }
}

resource "nsxt_policy_segment" "db" {
  display_name        = "db-tier"
  description         = "Terraform provisioned DB Segment"
  connectivity_path   = nsxt_policy_tier1_gateway.t1_gateway.path
  transport_zone_path = data.nsxt_policy_transport_zone.overlay_tz.path

  subnet {
    cidr        = "12.12.3.1/24"
    dhcp_ranges = ["12.12.3.100-12.12.3.160"]

    dhcp_v4_config {
      server_address = "12.12.3.2/24"
      lease_time     = 36000

      dhcp_option_121 {
        network  = "6.6.6.0/24"
        next_hop = "1.1.1.21"
      }
    }
  }

  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
  tag {
    scope = "tier"
    tag   = "db"
  }
}

#
# This part of the example shows creating Groups with dynamic membership
# criteria
#
# All Virtual machines with specific tag and scope
resource "nsxt_policy_group" "all_vms" {
  display_name = "All_VMs"
  description  = "Group consisting of ALL VMs"
  criteria {
    condition {
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      key         = "Tag"
      value       = var.nsx_tag

    }
  }
}

# All WEB VMs
resource "nsxt_policy_group" "web_group" {
  display_name = "Web_VMs"
  description  = "Group consisting of Web VMs"
  criteria {
    condition {
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      key         = "Tag"
      value       = "web"
    }
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

# All App VMs
resource "nsxt_policy_group" "app_group" {
  display_name = "App_VMs"
  description  = "Group consisting of App VMs"
  criteria {
    condition {
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      key         = "Tag"
      value       = "app"
    }
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

# All DB VMs
resource "nsxt_policy_group" "db_group" {
  display_name = "DB_VMs"
  description  = "Group consisting of DB VMs"
  criteria {
    condition {
      member_type = "VirtualMachine"
      operator    = "CONTAINS"
      key         = "Tag"
      value       = "db"
    }
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

# Static Group of IP addresses
resource "nsxt_policy_group" "ip_set" {
  display_name = "external_IPs"
  description  = "Group containing all external IPs"
  criteria {
    ipaddress_expression {
      ip_addresses = ["211.1.1.1", "212.1.1.1", "192.168.1.1-192.168.1.100"]
    }
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

#
# An example for Service for App that listens on port 8443
#
resource "nsxt_policy_service" "app_service" {
  display_name = "app_service_8443"
  description  = "Service for App that listens on port 8443"
  l4_port_set_entry {
    description       = "TCP Port 8443"
    protocol          = "TCP"
    destination_ports = ["8443"]
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

#
# Here we have examples of create data sources for Services
#
data "nsxt_policy_service" "https" {
  display_name = "HTTPS"
}

data "nsxt_policy_service" "mysql" {
  display_name = "MySQL"
}

data "nsxt_policy_service" "ssh" {
  display_name = "SSH"
}


#
# In this section, we have example to create Firewall sections and rules
# All rules in this section will be applied to VMs that are part of the
# Gropus we created earlier
#
resource "nsxt_policy_security_policy" "firewall_section" {
  display_name = "DFW Section"
  description  = "Firewall section created by Terraform"
  scope        = [nsxt_policy_group.all_vms.path]
  category     = "Application"
  locked       = "false"
  stateful     = "true"

  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }

  # Allow communication to any VMs only on the ports defined earlier
  rule {
    display_name       = "Allow HTTPS"
    description        = "In going rule"
    action             = "ALLOW"
    logged             = "false"
    ip_version         = "IPV4"
    destination_groups = [nsxt_policy_group.web_group.path]
    services           = [data.nsxt_policy_service.https.path]
  }

  rule {
    display_name       = "Allow SSH"
    description        = "In going rule"
    action             = "ALLOW"
    logged             = "false"
    ip_version         = "IPV4"
    destination_groups = [nsxt_policy_group.web_group.path]
    services           = [data.nsxt_policy_service.ssh.path]
  }

  # Web to App communication
  rule {
    display_name       = "Allow Web to App"
    description        = "Web to App communication"
    action             = "ALLOW"
    logged             = "false"
    ip_version         = "IPV4"
    source_groups      = [nsxt_policy_group.web_group.path]
    destination_groups = [nsxt_policy_group.app_group.path]
    services           = [nsxt_policy_service.app_service.path]
  }

  # App to DB communication
  rule {
    display_name       = "Allow App to DB"
    description        = "App to DB communication"
    action             = "ALLOW"
    logged             = "false"
    ip_version         = "IPV4"
    source_groups      = [nsxt_policy_group.app_group.path]
    destination_groups = [nsxt_policy_group.db_group.path]
    services           = [data.nsxt_policy_service.mysql.path]
  }

  # Allow External IPs to communicate with VMs
  rule {
    display_name       = "Allow Infrastructure"
    description        = "Allow DNS and Management servers"
    action             = "ALLOW"
    logged             = "true"
    ip_version         = "IPV4"
    source_groups      = [nsxt_policy_group.ip_set.path]
    destination_groups = [nsxt_policy_group.all_vms.path]
  }

  # Allow VMs to communicate with outside
  rule {
    display_name  = "Allow out"
    description   = "Outgoing rule"
    action        = "ALLOW"
    logged        = "true"
    ip_version    = "IPV4"
    source_groups = [nsxt_policy_group.all_vms.path]
  }

  # Reject everything else
  rule {
    display_name = "Deny ANY"
    description  = "Default Deny the traffic"
    action       = "REJECT"
    logged       = "true"
    ip_version   = "IPV4"
  }
}

#
# Here we have examples for creating NAT rules. The example here assumes
# the Web IP addresses are reachable from outside and no NAT is required.
#
resource "nsxt_policy_nat_rule" "rule1" {
  display_name        = "App 1-to-1 In"
  action              = "SNAT"
  translated_networks = ["102.10.22.1"] # NAT IP
  source_networks     = ["12.12.2.0/24"]
  gateway_path        = nsxt_policy_tier1_gateway.t1_gateway.path
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

resource "nsxt_policy_nat_rule" "rule2" {
  display_name         = "App 1-to-1 Out"
  action               = "DNAT"
  translated_networks  = ["102.10.22.2"]
  destination_networks = ["102.10.22.1/32"]
  source_networks      = ["12.12.2.0/24"]
  gateway_path         = nsxt_policy_tier1_gateway.t1_gateway.path
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

resource "nsxt_policy_nat_rule" "rule3" {
  display_name        = "DB 1-to-1 In"
  action              = "SNAT"
  translated_networks = ["102.10.23.1"] # NAT IP
  source_networks     = ["12.12.3.0/24"]
  gateway_path        = nsxt_policy_tier1_gateway.t1_gateway.path
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

resource "nsxt_policy_nat_rule" "rule4" {
  display_name         = "App 1-to-1 Out"
  action               = "DNAT"
  translated_networks  = ["102.10.23.3"]
  destination_networks = ["102.10.23.1/32"]
  source_networks      = ["12.12.3.0/24"]
  gateway_path         = nsxt_policy_tier1_gateway.t1_gateway.path
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

# The 3 VMs available in the NSX Inventory
data "nsxt_policy_vm" "web_vm" {
  display_name = "web-vm"
}

data "nsxt_policy_vm" "app_vm" {
  display_name = "app-vm"
}

data "nsxt_policy_vm" "db_vm" {
  display_name = "db-vm"
}


# Assign the right tags to the VMs so that they get included in the
# dynamic groups created above
resource "nsxt_policy_vm_tags" "web_vm_tag" {
  instance_id = data.nsxt_policy_vm.web_vm.instance_id
  tag {
    scope = "tier"
    tag   = "web"
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

resource "nsxt_policy_vm_tags" "app_vm_tag" {
  instance_id = data.nsxt_policy_vm.app_vm.instance_id
  tag {
    scope = "tier"
    tag   = "app"
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}

resource "nsxt_policy_vm_tags" "db_vm_tag" {
  instance_id = data.nsxt_policy_vm.db_vm.instance_id
  tag {
    scope = "tier"
    tag   = "db"
  }
  tag {
    scope = var.nsx_tag_scope
    tag   = var.nsx_tag
  }
}



```

In order to be able to connect VMs to the newly created logical switch a new
`vpshere_network` datasource need to be defined.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "Datacenter"
}

# Data source for the Segments we created earlier
data "vsphere_network" "tf_web" {
  name          = "web-tier"
  datacenter_id = data.vsphere_datacenter.datacenter.id
  depends_on    = [nsxt_policy_segment.web]
}

data "vsphere_network" "tf_app" {
  name          = "app-tier"
  datacenter_id = data.vsphere_datacenter.datacenter.id
  depends_on    = [nsxt_policy_segment.app]
}

data "vsphere_network" "tf_db" {
  name          = "db-tier"
  datacenter_id = data.vsphere_datacenter.datacenter.id
  depends_on    = [nsxt_policy_segment.db]
}
```

The datasource in the above example should be referred in `network_id` inside
`network_interface` section for `vsphere_virtual_machine` resource.

```hcl
resource "vsphere_virtual_machine" "appvm" {
  # ...
  # Attach the VM to the network data source that refers to the newly created Segments
  network_interface {
    network_id = data.vsphere_network.tf_app.id
  }
}
```

## Feature Requests, Bug Reports, and Contributing

For more information how how to submit feature requests, bug reports, or
details on how to make your own contributions to the provider, see the [NSX-T
provider project page][nsxt-provider-project-page].

[nsxt-provider-project-page]: https://github.com/vmware/terraform-provider-nsxt
