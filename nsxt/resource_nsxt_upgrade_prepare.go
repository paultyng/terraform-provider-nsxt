// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	liberrors "errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade/bundles"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade/eula"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/upgrade/pre_upgrade_checks"
)

var precheckComponentTypes = []string{"EDGE", "HOST", "MP"}

const bundleUploadTimeout int = 3600
const ucUpgradeTimeout int = 3600
const precheckTimeout int = 3600

func resourceNsxtUpgradePrepare() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtUpgradePrepareCreate,
		Read:   resourceNsxtUpgradePrepareRead,
		Update: resourceNsxtUpgradePrepareUpdate,
		Delete: resourceNsxtUpgradePrepareDelete,

		Schema: map[string]*schema.Schema{
			"version": {
				Type:        schema.TypeString,
				Description: "Target upgrade version for NSX, format is x.x.x..., should include at least 3 digits, example: 4.1.2",
				Optional:    true,
			},
			"upgrade_bundle_url": {
				Type:         schema.TypeString,
				Description:  "URL of the NSXT Upgrade bundle",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"precheck_bundle_url": {
				Type:        schema.TypeString,
				Description: "URL of the NSXT Upgrade precheck bundle (Only applied to NSXT version >= 4.1.1)",
				Optional:    true,
			},
			"accept_user_agreement": {
				Type:        schema.TypeBool,
				Description: "Whether to accept the user agreement",
				Required:    true,
			},
			"bundle_upload_timeout": {
				Type:        schema.TypeInt,
				Description: "Timeout for uploading bundle in seconds",
				Optional:    true,
				Default:     bundleUploadTimeout,
			},
			"uc_upgrade_timeout": {
				Type:        schema.TypeInt,
				Description: "Timeout for upgrading upgrade coordinator in seconds",
				Optional:    true,
				Default:     ucUpgradeTimeout,
			},
			"precheck_timeout": {
				Type:        schema.TypeInt,
				Description: "Timeout for executing pre-upgrade checks in seconds",
				Optional:    true,
				Default:     precheckTimeout,
			},
			"failed_prechecks": {
				Type:        schema.TypeList,
				Description: "List of failed prechecks for the upgrade, including both warnings and errors",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Description: "Type of the precheck failure (warning or error)",
							Computed:    true,
						},
						"id": {
							Type:        schema.TypeString,
							Description: "ID of failed precheck",
							Computed:    true,
						},
						"message": {
							Type:        schema.TypeString,
							Description: "Message of the failed precheck",
							Computed:    true,
						},
						"needs_ack": {
							Type:        schema.TypeBool,
							Description: "Boolean value which identifies if acknowledgement is required for the precheck",
							Computed:    true,
						},
						"needs_resolve": {
							Type:        schema.TypeBool,
							Description: "Boolean value identifies if resolution is required for the precheck",
							Computed:    true,
						},
						"acked": {
							Type:        schema.TypeBool,
							Description: "Boolean value which identifies if precheck has been acknowledged",
							Computed:    true,
						},
						"resolution_status": {
							Type:        schema.TypeString,
							Description: "The resolution status of the precheck failure",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},
			"target_version": {
				Type:        schema.TypeString,
				Description: "Target system version",
				Computed:    true,
			},
		},
	}
}

func resourceNsxtUpgradePrepareCreate(d *schema.ResourceData, m interface{}) error {
	id := util.GetVerifiableID(newUUID(), "nsxt_upgrade_prepare")
	err := prepareForUpgrade(d, m)
	if err != nil {
		return handleCreateError("NsxtUpgradePrepare", id, err)
	}
	d.SetId(id)
	return resourceNsxtUpgradePrepareRead(d, m)
}

func prepareForUpgrade(d *schema.ResourceData, m interface{}) error {
	// 1. Upload upgrade bundle and wait for upload to complete
	err := uploadPrecheckAndUpgradeBundle(d, m)
	if err != nil {
		return logAPIError("Failed to upload bundle", err)
	}
	// 2. Accept eula
	err = acceptUserAgreement(d, m)
	if err != nil {
		return err
	}
	// 3. Upgrade UC and check for its upgrade status
	err = upgradeUc(d, m)
	if err != nil {
		return logAPIError("Failed to upgrade Upgrade Coordinator", err)
	}
	return nil
}

