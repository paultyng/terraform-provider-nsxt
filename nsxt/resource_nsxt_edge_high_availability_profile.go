// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func resourceNsxtEdgeHighAvailabilityProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtEdgeHighAvailabilityProfileCreate,
		Read:   resourceNsxtEdgeHighAvailabilityProfileRead,
		Update: resourceNsxtEdgeHighAvailabilityProfileUpdate,
		Delete: resourceNsxtEdgeHighAvailabilityProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"revision":     getRevisionSchema(),
			"description":  getDescriptionSchema(),
			"display_name": getDisplayNameSchema(),
			"tag":          getTagsSchema(),
			"bfd_allowed_hops": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      255,
				Description:  "BFD allowed hops",
				ValidateFunc: validation.IntBetween(1, 255),
			},
			"bfd_declare_dead_multiple": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      3,
				Description:  "Number of times a packet is missed before BFD declares the neighbor down",
				ValidateFunc: validation.IntBetween(2, 16),
			},
			"bfd_probe_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      500,
				Description:  "the time interval (in millisecond) between probe packets for heartbeat purpose",
				ValidateFunc: validation.IntBetween(50, 60000),
			},
			"standby_relocation_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      30,
				Description:  "Standby service context relocation wait time",
				ValidateFunc: validation.IntBetween(10, 20000),
			},
		},
	}
}

func resourceNsxtEdgeHighAvailabilityProfileCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewClusterProfilesClient(connector)
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getMPTagsFromSchema(d)
	bfdAllowedHops := int64(d.Get("bfd_allowed_hops").(int))
	bfdDeclareDeadMultiple := int64(d.Get("bfd_declare_dead_multiple").(int))
	bfdProbeInterval := int64(d.Get("bfd_probe_interval").(int))
	standbyRelocationThreshold := int64(d.Get("standby_relocation_threshold").(int))

	obj := model.EdgeHighAvailabilityProfile{
		Description:            &description,
		DisplayName:            &displayName,
		Tags:                   tags,
		BfdAllowedHops:         &bfdAllowedHops,
		BfdDeclareDeadMultiple: &bfdDeclareDeadMultiple,
		BfdProbeInterval:       &bfdProbeInterval,
		StandbyRelocationConfig: &model.StandbyRelocationConfig{
			StandbyRelocationThreshold: &standbyRelocationThreshold,
		},
		ResourceType: model.EdgeHighAvailabilityProfile__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errs := converter.ConvertToVapi(obj, model.EdgeHighAvailabilityProfileBindingType())
	if errs != nil {
		return errs[0]
	}

	structValue, err := client.Create(dataValue.(*data.StructValue))
	if err != nil {
		return handleCreateError("Edge High Availability Profile", displayName, err)
	}
	o, errs := converter.ConvertToGolang(structValue, model.EdgeHighAvailabilityProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	obj = o.(model.EdgeHighAvailabilityProfile)

	d.SetId(*obj.Id)
	return resourceNsxtEdgeHighAvailabilityProfileRead(d, m)
}

func resourceNsxtEdgeHighAvailabilityProfileRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	client := nsx.NewClusterProfilesClient(connector)
	structValue, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Edge High Availability Profile", id, err)
	}
	converter := bindings.NewTypeConverter()
	o, errs := converter.ConvertToGolang(structValue, model.EdgeHighAvailabilityProfileBindingType())
	if errs != nil {
		return errs[0]
	}
	obj := o.(model.EdgeHighAvailabilityProfile)
	d.Set("revision", obj.Revision)
	d.Set("description", obj.Description)
	d.Set("display_name", obj.DisplayName)
	setMPTagsInSchema(d, obj.Tags)
	d.Set("bfd_allowed_hops", obj.BfdAllowedHops)
	d.Set("bfd_declare_dead_multiple", obj.BfdDeclareDeadMultiple)
	d.Set("bfd_probe_interval", obj.BfdProbeInterval)
	if obj.StandbyRelocationConfig != nil {
		d.Set("standby_relocation_threshold", obj.StandbyRelocationConfig.StandbyRelocationThreshold)
	}

	return nil
}

func resourceNsxtEdgeHighAvailabilityProfileUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	client := nsx.NewClusterProfilesClient(connector)
	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getMPTagsFromSchema(d)
	bfdAllowedHops := int64(d.Get("bfd_allowed_hops").(int))
	bfdDeclareDeadMultiple := int64(d.Get("bfd_declare_dead_multiple").(int))
	bfdProbeInterval := int64(d.Get("bfd_probe_interval").(int))
	standbyRelocationThreshold := int64(d.Get("standby_relocation_threshold").(int))

	obj := model.EdgeHighAvailabilityProfile{
		Revision:               &revision,
		Description:            &description,
		DisplayName:            &displayName,
		Tags:                   tags,
		BfdAllowedHops:         &bfdAllowedHops,
		BfdDeclareDeadMultiple: &bfdDeclareDeadMultiple,
		BfdProbeInterval:       &bfdProbeInterval,
		StandbyRelocationConfig: &model.StandbyRelocationConfig{
			StandbyRelocationThreshold: &standbyRelocationThreshold,
		},
		ResourceType: model.EdgeHighAvailabilityProfile__TYPE_IDENTIFIER,
	}
	converter := bindings.NewTypeConverter()
	dataValue, errs := converter.ConvertToVapi(obj, model.EdgeHighAvailabilityProfileBindingType())
	if errs != nil {
		return errs[0]
	}

	_, err := client.Update(id, dataValue.(*data.StructValue))
	if err != nil {
		return handleUpdateError("Edge High Availability Profile", id, err)
	}

	return resourceNsxtEdgeHighAvailabilityProfileRead(d, m)
}

func resourceNsxtEdgeHighAvailabilityProfileDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := nsx.NewClusterProfilesClient(connector)

	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("Edge High Availability Profile", id, err)
	}
	return nil
}
