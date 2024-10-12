// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points/host_transport_nodes"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	model2 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyHostTransportNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyHostTransportNodeCreate,
		Read:   resourceNsxtPolicyHostTransportNodeRead,
		Update: resourceNsxtPolicyHostTransportNodeUpdate,
		Delete: resourceNsxtPolicyHostTransportNodeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyHostTransportNodeImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"site_path": {
				Type:         schema.TypeString,
				Description:  "Path to the site this Host Transport Node belongs to",
				Optional:     true,
				ForceNew:     true,
				Default:      defaultInfraSitePath,
				ValidateFunc: validatePolicyPath(),
			},
			"enforcement_point": {
				Type:        schema.TypeString,
				Description: "ID of the enforcement point this Host Transport Node belongs to",
				Optional:    true,
				ForceNew:    true,
				Default:     "default",
			},
			"discovered_node_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Discovered node id to create Host Transport Node",
			},
			// host_switch_spec
			"standard_host_switch": getStandardHostSwitchSchema(nodeTypeHost),
			"remove_nsx_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicate whether NSX service should be removed from hypervisor during resource deletion",
				Default:     removeOnDestroyDefault,
			},
		},
	}
}

func resourceNsxtPolicyHostTransportNodeRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	htnClient := enforcement_points.NewHostTransportNodesClient(connector)

	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	obj, err := htnClient.Get(siteID, epID, id)
	if err != nil {
		return handleReadError(d, "HostTransportNode", id, err)
	}
	sitePath, err := getSitePathFromChildResourcePath(*obj.ParentPath)
	if err != nil {
		return handleReadError(d, "HostTransportNode", id, err)
	}

	d.Set("site_path", sitePath)
	d.Set("enforcement_point", epID)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	if obj.NodeDeploymentInfo != nil {
		d.Set("discovered_node_id", obj.NodeDeploymentInfo.DiscoveredNodeId)
	}
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	err = setHostSwitchSpecInSchema(d, obj.HostSwitchSpec, nodeTypeHost)
	if err != nil {
		return err
	}

	return nil
}

func resourceNsxtPolicyHostTransportNodeExists(siteID, epID, tzID string, connector client.Connector) (bool, error) {
	var err error

	// Check site existence first
	siteClient := infra.NewSitesClient(connector)
	_, err = siteClient.Get(siteID)
	if err != nil {
		msg := fmt.Sprintf("failed to read site %s", siteID)
		return false, logAPIError(msg, err)
	}

	// Check (ep, htn) existence. In case of ep not found, NSX returns BAD_REQUEST
	htnClient := enforcement_points.NewHostTransportNodesClient(connector)
	_, err = htnClient.Get(siteID, epID, tzID)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func policyHostTransportNodePatch(siteID, epID, htnID string, d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	htnClient := enforcement_points.NewHostTransportNodesClient(connector)

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getPolicyTagsFromSchema(d)
	discoveredNodeID := d.Get("discovered_node_id").(string)
	hostSwitchSpec, err := getHostSwitchSpecFromSchema(d, m, nodeTypeHost)
	revision := int64(d.Get("revision").(int))
	if err != nil {
		return fmt.Errorf("failed to create hostSwitchSpec of HostTransportNode: %v", err)
	}

	obj := model2.HostTransportNode{
		Description:               &description,
		DisplayName:               &displayName,
		Tags:                      tags,
		HostSwitchSpec:            hostSwitchSpec,
		DiscoveredNodeIdForCreate: &discoveredNodeID,
		Revision:                  &revision,
	}

	return htnClient.Patch(siteID, epID, htnID, obj, nil, nil, nil, nil, nil, nil, nil)
}

func resourceNsxtPolicyHostTransportNodeCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	}
	sitePath := d.Get("site_path").(string)
	siteID := getResourceIDFromResourcePath(sitePath, "sites")
	if siteID == "" {
		return fmt.Errorf("error obtaining Site ID from site path %s", sitePath)
	}
	epID := d.Get("enforcement_point").(string)
	if epID == "" {
		epID = getPolicyEnforcementPoint(m)
	}
	exists, err := resourceNsxtPolicyHostTransportNodeExists(siteID, epID, id, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("resource with ID %s already exists", id)
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating HostTransportNode with ID %s under site %s enforcement point %s", id, siteID, epID)
	err = policyHostTransportNodePatch(siteID, epID, id, d, m)
	if err != nil {
		return handleCreateError("HostTransportNode", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyHostTransportNodeRead(d, m)
}

func resourceNsxtPolicyHostTransportNodeUpdate(d *schema.ResourceData, m interface{}) error {
	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating HostTransportNode with ID %s", id)
	err = policyHostTransportNodePatch(siteID, epID, id, d, m)
	if err != nil {
		return handleUpdateError("HostTransportNode", id, err)
	}

	return resourceNsxtPolicyHostTransportNodeRead(d, m)
}

func getHostTransportNodeStateConf(connector client.Connector, id, siteID, epID string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{"notyet"},
		Target:  []string{"success", "failed"},
		Refresh: func() (interface{}, string, error) {
			client := host_transport_nodes.NewStateClient(connector)
			_, err := client.Get(siteID, epID, id)

			if isNotFoundError(err) {
				return "success", "success", nil
			}

			if err != nil {
				log.Printf("[DEBUG]: NSX Failed to retrieve HostTransportNode state: %v", err)
				return nil, "failed", err
			}

			return "notyet", "notyet", nil
		},
		Delay:        time.Duration(5) * time.Second,
		Timeout:      time.Duration(1200) * time.Second,
		PollInterval: time.Duration(5) * time.Second,
	}
}

func resourceNsxtPolicyHostTransportNodeDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	htnClient := enforcement_points.NewHostTransportNodesClient(connector)

	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	removeNsxOnDestroy := d.Get("remove_nsx_on_destroy").(bool)

	log.Printf("[INFO] Deleting HostTransportNode with ID %s", id)
	err = htnClient.Delete(siteID, epID, id, nil, &removeNsxOnDestroy)
	if err != nil {
		return handleDeleteError("HostTransportNode", id, err)
	}

	if removeNsxOnDestroy {
		log.Printf("[INFO] Removing NSX from host HostTransportNode with ID %s", id)

		// Busy-wait until removal is complete
		stateConf := getHostTransportNodeStateConf(connector, id, siteID, epID)
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("failed to remove NSX bits from hosts: %v", err)
		}
	}

	return nil
}

func resourceNsxtPolicyHostTransportNodeImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}

	epID, err := getParameterFromPolicyPath("/enforcement-points/", "/host-transport-nodes/", importID)
	if err != nil {
		return nil, err
	}
	d.Set("enforcement_point", epID)
	sitePath, err := getSitePathFromChildResourcePath(importID)
	if err != nil {
		return rd, err
	}
	d.Set("site_path", sitePath)
	d.Set("remove_nsx_on_destroy", removeOnDestroyDefault)

	return rd, nil
}