func isVCF9HostUpgrade(m interface{}, targetVersion string) (bool, error) {
	// In VCF9.0 and newer, overall upgrade status can be PAUSED when hosts are pre-upgraded. Pre-checks should still
	// execute in this case
	if util.VersionLower(targetVersion, "9.0.0") {
		return false, nil
	}
	connector := getPolicyConnector(m)
	client := upgrade.NewStatusSummaryClient(connector)
	statusSummary, err := client.Get(nil, nil, nil)
	if err != nil {
		return false, err
	}
	if *statusSummary.OverallUpgradeStatus != nsxModel.UpgradeStatus_OVERALL_UPGRADE_STATUS_PAUSED {
		return false, nil
	}
	// In this case, all components other than host will be NOT_STARTED, but hosts will be with status SUCCESS
	// Finalize component can be NOT_STARTED or PAUSED
	for _, c := range statusSummary.ComponentStatus {
		if *c.ComponentType == hostUpgradeGroup {
			if *c.Status != nsxModel.ComponentUpgradeStatus_STATUS_SUCCESS {
				return false, nil
			}
		} else if *c.ComponentType == finalizeUpgradeGroup {
			if *c.Status != nsxModel.ComponentUpgradeStatus_STATUS_PAUSED && *c.Status != nsxModel.ComponentUpgradeStatus_STATUS_NOT_STARTED {
				return false, nil
			}
		} else {
			// It's not a host - should be NOT_STARTED
			if *c.Status != nsxModel.ComponentUpgradeStatus_STATUS_NOT_STARTED {
				return false, nil
			}
		}
	}
	return true, nil
}

func getSummaryInfo(m interface{}) (string, bool, error) {
	connector := getPolicyConnector(m)
	summaryClient := upgrade.NewSummaryClient(connector)
	summary, err := summaryClient.Get()
	if err != nil {
		return "", false, err
	}
	var targetVersion string
	if summary.TargetVersion != nil {
		targetVersion = *summary.TargetVersion
	}
	if summary.UpgradeCoordinatorUpdated == nil || !(*summary.UpgradeCoordinatorUpdated) {
		log.Printf("Upgrade coordinated is not upgraded, skip running precheck")
		return targetVersion, false, nil
	}
	if summary.UpgradeStatus == nil || (*summary.UpgradeStatus) != nsxModel.UpgradeSummary_UPGRADE_STATUS_NOT_STARTED {
		is9, err := isVCF9HostUpgrade(m, targetVersion)
		if err != nil {
			return "", false, err
		}
		if is9 {
			log.Printf("Hosts have been pre-upgraded, prechecks should be running")
		} else {
			log.Printf("Upgrade process has started, skip running precheck")
			return targetVersion, false, nil
		}
	}
	return targetVersion, true, nil
}

func resourceNsxtUpgradePrepareRead(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	var err error
	// Execute precheck in Read function if upload bundle has been uploaded and upgrade not started
	targetVersion, precheckNeeded, err := getSummaryInfo(m)
	if err != nil {
		return logAPIError("Failed to get previous precheck result", err)
	}
	d.Set("target_version", targetVersion)
	if precheckNeeded {
		previousAcknowledgedPrecheckIDs, err := getAcknowledgedPrecheckIDs(m)
		if err != nil {
			return logAPIError("Failed to get previous precheck result", err)
		}
		err = executePreupgradeChecks(d, m)
		if err != nil {
			return logAPIError("Failed to execute pre-upgrade checks", err)
		}
		err = acknowledgePrecheckWarnings(m, previousAcknowledgedPrecheckIDs)
		if err != nil {
			return err
		}
	}
	precheckFailures, err := getPrecheckErrors(m, nil)
	if err != nil {
		return handleReadError(d, "NsxtUpgradePrepare", id, err)
	}
	err = setFailedPrechecksInSchema(d, precheckFailures)
	if err != nil {
		return handleReadError(d, "NsxtUpgradePrepare", id, err)
	}
	return nil
}

func resourceNsxtUpgradePrepareUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	err := prepareForUpgrade(d, m)
	if err != nil {
		return handleUpdateError("NsxtUpgradePrepare", id, err)
	}
	return resourceNsxtUpgradePrepareRead(d, m)
}

func resourceNsxtUpgradePrepareDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}

func uploadPrecheckAndUpgradeBundle(d *schema.ResourceData, m interface{}) error {
	upgradeBundleType := nsxModel.UpgradeBundleFetchRequest_BUNDLE_TYPE_UPGRADE
	precheckBundleType := nsxModel.UpgradeBundleFetchRequest_BUNDLE_TYPE_PRE_UPGRADE
	precheckBundleURL := d.Get("precheck_bundle_url").(string)
	if !precheckBundleCompatibilityCheck(precheckBundleURL) {
		return fmt.Errorf("Precheck bundle is only supported for NSXT version >= 4.1.1")
	}
	if len(precheckBundleURL) > 0 {
		err := uploadUpgradeBundle(d, m, precheckBundleType)
		if err != nil {
			return fmt.Errorf("Failed to upload precheck bundle: %s", err)
		}
	}
	err := uploadUpgradeBundle(d, m, upgradeBundleType)
	if err != nil {
		return fmt.Errorf("Failed to upload upgrade bundle: %s", err)
	}
	return nil
}

