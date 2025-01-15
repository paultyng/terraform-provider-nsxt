module github.com/vmware/terraform-provider-nsxt

go 1.22.0

toolchain go1.23.3

replace (
	github.com/vmware/vsphere-automation-sdk-go/lib => github.com/vmware/vsphere-automation-sdk-go/lib v0.7.1-0.20241113023437-5938c535c194
	github.com/vmware/vsphere-automation-sdk-go/runtime => github.com/vmware/vsphere-automation-sdk-go/runtime v0.7.1-0.20241113023437-5938c535c194
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt => github.com/vmware/vsphere-automation-sdk-go/services/nsxt v0.12.1-0.20241113023437-5938c535c194
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm => github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm v0.9.1-0.20241113023437-5938c535c194
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp => github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp v0.6.1-0.20241113023437-5938c535c194
)

require (
	github.com/google/uuid v1.6.0
	github.com/hashicorp/go-version v1.7.0
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.35.0
	github.com/stretchr/testify v1.10.0
	github.com/vmware/go-vmware-nsxt v0.0.0-20220328155605-f49a14c1ef5f
	github.com/vmware/vsphere-automation-sdk-go/lib v0.7.0
	github.com/vmware/vsphere-automation-sdk-go/runtime v0.7.1-0.20240611083326-25a4e1834c4d
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt v0.12.0
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-gm v0.9.0
	github.com/vmware/vsphere-automation-sdk-go/services/nsxt-mp v0.6.0
	golang.org/x/exp v0.0.0-20230801115018-d63ba01acd4b
)

require (
	github.com/ProtonMail/go-crypto v1.1.0-alpha.2 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/antihax/optional v1.0.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/beevik/etree v1.1.0 // indirect
	github.com/cloudflare/circl v1.3.7 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/gibson042/canonicaljson-go v1.0.3 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-checkpoint v0.5.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-cty v1.4.1-0.20200414143053-d3edf31b6320 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.6.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.7 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/hc-install v0.9.0 // indirect
	github.com/hashicorp/hcl/v2 v2.22.0 // indirect
	github.com/hashicorp/logutils v1.0.0 // indirect
	github.com/hashicorp/terraform-exec v0.21.0 // indirect
	github.com/hashicorp/terraform-json v0.23.0 // indirect
	github.com/hashicorp/terraform-plugin-go v0.25.0 // indirect
	github.com/hashicorp/terraform-plugin-log v0.9.0 // indirect
	github.com/hashicorp/terraform-registry-address v0.2.3 // indirect
	github.com/hashicorp/terraform-svchost v0.1.1 // indirect
	github.com/hashicorp/yamux v0.1.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/vmihailenco/msgpack v4.0.4+incompatible // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/zclconf/go-cty v1.15.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/mod v0.21.0 // indirect
	golang.org/x/net v0.33.0 // indirect
	golang.org/x/oauth2 v0.22.0 // indirect
	golang.org/x/sync v0.10.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
	google.golang.org/appengine v1.6.8 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240814211410-ddb44dafa142 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
