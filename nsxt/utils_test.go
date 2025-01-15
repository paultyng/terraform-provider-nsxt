/* Copyright © 2018 VMware, Inc. All Rights Reserved.
   SPDX-License-Identifier: MPL-2.0 */

package nsxt

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vmware-nsxt/trust"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/bindings"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"

	tf_api "github.com/vmware/terraform-provider-nsxt/api/utl"
)

// Default names or prefixed of NSX backend existing objects used in the acceptance tests.
// Those defaults can be overridden using environment parameters
const tier0RouterDefaultName string = "PLR-1 LogicalRouterTier0"
const edgeClusterDefaultName string = "EDGECLUSTER1"
const vlanTransportZoneName string = "transportzone2"
const overlayTransportZoneNamePrefix string = "1-transportzone"
const macPoolDefaultName string = "DefaultMacPool"
const defaultEnforcementPoint string = "default"

const realizationResourceName string = "data.nsxt_policy_realization_info.realization_info"
const defaultTestResourceName string = "terraform-acctest"

const testAccDataSourceName string = "terraform-acctest-data"
const testAccResourceName string = "terraform-acctest-resource"

const singleTag string = `
  tag {
    scope = "scope1"
    tag   = "tag1"
  }`

const doubleTags string = `
  tag {
    scope = "scope1"
    tag   = "tag1"
  }
  tag {
    scope = "scope2"
    tag   = "tag2"
  }`

var (
	randomized  bool
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	rnd         *rand.Rand
)

func initRand() {
	if randomized {
		return
	}
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
	randomized = true
}

func getAccTestDataSourceName() string {
	initRand()
	return fmt.Sprintf("%s-%d", testAccDataSourceName, rand.Intn(100000))
}

func getAccTestResourceName() string {
	initRand()
	return fmt.Sprintf("%s-%d", testAccResourceName, rand.Intn(100000))
}

func getAccTestRandomString(length int) string {
	initRand()
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(length)]
	}
	return string(b)
}

func getAccTestFQDN() string {
	return fmt.Sprintf("test.%s.org", getAccTestRandomString(10))
}

func getTier0RouterName() string {
	name := os.Getenv("NSXT_TEST_TIER0_ROUTER")
	if name == "" {
		name = tier0RouterDefaultName
	}
	return name
}

func getTier0RouterPath(connector client.Connector) string {
	// Retrieve Tier0 path
	routerName := getTier0RouterName()
	t0client := infra.NewTier0sClient(connector)
	tier0, _ := t0client.Get(routerName)

	return *tier0.Path
}

func getEdgeClusterName() string {
	name := os.Getenv("NSXT_TEST_EDGE_CLUSTER")
	if name == "" {
		name = edgeClusterDefaultName
	}
	return name
}

func getComputeCollectionName() string {
	return os.Getenv("NSXT_TEST_COMPUTE_COLLECTION")
}

func getComputeManagerName() string {
	return os.Getenv("NSXT_TEST_COMPUTE_MANAGER")
}

func getHostTransportNodeName() string {
	return os.Getenv("NSXT_TEST_HOST_TRANSPORT_NODE")
}

func getHostTransportNodeProfileName() string {
	return os.Getenv("NSXT_TEST_HOST_TRANSPORT_NODE_PROFILE")
}

func getEdgeTransportNodeName() string {
	return os.Getenv("NSXT_TEST_EDGE_TRANSPORT_NODE")
}

func getVlanTransportZoneName() string {
	name := os.Getenv("NSXT_TEST_VLAN_TRANSPORT_ZONE")
	if name == "" {
		name = vlanTransportZoneName
	}
	return name
}

func getOverlayTransportZoneName() string {
	name := os.Getenv("NSXT_TEST_OVERLAY_TRANSPORT_ZONE")
	if name == "" {
		name = overlayTransportZoneNamePrefix
	}
	return name
}

func getMacPoolName() string {
	name := os.Getenv("NSXT_TEST_MAC_POOL")
	if name == "" {
		name = macPoolDefaultName
	}
	return name
}

func getIPPoolName() string {
	return os.Getenv("NSXT_TEST_IP_POOL")
}

func getTestVMID() string {
	return os.Getenv("NSXT_TEST_VM_ID")
}

func getTestVMSegmentID() string {
	return os.Getenv("NSXT_TEST_VM_SEGMENT_ID")
}

