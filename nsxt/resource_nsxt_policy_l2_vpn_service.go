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
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	t0_locale_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s"
	t1_locale_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/locale_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var L2VpnServiceModes = []string{
	model.L2VPNService_MODE_SERVER,
	model.L2VPNService_MODE_CLIENT,
}

func resourceNsxtPolicyL2VpnService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyL2VpnServiceCreate,
		Read:   resourceNsxtPolicyL2VpnServiceRead,
		Update: resourceNsxtPolicyL2VpnServiceUpdate,
		Delete: resourceNsxtPolicyL2VpnServiceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyL2VpnServiceImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"gateway_path": getPolicyPathSchema(false, true, "Policy path for the gateway."),
			"locale_service_path": {
				Type:         schema.TypeString,
				Description:  "Polciy path for the locale service.",
				Optional:     true,
				ForceNew:     true,
				Deprecated:   "Use gateway_path instead.",
				ValidateFunc: validatePolicyPath(),
			},
			"enable_hub": {
				Type:        schema.TypeBool,
				Description: "This property applies only in SERVER mode. If set to true, traffic from any client will be replicated to all other clients. If set to false, traffic received from clients is only replicated to the local VPN endpoint.",
				Optional:    true,
				Default:     true,
			},
			"encap_ip_pool": {
				Type:        schema.TypeList,
				Description: "IP Pool to allocate local and peer endpoint IPs.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validateIPCidr(),
				},
				Optional: true,
			},
			"mode": {
				Type:         schema.TypeString,
				Description:  "Specify an L2VPN service mode as SERVER or CLIENT.",
				ValidateFunc: validation.StringInSlice(L2VpnServiceModes, false),
				Optional:     true,
				Default:      model.L2VPNService_MODE_SERVER,
			},
		},
	}
}

func getNsxtPolicyL2VpnServiceByID(connector client.Connector, gwID string, isT0 bool, localeServiceID string, serviceID string, isGlobalManager bool) (model.L2VPNService, error) {
	if localeServiceID == "" {
		if isT0 {
			client := tier_0s.NewL2vpnServicesClient(connector)
			return client.Get(gwID, serviceID)
		}
		client := tier_1s.NewL2vpnServicesClient(connector)
		return client.Get(gwID, serviceID)
	}
	if isT0 {
		client := t0_locale_service.NewL2vpnServicesClient(connector)
		return client.Get(gwID, localeServiceID, serviceID)
	}
	client := t1_locale_service.NewL2vpnServicesClient(connector)
	return client.Get(gwID, localeServiceID, serviceID)
}

func patchNsxtPolicyL2VpnService(connector client.Connector, gwID string, localeServiceID string, l2VpnService model.L2VPNService, isT0 bool) error {
	id := *l2VpnService.Id
	if localeServiceID == "" {
		if isT0 {
			client := tier_0s.NewL2vpnServicesClient(connector)
			return client.Patch(gwID, id, l2VpnService)
		}
		client := tier_1s.NewL2vpnServicesClient(connector)
		return client.Patch(gwID, id, l2VpnService)
	}
	if isT0 {
		client := t0_locale_service.NewL2vpnServicesClient(connector)
		return client.Patch(gwID, localeServiceID, id, l2VpnService)
	}
	client := t1_locale_service.NewL2vpnServicesClient(connector)
	return client.Patch(gwID, localeServiceID, id, l2VpnService)

}

func deleteNsxtPolicyL2VpnService(connector client.Connector, gwID string, localeServiceID string, isT0 bool, id string) error {
	if localeServiceID == "" {
		if isT0 {
			client := tier_0s.NewL2vpnServicesClient(connector)
			return client.Delete(gwID, id)
		}
		client := tier_1s.NewL2vpnServicesClient(connector)
		return client.Delete(gwID, id)
	}
	if isT0 {
		client := t0_locale_service.NewL2vpnServicesClient(connector)
		return client.Delete(gwID, localeServiceID, id)
	}
	client := t1_locale_service.NewL2vpnServicesClient(connector)
	return client.Delete(gwID, localeServiceID, id)
}

func resourceNsxtPolicyL2VpnServiceImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	err := fmt.Errorf("Expected policy path for the L2 VPN Service, got %s", importID)
	// The policy path of L2 VPN Service should be like /infra/tier-0s/aaa/locale-services/bbb/l2vpn-services/ccc
	// or /infra/tier-0s/aaa/l2vpn-services/bbb
	if len(s) != 8 && len(s) != 6 {
		return nil, err
	}
	useLocaleService := (len(s) == 8)
	d.SetId(s[len(s)-1])
	s = strings.Split(importID, "/l2vpn-services/")
	if len(s) != 2 {
		return []*schema.ResourceData{d}, err
	}
	if useLocaleService {
		d.Set("locale_service_path", s[0])
	} else {
		d.Set("gateway_path", s[0])
	}
	return []*schema.ResourceData{d}, nil
}

