// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Validations for Port objects

func isSinglePort(vi interface{}) bool {
	var i uint64
	switch vi := vi.(type) {
	case int:
		i = uint64(vi)
	case string:
		var err error
		i, err = strconv.ParseUint(vi, 10, 32)
		if err != nil {
			return false
		}
	case uint64:
		i = vi
	default:
		return false
	}
	return i <= 65536
}

func isPortRange(v string) bool {
	s := strings.Split(v, "-")
	if len(s) != 2 {
		return false
	}
	if !isSinglePort(s[0]) || !isSinglePort(s[1]) {
		return false
	}
	return true
}

func validatePortRange() schema.SchemaValidateFunc {
	// A single port num or a range of ports
	return func(v interface{}, k string) (ws []string, errors []error) {
		value := v.(string)
		if !isPortRange(value) && !isSinglePort(value) {
			errors = append(errors, fmt.Errorf(
				"expected %q to be a port range or a single port. Got %s", k, value))
		}
		return
	}
}

func validateSinglePort() schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if !isSinglePort(v) {
			errors = append(errors, fmt.Errorf(
				"expected %q to be a single port number. Got %v", k, v))
		}
		return
	}
}

// Validations for IP objects
func isIPRange(v string) bool {
	s := strings.Split(v, "-")
	if len(s) != 2 {
		return false
	}
	ip1 := net.ParseIP(strings.TrimSpace(s[0]))
	ip2 := net.ParseIP(strings.TrimSpace(s[1]))
	if ip1 == nil || ip2 == nil {
		return false
	}
	return true
}

func isSingleIP(v string) bool {
	v = strings.TrimSpace(v)
	ip := net.ParseIP(v)
	return ip != nil
}

func isCidr(v string, allowMaxPrefix bool, isIP bool) bool {
	v = strings.TrimSpace(v)
	_, ipnet, err := net.ParseCIDR(v)
	if err != nil {
		return false
	}
	if ipnet == nil {
		return false
	}
	if isIP && (v == ipnet.String()) && !allowMaxPrefix {
		return false
	}
	if !isIP && (v != ipnet.String()) {
		return false
	}

	_, bits := ipnet.Mask.Size()
	if !allowMaxPrefix && bits == 1 {
		return false
	}

	return true
}

func validatePortAddress() schema.SchemaValidateFunc {
	// Expects ip_address/prefix (prefix < 32)
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isCidr(v, false, true) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid port address/prefix, got: %s", k, v))
		}
		return
	}
}

func validateCidrOrIPOrRange() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isCidr(v, true, false) && !isSingleIP(v) && !isIPRange(v) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid CIDR or IP or Range, got: %s", k, v))
		}
		return
	}
}

func validateCidrOrIPOrRangeList() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		tokens := strings.Split(v, ",")
		for _, t := range tokens {
			if !isCidr(t, true, false) && !isSingleIP(t) && !isIPRange(t) {
				es = append(es, fmt.Errorf(
					"expected %s to contain a list of valid CIDRs or IPs or Ranges, got: %s", k, t))
				return
			}
		}
		return
	}
}

func validateIPOrRange() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isSingleIP(v) && !isIPRange(v) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP or Range, got: %s", k, v))
		}
		return
	}
}

func validateSingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isSingleIP(v) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP, got: %s", k, v))
		}
		return
	}
}

func validateSingleIPOrHostName() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isSingleIP(v) {
			match, _ := regexp.MatchString(`^[a-zA-Z0-9\.\-]+$`, v)
			if !match {
				es = append(es, fmt.Errorf(
					"expected %s to contain a valid IP or hostname, got: %s", k, v))
			}
		}
		return
	}
}

func isValidStringUint(value string, bits int) bool {
	_, err := strconv.ParseUint(value, 10, bits)
	return (err == nil)
}