func getTestVMName() string {
	return os.Getenv("NSXT_TEST_VM_NAME")
}

func getTestSiteName() string {
	return os.Getenv("NSXT_TEST_SITE_NAME")
}

func getTestAnotherSiteName() string {
	return os.Getenv("NSXT_TEST_ANOTHER_SITE_NAME")
}

func getTestCertificateName(isClient bool) string {
	if isClient {
		return os.Getenv("NSXT_TEST_CLIENT_CERTIFICATE_NAME")
	}
	return os.Getenv("NSXT_TEST_CERTIFICATE_NAME")
}

func getTestLBServiceName() string {
	return os.Getenv("NSXT_TEST_LB_SERVICE_NAME")
}

func getTestLdapUser() string {
	return os.Getenv("NSXT_TEST_LDAP_USER")
}

func getTestLdapPassword() string {
	return os.Getenv("NSXT_TEST_LDAP_PASSWORD")
}

func getTestLdapURL() string {
	return os.Getenv("NSXT_TEST_LDAP_URL")
}

func getTestLdapCert() string {
	return os.Getenv("NSXT_TEST_LDAP_CERT")
}

func getTestLdapDomain() string {
	return os.Getenv("NSXT_TEST_LDAP_DOMAIN")
}

func getTestLdapBaseDN() string {
	return os.Getenv("NSXT_TEST_LDAP_BASE_DN")
}

func getTestManagerClusterNode() string {
	return os.Getenv("NSXT_TEST_MANAGER_CLUSTER_NODE")
}

func testAccEnvDefined(t *testing.T, envVar string) {
	if len(os.Getenv(envVar)) == 0 {
		t.Skipf("This test requires %s environment variable to be set", envVar)
	}
}

func testAccIsGlobalManager() bool {
	return os.Getenv("NSXT_GLOBAL_MANAGER") == "true" || os.Getenv("NSXT_GLOBAL_MANAGER") == "1"
}

func testAccIsMultitenancy() bool {
	return os.Getenv("NSXT_PROJECT_ID") != ""
}

func testAccIsVPC() bool {
	return os.Getenv("NSXT_VPC_PROJECT_ID") != "" && os.Getenv("NSXT_VPC_ID") != ""
}

func testAccIsFabric() bool {
	return os.Getenv("NSXT_TEST_FABRIC") != ""
}

func testAccGetSessionContext() tf_api.SessionContext {
	clientType := testAccIsGlobalManager2()
	projectID := os.Getenv("NSXT_PROJECT_ID")
	vpcID := ""
	if clientType == tf_api.VPC {
		projectID = os.Getenv("NSXT_VPC_PROJECT_ID")
		vpcID = os.Getenv("NSXT_VPC_ID")
	}
	return tf_api.SessionContext{ProjectID: projectID, ClientType: clientType, VPCID: vpcID}
}

func testAccGetSessionProjectContext() tf_api.SessionContext {
	clientType := testAccIsGlobalManager2()
	projectID := os.Getenv("NSXT_PROJECT_ID")
	vpcID := ""
	return tf_api.SessionContext{ProjectID: projectID, ClientType: clientType, VPCID: vpcID}
}

func testAccGetProjectContext() tf_api.SessionContext {
	projectID := os.Getenv("NSXT_PROJECT_ID")
	return tf_api.SessionContext{ProjectID: projectID, ClientType: tf_api.Multitenancy}
}

func testAccIsGlobalManager2() tf_api.ClientType {
	if testAccIsVPC() {
		return tf_api.VPC
	}
	if os.Getenv("NSXT_PROJECT_ID") != "" {
		return tf_api.Multitenancy
	}
	if os.Getenv("NSXT_GLOBAL_MANAGER") == "true" || os.Getenv("NSXT_GLOBAL_MANAGER") == "1" {
		return tf_api.Global
	}
	return tf_api.Local
}

func testAccNotGlobalManager(t *testing.T) {
	if testAccIsGlobalManager() {
		t.Skipf("This test requires a global manager environment")
	}
}

func testAccOnlyGlobalManager(t *testing.T) {
	if !testAccIsGlobalManager() {
		t.Skipf("This test requires a global manager environment")
	}
}

func testAccOnlyLocalManager(t *testing.T) {
	if testAccIsGlobalManager() || testAccIsMultitenancy() {
		t.Skipf("This test requires a local manager environment")
	}
}

