// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/aaa"
	nsxModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

const (
	activeDirectoryType = "ActiveDirectory"
	openLdapType        = "OpenLdap"
)

var ldapServerTypes = []string{activeDirectoryType, openLdapType}

func resourceNsxtPolicyLdapIdentitySource() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyLdapIdentitySourceCreate,
		Read:   resourceNsxtPolicyLdapIdentitySourceRead,
		Update: resourceNsxtPolicyLdapIdentitySourceUpdate,
		Delete: resourceNsxtPolicyLdapIdentitySourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"nsx_id": {
				Type:        schema.TypeString,
				Description: "NSX ID for this resource",
				Required:    true,
				ForceNew:    true,
			},
			"description": getDescriptionSchema(),
			"revision":    getRevisionSchema(),
			"tag":         getTagsSchema(),
			"type": {
				Type:         schema.TypeString,
				Description:  "Indicates the type of LDAP server",
				Required:     true,
				ValidateFunc: validation.StringInSlice(ldapServerTypes, false),
				ForceNew:     true,
			},
			"domain_name": {
				Type:        schema.TypeString,
				Description: "Authentication domain name",
				Required:    true,
			},
			"base_dn": {
				Type:        schema.TypeString,
				Description: "DN of subtree for user and group searches",
				Required:    true,
			},
			"alternative_domain_names": {
				Type:        schema.TypeList,
				Description: "Additional domains to be directed to this identity source",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ldap_server": {
				Type:        schema.TypeList,
				Description: "LDAP servers for this identity source",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bind_identity": {
							Type:        schema.TypeString,
							Description: "Username or DN for LDAP authentication",
							Optional:    true,
						},
						"certificates": {
							Type:        schema.TypeList,
							Description: "TLS certificate(s) for LDAP server(s)",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "If true, this LDAP server is enabled",
							Optional:    true,
							Default:     true,
						},
						"password": {
							Type:        schema.TypeString,
							Description: "The authentication password for login",
							Optional:    true,
							Sensitive:   true,
						},
						"url": {
							Type:         schema.TypeString,
							Description:  "The URL for the LDAP server",
							Required:     true,
							ValidateFunc: validateLdapOrLdapsURL(),
						},
						"use_starttls": {
							Type:        schema.TypeBool,
							Description: "Enable/disable StartTLS",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
		},
	}
}

func getLdapServersFromSchema(d *schema.ResourceData) []nsxModel.IdentitySourceLdapServer {
	servers := d.Get("ldap_server").([]interface{})
	serverList := make([]nsxModel.IdentitySourceLdapServer, 0)
	for _, server := range servers {
		data := server.(map[string]interface{})
		bindIdentity := data["bind_identity"].(string)
		certificates := interface2StringList(data["certificates"].([]interface{}))
		enabled := data["enabled"].(bool)
		password := data["password"].(string)
		url := data["url"].(string)
		useStarttls := data["use_starttls"].(bool)
		elem := nsxModel.IdentitySourceLdapServer{
			BindIdentity: &bindIdentity,
			Certificates: certificates,
			Enabled:      &enabled,
			Password:     &password,
			Url:          &url,
			UseStarttls:  &useStarttls,
		}

		serverList = append(serverList, elem)
	}
	return serverList
}

// getLdapServerPasswordMap caches password of ldap servers for setting back to schema after read
func getLdapServerPasswordMap(d *schema.ResourceData) map[string]string {
	passwordMap := make(map[string]string)
	servers := d.Get("ldap_server").([]interface{})
	for _, server := range servers {
		data := server.(map[string]interface{})
		password := data["password"].(string)
		url := data["url"].(string)
		passwordMap[url] = password
	}

	return passwordMap
}

func setLdapServersInSchema(d *schema.ResourceData, nsxLdapServerList []nsxModel.IdentitySourceLdapServer, passwordMap map[string]string) {
	var ldapServerList []map[string]interface{}
	for _, ldapServer := range nsxLdapServerList {
		elem := make(map[string]interface{})
		elem["bind_identity"] = ldapServer.BindIdentity
		elem["certificates"] = ldapServer.Certificates
		elem["enabled"] = ldapServer.Enabled
		elem["url"] = ldapServer.Url
		elem["use_starttls"] = ldapServer.UseStarttls
		if val, ok := passwordMap[*ldapServer.Url]; ok {
			elem["password"] = val
		}
		ldapServerList = append(ldapServerList, elem)
	}
	err := d.Set("ldap_server", ldapServerList)
	if err != nil {
		log.Printf("[WARNING] Failed to set ldap_server in schema: %v", err)
	}
}

func resourceNsxtPolicyLdapIdentitySourceExists(id string, connector client.Connector, isGlobalManager bool) (bool, error) {
	var err error

	ldapClient := aaa.NewLdapIdentitySourcesClient(connector)
	_, err = ldapClient.Get(id)
	if err == nil {
		return true, nil
	}
	if isNotFoundError(err) {
		return false, nil
	}

	return false, logAPIError("Error retrieving resource", err)
}

