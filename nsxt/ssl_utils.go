// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getSSLProtocolsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "SSL versions TLS1.1 and TLS1.2 are supported and enabled by default. SSLv2, SSLv3, and TLS1.0 are supported, but disabled by default",
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateSSLProtocols(),
		},
		Optional: true,
		Computed: true,
	}
}

func getSSLCiphersSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Description: "Supported SSL cipher list",
		Elem: &schema.Schema{
			Type:         schema.TypeString,
			ValidateFunc: validateSSLCiphers(),
		},
		Optional: true,
		Computed: true,
	}
}

func getIsSecureSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Description: "This flag is set to true when all the ciphers and protocols are secure. It is set to false when one of the ciphers or protocols is insecure",
		Computed:    true,
	}
}

func getCertificateChainDepthSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeInt,
		Description: "Verification depth in the server certificate chain",
		Optional:    true,
		Default:     3,
	}
}
