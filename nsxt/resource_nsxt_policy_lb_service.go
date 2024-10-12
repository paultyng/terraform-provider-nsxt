// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var lBServiceErrorLogLevelValues = []string{
	model.LBService_ERROR_LOG_LEVEL_ERROR,
	model.LBService_ERROR_LOG_LEVEL_EMERGENCY,
	model.LBService_ERROR_LOG_LEVEL_ALERT,
	model.LBService_ERROR_LOG_LEVEL_DEBUG,
	model.LBService_ERROR_LOG_LEVEL_CRITICAL,
	model.LBService_ERROR_LOG_LEVEL_INFO,
	model.LBService_ERROR_LOG_LEVEL_WARNING,
}

var lBServiceSizeValues = []string{
	model.LBService_SIZE_XLARGE,
	model.LBService_SIZE_LARGE,
	model.LBService_SIZE_MEDIUM,
	model.LBService_SIZE_DLB,
	model.LBService_SIZE_SMALL,
}

func resourceNsxtPolicyLBService() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLBServiceCreate,
		Read:   resourceNsxtPolicyLBServiceRead,
		Update: resourceNsxtPolicyLBServiceUpdate,
		Delete: resourceNsxtPolicyLBServiceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":            getNsxIDSchema(),
			"path":              getPathSchema(),
			"display_name":      getDisplayNameSchema(),
			"description":       getDescriptionSchema(),
			"revision":          getRevisionSchema(),
			"tag":               getTagsSchema(),
			"connectivity_path": getPolicyPathSchema(false, false, "Policy path for connected policy object"),
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Flag to enable the Service",
				Optional:    true,
				Default:     true,
			},
			"error_log_level": {
				Type:         schema.TypeString,
				ValidateFunc: validation.StringInSlice(lBServiceErrorLogLevelValues, false),
				Description:  "Log level for Load Balancer Service messages",
				Optional:     true,
				Default:      model.LBService_ERROR_LOG_LEVEL_INFO,
			},
			"size": {
				Type:         schema.TypeString,
				Description:  "Load Balancer Service size",
				ValidateFunc: validation.StringInSlice(lBServiceSizeValues, false),
				Optional:     true,
				Default:      model.LBService_SIZE_SMALL,
			},
		},
	}
}

func resourceNsxtPolicyLBServiceExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	client := infra.NewLbServicesClient(connector)

	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyLBServiceCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbServicesClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLBServiceExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	connectivityPath := d.Get("connectivity_path").(string)
	enabled := d.Get("enabled").(bool)
	errorLogLevel := d.Get("error_log_level").(string)
	size := d.Get("size").(string)
	if size == "XLARGE" && util.NsxVersionLower("3.0.0") {
		return fmt.Errorf("XLARGE size is not supported before NSX version 3.0.0")
	}

	obj := model.LBService{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		ConnectivityPath: &connectivityPath,
		Enabled:          &enabled,
		ErrorLogLevel:    &errorLogLevel,
		Size:             &size,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating LBService with ID %s", id)
	err = client.Patch(id, obj)
	if err != nil {
		return handleCreateError("LBService", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLBServiceRead(d, m)
}

func resourceNsxtPolicyLBServiceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbServicesClient(connector)

	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBService ID")
	}

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "LBService", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("connectivity_path", obj.ConnectivityPath)
	d.Set("enabled", obj.Enabled)
	d.Set("error_log_level", obj.ErrorLogLevel)
	d.Set("size", obj.Size)

	return nil
}

func resourceNsxtPolicyLBServiceUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewLbServicesClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBService ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	connectivityPath := d.Get("connectivity_path").(string)
	enabled := d.Get("enabled").(bool)
	errorLogLevel := d.Get("error_log_level").(string)
	size := d.Get("size").(string)
	if size == "XLARGE" && util.NsxVersionLower("3.0.0") {
		return fmt.Errorf("XLARGE size is not supported before NSX version 3.0.0")
	}

	obj := model.LBService{
		DisplayName:      &displayName,
		Description:      &description,
		Tags:             tags,
		ConnectivityPath: &connectivityPath,
		Enabled:          &enabled,
		ErrorLogLevel:    &errorLogLevel,
		Size:             &size,
	}

	// Update the resource using PATCH
	err := client.Patch(id, obj)
	if err != nil {
		return handleUpdateError("LBService", id, err)
	}

	return resourceNsxtPolicyLBServiceRead(d, m)
}

func resourceNsxtPolicyLBServiceDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining LBService ID")
	}

	connector := getPolicyConnector(m)
	client := infra.NewLbServicesClient(connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	force := true
	err := client.Delete(id, &force)
	if err != nil {
		return handleDeleteError("LBService", id, err)
	}

	return nil
}