func testAccTestFabric(t *testing.T) {
	if !testAccIsFabric() {
		t.Skipf("Fabric testing is not enabled")
	}
}

func getTestVCUsername() string {
	return os.Getenv("NSXT_TEST_VC_USERNAME")
}

func getTestVCPassword() string {
	return os.Getenv("NSXT_TEST_VC_PASSWORD")
}

func getTestVCIPAddress() string {
	return os.Getenv("NSXT_TEST_VC_IPADDRESS")
}

func getTestVCThumbprint() string {
	return os.Getenv("NSXT_TEST_VC_THUMBPRINT")
}

func testAccTestVCCredentials(t *testing.T) {
	if getTestVCUsername() == "" || getTestVCPassword() == "" || getTestVCIPAddress() == "" || getTestVCThumbprint() == "" {
		t.Skipf("This test requires a vCenter configuration environment")
	}
}

func testAccOnlyMultitenancy(t *testing.T) {
	testAccNSXVersion(t, "4.1.0")
	if !testAccIsMultitenancy() {
		t.Skipf("This test requires a multitenancy environment")
	}
}

func testAccOnlyVPC(t *testing.T) {
	testAccNSXVersion(t, "4.1.2")
	if !testAccIsVPC() {
		t.Skipf("This test requires a VPC environment")
	}
}

func testAccNSXGlobalManagerSitePrecheck(t *testing.T) {
	if testAccIsGlobalManager() && getTestSiteName() == "" {
		str := fmt.Sprintf("%s must be set for this acceptance test", "NSXT_TEST_SITE_NAME")
		t.Fatal(str)
	}
}

func testAccTestDeprecated(t *testing.T) {
	if os.Getenv("NSXT_TEST_DEPRECATED") != "true" && os.Getenv("NSXT_TEST_DEPRECATED") != "1" {
		t.Skipf("To run deprecated test suite, please enable NSXT_TEST_DEPRECATED in your environment.")
	}
}

