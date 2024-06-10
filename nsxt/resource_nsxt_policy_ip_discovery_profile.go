/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

func resourceNsxtPolicyIPDiscoveryProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPDiscoveryProfileCreate,
		Read:   resourceNsxtPolicyIPDiscoveryProfileRead,
		Update: resourceNsxtPolicyIPDiscoveryProfileUpdate,
		Delete: resourceNsxtPolicyIPDiscoveryProfileDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"context":      getContextSchema(false, false, false),
			"arp_nd_binding_timeout": {
				Type:         schema.TypeInt,
				Description:  "ARP and ND cache timeout (in minutes)",
				Optional:     true,
				Default:      10,
				ValidateFunc: validation.IntBetween(5, 120),
			},
			"duplicate_ip_detection_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Duplicate IP detection",
			},
			"arp_binding_limit": {
				Type:         schema.TypeInt,
				Description:  "Maximum number of ARP bindings",
				Optional:     true,
				Default:      1,
				ValidateFunc: validation.IntBetween(1, 256),
			},
			"arp_snooping_enabled": {
				Type:        schema.TypeBool,
				Description: "Is ARP snooping enabled or not",
				Optional:    true,
				Default:     true,
			},
			"dhcp_snooping_enabled": {
				Type:        schema.TypeBool,
				Description: "Is DHCP snooping enabled or not",
				Optional:    true,
				Default:     true,
			},
			"vmtools_enabled": {
				Type:        schema.TypeBool,
				Description: "Is VM tools enabled or not",
				Optional:    true,
				Default:     true,
			},
			"dhcp_snooping_v6_enabled": {
				Type:        schema.TypeBool,
				Description: "Is DHCP snoping v6 enabled or not",
				Optional:    true,
				Default:     false,
			},
			"nd_snooping_enabled": {
				Type:        schema.TypeBool,
				Description: "Is ND snooping enabled or not",
				Optional:    true,
				Default:     false,
			},
			"nd_snooping_limit": {
				Type:         schema.TypeInt,
				Description:  "Maximum number of ND (Neighbor Discovery Protocol) bindings",
				Optional:     true,
				ValidateFunc: validation.IntBetween(2, 15),
				Default:      3,
			},
			"vmtools_v6_enabled": {
				Type:        schema.TypeBool,
				Description: "Is VM tools enabled or not",
				Optional:    true,
				Default:     false,
			},
			"tofu_enabled": {
				Type:        schema.TypeBool,
				Description: "Is TOFU enabled or not",
				Optional:    true,
				Default:     true,
			},
		},
	}
}

func ipDiscoveryProfileObjFromSchema(d *schema.ResourceData) model.IPDiscoveryProfile {
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	arpNdBindingTimeout := int64(d.Get("arp_nd_binding_timeout").(int))
	duplicateIPDetectionEnabled := d.Get("duplicate_ip_detection_enabled").(bool)

	arpBindingLimit := int64(d.Get("arp_binding_limit").(int))
	arpSnoopingEnabled := d.Get("arp_snooping_enabled").(bool)
	dhcpSnoopingEnabled := d.Get("dhcp_snooping_enabled").(bool)
	vmtoolsEnabled := d.Get("vmtools_enabled").(bool)

	dhcpSnoopingV6Enabled := d.Get("dhcp_snooping_v6_enabled").(bool)
	ndSnoopingEnabled := d.Get("nd_snooping_enabled").(bool)
	ndSnoopingLimit := int64(d.Get("nd_snooping_limit").(int))
	vmtoolsV6Enabled := d.Get("vmtools_v6_enabled").(bool)
	toFuEnabled := d.Get("tofu_enabled").(bool)

	return model.IPDiscoveryProfile{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		ArpNdBindingTimeout: &arpNdBindingTimeout,
		DuplicateIpDetection: &model.DuplicateIPDetectionOptions{
			DuplicateIpDetectionEnabled: &duplicateIPDetectionEnabled,
		},
		IpV4DiscoveryOptions: &model.IPv4DiscoveryOptions{
			ArpSnoopingConfig: &model.ArpSnoopingConfig{
				ArpBindingLimit:    &arpBindingLimit,
				ArpSnoopingEnabled: &arpSnoopingEnabled,
			},
			DhcpSnoopingEnabled: &dhcpSnoopingEnabled,
			VmtoolsEnabled:      &vmtoolsEnabled,
		},
		IpV6DiscoveryOptions: &model.IPv6DiscoveryOptions{
			DhcpSnoopingV6Enabled: &dhcpSnoopingV6Enabled,
			NdSnoopingConfig: &model.NdSnoopingConfig{
				NdSnoopingEnabled: &ndSnoopingEnabled,
				NdSnoopingLimit:   &ndSnoopingLimit,
			},
			VmtoolsV6Enabled: &vmtoolsV6Enabled,
		},
		TofuEnabled: &toFuEnabled,
	}
}

func resourceNsxtPolicyIPDiscoveryProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewIpDiscoveryProfilesClient(sessionContext, connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPDiscoveryProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyIPDiscoveryProfileExists)
	if err != nil {
		return err
	}

	obj := ipDiscoveryProfileObjFromSchema(d)

	// Create the resource using PATCH
	log.Printf("[INFO] Creating IPDiscoveryProfile with ID %s", id)
	boolFalse := false
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewIpDiscoveryProfilesClient(context, connector)
	err = client.Patch(id, obj, &boolFalse)
	if err != nil {
		return handleCreateError("IPDiscoveryProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPDiscoveryProfileRead(d, m)
}

func resourceNsxtPolicyIPDiscoveryProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPDiscoveryProfile ID")
	}

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewIpDiscoveryProfilesClient(context, connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "IPDiscoveryProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("arp_nd_binding_timeout", obj.ArpNdBindingTimeout)
	d.Set("duplicate_ip_detection_enabled", obj.DuplicateIpDetection.DuplicateIpDetectionEnabled)
	d.Set("arp_binding_limit", obj.IpV4DiscoveryOptions.ArpSnoopingConfig.ArpBindingLimit)
	d.Set("arp_snooping_enabled", obj.IpV4DiscoveryOptions.ArpSnoopingConfig.ArpSnoopingEnabled)
	d.Set("dhcp_snooping_enabled", obj.IpV4DiscoveryOptions.DhcpSnoopingEnabled)
	d.Set("vmtools_enabled", obj.IpV4DiscoveryOptions.VmtoolsEnabled)
	d.Set("dhcp_snooping_v6_enabled", obj.IpV6DiscoveryOptions.DhcpSnoopingV6Enabled)
	d.Set("nd_snooping_enabled", obj.IpV6DiscoveryOptions.NdSnoopingConfig.NdSnoopingEnabled)
	d.Set("nd_snooping_limit", obj.IpV6DiscoveryOptions.NdSnoopingConfig.NdSnoopingLimit)
	d.Set("vmtools_v6_enabled", obj.IpV6DiscoveryOptions.VmtoolsV6Enabled)
	d.Set("tofu_enabled", obj.TofuEnabled)

	return nil
}

func resourceNsxtPolicyIPDiscoveryProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewIpDiscoveryProfilesClient(context, connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPDiscoveryProfile ID")
	}

	// Read the rest of the configured parameters
	obj := ipDiscoveryProfileObjFromSchema(d)

	// Create the resource using PATCH
	log.Printf("[INFO] Updating IPDiscoveryProfile with ID %s", id)
	boolFalse := false
	err = client.Patch(id, obj, &boolFalse)
	if err != nil {
		return handleUpdateError("IPDiscoveryProfile", id, err)
	}

	return resourceNsxtPolicyIPDiscoveryProfileRead(d, m)
}

func resourceNsxtPolicyIPDiscoveryProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPDiscoveryProfile ID")
	}

	var err error
	connector := getPolicyConnector(m)
	boolFalse := false

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := infra.NewIpDiscoveryProfilesClient(context, connector)
	err = client.Delete(id, &boolFalse)

	if err != nil {
		return handleDeleteError("IPDiscoveryProfile", id, err)
	}

	return nil
}
