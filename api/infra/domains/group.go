//nolint:revive
package domains

// The following file has been autogenerated. Please avoid any changes!
import (
	"errors"

	vapiProtocolClient_ "github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	client1 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/global_infra/domains"
	model1 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm/model"
	client0 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/domains"
	model0 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
	client2 "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/orgs/projects/infra/domains"

	utl "github.com/vmware/terraform-provider-nsxt/api/utl"
)

type GroupClientContext utl.ClientContext

func NewGroupsClient(sessionContext utl.SessionContext, connector vapiProtocolClient_.Connector) *GroupClientContext {
	var client interface{}

	switch sessionContext.ClientType {

	case utl.Local:
		client = client0.NewGroupsClient(connector)

	case utl.Global:
		client = client1.NewGroupsClient(connector)

	case utl.Multitenancy:
		client = client2.NewGroupsClient(connector)

	default:
		return nil
	}
	return &GroupClientContext{Client: client, ClientType: sessionContext.ClientType, ProjectID: sessionContext.ProjectID, VPCID: sessionContext.VPCID}
}

func (c GroupClientContext) Get(domainIdParam string, groupIdParam string) (model0.Group, error) {
	var obj model0.Group
	var err error

	switch c.ClientType {

	case utl.Local:
		client := c.Client.(client0.GroupsClient)
		obj, err = client.Get(domainIdParam, groupIdParam)
		if err != nil {
			return obj, err
		}

	case utl.Global:
		client := c.Client.(client1.GroupsClient)
		gmObj, err1 := client.Get(domainIdParam, groupIdParam)
		if err1 != nil {
			return obj, err1
		}
		var rawObj interface{}
		rawObj, err = utl.ConvertModelBindingType(gmObj, model1.GroupBindingType(), model0.GroupBindingType())
		obj = rawObj.(model0.Group)

	case utl.Multitenancy:
		client := c.Client.(client2.GroupsClient)
		obj, err = client.Get(utl.DefaultOrgID, c.ProjectID, domainIdParam, groupIdParam)
		if err != nil {
			return obj, err
		}

	default:
		return obj, errors.New("invalid infrastructure for model")
	}
	return obj, err
}

func (c GroupClientContext) Patch(domainIdParam string, groupIdParam string, groupParam model0.Group) error {
	var err error

	switch c.ClientType {

	case utl.Local:
		client := c.Client.(client0.GroupsClient)
		err = client.Patch(domainIdParam, groupIdParam, groupParam)

	case utl.Global:
		client := c.Client.(client1.GroupsClient)
		gmObj, err1 := utl.ConvertModelBindingType(groupParam, model0.GroupBindingType(), model1.GroupBindingType())
		if err1 != nil {
			return err1
		}
		err = client.Patch(domainIdParam, groupIdParam, gmObj.(model1.Group))

	case utl.Multitenancy:
		client := c.Client.(client2.GroupsClient)
		err = client.Patch(utl.DefaultOrgID, c.ProjectID, domainIdParam, groupIdParam, groupParam)

	default:
		err = errors.New("invalid infrastructure for model")
	}
	return err
}

func (c GroupClientContext) Update(domainIdParam string, groupIdParam string, groupParam model0.Group) (model0.Group, error) {
	var err error
	var obj model0.Group

	switch c.ClientType {

	case utl.Local:
		client := c.Client.(client0.GroupsClient)
		obj, err = client.Update(domainIdParam, groupIdParam, groupParam)

	case utl.Global:
		client := c.Client.(client1.GroupsClient)
		gmObj, err := utl.ConvertModelBindingType(groupParam, model0.GroupBindingType(), model1.GroupBindingType())
		if err != nil {
			return obj, err
		}
		gmObj, err = client.Update(domainIdParam, groupIdParam, gmObj.(model1.Group))
		if err != nil {
			return obj, err
		}
		obj1, err1 := utl.ConvertModelBindingType(gmObj, model1.GroupBindingType(), model0.GroupBindingType())
		if err1 != nil {
			return obj, err1
		}
		obj = obj1.(model0.Group)

	case utl.Multitenancy:
		client := c.Client.(client2.GroupsClient)
		obj, err = client.Update(utl.DefaultOrgID, c.ProjectID, domainIdParam, groupIdParam, groupParam)

	default:
		err = errors.New("invalid infrastructure for model")
	}
	return obj, err
}

func (c GroupClientContext) Delete(domainIdParam string, groupIdParam string, failIfSubtreeExistsParam *bool, forceParam *bool) error {
	var err error

	switch c.ClientType {

	case utl.Local:
		client := c.Client.(client0.GroupsClient)
		err = client.Delete(domainIdParam, groupIdParam, failIfSubtreeExistsParam, forceParam)

	case utl.Global:
		client := c.Client.(client1.GroupsClient)
		err = client.Delete(domainIdParam, groupIdParam, failIfSubtreeExistsParam, forceParam)

	case utl.Multitenancy:
		client := c.Client.(client2.GroupsClient)
		err = client.Delete(utl.DefaultOrgID, c.ProjectID, domainIdParam, groupIdParam, failIfSubtreeExistsParam, forceParam)

	default:
		err = errors.New("invalid infrastructure for model")
	}
	return err
}

func (c GroupClientContext) List(domainIdParam string, cursorParam *string, includeMarkForDeleteObjectsParam *bool, includedFieldsParam *string, memberTypesParam *string, pageSizeParam *int64, sortAscendingParam *bool, sortByParam *string) (model0.GroupListResult, error) {
	var err error
	var obj model0.GroupListResult

	switch c.ClientType {

	case utl.Local:
		client := c.Client.(client0.GroupsClient)
		obj, err = client.List(domainIdParam, cursorParam, includeMarkForDeleteObjectsParam, includedFieldsParam, memberTypesParam, pageSizeParam, sortAscendingParam, sortByParam)

	case utl.Global:
		client := c.Client.(client1.GroupsClient)
		gmObj, err := client.List(domainIdParam, cursorParam, includeMarkForDeleteObjectsParam, includedFieldsParam, memberTypesParam, pageSizeParam, sortAscendingParam, sortByParam)
		if err != nil {
			return obj, err
		}
		obj1, err1 := utl.ConvertModelBindingType(gmObj, model1.GroupListResultBindingType(), model0.GroupListResultBindingType())
		if err1 != nil {
			return obj, err1
		}
		obj = obj1.(model0.GroupListResult)

	case utl.Multitenancy:
		client := c.Client.(client2.GroupsClient)
		obj, err = client.List(utl.DefaultOrgID, c.ProjectID, domainIdParam, cursorParam, includeMarkForDeleteObjectsParam, includedFieldsParam, memberTypesParam, pageSizeParam, sortAscendingParam, sortByParam)

	default:
		err = errors.New("invalid infrastructure for model")
	}
	return obj, err
}
