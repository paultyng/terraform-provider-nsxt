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
	"github.com/vmware/go-vmware-nsxt/loadbalancer"
)

var cookieModeTypes = []string{"INSERT", "PREFIX", "REWRITE"}
var cookieExpiryTypes = []string{"SESSION_COOKIE_TIME", "PERSISTENCE_COOKIE_TIME"}

func resourceNsxtLbCookiePersistenceProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtLbCookiePersistenceProfileCreate,
		Read:   resourceNsxtLbCookiePersistenceProfileRead,
		Update: resourceNsxtLbCookiePersistenceProfileUpdate,
		Delete: resourceNsxtLbCookiePersistenceProfileDelete,
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
			"cookie_mode": {
				Type:         schema.TypeString,
				Description:  "The cookie persistence mode",
				Optional:     true,
				Default:      "INSERT",
				ValidateFunc: validation.StringInSlice(cookieModeTypes, false),
			},
			"cookie_name": {
				Type:        schema.TypeString,
				Description: "The name of the cookie",
				Required:    true,
			},
			"persistence_shared": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether the cookie persistence is private or shared",
				Optional:    true,
				Default:     false,
			},
			"cookie_fallback": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether once the server points by this cookie is down, a new server is selected, or the requests will be rejected",
				Optional:    true,
				Default:     true,
			},
			"cookie_garble": {
				Type:        schema.TypeBool,
				Description: "A boolean flag which reflects whether the cookie value (server IP and port) would be encrypted or in plain text",
				Optional:    true,
				Default:     true,
			},
			"insert_mode_params": {
				Type:        schema.TypeList,
				Description: "Additional parameters for the INSERT cookie mode",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_domain": {
							Type:        schema.TypeString,
							Description: "HTTP cookie domain",
							Optional:    true,
							// TODO(asarfaty) add validation on cookie domain values
						},
						"cookie_path": {
							Type:        schema.TypeString,
							Description: "HTTP cookie path",
							Optional:    true,
						},
						"cookie_expiry_type": {
							Type:         schema.TypeString,
							Description:  "Type of cookie expiration timing",
							Optional:     true,
							ValidateFunc: validation.StringInSlice(cookieExpiryTypes, false),
						},
						"max_idle_time": {
							Type:        schema.TypeInt,
							Description: "Maximum interval (in seconds) the cookie is valid for from the last time it was seen in a request (required if cookie_expiry_type is set)",
							Optional:    true,
							Computed:    true,
						},
						"max_life_time": {
							Type:        schema.TypeInt,
							Description: "Maximum interval (in seconds) the cookie is valid for from the first time the cookie was seen in a request (required if cookie_expiry_type is SESSION_COOKIE_TIME expiration)",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func getInsertParamsFromSchema(d *schema.ResourceData) (string, string, *loadbalancer.LbCookieTime) {
	confs := d.Get("insert_mode_params").([]interface{})
	cookieMode := d.Get("cookie_mode").(string)
	if cookieMode != "INSERT" {
		return "", "", nil
	}
	for _, conf := range confs {
		// only 1 insert mode config is allowed so return the first 1
		data := conf.(map[string]interface{})
		cookieDomain := data["cookie_domain"].(string)
		cookiePath := data["cookie_path"].(string)
		expiryType := data["cookie_expiry_type"].(string)

		if expiryType == "" {
			// Cookie time is supported only in insert mode
			return cookieDomain, cookiePath, nil
		}

		if expiryType == "SESSION_COOKIE_TIME" {
			return cookieDomain, cookiePath, &loadbalancer.LbCookieTime{
				Type_:         "LbSessionCookieTime",
				CookieMaxIdle: int64(data["max_idle_time"].(int)),
				CookieMaxLife: int64(data["max_life_time"].(int)),
			}
		}
		return cookieDomain, cookiePath, &loadbalancer.LbCookieTime{
			Type_:         "LbPersistenceCookieTime",
			CookieMaxIdle: int64(data["max_idle_time"].(int)),
		}
	}
	return "", "", nil
}

func setInsertParamsInSchema(d *schema.ResourceData, cookieDomain string, cookiePath string, cookieTime *loadbalancer.LbCookieTime) {
	var insertConfigList []map[string]interface{}
	elem := make(map[string]interface{})
	elem["cookie_domain"] = cookieDomain
	elem["cookie_path"] = cookiePath
	if cookieTime != nil {
		elem["max_idle_time"] = cookieTime.CookieMaxIdle
		if cookieTime.Type_ == "LbSessionCookieTime" {
			elem["cookie_expiry_type"] = "SESSION_COOKIE_TIME"
			elem["max_life_time"] = cookieTime.CookieMaxLife
		} else {
			elem["cookie_expiry_type"] = "PERSISTENCE_COOKIE_TIME"
		}
	}
	insertConfigList = append(insertConfigList, elem)
	err := d.Set("insert_mode_params", insertConfigList)
	if err != nil {
		log.Printf("[WARNING]: Failed to set insert_mode_params in schema: %v", err)
	}
}

func resourceNsxtLbCookiePersistenceProfileCreate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	persistenceShared := d.Get("persistence_shared").(bool)
	cookieFallback := d.Get("cookie_fallback").(bool)
	cookieGarble := d.Get("cookie_garble").(bool)
	cookieMode := d.Get("cookie_mode").(string)
	cookieName := d.Get("cookie_name").(string)
	cookieDomain, cookiePath, cookieTime := getInsertParamsFromSchema(d)
	lbCookiePersistenceProfile := loadbalancer.LbCookiePersistenceProfile{
		Description:       description,
		DisplayName:       displayName,
		Tags:              tags,
		PersistenceShared: persistenceShared,
		CookieDomain:      cookieDomain,
		CookieFallback:    cookieFallback,
		CookieGarble:      cookieGarble,
		CookieMode:        cookieMode,
		CookieName:        cookieName,
		CookiePath:        cookiePath,
		CookieTime:        cookieTime,
	}

	lbCookiePersistenceProfile, resp, err := nsxClient.ServicesApi.CreateLoadBalancerCookiePersistenceProfile(nsxClient.Context, lbCookiePersistenceProfile)

	if err != nil {
		return fmt.Errorf("Error during LbCookiePersistenceProfile create: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status returned during LbCookiePersistenceProfile create: %v", resp.StatusCode)
	}
	d.SetId(lbCookiePersistenceProfile.Id)

	return resourceNsxtLbCookiePersistenceProfileRead(d, m)
}

