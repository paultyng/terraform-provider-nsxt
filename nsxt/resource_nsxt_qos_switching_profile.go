// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var ingressRateShaperIndex = 0
var ingressBroadcastRateShaperIndex = 1
var egressRateShaperIndex = 2

var rateShaperSchemaNames = []string{"ingress_rate_shaper", "ingress_broadcast_rate_shaper", "egress_rate_shaper"}
var rateShaperScales = []string{"mbps", "kbps", "mbps"}
var rateShaperResourceTypes = []string{"IngressRateShaper", "IngressBroadcastRateShaper", "EgressRateShaper"}

func resourceNsxtQosSwitchingProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtQosSwitchingProfileCreate,
		Read:   resourceNsxtQosSwitchingProfileRead,
		Update: resourceNsxtQosSwitchingProfileUpdate,
		Delete: resourceNsxtQosSwitchingProfileDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		DeprecationMessage: mpObjectResourceDeprecationMessage,
		Schema: map[string]*schema.Schema{
			"revision": getRevisionSchema(),
			"description": {
				Type:        schema.TypeString,
				Description: "Description of this resource",
				Optional:    true,
			},
			"display_name": {
				Type:        schema.TypeString,
				Description: "The display name of this resource. Defaults to ID if not set",
				Optional:    true,
				Computed:    true,
			},
			"tag": getTagsSchema(),
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

func getQosRateShaperSchema(index int) *schema.Schema {
	scale := rateShaperScales[index]
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Type:        schema.TypeBool,
					Description: "Whether this rate shaper is enabled",
					Optional:    true,
					Default:     true,
				},
				fmt.Sprintf("average_bw_%s", scale): {
					Type:        schema.TypeInt,
					Description: fmt.Sprintf("Average Bandwidth in %s", scale),
					Optional:    true,
				},
				"burst_size": {
					Type:        schema.TypeInt,
					Description: "Burst size in bytes",
					Optional:    true,
				},
				fmt.Sprintf("peak_bw_%s", scale): {
					Type:        schema.TypeInt,
					Description: fmt.Sprintf("Peak Bandwidth in %s", scale),
					Optional:    true,
				},
			},
		},
	}
}

func getQosRateShaperFromSchema(d *schema.ResourceData, index int) *manager.QosBaseRateShaper {
	scale := rateShaperScales[index]
	schemaName := rateShaperSchemaNames[index]
	resourceType := rateShaperResourceTypes[index]
	shaperConfs := d.Get(schemaName).([]interface{})
	for _, shaperConf := range shaperConfs {
		// only 1 is allowed for each type
		data := shaperConf.(map[string]interface{})
		enabled := data["enabled"].(bool)
		averageBW := int32(data[fmt.Sprintf("average_bw_%s", scale)].(int))
		burstSize := int32(data["burst_size"].(int))
		peakBW := int32(data[fmt.Sprintf("peak_bw_%s", scale)].(int))

		shaper := manager.QosBaseRateShaper{
			ResourceType:   resourceType,
			Enabled:        enabled,
			BurstSizeBytes: burstSize,
		}

		if scale == "mbps" {
			shaper.AverageBandwidthMbps = averageBW
			shaper.PeakBandwidthMbps = peakBW
		} else {
			shaper.AverageBandwidthKbps = averageBW
			shaper.PeakBandwidthKbps = peakBW
		}

		return &shaper

	}

	return nil
}

func setQosRateShaperInSchema(d *schema.ResourceData, shaperConf []manager.QosBaseRateShaper, index int) {
	scale := rateShaperScales[index]
	schemaName := rateShaperSchemaNames[index]
	resourceType := rateShaperResourceTypes[index]
	var shapers []map[string]interface{}
	for _, shaper := range shaperConf {
		if shaper.ResourceType == resourceType {
			average := shaper.AverageBandwidthKbps
			peak := shaper.PeakBandwidthKbps
			if scale == "mbps" {
				average = shaper.AverageBandwidthMbps
				peak = shaper.PeakBandwidthMbps
			}
			// Do not define schema for default shaper to avoid non-empty plan
			if !shaper.Enabled && (average+shaper.BurstSizeBytes+peak == 0) {
				return
			}
			elem := make(map[string]interface{})
			elem["enabled"] = shaper.Enabled
			elem["burst_size"] = shaper.BurstSizeBytes

			elem[fmt.Sprintf("average_bw_%s", scale)] = average
			elem[fmt.Sprintf("peak_bw_%s", scale)] = peak

			shapers = append(shapers, elem)
			err := d.Set(schemaName, shapers)
			if err != nil {
				log.Printf("[WARNING] Failed to set shapers in schema: %v", err)
			}
		}
	}

}

func resourceNsxtQosSwitchingProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	classOfService := int32(d.Get("class_of_service").(int))
	dscpTrusted := "UNTRUSTED"
	if d.Get("dscp_trusted").(bool) {
		dscpTrusted = "TRUSTED"
	}
	dscpPriority := int32(d.Get("dscp_priority").(int))
	shapers := []manager.QosBaseRateShaper{}
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {

		shaper := getQosRateShaperFromSchema(d, index)
		if shaper != nil {
			shapers = append(shapers, *shaper)
		}
	}

	qosSwitchingProfile := manager.QosSwitchingProfile{
		Description:    description,
		DisplayName:    displayName,
		Tags:           tags,
		ClassOfService: classOfService,
		Dscp: &manager.Dscp{
			Mode:     dscpTrusted,
			Priority: dscpPriority,
		},
		ShaperConfiguration: shapers,
	}

	qosSwitchingProfile, resp, err := nsxClient.LogicalSwitchingApi.CreateQosSwitchingProfile(nsxClient.Context, qosSwitchingProfile)

	if err != nil {
		return fmt.Errorf("Error during QosSwitchingProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned during QosSwitchingProfile create: %v", resp.StatusCode)
	}
	d.SetId(qosSwitchingProfile.Id)

	return resourceNsxtQosSwitchingProfileRead(d, m)
}

func resourceNsxtQosSwitchingProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	qosSwitchingProfile, resp, err := nsxClient.LogicalSwitchingApi.GetQosSwitchingProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] QosSwitchingProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during QosSwitchingProfile read: %v", err)
	}

	d.Set("revision", qosSwitchingProfile.Revision)
	d.Set("description", qosSwitchingProfile.Description)
	d.Set("display_name", qosSwitchingProfile.DisplayName)
	setTagsInSchema(d, qosSwitchingProfile.Tags)
	d.Set("class_of_service", qosSwitchingProfile.ClassOfService)
	if qosSwitchingProfile.Dscp.Mode == "TRUSTED" {
		d.Set("dscp_trusted", true)
	} else {
		d.Set("dscp_trusted", false)
	}
	d.Set("dscp_priority", qosSwitchingProfile.Dscp.Priority)
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {
		setQosRateShaperInSchema(d, qosSwitchingProfile.ShaperConfiguration, index)
	}

	return nil
}

func resourceNsxtQosSwitchingProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	revision := int64(d.Get("revision").(int))
	classOfService := int32(d.Get("class_of_service").(int))
	dscpTrusted := "UNTRUSTED"
	if d.Get("dscp_trusted").(bool) {
		dscpTrusted = "TRUSTED"
	}
	dscpPriority := int32(d.Get("dscp_priority").(int))
	shapers := []manager.QosBaseRateShaper{}
	for index := ingressRateShaperIndex; index <= egressRateShaperIndex; index++ {
		shaper := getQosRateShaperFromSchema(d, index)
		if shaper != nil {
			shapers = append(shapers, *shaper)
		}
	}

	qosSwitchingProfile := manager.QosSwitchingProfile{
		Description:    description,
		DisplayName:    displayName,
		Tags:           tags,
		ClassOfService: classOfService,
		Dscp: &manager.Dscp{
			Mode:     dscpTrusted,
			Priority: dscpPriority,
		},
		ShaperConfiguration: shapers,
		Revision:            revision,
	}

	_, resp, err := nsxClient.LogicalSwitchingApi.UpdateQosSwitchingProfile(nsxClient.Context, id, qosSwitchingProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during QosSwitchingProfile update: %v", err)
	}

	return resourceNsxtQosSwitchingProfileRead(d, m)
}

func resourceNsxtQosSwitchingProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.LogicalSwitchingApi.DeleteSwitchingProfile(nsxClient.Context, id, nil)
	if err != nil {
		return fmt.Errorf("Error during QosSwitchingProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] QosSwitchingProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