func precheckBundleCompatibilityCheck(precheckBundleURL string) bool {
	if util.NsxVersionLower("4.1.1") && len(precheckBundleURL) > 0 {
		return false
	}
	return true
}

func uploadUpgradeBundle(d *schema.ResourceData, m interface{}, bundleType string) error {
	upgradeBundleURL := d.Get("upgrade_bundle_url").(string)
	precheckBundleURL := d.Get("precheck_bundle_url").(string)
	var url string
	timeout := d.Get("bundle_upload_timeout").(int)
	c := m.(nsxtClients)
	userName := c.NsxtClientConfig.UserName
	password := c.NsxtClientConfig.Password
	connector := getPolicyConnector(m)
	summaryClient := upgrade.NewSummaryClient(connector)
	summary, err := summaryClient.Get()
	if err != nil {
		return err
	}
	if bundleType == nsxModel.UpgradeBundleFetchRequest_BUNDLE_TYPE_UPGRADE {
		if !d.HasChange("upgrade_bundle_url") {
			return nil
		}
		url = upgradeBundleURL
	} else {
		// Check if same precheck bundle has already been uploaded by comparing input target version
		// with the precheck bundler version in upgrade summary
		// version format is x.x.x...., contains at least 3 numbers
		version := d.Get("version").(string)
		if summary.PreUpgradeBundleVersion != nil && version != "" {
			existingBundleVersion := *summary.PreUpgradeBundleVersion
			arr := strings.Split(version, ".")
			if len(arr) < 3 {
				log.Printf("Invalid input version format, cannot compare version, input version format must be x.x.x....")
			} else {
				if strings.HasPrefix(existingBundleVersion, version) {
					log.Printf("Precheck bundle already uploaded to NSX, skip uploading precheck bundle")
					return nil
				}
			}
		}
		url = precheckBundleURL
	}
	client := upgrade.NewBundlesClient(connector)
	bundleFetchRequest := nsxModel.UpgradeBundleFetchRequest{
		Url: &url,
	}
	if util.NsxVersionHigherOrEqual("4.1.1") {
		bundleFetchRequest.BundleType = &bundleType
		bundleFetchRequest.Password = &password
		bundleFetchRequest.Username = &userName
	}
	bundleID, err := client.Create(bundleFetchRequest, nil)
	if err != nil {
		return fmt.Errorf("Failed to upload upgrade bundle of type %s: %v", bundleType, err)
	} else if bundleID.BundleId == nil {
		return fmt.Errorf("Failed to upload upgrade bundle of type %s: bundle is apparently invalid", bundleType)
	}
	return waitForBundleUpload(m, *bundleID.BundleId, timeout)
}

func acceptUserAgreement(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	acceptUserAgreement := d.Get("accept_user_agreement").(bool)
	if !acceptUserAgreement {
		return fmt.Errorf("To proceed with upgrade, you must accept user agreement")
	}
	client := eula.NewAcceptClient(connector)
	err := client.Create()
	if err != nil {
		return err
	}
	return nil
}

func upgradeUc(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	summaryClient := upgrade.NewSummaryClient(connector)
	summary, err := summaryClient.Get()
	if err != nil {
		return err
	}
	if *summary.UpgradeCoordinatorUpdated {
		log.Printf("Upgrade coordinator already upgraded")
		return nil
	}
	client := nsx.NewUpgradeClient(connector)
	err = client.Upgradeuc()
	if err != nil {
		return err
	}
	timeout := d.Get("uc_upgrade_timeout").(int)
	return waitForUcUpgrade(m, timeout)
}

