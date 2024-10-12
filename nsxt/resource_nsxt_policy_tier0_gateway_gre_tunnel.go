// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyTier0GatewayGRETunnel() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyTier0GatewayGRETunnelCreate,
		Read:   resourceNsxtPolicyTier0GatewayGRETunnelRead,
		Update: resourceNsxtPolicyTier0GatewayGRETunnelUpdate,
		Delete: resourceNsxtPolicyTier0GatewayGRETunnelDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0GatewayGRETunnelImport,
		},
		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"locale_service_path": {
				Type:        schema.TypeString,
				Description: "Policy path of associated Gateway Locale Service on NSX",
				Required:    true,
				ForceNew:    true,
			},
			"destination_address": {
				Type:         schema.TypeString,
				Description:  "Destination IPv4 address",
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Enable/Disable Tunnel",
				Optional:    true,
				Default:     true,
			},
			"mtu": {
				Type:         schema.TypeInt,
				Description:  "Maximum transmission unit",
				Optional:     true,
				Default:      1476,
				ValidateFunc: validation.IntAtLeast(64),
			},
			"tunnel_address": {
				Type:        schema.TypeList,
				Description: "Tunnel Address object parameter",
				Required:    true,
				MinItems:    1,
				MaxItems:    8,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"edge_path": {
							Type:         schema.TypeString,
							Description:  "Policy edge node path",
							Required:     true,
							ValidateFunc: validatePolicyPath(),
						},
						"source_address": {
							Type:         schema.TypeString,
							Description:  "IPv4 source address",
							Required:     true,
							ValidateFunc: validateSingleIP(),
						},
						"tunnel_interface_subnet": {
							Type:        schema.TypeList,
							Description: "Interface Subnet object parameter",
							Required:    true,
							MinItems:    1,
							MaxItems:    2,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_addresses": {
										Type:        schema.TypeList,
										Description: "IP addresses assigned to interface",
										Required:    true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateSingleIP(),
										},
									},
									"prefix_len": {
										Type:         schema.TypeInt,
										Description:  "Subnet prefix length",
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 128),
									},
								},
							},
						},
					},
				},
			},
			"tunnel_keepalive": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "tunnel keep alive object",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dead_time_multiplier": {
							Type:         schema.TypeInt,
							Description:  "Dead time multiplier",
							Optional:     true,
							ValidateFunc: validation.IntBetween(3, 5),
							Default:      3,
						},
						"enable_keepalive_ack": {
							Type:        schema.TypeBool,
							Description: "Enable tunnel keep alive acknowledge",
							Optional:    true,
							Default:     true,
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enable/Disable tunnel keep alive",
							Optional:    true,
							Default:     false,
						},
						"keepalive_interval": {
							Type:         schema.TypeInt,
							Description:  "Keep alive interval",
							ValidateFunc: validation.IntBetween(2, 120),
							Optional:     true,
							Default:      10,
						},
					},
				},
			},
		},
	}
}

