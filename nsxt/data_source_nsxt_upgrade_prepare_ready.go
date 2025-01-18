// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade"
	"golang.org/x/exp/slices"
)

func dataSourceNsxtUpgradePrepareReady() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceNsxtUpgradePrepareReadyRead,

		Schema: map[string]*schema.Schema{
			"id": getDataSourceIDSchema(),
			"upgrade_prepare_id": {
				Type:        schema.TypeString,
				Description: "ID of corresponding nsxt_upgrade_prepare resource",
				Required:    true,
			},
		},
	}
}

func getPrechecksText(m interface{}, precheckIDs []string) (string, error) {
	precheckText := ""
	connector := getPolicyConnector(m)
	client := upgrade.NewUpgradeChecksInfoClient(connector)
	checkInfoResults, err := client.List(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return "", err
	}
	for _, checkInfo := range checkInfoResults.Results {
		for _, ci := range checkInfo.PreUpgradeChecksInfo {
			if slices.Contains(precheckIDs, *ci.Id) {
				precheckText += fmt.Sprintf("  Component: %s, code: %s, description: %s\n", *checkInfo.ComponentType, *ci.Id, *ci.Description)
			}
		}
	}
	return precheckText, nil
}

func dataSourceNsxtUpgradePrepareReadyRead(d *schema.ResourceData, m interface{}) error {
	// Validate that upgrade_prepare_id is actually from the nsxt_upgrade_prepare resource
	upgradePrepareID := d.Get("upgrade_prepare_id").(string)
	if !util.VerifyVerifiableID(upgradePrepareID, "nsxt_upgrade_prepare") {
		return fmt.Errorf("value for upgrade_prepare_id is invalid: %s", upgradePrepareID)
	}
	precheckErrors, err := getPrecheckErrors(m, nil)
	if err != nil {
		return fmt.Errorf("Error while reading precheck failures: %v", err)
	}
	var precheckFailureIDs []string
	var unacknowledgedWarningIDs []string
	for _, precheckError := range precheckErrors {
		errType := *precheckError.Type_
		if errType == nsxModel.UpgradeCheckFailure_TYPE_FAILURE {
			precheckFailureIDs = append(precheckFailureIDs, *precheckError.Id)
		} else {
			if !(*precheckError.Acked) {
				unacknowledgedWarningIDs = append(unacknowledgedWarningIDs, *precheckError.Id)
			}
		}
	}
	var errMessage string
	if len(precheckFailureIDs) > 0 {
		preCheckText, err := getPrechecksText(m, precheckFailureIDs)
		if err != nil {
			errMessage += fmt.Sprintf("Error while reading precheck failures text: %v", err)
		}
		errMessage += fmt.Sprintf("There are failures in prechecks:\n%s\nPlease check their status from nsxt_upgrade_prepare resource and address these failures on NSX", preCheckText)
	}
	if len(unacknowledgedWarningIDs) > 0 {
		preCheckText, err := getPrechecksText(m, unacknowledgedWarningIDs)
		if err != nil {
			errMessage += fmt.Sprintf("Error while reading precheck failures text: %v", err)
		}
		errMessage += fmt.Sprintf("\nThere are unacknowledged warnings in prechecks:\n%s\nPlease address these errors from NSX or using nsxt_upgrade_precheck_acknowledge resource", preCheckText)
	}
	if len(errMessage) > 0 {
		return errors.New(errMessage)
	}
	objID := util.GetVerifiableID(newUUID(), "nsxt_upgrade_prepare_ready")
	d.SetId(objID)

	return nil
}
