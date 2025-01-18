// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

var rateLimiterResourceTypes = []string{model.QosBaseRateLimiter_RESOURCE_TYPE_INGRESSRATELIMITER, model.QosBaseRateLimiter_RESOURCE_TYPE_INGRESSBROADCASTRATELIMITER, model.QosBaseRateLimiter_RESOURCE_TYPE_EGRESSRATELIMITER}

func resourceNsxtPolicyQosProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyQosProfileCreate,
		Read:   resourceNsxtPolicyQosProfileRead,
		Update: resourceNsxtPolicyQosProfileUpdate,
		Delete: resourceNsxtPolicyQosProfileDelete,
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
			"class_of_service": {
				Type:         schema.TypeInt,
				Description:  "Class of service",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 7),
			},
			"dscp_trusted": {
				Type:        schema.TypeBool,
				Description: "Trust mode for DSCP",
				Optional:    true,
				Default:     false,
			},

			"dscp_priority": {
				Type:         schema.TypeInt,
				Description:  "DSCP Priority",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 63),
			},
			"ingress_rate_shaper":           getQosRateShaperSchema(ingressRateShaperIndex),
			"ingress_broadcast_rate_shaper": getQosRateShaperSchema(ingressBroadcastRateShaperIndex),
			"egress_rate_shaper":            getQosRateShaperSchema(egressRateShaperIndex),
		},
	}
}

func resourceNsxtPolicyQosProfileExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {
	client := infra.NewQosProfilesClient(sessionContext, connector)
	if client == nil {
		return false, policyResourceNotSupportedError()
	}
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func getPolicyQosRateShaperFromSchema(d *schema.ResourceData, index int) *data.StructValue {
	scale := rateShaperScales[index]
	schemaName := rateShaperSchemaNames[index]
	resourceType := rateLimiterResourceTypes[index]
	shaperConfs := d.Get(schemaName).([]interface{})
	for _, shaperConf := range shaperConfs {
		// only 1 is allowed for each type
		shaperData := shaperConf.(map[string]interface{})
		enabled := shaperData["enabled"].(bool)
		averageBW := int64(shaperData[fmt.Sprintf("average_bw_%s", scale)].(int))
		burstSize := int64(shaperData["burst_size"].(int))
		peakBW := int64(shaperData[fmt.Sprintf("peak_bw_%s", scale)].(int))

		shaper := model.IngressRateLimiter{
			ResourceType:     resourceType,
			Enabled:          &enabled,
			BurstSize:        &burstSize,
			AverageBandwidth: &averageBW,
			PeakBandwidth:    &peakBW,
		}
		converter := bindings.NewTypeConverter()
		dataValue, _ := converter.ConvertToVapi(shaper, model.IngressRateLimiterBindingType())
		return dataValue.(*data.StructValue)
	}

	return nil
}

func setPolicyQosRateShaperInSchema(d *schema.ResourceData, shaperConf []*data.StructValue, index int) {
	scale := rateShaperScales[index]
	schemaName := rateShaperSchemaNames[index]
	resourceType := rateLimiterResourceTypes[index]
	var shapers []map[string]interface{}
	converter := bindings.NewTypeConverter()
	for _, dataShaper := range shaperConf {
		dataValue, _ := converter.ConvertToGolang(dataShaper, model.IngressRateLimiterBindingType())
		shaper := dataValue.(model.IngressRateLimiter)
		if shaper.ResourceType == resourceType {
			elem := make(map[string]interface{})
			elem["enabled"] = shaper.Enabled
			elem["burst_size"] = shaper.BurstSize
			elem[fmt.Sprintf("average_bw_%s", scale)] = shaper.AverageBandwidth
			elem[fmt.Sprintf("peak_bw_%s", scale)] = shaper.PeakBandwidth

			shapers = append(shapers, elem)
			err := d.Set(schemaName, shapers)
			if err != nil {
				log.Printf("[WARNING] Failed to set shapers in schema: %v", err)
			}
		}
	}
}

func resourceNsxtPolicyQosProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID2(d, m, resourceNsxtPolicyQosProfileExists)
	if err != nil {
		return err
	}

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	classOfService := int64(d.Get("class_of_service").(int))
	dscpTrusted := "UNTRUSTED"
	if d.Get("dscp_trusted").(bool) {
		dscpTrusted = "TRUSTED"
	}
	dscpPriority := int64(d.Get("dscp_priority").(int))
	var shapers []*data.StructValue
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {

		shaper := getPolicyQosRateShaperFromSchema(d, index)
		if shaper != nil {
			shapers = append(shapers, shaper)
		}
	}

	obj := model.QosProfile{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		ClassOfService: &classOfService,
		Dscp: &model.QosDscp{
			Mode:     &dscpTrusted,
			Priority: &dscpPriority,
		},
		ShaperConfigurations: shapers,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating QosProfile with ID %s", id)
	boolFalse := false
	client := infra.NewQosProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err = client.Patch(id, obj, &boolFalse)
	if err != nil {
		return handleCreateError("QosProfile", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyQosProfileRead(d, m)
}

func resourceNsxtPolicyQosProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining QosProfile ID")
	}

	client := infra.NewQosProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "QosProfile", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("class_of_service", obj.ClassOfService)
	if *obj.Dscp.Mode == "TRUSTED" {
		d.Set("dscp_trusted", true)
	} else {
		d.Set("dscp_trusted", false)
	}
	d.Set("dscp_priority", obj.Dscp.Priority)
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {
		setPolicyQosRateShaperInSchema(d, obj.ShaperConfigurations, index)
	}

	return nil
}

func resourceNsxtPolicyQosProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := infra.NewQosProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining QosProfile ID")
	}

	// Read the rest of the configured parameters
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)

	classOfService := int64(d.Get("class_of_service").(int))
	dscpTrusted := "UNTRUSTED"
	if d.Get("dscp_trusted").(bool) {
		dscpTrusted = "TRUSTED"
	}
	dscpPriority := int64(d.Get("dscp_priority").(int))
	var shapers []*data.StructValue
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {

		shaper := getPolicyQosRateShaperFromSchema(d, index)
		if shaper != nil {
			shapers = append(shapers, shaper)
		}
	}

	obj := model.QosProfile{
		DisplayName:    &displayName,
		Description:    &description,
		Tags:           tags,
		ClassOfService: &classOfService,
		Dscp: &model.QosDscp{
			Mode:     &dscpTrusted,
			Priority: &dscpPriority,
		},
		ShaperConfigurations: shapers,
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Updating QosProfile with ID %s", id)
	boolFalse := false
	err := client.Patch(id, obj, &boolFalse)
	if err != nil {
		d.Partial(true)
		return handleUpdateError("QosProfile", id, err)
	}

	return resourceNsxtPolicyQosProfileRead(d, m)
}

func resourceNsxtPolicyQosProfileDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining QosProfile ID")
	}

	connector := getPolicyConnector(m)
	boolFalse := false
	client := infra.NewQosProfilesClient(getSessionContext(d, m), connector)
	if client == nil {
		return policyResourceNotSupportedError()
	}
	err := client.Delete(id, &boolFalse)

	if err != nil {
		return handleDeleteError("QosProfile", id, err)
	}

	return nil
}