func tier0GatewayGRETunnelFromSchema(d *schema.ResourceData) (*data.StructValue, error) {
	converter := bindings.NewTypeConverter()

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	destinationAddress := d.Get("destination_address").(string)
	enabled := d.Get("enabled").(bool)
	mtu := int64(d.Get("mtu").(int))
	var tunnelAddresses []model.TunnelAddress
	tunnelAddressList := d.Get("tunnel_address").([]interface{})
	for _, ta := range tunnelAddressList {
		tunnelAddress := ta.(map[string]interface{})
		edgePath := tunnelAddress["edge_path"].(string)
		sourceAddress := tunnelAddress["source_address"].(string)
		var tunnelInterfaceSubnets []model.InterfaceSubnet
		tunnelInterfaceSubnet := tunnelAddress["tunnel_interface_subnet"].([]interface{})
		for _, tis := range tunnelInterfaceSubnet {
			t := tis.(map[string]interface{})
			ipAddresses := interfaceListToStringList(t["ip_addresses"].([]interface{}))
			prefixLen := int64(t["prefix_len"].(int))
			elem := model.InterfaceSubnet{
				IpAddresses: ipAddresses,
				PrefixLen:   &prefixLen,
			}
			tunnelInterfaceSubnets = append(tunnelInterfaceSubnets, elem)
		}
		elem := model.TunnelAddress{
			EdgePath:              &edgePath,
			SourceAddress:         &sourceAddress,
			TunnelInterfaceSubnet: tunnelInterfaceSubnets,
		}
		tunnelAddresses = append(tunnelAddresses, elem)
	}
	var tunnelKeepalive model.TunnelKeepAlive
	tunnelKeepAliveList := d.Get("tunnel_keepalive").([]interface{})
	for _, tka := range tunnelKeepAliveList {
		tkaElement := tka.(map[string]interface{})
		deadTimeMultiplier := int64(tkaElement["dead_time_multiplier"].(int))
		enableKeepaliveAck := tkaElement["enable_keepalive_ack"].(bool)
		tkaEnabled := tkaElement["enabled"].(bool)
		keepaliveInterval := int64(tkaElement["keepalive_interval"].(int))

		tunnelKeepalive = model.TunnelKeepAlive{
			DeadTimeMultiplier: &deadTimeMultiplier,
			EnableKeepaliveAck: &enableKeepaliveAck,
			Enabled:            &tkaEnabled,
			KeepaliveInterval:  &keepaliveInterval,
		}
		break // Only one element is allowed.
	}

	greTunnel := model.GreTunnel{
		DisplayName:        &displayName,
		Description:        &description,
		Tags:               tags,
		Enabled:            &enabled,
		Mtu:                &mtu,
		DestinationAddress: &destinationAddress,
		TunnelAddress:      tunnelAddresses,
		TunnelKeepalive:    &tunnelKeepalive,
		ResourceType:       model.GreTunnel__TYPE_IDENTIFIER,
	}
	dataValue, errs := converter.ConvertToVapi(greTunnel, model.GreTunnelBindingType())
	if errs != nil {
		return nil, errs[0]
	} else if dataValue != nil {
		return dataValue.(*data.StructValue), nil
	}
	return nil, nil

}

func resourceNsxtPolicyTier0GatewayGRETunnelExists(id, localeServicePath string, connector client.Connector, isGlobalManager bool) (bool, error) {
	isT0, tier0id, localeSvcID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil {
		return false, err
	}
	if !isT0 {
		return false, fmt.Errorf("locale service path %s is not associated with a tier0 router", localeServicePath)
	}

	client := locale_services.NewTunnelsClient(connector)
	_, err = client.Get(tier0id, localeSvcID, id)

	if isNotFoundError(err) {
		return false, nil
	}
	return true, nil
}