func resourceNsxtPolicyLdapIdentitySourceProbeAndUpdate(d *schema.ResourceData, m interface{}, id string) error {
	connector := getPolicyConnector(m)
	ldapClient := aaa.NewLdapIdentitySourcesClient(connector)
	converter := bindings.NewTypeConverter()
	serverType := d.Get("type").(string)

	description := d.Get("description").(string)
	revision := int64(d.Get("revision").(int))
	tags := getPolicyTagsFromSchema(d)
	domainName := d.Get("domain_name").(string)
	baseDn := d.Get("base_dn").(string)
	altDomainNames := getStringListFromSchemaList(d, "alternative_domain_names")
	ldapServers := getLdapServersFromSchema(d)

	var dataValue data.DataValue
	var errs []error
	if serverType == activeDirectoryType {
		obj := nsxModel.ActiveDirectoryIdentitySource{
			Description:            &description,
			Revision:               &revision,
			Tags:                   tags,
			DomainName:             &domainName,
			BaseDn:                 &baseDn,
			AlternativeDomainNames: altDomainNames,
			LdapServers:            ldapServers,
			ResourceType:           nsxModel.LdapIdentitySource_RESOURCE_TYPE_ACTIVEDIRECTORYIDENTITYSOURCE,
		}
		dataValue, errs = converter.ConvertToVapi(obj, nsxModel.ActiveDirectoryIdentitySourceBindingType())
	} else if serverType == openLdapType {
		obj := nsxModel.OpenLdapIdentitySource{
			Description:            &description,
			Revision:               &revision,
			Tags:                   tags,
			DomainName:             &domainName,
			BaseDn:                 &baseDn,
			AlternativeDomainNames: altDomainNames,
			LdapServers:            ldapServers,
			ResourceType:           nsxModel.LdapIdentitySource_RESOURCE_TYPE_OPENLDAPIDENTITYSOURCE,
		}
		dataValue, errs = converter.ConvertToVapi(obj, nsxModel.OpenLdapIdentitySourceBindingType())
	}
	if errs != nil {
		return errs[0]
	}
	structValue := dataValue.(*data.StructValue)

	log.Printf("[INFO] Probing LDAP Identity Source with ID %s", id)
	probeResult, err := ldapClient.Probeidentitysource(structValue)
	if err != nil {
		return logAPIError("Error probing LDAP Identity Source", err)
	}
	for _, result := range probeResult.Results {
		if result.Result != nil && *result.Result != nsxModel.IdentitySourceLdapServerProbeResult_RESULT_SUCCESS {
			probeErrs := make([]string, 0)
			for _, probeErr := range result.Errors {
				if probeErr.ErrorType != nil {
					probeErrs = append(probeErrs, *probeErr.ErrorType)
				}
			}
			return fmt.Errorf("LDAP Identity Source server %s probe failed with errors: %s",
				*result.Url, strings.Join(probeErrs, ","))
		}
	}

	log.Printf("[INFO] PUT LDAP Identity Source with ID %s", id)
	if _, err = ldapClient.Update(id, structValue); err != nil {
		return handleUpdateError(serverType, id, err)
	}

	return nil
}

func resourceNsxtPolicyLdapIdentitySourceCreate(d *schema.ResourceData, m interface{}) error {
	// Initialize resource Id and verify this ID is not yet used
	id, err := getOrGenerateID(d, m, resourceNsxtPolicyLdapIdentitySourceExists)
	if err != nil {
		return err
	}

	if err := resourceNsxtPolicyLdapIdentitySourceProbeAndUpdate(d, m, id); err != nil {
		return err
	}
	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyLdapIdentitySourceRead(d, m)
}

func resourceNsxtPolicyLdapIdentitySourceRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	converter := bindings.NewTypeConverter()

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining LDAPIdentitySource ID")
	}

	ldapClient := aaa.NewLdapIdentitySourcesClient(connector)
	structObj, err := ldapClient.Get(id)
	if err != nil {
		return handleReadError(d, "LDAPIdentitySource", id, err)
	}

	obj, errs := converter.ConvertToGolang(structObj, nsxModel.LdapIdentitySourceBindingType())
	if errs != nil {
		return errs[0]
	}

	ldapObj := obj.(nsxModel.LdapIdentitySource)
	resourceType := ldapObj.ResourceType
	var dServerType string
	if resourceType == nsxModel.LdapIdentitySource_RESOURCE_TYPE_ACTIVEDIRECTORYIDENTITYSOURCE {
		dServerType = activeDirectoryType
	} else if resourceType == nsxModel.LdapIdentitySource_RESOURCE_TYPE_OPENLDAPIDENTITYSOURCE {
		dServerType = openLdapType
	} else {
		return fmt.Errorf("unrecognized LdapIdentitySource Resource Type %s", resourceType)
	}

	passwordMap := getLdapServerPasswordMap(d)

	d.Set("nsx_id", id)
	d.Set("description", ldapObj.Description)
	d.Set("revision", ldapObj.Revision)
	setPolicyTagsInSchema(d, ldapObj.Tags)
	d.Set("type", dServerType)
	d.Set("domain_name", ldapObj.DomainName)
	d.Set("base_dn", ldapObj.BaseDn)
	d.Set("alternative_domain_names", ldapObj.AlternativeDomainNames)
	setLdapServersInSchema(d, ldapObj.LdapServers, passwordMap)
	return nil
}

func resourceNsxtPolicyLdapIdentitySourceUpdate(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("rrror obtaining LDAPIdentitySource ID")
	}

	if err := resourceNsxtPolicyLdapIdentitySourceProbeAndUpdate(d, m, id); err != nil {
		return handleUpdateError("LDAPIdentitySource", id, err)
	}

	return resourceNsxtPolicyLdapIdentitySourceRead(d, m)
}

func resourceNsxtPolicyLdapIdentitySourceDelete(d *schema.ResourceData, m interface{}) error {
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining LDAPIdentitySource ID")
	}

	connector := getPolicyConnector(m)
	ldapClient := aaa.NewLdapIdentitySourcesClient(connector)
	if err := ldapClient.Delete(id); err != nil {
		return handleDeleteError("LDAPIdentitySource", id, err)
	}

	return nil
}
