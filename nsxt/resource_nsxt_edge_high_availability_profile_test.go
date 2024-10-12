// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp/nsx/model"
)

func TestAccResourceNsxtEdgeHighAvailabilityProfile_basic(t *testing.T) {
	profileName := getAccTestResourceName()
	updateProfileName := "updated-" + profileName
	testResourceName := "nsxt_edge_high_availability_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestFabric(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXEdgeHighAvailabilityProfileCheckDestroy(state, updateProfileName)
		},

		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeHighAvailabilityProfileCreateTemplate(profileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXEdgeHighAvailabilityProfileExists(profileName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", profileName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test edge high availability profile"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
			{
				Config: testAccNSXEdgeHighAvailabilityProfileCreateTemplate(updateProfileName),
				Check: resource.ComposeTestCheckFunc(
					testAccNSXEdgeHighAvailabilityProfileExists(updateProfileName, testResourceName),
					resource.TestCheckResourceAttr(testResourceName, "display_name", updateProfileName),
					resource.TestCheckResourceAttr(testResourceName, "description", "Terraform test edge high availability profile"),
					resource.TestCheckResourceAttr(testResourceName, "tag.#", "1"),
				),
			},
		},
	})
}

func TestAccResourceNsxtEdgeHighAvailabilityProfile_importBasic(t *testing.T) {
	name := getAccTestResourceName()
	profileName := getAccTestResourceName()
	testResourceName := "nsxt_edge_high_availability_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccOnlyLocalManager(t); testAccTestFabric(t); testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			return testAccNSXEdgeHighAvailabilityProfileCheckDestroy(state, name)
		},
		Steps: []resource.TestStep{
			{
				Config: testAccNSXEdgeHighAvailabilityProfileCreateTemplate(profileName),
			},
			{
				ResourceName:      testResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccNSXEdgeHighAvailabilityProfileExists(displayName string, resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("NSX Edge High Availability Profile resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("NSX Edge High Availability Profile resource ID not set in resources ")
		}

		client := nsx.NewClusterProfilesClient(connector)
		structValue, err := client.Get(resourceID)
		if err != nil {
			return fmt.Errorf("error while retrieving Edge High Availability Profile ID %s. Error: %v", resourceID, err)
		}
		converter := bindings.NewTypeConverter()
		o, errs := converter.ConvertToGolang(structValue, model.EdgeHighAvailabilityProfileBindingType())
		if errs != nil {
			return errs[0]
		}
		obj := o.(model.EdgeHighAvailabilityProfile)

		if displayName == *obj.DisplayName {
			return nil
		}
		return fmt.Errorf("NSX Edge High Availability Profile %s wasn't found", displayName)
	}
}

func testAccNSXEdgeHighAvailabilityProfileCheckDestroy(state *terraform.State, displayName string) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

	// This addresses the fact that object is retrieved even though it had been deleted
	time.Sleep(1 * time.Second)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "nsxt_edge_high_availability_profile" {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		client := nsx.NewClusterProfilesClient(connector)
		structValue, err := client.Get(resourceID)

		if isNotFoundError(err) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error while retrieving Edge High Availability Profile ID %s. Error: %v", resourceID, err)
		}
		converter := bindings.NewTypeConverter()
		o, errs := converter.ConvertToGolang(structValue, model.EdgeHighAvailabilityProfileBindingType())
		if errs != nil {
			return errs[0]
		}
		obj := o.(model.EdgeHighAvailabilityProfile)

		if obj.DisplayName != nil && displayName == *obj.DisplayName {
			return fmt.Errorf("NSX Edge High Availability Profile %s still exists", displayName)
		}
	}

	return nil
}

func testAccNSXEdgeHighAvailabilityProfileCreateTemplate(displayName string) string {
	return fmt.Sprintf(`
resource "nsxt_edge_high_availability_profile" "test" {
    description  = "Terraform test edge high availability profile"
    display_name = "%s"
    tag {
    	scope = "scope1"
        tag   = "tag1"
    }
}
`, displayName)
}
