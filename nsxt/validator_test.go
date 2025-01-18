// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"testing"
)

func TestValidateCidrOrIPOrRange(t *testing.T) {

	cases := map[string]struct {
		value  interface{}
		result bool
	}{
		"NotString": {
			value:  777,
			result: false,
		},
		"Empty": {
			value:  "",
			result: false,
		},
		"IpRangeCidrMix": {
			value:  "1.1.1.1,3.3.3.0/25, 2.2.2.3-2.2.2.21",
			result: false,
		},
		"Range": {
			value:  "2.2.2.3 - 2.2.2.21",
			result: true,
		},
		"NotIP": {
			value:  "1.2.3.no",
			result: false,
		},
		"Text": {
			value:  "frog",
			result: false,
		},
		"ipv4": {
			value:  "192.218.3.1",
			result: true,
		},
		"ipv6": {
			value:  "1231:3213:abcd::15",
			result: true,
		},
		"withSpaces": {
			value:  "20.2.5.5 ",
			result: true,
		},
	}

	validator := validateCidrOrIPOrRange()
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := validator(tc.value, tn)

			if len(errors) > 0 && tc.result {
				t.Errorf("validateCidrOrIPOrRange (%s) produced an unexpected error %s", tc.value, errors)
			} else if len(errors) == 0 && !tc.result {
				t.Errorf("validateCidrOrIPOrRange (%s) did not error", tc.value)
			}
		})
	}
}

func TestValidateCidrOrIPOrRangeList(t *testing.T) {

	cases := map[string]struct {
		value  interface{}
		result bool
	}{
		"NotString": {
			value:  777,
			result: false,
		},
		"Empty": {
			value:  "",
			result: false,
		},
		"IpRangeCidrMix": {
			value:  "1.1.1.1,3.3.3.0/25, 2.2.2.3-2.2.2.21",
			result: true,
		},
		"NotIP": {
			value:  "1.2.3.no",
			result: false,
		},
		"Text": {
			value:  "frog",
			result: false,
		},
		"ipv6": {
			value:  "9870::2-9870::5, 1231:3213:abcd::15",
			result: true,
		},
		"ipVersionMix": {
			value:  "1.1.1.1,3.3.3.0/25, 2.2.2.3-2.2.2.21",
			result: true,
		},
	}

	validator := validateCidrOrIPOrRangeList()
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := validator(tc.value, tn)

			if len(errors) > 0 && tc.result {
				t.Errorf("validateCidrOrIPOrRangeList (%s) produced an unexpected error %s", tc.value, errors)
			} else if len(errors) == 0 && !tc.result {
				t.Errorf("validateCidrOrIPOrRangeList (%s) did not error", tc.value)
			}
		})
	}
}

func TestValidateSingleIP(t *testing.T) {

	cases := map[string]struct {
		value  interface{}
		result bool
	}{
		"NotString": {
			value:  777,
			result: false,
		},
		"Empty": {
			value:  "",
			result: false,
		},
		"ipv4": {
			value:  "192.218.3.1",
			result: true,
		},
		"ipv6": {
			value:  "1231:3213:abcd::15",
			result: true,
		},
		"withSpaces": {
			value:  "20.2.5.5 ",
			result: true,
		},
		"badIP": {
			value:  "192.278.3.1",
			result: true,
		},
	}

	validator := validateSingleIP()
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			_, errors := validator(tc.value, tn)

			if len(errors) > 0 && tc.result {
				t.Errorf("validateSingleIP (%s) produced an unexpected error %s", tc.value, errors)
			} else if len(errors) == 0 && !tc.result {
				t.Errorf("validateSingleIP (%s) did not error", tc.value)
			}
		})
	}
}
