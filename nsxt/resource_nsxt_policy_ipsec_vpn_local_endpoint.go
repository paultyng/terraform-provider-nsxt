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
	t0_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/ipsec_vpn_services"
	t0_nested_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services/ipsec_vpn_services"
	t1_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/ipsec_vpn_services"
	t1_nested_service "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_1s/locale_services/ipsec_vpn_services"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyIPSecVpnLocalEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPSecVpnLocalEndpointCreate,
		Read:   resourceNsxtPolicyIPSecVpnLocalEndpointRead,
		Update: resourceNsxtPolicyIPSecVpnLocalEndpointUpdate,
		Delete: resourceNsxtPolicyIPSecVpnLocalEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtVPNServiceResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"service_path": getPolicyPathSchema(true, true, "Policy path for IPSec VPN service"),
			"local_address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"certificate_path": getPolicyPathSchema(false, false, "Policy path referencing site certificate"),
			"local_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"trust_ca_paths": {
				Type:     schema.TypeSet,
				Elem:     getElemPolicyPathSchema(),
				Optional: true,
			},
			"trust_crl_paths": {
				Type:     schema.TypeSet,
				Elem:     getElemPolicyPathSchema(),
				Optional: true,
			},
		},
	}
}

type localEndpointClient struct {
	isT0            bool
	gwID            string
	localeServiceID string
	serviceID       string
}

func newLocalEndpointClient(servicePath string) (*localEndpointClient, error) {
	isT0, gwID, localeServiceID, serviceID, err := parseIPSecVPNServicePolicyPath(servicePath)
	if err != nil {
		return nil, err
	}

	return &localEndpointClient{
		isT0:            isT0,
		gwID:            gwID,
		localeServiceID: localeServiceID,
		serviceID:       serviceID,
	}, nil
}

func (c *localEndpointClient) Get(connector client.Connector, id string) (model.IPSecVpnLocalEndpoint, error) {
	if c.isT0 {
		if len(c.localeServiceID) > 0 {
			client := t0_nested_service.NewLocalEndpointsClient(connector)
			return client.Get(c.gwID, c.localeServiceID, c.serviceID, id)
		}
		client := t0_service.NewLocalEndpointsClient(connector)
		return client.Get(c.gwID, c.serviceID, id)

	}
	if len(c.localeServiceID) > 0 {
		client := t1_nested_service.NewLocalEndpointsClient(connector)
		return client.Get(c.gwID, c.localeServiceID, c.serviceID, id)
	}
	client := t1_service.NewLocalEndpointsClient(connector)
	return client.Get(c.gwID, c.serviceID, id)
}

// Note: we don't expect pagination to be relevant here
func (c *localEndpointClient) List(connector client.Connector) ([]model.IPSecVpnLocalEndpoint, error) {
	boolFalse := false
	var cursor string
	var result model.IPSecVpnLocalEndpointListResult
	var err error
	if c.isT0 {
		if len(c.localeServiceID) > 0 {
			client := t0_nested_service.NewLocalEndpointsClient(connector)
			result, err = client.List(c.gwID, c.localeServiceID, c.serviceID, &cursor, &boolFalse, nil, nil, nil, nil)
		} else {
			client := t0_service.NewLocalEndpointsClient(connector)
			result, err = client.List(c.gwID, c.serviceID, &cursor, &boolFalse, nil, nil, nil, nil)
		}

	} else {
		if len(c.localeServiceID) > 0 {
			client := t1_nested_service.NewLocalEndpointsClient(connector)
			result, err = client.List(c.gwID, c.localeServiceID, c.serviceID, &cursor, &boolFalse, nil, nil, nil, nil)
		} else {
			client := t1_service.NewLocalEndpointsClient(connector)
			result, err = client.List(c.gwID, c.serviceID, &cursor, &boolFalse, nil, nil, nil, nil)
		}
	}

	return result.Results, err
}

func (c *localEndpointClient) Patch(connector client.Connector, id string, obj model.IPSecVpnLocalEndpoint) error {
	if c.isT0 {
		if len(c.localeServiceID) > 0 {
			client := t0_nested_service.NewLocalEndpointsClient(connector)
			return client.Patch(c.gwID, c.localeServiceID, c.serviceID, id, obj)
		}
		client := t0_service.NewLocalEndpointsClient(connector)
		return client.Patch(c.gwID, c.serviceID, id, obj)

	}
	if len(c.localeServiceID) > 0 {
		client := t1_nested_service.NewLocalEndpointsClient(connector)
		return client.Patch(c.gwID, c.localeServiceID, c.serviceID, id, obj)
	}
	client := t1_service.NewLocalEndpointsClient(connector)
	return client.Patch(c.gwID, c.serviceID, id, obj)
}

func (c *localEndpointClient) Delete(connector client.Connector, id string) error {
	if c.isT0 {
		if len(c.localeServiceID) > 0 {
			client := t0_nested_service.NewLocalEndpointsClient(connector)
			return client.Delete(c.gwID, c.localeServiceID, c.serviceID, id)
		}
		client := t0_service.NewLocalEndpointsClient(connector)
		return client.Delete(c.gwID, c.serviceID, id)

	}
	if len(c.localeServiceID) > 0 {
		client := t1_nested_service.NewLocalEndpointsClient(connector)
		return client.Delete(c.gwID, c.localeServiceID, c.serviceID, id)
	}
	client := t1_service.NewLocalEndpointsClient(connector)
	return client.Delete(c.gwID, c.serviceID, id)
}