func resourceNsxtPolicyL2VpnServiceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2VpnService ID")
	}
	gatewayPath, localeServicePath, err := getLocaleServiceAndGatewayPath(d)
	if err != nil {
		return nil
	}
	isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil && gatewayPath == "" {
		return err
	}
	if localeServiceID == "" {
		isT0, gwID = parseGatewayPolicyPath(gatewayPath)
	}
	obj, err := getNsxtPolicyL2VpnServiceByID(connector, gwID, isT0, localeServiceID, id, isPolicyGlobalManager(m))
	if err != nil {
		return handleReadError(d, "L2VpnService", id, err)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("enable_hub", obj.EnableHub)
	d.Set("mode", obj.Mode)

	if obj.EncapIpPool != nil {
		d.Set("encap_ip_pool", obj.EncapIpPool)
	}
	return nil
}

func resourceNsxtPolicyL2VpnServiceCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	gatewayPath, localeServicePath, err := getLocaleServiceAndGatewayPath(d)
	if err != nil {
		return nil
	}
	isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil && gatewayPath == "" {
		return err
	}
	if localeServiceID == "" {
		isT0, gwID = parseGatewayPolicyPath(gatewayPath)
	}
	isGlobalManager := isPolicyGlobalManager(m)
	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := getNsxtPolicyL2VpnServiceByID(connector, gwID, isT0, localeServiceID, id, isGlobalManager)
		if err == nil {
			return fmt.Errorf("L2VpnService with nsx_id '%s' already exists", id)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	enableHub := d.Get("enable_hub").(bool)
	mode := d.Get("mode").(string)
	tags := getPolicyTagsFromSchema(d)

	l2VpnService := model.L2VPNService{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		EnableHub:   &enableHub,
		Mode:        &mode,
	}

	ipPool := d.Get("encap_ip_pool").([]interface{})
	encapIPPool := make([]string, 0, len(ipPool))
	for _, s := range ipPool {
		encapIPPool = append(encapIPPool, s.(string))
	}
	if len(encapIPPool) != 0 {
		l2VpnService.EncapIpPool = encapIPPool
	}

	err = patchNsxtPolicyL2VpnService(connector, gwID, localeServiceID, l2VpnService, isT0)
	if err != nil {
		return handleCreateError("L2VpnService", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyL2VpnServiceRead(d, m)
}

func resourceNsxtPolicyL2VpnServiceUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2 VPN Service ID")
	}
	gatewayPath, localeServicePath, err := getLocaleServiceAndGatewayPath(d)
	if err != nil {
		return nil
	}
	isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil && gatewayPath == "" {
		return err
	}
	if localeServiceID == "" {
		isT0, gwID = parseGatewayPolicyPath(gatewayPath)
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	enableHub := d.Get("enable_hub").(bool)
	revision := int64(d.Get("revision").(int))
	mode := d.Get("mode").(string)
	tags := getPolicyTagsFromSchema(d)

	l2VpnService := model.L2VPNService{
		Id:          &id,
		DisplayName: &displayName,
		Description: &description,
		Tags:        tags,
		EnableHub:   &enableHub,
		Mode:        &mode,
		Revision:    &revision,
	}

	ipPool := d.Get("encap_ip_pool").([]interface{})
	encapIPPool := make([]string, 0, len(ipPool))
	for _, s := range ipPool {
		encapIPPool = append(encapIPPool, s.(string))
	}
	if len(encapIPPool) != 0 {
		l2VpnService.EncapIpPool = encapIPPool
	}

	log.Printf("[INFO] Updating L2VpnService with ID %s", id)
	err = patchNsxtPolicyL2VpnService(connector, gwID, localeServiceID, l2VpnService, isT0)
	if err != nil {
		return handleUpdateError("L2VpnService", id, err)
	}
	d.SetId(id)
	d.Set("nsx_id", id)
	return resourceNsxtPolicyL2VpnServiceRead(d, m)
}

func resourceNsxtPolicyL2VpnServiceDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining L2 VPN Service ID")
	}

	gatewayPath, localeServicePath, err := getLocaleServiceAndGatewayPath(d)
	if err != nil {
		return nil
	}
	isT0, gwID, localeServiceID, err := parseLocaleServicePolicyPath(localeServicePath)
	if err != nil && gatewayPath == "" {
		return err
	}
	if localeServiceID == "" {
		isT0, gwID = parseGatewayPolicyPath(gatewayPath)
	}

	err = deleteNsxtPolicyL2VpnService(getPolicyConnector(m), gwID, localeServiceID, isT0, id)
	if err != nil {
		return handleDeleteError("L2VpnService", id, err)
	}
	return nil
}