func resourceNsxtLbCookiePersistenceProfileRead(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	lbCookiePersistenceProfile, resp, err := nsxClient.ServicesApi.ReadLoadBalancerCookiePersistenceProfile(nsxClient.Context, id)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbCookiePersistenceProfile %s not found", id)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error during LbCookiePersistenceProfile read: %v", err)
	}

	d.Set("revision", lbCookiePersistenceProfile.Revision)
	d.Set("description", lbCookiePersistenceProfile.Description)
	d.Set("display_name", lbCookiePersistenceProfile.DisplayName)
	setTagsInSchema(d, lbCookiePersistenceProfile.Tags)
	d.Set("persistence_shared", lbCookiePersistenceProfile.PersistenceShared)
	d.Set("cookie_fallback", lbCookiePersistenceProfile.CookieFallback)
	d.Set("cookie_garble", lbCookiePersistenceProfile.CookieGarble)
	d.Set("cookie_mode", lbCookiePersistenceProfile.CookieMode)
	d.Set("cookie_name", lbCookiePersistenceProfile.CookieName)
	setInsertParamsInSchema(d, lbCookiePersistenceProfile.CookieDomain, lbCookiePersistenceProfile.CookiePath, lbCookiePersistenceProfile.CookieTime)

	return nil
}

func resourceNsxtLbCookiePersistenceProfileUpdate(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	revision := int32(d.Get("revision").(int))
	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getTagsFromSchema(d)
	persistenceShared := d.Get("persistence_shared").(bool)
	cookieFallback := d.Get("cookie_fallback").(bool)
	cookieGarble := d.Get("cookie_garble").(bool)
	cookieMode := d.Get("cookie_mode").(string)
	cookieName := d.Get("cookie_name").(string)
	cookieDomain, cookiePath, cookieTime := getInsertParamsFromSchema(d)
	lbCookiePersistenceProfile := loadbalancer.LbCookiePersistenceProfile{
		Revision:          revision,
		Description:       description,
		DisplayName:       displayName,
		Tags:              tags,
		PersistenceShared: persistenceShared,
		CookieDomain:      cookieDomain,
		CookieFallback:    cookieFallback,
		CookieGarble:      cookieGarble,
		CookieMode:        cookieMode,
		CookieName:        cookieName,
		CookiePath:        cookiePath,
		CookieTime:        cookieTime,
	}

	_, resp, err := nsxClient.ServicesApi.UpdateLoadBalancerCookiePersistenceProfile(nsxClient.Context, id, lbCookiePersistenceProfile)

	if err != nil || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("Error during LbCookiePersistenceProfile update: %v", err)
	}

	return resourceNsxtLbCookiePersistenceProfileRead(d, m)
}

func resourceNsxtLbCookiePersistenceProfileDelete(d *schema.ResourceData, m interface{}) error {
	nsxClient := m.(nsxtClients).NsxtClient
	if nsxClient == nil {
		return resourceNotSupportedError()
	}

	id := d.Id()
	if id == "" {
		return fmt.Errorf("Error obtaining logical object id")
	}

	resp, err := nsxClient.ServicesApi.DeleteLoadBalancerPersistenceProfile(nsxClient.Context, id)
	if err != nil {
		return fmt.Errorf("Error during LbCookiePersistenceProfile delete: %v", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		log.Printf("[DEBUG] LbCookiePersistenceProfile %s not found", id)
		d.SetId("")
	}
	return nil
}
