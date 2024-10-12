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

var lBClientSslProfileCipherGroupLabelValues = []string{
	model.LBClientSslProfile_CIPHER_GROUP_LABEL_HIGH_COMPATIBILITY,
	model.LBClientSslProfile_CIPHER_GROUP_LABEL_HIGH_SECURITY,
	model.LBClientSslProfile_CIPHER_GROUP_LABEL_CUSTOM,
	model.LBClientSslProfile_CIPHER_GROUP_LABEL_BALANCED,
}

func resourceNsxtPolicyLBClientSslProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBClientSslProfileCreate,
		Read:   resourceNsxtPolicyLBClientSslProfileRead,
		Update: resourceNsxtPolicyLBClientSslProfileUpdate,
		Delete: resourceNsxtPolicyLBClientSslProfileDelete,
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
			"cipher_group_label": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(lBClientSslProfileCipherGroupLabelValues, false),
				Optional:     true,
				Default:      model.LBClientSslProfile_CIPHER_GROUP_LABEL_BALANCED,
				Description:  "A label of cipher group which is mostly consumed by GUI. Default value is BALANCED.",
			},
			"ciphers": getSSLCiphersSchema(),
			"is_fips": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "This flag is set to true when all the ciphers and protocols are FIPS compliant. It is set to false when one of the ciphers or protocols are not FIPS compliant.",
			},
			"is_secure": getIsSecureSchema(),
			"prefer_server_ciphers": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "During SSL handshake as part of the SSL client Hello client sends an ordered list of ciphers that it can support (or prefers) and typically server selects the first one from the top of that list it can also support. For Perfect Forward Secrecy(PFS), server could override the client's preference.",
			},
			"protocols": getSSLProtocolsSchema(),
			"session_cache_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "If set to true, SSL session caching allows SSL client and server to reuse previously negotiated security parameters avoiding the expensive public key operation during handshake.",
			},
			"session_cache_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     300,
				Description: "Session cache timeout specifies how long the SSL session parameters are held on to and can be reused. Default value is 300.",
			},
		},
	}
}

func resourceNsxtPolicyLBClientSslProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	client := infra.NewLbClientSslProfilesClient(connector)
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyLBClientSslProfilePatch(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	cipherGroupLabel := d.Get("cipher_group_label").(string)
	ciphers := getStringListFromSchemaSet(d, "ciphers")
	preferServerCiphers := d.Get("prefer_server_ciphers").(bool)
	protocols := getStringListFromSchemaSet(d, "protocols")
	sessionCacheEnabled := d.Get("session_cache_enabled").(bool)
	sessionCacheTimeout := int64(d.Get("session_cache_timeout").(int))

	obj := model.LBClientSslProfile{
		DisplayName:         &displayName,
		Description:         &description,
		Tags:                tags,
		CipherGroupLabel:    &cipherGroupLabel,
		Ciphers:             ciphers,
		PreferServerCiphers: &preferServerCiphers,
		Protocols:           protocols,
		SessionCacheEnabled: &sessionCacheEnabled,
		SessionCacheTimeout: &sessionCacheTimeout,
	}

	log.Printf("[INFO] Patching LBClientSslProfile with ID %s", id)

	client := infra.NewLbClientSslProfilesClient(connector)
	return client.Patch(id, obj)
}

func resourceNsxtPolicyLBClientSslProfileCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBClientSslProfileExists)
	if err != nil {
		return err
	}

	err = resourceNsxtPolicyLBClientSslProfilePatch(d, m, id)
	if err != nil {
		return handleCreateError("LBClientSslProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBClientSslProfileRead(d, m)
}

func resourceNsxtPolicyLBClientSslProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBClientSslProfile ID")
	}

	var obj model.LBClientSslProfile
	client := infra.NewLbClientSslProfilesClient(connector)
	var err error
	obj, err = client.Get(id)
	if err != nil {
		return handleReadError(d, "LBClientSslProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("cipher_group_label", obj.CipherGroupLabel)
	d.Set("ciphers", obj.Ciphers)
	d.Set("is_fips", obj.IsFips)
	d.Set("is_secure", obj.IsSecure)
	d.Set("prefer_server_ciphers", obj.PreferServerCiphers)
	d.Set("protocols", obj.Protocols)
	d.Set("session_cache_enabled", obj.SessionCacheEnabled)
	d.Set("session_cache_timeout", obj.SessionCacheTimeout)

	return nil
}

func resourceNsxtPolicyLBClientSslProfileUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBClientSslProfile ID")
	}

	err := resourceNsxtPolicyLBClientSslProfilePatch(d, m, id)
	if err != nil {
		return handleUpdateError("LBClientSslProfile", id, err)
	}

	return resourceNsxtPolicyLBClientSslProfileRead(d, m)
}

func resourceNsxtPolicyLBClientSslProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBClientSslProfile ID")
	}

	forceParam := true
	connector := getPolicyConnector(m)
	var err error
	client := infra.NewLbClientSslProfilesClient(connector)
	err = client.Delete(id, &forceParam)

	if err != nil {
		return handleDeleteError("LBClientSslProfile", id, err)
	}

	return nil
}
