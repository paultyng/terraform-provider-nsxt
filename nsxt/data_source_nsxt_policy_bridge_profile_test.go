/* Copyright © 2022 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"fmt"
	"github.com/vmware/vsphere-automation-sdk-go/lib/vapi/std/errors"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	ep "github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/sites/enforcement_points/edge_clusters"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

func TestAccDataSourceNsxtPolicyBridgeProfile_basic(t *testing.T) {
	name := getAccTestDataSourceName()
	testResourceName := "data.nsxt_policy_bridge_profile.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccOnlyLocalManager(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccDataSourceNsxtPolicyBridgeProfileDeleteByName(name)
		},
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if err := testAccDataSourceNsxtPolicyBridgeProfileCreate(name); err != nil {
						t.Error(err)
					}
				},
				Config: testAccNsxtPolicyBridgeProfileReadTemplate(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "display_name", name),
					resource.TestCheckResourceAttr(testResourceName, "description", name),
					resource.TestCheckResourceAttrSet(testResourceName, "path"),
				),
			},
		},
	})
}

func getEdgeClusterPath() (string, error) {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return "", fmt.Errorf("error during test client initialization: %v", err)
	}
	client := nsx.NewEdgeClustersClient(connector)
	clusterList, err := client.List(nil, nil, nil, nil, nil)
	if err != nil {
		return "", fmt.Errorf("error during edge cluster list retrieval: %v", err)
	}
	clusterName := getEdgeClusterName()
	clusterID := ""
	if clusterList.Results != nil {
		for _, objList := range clusterList.Results {
			if *objList.DisplayName == clusterName {
				clusterID = *objList.Id
				break
			}
		}
	}
	if clusterID != "" {
		enClient := edge_clusters.NewEdgeNodesClient(connector)
		objList, err := enClient.List(defaultSite, defaultEnforcementPoint, clusterID, nil, nil, nil, nil, nil, nil)
		if err != nil {
			return "", fmt.Errorf("error during edge node list retrieval: %v", err)
		}
		if len(objList.Results) > 0 {
			return *objList.Results[0].Path, nil
		}
	}
	return "", errors.NotFound{}
}

func testAccDataSourceNsxtPolicyBridgeProfileCreate(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	displayName := name
	description := name
	obj := model.L2BridgeEndpointProfile{
		Description: &description,
		DisplayName: &displayName,
	}

	// Generate a random ID for the resource
	uuid, _ := uuid.NewRandom()
	id := uuid.String()

	// Get Edge path
	edgePath, err := getEdgeClusterPath()
	if err == nil {
		obj.EdgePaths = []string{edgePath}
	} else if !isNotFoundError(err) {
		return err
	}
	client := ep.NewEdgeBridgeProfilesClient(connector)
	err = client.Patch(defaultSite, defaultEnforcementPoint, id, obj)

	if err != nil {
		return fmt.Errorf("Error during Bridge Profile creation: %v", err)
	}
	return nil
}

func testAccDataSourceNsxtPolicyBridgeProfileDeleteByName(name string) error {
	connector, err := testAccGetPolicyConnector()
	if err != nil {
		return fmt.Errorf("Error during test client initialization: %v", err)
	}

	// Find the object by name
	objID, err := testGetObjIDByName(name, "L2BridgeEndpointProfile")
	if err != nil {
		return nil
	}
	client := ep.NewEdgeBridgeProfilesClient(connector)
	err = client.Delete(defaultSite, defaultEnforcementPoint, objID)
	if err != nil {
		return fmt.Errorf("Error during Bridge Profile deletion: %v", err)
	}
	return nil
}

func testAccNsxtPolicyBridgeProfileReadTemplate(name string) string {
	return fmt.Sprintf(`
data "nsxt_policy_bridge_profile" "test" {
  display_name = "%s"
}`, name)
}
