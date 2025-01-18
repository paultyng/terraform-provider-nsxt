// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyEvpnTenant() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyEvpnTenantCreate,
		Read:   resourceNsxtPolicyEvpnTenantRead,
		Update: resourceNsxtPolicyEvpnTenantUpdate,
		Delete: resourceNsxtPolicyEvpnTenantDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":              getNsxIDSchema(),
			"path":                getPathSchema(),
			"display_name":        getDisplayNameSchema(),
			"description":         getDescriptionSchema(),
			"revision":            getRevisionSchema(),
			"tag":                 getTagsSchema(),
			"transport_zone_path": getPolicyPathSchema(true, false, "Policy path to overlay transport zone"),
			"vni_pool_path":       getPolicyPathSchema(true, false, "Policy path to the vni pool used for Evpn in ROUTE-SERVER mode"),
			"mapping": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vlans": {
							Type:        schema.TypeString,
							Description: "Values for attribute key",
							Required:    true,
						},
						"vnis": {
							Type:        schema.TypeString,
							Description: "Values for attribute key",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func resourceNsxtPolicyEvpnTenantExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error
	client := infra.NewEvpnTenantConfigsClient(connector)
	_, err = client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func getEvpnTenantMappingsFromSchema(d *schema.ResourceData) []model.VlanVniRangePair {
	var results []model.VlanVniRangePair
	mappings := d.Get("mapping").(*schema.Set).List()
	for _, item := range mappings {
		mapping := item.(map[string]interface{})
		vlans := mapping["vlans"].(string)
		vnis := mapping["vnis"].(string)
		result := model.VlanVniRangePair{
			Vlans: &vlans,
			Vnis:  &vnis,
		}
		results = append(results, result)
	}

	return results
}

func setEvpnTenantMappingsInSchema(d *schema.ResourceData, mappings []model.VlanVniRangePair) error {
	var mappingList []map[string]interface{}
	for _, item := range mappings {
		mapping := make(map[string]interface{})
		mapping["vlans"] = item.Vlans
		mapping["vnis"] = item.Vnis
		mappingList = append(mappingList, mapping)
	}

	return d.Set("mapping", mappingList)
}

func policyEvpnTenantPatch(id string, d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	tzPath := d.Get("transport_zone_path").(string)
	vniPoolPath := d.Get("vni_pool_path").(string)
	mappings := getEvpnTenantMappingsFromSchema(d)

	obj := model.EvpnTenantConfig{
		DisplayName:       &displayName,
		Description:       &description,
		Tags:              tags,
		TransportZonePath: &tzPath,
		VniPoolPath:       &vniPoolPath,
		Mappings:          mappings,
	}

	// Create the resource using PATCH
	client := infra.NewEvpnTenantConfigsClient(connector)
	return client.Patch(id, obj)
}

func resourceNsxtPolicyEvpnTenantCreate(d *schema.ResourceData, m interface{}) error {

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyEvpnTenantExists)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Creating Evpn Tenant with ID %s", id)
	err = policyEvpnTenantPatch(id, d, m)
	if err != nil {
		return handleCreateError("Evpn Tenant", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyEvpnTenantRead(d, m)
}

func resourceNsxtPolicyEvpnTenantRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Evpn Tenant ID")
	}

	client := infra.NewEvpnTenantConfigsClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Evpn Tenant", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("transport_zone_path", obj.TransportZonePath)
	d.Set("vni_pool_path", obj.VniPoolPath)

	err = setEvpnTenantMappingsInSchema(d, obj.Mappings)
	if err != nil {
		return err
	}
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	return nil
}

func resourceNsxtPolicyEvpnTenantUpdate(d *schema.ResourceData, m interface{}) error {

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Evpn Tenant ID")
	}

	log.Printf("[INFO] Creating Evpn Tenant with ID %s", id)
	err := policyEvpnTenantPatch(id, d, m)
	if err != nil {
		return handleUpdateError("Evpn Tenant", id, err)
	}

	return resourceNsxtPolicyEvpnTenantRead(d, m)
}

func resourceNsxtPolicyEvpnTenantDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining Evpn Tenant ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewEvpnTenantConfigsClient(connector)
	err := client.Delete(id)

	if err != nil {
		return handleDeleteError("Evpn Tenant", id, err)
	}

	return nil
}
