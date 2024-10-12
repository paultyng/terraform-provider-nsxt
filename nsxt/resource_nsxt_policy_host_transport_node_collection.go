// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/fabric/compute_collections"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

const removeOnDestroyDefault = true

func resourceNsxtPolicyHostTransportNodeCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyHostTransportNodeCollectionCreate,
		Read:   resourceNsxtPolicyHostTransportNodeCollectionRead,
		Update: resourceNsxtPolicyHostTransportNodeCollectionUpdate,
		Delete: resourceNsxtPolicyHostTransportNodeCollectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyHostTransportNodeCollectionImporter,
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
				Description:  "Path to the site this resource belongs to",
				Optional:     true,
				ForceNew:     true,
				Default:      defaultInfraSitePath,
				ValidateFunc: validatePolicyPath(),
			},
			"enforcement_point": {
				Type:        schema.TypeString,
				Description: "ID of the enforcement point this resource belongs to",
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
			},
			"compute_collection_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Compute collection id",
			},
			"sub_cluster_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of sub-cluster configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"host_switch_config_source": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "List of overridden HostSwitch configuration",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"host_switch_id": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "HostSwitch Id",
									},
									"transport_node_profile_sub_config_name": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Name of the TransportNodeProfile sub configuration to be used",
									},
								},
							},
						},
						"sub_cluster_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "sub-cluster Id",
						},
					},
				},
			},
			"transport_node_profile_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Transport Node Profile Path",
			},
			"remove_nsx_on_destroy": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicate whether NSX service should be removed from hypervisors during resource deletion",
				Default:     removeOnDestroyDefault,
			},
		},
	}
}

func resourceNsxtPolicyHostTransportNodeCollectionExists(siteID, epID, id string, connector client.Connector) (bool, error) {
	// Check site existence first
	siteClient := infra.NewSitesClient(connector)
	_, err := siteClient.Get(siteID)
	if err != nil {
		msg := fmt.Sprintf("Failed to read site %s", siteID)
		return false, logAPIError(msg, err)
	}

	client := enforcement_points.NewTransportNodeCollectionsClient(connector)
	_, err = client.Get(siteID, epID, id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func policyHostTransportNodeCollectionUpdate(siteID, epID, id string, isCreate bool, d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)

	computeCollectionID := d.Get("compute_collection_id").(string)
	transportNodeProfileID := d.Get("transport_node_profile_path").(string)
	var subClusterConfigs []model.SubClusterConfig
	for _, scc := range d.Get("sub_cluster_config").([]interface{}) {
		subClusterConfig := scc.(map[string]interface{})
		var hostSwitchConfigSources []model.HostSwitchConfigSource
		for _, hss := range subClusterConfig["host_switch_config_source"].([]interface{}) {
			hsSource := hss.(map[string]interface{})
			hostSwitchID := hsSource["host_switch_id"].(string)
			transportNodeProfileSubConfigName := hsSource["transport_node_profile_sub_config_name"].(string)
			elem := model.HostSwitchConfigSource{
				HostSwitchId:                      &hostSwitchID,
				TransportNodeProfileSubConfigName: &transportNodeProfileSubConfigName,
			}
			hostSwitchConfigSources = append(hostSwitchConfigSources, elem)
		}
		subClusterID := subClusterConfig["sub_cluster_id"].(string)
		elem := model.SubClusterConfig{
			SubClusterId:            &subClusterID,
			HostSwitchConfigSources: hostSwitchConfigSources,
		}
		subClusterConfigs = append(subClusterConfigs, elem)
	}
	obj := model.HostTransportNodeCollection{
		DisplayName:            &displayName,
		Description:            &description,
		Tags:                   tags,
		ComputeCollectionId:    &computeCollectionID,
		TransportNodeProfileId: &transportNodeProfileID,
		SubClusterConfig:       subClusterConfigs,
	}

	if !isCreate {
		revision := int64(d.Get("revision").(int))
		obj.Revision = &revision
	}
	client := enforcement_points.NewTransportNodeCollectionsClient(connector)
	_, err := client.Update(siteID, epID, id, obj, &isCreate, nil)

	return err
}

func resourceNsxtPolicyHostTransportNodeCollectionCreate(d *schema.ResourceData, m interface{}) error {
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

	exists, err := resourceNsxtPolicyHostTransportNodeCollectionExists(siteID, epID, id, connector)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("resource with ID %s already exists", id)
	}

	// Create the resource using PATCH
	log.Printf("[INFO] Creating HostTransportNodeCollection with ID %s under site %s enforcement point %s", id, siteID, epID)
	err = policyHostTransportNodeCollectionUpdate(siteID, epID, id, true, d, m)
	if err != nil {
		return handleCreateError("HostTransportNodeCollection", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyHostTransportNodeCollectionRead(d, m)
}

func resourceNsxtPolicyHostTransportNodeCollectionRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := enforcement_points.NewTransportNodeCollectionsClient(connector)
	// (TODO) Reusing this code here - maybe worthwhile renaming this func as it's usable for other resources
	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}
	obj, err := client.Get(siteID, epID, id)
	if err != nil {
		return handleReadError(d, "HostTransportNodeCollection", id, err)
	}

	d.Set("enforcement_point", epID)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setPolicyTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)
	d.Set("compute_collection_id", obj.ComputeCollectionId)
	if obj.SubClusterConfig != nil {
		var sccList []map[string]interface{}
		for _, cfg := range obj.SubClusterConfig {
			scc := make(map[string]interface{})

			if cfg.HostSwitchConfigSources != nil {
				var hscsList []map[string]interface{}
				for _, src := range cfg.HostSwitchConfigSources {
					hscs := make(map[string]interface{})
					hscs["host_switch_id"] = src.HostSwitchId
					hscs["transport_node_profile_sub_config_name"] = src.TransportNodeProfileSubConfigName
					hscsList = append(hscsList, hscs)
				}
				scc["host_switch_config_source"] = hscsList
			}

			scc["sub_cluster_id"] = cfg.SubClusterId
			sccList = append(sccList, scc)
		}
		d.Set("sub_cluster_config", sccList)
	}
	d.Set("transport_node_profile_path", obj.TransportNodeProfileId)
	return nil
}

