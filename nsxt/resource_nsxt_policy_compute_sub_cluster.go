// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"

	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPolicyComputeSubCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyComputeSubClusterCreate,
		Read:   resourceNsxtPolicyComputeSubClusterRead,
		Update: resourceNsxtPolicyComputeSubClusterUpdate,
		Delete: resourceNsxtPolicyComputeSubClusterDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyComputeSubClusterImporter,
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
				Description:  "Path to the site this subcluster belongs to",
				Optional:     true,
				ForceNew:     true,
				Default:      defaultInfraSitePath,
				ValidateFunc: validatePolicyPath(),
			},
			"enforcement_point": {
				Type:        schema.TypeString,
				Description: "ID of the enforcement point this subcluster belongs to",
				Optional:    true,
				ForceNew:    true,
				Default:     "default",
			},
			"compute_collection_id": {
				Type:        schema.TypeString,
				Description: "Compute collection ID under which subcluster is created",
				Required:    true,
			},
			"discovered_node_ids": {
				Type:        schema.TypeList,
				Description: "Discovered node IDs under this subcluster",
				Optional:    true,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceNsxtPolicyComputeSubClusterRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	scClient := enforcement_points.NewSubClustersClient(connector)

	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	obj, err := scClient.Get(siteID, epID, id)
	if err != nil {
		return handleReadError(d, "SubCluster", id, err)
	}

	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("compute_collection_id", obj.ComputeCollectionId)
	if obj.SubClusterInfo != nil {
		d.Set("discovered_node_ids", obj.SubClusterInfo.DiscoveredNodeIds)
	}

	return nil
}

func resourceNsxtPolicyComputeSubClusterExists(siteID, epID, id string, connector client.Connector) (bool, error) {
	// Check site existence first
	siteClient := infra.NewSitesClient(connector)
	_, err := siteClient.Get(siteID)
	if err != nil {
		msg := fmt.Sprintf("failed to read site %s", siteID)
		return false, logAPIError(msg, err)
	}
	scClient := enforcement_points.NewSubClustersClient(connector)
	_, err = scClient.Get(siteID, epID, id)

	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving service", err)
}

func policyComputeSubClusterPatch(siteID, epID, id string, d *schema.ResourceData, m interface{}, isUpdate bool, isDelete bool) error {
	connector := getPolicyConnector(m)
	scClient := enforcement_points.NewSubClustersClient(connector)

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	collectionID := d.Get("compute_collection_id").(string)
	tags := getPolicyTagsFromSchema(d)
	discoveredNodes := getStringListFromSchemaList(d, "discovered_node_ids")
	revision := int64(d.Get("revision").(int))
	// Only manual type is supported for now
	manual := model.SubClusterInfo_SUB_CLUSTER_TYPE_MANUAL
	subClusterInfo := &model.SubClusterInfo{
		SubClusterType: &manual,
	}
	obj := model.SubCluster{
		Description:         &description,
		DisplayName:         &displayName,
		Tags:                tags,
		ComputeCollectionId: &collectionID,
		SubClusterInfo:      subClusterInfo,
	}

	if len(discoveredNodes) > 0 && !isDelete {
		// NSX requires us to remove all nodes before deletion,
		// This Patch call will remove all nodes during delete flow,
		// hence the line below is skipped
		obj.SubClusterInfo.DiscoveredNodeIds = discoveredNodes
	}

	if isUpdate {
		obj.Revision = &revision
		_, err := scClient.Update(siteID, epID, id, obj)
		return err
	}

	_, err := scClient.Patch(siteID, epID, id, obj)
	return err
}

func resourceNsxtPolicyComputeSubClusterCreate(d *schema.ResourceData, m interface{}) error {
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
	exists, err := resourceNsxtPolicyComputeSubClusterExists(siteID, epID, id, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("resource with ID %s already exists", id)
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating SubCluster with ID %s under site %s enforcement point %s", id, siteID, epID)
	err = policyComputeSubClusterPatch(siteID, epID, id, d, m, false, false)
	if err != nil {
		return handleCreateError("SubCluster", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyComputeSubClusterRead(d, m)
}

func resourceNsxtPolicyComputeSubClusterUpdate(d *schema.ResourceData, m interface{}) error {
	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating SubCluster with ID %s", id)
	err = policyComputeSubClusterPatch(siteID, epID, id, d, m, true, false)
	if err != nil {
		return handleUpdateError("SubCluster", id, err)
	}

	return resourceNsxtPolicyComputeSubClusterRead(d, m)
}

func resourceNsxtPolicyComputeSubClusterDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	htnClient := enforcement_points.NewSubClustersClient(connector)

	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	// If nodes are present, NSX requires to delete them first
	discoveredNodes := getStringListFromSchemaList(d, "discovered_node_ids")
	if len(discoveredNodes) > 0 {
		err = policyComputeSubClusterPatch(siteID, epID, id, d, m, true, true)
		if err != nil {
			return handleUpdateError("SubCluster", id, err)
		}
	}

	log.Printf("[INFO] Deleting SubCluster with ID %s", id)
	err = htnClient.Delete(siteID, epID, id)
	if err != nil {
		return handleDeleteError("SubCluster", id, err)
	}

	return nil
}

func resourceNsxtPolicyComputeSubClusterImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}

	epID, err := getParameterFromPolicyPath("/enforcement-points/", "/sub-clusters/", importID)
	if err != nil {
		return nil, err
	}
	d.Set("enforcement_point", epID)
	sitePath, err := getSitePathFromChildResourcePath(importID)
	if err != nil {
		return rd, err
	}
	d.Set("site_path", sitePath)

	return rd, nil
}
