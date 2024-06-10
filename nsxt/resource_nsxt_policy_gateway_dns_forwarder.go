/* Copyright © 2020 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	tier0s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_0s"
	tier1s "github.com/vmware/terraform-provider-nsxt/api/infra/tier_1s"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var gatewayDNSForwarderLogLevelTypeValues = []string{
	model.PolicyDnsForwarder_LOG_LEVEL_DEBUG,
	model.PolicyDnsForwarder_LOG_LEVEL_INFO,
	model.PolicyDnsForwarder_LOG_LEVEL_WARNING,
	model.PolicyDnsForwarder_LOG_LEVEL_ERROR,
	model.PolicyDnsForwarder_LOG_LEVEL_FATAL,
}

func resourceNsxtPolicyGatewayDNSForwarder() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyGatewayDNSForwarderCreate,
		Read:   resourceNsxtPolicyGatewayDNSForwarderRead,
		Update: resourceNsxtPolicyGatewayDNSForwarderUpdate,
		Delete: resourceNsxtPolicyGatewayDNSForwarderDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyGatewayDNSForwarderImport,
		},

		Schema: map[string]*schema.Schema{
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"gateway_path": getPolicyPathSchema(true, true, "Policy path for the Gateway"),
			"listener_ip": {
				Type:         schema.TypeString,
				Description:  "IP on which the DNS Forwarder listens",
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"default_forwarder_zone_path": getPolicyPathSchema(true, false, "Zone to which DNS requests are forwarded by default"),
			"conditional_forwarder_zone_paths": {
				Type:        schema.TypeSet,
				Description: "List of conditional (FQDN) forwarder zone paths",
				Optional:    true,
				MaxItems:    5,
				Elem:        getElemPolicyPathSchema(),
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"log_level": {
				Type:         schema.TypeString,
				Description:  "Log level",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(gatewayDNSForwarderLogLevelTypeValues, false),
				Default:      model.PolicyDnsForwarder_LOG_LEVEL_INFO,
			},
			"cache_size": {
				Type:        schema.TypeInt,
				Description: "Cache size in KB",
				Optional:    true,
				Default:     1024,
			},
		},
	}
}

func policyGatewayDNSForwarderGet(sessionContext utl.SessionContext, connector client.Connector, gwID string, isT0 bool) (model.PolicyDnsForwarder, error) {
	if isT0 {
		client := tier0s.NewDnsForwarderClient(sessionContext, connector)
		return client.Get(gwID)
	}
	client := tier1s.NewDnsForwarderClient(sessionContext, connector)
	return client.Get(gwID)
}

func resourceNsxtPolicyGatewayDNSForwarderRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	obj, err := policyGatewayDNSForwarderGet(context, connector, gwID, isT0)

	if err != nil {
		return handleReadError(d, "Gateway Dns Forwarder", gwID, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("listener_ip", obj.ListenerIp)
	d.Set("default_forwarder_zone_path", obj.DefaultForwarderZonePath)
	d.Set("conditional_forwarder_zone_paths", obj.ConditionalForwarderZonePaths)
	d.Set("enabled", obj.Enabled)
	d.Set("log_level", obj.LogLevel)
	d.Set("cache_size", obj.CacheSize)

	return nil
}

func patchNsxtPolicyGatewayDNSForwarder(sessionContext utl.SessionContext, connector client.Connector, d *schema.ResourceData, gwID string, isT0 bool) error {

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	listenerIP := d.Get("listener_ip").(string)
	defaultZonePath := d.Get("default_forwarder_zone_path").(string)
	conditionalZonePaths := getStringListFromSchemaSet(d, "conditional_forwarder_zone_paths")
	enabled := d.Get("enabled").(bool)
	logLevel := d.Get("log_level").(string)
	cacheSize := int64(d.Get("cache_size").(int))

	obj := model.PolicyDnsForwarder{
		DisplayName:              &displayName,
		Description:              &description,
		Tags:                     tags,
		ListenerIp:               &listenerIP,
		DefaultForwarderZonePath: &defaultZonePath,
		Enabled:                  &enabled,
		LogLevel:                 &logLevel,
	}

	if len(conditionalZonePaths) > 0 {
		obj.ConditionalForwarderZonePaths = conditionalZonePaths
	}

	if util.NsxVersionHigherOrEqual("3.2.0") {
		obj.CacheSize = &cacheSize
	}

	if isT0 {
		client := tier0s.NewDnsForwarderClient(sessionContext, connector)
		return client.Patch(gwID, obj)
	}
	client := tier1s.NewDnsForwarderClient(sessionContext, connector)
	return client.Patch(gwID, obj)
}

func resourceNsxtPolicyGatewayDNSForwarderCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	// Verify DNS forwarder is not yet defined for this Gateway
	var err error
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}

	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	if isT0 {
		client := tier0s.NewDnsForwarderClient(context, connector)
		_, err = client.Get(gwID)
	} else {
		client := tier1s.NewDnsForwarderClient(context, connector)
		_, err = client.Get(gwID)
	}
	if err == nil {
		return fmt.Errorf("Gateway Dns Forwarder already exists for Gateway '%s'", gwID)
	} else if !isNotFoundError(err) {
		return err
	}

	log.Printf("[INFO] Creating Dns Forwarder for Gateway %s", gwID)

	err = patchNsxtPolicyGatewayDNSForwarder(context, connector, d, gwID, isT0)
	if err != nil {
		return handleCreateError("Gateway Dns Forwarder", gwID, err)
	}

	d.SetId(gwID)

	return resourceNsxtPolicyGatewayDNSForwarderRead(d, m)
}

func resourceNsxtPolicyGatewayDNSForwarderUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}
	log.Printf("[INFO] Updating Gateway Dns Forwarder with ID %s", gwID)
	err = patchNsxtPolicyGatewayDNSForwarder(context, connector, d, gwID, isT0)
	if err != nil {
		return handleUpdateError("Gateway Dns Forwarder", gwID, err)
	}

	return resourceNsxtPolicyGatewayDNSForwarderRead(d, m)
}

func resourceNsxtPolicyGatewayDNSForwarderDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}

	var err error
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	if isT0 && context.ClientType == utl.Multitenancy {
		return handleMultitenancyTier0Error()
	}

	if isT0 {
		client := tier0s.NewDnsForwarderClient(context, connector)
		err = client.Delete(gwID)
	} else {
		client := tier1s.NewDnsForwarderClient(context, connector)
		err = client.Delete(gwID)
	}
	if err != nil {
		return handleDeleteError("Gateway Dns Forwarder", gwID, err)
	}

	return nil
}

func resourceNsxtPolicyGatewayDNSForwarderImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	gwPath := d.Id()

	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}

	d.Set("gateway_path", gwPath)
	_, gwID := parseGatewayPolicyPath(gwPath)
	d.SetId(gwID)

	return []*schema.ResourceData{d}, nil
}
