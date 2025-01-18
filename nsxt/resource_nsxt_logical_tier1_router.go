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
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/go-vmware-nsxt/manager"
)

var failOverModeValues = []string{"PREEMPTIVE", "NON_PREEMPTIVE"}
var failOverModeDefaultValue = "PREEMPTIVE"

// formatLogicalRouterRollbackError defines the verbose error when
// rollback fails on a logical router create operation.
const formatLogicalRouterRollbackError = `
WARNING:
There was an error during the creation of logical router %s:
%s
Additionally, there was an error deleting the logical router during rollback:
%s
The logical router may still exist in the NSX. If it does, please manually delete it
and try again.
`

// TODO: add advanced config
func resourceNsxtLogicalTier1Router() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLogicalTier1RouterCreate,
		Read:   resourceNsxtLogicalTier1RouterRead,
		Update: resourceNsxtLogicalTier1RouterUpdate,
		Delete: resourceNsxtLogicalTier1RouterDelete,
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
			"failover_mode": {
				Type:         schema.TypeString,
				Description:  "Failover mode which determines whether the preferred service router instance for given logical router will preempt the peer",
				Default:      failOverModeDefaultValue,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(failOverModeValues, false),
			},
			"firewall_sections": getResourceReferencesSchema(false, true, []string{}, "List of Firewall sections related to the logical router"),
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Description: "Edge Cluster Id",
				// This is needed since updating edge cluster requires special handling
				// which is not yet available in sdk
				ForceNew: true,
				Optional: true,
			},
			"enable_router_advertisement": {
				Type:        schema.TypeBool,
				Description: "Enable router advertisement",
				Default:     false,
				Optional:    true,
			},
			"advertise_connected_routes": {
				Type:        schema.TypeBool,
				Description: "Enable connected NSX routes advertisement",
				Default:     false,
				Optional:    true,
			},
			"advertise_static_routes": {
				Type:        schema.TypeBool,
				Description: "Enable static routes advertisement",
				Default:     false,
				Optional:    true,
			},
			"advertise_nat_routes": {
				Type:        schema.TypeBool,
				Description: "Enable NAT routes advertisement",
				Default:     false,
				Optional:    true,
			},
			"advertise_lb_vip_routes": {
				Type:        schema.TypeBool,
				Description: "Enable LB VIP routes advertisement",
				Default:     false,
				Optional:    true,
			},
			"advertise_lb_snat_ip_routes": {
				Type:        schema.TypeBool,
				Description: "Enable LB SNAT IP routes advertisement",
				Default:     false,
				Optional:    true,
			},

			"advertise_config_revision": getRevisionSchema(),
		},
	}
}

func resourceNsxtLogicalTier1RouterReadAdv(d *schema.ResourceData, nsxClient *api.APIClient, id string) error {
	advConfig, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadAdvertisementConfig(nsxClient.Context, id)
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("LogicalTier1Router %s Advertisement config not found", id)
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router Advertisement config read: %v", err)
	}
	d.Set("enable_router_advertisement", advConfig.Enabled)
	d.Set("advertise_connected_routes", advConfig.AdvertiseNsxConnectedRoutes)
	d.Set("advertise_static_routes", advConfig.AdvertiseStaticRoutes)
	d.Set("advertise_nat_routes", advConfig.AdvertiseNatRoutes)
	d.Set("advertise_lb_vip_routes", advConfig.AdvertiseLbVip)
	d.Set("advertise_lb_snat_ip_routes", advConfig.AdvertiseLbSnatIp)
	d.Set("advertise_config_revision", advConfig.Revision)
	return nil
}

func resourceNsxtLogicalTier1RouterCreateAdv(d *schema.ResourceData, nsxClient *api.APIClient, id string) error {
	enableRouterAdvertisement := d.Get("enable_router_advertisement").(bool)
	if enableRouterAdvertisement {
		advConnected := d.Get("advertise_connected_routes").(bool)
		advStatic := d.Get("advertise_static_routes").(bool)
		advNat := d.Get("advertise_nat_routes").(bool)
		advLbVip := d.Get("advertise_lb_vip_routes").(bool)
		advLbSnatIP := d.Get("advertise_lb_snat_ip_routes").(bool)
		advConfig := manager.AdvertisementConfig{
			Enabled:                     true,
			AdvertiseNsxConnectedRoutes: advConnected,
			AdvertiseStaticRoutes:       advStatic,
			AdvertiseNatRoutes:          advNat,
			AdvertiseLbVip:              advLbVip,
			AdvertiseLbSnatIp:           advLbSnatIP,
		}
		_, _, err := nsxClient.LogicalRoutingAndServicesApi.UpdateAdvertisementConfig(nsxClient.Context, id, advConfig)
		return err
	}
	return nil
}

