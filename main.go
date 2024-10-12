// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/vmware/terraform-provider-nsxt/nsxt"
)

func main() {
	debugMode := flag.Bool("debug", false, "Enable debug mode for debuggers like Delve (dlv).")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return nsxt.Provider()
		},
	}

	if *debugMode {
		opts.Debug = true
		opts.ProviderAddr = "registry.terraform.io/vmware/nsxt"
	}

	plugin.Serve(opts)
}
