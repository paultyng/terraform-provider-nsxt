// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade"
)

func resourceNsxtUpgradePrecheckAcknowledge() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtUpgradePrecheckAcknowledgeCreate,
		Read:   resourceNsxtUpgradePrecheckAcknowledgeRead,
		Update: resourceNsxtUpgradePrecheckAcknowledgeUpdate,
		Delete: resourceNsxtUpgradePrecheckAcknowledgeDelete,

		Schema: map[string]*schema.Schema{
			"precheck_ids": {
				Type:        schema.TypeList,
				Description: "IDs of precheck warnings that need to be acknowledged",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required: true,
			},
			"precheck_warnings": {
				Type:        schema.TypeList,
				Description: "List of warnings from precheck",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "ID of the precheck warning",
							Computed:    true,
						},
						"message": {
							Type:        schema.TypeString,
							Description: "Message of the precheck warning",
							Computed:    true,
						},
						"is_acknowledged": {
							Type:        schema.TypeBool,
							Description: "Boolean value which identifies if the precheck warning has been acknowledged.",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
			"target_version": {
				Type:        schema.TypeString,
				Description: "Target system version",
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceNsxtUpgradePrecheckAcknowledgeCreate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		id = newUUID()
	}
	d.SetId(id)
	precheckIDs := interface2StringList(d.Get("precheck_ids").([]interface{}))
	err := acknowledgePrecheckWarnings(m, precheckIDs)
	if err != nil {
		return handleCreateError("NsxtPrecheckAcknowledge", id, err)
	}
	return resourceNsxtUpgradePrecheckAcknowledgeRead(d, m)
}

func acknowledgePrecheckWarnings(m interface{}, precheckIDs []string) error {
	connector := getPolicyConnector(m)
	client := upgrade.NewPreUpgradeChecksClient(connector)
	for _, precheckID := range precheckIDs {
		err := client.Acknowledge(precheckID)
		if err != nil {
			msg := fmt.Sprintf("Failed to acknowledge precheck warning with ID %s", precheckID)
			return logAPIError(msg, err)
		}
	}
	return nil
}

func resourceNsxtUpgradePrecheckAcknowledgeRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	typeParam := nsxModel.UpgradeCheckFailure_TYPE_WARNING
	precheckWarnings, err := getPrecheckErrors(m, &typeParam)
	if err != nil {
		return handleReadError(d, "NsxtUpgradePrecheckAcknowledge", id, err)
	}
	err = setPrecheckWarningsInSchema(d, precheckWarnings)
	if err != nil {
		return handleReadError(d, "NsxtUpgradePrecheckAcknowledge", id, err)
	}
	return nil
}

func setPrecheckWarningsInSchema(d *schema.ResourceData, precheckWarnings []nsxModel.UpgradeCheckFailure) error {
	var precheckWarningList []map[string]interface{}
	for _, precheckWarning := range precheckWarnings {
		elem := make(map[string]interface{})
		elem["id"] = precheckWarning.Id
		elem["message"] = precheckWarning.Message.Message
		elem["is_acknowledged"] = precheckWarning.Acked
		precheckWarningList = append(precheckWarningList, elem)
	}
	return d.Set("precheck_warnings", precheckWarningList)
}

func resourceNsxtUpgradePrecheckAcknowledgeUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	precheckIDs := interface2StringList(d.Get("precheck_ids").([]interface{}))
	err := acknowledgePrecheckWarnings(m, precheckIDs)
	if err != nil {
		return handleUpdateError("NsxtPrecheckAcknowledge", id, err)
	}
	return resourceNsxtUpgradePrecheckAcknowledgeRead(d, m)
}

func resourceNsxtUpgradePrecheckAcknowledgeDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func getAcknowledgedPrecheckIDs(m interface{}) ([]string, error) {
	var result []string
	typeParam := nsxModel.UpgradeCheckFailure_TYPE_WARNING
	precheckWarnings, err := getPrecheckErrors(m, &typeParam)
	if err != nil {
		return result, err
	}
	for _, warning := range precheckWarnings {
		acked := *warning.Acked
		if acked {
			result = append(result, *warning.Id)
		}
	}
	return result, nil
}
