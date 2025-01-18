// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var lbPoolAlgorithmValues = []string{
	model.LBPool_ALGORITHM_IP_HASH,
	model.LBPool_ALGORITHM_WEIGHTED_ROUND_ROBIN,
	model.LBPool_ALGORITHM_ROUND_ROBIN,
	model.LBPool_ALGORITHM_WEIGHTED_LEAST_CONNECTION,
	model.LBPool_ALGORITHM_LEAST_CONNECTION,
}

var lbPoolSnatTypeValues = []string{
	"AUTOMAP", "IPPOOL", "DISABLED",
}

func resourceNsxtPolicyLBPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBPoolCreate,
		Read:   resourceNsxtPolicyLBPoolRead,
		Update: resourceNsxtPolicyLBPoolUpdate,
		Delete: resourceNsxtPolicyLBPoolDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"member":       getPoolMembersSchema(),
			"member_group": getPolicyPoolMemberGroupSchema(),
			"active_monitor_paths": {
				Type:          schema.TypeList,
				Description:   "Used by the load balancer to initiate new connections to the servers to check their health. Active healthchecks are deactivated by default and can be activated using this setting",
				Elem:          getPolicyPathSchemaSimple(),
				Optional:      true,
				ConflictsWith: []string{"active_monitor_path"},
			},
			"active_monitor_path": getPolicyPathSchemaExtended(false, false, "Active healthcheck is disabled by default and can be enabled using this setting", "This attribute is deprecated, please use active_monitor_paths", []string{"active_monitor_paths"}),
			"algorithm": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(lbPoolAlgorithmValues, false),
				Optional:     true,
				Default:      model.LBPool_ALGORITHM_ROUND_ROBIN,
			},
			"min_active_members": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Default:      1,
			},
			"passive_monitor_path": getPolicyPathSchema(false, false, "Policy path for passive health monitor"),
			"tcp_multiplexing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tcp_multiplexing_number": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"snat": getPolicyPoolSnatSchema(),
		},
	}
}

func getPolicyPoolMemberGroupSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "Dynamic pool members for the loadbalancing pool. When member group is defined, members setting should not be specified",
		Optional:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"group_path": getPolicyPathSchema(true, false, "The IP list of the Group would be used as pool member IP setting"),
				"allow_ipv4": {
					Type:        schema.TypeBool,
					Description: "Use IPv4 addresses as server IPs",
					Optional:    true,
					Default:     true,
				},
				"allow_ipv6": {
					Type:        schema.TypeBool,
					Description: "Use IPv6 addresses as server IPs",
					Optional:    true,
					Default:     false,
				},
				"max_ip_list_size": {
					Type:        schema.TypeInt,
					Description: "Limits the max number of pool members to the specified value",
					Optional:    true,
				},
				"port": {
					Type:         schema.TypeString,
					Description:  "If port is specified, all connections will be sent to this port. If unset, the same port the client connected to will be used",
					Optional:     true,
					ValidateFunc: validateSinglePort(),
				},
			},
		},
	}
}

func getPolicyPoolSnatSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Description: "SNAT configuration",
		Optional:    true,
		MaxItems:    1,
		Computed:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:         schema.TypeString,
					Description:  "Type of SNAT performed to ensure reverse traffic from the server can be received and processed by the loadbalancer",
					ValidateFunc: validation.StringInSlice(lbPoolSnatTypeValues, false),
					Optional:     true,
					Default:      "AUTOMAP",
				},
				"ip_pool_addresses": {
					Type:        schema.TypeList,
					Description: "List of IP CIDRs or IP ranges for SNAT of type SNAT_IP_POOL",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateCidrOrIPOrRange(),
					},
					Optional: true,
				},
			},
		},
	}

}

func getPolicyPoolMembersFromSchema(d *schema.ResourceData) []model.LBPoolMember {
	members := d.Get("member").([]interface{})
	var memberList []model.LBPoolMember
	for _, member := range members {
		data := member.(map[string]interface{})
		displayName := data["display_name"].(string)
		adminState := data["admin_state"].(string)
		backupMember := data["backup_member"].(bool)
		port := data["port"].(string)
		weight := int64(data["weight"].(int))
		maxConnections := int64(data["max_concurrent_connections"].(int))
		address := data["ip_address"].(string)
		elem := model.LBPoolMember{
			AdminState:   &adminState,
			BackupMember: &backupMember,
			DisplayName:  &displayName,
			IpAddress:    &address,
			Weight:       &weight,
		}

		if maxConnections > 0 {
			elem.MaxConcurrentConnections = &maxConnections
		}
		if port != "" {
			elem.Port = &port
		}

		memberList = append(memberList, elem)
	}

	return memberList
}

