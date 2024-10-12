// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s/locale_services/ospf"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var ospfAreaTypeValues = []string{
	model.OspfAreaConfig_AREA_TYPE_NORMAL,
	model.OspfAreaConfig_AREA_TYPE_NSSA,
}

var ospfAreaAuthModeValues = []string{
	model.OspfAuthenticationConfig_MODE_NONE,
	model.OspfAuthenticationConfig_MODE_PASSWORD,
	model.OspfAuthenticationConfig_MODE_MD5,
}

func resourceNsxtPolicyOspfArea() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyOspfAreaCreate,
		Read:   resourceNsxtPolicyOspfAreaRead,
		Update: resourceNsxtPolicyOspfAreaUpdate,
		Delete: resourceNsxtPolicyOspfAreaDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyOspfAreaImport,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"ospf_path":    getPolicyPathSchema(true, true, "Policy path to the OSPF config for this area"),
			"area_id": {
				// TODO: add validator
				Description: "OSPF area ID in decimal or dotted format",
				Type:        schema.TypeString,
				Required:    true,
			},
			"area_type": {
				Type:         schema.TypeString,
				Description:  "OSPF area type",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ospfAreaTypeValues, false),
				Default:      model.OspfAreaConfig_AREA_TYPE_NSSA,
			},
			"auth_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Authentication mode",
				ValidateFunc: validation.StringInSlice(ospfAreaAuthModeValues, false),
				Default:      model.OspfAuthenticationConfig_MODE_NONE,
			},
			"key_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 255),
				Description:  "Authentication secret key id for MD5 auth mode",
				Sensitive:    true,
			},
			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Authentication secret",
				Sensitive:   true,
			},
		},
	}
}

func resourceNsxtPolicyOspfAreaExists(gwID string, localeServiceID string, areaID string, isGlobalManager bool, connector client.Connector) (bool, error) {

	client := ospf.NewAreasClient(connector)
	_, err := client.Get(gwID, localeServiceID, areaID)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func parseOspfConfigPath(path string) (string, string) {
	segs := strings.Split(path, "/")
	if len(segs) < 7 {
		// Bad ospf path
		return "", ""
	}
	// Tier-0 ID and locale service ID
	return segs[3], segs[5]
}

func policyOspfAreaPatch(d *schema.ResourceData, m interface{}, id string) error {

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	ospfPath := d.Get("ospf_path").(string)
	areaID := d.Get("area_id").(string)
	areaType := d.Get("area_type").(string)
	authMode := d.Get("auth_mode").(string)
	keyID := int64(d.Get("key_id").(int))
	secretKey := d.Get("secret_key").(string)

	gwID, localeServiceID := parseOspfConfigPath(ospfPath)
	if gwID == "" {
		return fmt.Errorf("Expecting OSPF config path, got %s", ospfPath)
	}

	authConfig := model.OspfAuthenticationConfig{
		Mode: &authMode,
	}

	if authMode != model.OspfAuthenticationConfig_MODE_NONE {
		authConfig.SecretKey = &secretKey
	}

	if authMode == model.OspfAuthenticationConfig_MODE_MD5 {
		authConfig.KeyId = &keyID
	}

	obj := model.OspfAreaConfig{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		AreaId:         &areaID,
		AreaType:       &areaType,
		Authentication: &authConfig,
	}

	connector := getPolicyConnector(m)
	client := ospf.NewAreasClient(connector)
	_, err := client.Patch(gwID, localeServiceID, id, obj)
	return err
}

func resourceNsxtPolicyOspfAreaCreate(d *schema.ResourceData, m interface{}) error {

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}

	// Only a single OSPF area is supported so far per OSPF config, thus we don't check
	// here for nsx id existence

	err := policyOspfAreaPatch(d, m, id)
	if err != nil {
		return handleCreateError("Ospf Area", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyOspfAreaRead(d, m)
}

func resourceNsxtPolicyOspfAreaRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining OspfArea ID")
	}

	ospfPath := d.Get("ospf_path").(string)
	gwID, localeServiceID := parseOspfConfigPath(ospfPath)
	if gwID == "" {
		return fmt.Errorf("Expecting OSPF config path, got %s", ospfPath)
	}

	client := ospf.NewAreasClient(connector)
	obj, err := client.Get(gwID, localeServiceID, id)
	if err != nil {
		return handleReadError(d, "Ospf Area", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("area_id", obj.AreaId)
	d.Set("area_type", obj.AreaType)
	if obj.Authentication == nil {
		d.Set("auth_mode", model.OspfAuthenticationConfig_MODE_NONE)
	} else {
		// Key ID and secret key are sensitive and not returned on API
		d.Set("auth_mode", obj.Authentication.Mode)
	}
	d.Set("path", obj.Path)
	d.Set("nsx_id", obj.Id)
	d.Set("revision", obj.Revision)

	return nil
}

func resourceNsxtPolicyOspfAreaUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Ospf Area ID")
	}

	err := policyOspfAreaPatch(d, m, id)
	if err != nil {
		return handleUpdateError("Ospf Area", id, err)
	}

	return resourceNsxtPolicyOspfAreaRead(d, m)
}

func resourceNsxtPolicyOspfAreaDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Ospf Area ID")
	}

	connector := getPolicyConnector(m)
	ospfPath := d.Get("ospf_path").(string)
	gwID, localeServiceID := parseOspfConfigPath(ospfPath)
	if gwID == "" {
		return fmt.Errorf("Expecting OSPF config path, got %s", ospfPath)
	}

	client := ospf.NewAreasClient(connector)
	err := client.Delete(gwID, localeServiceID, id)
	if err != nil {
		return handleDeleteError("Ospf Area", id, err)
	}

	return nil
}

func resourceNsxtPolicyOspfAreaImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	s := strings.Split(importID, "/")
	if len(s) != 3 {
		return nil, fmt.Errorf("Please provide <tier0-id>/<locale-service-id>/<area-id> as an input")
	}

	gwID := s[0]
	serviceID := s[1]
	areaID := s[2]
	connector := getPolicyConnector(m)
	client := ospf.NewAreasClient(connector)

	area, err := client.Get(gwID, serviceID, areaID)
	if err != nil {
		return nil, err
	}
	d.Set("ospf_path", area.ParentPath)

	d.SetId(areaID)

	return []*schema.ResourceData{d}, nil
}
