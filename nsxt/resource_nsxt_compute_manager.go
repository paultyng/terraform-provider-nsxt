// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/terraform-provider-nsxt/nsxt/util"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/data"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/fabric"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

var accessLevelForOidcValues = []string{
	model.ComputeManager_ACCESS_LEVEL_FOR_OIDC_FULL,
	model.ComputeManager_ACCESS_LEVEL_FOR_OIDC_LIMITED,
}

func resourceNsxtComputeManager() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtComputeManagerCreate,
		Read:   resourceNsxtComputeManagerRead,
		Update: resourceNsxtComputeManagerUpdate,
		Delete: resourceNsxtComputeManagerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"revision":     getRevisionSchema(),
			"description":  getDescriptionSchema(),
			"display_name": getDisplayNameSchema(),
			"tag":          getTagsSchema(),
			"access_level_for_oidc": {
				Type:         schema.TypeString,
				Description:  "Specifies access level to NSX from the compute manager",
				Optional:     true,
				ValidateFunc: validation.StringInSlice(accessLevelForOidcValues, false),
				Default:      model.ComputeManager_ACCESS_LEVEL_FOR_OIDC_FULL,
			},
			"create_service_account": {
				Type:        schema.TypeBool,
				Description: "Specifies whether service account is created or not on compute manager",
				Optional:    true,
				Computed:    true,
			},
			"credential": {
				Type:        schema.TypeList,
				Description: "Login credentials for the compute manager",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"saml_login": {
							Type:        schema.TypeList,
							Description: "A login credential specifying saml token",
							Optional:    true,
							MaxItems:    1,
							ExactlyOneOf: []string{
								"credential.0.saml_login",
								"credential.0.session_login",
								"credential.0.username_password_login",
								"credential.0.verifiable_asymmetric_login",
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"thumbprint": {
										Type:        schema.TypeString,
										Description: "Thumbprint of the server",
										Required:    true,
									},
									"token": {
										Type:        schema.TypeString,
										Description: "The saml token to login to server",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
						},
						"session_login": {
							Type:        schema.TypeList,
							Description: "A login credential specifying session_id",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"session_id": {
										Type:        schema.TypeString,
										Description: "The session_id to login to server",
										Required:    true,
										Sensitive:   true,
									},
									"thumbprint": {
										Type:        schema.TypeString,
										Description: "Thumbprint of the login server",
										Required:    true,
									},
								},
							},
						},
						"username_password_login": {
							Type:        schema.TypeList,
							Description: "A login credential specifying a username and password",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"password": {
										Type:        schema.TypeString,
										Description: "The authentication password for login",
										Required:    true,
										Sensitive:   true,
									},
									"thumbprint": {
										Type:        schema.TypeString,
										Description: "Thumbprint of the login server",
										Required:    true,
									},
									"username": {
										Type:        schema.TypeString,
										Description: "The username for login",
										Required:    true,
									},
								},
							},
						},
						"verifiable_asymmetric_login": {
							Type:        schema.TypeList,
							Description: "A verifiable asymmetric login credential",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"asymmetric_credential": {
										Type:        schema.TypeString,
										Description: "Asymmetric login credential",
										Required:    true,
										Sensitive:   true,
									},
									"credential_key": {
										Type:        schema.TypeString,
										Description: "Credential key",
										Required:    true,
										Sensitive:   true,
									},
									"credential_verifier": {
										Type:        schema.TypeString,
										Description: "Credential verifier",
										Required:    true,
										Sensitive:   true,
									},
								},
							},
						},
					},
				},
			},
			"extension_certificate": {
				Type:        schema.TypeList,
				Description: "Specifies certificate for compute manager extension",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pem_encoded": {
							Type:        schema.TypeString,
							Description: "PEM encoded certificate data",
							Required:    true,
						},
						"private_key": {
							Type:        schema.TypeString,
							Description: "Private key of certificate",
							Required:    true,
							Sensitive:   true,
						},
					},
				},
			},
			"multi_nsx": {
				Type:        schema.TypeBool,
				Description: "Specifies whether multi nsx feature is enabled for compute manager",
				Optional:    true,
				Default:     false,
			},
			"origin_type": {
				Type:        schema.TypeString,
				Description: "Compute manager type like vCenter",
				Optional:    true,
				Default:     "vCenter",
			},
			"reverse_proxy_https_port": {
				Type:         schema.TypeInt,
				Description:  "Proxy https port of compute manager",
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 65535),
				Default:      443,
			},
			"server": {
				Type:        schema.TypeString,
				Description: "IP address or hostname of compute manager",
				Required:    true,
			},
			"set_as_oidc_provider": {
				Type:        schema.TypeBool,
				Description: "Specifies whether compute manager has been set as OIDC provider",
				Optional:    true,
				Computed:    true, // default varies based on NSX version
			},
		},
	}
}

func resourceNsxtComputeManagerCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := fabric.NewComputeManagersClient(connector)

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	tags := getMPTagsFromSchema(d)

	var accessLevelForOidc *string
	alfo := d.Get("access_level_for_oidc").(string)
	if alfo != "" {
		accessLevelForOidc = &alfo
	}
	createServiceAccount := d.Get("create_service_account").(bool)
	multiNSX := d.Get("multi_nsx").(bool)
	originType := d.Get("origin_type").(string)
	var reverseProxyHTTPSsPort *int64
	port := int64(d.Get("reverse_proxy_https_port").(int))
	if port != 0 {
		reverseProxyHTTPSsPort = &port
	}
	server := d.Get("server").(string)
	setAsOidcProvider := d.Get("set_as_oidc_provider").(bool)
	credential, err := getCredentialValues(d)
	if err != nil {
		return handleCreateError("ComputeManager", displayName, err)
	}

	obj := model.ComputeManager{
		Description:           &description,
		DisplayName:           &displayName,
		Tags:                  tags,
		AccessLevelForOidc:    accessLevelForOidc,
		Credential:            credential,
		ExtensionCertificate:  getExtensionCertificate(d),
		MultiNsx:              &multiNSX,
		OriginType:            &originType,
		ReverseProxyHttpsPort: reverseProxyHTTPSsPort,
		Server:                &server,
		SetAsOidcProvider:     &setAsOidcProvider,
	}

	// From 9.0.0 onwards CreateServiceAccount can not be false
	// so we can effetively ignore this field
	if util.NsxVersionLower("9.0.0") {
		obj.CreateServiceAccount = &createServiceAccount
	}

	log.Printf("[INFO] Creating Compute Manager %s", displayName)
	obj, err = client.Create(obj)
	if err != nil {
		return handleCreateError("Compute Manager", displayName, err)
	}

	d.SetId(*obj.Id)
	return resourceNsxtComputeManagerRead(d, m)
}

func getCredentialData(data map[string]interface{}) (string, map[string]interface{}) {
	credTypes := []string{
		"saml_login",
		"session_login",
		"username_password_login",
		"verifiable_asymmetric_login",
	}
	for _, credType := range credTypes {
		if data[credType] != nil && len(data[credType].([]interface{})) > 0 {
			return credType, data[credType].([]interface{})[0].(map[string]interface{})
		}
	}
	return "", nil
}

func getCredentialValues(d *schema.ResourceData) (*data.StructValue, error) {
	credentialList := d.Get("credential").([]interface{})
	converter := bindings.NewTypeConverter()

	for _, c := range credentialList {
		credType, cData := getCredentialData(c.(map[string]interface{}))
		var dataValue data.DataValue
		var errs []error
		switch credType {
		case "saml_login":
			thumbPrint := cData["thumbprint"].(string)
			token := cData["token"].(string)
			cred := model.SamlTokenLoginCredential{
				Thumbprint:     &thumbPrint,
				Token:          &token,
				CredentialType: model.SamlTokenLoginCredential__TYPE_IDENTIFIER,
			}
			dataValue, errs = converter.ConvertToVapi(cred, model.SamlTokenLoginCredentialBindingType())

		case "session_login":
			sessionID := cData["session_id"].(string)
			thumbPrint := cData["thumbprint"].(string)
			cred := model.SessionLoginCredential{
				SessionId:      &sessionID,
				Thumbprint:     &thumbPrint,
				CredentialType: model.SessionLoginCredential__TYPE_IDENTIFIER,
			}
			dataValue, errs = converter.ConvertToVapi(cred, model.SessionLoginCredentialBindingType())

		case "username_password_login":
			password := cData["password"].(string)
			thumbPrint := cData["thumbprint"].(string)
			username := cData["username"].(string)

			cred := model.UsernamePasswordLoginCredential{
				Thumbprint:     &thumbPrint,
				Password:       &password,
				Username:       &username,
				CredentialType: model.UsernamePasswordLoginCredential__TYPE_IDENTIFIER,
			}
			dataValue, errs = converter.ConvertToVapi(cred, model.UsernamePasswordLoginCredentialBindingType())

		case "verifiable_asymmetric_login":
			asymmetricCredential := cData["asymmetric_credential"].(string)
			credentialKey := cData["credential_key"].(string)
			credentialVerifier := cData["credential_verifier"].(string)

			cred := model.VerifiableAsymmetricLoginCredential{
				AsymmetricCredential: &asymmetricCredential,
				CredentialKey:        &credentialKey,
				CredentialVerifier:   &credentialVerifier,
				CredentialType:       model.VerifiableAsymmetricLoginCredential__TYPE_IDENTIFIER,
			}
			dataValue, errs = converter.ConvertToVapi(cred, model.VerifiableAsymmetricLoginCredentialBindingType())

		default:
			return nil, errors.New("no valid credential found")
		}
		if errs != nil {
			return nil, errs[0]
		}
		entryStruct := dataValue.(*data.StructValue)
		return entryStruct, nil
	}
	return nil, nil
}