func resourceNsxtPolicyTier0GatewayGRETunnelCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}
	localeServicePath := d.Get("locale_service_path").(string)
	isT0, tier0id, localeSvcID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil {
		return handleCreateError("GreTunnel", id, err)
	}
	if !isT0 {
		return fmt.Errorf("locale service path %s is not associated with a tier0 router", localeServicePath)
	}
	client := locale_services.NewTunnelsClient(connector)

	obj, err := tier0GatewayGRETunnelFromSchema(d)
	if err != nil {
		return fmt.Errorf("unable to create the GRE Tunnel: %v", err)
	}
	err = client.Patch(tier0id, localeSvcID, id, obj)
	if err != nil {
		return handleCreateError("GreTunnel", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyTier0GatewayGRETunnelRead(d, m)
}

func resourceNsxtPolicyTier0GatewayGRETunnelRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining GreTunnel ID")
	}
	localeServicePath := d.Get("locale_service_path").(string)
	isT0, tier0id, localeSvcID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil {
		return handleCreateError("GreTunnel", id, err)
	}
	if !isT0 {
		return fmt.Errorf("locale service path %s is not associated with a tier0 router", localeServicePath)
	}

	client := locale_services.NewTunnelsClient(connector)
	obj, err := client.Get(tier0id, localeSvcID, id)
	if err != nil {
		return handleReadError(d, "GreTunnel", id, err)
	}

	converter := bindings.NewTypeConverter()
	gt, errs := converter.ConvertToGolang(obj, model.GreTunnelBindingType())
	if errs != nil {
		return handleReadError(d, "GreTunnel", id, errs[0])
	}
	greTunnel := gt.(model.GreTunnel)
	d.Set("nsx_id", id)
	d.Set("path", greTunnel.Path)
	d.Set("display_name", greTunnel.DisplayName)
	d.Set("description", greTunnel.Description)
	d.Set("revision", greTunnel.Revision)
	setPolicyTagsInSchema(d, greTunnel.Tags)
	d.Set("destination_address", greTunnel.DestinationAddress)
	d.Set("enabled", greTunnel.Enabled)
	d.Set("mtu", greTunnel.Mtu)

	var tunnelAddress []interface{}
	for _, ta := range greTunnel.TunnelAddress {
		elem := make(map[string]interface{})
		elem["edge_path"] = ta.EdgePath
		elem["source_address"] = ta.SourceAddress
		var tunnelInterfaceSubnet []interface{}
		for _, tis := range ta.TunnelInterfaceSubnet {
			elem := make(map[string]interface{})
			elem["ip_addresses"] = stringList2Interface(tis.IpAddresses)
			elem["prefix_len"] = tis.PrefixLen

			tunnelInterfaceSubnet = append(tunnelInterfaceSubnet, elem)
		}
		elem["tunnel_interface_subnet"] = tunnelInterfaceSubnet
		tunnelAddress = append(tunnelAddress, elem)
	}
	d.Set("tunnel_address", tunnelAddress)

	tunnelKeepalive := make(map[string]interface{})
	tunnelKeepalive["dead_time_multiplier"] = greTunnel.TunnelKeepalive.DeadTimeMultiplier
	tunnelKeepalive["enable_keepalive_ack"] = greTunnel.TunnelKeepalive.EnableKeepaliveAck
	tunnelKeepalive["enabled"] = greTunnel.TunnelKeepalive.Enabled
	tunnelKeepalive["keepalive_interval"] = greTunnel.TunnelKeepalive.KeepaliveInterval
	d.Set("tunnel_keepalive", []interface{}{tunnelKeepalive})

	return nil
}

func resourceNsxtPolicyTier0GatewayGRETunnelUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	localeServicePath := d.Get("locale_service_path").(string)
	isT0, tier0id, localeSvcID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil {
		return handleCreateError("GreTunnel", id, err)
	}
	if !isT0 {
		return fmt.Errorf("locale service path %s is not associated with a tier0 router", localeServicePath)
	}
	client := locale_services.NewTunnelsClient(connector)

	obj, err := tier0GatewayGRETunnelFromSchema(d)
	if err != nil {
		return fmt.Errorf("failed to update GRE Tunnel: %v", err)
	}
	err = client.Patch(tier0id, localeSvcID, id, obj)
	if err != nil {
		return handleUpdateError("GreTunnel", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyTier0GatewayGRETunnelRead(d, m)
}

func resourceNsxtPolicyTier0GatewayGRETunnelDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Get("nsx_id").(string)
	localeServicePath := d.Get("locale_service_path").(string)
	isT0, tier0id, localeSvcID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil {
		return handleCreateError("GreTunnel", id, err)
	}
	if !isT0 {
		return fmt.Errorf("locale service path %s is not associated with a tier0 router", localeServicePath)
	}
	client := locale_services.NewTunnelsClient(connector)

	log.Printf("[INFO] Deleting GRE Tunnel with ID %s", id)
	err = client.Delete(tier0id, localeSvcID, id)
	if err != nil {
		return handleDeleteError("GreTunnel", id, err)
	}

	return nil
}

func resourceNsxtPolicyTier0GatewayGRETunnelImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	// Path should be like /infra/tier-0s/test/locale-services/default/tunnels/test
	importID := d.Id()
	err := fmt.Errorf("expected GRE Tunnel path, got %s", importID)
	s := strings.Split(importID, "/")
	if len(s) < 8 {
		return []*schema.ResourceData{d}, err
	}

	d.SetId(s[len(s)-1])

	s = strings.Split(importID, "/tunnels/")
	if len(s) != 2 {
		return []*schema.ResourceData{d}, err
	}
	d.Set("locale_service_path", s[0])

	return []*schema.ResourceData{d}, nil
}