// Create and delete CA and client cert for various tests
func testAccNSXCreateCert(t *testing.T, name string, certPem string, certPK string, certType string) string {
	nsxClient, err := testAccGetClient()
	if err != nil {
		t.Fatal(err)
	}

	object := trust.TrustObjectData{
		DisplayName:  name,
		ResourceType: certType,
		PemEncoded:   certPem,
		PrivateKey:   certPK,
	}

	certList, response, err := nsxClient.NsxComponentAdministrationApi.AddCertificateImport(nsxClient.Context, object)

	if err != nil {
		t.Fatalf("Error while creating %s certificate. Error: %v", certType, err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("Error while creating %s certificate. HTTP return code %d", certType, response.StatusCode)
	}
	certID := ""
	for _, cert := range certList.Results {
		certID = cert.Id
	}

	return certID
}

func testAccNSXDeleteCert(t *testing.T, id string) {
	nsxClient, err := testAccGetClient()
	if err != nil {
		t.Fatal(err)
	}

	response, err := nsxClient.NsxComponentAdministrationApi.DeleteCertificate(nsxClient.Context, id)

	if err != nil {
		t.Fatalf("Error while deleting certificate %s. Error: %v", id, err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("Error while deleting certificate %s. HTTP return code %d", id, response.StatusCode)
	}
}

func testAccNSXCreateCerts(t *testing.T) (string, string, string) {

	certPem := "-----BEGIN CERTIFICATE-----\n" +
		"MIICVjCCAb8CAg37MA0GCSqGSIb3DQEBBQUAMIGbMQswCQYDVQQGEwJKUDEOMAwG\n" +
		"A1UECBMFVG9reW8xEDAOBgNVBAcTB0NodW8ta3UxETAPBgNVBAoTCEZyYW5rNERE\n" +
		"MRgwFgYDVQQLEw9XZWJDZXJ0IFN1cHBvcnQxGDAWBgNVBAMTD0ZyYW5rNEREIFdl\n" +
		"YiBDQTEjMCEGCSqGSIb3DQEJARYUc3VwcG9ydEBmcmFuazRkZC5jb20wHhcNMTIw\n" +
		"ODIyMDUyNzIzWhcNMTcwODIxMDUyNzIzWjBKMQswCQYDVQQGEwJKUDEOMAwGA1UE\n" +
		"CAwFVG9reW8xETAPBgNVBAoMCEZyYW5rNEREMRgwFgYDVQQDDA93d3cuZXhhbXBs\n" +
		"ZS5jb20wgZ8wDQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBAMYBBrx5PlP0WNI/ZdzD\n" +
		"+6Pktmurn+F2kQYbtc7XQh8/LTBvCo+P6iZoLEmUA9e7EXLRxgU1CVqeAi7QcAn9\n" +
		"MwBlc8ksFJHB0rtf9pmf8Oza9E0Bynlq/4/Kb1x+d+AyhL7oK9tQwB24uHOueHi1\n" +
		"C/iVv8CSWKiYe6hzN1txYe8rAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAASPdjigJ\n" +
		"kXCqKWpnZ/Oc75EUcMi6HztaW8abUMlYXPIgkV2F7YanHOB7K4f7OOLjiz8DTPFf\n" +
		"jC9UeuErhaA/zzWi8ewMTFZW/WshOrm3fNvcMrMLKtH534JKvcdMg6qIdjTFINIr\n" +
		"evnAhf0cwULaebn+lMs8Pdl7y37+sfluVok=\n" +
		"-----END CERTIFICATE-----"

	certPKPem := "-----BEGIN RSA PRIVATE KEY-----\n" +
		"MIICWwIBAAKBgQDGAQa8eT5T9FjSP2Xcw/uj5LZrq5/hdpEGG7XO10IfPy0wbwqP\n" +
		"j+omaCxJlAPXuxFy0cYFNQlangIu0HAJ/TMAZXPJLBSRwdK7X/aZn/Ds2vRNAcp5\n" +
		"av+Pym9cfnfgMoS+6CvbUMAduLhzrnh4tQv4lb/AkliomHuoczdbcWHvKwIDAQAB\n" +
		"AoGAXzxrIwgmBHeIqUe5FOBnDsOZQlyAQA+pXYjCf8Rll2XptFwUdkzAUMzWUGWT\n" +
		"G5ZspA9l8Wc7IozRe/bhjMxuVK5yZhPDKbjqRdWICA95Jd7fxlIirHOVMQRdzI7x\n" +
		"NKqMNQN05MLJfsEHUYtOLhZE+tfhJTJnnmB7TMwnJgc4O5ECQQD8oOJ45tyr46zc\n" +
		"OAt6ao7PefVLiW5Qu+PxfoHmZmDV2UQqeM5XtZg4O97VBSugOs3+quIdAC6LotYl\n" +
		"/6N+E4y3AkEAyKWD2JNCrAgtjk2bfF1HYt24tq8+q7x2ek3/cUhqwInkrZqOFoke\n" +
		"x3+yBB879TuUOadvBXndgMHHcJQKSAJlLQJAXRuGnHyptAhTe06EnHeNbtZKG67p\n" +
		"I4Q8PJMdmSb+ZZKP1v9zPUxGb+NQ+z3OmF1T8ppUf8/DV9+KAbM4NI1L/QJAdGBs\n" +
		"BKYFObrUkYE5+fwwd4uao3sponqBTZcH3jDemiZg2MCYQUHu9E+AdRuYrziLVJVk\n" +
		"s4xniVLb1tRG0lVxUQJASfjdGT81HDJSzTseigrM+JnBKPPrzpeEp0RbTP52Lm23\n" +
		"YARjLCwmPMMdAwYZsvqeTuHEDQcOHxLHWuyN/zgP2A==\n" +
		"-----END RSA PRIVATE KEY-----"

	certID := testAccNSXCreateCert(t, "test", certPem, certPKPem, "certificate_signed")

	clientCertPem := "-----BEGIN CERTIFICATE-----\n" +
		"MIID6jCCAtKgAwIBAgIJAOtKKdMP6oZcMA0GCSqGSIb3DQEBCwUAMHYxCzAJBgNV\n" +
		"BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX\n" +
		"aWRnaXRzIFB0eSBMdGQxETAPBgNVBAMMCHRlc3QuY29tMRwwGgYJKoZIhvcNAQkB\n" +
		"Fg10ZXN0QHRlc3QuY29tMB4XDTE4MDUwNzE3MTkxOVoXDTE5MDUwNzE3MTkxOVow\n" +
		"djELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGElu\n" +
		"dGVybmV0IFdpZGdpdHMgUHR5IEx0ZDERMA8GA1UEAwwIdGVzdC5jb20xHDAaBgkq\n" +
		"hkiG9w0BCQEWDXRlc3RAdGVzdC5jb20wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw\n" +
		"ggEKAoIBAQC8GhgSX8nV0QEStgQ208PDXsl62xefAWVYMVjF9+xUWHY6zp/sTTQA\n" +
		"uFGD6DBiyljYYShj85ha7N3m70Xmur5Bbpa7ca010R0oCLPQRcySvsom190OfeUG\n" +
		"LW396gZUjiY9+8OQqI8BrTcNkI1DOpC5fBbbXkvLN4pOIf+/JqZNtr1igQBv2g+E\n" +
		"FCiPxDuWPOagi2xeEQv3SLLBhfF92UAlV6JRziCwaILlC3Zn/6B8E8rpuqVxUShW\n" +
		"wDmLowoT4BZP6/OHRneyEvVi2L/Ucsdk3nwUTe5Q4ojHY21ftW6by6uXYvK4IePC\n" +
		"ZXeTjdNtRX4MaAtJ4iLY7E/9d0BdTNRvAgMBAAGjezB5MAkGA1UdEwQCMAAwLAYJ\n" +
		"YIZIAYb4QgENBB8WHU9wZW5TU0wgR2VuZXJhdGVkIENlcnRpZmljYXRlMB0GA1Ud\n" +
		"DgQWBBTP56rycJh9hpWKOmvTzYg3YEkqiDAfBgNVHSMEGDAWgBTP56rycJh9hpWK\n" +
		"OmvTzYg3YEkqiDANBgkqhkiG9w0BAQsFAAOCAQEAZ90xoDf4gRbh7/PHxbokP69K\n" +
		"uqt7s78JDKrGxCDwiezUhZrdOBwi2r/sg4cWi43uuwCfjGNCd824EQYRaSCjanAn\n" +
		"5OH14KshCOBi66CaWDzJK6v4X/hbrKtUmXvbvUjrQCEHVuLueEQfvJB5/8O0dpEA\n" +
		"xF4MhnSaF2Id5O7tlXnFoXZh0YI9QnTwHLQ4L9+3PS5LqWd0peV1XqgWy0CXwjZF\n" +
		"nEpHq+TGDwLRoAgnoBGrbaFJRmvm+iVU4J76AtV7B3keckVMyMIeBR9CWB7kDm64\n" +
		"86qiRcEGN7V5mMJtxF49l0F01qdOgrictZRf+gMMrtGmX4KkZ6DKrl278HPs7A==\n" +
		"-----END CERTIFICATE-----"

	clientPKPem := "-----BEGIN PRIVATE KEY-----\n" +
		"MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC8GhgSX8nV0QES\n" +
		"tgQ208PDXsl62xefAWVYMVjF9+xUWHY6zp/sTTQAuFGD6DBiyljYYShj85ha7N3m\n" +
		"70Xmur5Bbpa7ca010R0oCLPQRcySvsom190OfeUGLW396gZUjiY9+8OQqI8BrTcN\n" +
		"kI1DOpC5fBbbXkvLN4pOIf+/JqZNtr1igQBv2g+EFCiPxDuWPOagi2xeEQv3SLLB\n" +
		"hfF92UAlV6JRziCwaILlC3Zn/6B8E8rpuqVxUShWwDmLowoT4BZP6/OHRneyEvVi\n" +
		"2L/Ucsdk3nwUTe5Q4ojHY21ftW6by6uXYvK4IePCZXeTjdNtRX4MaAtJ4iLY7E/9\n" +
		"d0BdTNRvAgMBAAECggEAYS8KKNQcr7/gUg6AduNKTXJ3nmX7+kb6WWqFdyL0k09x\n" +
		"JkkDD0+EAqs4UnJwLVpmNMVd3OZKXQ/sRhhxgRgSnDPK5OWCnD+CVODKJl0pqNey\n" +
		"EgeNSqN45IwsO/fhdWZME9Iz5FVyLWeU/gklMwrbIzodhRFfD4uOhXfDbrtFSPix\n" +
		"KGzhDNCn3z/Y9Sml6tUMgCLPisQbxWcQ+bbTHCgXXEdK2yG45j4q3FqBGy8IExou\n" +
		"cxw0aWu8cIqIGlXBQA5SZ78csBisKxEmlHjkV1i7Jn2DIARUluU/nmYKGC4ElWGp\n" +
		"yuURRJhcjO17MkCEr+niTxnTptkHi9e1p4d/k6HCgQKBgQD0Uj/JPr9vIM6Mt3vm\n" +
		"kK0GSTfpMUDDwaAMR+Pip6qocEYD5nZ6rnSuXIqlKyP/YxtQqsDqFdYHbPJet/Lg\n" +
		"3HFKerQ7xHiral95rsP0TWiRtxCHaF1xahgQWlVaYqGDeBcL7EavGJwQy3crFiVo\n" +
		"6wHwbVfNV/hoXHHK7pX4J8M3sQKBgQDFF+P/gM86QrvDbJnH/7PuF6QzcdvvtacH\n" +
		"i+5aTG9YQaUGldQIx2Fjpn48V73/YGK04DTrXqACjE2XxgEu0/B0wAGkOwGZztAr\n" +
		"q54O4AEWgVAyUukarkjdGtgeUAf/UMeXf7D2tGhjyNtORrOyZCdKS3kZ3GfZuEXj\n" +
		"FjhbJbX2HwKBgCMh8Kovq7d/MDRr7hUpmLfer3uI6Zc8sJcTf2GIWrH98xN8gG0D\n" +
		"ySOJiyZVHcgLqFHhO/xtR2mp8PBN408SY/ghzOkLR47erPwCdYsb1n2dpXLTPxyf\n" +
		"9PXlB4EHzdHp4uaEA2YKU+bWWzyG4rpDkPPRxV5x1/ap1HMp+8bDcP8BAoGAJTdQ\n" +
		"pxNUjgTB3bHpC9ndyOyP5eLvC8F6S7OBi215bOngVnD+O7YiTqXGmnBbARjbKppX\n" +
		"g8Y3YqPJlwodeREuC22iIbe+oqNprYVXcCmeKvi6Avai65XTTmTeQEMOb4h6V8IV\n" +
		"0U/ZklYACzTQg7Pjs2Sy9k4nEfZ4w9uTQqrJRDMCgYEAvgB29WCwEJ81pjBIGiYw\n" +
		"FbqLxQtIFiXKvRW+Qr1EKcV89kUYvsH3+TO/hZpf6rBtRf9yLJvw2fR4tVouowbn\n" +
		"/w+mKvKmcm11NOuobla4w9gztptqeUzcW03JmHmcnyHlnnUwpsU/XcgnsToV6kJB\n" +
		"bS6lN3HXkJnFnAzK2BKoZCA=\n" +
		"-----END PRIVATE KEY-----\n"

	clientCertID := testAccNSXCreateCert(t, "test_client", clientCertPem, clientPKPem, "certificate_self_signed")

	caCertPem := "-----BEGIN CERTIFICATE-----\n" +
		"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\n" +
		"-----END CERTIFICATE-----"

	caCertID := testAccNSXCreateCert(t, "test_ca", caCertPem, "", "certificate_ca")

	return certID, clientCertID, caCertID
}

func testAccNSXDeleteCerts(t *testing.T, certID string, clientCertID string, caCertID string) {

	testAccNSXDeleteCert(t, certID)
	testAccNSXDeleteCert(t, clientCertID)
	testAccNSXDeleteCert(t, caCertID)
}

func testGetObjIDByName(objName string, resourceType string) (string, error) {
	connector, err1 := testAccGetPolicyConnector()
	if err1 != nil {
		return "", fmt.Errorf("Error during test client initialization: %v", err1)
	}

	resultValues, err2 := listPolicyResourcesByNameAndType(connector, testAccGetSessionContext(), objName, resourceType, nil)
	if err2 != nil {
		return "", err2
	}

	converter := bindings.NewTypeConverter()

	for _, result := range resultValues {
		dataValue, errors := converter.ConvertToGolang(result, model.PolicyResourceBindingType())
		if len(errors) > 0 {
			return "", errors[0]
		}
		policyResource := dataValue.(model.PolicyResource)

		if *policyResource.DisplayName == objName {
			return *policyResource.Id, nil
		}
	}

	return "", fmt.Errorf("%s with name '%s' was not found", resourceType, objName)
}

func testAccNsxtGlobalPolicySite(domainName string) string {
	return fmt.Sprintf(`
data "nsxt_policy_site" "test" {
  display_name = "%s"
}`, domainName)
}

func testAccAdjustPolicyInfraConfig(config string) string {
	if testAccIsGlobalManager() {
		return strings.Replace(config, "/infra/", "/global-infra/", -1)
	}

	return config
}

func testAccNsxtPolicyTier0WithEdgeClusterTemplate(edgeClusterName string, standby bool) string {
	return testAccNsxtPolicyGatewayWithEdgeClusterTemplate(edgeClusterName, true, standby, false)
}

func testAccNsxtPolicyTier1WithEdgeClusterTemplate(edgeClusterName string, standby, withContext bool) string {
	return testAccNsxtPolicyGatewayWithEdgeClusterTemplate(edgeClusterName, false, standby, withContext)
}

func testAccNsxtPolicyGatewayWithEdgeClusterTemplate(edgeClusterName string, tier0, standby, withContext bool) string {
	var tier string
	if tier0 {
		tier = "0"
	} else {
		tier = "1"
	}
	var haMode string
	if standby && tier0 {
		haMode = fmt.Sprintf(`ha_mode = "%s"`, model.Tier0_HA_MODE_STANDBY)
	}
	if testAccIsGlobalManager() {
		return fmt.Sprintf(`
resource "nsxt_policy_tier%s_gateway" "test" {
  display_name = "terraform-t%s-gw"
  description  = "Acceptance Test"
  locale_service {
    edge_cluster_path = data.nsxt_policy_edge_cluster.%s.path
  }
  %s
}`, tier, tier, edgeClusterName, haMode)
	}
	context := ""
	if withContext {
		context = testAccNsxtPolicyMultitenancyContext()
	}
	return fmt.Sprintf(`
resource "nsxt_policy_tier%s_gateway" "test" {
%s
  display_name      = "terraform-t%s-gw"
  description       = "Acceptance Test"
  edge_cluster_path = data.nsxt_policy_edge_cluster.%s.path
  %s
}`, tier, context, tier, edgeClusterName, haMode)
}

var testAccNsxtPolicyVPNGatewayHelperName = getAccTestResourceName()

func testAccNsxtPolicyTier0WithEdgeClusterForVPN() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) + fmt.Sprintf(`
resource "nsxt_policy_tier0_gateway" "test" {
  display_name = "%s"
  description  = "Acceptance Test"
  locale_service {
    edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
  }
  ha_mode = "ACTIVE_STANDBY"
}`, testAccNsxtPolicyVPNGatewayHelperName)
}