func isValidASN(value string) bool {
	tokens := strings.Split(value, ":")
	if len(tokens) != 2 {
		return false
	}
	// ASN is limited to 2 bytes
	if !isValidStringUint(tokens[0], 16) {
		return false
	}

	// Number is limited to 4 bytes
	if !isValidStringUint(tokens[1], 32) {
		return false
	}

	return true
}

func validateASNPair(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if !isValidASN(v) {
		err := fmt.Errorf("expected %s to contain ASN:<number> with ASN limited to 2 bytes, got: %s", k, v)
		es = append(es, err)
	}
	return
}

func validateIPorASNPair(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	err := fmt.Errorf("expected %s to contain <ASN>:<number> or <IP>:<number> with ASN limited to 2 bytes, got: %s", k, v)

	tokens := strings.Split(v, ":")
	if len(tokens) != 2 {
		es = append(es, err)
		return
	}

	if isSingleIP(tokens[0]) {
		if !isValidStringUint(tokens[1], 32) {
			es = append(es, err)
		}
		return
	}

	if !isValidASN(v) {
		es = append(es, err)
	}

	return
}

func validateIPRange() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isIPRange(v) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP range, got: %s", k, v))
		}
		return
	}
}

func validateCidr() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isCidr(v, true, false) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid CIDR, got: %s", k, v))
		}
		return
	}
}

func validateIPCidr() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isCidr(v, true, true) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid CIDR, got: %s", k, v))
		}
		return
	}
}

func isPowerOfTwo(num int) bool {
	for num >= 2 {
		if num%2 != 0 {
			return false
		}
		num = num / 2
	}
	return num == 1
}

func validatePowerOf2(allowZero bool, maxVal int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(int)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be int", k))
			return
		}

		if allowZero && v == 0 {
			return
		}

		if maxVal != 0 && v > maxVal {
			es = append(es, fmt.Errorf(
				"expected %s to be <= %d, got: %d", k, maxVal, v))
			return
		}

		if !isPowerOfTwo(v) {
			es = append(es, fmt.Errorf(
				"expected %s to be a power of 2, got: %d", k, v))
		}
		return
	}
}

var supportedSSLProtocols = []string{"SSL_V2", "SSL_V3", "TLS_V1", "TLS_V1_1", "TLS_V1_2"}

func validateSSLProtocols() schema.SchemaValidateFunc {
	return validation.StringInSlice(supportedSSLProtocols, false)
}

var supportedSSLCiphers = []string{
	"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA",
	"TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA",
	"TLS_ECDH_RSA_WITH_AES_256_CBC_SHA",
	"TLS_RSA_WITH_AES_256_CBC_SHA",
	"TLS_RSA_WITH_AES_128_CBC_SHA",
	"TLS_RSA_WITH_3DES_EDE_CBC_SHA",
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256",
	"TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384",
	"TLS_RSA_WITH_AES_128_CBC_SHA256",
	"TLS_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_RSA_WITH_AES_256_CBC_SHA256",
	"TLS_RSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384",
	"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDH_ECDSA_WITH_AES_128_CBC_SHA256",
	"TLS_ECDH_ECDSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDH_ECDSA_WITH_AES_256_CBC_SHA384",
	"TLS_ECDH_ECDSA_WITH_AES_256_GCM_SHA384",
	"TLS_ECDH_RSA_WITH_AES_128_CBC_SHA",
	"TLS_ECDH_RSA_WITH_AES_128_CBC_SHA256",
	"TLS_ECDH_RSA_WITH_AES_128_GCM_SHA256",
	"TLS_ECDH_RSA_WITH_AES_256_CBC_SHA384",
	"TLS_ECDH_RSA_WITH_AES_256_GCM_SHA384",
}

func validateSSLCiphers() schema.SchemaValidateFunc {
	return validation.StringInSlice(supportedSSLCiphers, false)
}

func validatePolicySourceDestinationGroups() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isCidr(v, true, false) && !isSingleIP(v) && !isIPRange(v) && !isPolicyPath(v) {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP, Range, CIDR, or Group Path. Got: %s", k, v))
		}
		return

	}
}

