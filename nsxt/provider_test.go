// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/vmware/terraform-provider-nsxt/nsxt/util"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	api "github.com/vmware/go-vmware-nsxt"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/core"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/security"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider
var testAccConnector client.Connector

func init() {

	testAccProvider = Provider()
	testAccProviders = map[string]*schema.Provider{
		"nsxt": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	var requiredVariables = []string{"NSXT_USERNAME", "NSXT_PASSWORD", "NSXT_MANAGER_HOST", "NSXT_ALLOW_UNVERIFIED_SSL"}
	for _, element := range requiredVariables {
		if v := os.Getenv(element); v == "" {
			str := fmt.Sprintf("%s must be set for acceptance tests", element)
			t.Fatal(str)
		}
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func testAccGetClient() (*api.APIClient, error) {
	if os.Getenv("NSXT_MANAGER_HOST") == "" {
		return nil, fmt.Errorf("NSXT_MANAGER_HOST is not set in environment")
	}

	client, ok := testAccProvider.Meta().(*api.APIClient)
	if ok {
		return client, nil
	}

	// Try to create a temporary client using the tests configuration
	// This is necessary since the test PreCheck is called before the client is initialized.
	insecure := false
	if v := strings.ToLower(os.Getenv("NSXT_ALLOW_UNVERIFIED_SSL")); v != "false" && v != "0" {
		insecure = true
	}

	cfg := api.Configuration{
		BasePath:   "/api/v1",
		Host:       os.Getenv("NSXT_MANAGER_HOST"),
		Scheme:     "https",
		UserAgent:  "terraform-provider-nsxt/1.0",
		UserName:   os.Getenv("NSXT_USERNAME"),
		Password:   os.Getenv("NSXT_PASSWORD"),
		RemoteAuth: false,
		Insecure:   insecure,
	}

	return api.NewAPIClient(&cfg)
}

func testAccNSXVersion(t *testing.T, requiredVersion string) {
	if util.NsxVersion == "" {
		connector, err := testAccGetPolicyConnector()
		if err != nil {
			t.Errorf("Failed to get policy connector")
			return
		}

		err = initNSXVersion(connector)
		if err != nil {
			t.Errorf("Failed to retrieve NSX version")
			return
		}
	}

	if util.NsxVersionLower(requiredVersion) {
		t.Skipf("This test can only run in NSX %s or above (Current version %s)", requiredVersion, util.NsxVersion)
	}
}

func testAccNSXVersionLessThan(t *testing.T, requiredVersion string) {
	if util.NsxVersion == "" {
		connector, err := testAccGetPolicyConnector()
		if err != nil {
			t.Errorf("Failed to get policy connector")
			return
		}

		err = initNSXVersion(connector)
		if err != nil {
			t.Errorf("Failed to retrieve NSX version")
			return
		}
	}

	if util.NsxVersionHigherOrEqual(requiredVersion) {
		t.Skipf("This test can only run in NSX below %s (Current version %s)", requiredVersion, util.NsxVersion)
	}
}

func testAccGetPolicyConnector() (client.Connector, error) {
	if testAccConnector != nil {
		return testAccConnector, nil
	}

	if os.Getenv("NSXT_MANAGER_HOST") == "" {
		return nil, fmt.Errorf("NSXT_MANAGER_HOST is not set in environment")
	}

	// Try to create a temporary client using the tests configuration
	// This is necessary since the test PreCheck is called before the client is initialized.
	insecure := false
	if v := strings.ToLower(os.Getenv("NSXT_ALLOW_UNVERIFIED_SSL")); v != "false" && v != "0" {
		insecure = true
	}

	hostIP := os.Getenv("NSXT_MANAGER_HOST")
	host := fmt.Sprintf("https://%s", hostIP)
	username := os.Getenv("NSXT_USERNAME")
	password := os.Getenv("NSXT_PASSWORD")

	//TODO: add error handling
	securityCtx := core.NewSecurityContextImpl()
	securityCtx.SetProperty(security.AUTHENTICATION_SCHEME_ID, security.USER_PASSWORD_SCHEME_ID)
	securityCtx.SetProperty(security.USER_KEY, username)
	securityCtx.SetProperty(security.PASSWORD_KEY, password)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
		Proxy:           http.ProxyFromEnvironment,
	}
	httpClient := http.Client{Transport: tr}
	connector := client.NewConnector(host, client.UsingRest(nil), client.WithHttpClient(&httpClient), client.WithSecurityContext(securityCtx))

	testAccConnector = connector

	return testAccConnector, nil
}