func getExtensionCertificate(d *schema.ResourceData) *model.CertificateData {
	extensionCertificateList := d.Get("extension_certificate").([]interface{})
	for _, ec := range extensionCertificateList {
		data := ec.(map[string]interface{})
		pemEncoded := data["pem_encoded"].(string)
		privateKey := data["private_key"].(string)
		return &model.CertificateData{
			PemEncoded: &pemEncoded,
			PrivateKey: &privateKey,
		}
	}
	return nil
}

func resourceNsxtComputeManagerRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := fabric.NewComputeManagersClient(connector)

	obj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "ComputeManager", id, err)
	}

	d.Set("revision", obj.Revision)
	d.Set("description", obj.Description)
	d.Set("display_name", obj.DisplayName)
	setMPTagsInSchema(d, obj.Tags)

	d.Set("access_level_for_oidc", obj.AccessLevelForOidc)
	d.Set("create_service_account", obj.CreateServiceAccount)
	setCredentialValuesInSchema(d, obj.Credential)
	d.Set("extension_certificate", obj.ExtensionCertificate)
	d.Set("multi_nsx", obj.MultiNsx)
	d.Set("origin_type", obj.OriginType)
	d.Set("reverse_proxy_https_port", obj.ReverseProxyHttpsPort)
	d.Set("server", obj.Server)
	d.Set("set_as_oidc_provider", obj.SetAsOidcProvider)

	return nil
}