func testAccNsxtPolicyTier1WithEdgeClusterForVPN() string {
	return testAccNsxtPolicyEdgeClusterReadTemplate(getEdgeClusterName()) + fmt.Sprintf(`
resource "nsxt_policy_tier0_gateway" "test" {
	display_name = "%s"
	description  = "Acceptance Test"
	ha_mode = "ACTIVE_STANDBY"
}
resource "nsxt_policy_tier1_gateway" "test" {
	description               = "Acceptance Test"
	display_name              = "%s"
	locale_service {
		edge_cluster_path = data.nsxt_policy_edge_cluster.test.path
	}
	tier0_path                = nsxt_policy_tier0_gateway.test.path
}`, testAccNsxtPolicyVPNGatewayHelperName, testAccNsxtPolicyVPNGatewayHelperName)
}

func testAccNsxtPolicyGatewayTemplate(isT0 bool) string {
	if isT0 {
		return testAccNsxtPolicyTier0WithEdgeClusterForVPN()
	}
	return testAccNsxtPolicyTier1WithEdgeClusterForVPN()
}

func testAccNsxtPolicyResourceExists(context tf_api.SessionContext, resourceName string, presenceChecker func(tf_api.SessionContext, string, client.Connector) (bool, error)) resource.TestCheckFunc {
	return func(state *terraform.State) error {

		connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))

		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Policy resource %s not found in resources", resourceName)
		}

		resourceID := rs.Primary.ID
		if resourceID == "" {
			return fmt.Errorf("Policy resource ID not set in resources")
		}

		exists, err := presenceChecker(context, resourceID, connector)
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Policy resource %s does not exist", resourceID)
		}

		return nil
	}
}