func setPolicyPoolMembersInSchema(d *schema.ResourceData, members []model.LBPoolMember) error {
	var membersList []map[string]interface{}
	for _, member := range members {
		elem := make(map[string]interface{})
		if member.DisplayName != nil {
			elem["display_name"] = *member.DisplayName
		}
		if member.AdminState != nil {
			elem["admin_state"] = *member.AdminState
		}
		if member.BackupMember != nil {
			elem["backup_member"] = *member.BackupMember
		}
		elem["ip_address"] = member.IpAddress
		if member.MaxConcurrentConnections != nil {
			elem["max_concurrent_connections"] = *member.MaxConcurrentConnections
		}
		if member.Port != nil {
			elem["port"] = *member.Port
		}
		if member.Weight != nil {
			elem["weight"] = *member.Weight
		}

		membersList = append(membersList, elem)
	}
	err := d.Set("member", membersList)
	return err
}

func getPolicyPoolMemberGroupFromSchema(d *schema.ResourceData) *model.LBPoolMemberGroup {
	members := d.Get("member_group").([]interface{})
	for _, member := range members {
		data := member.(map[string]interface{})
		groupPath := data["group_path"].(string)
		allowIPv4 := data["allow_ipv4"].(bool)
		allowIPv6 := data["allow_ipv6"].(bool)
		port := data["port"].(string)
		maxIPListSize := int64(data["max_ip_list_size"].(int))
		ipRevisionFilter := model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV4
		if !allowIPv4 {
			ipRevisionFilter = model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV6
		} else if allowIPv6 {
			ipRevisionFilter = model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV4_IPV6
		}

		elem := model.LBPoolMemberGroup{
			GroupPath:        &groupPath,
			IpRevisionFilter: &ipRevisionFilter,
		}

		if port != "" {
			portInt, _ := strconv.Atoi(port)
			portInt64 := int64(portInt)
			elem.Port = &portInt64
		}

		if maxIPListSize > 0 {
			elem.MaxIpListSize = &maxIPListSize
		}

		// Only single element is allowed
		return &elem
	}

	return nil
}

func setPolicyPoolMemberGroupInSchema(d *schema.ResourceData, groupMember *model.LBPoolMemberGroup) error {
	var groupMembersList []map[string]interface{}
	if groupMember != nil {
		elem := make(map[string]interface{})
		elem["group_path"] = groupMember.GroupPath
		if groupMember.IpRevisionFilter != nil {
			allowIPv4 := (*groupMember.IpRevisionFilter == model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV4) || (*groupMember.IpRevisionFilter == model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV4_IPV6)
			allowIPv6 := (*groupMember.IpRevisionFilter == model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV6) || (*groupMember.IpRevisionFilter == model.LBPoolMemberGroup_IP_REVISION_FILTER_IPV4_IPV6)
			elem["allow_ipv4"] = allowIPv4
			elem["allow_ipv6"] = allowIPv6
		}
		if groupMember.MaxIpListSize != nil {
			elem["max_ip_list_size"] = *groupMember.MaxIpListSize
		}
		if groupMember.Port != nil {
			elem["port"] = fmt.Sprintf("%d", *groupMember.Port)
		}

		groupMembersList = append(groupMembersList, elem)
	}
	err := d.Set("member_group", groupMembersList)
	return err
}

