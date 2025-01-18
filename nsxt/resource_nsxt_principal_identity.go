// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mpModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/trust_management"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/trust_management/principal_identities"
	policyModel "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func resourceNsxtPrincipalIdentity() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPrincipalIdentityCreate,
		Read:   resourceNsxtPrincipalIdentityRead,
		Delete: resourceNsxtPrincipalIdentityDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"tag": getTagsSchemaForceNew(),
			"is_protected": {
				Type:        schema.TypeBool,
				Description: "Indicates whether the entities created by this principal should be protected",
				Optional:    true,
				Default:     true,
				ForceNew:    true,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "Name of the principal",
				ValidateFunc: validatePINameOrNodeID(),
				Required:     true,
				ForceNew:     true,
			},
			"node_id": {
				Type:         schema.TypeString,
				Description:  "Unique node-id of a principal",
				ValidateFunc: validatePINameOrNodeID(),
				Required:     true,
				ForceNew:     true,
			},
			"certificate_id": {
				Type:        schema.TypeString,
				Description: "Id of the imported certificate pem",
				Computed:    true,
				ForceNew:    true,
			},
			"certificate_pem": {
				Type:        schema.TypeString,
				Description: "PEM encoding of the new certificate",
				Required:    true,
				ForceNew:    true,
			},
			"roles_for_path": getRolesForPathSchema(true),
		},
	}
}

func validatePINameOrNodeID() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		r := regexp.MustCompile(`^[a-zA-Z0-9]+([-._]?[a-zA-Z0-9]+)*$`)
		if ok := r.MatchString(v); !ok {
			es = append(es, fmt.Errorf("%s %s is invalid", k, v))
			return
		}

		if len(v) < 1 || len(v) > 255 {
			es = append(es, fmt.Errorf("expected length of %s to be in the range (%d - %d), got %d", k, 1, 255, len(v)))
			return
		}
		return
	}
}

func convertToMPRolesForPath(policyRolesForPath []policyModel.RolesForPath) []mpModel.RolesForPath {
	mpRolesForPath := make([]mpModel.RolesForPath, len(policyRolesForPath))
	for i, pRolesForPath := range policyRolesForPath {
		mpRolesForPath[i].DeletePath = pRolesForPath.DeletePath
		mpRolesForPath[i].Path = pRolesForPath.Path
		mpRolesForPath[i].Roles = make([]mpModel.Role, len(pRolesForPath.Roles))
		for j, pRole := range pRolesForPath.Roles {
			mpRolesForPath[i].Roles[j].Role = pRole.Role
			mpRolesForPath[i].Roles[j].RoleDisplayName = pRole.RoleDisplayName
		}
	}
	return mpRolesForPath
}

func convertToPolicyRolesForPath(mpRolesForPath []mpModel.RolesForPath) []policyModel.RolesForPath {
	pRolesForPath := make([]policyModel.RolesForPath, len(mpRolesForPath))
	for i, mRolesForPath := range mpRolesForPath {
		pRolesForPath[i].DeletePath = mRolesForPath.DeletePath
		pRolesForPath[i].Path = mRolesForPath.Path
		pRolesForPath[i].Roles = make([]policyModel.Role, len(mRolesForPath.Roles))
		for j, mRole := range mRolesForPath.Roles {
			pRolesForPath[i].Roles[j].Role = mRole.Role
			pRolesForPath[i].Roles[j].RoleDisplayName = mRole.RoleDisplayName
		}
	}
	return pRolesForPath
}

func resourceNsxtPrincipalIdentityCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := principal_identities.NewWithCertificateClient(connector)

	tags := getMPTagsFromSchema(d)
	isProtected := d.Get("is_protected").(bool)
	name := d.Get("name").(string)
	nodeID := d.Get("node_id").(string)
	certificatePem := d.Get("certificate_pem").(string)
	rolesForPaths := convertToMPRolesForPath(getRolesForPathList(d, rolesForPath{}))

	piObj := mpModel.PrincipalIdentityWithCertificate{
		Tags:           tags,
		IsProtected:    &isProtected,
		Name:           &name,
		NodeId:         &nodeID,
		CertificatePem: &certificatePem,
		RolesForPaths:  rolesForPaths,
	}

	pi, err := client.Create(piObj)
	if err != nil {
		return handleCreateError("PrincipalIdentity", name, err)
	}
	d.SetId(*pi.Id)

	return resourceNsxtPrincipalIdentityRead(d, m)
}

func resourceNsxtPrincipalIdentityRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := trust_management.NewPrincipalIdentitiesClient(connector)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	piObj, err := client.Get(id)
	if err != nil {
		return handleReadError(d, "PrincipalIdentity", id, err)
	}

	setMPTagsInSchema(d, piObj.Tags)
	d.Set("is_protected", piObj.IsProtected)
	d.Set("name", piObj.Name)
	d.Set("node_id", piObj.NodeId)
	d.Set("certificate_id", piObj.CertificateId)
	setRolesForPathInSchema(d, convertToPolicyRolesForPath(piObj.RolesForPaths))
	return nil
}

func resourceNsxtPrincipalIdentityDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	piClient := trust_management.NewPrincipalIdentitiesClient(connector)
	id := d.Id()
	if id == "" {
		return fmt.Errorf("error obtaining logical object id")
	}
	certID := d.Get("certificate_id").(string)
	err := piClient.Delete(id)
	if err != nil {
		// In case PI is already deleted, do not return here to attempt cert deletion.
		if handledErr := handleDeleteError("PrincipalIdentity", id, err); handledErr != nil {
			return handledErr
		}
	}

	// Clean up underlying cert imported by NSX if exists
	if len(certID) > 0 {
		certClient := trust_management.NewCertificatesClient(connector)
		err := certClient.Delete(certID)
		if err != nil {
			return handleDeleteError("Certificate", certID, err)
		}
	}

	return err
}
