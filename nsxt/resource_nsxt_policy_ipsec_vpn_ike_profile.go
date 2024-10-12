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

var iPSecVpnIkeProfileDhGroupsValues = []string{
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP2,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP5,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP20,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP16,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP15,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP14,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP21,
	model.IPSecVpnIkeProfile_DH_GROUPS_GROUP19,
}

var iPSecVpnIkeProfileDigestAlgorithmsValues = []string{
	model.IPSecVpnIkeProfile_DIGEST_ALGORITHMS_SHA2_256,
	model.IPSecVpnIkeProfile_DIGEST_ALGORITHMS_SHA2_512,
	model.IPSecVpnIkeProfile_DIGEST_ALGORITHMS_SHA2_384,
	model.IPSecVpnIkeProfile_DIGEST_ALGORITHMS_SHA1,
}

var iPSecVpnIkeProfileEncryptionAlgorithmsValues = []string{
	model.IPSecVpnIkeProfile_ENCRYPTION_ALGORITHMS_GCM_128,
	model.IPSecVpnIkeProfile_ENCRYPTION_ALGORITHMS_128,
	model.IPSecVpnIkeProfile_ENCRYPTION_ALGORITHMS_256,
	model.IPSecVpnIkeProfile_ENCRYPTION_ALGORITHMS_GCM_256,
	model.IPSecVpnIkeProfile_ENCRYPTION_ALGORITHMS_GCM_192,
}

var iPSecVpnIkeProfileIkeVersionValues = []string{
	model.IPSecVpnIkeProfile_IKE_VERSION_V1,
	model.IPSecVpnIkeProfile_IKE_VERSION_V2,
	model.IPSecVpnIkeProfile_IKE_VERSION_FLEX,
}

func resourceNsxtPolicyIPSecVpnIkeProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyIPSecVpnIkeProfileCreate,
		Read:   resourceNsxtPolicyIPSecVpnIkeProfileRead,
		Update: resourceNsxtPolicyIPSecVpnIkeProfileUpdate,
		Delete: resourceNsxtPolicyIPSecVpnIkeProfileDelete,
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
			"dh_groups": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(iPSecVpnIkeProfileDhGroupsValues, false),
				},
				Required: true,
			},
			"digest_algorithms": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(iPSecVpnIkeProfileDigestAlgorithmsValues, false),
				},
				Optional: true,
			},
			"encryption_algorithms": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(iPSecVpnIkeProfileEncryptionAlgorithmsValues, false),
				},
				Required: true,
			},
			"ike_version": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(iPSecVpnIkeProfileIkeVersionValues, false),
				Default:      model.IPSecVpnIkeProfile_IKE_VERSION_V2,
				Optional:     true,
			},
			"sa_life_time": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  86400,
			},
		},
	}
}

func resourceNsxtPolicyIPSecVpnIkeProfileExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewIpsecVpnIkeProfilesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyIPSecVpnIkeProfileCreate(d *schema.ResourceData, m interface{}) error {
	if isPolicyGlobalManager(m) {
		return resourceNotSupportedError()
	}

	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyIPSecVpnIkeProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	dhGroups := getStringListFromSchemaSet(d, "dh_groups")
	digestAlgorithms := getStringListFromSchemaSet(d, "digest_algorithms")
	encryptionAlgorithms := getStringListFromSchemaSet(d, "encryption_algorithms")
	ikeVersion := d.Get("ike_version").(string)
	saLifeTime := int64(d.Get("sa_life_time").(int))

	obj := model.IPSecVpnIkeProfile{
		DisplayName:          &displayName,
		Description:          &description,
		Tags:                 tags,
		DhGroups:             dhGroups,
		DigestAlgorithms:     digestAlgorithms,
		EncryptionAlgorithms: encryptionAlgorithms,
		IkeVersion:           &ikeVersion,
		SaLifeTime:           &saLifeTime,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating IPSecVpnIkeProfile with ID %s", id)
	client := infra.NewIpsecVpnIkeProfilesClient(connector)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("IPSecVpnIkeProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyIPSecVpnIkeProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnIkeProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnIkeProfile ID")
	}

	client := infra.NewIpsecVpnIkeProfilesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "IPSecVpnIkeProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("dh_groups", obj.DhGroups)
	d.Set("digest_algorithms", obj.DigestAlgorithms)
	d.Set("encryption_algorithms", obj.EncryptionAlgorithms)
	d.Set("ike_version", obj.IkeVersion)
	d.Set("sa_life_time", obj.SaLifeTime)

	return nil
}

func resourceNsxtPolicyIPSecVpnIkeProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnIkeProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	dhGroups := getStringListFromSchemaSet(d, "dh_groups")
	digestAlgorithms := getStringListFromSchemaSet(d, "digest_algorithms")
	encryptionAlgorithms := getStringListFromSchemaSet(d, "encryption_algorithms")
	ikeVersion := d.Get("ike_version").(string)
	saLifeTime := int64(d.Get("sa_life_time").(int))

	obj := model.IPSecVpnIkeProfile{
		DisplayName:          &displayName,
		Description:          &description,
		Tags:                 tags,
		DhGroups:             dhGroups,
		DigestAlgorithms:     digestAlgorithms,
		EncryptionAlgorithms: encryptionAlgorithms,
		IkeVersion:           &ikeVersion,
		SaLifeTime:           &saLifeTime,
	}

	client := infra.NewIpsecVpnIkeProfilesClient(connector)
	err := client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("IPSecVpnIkeProfile", id, err)
	}

	return resourceNsxtPolicyIPSecVpnIkeProfileRead(d, m)
}

func resourceNsxtPolicyIPSecVpnIkeProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining IPSecVpnIkeProfile ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewIpsecVpnIkeProfilesClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("IPSecVpnIkeProfile", id, err)
	}

	return nil
}
