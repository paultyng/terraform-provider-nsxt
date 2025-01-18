// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNsxtPolicyFixedSegment() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyFixedSegmentCreate,
		Read:   resourceNsxtPolicyFixedSegmentRead,
		Update: resourceNsxtPolicyFixedSegmentUpdate,
		Delete: resourceNsxtPolicyFixedSegmentDelete,
		Importer: &schema.ResourceImporter{
			State: nsxtGatewayResourceImporter,
		},

		Schema: getPolicyCommonSegmentSchema(false, true),
	}
}

func resourceNsxtPolicyFixedSegmentCreate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentCreate(d, m, false, true)
}

func resourceNsxtPolicyFixedSegmentRead(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentRead(d, m, false, true)
}

func resourceNsxtPolicyFixedSegmentUpdate(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentUpdate(d, m, false, true)
}

func resourceNsxtPolicyFixedSegmentDelete(d *schema.ResourceData, m interface{}) error {
	return nsxtPolicySegmentDelete(d, m, true)
}

func nsxtGatewayResourceImporter(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()
	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		connectivityPath, err := getParameterFromPolicyPath("", "/segments/", importID)
		if err != nil {
			return nil, err
		}
		d.Set("connectivity_path", connectivityPath)
		return rd, nil
	} else if !errors.Is(err, ErrNotAPolicyPath) {
		return rd, err
	}
	importGW := ""
	s := strings.Split(importID, "/")
	if len(s) < 2 {
		return []*schema.ResourceData{d}, fmt.Errorf("Import format gatewayID/segmentID expected, got %s", importID)
	}

	d.SetId(s[1])
	importGW = s[0]

	infra := "infra"
	if isPolicyGlobalManager(m) {
		infra = "global-infra"
	}

	gwPath := fmt.Sprintf("/%s/tier-1s/%s", infra, importGW)
	d.Set("connectivity_path", gwPath)

	return []*schema.ResourceData{d}, nil
}
