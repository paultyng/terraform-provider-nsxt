// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var cryptoProtocolsValues = []string{"TLS_V1", "TLS_V1_1", "TLS_V1_2"}

func resourceNsxtPolicyMetadataProxy() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyMetadataProxyCreate,
		Read:   resourceNsxtPolicyMetadataProxyRead,
		Update: resourceNsxtPolicyMetadataProxyUpdate,
		Delete: resourceNsxtPolicyMetadataProxyDelete,
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
			"crypto_protocols": {
				Type:        schema.TypeList,
				Computed:    true,
				Optional:    true,
				Description: "Metadata proxy supported cryptographic protocols",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(cryptoProtocolsValues, false),
				},
			},
			"edge_cluster_path": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Policy path to Edge Cluster",
				ValidateFunc: validatePolicyPath(),
			},
			"enable_standby_relocation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag to enable standby relocation",
			},
			"preferred_edge_paths": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Preferred Edge Paths",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validatePolicyPath(),
				},
			},
			"secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Secret",
			},
			"server_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Server Address",
			},
			"server_certificates": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Policy paths to Certificate Authority (CA) certificates",
				Sensitive:   true,
				Elem: &schema.Schema{
					Type:      schema.TypeString,
					Sensitive: true,
				},
			},
		},
	}
}

func resourceNsxtPolicyMetadataProxyExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewMetadataProxiesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func getMetadataProxyFromSchema(d *schema.ResourceData) model.MetadataProxyConfig {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	cryptoProtocols := interfaceListToStringList(d.Get("crypto_protocols").([]interface{}))
	edgeClusterPath := d.Get("edge_cluster_path").(string)
	enableStandbyRelocation := d.Get("enable_standby_relocation").(bool)
	preferredEdgePaths := interfaceListToStringList(d.Get("preferred_edge_paths").([]interface{}))
	secret := d.Get("secret").(string)
	serverAddress := d.Get("server_address").(string)
	serverCertificates := interfaceListToStringList(d.Get("server_certificates").([]interface{}))
	return model.MetadataProxyConfig{
		DisplayName:             &displayName,
		Description:             &description,
		Tags:                    tags,
		CryptoProtocols:         cryptoProtocols,
		EdgeClusterPath:         &edgeClusterPath,
		EnableStandbyRelocation: &enableStandbyRelocation,
		PreferredEdgePaths:      preferredEdgePaths,
		Secret:                  &secret,
		ServerAddress:           &serverAddress,
		ServerCertificates:      serverCertificates,
	}
}

func resourceNsxtPolicyMetadataProxyCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyMetadataProxyExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := infra.NewMetadataProxiesClient(connector)
	obj := getMetadataProxyFromSchema(d)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("PolicyMetadataProxy", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyMetadataProxyRead(d, m)
}

func resourceNsxtPolicyMetadataProxyRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining PolicyMetadataProxy ID")
	}
	client := infra.NewMetadataProxiesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "PolicyMetadataProxy", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("crypto_protocols", stringList2Interface(obj.CryptoProtocols))
	d.Set("edge_cluster_path", obj.EdgeClusterPath)
	d.Set("enable_standby_relocation", obj.EnableStandbyRelocation)
	d.Set("preferred_edge_paths", stringList2Interface(obj.PreferredEdgePaths))
	d.Set("server_address", obj.ServerAddress)

	return nil
}

func resourceNsxtPolicyMetadataProxyUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining PolicyMetadataProxy ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewMetadataProxiesClient(connector)

	obj := getMetadataProxyFromSchema(d)
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision
	_, err := client.Update(id, obj)
	if err != nil {
		return handleUpdateError("PolicyMetadataProxy", id, err)
	}

	return resourceNsxtPolicyMetadataProxyRead(d, m)
}

func resourceNsxtPolicyMetadataProxyDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining PolicyMetadataProxy ID")
	}
	connector := getPolicyConnector(m)
	client := infra.NewMetadataProxiesClient(connector)
	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("PolicyMetadataProxy", id, err)
	}
	return nil
}