func executePreupgradeChecks(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := nsx.NewUpgradeClient(connector)
	err := client.Executepreupgradechecks(nil, nil, nil, nil, nil, nil)
	if err != nil {
		return err
	}
	timeout := d.Get("precheck_timeout").(int)
	for _, componentType := range precheckComponentTypes {
		log.Printf("Execute pre-upgrade check on %s", componentType)
		err = waitForPrecheckComplete(m, componentType, timeout)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPrecheckErrors(m interface{}, typeParam *string) ([]nsxModel.UpgradeCheckFailure, error) {
	connector := getPolicyConnector(m)
	client := pre_upgrade_checks.NewFailuresClient(connector)
	resultList, err := client.List(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, typeParam, nil, nil)
	if err != nil {
		return nil, err
	}
	return resultList.Results, nil
}

func setFailedPrechecksInSchema(d *schema.ResourceData, precheckErrors []nsxModel.UpgradeCheckFailure) error {
	var failedPrechecksList []map[string]interface{}
	for _, precheckError := range precheckErrors {
		elem := make(map[string]interface{})
		elem["id"] = precheckError.Id
		elem["message"] = precheckError.Message.Message
		elem["type"] = precheckError.Type_
		elem["needs_ack"] = precheckError.NeedsAck
		elem["needs_resolve"] = precheckError.NeedsResolve
		elem["acked"] = precheckError.Acked
		elem["resolution_status"] = precheckError.ResolutionStatus
		failedPrechecksList = append(failedPrechecksList, elem)
	}
	return d.Set("failed_prechecks", failedPrechecksList)
}

func waitForBundleUpload(m interface{}, bundleID string, timeout int) error {
	connector := getPolicyConnector(m)
	client := bundles.NewUploadStatusClient(connector)
	pendingStates := []string{
		nsxModel.UpgradeBundleUploadStatus_STATUS_UPLOADING,
		nsxModel.UpgradeBundleUploadStatus_STATUS_VERIFYING,
	}
	targetStates := []string{
		nsxModel.UpgradeBundleUploadStatus_STATUS_SUCCESS,
		nsxModel.UpgradeBundleUploadStatus_STATUS_FAILED,
	}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(bundleID)
			if err != nil {
				msg := fmt.Sprintf("Error while retrieving upload status of bundle %s", bundleID)
				return state, nsxModel.UpgradeBundleUploadStatus_STATUS_FAILED, logAPIError(msg, err)
			}

			log.Printf("[DEBUG] Current status for uploading bundle %s is %s", bundleID, *state.Status)
			if *state.Status == nsxModel.UpgradeBundleUploadStatus_STATUS_FAILED {
				return state, nsxModel.UpgradeBundleUploadStatus_STATUS_FAILED, liberrors.New(*state.DetailedStatus)
			}

			return state, *state.Status, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to upload bundle %s: %s", bundleID, err)
	}
	return nil
}

func waitForUcUpgrade(m interface{}, timeout int) error {
	connector := getPolicyConnector(m)
	client := upgrade.NewUcUpgradeStatusClient(connector)
	pendingStates := []string{
		nsxModel.UcUpgradeStatus_STATE_NOT_STARTED,
		nsxModel.UcUpgradeStatus_STATE_IN_PROGRESS,
	}
	targetStates := []string{
		nsxModel.UcUpgradeStatus_STATE_SUCCESS,
		nsxModel.UcUpgradeStatus_STATE_FAILED,
	}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get()
			if err != nil {
				if _, ok := err.(errors.InternalServerError); ok {
					log.Printf("[DEBUG] Temporary upstream connection error, retry retrieving UC upgrade status")
				} else {
					return state, nsxModel.UcUpgradeStatus_STATE_FAILED, logAPIError("Error while retrieving UC upgrade status", err)
				}
			}
			if state.State != nil {
				log.Printf("[DEBUG] Current status for UC Upgrade is %s", *state.State)
				return state, *state.State, nil
			}
			return state, nsxModel.UcUpgradeStatus_STATE_IN_PROGRESS, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Failed to upgrade UC: %s", err)
	}
	return nil
}

func waitForPrecheckComplete(m interface{}, componentType string, timeout int) error {
	connector := getPolicyConnector(m)
	client := upgrade.NewStatusSummaryClient(connector)
	pendingStates := []string{
		nsxModel.UpgradeChecksExecutionStatus_STATUS_NOT_STARTED,
		nsxModel.UpgradeChecksExecutionStatus_STATUS_IN_PROGRESS,
		nsxModel.UpgradeChecksExecutionStatus_STATUS_ABORTING,
	}
	targetStates := []string{
		nsxModel.UpgradeChecksExecutionStatus_STATUS_ABORTED,
		nsxModel.UpgradeChecksExecutionStatus_STATUS_COMPLETED,
	}
	stateConf := &resource.StateChangeConf{
		Pending: pendingStates,
		Target:  targetStates,
		Refresh: func() (interface{}, string, error) {
			state, err := client.Get(&componentType, nil, nil)
			if err != nil {
				return state, nsxModel.UpgradeChecksExecutionStatus_STATUS_ABORTED, logAPIError("Error retrieving component upgrade status", err)
			}
			componentStatus := state.ComponentStatus
			if len(componentStatus) != 1 {
				return state, nsxModel.UpgradeChecksExecutionStatus_STATUS_ABORTED, logAPIError("Error retrieving component upgrade status", err)
			}
			return state, *componentStatus[0].PreUpgradeStatus.Status, nil
		},
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 1 * time.Second,
		Delay:      1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Encounter error while running precheck on component type %s: %s", componentType, err)
	}
	return nil
}