func testAccNsxtPolicyResourceCheckDestroy(context tf_api.SessionContext, state *terraform.State, displayName string, resourceType string, presenceChecker func(tf_api.SessionContext, string, client.Connector) (bool, error)) error {
	connector := getPolicyConnector(testAccProvider.Meta().(nsxtClients))
	for _, rs := range state.RootModule().Resources {

		if rs.Type != resourceType {
			continue
		}

		resourceID := rs.Primary.Attributes["id"]
		exists, err := presenceChecker(context, resourceID, connector)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("Policy resource %s still exists", displayName)
		}
	}
	return nil
}

func testAccNsxtPolicyMultitenancyContext() string {
	if testAccIsVPC() {
		projectID := os.Getenv("NSXT_VPC_PROJECT_ID")
		vpcID := os.Getenv("NSXT_VPC_ID")
		return fmt.Sprintf(`
  context {
    project_id = "%s"
    vpc_id     = "%s"
  }
`, projectID, vpcID)
	}
	projectID := os.Getenv("NSXT_PROJECT_ID")
	if projectID != "" {
		return fmt.Sprintf(`
  context {
    project_id = "%s"
  }
`, projectID)
	}
	return ""
}

func testAccNsxtProjectContext() string {
	projectID := os.Getenv("NSXT_PROJECT_ID")
	return fmt.Sprintf(`
  context {
    project_id = "%s"
  }
`, projectID)
}

