/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sha/runbook_invocations"
)

func dataSourceNsxtPolicyODSRunbookInvocationReport() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtPolicyODSPRunbookInvocationReportRead,

		Schema: map[string]*schema.Schema{
			"invocation_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UUID of runbook invocation",
			},
			"target_node": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Identifier of an appliance node or transport node",
			},
			"error_detail": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "The report error detail",
			},
			"invalid_reason": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Invalid report reason",
			},
			"recommendation_code": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "Online Diagnostic System recommendation code",
			},
			"recommendation_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Online Diagnostic System recommendation message",
			},
			"result_code": {
				Type:        schema.TypeInt,
				Computed:    true,
				Optional:    true,
				Description: "Online Diagnostic System result code",
			},
			"result_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Online Diagnostic System result message",
			},
			"request_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Request status of a runbook invocation",
			},
			"operation_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Operation state of a runbook invocation on the target node",
			},
		},
	}
}

func dataSourceNsxtPolicyODSPRunbookInvocationReportRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	invocationID := d.Get("invocation_id").(string)
	client := runbook_invocations.NewReportClient(connector)

	obj, err := client.Get(invocationID)
	if err != nil {
		return handleDataSourceReadError(d, "OdsRunbookInvocationReport", invocationID, err)
	}

	d.SetId(invocationID)
	d.Set("target_node", obj.TargetNode)
	d.Set("error_detail", obj.ErrorDetail)
	d.Set("invalid_reason", obj.InvalidReason)
	d.Set("recommendation_code", obj.RecommendationCode)
	d.Set("recommendation_message", obj.RecommendationMessage)
	d.Set("result_code", obj.ResultCode)
	d.Set("result_message", obj.ResultMessage)
	if obj.Status != nil {
		d.Set("request_status", obj.Status.RequestStatus)
		d.Set("operation_state", obj.Status.OperationState)
	}

	return nil
}