func getPolicyPoolSnatFromSchema(d *schema.ResourceData) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	snats := d.Get("snat").([]interface{})
	for _, snat := range snats {
		snatMap := snat.(map[string]interface{})
		snatType := snatMap["type"].(string)

		if snatType == "DISABLED" {
			entry := model.LBSnatDisabled{
				Type_: model.LBSnatDisabled__TYPE_IDENTIFIER,
			}

			dataValue, errs := converter.ConvertToVapi(entry, model.LBSnatDisabledBindingType())
			if errs != nil {
				return nil, errs[0]
			}

			return dataValue.(*data.StructValue), nil
		}

		if snatType == "AUTOMAP" {

			entry := model.LBSnatAutoMap{
				Type_: model.LBSnatAutoMap__TYPE_IDENTIFIER,
			}

			dataValue, errs := converter.ConvertToVapi(entry, model.LBSnatAutoMapBindingType())
			if errs != nil {
				return nil, errs[0]
			}

			return dataValue.(*data.StructValue), nil
		}

		addresses := snatMap["ip_pool_addresses"].([]interface{})
		var addressList []model.LBSnatIpElement
		for _, address := range addresses {
			addressStr := address.(string)
			tokens := strings.Split(addressStr, "/")
			if len(tokens) == 2 {
				// cidr format
				prefix, err := strconv.Atoi(tokens[1])
				if err != nil {
					return nil, fmt.Errorf("Failed to convert snat address prefix: %s", err)
				}
				prefix64 := int64(prefix)
				element := model.LBSnatIpElement{
					IpAddress:    &tokens[0],
					PrefixLength: &prefix64,
				}

				addressList = append(addressList, element)
			} else {
				// single IP or range
				element := model.LBSnatIpElement{
					IpAddress: &addressStr,
				}
				addressList = append(addressList, element)
			}
		}

		// IP POOL type
		entry := model.LBSnatIpPool{
			Type_:       model.LBSnatIpPool__TYPE_IDENTIFIER,
			IpAddresses: addressList,
		}

		dataValue, errs := converter.ConvertToVapi(entry, model.LBSnatIpPoolBindingType())
		if errs != nil {
			return nil, errs[0]
		}

		return dataValue.(*data.StructValue), nil
	}

	return nil, nil
}

func setPolicyPoolSnatInSchema(d *schema.ResourceData, snat *data.StructValue) error {
	if snat == nil {
		return nil
	}

	converter := bindings.NewTypeConverter()
	var snatList []map[string]interface{}
	elem := make(map[string]interface{})

	basicType, errs := converter.ConvertToGolang(snat, model.LBSnatTranslationBindingType())
	if errs != nil {
		return errs[0]
	}

	snatType := basicType.(model.LBSnatTranslation).Type_

	if snatType == model.LBSnatDisabled__TYPE_IDENTIFIER {
		elem["type"] = "DISABLED"
		elem["ip_pool_addresses"] = nil
	} else if snatType == model.LBSnatAutoMap__TYPE_IDENTIFIER {
		elem["type"] = "AUTOMAP"
		elem["ip_pool_addresses"] = nil
	} else {
		if snatType != model.LBSnatIpPool__TYPE_IDENTIFIER {
			return fmt.Errorf("Unrecognized Snat Translation type %s", snatType)
		}
		// IP Pool
		data, errs := converter.ConvertToGolang(snat, model.LBSnatIpPoolBindingType())
		if errs != nil {
			return errs[0]
		}
		ipPoolSnat := data.(model.LBSnatIpPool)
		elem["type"] = "IPPOOL"
		var addressList []string
		for _, address := range ipPoolSnat.IpAddresses {
			if address.PrefixLength != nil {
				cidr := fmt.Sprintf("%s/%d", *address.IpAddress, *address.PrefixLength)
				addressList = append(addressList, cidr)
			} else {
				addressList = append(addressList, *address.IpAddress)
			}
		}
		elem["ip_pool_addresses"] = addressList
	}

	snatList = append(snatList, elem)
	d.Set("snat", snatList)
	return nil
}

func resourceNsxtPolicyLBPoolExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewLbPoolsClient(connector)

	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyLBPoolCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbPoolsClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBPoolExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	activeMonitorPaths := interfaceListToStringList(d.Get("active_monitor_paths").([]interface{}))
	if activeMonitorPaths == nil && d.Get("active_monitor_path") != "" {
		activeMonitorPath := d.Get("active_monitor_path").(string)
		activeMonitorPaths = []string{activeMonitorPath}
	}
	algorithm := d.Get("algorithm").(string)
	memberGroup := getPolicyPoolMemberGroupFromSchema(d)
	members := getPolicyPoolMembersFromSchema(d)
	minActiveMembers := int64(d.Get("min_active_members").(int))
	passiveMonitorPath := d.Get("passive_monitor_path").(string)
	snatTranslation, err := getPolicyPoolSnatFromSchema(d)
	if err != nil {
		return err
	}
	tcpMultiplexingEnabled := d.Get("tcp_multiplexing_enabled").(bool)
	tcpMultiplexingNumber := int64(d.Get("tcp_multiplexing_number").(int))

	obj := model.LBPool{
		DisplayName:            &displayName,
		Description:            &description,
		Tags:                   tags,
		ActiveMonitorPaths:     activeMonitorPaths,
		Algorithm:              &algorithm,
		MemberGroup:            memberGroup,
		Members:                members,
		PassiveMonitorPath:     &passiveMonitorPath,
		MinActiveMembers:       &minActiveMembers,
		SnatTranslation:        snatTranslation,
		TcpMultiplexingEnabled: &tcpMultiplexingEnabled,
		TcpMultiplexingNumber:  &tcpMultiplexingNumber,
	}

	log.Printf("[INFO] Creating LBPool with ID %s", id)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("LBPool", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBPoolRead(d, m)
}

func resourceNsxtPolicyLBPoolRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbPoolsClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBPool ID")
	}

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBPool", id, err)
	}

	if _, ok := d.GetOk("active_monitor_path"); ok {
		// deprecated value was set
		if obj.ActiveMonitorPaths != nil {
			d.Set("active_monitor_path", obj.ActiveMonitorPaths[0])
		}
	} else {
		// resource uses lists attribute (non-deprecated) or attribute
		// not specified. In this case set non-deprecated attr in state
		d.Set("active_monitor_paths", obj.ActiveMonitorPaths)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("algorithm", obj.Algorithm)
	err = setPolicyPoolMemberGroupInSchema(d, obj.MemberGroup)
	if err != nil {
		return err
	}
	err = setPolicyPoolMembersInSchema(d, obj.Members)
	if err != nil {
		return err
	}
	err = setPolicyPoolMemberGroupInSchema(d, obj.MemberGroup)
	if err != nil {
		return err
	}

	err = setPolicyPoolSnatInSchema(d, obj.SnatTranslation)
	if err != nil {
		return err
	}

	if obj.MinActiveMembers != nil {
		d.Set("min_active_members", obj.MinActiveMembers)
	} else {
		d.Set("min_active_members", 1)
	}
	d.Set("passive_monitor_path", obj.PassiveMonitorPath)
	d.Set("tcp_multiplexing_enabled", obj.TcpMultiplexingEnabled)
	d.Set("tcp_multiplexing_number", obj.TcpMultiplexingNumber)

	return nil
}

func resourceNsxtPolicyLBPoolUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbPoolsClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBPool ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	activeMonitorPaths := interfaceListToStringList(d.Get("active_monitor_paths").([]interface{}))
	if activeMonitorPaths == nil && d.Get("active_monitor_path") != "" {
		activeMonitorPath := d.Get("active_monitor_path").(string)
		activeMonitorPaths = []string{activeMonitorPath}
	}
	algorithm := d.Get("algorithm").(string)
	memberGroup := getPolicyPoolMemberGroupFromSchema(d)
	members := getPolicyPoolMembersFromSchema(d)
	passiveMonitorPath := d.Get("passive_monitor_path").(string)
	snatTranslation, err := getPolicyPoolSnatFromSchema(d)
	if err != nil {
		return err
	}
	tcpMultiplexingEnabled := d.Get("tcp_multiplexing_enabled").(bool)
	tcpMultiplexingNumber := int64(d.Get("tcp_multiplexing_number").(int))
	minActiveMembers := int64(d.Get("min_active_members").(int))
	revision := int64(d.Get("revision").(int))

	obj := model.LBPool{
		DisplayName:            &displayName,
		Description:            &description,
		Tags:                   tags,
		ActiveMonitorPaths:     activeMonitorPaths,
		Algorithm:              &algorithm,
		MemberGroup:            memberGroup,
		Members:                members,
		MinActiveMembers:       &minActiveMembers,
		PassiveMonitorPath:     &passiveMonitorPath,
		SnatTranslation:        snatTranslation,
		TcpMultiplexingEnabled: &tcpMultiplexingEnabled,
		TcpMultiplexingNumber:  &tcpMultiplexingNumber,
		Revision:               &revision,
	}

	_, err = client.Update(id, obj)
	if err != nil {
		return handleUpdateError("LBPool", id, err)
	}

	return resourceNsxtPolicyLBPoolRead(d, m)
}

func resourceNsxtPolicyLBPoolDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBPool ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewLbPoolsClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	force := true
	err := client.Delete(id, &force)
	if err != nil {
		return handleDeleteError("LBPool", id, err)
	}

	return nil
}