func testAccNsxtMultitenancyContext(includeVpc bool) string {
	if testAccIsVPC() {
		// Some tests run in VPC context, however dependency resources are
		// not under VPC. In this case, we rely on VPC env configuration
		// but need to only list project in the context
		projectID := os.Getenv("NSXT_VPC_PROJECT_ID")
		if !includeVpc {
			return fmt.Sprintf(`
  context {
    project_id = "%s"
  }
`, projectID)
		}
		// VPC resource
		vpcID := os.Getenv("NSXT_VPC_ID")
		return fmt.Sprintf(`
  context {
    project_id = "%s"
    vpc_id     = "%s"
  }
`, projectID, vpcID)
	}
	// CLassic Multi Tenancy resource
	projectID := os.Getenv("NSXT_PROJECT_ID")
	if projectID != "" {
		return fmt.Sprintf(`
  context {
    project_id = "%s"
  }
`, projectID)
	}
	return ""
}

func testAccResourceNsxtPolicyImportIDRetriever(resourceID string) func(*terraform.State) (string, error) {
	return func(s *terraform.State) (string, error) {

		rs, ok := s.RootModule().Resources[resourceID]
		if !ok {
			return "", fmt.Errorf("NSX Policy %s resource not found in resources", resourceID)
		}
		path := rs.Primary.Attributes["path"]
		if path == "" {
			return "", fmt.Errorf("NSX Policy %s path not set in resources ", resourceID)
		}
		return path, nil
	}
}