func setCredentialValuesInSchema(d *schema.ResourceData, credential *data.StructValue) error {
	converter := bindings.NewTypeConverter()
	parentElem := getElemOrEmptyMapFromSchema(d, "credential")

	base, errs := converter.ConvertToGolang(credential, model.LoginCredentialBindingType())
	if errs != nil {
		return errs[0]
	}
	credType := base.(model.LoginCredential).CredentialType

	switch credType {
	case model.SamlTokenLoginCredential__TYPE_IDENTIFIER:
		elem := getElemOrEmptyMapFromMap(parentElem, "saml_login")
		entry, errs := converter.ConvertToGolang(credential, model.SamlTokenLoginCredentialBindingType())
		if errs != nil {
			return errs[0]
		}
		credEntry := entry.(model.SamlTokenLoginCredential)
		// Normally NSX won't return sensitive info, in which case
		// we need to keep values from intent to avoid permadiff
		if credEntry.Thumbprint != nil {
			elem["thumbprint"] = credEntry.Thumbprint
		}
		if credEntry.Token != nil {
			elem["token"] = credEntry.Token
		}
		parentElem["saml_login"] = []interface{}{elem}

	case model.SessionLoginCredential__TYPE_IDENTIFIER:
		elem := getElemOrEmptyMapFromMap(parentElem, "session_login")
		entry, errs := converter.ConvertToGolang(credential, model.SessionLoginCredentialBindingType())
		if errs != nil {
			return errs[0]
		}
		credEntry := entry.(model.SessionLoginCredential)
		// Normally NSX won't return sensitive info, in which case
		// we need to keep values from intent to avoid permadiff
		if credEntry.SessionId != nil {
			elem["session_id"] = credEntry.SessionId
		}
		if credEntry.Thumbprint != nil {
			elem["thumbprint"] = credEntry.Thumbprint
		}
		parentElem["session_login"] = []interface{}{elem}

	case model.UsernamePasswordLoginCredential__TYPE_IDENTIFIER:
		elem := getElemOrEmptyMapFromMap(parentElem, "username_password_login")
		entry, errs := converter.ConvertToGolang(credential, model.UsernamePasswordLoginCredentialBindingType())
		if errs != nil {
			return errs[0]
		}
		credEntry := entry.(model.UsernamePasswordLoginCredential)
		// Normally NSX won't return sensitive info, in which case
		// we need to keep values from intent to avoid permadiff
		if credEntry.Username != nil {
			elem["username"] = credEntry.Username
		}
		if credEntry.Password != nil {
			elem["password"] = credEntry.Password
		}
		if credEntry.Thumbprint != nil {
			elem["thumbprint"] = credEntry.Thumbprint
		}
		parentElem["username_password_login"] = []interface{}{elem}

	case model.VerifiableAsymmetricLoginCredential__TYPE_IDENTIFIER:
		elem := getElemOrEmptyMapFromMap(parentElem, "verifiable_asymmetric_login")
		entry, errs := converter.ConvertToGolang(credential, model.VerifiableAsymmetricLoginCredentialBindingType())
		if errs != nil {
			return errs[0]
		}
		credEntry := entry.(model.VerifiableAsymmetricLoginCredential)
		// Normally NSX won't return sensitive info, in which case
		// we need to keep values from intent to avoid permadiff
		if credEntry.AsymmetricCredential != nil {
			elem["asymmetric_credential"] = credEntry.AsymmetricCredential
		}
		if credEntry.CredentialKey != nil {
			elem["credential_key"] = credEntry.CredentialKey
		}
		if credEntry.CredentialVerifier != nil {
			elem["credential_verifier"] = credEntry.CredentialVerifier
		}
		parentElem["verifiable_asymmetric_login"] = []interface{}{elem}

	default:
		return errors.New("no valid credential found")
	}

	d.Set("credential", []interface{}{parentElem})
	return nil
}

func resourceNsxtComputeManagerUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := fabric.NewComputeManagersClient(connector)

	description := d.Get("description").(string)
	displayName := d.Get("display_name").(string)
	revision := int64(d.Get("revision").(int))
	tags := getMPTagsFromSchema(d)
	var accessLevelForOidc *string
	alfo := d.Get("access_level_for_oidc").(string)
	if alfo != "" {
		accessLevelForOidc = &alfo
	}
	createServiceAccount := d.Get("create_service_account").(bool)
	multiNSX := d.Get("multi_nsx").(bool)
	originType := d.Get("origin_type").(string)
	var reverseProxyHTTPSPort *int64
	port := int64(d.Get("reverse_proxy_https_port").(int))
	if port != 0 {
		reverseProxyHTTPSPort = &port
	}
	server := d.Get("server").(string)
	setAsOidcProvider := d.Get("set_as_oidc_provider").(bool)
	credential, err := getCredentialValues(d)
	if err != nil {
		return handleUpdateError("ComputeManager", id, err)
	}

	obj := model.ComputeManager{
		Description:           &description,
		DisplayName:           &displayName,
		Tags:                  tags,
		AccessLevelForOidc:    accessLevelForOidc,
		Credential:            credential,
		ExtensionCertificate:  getExtensionCertificate(d),
		MultiNsx:              &multiNSX,
		OriginType:            &originType,
		ReverseProxyHttpsPort: reverseProxyHTTPSPort,
		Server:                &server,
		SetAsOidcProvider:     &setAsOidcProvider,
		Revision:              &revision,
	}

	// From 9.0.0 onwards CreateServiceAccount can not be false
	// so we can effetively ignore this field
	if util.NsxVersionLower("9.0.0") {
		obj.CreateServiceAccount = &createServiceAccount
	}

	_, err = client.Update(id, obj)
	if err != nil {
		return handleUpdateError("ComputeManager", id, err)
	}

	return resourceNsxtComputeManagerRead(d, m)
}

func resourceNsxtComputeManagerDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)

	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}

	client := fabric.NewComputeManagersClient(connector)

	err := client.Delete(id)
	if err != nil {
		return handleDeleteError("ComputeManager", id, err)
	}
	return nil
}
