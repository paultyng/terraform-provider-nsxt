// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var ipSecVpnTunnelProfileDfPolicyValues = []string{
	model.IPSecVpnTunnelProfile_DF_POLICY_COPY,
	model.IPSecVpnTunnelProfile_DF_POLICY_CLEAR,
}

var ipSecVpnTunnelProfileDhGroupsValues = []string{
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP2,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP5,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP19,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP14,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP16,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP15,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP20,
	model.IPSecVpnTunnelProfile_DH_GROUPS_GROUP21,
}

var ipSecVpnTunnelProfileDigestAlgorithmsValues = []string{
	model.IPSecVpnTunnelProfile_DIGEST_ALGORITHMS_SHA2_256,
	model.IPSecVpnTunnelProfile_DIGEST_ALGORITHMS_SHA1,
	model.IPSecVpnTunnelProfile_DIGEST_ALGORITHMS_SHA2_512,
	model.IPSecVpnTunnelProfile_DIGEST_ALGORITHMS_SHA2_384,
}

var ipSecVpnTunnelProfileEncryptionAlgorithmsValues = []string{
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_AES_256,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_AES_GCM_192,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_AES_GCM_256,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_NO_ENCRYPTION_AUTH_AES_GMAC_128,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_AES_GCM_128,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_AES_128,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_NO_ENCRYPTION_AUTH_AES_GMAC_256,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_NO_ENCRYPTION_AUTH_AES_GMAC_192,
	model.IPSecVpnTunnelProfile_ENCRYPTION_ALGORITHMS_NO_ENCRYPTION,
}

func resourceNsxtPolicyIPSecVpnTunnelProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPSecVpnTunnelProfileCreate,
		Read:   resourceNsxtPolicyIPSecVpnTunnelProfileRead,
		Update: resourceNsxtPolicyIPSecVpnTunnelProfileUpdate,
		Delete: resourceNsxtPolicyIPSecVpnTunnelProfileDelete,
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
			"df_policy": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(ipSecVpnTunnelProfileDfPolicyValues, false),
				Optional:     true,
				Default:      model.IPSecVpnTunnelProfile_DF_POLICY_COPY,
			},
			"dh_groups": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ipSecVpnTunnelProfileDhGroupsValues, false),
				},
				Required: true,
			},
			"digest_algorithms": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ipSecVpnTunnelProfileDigestAlgorithmsValues, false),
				},
				Optional: true,
			},
			"enable_perfect_forward_secrecy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"encryption_algorithms": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ipSecVpnTunnelProfileEncryptionAlgorithmsValues, false),
				},
				Required: true,
			},
			"sa_life_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
			},
		},
	}
}

func resourceNsxtPolicyIPSecVpnTunnelProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewIpsecVpnTunnelProfilesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPSecVpnTunnelProfileCreate(d *schema.ResourceData, m interface{}) error {
	if isPolicyGlobalManager(m) {
		return resourceNotSupportedError()
	}
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyIPSecVpnTunnelProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	dfPolicy := d.Get("df_policy").(string)
	dhGroups := getStringListFromSchemaSet(d, "dh_groups")
	digestAlgorithms := getStringListFromSchemaSet(d, "digest_algorithms")
	enablePerfectForwardSecrecy := d.Get("enable_perfect_forward_secrecy").(bool)
	encryptionAlgorithms := getStringListFromSchemaSet(d, "encryption_algorithms")
	saLifeTime := int64(d.Get("sa_life_time").(int))

	obj := model.IPSecVpnTunnelProfile{
		DisplayName:                 &displayName,
		Description:                 &description,
		Tags:                        tags,
		DfPolicy:                    &dfPolicy,
		DhGroups:                    dhGroups,
		DigestAlgorithms:            digestAlgorithms,
		EnablePerfectForwardSecrecy: &enablePerfectForwardSecrecy,
		EncryptionAlgorithms:        encryptionAlgorithms,
		SaLifeTime:                  &saLifeTime,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating IPSecVpnTunnelProfile with ID %s", id)
	client := infra.NewIpsecVpnTunnelProfilesClient(connector)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("IPSecVpnTunnelProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPSecVpnTunnelProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnTunnelProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnTunnelProfile ID")
	}

	client := infra.NewIpsecVpnTunnelProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "IPSecVpnTunnelProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("df_policy", obj.DfPolicy)
	d.Set("dh_groups", obj.DhGroups)
	d.Set("digest_algorithms", obj.DigestAlgorithms)
	d.Set("enable_perfect_forward_secrecy", obj.EnablePerfectForwardSecrecy)
	d.Set("encryption_algorithms", obj.EncryptionAlgorithms)
	d.Set("sa_life_time", obj.SaLifeTime)

	return nil
}

func resourceNsxtPolicyIPSecVpnTunnelProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnTunnelProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	dfPolicy := d.Get("df_policy").(string)
	dhGroups := getStringListFromSchemaSet(d, "dh_groups")
	digestAlgorithms := getStringListFromSchemaSet(d, "digest_algorithms")
	enablePerfectForwardSecrecy := d.Get("enable_perfect_forward_secrecy").(bool)
	encryptionAlgorithms := getStringListFromSchemaSet(d, "encryption_algorithms")
	saLifeTime := int64(d.Get("sa_life_time").(int))

	obj := model.IPSecVpnTunnelProfile{
		DisplayName:                 &displayName,
		Description:                 &description,
		Tags:                        tags,
		DfPolicy:                    &dfPolicy,
		DhGroups:                    dhGroups,
		DigestAlgorithms:            digestAlgorithms,
		EnablePerfectForwardSecrecy: &enablePerfectForwardSecrecy,
		EncryptionAlgorithms:        encryptionAlgorithms,
		SaLifeTime:                  &saLifeTime,
	}

	client := infra.NewIpsecVpnTunnelProfilesClient(connector)
	err := client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("IPSecVpnTunnelProfile", id, err)
	}

	return resourceNsxtPolicyIPSecVpnTunnelProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnTunnelProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnTunnelProfile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewIpsecVpnTunnelProfilesClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("IPSecVpnTunnelProfile", id, err)
	}

	return nil
}