func testAccGenerateTLSKeyPair() (string, string, error) {
	// Ref: https://go.dev/src/crypto/tls/generate_cert.go
	var publicPem, privatePem string
	priv, err := ecdsa.GenerateKey(elliptic.P384(), cryptorand.Reader)
	if err != nil {
		return publicPem, privatePem, err
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 180),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(cryptorand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return publicPem, privatePem, err
	}
	buf := &bytes.Buffer{}

	if err := pem.Encode(buf, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return publicPem, privatePem, err
	}
	publicPem = buf.String()
	buf.Reset()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return publicPem, privatePem, err
	}
	if err := pem.Encode(buf, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return publicPem, privatePem, err
	}
	privatePem = buf.String()
	return publicPem, privatePem, nil
}

func testAccNsxtProjectShareAll(sharedResourcePath string) string {
	name := getAccTestResourceName()
	projectPath := fmt.Sprintf("/orgs/default/projects/%s", os.Getenv("NSXT_VPC_PROJECT_ID"))
	context := testAccNsxtMultitenancyContext(false)
	return fmt.Sprintf(`
resource "nsxt_policy_share" "test" {
%s
  display_name = "%s"

  sharing_strategy = "ALL_DESCENDANTS"
  shared_with      = ["%s"]
}

resource "nsxt_policy_shared_resource" "test" {
  display_name = "%s"

  share_path   = nsxt_policy_share.test.path
  resource_object {
    resource_path    = %s
    include_children = true
  }
}`, context, name, projectPath, name, sharedResourcePath)
}