func resourceNsxtPolicyHostTransportNodeCollectionUpdate(d *schema.ResourceData, m interface{}) error {
	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Updating HostTransportNodeCollection with ID %s", id)
	err = policyHostTransportNodeCollectionUpdate(siteID, epID, id, false, d, m)

	if err != nil {
		return handleUpdateError("HostTransportNodeCollection", id, err)
	}

	return resourceNsxtPolicyHostTransportNodeCollectionRead(d, m)
}

func getComputeCollectionMemberStateConf(connector client.Connector, id string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{"notyet"},
		Target:  []string{"success", "failed"},
		Refresh: func() (interface{}, string, error) {
			client := compute_collections.NewMemberStatusClient(connector)
			statuses, err := client.List(id)
			if err != nil {
				log.Printf("[DEBUG]: NSX Failed to retrieve compute collection member statuses: %v", err)
				return nil, "failed", err
			}

			// When NSX bits are successfully, no member statuses will remain in the results list
			if len(statuses.Results) > 0 {
				return "notyet", "notyet", nil
			}
			return "success", "success", nil
		},
		Delay:        time.Duration(5) * time.Second,
		Timeout:      time.Duration(1200) * time.Second,
		PollInterval: time.Duration(5) * time.Second,
	}
}

func resourceNsxtPolicyHostTransportNodeCollectionDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := enforcement_points.NewTransportNodeCollectionsClient(connector)
	id, siteID, epID, err := policyIDSiteEPTuple(d, m)
	if err != nil {
		return err
	}

	removeNsxOnDestroy := d.Get("remove_nsx_on_destroy").(bool)
	if removeNsxOnDestroy {
		log.Printf("[INFO] Removing NSX from hosts associated with HostTransportNodeCollection with ID %s", id)
		err = client.Removensx(siteID, epID, id)
		if err != nil {
			return handleDeleteError("HostTransportNodeCollection", id, err)
		}

		// Busy-wait until removal is complete
		ccID := d.Get("compute_collection_id").(string)
		stateConf := getComputeCollectionMemberStateConf(connector, ccID)
		_, err := stateConf.WaitForState()
		if err != nil {
			return fmt.Errorf("failed to remove NSX bits from hosts: %v", err)
		}
	}
	log.Printf("[INFO] Deleting HostTransportNodeCollection with ID %s", id)
	err = client.Delete(siteID, epID, id)
	if err != nil {
		return handleDeleteError("HostTransportNodeCollection", id, err)
	}

	return nil
}

func resourceNsxtPolicyHostTransportNodeCollectionImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err != nil {
		return rd, err
	}

	epID, err := getParameterFromPolicyPath("/enforcement-points/", "/transport-node-collections/", importID)
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