func resourceNsxtLogicalTier1RouterUpdateAdv(d *schema.ResourceData, nsxClient *api.APIClient, id string) error {
	enableRouterAdvertisement := d.Get("enable_router_advertisement").(bool)
	advConnected := d.Get("advertise_connected_routes").(bool)
	advStatic := d.Get("advertise_static_routes").(bool)
	advNat := d.Get("advertise_nat_routes").(bool)
	advLbVip := d.Get("advertise_lb_vip_routes").(bool)
	advLbSnatIP := d.Get("advertise_lb_snat_ip_routes").(bool)
	advRevision := int64(d.Get("advertise_config_revision").(int))
	advConfig := manager.AdvertisementConfig{
		Enabled:                     enableRouterAdvertisement,
		AdvertiseNsxConnectedRoutes: advConnected,
		AdvertiseStaticRoutes:       advStatic,
		AdvertiseNatRoutes:          advNat,
		AdvertiseLbVip:              advLbVip,
		AdvertiseLbSnatIp:           advLbSnatIP,
		Revision:                    advRevision,
	}
	_, _, err := nsxClient.LogicalRoutingAndServicesApi.UpdateAdvertisementConfig(nsxClient.Context, id, advConfig)
	return err
}

func resourceNsxtLogicalTier1RouterCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	failoverMode := d.Get("failover_mode").(string)
	routerType := "TIER1"
	edgeClusterID := d.Get("edge_cluster_id").(string)
	logicalRouter := manager.LogicalRouter{
		Description:   description,
		DisplayName:   displayName,
		Tags:          tags,
		FailoverMode:  failoverMode,
		RouterType:    routerType,
		EdgeClusterId: edgeClusterID,
	}

	logicalRouter, resp, err := nsxClient.LogicalRoutingAndServicesApi.CreateLogicalRouter(nsxClient.Context, logicalRouter)

	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("Unexpected status returned: %d", resp.StatusCode)
	}

	// Add advertisement config
	err = resourceNsxtLogicalTier1RouterCreateAdv(d, nsxClient, logicalRouter.Id)
	if err != nil {
		// Advertisement configuration creation failed: rollback & delete the router
		log.Printf("[ERROR] Rollback router %s creation", logicalRouter.Id)
		localVarOptionals := make(map[string]interface{})
		_, derr := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouter(nsxClient.Context, logicalRouter.Id, localVarOptionals)
		if derr != nil {
			// rollback failed
			return fmt.Errorf(formatLogicalRouterRollbackError, logicalRouter.Id, err, derr)
		}
		return fmt.Errorf("Error while setting advertisement configuration: %v", err)
	}

	d.SetId(logicalRouter.Id)

	return resourceNsxtLogicalTier1RouterRead(d, m)
}

func resourceNsxtLogicalTier1RouterRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical Tier-1 router id")
	}

	logicalRouter, resp, err := nsxClient.LogicalRoutingAndServicesApi.ReadLogicalRouter(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalTier1Router %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router read: %v", err)
	}

	d.Set("revision", logicalRouter.Revision)
	d.Set("description", logicalRouter.Description)
	d.Set("display_name", logicalRouter.DisplayName)
	setTagsInSchema(d, logicalRouter.Tags)
	d.Set("edge_cluster_id", logicalRouter.EdgeClusterId)
	if logicalRouter.FailoverMode != "" {
		d.Set("failover_mode", logicalRouter.FailoverMode)
	} else {
		// If router is not connected to edge cluster, failover mode will not be set
		// Reset to default value to avoid non-empty plan
		d.Set("failover_mode", failOverModeDefaultValue)
	}
	err = setResourceReferencesInSchema(d, logicalRouter.FirewallSections, "firewall_sections")
	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router firewall sections set in schema: %v", err)
	}

	err = resourceNsxtLogicalTier1RouterReadAdv(d, nsxClient, id)
	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router read: %v", err)
	}

	return nil
}

func resourceNsxtLogicalTier1RouterUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical tier1 router id")
	}

	revision := int64(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	failoverMode := d.Get("failover_mode").(string)
	routerType := "TIER1"
	edgeClusterID := d.Get("edge_cluster_id").(string)
	logicalRouter := manager.LogicalRouter{
		Revision:      revision,
		Description:   description,
		DisplayName:   displayName,
		Tags:          tags,
		FailoverMode:  failoverMode,
		RouterType:    routerType,
		EdgeClusterId: edgeClusterID,
	}
	_, resp, err := nsxClient.LogicalRoutingAndServicesApi.UpdateLogicalRouter(nsxClient.Context, id, logicalRouter)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LogicalTier1Router update error: %v, resp %+v", err, resp)
	}

	// Update advertisement config
	err = resourceNsxtLogicalTier1RouterUpdateAdv(d, nsxClient, id)
	if err != nil {
		return fmt.Errorf("Error while setting config advertisement state: %v", err)
	}

	return resourceNsxtLogicalTier1RouterRead(d, m)
}

func resourceNsxtLogicalTier1RouterDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical tier1 router id")
	}

	localVarOptionals := make(map[string]interface{})
	resp, err := nsxClient.LogicalRoutingAndServicesApi.DeleteLogicalRouter(nsxClient.Context, id, localVarOptionals)
	if err != nil {
		return fmt.Errorf("Error during LogicalTier1Router delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LogicalTier1Router %s not found", id)
		d.SetId("")
	}
	return nil
}
