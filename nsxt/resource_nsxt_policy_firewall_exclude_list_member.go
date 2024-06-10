/* Copyright © 2023 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	"github.com/vmware/terraform-provider-nsxt/api/infra/settings/firewall/security"
)

func resourceNsxtPolicyFirewallExcludeListMember() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyFirewallExcludeListMemberCreate,
		Read:   resourceNsxtPolicyFirewallExcludeListMemberRead,
		Delete: resourceNsxtPolicyFirewallExcludeListMemberDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"member": {
				Type:         schema.TypeString,
				Description:  "ExcludeList member",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validatePolicyPath(),
			},
		},
	}
}

func memberInList(member string, members []string) int {
	for i, mem := range members {
		if mem == member {
			return i
		}
	}
	return -1
}

func resourceNsxtPolicyFirewallExcludeListMemberExists(sessionContext utl.SessionContext, id string, connector client.Connector) (bool, error) {

	client := security.NewExcludeListClient(sessionContext, connector)
	obj, err := client.Get()
	if isNotFoundError(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if 0 <= memberInList(id, obj.Members) {
		return true, nil
	}

	return false, nil
}

func resourceNsxtPolicyFirewallExcludeListMemberCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	member := d.Get("member").(string)

	doUpdate := func() error {
		var obj model.PolicyExcludeList

		context, err := getSessionContext(d, m)
		if err != nil {
			return err
		}
		client := security.NewExcludeListClient(context, connector)
		obj, err = client.Get()
		if isNotFoundError(err) {
			obj = model.PolicyExcludeList{
				Members: []string{member},
			}
		} else if err != nil {
			return err
		}
		if 0 <= memberInList(member, obj.Members) {
			return errors.AlreadyExists{}
		}
		obj.Members = append(obj.Members, member)
		_, err = client.Update(obj)
		if err != nil {
			return err
		}

		d.SetId(member)

		return nil
	}
	commonProviderConfig := getCommonProviderConfig(m)
	err := retryUponPreconditionFailed(doUpdate, commonProviderConfig.MaxRetries)
	if err != nil {
		return handleCreateError("PolicyFirewallExcludeListMember", member, err)
	}

	return resourceNsxtPolicyFirewallExcludeListMemberRead(d, m)
}

func resourceNsxtPolicyFirewallExcludeListMemberRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	member := d.Id()

	context, err := getSessionContext(d, m)
	if err != nil {
		return err
	}
	client := security.NewExcludeListClient(context, connector)
	obj, err := client.Get()
	if err != nil {
		return handleReadError(d, "PolicyFirewallExcludeListMember", member, err)
	}
	if 0 > memberInList(member, obj.Members) {
		return errors.NotFound{}
	}
	d.Set("member", member)
	return nil
}

func resourceNsxtPolicyFirewallExcludeListMemberDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	member := d.Get("member").(string)

	doUpdate := func() error {
		var obj model.PolicyExcludeList

		context, err := getSessionContext(d, m)
		if err != nil {
			return err
		}
		client := security.NewExcludeListClient(context, connector)
		obj, err = client.Get()
		if isNotFoundError(err) {
			return nil
		} else if err != nil {
			return err
		}
		i := memberInList(member, obj.Members)
		if i < 0 {
			return errors.NotFound{}
		}

		obj.Members = append(obj.Members[:i], obj.Members[i+1:]...)
		_, err = client.Update(obj)
		return err
	}
	commonProviderConfig := getCommonProviderConfig(m)
	err := retryUponPreconditionFailed(doUpdate, commonProviderConfig.MaxRetries)
	if err != nil {
		return handleDeleteError("PolicyFirewallExcludeListMember", member, err)
	}
	return nil
}
