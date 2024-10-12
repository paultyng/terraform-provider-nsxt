// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra"
	gm_model "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
)

var siteTypeValues = []string{
	gm_model.Site_SITE_TYPE_ONPREM_LM,
	gm_model.Site_SITE_TYPE_SDDC_LM,
}

func resourceNsxtPolicySite() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicySiteCreate,
		Read:   resourceNsxtPolicySiteRead,
		Update: resourceNsxtPolicySiteUpdate,
		Delete: resourceNsxtPolicySiteDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtPolicyPathResourceImporter,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"fail_if_rtep_misconfigured": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Fail onboarding if RTEPs misconfigured",
				Default:     true,
			},
			"fail_if_rtt_exceeded": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Fail onboarding if maximum RTT exceeded",
				Default:     true,
			},
			"maximum_rtt": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Maximum acceptable packet round trip time (RTT)",
				Default:      250,
				ValidateFunc: validation.IntBetween(0, 1000),
			},
			"site_connection_info": getConnectionInfoSchema(),
			"site_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Persistent Site Type",
				ValidateFunc: validation.StringInSlice(siteTypeValues, false),
			},
		},
	}
}

func getConnectionInfoSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Connection information",
		MaxItems:    3,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"fqdn": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Fully Qualified Domain Name of the Management Node",
				},
				"password": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "Password",
				},
				"site_uuid": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "id of Site",
				},
				"thumbprint": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Thumbprint of Enforcement Point",
				},
				"username": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Username",
				},
			},
		},
	}
}

func getConnectionInfosFromSchema(d *schema.ResourceData, key string) []gm_model.SiteNodeConnectionInfo {
	var connectionInfos []gm_model.SiteNodeConnectionInfo
	connectionInfoList := d.Get(key).([]interface{})
	for _, sci := range connectionInfoList {
		elem := sci.(map[string]interface{})
		fqdn := elem["fqdn"].(string)
		password := elem["password"].(string)
		siteUUID := elem["site_uuid"].(string)
		thumbprint := elem["thumbprint"].(string)
		username := elem["username"].(string)

		connectionInfo := gm_model.SiteNodeConnectionInfo{
			Fqdn:       &fqdn,
			Password:   &password,
			SiteUuid:   &siteUUID,
			Thumbprint: &thumbprint,
			Username:   &username,
		}
		connectionInfos = append(connectionInfos, connectionInfo)
	}
	return connectionInfos
}

func setConnectionInfosInSchema(d *schema.ResourceData, connectionInfo []gm_model.SiteNodeConnectionInfo, key string) {
	var connectionInfos []interface{}
	for _, sci := range connectionInfo {
		elem := make(map[string]interface{})
		elem["fqdn"] = sci.Fqdn
		elem["password"] = sci.Password
		elem["site_uuid"] = sci.SiteUuid
		elem["thumbprint"] = sci.Thumbprint
		elem["username"] = sci.Username

		connectionInfos = append(connectionInfos, elem)
	}
	d.Set(key, connectionInfos)
}

func getSiteFromSchema(d *schema.ResourceData) gm_model.Site {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getGMTagsFromSchema(d)
	failIfRtepMisconfigured := d.Get("fail_if_rtep_misconfigured").(bool)
	failIfRttExceeded := d.Get("fail_if_rtt_exceeded").(bool)
	maximumRtt := int64(d.Get("maximum_rtt").(int))
	siteConnectionInfos := getConnectionInfosFromSchema(d, "site_connection_info")
	siteType := d.Get("site_type").(string)

	return gm_model.Site{
		DisplayName:             &displayName,
		Description:             &description,
		Tags:                    tags,
		FailIfRtepMisconfigured: &failIfRtepMisconfigured,
		FailIfRttExceeded:       &failIfRttExceeded,
		MaximumRtt:              &maximumRtt,
		SiteConnectionInfo:      siteConnectionInfos,
		SiteType:                &siteType,
	}
}

func resourceNsxtPolicySiteExists(id string, connector client.Connector, isGlobal bool) (bool, error) {
	client := global_infra.NewSitesClient(connector)
	_, err := client.Get(id)
	if err == nil {
		return true, nil
	}

	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicySiteCreate(d *schema.ResourceData, m interface{}) error {
	if !isPolicyGlobalManager(m) {
		return globalManagerOnlyError()
	}
	id, err := getOrGenerateID(d, m, resourceNsxtPolicySiteExists)
	if err != nil {
		return err
	}

	connector := getPolicyConnector(m)
	client := global_infra.NewSitesClient(connector)
	site := getSiteFromSchema(d)

	err = client.Patch(id, site)
	if err != nil {
		return handleCreateError("Site", id, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicySiteRead(d, m)
}

func resourceNsxtPolicySiteRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Site ID")
	}
	client := global_infra.NewSitesClient(connector)
	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "Site", id, err)
	}
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	setGMTagsInSchema(d, obj.Tags)
	d.Set("nsx_id", id)
	d.Set("path", obj.Path)
	d.Set("revision", obj.Revision)

	d.Set("fail_if_rtep_misconfigured", obj.FailIfRtepMisconfigured)
	d.Set("fail_if_rtt_exceeded", obj.FailIfRttExceeded)
	d.Set("maximum_rtt", obj.MaximumRtt)

	setConnectionInfosInSchema(d, obj.SiteConnectionInfo, "site_connection_info")
	d.Set("site_type", obj.SiteType)

	return nil
}

func resourceNsxtPolicySiteUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Site ID")
	}

	connector := getPolicyConnector(m)
	client := global_infra.NewSitesClient(connector)

	obj := getSiteFromSchema(d)
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision

	_, err := client.Update(id, obj)
	if err != nil {
		return handleUpdateError("Site", id, err)
	}

	return nil
}

func resourceNsxtPolicySiteDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining Site ID")
	}
	connector := getPolicyConnector(m)
	client := global_infra.NewSitesClient(connector)
	err := client.Delete(id, nil)
	if err != nil {
		return handleDeleteError("Site", id, err)
	}

	return nil
}
