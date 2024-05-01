## 3.6.1 (May 2, 2024)

BUG FIXES:
* `resource/nsxt_policy_security_policy`: Fix crash when reading policy that does not exist on backend ([#1202](https://github.com/vmware/terraform-provider-nsxt/pull/1202))
* `resource/nsxt_policy_lb_virtual_server`: Fix SNI path assignment ([#1197](https://github.com/vmware/terraform-provider-nsxt/pull/1197))
* `data/nsxt_policy_vm`, `data/nsxt_policy_vms`, `resource/nsxt_policy_vm_tags`: Fix failure to find certain VMs due to inconsistent pagination ([#1194](https://github.com/vmware/terraform-provider-nsxt/pull/1194))
* `data/nsxt_policy_vm`, `data/nsxt_policy_vms`, `resource/nsxt_policy_vm_tags`: Allow non-admin user to pull VMs and modify VM tags, by using a new API available with NSX 4.1.1 onwards ([#1187](https://github.com/vmware/terraform-provider-nsxt/pull/1187))

IMPROVEMENTS:
* `data/nsxt_policy_vm`, `data/nsxt_policy_vms`, `resource/nsxt_policy_vm_tags`: Improve performance by using search API for NSX 4.1.2 and above ([#1196](https://github.com/vmware/terraform-provider-nsxt/pull/1196))
* `resource/nsxt_policy_ip_block`: Add `visibility` attribute for NSX 4.2.0 onwards ([#1195](https://github.com/vmware/terraform-provider-nsxt/pull/1195))
* `resource/nsxt_policy_vm_tags`: Add support for fixed segments in VM interface tagging ([#1187](https://github.com/vmware/terraform-provider-nsxt/pull/1187))

BUG FIXES ON EXPERIMENTAL FEATURES:
* `resource/nsxt_edge_transport_node`: fix crash when syslog port is specified ([#1190](https://github.com/vmware/terraform-provider-nsxt/pull/1190))

## 3.6.0 (April 16, 2024)

IMPROVEMENTS:
* Improve error messaging on NSX connection initiation ([#1164](https://github.com/vmware/terraform-provider-nsxt/pull/1164))
* `resource/nsxt_policy_nat_rule`: Support `policy_based_vpn_mode` attribute ([#1143](https://github.com/vmware/terraform-provider-nsxt/pull/1143))

BUG FIXES:
* `resource/nsxt_policy_lb_virtual_server`: Fix crash while configuring SSL conditions ([#1156](https://github.com/vmware/terraform-provider-nsxt/pull/1156))
* `resource/nsxt_policy_segment`, `resource/nsxt_policy_vlan_segment`, `resource/nsxt_policy_fixed_segment`: Fix segment import issues related to QoS profile ([#1139](https://github.com/vmware/terraform-provider-nsxt/pull/1139))
* `resource/nsxt_policy_qos_profile`: Fix incorrect state update after failed apply ([#1138](https://github.com/vmware/terraform-provider-nsxt/pull/1138))

IMPROVEMENTS AND BUG FIXES ON EXPERIMENTAL FEATURES:
* `resource/nsxt_edge_transport_node`: fix IP assignment; communicate lack of ipv6 support with appropriate error messaging; split out RTEP config into dedicated resource; allow to configure host switch name; fix timing issues on edge deletion
* `resource/nsxt_policy_host_transport_node`: add `cpu_config` attribute; make sure discovered node ID is assigned
* `data/nsxt_discover_node`: renamed to `data/nsxt_discovered_node`; `compute_manager_state` attribute added

EXPERIMENTAL FEATURES:
* `data/nsxt_policy_vtep_ha_host_switch_profile`
* `data/nsxt_policy_host_transport_node_collection`
* `data/nsxt_policy_host_transport_node_collection_realization`
* `data/nsxt_policy_distributed_flood_protection_profile`
* `data/nsxt_policy_gateway_flood_protection_profile`

* `resource/nsxt_policy_compute_sub_cluster`
* `resource/nsxt_policy_tier0_inter_vrf_routing`
* `resource/nsxt_edge_transport_node_rtep`
* `resource/nsxt_policy_metadata_proxy`
* `resource/nsxt_policy_vtep_ha_host_switch_profile`
* `resource/nsxt_policy_global_manager`
* `resource/nsxt_policy_site`
* `resource/nsxt_policy_distributed_flood_protection_profile`
* `resource/nsxt_policy_distributed_flood_protection_profile_binding`
* `resource/nsxt_policy_gateway_flood_protection_profile`
* `resource/nsxt_policy_gateway_flood_protection_profile_binding`

## 3.5.0 (February 26, 2024)

IMPROVEMENTS:
* Multitenancy support for IDS/IPS resources ([#1096](https://github.com/vmware/terraform-provider-nsxt/pull/1096))
* Support OAuth on VMC ([#1080](https://github.com/vmware/terraform-provider-nsxt/pull/1080))
* `resource/nsxt_policy_tier0_gateway`: support vrf_transit_subnets attribute ([#1026](https://github.com/vmware/terraform-provider-nsxt/pull/1026))

BUG FIXES:
* `resource/nsxt_policy_tier0_gateway`, `resource/nsxt_policy_tier1_gateway`: Fix assignment of ipv6 profiles with multitenancy ([#1093](https://github.com/vmware/terraform-provider-nsxt/pull/1093))
* `data/nsxt_certificate`: Fix an issue when certificate fails to be found by supporting pagination ([#1075](https://github.com/vmware/terraform-provider-nsxt/pull/1075))
* `resource/nsxt_policy_bgp_config`: Support local_as_num for VRF gateways ([#1074](https://github.com/vmware/terraform-provider-nsxt/pull/1074))
* `resource/nsxt_policy_bgp_neighbor`: Clear password only if password was changed by user, this is to avoid resetting the password following import of the resource ([#1046](https://github.com/vmware/terraform-provider-nsxt/pull/1046))
* `resource/nsxt_policy_tier1_gateway`: Respect ha_mode value when NSX auto-assigns it, to avoid non-empty diffs in the provider ([#1042](https://github.com/vmware/terraform-provider-nsxt/pull/1042))
* `resource/nsxt_policy_tier0_gateway`, `resource/nsxt_policy_tier1_gateway`: Allow locking gateway locale_service by specifying nsx_id. This fixes the issue of dependant interfacing being deleted after applying a change in locale service ([#1040](https://github.com/vmware/terraform-provider-nsxt/pull/1040))
* `resource/nsxt_policy_lb_pool`: Allow to specify multiple active monitor paths by adding new attribute and deprecating the old one ([#1039](https://github.com/vmware/terraform-provider-nsxt/pull/1039))
* `data/nsxt_policy_tier0_gateway`, `data/nsxt_policy_edge_cluster`: Fix issue that prevented tenant user to accessing these data sources ([#1015](https://github.com/vmware/terraform-provider-nsxt/pull/1015))

EXPERIMENTAL FEATURES:
Upgrade support is offered as Beta with this release. Please note that upgrade feature requires a separate terraform config and state.
Additional Fabric objects are also offered as Beta in this release.
Additional Beta features include separate security policy rule resource, policy LB profiles and GRE tunnel.

* `data/nsxt_upgrade_postcheck`
* `data/nsxt_upgrade_prepare_ready`
* `data/nsxt_edge_upgrade_group`
* `data/nsxt_host_upgrade_group`

* `data/nsxt_compute_collection`
* `data/nsxt_manager_cluster_node`
* `data/nsxt_discover_node`
* `data/nsxt_edge_transport_node` (renamed from `nsxt_transport_node`)

* `data/nsxt_policy_gateway_dns_forwarder`
* `data/nsxt_policy_gateway_interface_realization`

* `resource/nsxt_upgrade_prepare`
* `resource/nsxt_upgrade_run`
* `resource/nsxt_upgrade_precheck_acknowledge`

* `resource/nsxt_policy_tier0_gateway_gre_tunnel`

* `resource/nsxt_policy_parent_security_policy`
* `resource/nsxt_policy_security_policy_rule`

* `resource/nsxt_policy_lb_client_ssl_profile`
* `resource/nsxt_policy_lb_http_application_profile`
* `resource/nsxt_policy_lb_http_monitor_profile`
* `resource/nsxt_policy_lb_https_monitor_profile`
* `resource/nsxt_policy_lb_icmp_monitor_profile`
* `resource/nsxt_policy_lb_passive_monitor_profile`
* `resource/nsxt_policy_lb_tcp_monitor_profile`
* `resource/nsxt_policy_lb_udp_monitor_profile`

* `resource/nsxt_edge_transport_node` (renamed from `nsxt_transport_node`)

* `resource/nsxt_node_user`
* `resource/nsxt_principal_identity`
* `resource/nsxt_policy_user_management_role_binding`
* `resource/nsxt_policy_user_management_role`
* `resource/policy_ldap_identity_source`

## 3.4.0 (October 27, 2023)

FEATURES:
* Multitenancy support. Supported resources and data sources can now be created within a context of a project. In order to specify a project, use `context` block within resource or data source, and specify `project_id` within. For the full list of supported resources, please refer to [Multitenancy Guide](https://registry.terraform.io/providers/vmware/nsxt/latest/docs/guides/multitenancy).
As part of multitenancy support, project resource and data source are offered, as listed below.

* `data/nsxt_policy_project`

* `resource/nsxt_policy_project`

BUG FIXES:
* `resource/nsxt_policy_security_policy`, `resource/nsxt_policy_gateway_policy`: Validate correctness of sequence numbers only on policy creation, and skip this check on update, but rather auto-correct sequence numbers if needed. This is in order to avoid erroring out in case of incorrect sequence numbers that got assigned with previous provider version ([#1001](https://github.com/vmware/terraform-provider-nsxt/pull/1001))
* Escape all special characters in data sources, as required by search API. This fixes and issue with search by `display_name` that was not working as expected in case it contained some special characters ([#993](https://github.com/vmware/terraform-provider-nsxt/pull/993))

EXPERIMENTAL FEATURES:
Fabric support is offered as Beta with this release:

* `data/nsxt_compute_collection`
* `data/nsxt_compute_manager`
* `data/nsxt_failure_domain`
* `data/nsxt_policy_uplink_host_switch_profile`
* `data/nsxt_transport_node_realization`
* `data/nsxt_compute_manager_realization`

* `resource/nsxt_cluster_virtual_ip`
* `resource/nsxt_compute_manager`
* `resource/nsxt_edge_cluster`
* `resource/nsxt_failure_domain`
* `resource/nsxt_manager_cluster`
* `resource/nsxt_policy_host_transport_node_profile`
* `resource/nsxt_policy_transport_zone`
* `resource/nsxt_transport_node`
* `resource/nsxt_policy_uplink_host_switch_profile`
* `resource/nsxt_policy_host_transport_node_collection`
* `resource/nsxt_edge_high_availability_profile`
* `resource/nsxt_policy_host_transport_node`
* `resource/nsxt_node_user`
* `resource/nsxt_policy_user_management_role_binding`
* `resource/nsxt_policy_user_management_role`
* `resource/nsxt_policy_transport_zone`

## 3.3.2 (September 22, 2023)

IMPROVEMENTS:
* Support on-demand connection init in the provider. This behavior is controlled with `on_demand_connection` flag and is useful is NSX manager is not available at the time of plan/apply ([#948](https://github.com/vmware/terraform-provider-nsxt/pull/948))
* `resource/nsxt_policy_tier1_gateway`: Support `type` argument. This argument helps with auto-configuring route advertisements and provides the user experience that is consistent with UI on VMC ([#909](https://github.com/vmware/terraform-provider-nsxt/pull/909))
* Improve debug logging by dumping NSX API requests and responses when `TF_LOG_PROVIDER_NSX_HTTP` env variable is set ([#963](https://github.com/vmware/terraform-provider-nsxt/pull/963))

BUG FIXES:
* `resource/nsxt_policy_security_policy`, `resource/nsxt_policy_gateway_policy`: Fix rule ordering issue by auto-assigning `sequence_number`. ([#967](https://github.com/vmware/terraform-provider-nsxt/pull/967))
* `resource/nsxt_policy_group`: Fix `group_type` assignment on VMC by using `node/version` API to determine underlying NSX version ([#970](https://github.com/vmware/terraform-provider-nsxt/pull/970))
* `resource/nsxt_nat_rule`: Ensure compatibility with NSX 4.1.0 and above by replacing removed 'nat_pass' property with 'firewall_match' ([#950](https://github.com/vmware/terraform-provider-nsxt/pull/950))

EXPERIMENTAL FEATURES:
* `data/nsxt_policy_gateway_prefix_list`
* `data/nsxt_policy_gateway_route_map`
* `data/nsxt_policy_project`

* `resource/nsxt_policy_vni_pool`
* `resource/nsxt_policy_project`

* Multitenancy support in selected resources, controlled by `context` argument
* Fabric resources and data sources (detailed list coming with next feature release)

## 3.3.1 (May 30, 2023)

FEATURES:

* **New Data Source**: `nsxt_policy_segment`.

* **New Resource**: `nsxt_policy_ip_discovery_profile`.
* **New Resource**: `nsxt_policy_gateway_qos_profile`.
* **New Resource**: `nsxt_policy_segment_security_profile`.
* **New Resource**: `nsxt_policy_spoof_guard_profile`.
* **New Resource**: `nsxt_policy_context_profile_custom_attribute`.

IMPROVEMENTS:
* `resource/nsxt_policy_ip_address_allocation`: Avoid recreation of resource if descriptive properties like `display_name` are updated ([#892](https://github.com/vmware/terraform-provider-nsxt/pull/892))
* `resource/nsxt_policy_gateway_dns_forwarder`: Add `cache_size` property ([#889](https://github.com/vmware/terraform-provider-nsxt/pull/889))
* `resource/nsxt_policy_dhcp_relay`: Add Global Manager support ([#883](https://github.com/vmware/terraform-provider-nsxt/pull/883))
* `resource/nsxt_policy_group`: Align enumeration values for `key`, `member_type`, `operator` with latest NSX spec. This would allow to configure values that were previously blocked by provider validation ([#882](https://github.com/vmware/terraform-provider-nsxt/pull/882))
* `resource/nsxt_policy_group`: Add `group_type` property ([#857](https://github.com/vmware/terraform-provider-nsxt/pull/857))
* `resource/nsxt_policy_gateway_policy`: Disallow creating policy in Read-Only category ([#860](https://github.com/vmware/terraform-provider-nsxt/pull/860))
* `resource/nsxt_policy_tier1_gateway`: Add `ha_mode` property ([#856](https://github.com/vmware/terraform-provider-nsxt/pull/856))
* `resource/nsxt_policy_context_profile`: Add support for custom URLs ([#840](https://github.com/vmware/terraform-provider-nsxt/pull/840))
* `resource/nsxt_policy_context_profile`: Add `custom_url_partial_match` property ([#850](https://github.com/vmware/terraform-provider-nsxt/pull/850))
* `resource/nsxt_policy_service`: Add support for nested service ([#836](https://github.com/vmware/terraform-provider-nsxt/pull/836))
* `resource/nsxt_policy_ip_discovery_profile`: Add support for `tofu_enabled` property ([#834](https://github.com/vmware/terraform-provider-nsxt/pull/834))
* `data/nsxt_policy_vms`: Add ability to filter Virtual Machines by `state` and `guest_os` ([#869](https://github.com/vmware/terraform-provider-nsxt/pull/869))
* Switch to new set of API for VPN objects (old set of API are deprecated on NSX). With new APIs, VPN objects are located directly under a gateway rather than under locale service as before. Deprecated API are still supported. ([#866](https://github.com/vmware/terraform-provider-nsxt/pull/866))
* Support session authentication for policy object. This support significantly improves performance for vIDM environments. This setting is controlled by `session_auth` provider property, and is enabled by default ([#846](https://github.com/vmware/terraform-provider-nsxt/pull/846))

BUG FIXES:

* `resource/nsxt_policy_ipsec_vpn_session`: Allow configuring `compliance_suite` ([#891](https://github.com/vmware/terraform-provider-nsxt/pull/891))
* `resource/nsxt_policy_ipsec_vpn_session`: Fix import for Policy-Based session ([#864](https://github.com/vmware/terraform-provider-nsxt/pull/864))
* `resource/nsxt_policy_security_policy`: Fix configuration of `Ethernet` category ([#844](https://github.com/vmware/terraform-provider-nsxt/pull/844))
* `resource/nsxt_policy_lb_virtual_server`: Fix a bug in detecting rule changes ([#843](https://github.com/vmware/terraform-provider-nsxt/pull/843))
* `resource/nsxt_policy_tier0_gateway`, `resource/nsxt_policy_tier1_gateway`: Ensure ordered list in `preferred_edge_paths` setting. This allows changing order of edge nodes ([#829](https://github.com/vmware/terraform-provider-nsxt/pull/829))


## 3.3.0 (January 16, 2023)

FEATURES:

* **New Data Source**: `nsxt_policy_gateway_locale_service`.
* **New Data Source**: `nsxt_policy_bridge_profile`.
* **New Data Source**: `nsxt_policy_ipsec_vpn_local_endpoint`.
* **New Data Source**: `nsxt_policy_ipsec_vpn_service`.
* **New Data Source**: `nsxt_policy_l2_vpn_service`.

* **New Resource**: `nsxt_policy_ipsec_vpn_ike_profile`.
* **New Resource**: `nsxt_policy_ipsec_vpn_tunnel_profile`.
* **New Resource**: `nsxt_policy_ipsec_vpn_dpd_profile`.
* **New Resource**: `nsxt_policy_ipsec_vpn_session`.
* **New Resource**: `nsxt_policy_ipsec_vpn_service`.
* **New Resource**: `nsxt_policy_ipsec_vpn_local_endpoint`.
* **New Resource**: `nsxt_policy_l2_vpn_session`.
* **New Resource**: `nsxt_policy_l2_vpn_service`.

EXPERIMENTAL FEATURES:

* **New Resource**: `nsxt_policy_ip_discovery_profile`.

BUG FIXES:

* `resource/nsxt_policy_gateway_route_map`: Fix issues around `local_preference` and `med` attributes in route map set clause. Those values were assigned incorrect default value when not specified by terraform user ([#818](https://github.com/vmware/terraform-provider-nsxt/pull/818))

DEPRECATIONS:

In this release we deprecate non-policy data sources and resources. Please use corresponding policy resources instead.

## 3.2.9 (October 28, 2022)

BUG FIXES:

* `resource/nsxt_policy_lb_virtual_server`: Fix change detection for rules. This solves a bug that resulted in rule change not being applied ([#774](https://github.com/vmware/terraform-provider-nsxt/pull/774))
* `resource/nsxt_policy_bgp_neighbor`: Fix import functionality on Global Manager ([#796](https://github.com/vmware/terraform-provider-nsxt/pull/796))

IMPROVEMENTS:

* `resource/nsxt_policy_security_policy`, `resource/nsxt_policy_gateway_policy`: Only update rules that have non-empty diff. Previously, all rules would be updated as part of parent policy resource, which would cause rule statistics to reset and unnecessarily increase realization time ([#786](https://github.com/vmware/terraform-provider-nsxt/pull/786))
* `resource/nsxt_policy_segment`, `resource/nsxt_policy_vlan_segment`: Support Bridge configuration on segments ([#784](https://github.com/vmware/terraform-provider-nsxt/pull/784))
* `resource/nsxt_policy_segment`, `resource/nsxt_policy_vlan_segment`: Support replication mode on segments ([#779](https://github.com/vmware/terraform-provider-nsxt/pull/779))

## 3.2.8 (June 20, 2022)

BUG FIXES:

* `resource/nsxt_policy_bgp_config`: Avoid assigning irrelevant values for VRF configurations, since those cause NSX validation error, even if those values are set to default ([#756](https://github.com/vmware/terraform-provider-nsxt/pull/756))
* `resource/nsxt_policy_bgp_config`: Mark `local_as_num` as Computed, this ensures configuration consistency for VRF use case where BGP configuration is inherited ([#762](https://github.com/vmware/terraform-provider-nsxt/pull/762))
* `resource/nsxt_policy_bgp_config`: Fix segmentation fault when edge cluster is not set on Gateway ([#756](https://github.com/vmware/terraform-provider-nsxt/pull/756))
* `resource/nsxt_policy_nat_rule`: Fix REFLECTIVE NAT rule configuration ([#759](https://github.com/vmware/terraform-provider-nsxt/pull/759))
* `resource/nsxt_policy_nat_rule`: Change translated_network attribute definition from Required to Optional for sake of NO_SNAT/NO_DNAT rule types([#759](https://github.com/vmware/terraform-provider-nsxt/pull/759))

IMPROVEMENTS:

* Include object scope (LM or GM) in policy search for data sources. This improvement would narrow down object search to scope relevant to current backend, for example, Global Manager objects will not show up in data source query on Local Manager([#755](https://github.com/vmware/terraform-provider-nsxt/pull/755))
* (Exterimental) Support `locale_service` configuration on Local Manager. This offers more flexibility to specify locale configuration such as edge cluster, preferred node, redistribution, and should not be used together with `edge_cluster_path` argument. Previsouly this clause was only supported on Global Manager([#764](https://github.com/vmware/terraform-provider-nsxt/pull/764))

## 3.2.7 (May 12, 2022)

BUG FIXES:

* Fix potential segmentation fault in API retry ([#746](https://github.com/vmware/terraform-provider-nsxt/pull/746))
* Support special characters in IDs of NSX objects in data source search ([#751](https://github.com/vmware/terraform-provider-nsxt/pull/751))

IMPROVEMENTS:
* `resource/nsxt_policy_group`: Support external IDs ([#733](https://github.com/vmware/terraform-provider-nsxt/pull/733))
* `resource/nsxt_policy_tier1_gateway`: Improve error handling in delete ([#746](https://github.com/vmware/terraform-provider-nsxt/pull/746))


## 3.2.6 (April 8, 2022)

BUG FIXES:

* `resource/nsxt_policy_nat_rule`: Support NAT64 action ([#725](https://github.com/vmware/terraform-provider-nsxt/pull/725))
* Segment resources: Fix IP pool assignment ([#712](https://github.com/vmware/terraform-provider-nsxt/pull/712))
* Fix SDK bug with Cookie header assignment in session create API for MP resources. This fix can improve performance in VIDM environments ([#730](https://github.com/vmware/terraform-provider-nsxt/pull/730))

EXPERIMENTAL FEATURES:

* **New Data Source**: `nsxt_policy_vms`. This data source is populated with a map of all VMS in inventory, and can be used as an alternative for `nsxt_policy_vm` to address scale issues.
* **New Resource**: `nsxt_policy_mac_discovery_profile`

IMPROVEMENTS:

* Support global retry for policy resources. Retry parameters are configured in provider section - please refer to documentation ([#708](https://github.com/vmware/terraform-provider-nsxt/pull/708))
* `resource/resource_nsxt_policy_lb_virtual_server`: Support rules for this resource ([#676](https://github.com/vmware/terraform-provider-nsxt/pull/676))
* `resource/nsxt_policy_vm_tags`: Avoid erroring out on refresh/apply if given VM no longer exists on backend. Note that `nsxt_policy_vm` data source would still error out when VM is not found, so users seeking behavior where `not found` error is swallowed, are encouraged to use `nsxt_policy_vms` data source instead, and look up VM names in `items` map ([#718](https://github.com/vmware/terraform-provider-nsxt/pull/718))


## 3.2.5 (October 15, 2021)

BUG FIXES:

* `resource/nsxt_policy_gateway_route_map`: Allow multiple areas in AS path validation ([#666](https://github.com/vmware/terraform-provider-nsxt/pull/666))
* `resource/nsxt_policy_intrusion_service_profile`: Fix potential non-empty plan issues by switching argument type from List to Set where appropriate. This fix is relevant with NSX 3.2 onwards ([#684](https://github.com/vmware/terraform-provider-nsxt/pull/684))

EXPERIMENTAL FEATURES:

* **New Data Source**: `nsxt_policy_lb_service`.
* **New Data Source**: `nsxt_ns_groups`. This data source is introduced to address scale issues. Please note this data source uses non-policy (Manager) API and should only be used with features that have limited Policy support.
* **New Data Source**: `nsxt_ns_services`. This data source is introduced to address scale issues. Please note this data source uses non-policy (Manager) API and should only be used with features that have limited Policy support.

IMPROVEMENTS:

* `resource/nsxt_policy_gateway_redistribution_config`: Add `bgp` and `ospf` markers to redistribution rules([#673](https://github.com/vmware/terraform-provider-nsxt/pull/673))
* Introduce retries in selected resources to avoid most common deletion syncronization issues. This measure is temporary until provider-wide retry is implemented with SDK enhancement ([#681](https://github.com/vmware/terraform-provider-nsxt/pull/681), [#686](https://github.com/vmware/terraform-provider-nsxt/pull/686), [#687](https://github.com/vmware/terraform-provider-nsxt/pull/687))

NSX 3.2.0 NOTES:

* `data/nsxt_policy_edge_node`: Policy API for edge node has changed in NSX 3.2.0. While backwards compatibility is not broken with this resource, it is important to note that `path` attribute to edge node no longer reflects node UUID, but rather its ordinal value ([#679](https://github.com/vmware/terraform-provider-nsxt/pull/679))
* Policy Segment resources: Following change in populating `advanced_config` segment sub-clause, there is a new nuance while importing segment resources with NSX 3.2.0. If you wish to import `advanced_config` settings, `advanced_config` needs to be specified in your terraform configuration prior to importing ([#671](https://github.com/vmware/terraform-provider-nsxt/pull/671))

## 3.2.4 (August 26, 2021)

BUG FIXES:

* Fix pagination for non-policy data sources. This fix is relevant for big scale (1K+ per object type) environments where non-policy data sources are still used ([#656](https://github.com/vmware/terraform-provider-nsxt/pull/656))
* `resource/nsxt_policy_tier0_gateway`: Fix potential apply error on Global Manager ([#659](https://github.com/vmware/terraform-provider-nsxt/pull/659))
* `resource/nsxt_policy_nat_rule`: Fix potential non-empty diff when multiple scopes are used ([#655](https://github.com/vmware/terraform-provider-nsxt/pull/655))
* `data/nsxt_policy_certificate`: Fix broken functionality on Global Manager ([#653](https://github.com/vmware/terraform-provider-nsxt/pull/653))

## 3.2.3 (July 29, 2021)

BUG FIXES:

* Fix import functionality for Tier0 resources that were preconfigured with non-default locale service ([#648](https://github.com/vmware/terraform-provider-nsxt/pull/648))

## 3.2.2 (June 21, 2021)

BUG FIXES:

* Fix realization issue for gateway route map ([#640](https://github.com/vmware/terraform-provider-nsxt/pull/640))

## 3.2.1 (June 10, 2021)

BUG FIXES:

* Fix provider compatibility with NSX 3.0.x version line ([#636](https://github.com/vmware/terraform-provider-nsxt/pull/636))

## 3.2.0 (June 7, 2021)

FEATURES:

* **New Data Source**: `nsxt_policy_bfd_profile`.
* **New Data Source**: `nsxt_policy_intrusion_service_profile`.

* **New Resource**: `nsxt_policy_dns_forwarder_zone`.
* **New Resource**: `nsxt_policy_gateway_dns_forwarder`.
* **New Resource**: `nsxt_policy_intrusion_service_profile` (Local Manager only).
* **New Resource**: `nsxt_policy_intrusion_service_policy` (Local Manager only).
* **New Resource**: `nsxt_policy_gateway_community_list`.
* **New Resource**: `nsxt_policy_fixed_segment` (VMC only).
* **New Resource**: `nsxt_policy_dns_forwarder_zone`.
* **New Resource**: `nsxt_policy_gateway_dns_forwarder`.
* **New Resource**: `nsxt_policy_gateway_community_list`.
* **New Resource**: `nsxt_policy_gateway_route_map`.
* **New Resource**: `nsxt_policy_static_route_bfd_peer`.
* **New Resource**: `nsxt_policy_evpn_tenant` (Local Manager only).
* **New Resource**: `nsxt_policy_evpn_config` (Local Manager only).
* **New Resource**: `nsxt_policy_evpn_tunnel_endpoint` (Local Manager only).
* **New Resource**: `nsxt_policy_ospf_config` (Local Manager only).
* **New Resource**: `nsxt_policy_ospf_area` (Local Manager only).
* **New Resource**: `nsxt_policy_gateway_redistribution_config`.
* **New Resource**: `nsxt_policy_qos_profile`.

IMPROVEMENTS:

* `resource/nsxt_policy_fixed_segment`: Add support for dhcp static bindings(([#557](https://github.com/vmware/terraform-provider-nsxt/pull/557))
* `resource/nsxt_policy_bgp_config`: Add support for Local Manager(([#572](https://github.com/vmware/terraform-provider-nsxt/pull/572))
* Support basic auth mode for VMC PCI use case([#577](https://github.com/vmware/terraform-provider-nsxt/pull/577))
* Security Policy and Gateway policy resources: Allow IP CIDR or Range as source/dest groups([#589](https://github.com/vmware/terraform-provider-nsxt/pull/589))
* `data/nsxt_policy_realization_info`: Introduce timeout and delay realization arguments([#590](https://github.com/vmware/terraform-provider-nsxt/pull/590))
* Segment resources: Support urpf_mode in advanced config (([#627](https://github.com/vmware/terraform-provider-nsxt/pull/627))
* Support darwin arm64 release([#628](https://github.com/vmware/terraform-provider-nsxt/pull/628))

BUG FIXES:

* `data/nsxt_policy_vm`: Fix fetching by `display_name` by adding pagination support([#570](https://github.com/vmware/terraform-provider-nsxt/pull/570))
* `resource/nsxt_policy_nat_rule`: Fix source network assignment that caused API error on VMC([#575](https://github.com/vmware/terraform-provider-nsxt/pull/575))
* Gateway resources: Fix ipv6 profiles assignment on Global Manager ([#582](https://github.com/vmware/terraform-provider-nsxt/pull/582))
* `data/nsxt_policy_group`: Fix fetching by `display_name` by adding pagination support([#586](https://github.com/vmware/terraform-provider-nsxt/pull/586))
* `resource/nsxt_policy_tier0_gateway`: Fix VRF realization error due to empty route configuration([#588](https://github.com/vmware/terraform-provider-nsxt/pull/588))
* `resource/nsxt_policy_group`: Fix provider crush due to empty configuration([#607](https://github.com/vmware/terraform-provider-nsxt/pull/607))
* `resource/nsxt_policy_fixed_segment`: Define `transport_zone_path` as optional force-new argument, rather than required([#617](https://github.com/vmware/terraform-provider-nsxt/pull/617))
* `resource/nsxt_policy_static_route`: Define `ip_address` as optional argument, rather than required([#621](https://github.com/vmware/terraform-provider-nsxt/pull/621))

DEPRECATIONS:

* `resource/nsxt_policy_tier0_gateway`: `redistribution_config` clause is now deprecated. Please use `nsxt_policy_gateway_redistribution_config` resource instead.


## 3.1.1 (January 4, 2021)

FEATURES:

* **New Resource**: `nsxt_policy_dhcp_server`.
* **New Resource**: `nsxt_policy_domain` (Global Manager only).
* **New Resource**: `nsxt_policy_dhcp_v4_static_binding`.
* **New Resource**: `nsxt_policy_dhcp_v6_static_binding`.

EXPERIMENTAL FEATURES:

* **New Data Source**: `nsxt_policy_bfd_profile`.

* **New Resource**: `nsxt_policy_dns_forwarder_zone`.
* **New Resource**: `nsxt_policy_gateway_dns_forwarder`.
* **New Resource**: `nsxt_policy_intrusion_service_policy`.
* **New Resource**: `nsxt_policy_gateway_community_list`.
* **New Resource**: `nsxt_policy_fixed_segment` (VMC only).

IMPROVEMENTS:

* New provider attributes `client_auth_cert`, `client_auth_key` to allow passing these values as string rather than a file ([#524](https://github.com/vmware/terraform-provider-nsxt/pull/524))
* Allow Bearer token authorization type for VMC deployments (Experimental). This behavior is configured by setting new provider attribute `vmc_auth_mode` to `Bearer` ([#539](https://github.com/vmware/terraform-provider-nsxt/pull/539))
* Complete Global Manager support for data sources (T1 Gateway, IPv6 Profiles, Ceritificate)
* `resource/nsxt_policy_tier1_gateway`: Enhance T0 Gateway resource with `rd_admin_address` attribute ([#503](https://github.com/vmware/terraform-provider-nsxt/pull/503))
* `resource/nsxt_policy_predefined_gateway_policy`: Add Importer for this resource to match user expectations in case predefined rules exist. Documentation was also extended to cover import and no-import usage ([#527](https://github.com/vmware/terraform-provider-nsxt/pull/527))


BUG FIXES:
* Allow maximum subnet length in Gateway Interface validation ([#528](https://github.com/vmware/terraform-provider-nsxt/pull/528))
* Make sure policy data sources ignore deleted objects ([#516](https://github.com/vmware/terraform-provider-nsxt/pull/516))
* `resource/nsxt_policy_segment`: Allow configuration of segment on Global Manager without transport zone ([#513](https://github.com/vmware/terraform-provider-nsxt/pull/513)).
* Determine major NSX version behind VMC deployment, thus making 3.0.0 features (such as segment DHCP) available for VMC. This requires a more robust solution in futire ([#531](https://github.com/vmware/terraform-provider-nsxt/pull/531)).

## 3.1.0 (October 20, 2020)

FEATURES:

* **New Data Source**: `nsxt_policy_security_policy`.
* **New Data Source**: `nsxt_policy_gateway_policy`.
* **New Data Source**: `nsxt_policy_group`.
* **New Data Source**: `nsxt_policy_context_profile` (official support).

* **New Resource**: `nsxt_policy_context_profile` (official support).
* **New Resource**: `nsxt_policy_tier0_gateway_ha_vip_config` (official support).
* **New Resource**: `nsxt_policy_gateway_prefix_list` (official support).

EXPERIMENTAL FEATURES:

* **New Data Source**: `nsxt_management_cluster`.

* **New Resource**: `nsxt_policy_predefined_security_policy`. This resource allows users to modify default security policy. Please refer to docs for more details.
* **New Resource**: `nsxt_policy_predefined_gateway_policy`. This resource enables gateway policy configuration for VMC. Please refer to docs for more details.

IMPROVEMENTS:

* Allow specifying `vlan_ids` for overlay segments ([#462](https://github.com/vmware/terraform-provider-nsxt/pull/462))
* Allow specifying NSX license via provider attribute. Note: the lisence is not considered part of configuration, and is applied at plan time! ([#423](https://github.com/vmware/terraform-provider-nsxt/pull/423))

BUG FIXES:

* `resource/nsxt_policy_tier0_gateway`: Fix non-empty state issue for VRF use case ([#478](https://github.com/vmware/terraform-provider-nsxt/pull/478))
* `resource/nsxt_policy_segment`: Fix a bug with `excluded_range` assignment ([#473](https://github.com/vmware/terraform-provider-nsxt/pull/473))
* `resource/nsxt_policy_lb_pool`: Fix read function for `member_group` attribute ([#473](https://github.com/vmware/terraform-provider-nsxt/pull/473))
* `resource/nsxt_policy_ip_address_allocation`: Fix address allocation with older NSX versions ([#468](https://github.com/vmware/terraform-provider-nsxt/pull/468))
* `data/nsxt_policy_realization_info`: Fix realization polling with older NSX versions ([#468](https://github.com/vmware/terraform-provider-nsxt/pull/468))
* `data/nsxt_ns_group`: Add pagination support to fix group retrieval with many group objects defined ([#440](https://github.com/vmware/terraform-provider-nsxt/pull/440))
* `resource/nsxt_policy_lb_virtual_server`: Preserve existing rules that are defined outside terraform ([#482](https://github.com/vmware/terraform-provider-nsxt/pull/482))

## 3.0.0 (August 24, 2020)

* The provider is extended to support NSXT Global Manager. Only a subset of objects is supported, check the documentation for more details.

* **New Data Source**: `nsxt_policy_site`. Applicable for NSX Global Manager only.

* **New Resource**: `nsxt_policy_bgp_config`. Applicable for NSX Global Manager only.

EXPERIMENTAL FEATURES:

* **New Data Source**: `nsxt_policy_context_profile`.

* **New Resource**: `nsxt_policy_context_profile`.
* **New Resource**: `nsxt_policy_tier0_gateway_ha_vip_config`.
* **New Resource**: `nsxt_policy_gateway_prefix_list`.

IMPROVEMENTS:

* Improve error handling for policy resources. This fixes some scenarios (mostly relevant for VMConAWS) where error was swallowed by the provider ([#428] (https://github.com/vmware/terraform-provider-nsxt/pull/428))
* Improve provider host validation and allow schema to be specified ([#413](https://github.com/vmware/terraform-provider-nsxt/pull/413))
* `resource/nsxt_policy_vm_tags`: Support tagging specific logical port on the VM, based on segment path ([#406](https://github.com/vmware/terraform-provider-nsxt/pull/406))
* `resource/nsxt_policy_group`: Support MAC address criteria ([#388](https://github.com/vmware/terraform-provider-nsxt/pull/388))
* `resource/nsxt_policy_segment`, `resource/nsxt_policy_vlan_segment`: Support assigning custome segment profiles ([#384](https://github.com/vmware/terraform-provider-nsxt/pull/384))
* `resource/nsxt_policy_segment`, `resource/nsxt_policy_vlan_segment`: Wait for VM ports to be deleted before proceeding with segment delete. This avoids potential dependency error on deletion ([#311](https://github.com/vmware/terraform-provider-nsxt/pull/311))
* `resource/nsxt_policy_vlan_segment`: Allow specifying vlan range ([#342](https://github.com/vmware/terraform-provider-nsxt/pull/342))
* `resource/nsxt_policy_tier0_gateway`: Support assigning custom segment profiles ([#363](https://github.com/vmware/terraform-provider-nsxt/pull/363))

BUG FIXES:

* Fix to bypass certificate validation against cert request ([#381](https://github.com/vmware/terraform-provider-nsxt/pull/381))
* Fix potential crashes in some policy resources ([#305](https://github.com/vmware/terraform-provider-nsxt/pull/305))
* `resource/nsxt_policy_segment`: Fix error reporting on segment deletion ([#321](https://github.com/vmware/terraform-provider-nsxt/pull/321))
* `resource/nsxt_policy_vlan_segment`: Allow to specify zero as vlan id ([#304](https://github.com/vmware/terraform-provider-nsxt/pull/304))
* `resource/nsxt_policy_bgp_neighbor`: Fix route filters configuration ([#387](https://github.com/vmware/terraform-provider-nsxt/pull/387))
* `resource/nsxt_ip_pool_allocation_ip_address`: Fix import ([#319](https://github.com/vmware/terraform-provider-nsxt/pull/319))

## 2.1.0 (June 09, 2020)

* The provider is extended to support NSXT on VMConAWS. Only a subset of objects is supported, check the documentation for more details.

BUG FIXES:

* Fix remote authentication(vIDM) for policy objects. This fix is relevant for NSX version below 3.0.0. ([#302](https://github.com/vmware/terraform-provider-nsxt/pull/302))
* Fix client certificate authentication for policy objects ([#292](https://github.com/vmware/terraform-provider-nsxt/pull/292))
* Fix an issue related to non-admin NSX credentials ([#293](https://github.com/vmware/terraform-provider-nsxt/pull/293))
* `resource/nsxt_policy_vlan_segment`: Allow to specify vlan range ([#342](https://github.com/vmware/terraform-provider-nsxt/pull/342))
* `resource/nsxt_policy_segment`: Fix handling of segment deletion error ([#321](https://github.com/vmware/terraform-provider-nsxt/pull/321))
* `resource/nsxt_policy_segment`: Wait for potential VMs to free segment port before deleting the segment. ([#311](https://github.com/vmware/terraform-provider-nsxt/pull/311))
* `resource/nsxt_policy_vlan_segment`: Allow zero vlan ID ([#297](https://github.com/vmware/terraform-provider-nsxt/pull/297))
* `resource/nsxt_policy_tierX_gateway_interface`: Fix a use case of preconfigured locale service on gateway ([#300](https://github.com/vmware/terraform-provider-nsxt/pull/300))
* `resource/nsxt_policy_security_policy`: Fix import crash ([#299](https://github.com/vmware/terraform-provider-nsxt/pull/299))
* `resource/nsxt_policy_security_policy`: Expose `log_label` argument ([#298](https://github.com/vmware/terraform-provider-nsxt/pull/298))
* `resource/nsxt_policy_group`: Fix issues with group subresource import ([#288](https://github.com/vmware/terraform-provider-nsxt/pull/288))
* `resource/nsxt_policy_nat_rule`: Make `source_networks` argument optional ([#294](https://github.com/vmware/terraform-provider-nsxt/pull/294))
* `resource/nsxt_ip_pool_allocation_ip_address`: Fix resource import ([#319](https://github.com/vmware/terraform-provider-nsxt/pull/319))
* `data/nsxt_policy_segment_realization`: Expose computed attribute network_name. This attribute can be used as network name in vsphere provider, which forms the necessary dependency ([#308](https://github.com/vmware/terraform-provider-nsxt/pull/308))

## 2.0.0 (March 30, 2020)

NOTES:

* The provider is extended to support NSX-T policy API. Policy API is intended to be primary consumtion for NSX-T logical constructs, thus users are encouraged to use new data sources/resources, with _policy_ in the name.

FEATURES:
* **New Data Source**: `nsxt_policy_certificate`
* **New Data Source**: `nsxt_policy_edge_cluster`
* **New Data Source**: `nsxt_policy_edge_node`
* **New Data Source**: `nsxt_policy_tier0_gateway`
* **New Data Source**: `nsxt_policy_tier1_gateway`
* **New Data Source**: `nsxt_policy_segment`
* **New Data Source**: `nsxt_policy_vlan_segment`
* **New Data Source**: `nsxt_policy_service`
* **New Data Source**: `nsxt_policy_ip_discovery_profile`
* **New Data Source**: `nsxt_policy_spoofguard_profile`
* **New Data Source**: `nsxt_policy_qos_profile`
* **New Data Source**: `nsxt_policy_segment_security_profile`
* **New Data Source**: `nsxt_policy_mac_discovery_profile`
* **New Data Source**: `nsxt_policy_ipv6_ndra_profile`
* **New Data Source**: `nsxt_policy_ipv6_dad_profile`
* **New Data Source**: `nsxt_policy_vm`
* **New Data Source**: `nsxt_policy_lb_app_profile`
* **New Data Source**: `nsxt_policy_lb_client_ssl_profile`
* **New Data Source**: `nsxt_policy_lb_server_ssl_profile`
* **New Data Source**: `nsxt_policy_lb_monitor`
* **New Data Source**: `nsxt_policy_lb_persistence_profile`
* **New Data Source**: `nsxt_policy_vni_pool`
* **New Data Source**: `nsxt_policy_realization_info`
* **New Data Source**: `nsxt_policy_segment_realization`
* **New Data Source**: `nsxt_firewall_section`

* **New Resource**: `nsxt_policy_tier0_gateway`
* **New Resource**: `nsxt_policy_tier1_gateway`
* **New Resource**: `nsxt_policy_tier0_gateway_interface`
* **New Resource**: `nsxt_policy_tier1_gateway_interface`
* **New Resource**: `nsxt_policy_group`
* **New Resource**: `nsxt_policy_service`
* **New Resource**: `nsxt_policy_security_policy`
* **New Resource**: `nsxt_policy_gateway_policy`
* **New Resource**: `nsxt_policy_segment`
* **New Resource**: `nsxt_policy_vlan_segment`
* **New Resource**: `nsxt_policy_static_route`
* **New Resource**: `nsxt_policy_nat_rule`
* **New Resource**: `nsxt_policy_vm_tags`
* **New Resource**: `nsxt_policy_ip_block`
* **New Resource**: `nsxt_policy_ip_pool`
* **New Resource**: `nsxt_policy_ip_pool_block_subnet`
* **New Resource**: `nsxt_policy_ip_pool_static_subnet`
* **New Resource**: `nsxt_policy_ip_address_allocation`
* **New Resource**: `nsxt_policy_lb_pool`
* **New Resource**: `nsxt_policy_lb_service`
* **New Resource**: `nsxt_policy_lb_virtual_server`
* **New Resource**: `nsxt_policy_bgp_neighbor`
* **New Resource**: `nsxt_policy_dhcp_relay`
* **New Resource**: `nsxt_policy_dhcp_server`

IMPROVEMENTS:
* Migrate to Terraform Plugin SDK ([#210](https://github.com/vmware/terraform-provider-nsxt/pull/210))
* `resource/nsxt_vm_tags`: Avoid backend calls if no change required in corresponding tags ([#261](https://github.com/vmware/terraform-provider-nsxt/pull/261))

BUG FIXES:
* Fix client authentication error that used to occur when client certificate is not self signed ([#207](https://github.com/vmware/terraform-provider-nsxt/pull/207))
* Allow IPv6 in IP addresses and CIDR validations ([#204](https://github.com/vmware/terraform-provider-nsxt/pull/204))
* `resource/nsxt_vm_tags`: Fix tag removal ([#240](https://github.com/vmware/terraform-provider-nsxt/pull/240))
* `resource/nsxt_vm_tags`: Apply tags to all logical ports on given vm ([#235](https://github.com/vmware/terraform-provider-nsxt/pull/235))
* `resource/nsxt_logical_dhcp_server`: Mark gateway_ip as optional rather than required ([#245](https://github.com/vmware/terraform-provider-nsxt/pull/245))


## 1.1.2 (November 18, 2019)

FEATURES:
* **New Data Source**: `nsxt_ip_pool` ([#190](https://github.com/vmware/terraform-provider-nsxt/pull/190))
* **New Resource**: `nsxt_ip_pool_allocation_ip_address` ([#190](https://github.com/vmware/terraform-provider-nsxt/pull/190))

IMPROVEMENTS:
* `resource/nsxt_ns_group`: Support IPSet type in membership criteria ([#195](https://github.com/vmware/terraform-provider-nsxt/pull/195))

BUG FIXES:
* Fix refresh failures for most of resources. When resource was deleted on backend, the provider is expected to refresh state, discover resource absence and re-create it on next apply. Instead, the provider errored out ([#195]https://github.com/vmware/terraform-provider-nsxt/pull/191))
* `resource/nsxt_ip_set`: Allow force-deletion of IPSet even if its referenced in ns groups.
* `resource/nsxt_logical_router_downlink_port`: Fix crash that happened during import with specific configuration ([#193](https://github.com/vmware/terraform-provider-nsxt/pull/193))
* `resource/nsxt_logical_router_link_port_on_tier1`: Fix crash that happened during import with specific configuration ([#193](https://github.com/vmware/terraform-provider-nsxt/pull/193))
* `resource/nsxt_*_switching_profile`: Fix update error that occured in some cases due to omitted revision ([#201](https://github.com/vmware/terraform-provider-nsxt/pull/201))
* `resource/nsxt_logical_switch`: On delete operation, detach logical switch in order to avoid possible dependency errors ([#202](https://github.com/vmware/terraform-provider-nsxt/pull/202))

## 1.1.1 (August 05, 2019)

NOTES:

* The provider is now aligned with Terraform 0.12 SDK which is required for Terraform 0.12 support. This version of terraform is more strict with syntax enforcement. If you old configuration errors out post upgrade, please verify syntax against the updated provider documentation.

IMPROVEMENTS:

* `resource/nsxt_vm_tag`: Support tagging of logical port for the VM ([#171](https://github.com/vmware/terraform-provider-nsxt/pull/171))
* `resource/nsxt_firewall_section`: Add ability to control order of FW sections ([#150](https://github.com/vmware/terraform-provider-nsxt/pull/150))
* `resource/nsxt_firewall_section`: Add support for LogicalRouter and LogicalRouterPort in as applied_to type ([#157](https://github.com/vmware/terraform-provider-nsxt/pull/157))
* Introduce flag to tolerate partial_success realization state. This can be controlled by tolerate_partial_success provider attribute or NSXT_TOLERATE_PARTIAL_SUCCESS environment variable. The default is False ([#181](https://github.com/vmware/terraform-provider-nsxt/pull/181))
* Add Go Modules support ([#155](https://github.com/vmware/terraform-provider-nsxt/pull/155))
* Fix syntax in documentation and tests according to terraform 0.12 requirements ([#178](https://github.com/vmware/terraform-provider-nsxt/pull/178))
* Verify interoperability with NSX 2.5
* Improve documentation and test coverage


BUG FIXES:

* `resource/nsxt_nat_rule`: Fix deletion of NAT rule that was due to a platform bug in versions 2.4 and below ([#166](https://github.com/vmware/terraform-provider-nsxt/pull/166)).
* `resource/nsxt_firewall_section`: Do not enforce order of services in rules. This fixes the bug of non-empty plan when services were registered on backend in order different that defined in terraform ([#156](https://github.com/vmware/terraform-provider-nsxt/pull/156))
* `resource/nsxt_firewall_section`: Prevent re-creation of rules by retaining rule ids ([#154](https://github.com/vmware/terraform-provider-nsxt/pull/154))
* `resource/nsxt_nat_rule`: Allow setting rule_priority ([#182](https://github.com/te    rraform-providers/terraform-provider-nsxt/pull/182))

## 1.1.0 (February 22, 2019)

NOTES:

* resource/nsxt_logical_switch: Attribute `vlan` is deprecated. Please use new resource `nsxt_vlan_logical_switch` to manage vlan based logical switches.

FEATURES:

* **New Data Source**: `nsxt_mac_pool`
* **New Data Source**: `nsxt_ns_group`
* **New Data Source**: `nsxt_ns_service`
* **New Data Source**: `nsxt_certificate`
* **New Resource**: `nsxt_dhcp_relay_profile`
* **New Resource**: `nsxt_dhcp_relay_service`
* **New Resource**: `nsxt_dhcp_server_profile`
* **New Resource**: `nsxt_logical_dhcp_server`
* **New Resource**: `nsxt_dhcp_server_ip_pool`
* **New Resource**: `nsxt_vlan_logical_switch`
* **New Resource**: `nsxt_logical_dhcp_port`
* **New Resource**: `nsxt_logical_tier0_router`
* **New Resource**: `nsxt_logical_router_centralized_service_port`
* **New Resource**: `nsxt_ip_block`
* **New Resource**: `nsxt_ip_block_subnet`
* **New Resource**: `nsxt_ip_pool`
* **New Resource**: `nsxt_ip_set`
* **New Resource**: `nsxt_lb_icmp_monitor`
* **New Resource**: `nsxt_lb_tcp_monitor`
* **New Resource**: `nsxt_lb_udp_monitor`
* **New Resource**: `nsxt_lb_http_monitor`
* **New Resource**: `nsxt_lb_https_monitor`
* **New Resource**: `nsxt_lb_passive_monitor`
* **New Resource**: `nsxt_lb_pool`
* **New Resource**: `nsxt_lb_tcp_virtual_server`
* **New Resource**: `nsxt_lb_udp_virtual_server`
* **New Resource**: `nsxt_lb_http_forwarding_rule`
* **New Resource**: `nsxt_lb_http_request_rewrite_rule`
* **New Resource**: `nsxt_lb_http_response_rewrite_rule`
* **New Resource**: `nsxt_lb_cookie_persistence_profile`
* **New Resource**: `nsxt_lb_source_ip_persistence_profile`
* **New Resource**: `nsxt_lb_client_ssl_profile`
* **New Resource**: `nsxt_lb_server_ssl_profile`
* **New Resource**: `nsxt_lb_service`
* **New Resource**: `nsxt_lb_fast_tcp_application_profile`
* **New Resource**: `nsxt_lb_fast_udp_application_profile`
* **New Resource**: `nsxt_lb_http_application_profile`

## 1.0.0 (April 09, 2018)

Initial release.