func resourceNsxtPolicyIPSecVpnLocalEndpointExistsOnService(id string, connector client.Connector, servicePath string) (bool, error) {
	client, err := newLocalEndpointClient(servicePath)
	if err != nil {
		return false, err
	}
	_, err = client.Get(connector, id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPSecVpnLocalEndpointExists(servicePath string) func(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	return func(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
		return resourceNsxtPolicyIPSecVpnLocalEndpointExistsOnService(id, connector, servicePath)
	}
}

func ipSecVpnLocalEndpointInitStruct(d *schema.ResourceData) model.IPSecVpnLocalEndpoint {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	certificatePath := d.Get("certificate_path").(string)
	localAddress := d.Get("local_address").(string)
	localID := d.Get("local_id").(string)
	trustCaPaths := getStringListFromSchemaSet(d, "trust_ca_paths")
	trustCrlPaths := getStringListFromSchemaSet(d, "trust_crl_paths")

	obj := model.IPSecVpnLocalEndpoint{
		DisplayName:  &displayName,
		Description:  &description,
		Tags:         tags,
		LocalAddress: &localAddress,
	}

	if len(localID) > 0 {
		obj.LocalId = &localID
	}

	if len(certificatePath) > 0 {
		obj.CertificatePath = &certificatePath
	}

	if len(trustCaPaths) > 0 {
		obj.TrustCaPaths = trustCaPaths
	}

	if len(trustCrlPaths) > 0 {
		obj.TrustCrlPaths = trustCrlPaths
	}

	return obj
}

func resourceNsxtPolicyIPSecVpnLocalEndpointCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	servicePath := d.Get("service_path").(string)
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyIPSecVpnLocalEndpointExists(servicePath))
	if err != nil {
		return err
	}

	obj := ipSecVpnLocalEndpointInitStruct(d)

	log.Printf("[INFO] Creating IPSecVpnLocalEndpoint with ID %s", id)
	client, err := newLocalEndpointClient(servicePath)
	if err != nil {
		return handleCreateError("IPSecVpnLocalEndpoint", id, err)
	}
	err = client.Patch(connector, id, obj)
	if err != nil {
		return handleCreateError("IPSecVpnLocalEndpoint", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPSecVpnLocalEndpointRead(d, m)
}

func resourceNsxtPolicyIPSecVpnLocalEndpointRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnLocalEndpoint ID")
	}

	servicePath := d.Get("service_path").(string)
	client, err := newLocalEndpointClient(servicePath)
	if err != nil {
		return handleReadError(d, "IPSecVpnLocalEndpoint", id, err)
	}
	obj, err := client.Get(connector, id)
	if err != nil {
		return handleReadError(d, "IPSecVpnLocalEndpoint", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("certificate_path", obj.CertificatePath)
	d.Set("local_address", obj.LocalAddress)
	d.Set("local_id", obj.LocalId)
	d.Set("trust_ca_paths", obj.TrustCaPaths)
	d.Set("trust_crl_paths", obj.TrustCrlPaths)

	return nil
}

func resourceNsxtPolicyIPSecVpnLocalEndpointUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnLocalEndpoint ID")
	}

	servicePath := d.Get("service_path").(string)

	client, err := newLocalEndpointClient(servicePath)
	if err != nil {
		return handleUpdateError("IPSecVpnLocalEndpoint", id, err)
	}

	obj := ipSecVpnLocalEndpointInitStruct(d)
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision

	log.Printf("[INFO] Updating IPSecVpnLocalEndpoint with ID %s", id)
	err = client.Patch(connector, id, obj)
	if err != nil {
		return handleUpdateError("IPSecVpnLocalEndpoint", id, err)
	}

	return resourceNsxtPolicyIPSecVpnLocalEndpointRead(d, m)
}

func resourceNsxtPolicyIPSecVpnLocalEndpointDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnLocalEndpoint ID")
	}

	servicePath := d.Get("service_path").(string)
	client, err := newLocalEndpointClient(servicePath)
	if err != nil {
		return handleUpdateError("IPSecVpnLocalEndpoint", id, err)
	}
	connector := getPolicyConnector(m)
	err = client.Delete(connector, id)

	if err != nil {
		return handleDeleteError("IPSecVpnLocalEndpoint", id, err)
	}

	return nil
}

func nsxtVPNServiceResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	err := fmt.Errorf("VPN Local Endpoint Path expected, got %s", importID)
	// path format example: /infra/tier-1s/aaa/locale-services/default/ipsec-vpn-services/bbb/local-endpoints/ccc
	s := strings.Split(importID, "/")
	if len(s) < 8 {
		return []*schema.ResourceData{d}, err
	}

	endpointID := s[len(s)-1]
	d.SetId(endpointID)
	s = strings.Split(importID, "/local-endpoints/")
	if len(s) != 2 {
		return []*schema.ResourceData{d}, err
	}

	d.Set("service_path", s[0])

	return []*schema.ResourceData{d}, nil
}