func validatePolicyPath() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isPolicyPath(v) {
			es = append(es, fmt.Errorf("Invalid policy path: %s", v))
		}

		return
	}
}

func validateID() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		if !isValidID(v) {
			es = append(es, fmt.Errorf("invalid ID atrribute: %s", v))
		}

		return
	}
}

func validateVLANId(i interface{}, k string) (s []string, es []error) {
	var vlan int
	vlan, ok := i.(int)
	if !ok {
		v, ok := i.(string)
		if ok {
			vlanVal, err := strconv.Atoi(v)
			if err != nil {
				es = append(es, fmt.Errorf("cannot convert %s to int", v))
				return
			}
			vlan = vlanVal
		} else {
			es = append(es, fmt.Errorf("Expected VLAN type to be int or string"))
			return
		}
	}
	if vlan < 0 || vlan > 4095 {
		es = append(es, fmt.Errorf("Invalid VLAN ID %d", vlan))
		return
	}
	return
}

func validateVLANIdOrRange(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		s, es = validateVLANId(i, k)
		return
	}

	tokens := strings.Split(v, "-")
	if len(tokens) > 2 {
		es = append(es, fmt.Errorf("Invalid vlan range %s", v))
		return
	}
	for _, token := range tokens {
		s, es = validateVLANId(token, k)
	}

	return
}

func validateStringIntBetween(start int, end int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}
		intVal, err := strconv.Atoi(v)
		if err != nil {
			es = append(es, fmt.Errorf("cannot convert %s to int", v))
			return
		}
		if intVal < start || intVal > end {
			es = append(es, fmt.Errorf("the value %s must be between %v and %v", v, start, end))
		}
		return
	}
}

func validateNsxtProviderHostFormat() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		withSchema := v
		if !strings.HasPrefix(v, "https://") {
			// Add schema for validation
			withSchema = fmt.Sprintf("https://%s", v)
		}

		s, es = validation.IsURLWithHTTPS(withSchema, k)
		return
	}
}

func validateASPlainOrDot(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		es = append(es, fmt.Errorf("String is expected, got %s", v))
		return
	}

	tokens := strings.Split(v, ".")
	if len(tokens) > 2 {
		es = append(es, fmt.Errorf("ASPlain/ASDot format is expected, got %s", v))
		return
	}
	intSize := 32
	if len(tokens) == 2 {
		// Dot notation
		intSize = 16
	}
	for _, token := range tokens {
		_, err := strconv.ParseUint(token, 10, intSize)
		if err != nil {
			es = append(es, fmt.Errorf("%dbit number is expected, got %s", intSize, token))
			return
		}
	}

	return
}

func validateASPath(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		es = append(es, fmt.Errorf("String is expected, got %s", v))
		return
	}

	tokens := strings.Split(v, " ")
	for _, token := range tokens {
		s, es = validateASPlainOrDot(token, k)
	}

	return
}

func validatePolicyBGPCommunity(i interface{}, k string) (s []string, es []error) {
	v, ok := i.(string)
	if !ok {
		es = append(es, fmt.Errorf("String is expected, got %s", v))
		return
	}

	if (v == "NO_EXPORT") || (v == "NO_ADVERTISE") || (v == "NO_EXPORT_SUBCONFED") {
		return
	}

	formatErr := "aa:nn or aa:bb:nn format is expected, got %s"
	tokens := strings.Split(v, ":")
	if (len(tokens) > 3) || (len(tokens) < 2) {
		es = append(es, fmt.Errorf(formatErr, v))
		return
	}

	for _, token := range tokens {
		_, err := strconv.Atoi(token)
		if err != nil {
			es = append(es, fmt.Errorf(formatErr, v))
			return
		}
	}

	return
}

// validateLdapOrLdapsURL( is a SchemaValidateFunc which tests if the url is of type string and a valid LDAP or LDAPs
func validateLdapOrLdapsURL() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		_, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		s, es = validation.IsURLWithScheme([]string{"ldap", "ldaps"})(i, k)
		return
	}
}
