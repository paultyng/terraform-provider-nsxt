// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package nsxt

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/infra/tier_0s"
	"github.com/vmware/vsphere-automation-sdk-go/services/nsxt/model"
)

var routeAdvertisementTypesValues = []string{
	"TIER0_STATIC",
	"TIER0_CONNECTED",
	"TIER0_NAT",
	"TIER0_DNS_FORWARDER_IP",
	"TIER0_IPSEC_LOCAL_ENDPOINT",
	"TIER1_STATIC",
	"TIER1_CONNECTED",
	"TIER1_LB_SNAT",
	"TIER1_LB_VIP",
	"TIER1_NAT",
	"TIER1_DNS_FORWARDER_IP",
	"TIER1_IPSEC_LOCAL_ENDPOINT",
}

func resourceNsxtPolicyTier0InterVRFRouting() *schema.Resource {
	return &schema.Resource{
		Create: resourceNsxtPolicyTier0InterVRFRoutingCreate,
		Read:   resourceNsxtPolicyTier0InterVRFRoutingRead,
		Update: resourceNsxtPolicyTier0InterVRFRoutingUpdate,
		Delete: resourceNsxtPolicyTier0InterVRFRoutingDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNsxtPolicyTier0InterVRFRoutingImport,
		},
		Schema: map[string]*schema.Schema{
			"nsx_id":       getNsxIDSchema(),
			"path":         getPathSchema(),
			"display_name": getDisplayNameSchema(),
			"description":  getDescriptionSchema(),
			"revision":     getRevisionSchema(),
			"tag":          getTagsSchema(),
			"gateway_path": getPolicyPathSchema(true, true, "Policy path for the Gateway"),
			"bgp_route_leaking": {
				Type:        schema.TypeList,
				Description: "Import / export BGP routes",
				Optional:    true,
				MaxItems:    2,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_family": {
							Type:         schema.TypeString,
							Description:  "Address family type",
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"IPV4", "IPV6"}, false),
						},
						"in_filter": {
							Type:        schema.TypeList,
							Description: "route map path for IN direction",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"out_filter": {
							Type:        schema.TypeList,
							Description: "route map path for OUT direction",
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"static_route_advertisement": {
				Type:        schema.TypeList,
				Description: "Advertise subnet to target peers as static routes",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"advertisement_rule": {
							Type:        schema.TypeList,
							Description: "Route advertisement rules",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:         schema.TypeString,
										Description:  "Action to advertise routes",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"PERMIT", "DENY"}, false),
										Default:      "PERMIT",
									},
									"name": {
										Type:        schema.TypeString,
										Description: "Display name for rule",
										Optional:    true,
									},
									"prefix_operator": {
										Type:         schema.TypeString,
										Description:  "Prefix operator to match subnets",
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"GE", "EQ"}, false),
										Default:      "GE",
									},
									"route_advertisement_types": {
										Type:        schema.TypeList,
										Description: "Enable different types of route advertisements",
										Optional:    true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(routeAdvertisementTypesValues, false),
										},
									},
									"subnets": {
										Type:        schema.TypeList,
										Description: "Network CIDRs",
										Optional:    true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validateCidr(),
										},
									},
								},
							},
						},
						"in_filter_prefix_list": {
							Type:        schema.TypeList,
							Description: "Paths of ordered Prefix list",
							Optional:    true,
							MaxItems:    5,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"target_path": getPolicyPathSchema(true, false, "Policy path to tier0/vrf belongs to the same parent tier0"),
		},
	}
}

func getPolicyInterVRFRoutingFromSchema(d *schema.ResourceData) model.PolicyInterVrfRoutingConfig {
	displayName := d.Get("display_name").(string)
	description := d.Get("description").(string)
	tags := getPolicyTagsFromSchema(d)
	var bgpRouteLeaking []model.BgpRouteLeaking
	bpgRouteLeakingList := d.Get("bgp_route_leaking")
	if bpgRouteLeakingList != nil {
		for _, brl := range bpgRouteLeakingList.([]interface{}) {
			brlMap := brl.(map[string]interface{})
			addressFamily := brlMap["address_family"].(string)
			var inFilter []string
			if brlMap["in_filter"] != nil {
				inFilter = interface2StringList(brlMap["in_filter"].([]interface{}))
			}
			var outFilter []string
			if brlMap["out_filter"] != nil {
				outFilter = interface2StringList(brlMap["out_filter"].([]interface{}))
			}
			bgpRouteLeaking = append(bgpRouteLeaking, model.BgpRouteLeaking{
				AddressFamily: &addressFamily,
				InFilter:      inFilter,
				OutFilter:     outFilter,
			})
		}
	}

	var staticRouteAdvertisement *model.PolicyStaticRouteAdvertisement
	staticRouteAdvert := d.Get("static_route_advertisement")
	if staticRouteAdvert != nil {
		for _, sAdvert := range staticRouteAdvert.([]interface{}) {
			// Should be one as list has MaxItems = 1
			sAdvertMap := sAdvert.(map[string]interface{})

			var advertisementRules []model.PolicyRouteAdvertisementRule
			advertRulesList := sAdvertMap["advertisement_rule"]
			if advertRulesList != nil {
				for _, advRule := range advertRulesList.([]interface{}) {
					advRuleMap := advRule.(map[string]interface{})
					action := advRuleMap["action"].(string)
					name := advRuleMap["name"].(string)
					prefixOperator := advRuleMap["prefix_operator"].(string)

					var routeAdvertisementTypes []string
					if advRuleMap["route_advertisement_types"] != nil {
						routeAdvertisementTypes = interface2StringList(advRuleMap["route_advertisement_types"].([]interface{}))
					}

					var subnets []string
					if advRuleMap["subnets"] != nil {
						subnets = interface2StringList(advRuleMap["subnets"].([]interface{}))
					}

					advertisementRules = append(advertisementRules, model.PolicyRouteAdvertisementRule{
						Action:                  &action,
						Name:                    &name,
						PrefixOperator:          &prefixOperator,
						RouteAdvertisementTypes: routeAdvertisementTypes,
						Subnets:                 subnets,
					})
				}
			}
			var inFilterPrefixList []string
			if sAdvertMap["in_filter_prefix_list"] != nil {
				inFilterPrefixList = interface2StringList(sAdvertMap["in_filter_prefix_list"].([]interface{}))
			}
			staticRouteAdvertisement = &model.PolicyStaticRouteAdvertisement{
				AdvertisementRules: advertisementRules,
				InFilterPrefixList: inFilterPrefixList,
			}
		}
	}

	targetPath := d.Get("target_path").(string)

	return model.PolicyInterVrfRoutingConfig{
		DisplayName:              &displayName,
		Description:              &description,
		Tags:                     tags,
		BgpRouteLeaking:          bgpRouteLeaking,
		StaticRouteAdvertisement: staticRouteAdvertisement,
		TargetPath:               &targetPath,
	}
}

func resourceNsxtPolicyTier0InterVRFRoutingCreate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := tier_0s.NewInterVrfRoutingClient(connector)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("tier0 gateway path expected, got %s", gwPolicyPath)
	}

	id := d.Get("nsx_id").(string)
	if id == "" {
		id = newUUID()
	} else {
		_, err := client.Get(gwID, id)
		if err == nil {
			return fmt.Errorf("inter VRF routing configuration with ID '%s' already exists on Tier0 %s", id, gwID)
		} else if !isNotFoundError(err) {
			return err
		}
	}

	obj := getPolicyInterVRFRoutingFromSchema(d)
	err := client.Patch(gwID, id, obj)
	if err != nil {
		return handleCreateError("InterVRFRouting", *obj.DisplayName, err)
	}

	d.SetId(id)
	d.Set("nsx_id", id)

	return resourceNsxtPolicyTier0InterVRFRoutingRead(d, m)
}

func resourceNsxtPolicyTier0InterVRFRoutingRead(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := tier_0s.NewInterVrfRoutingClient(connector)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("Tier0 gateway path expected, got %s", gwPolicyPath)
	}

	id := d.Get("nsx_id").(string)

	obj, err := client.Get(gwID, id)
	if err != nil {
		return handleReadError(d, "InterVRFRouting", id, err)
	}

	d.Set("path", obj.Path)
	d.Set("display_name", obj.DisplayName)
	d.Set("description", obj.Description)
	d.Set("revision", obj.Revision)
	setPolicyTagsInSchema(d, obj.Tags)

	var brlList []interface{}
	for _, brl := range obj.BgpRouteLeaking {
		brlMap := make(map[string]interface{})
		brlMap["address_family"] = brl.AddressFamily
		brlMap["in_filter"] = stringList2Interface(brl.InFilter)
		brlMap["out_filter"] = stringList2Interface(brl.OutFilter)

		brlList = append(brlList, brlMap)
	}
	d.Set("bgp_route_leaking", brlList)

	var sraList []interface{}
	if obj.StaticRouteAdvertisement != nil {
		sra := make(map[string]interface{})
		var advRuleList []interface{}
		for _, advRule := range obj.StaticRouteAdvertisement.AdvertisementRules {
			elem := make(map[string]interface{})
			elem["action"] = advRule.Action
			elem["name"] = advRule.Name
			elem["prefix_operator"] = advRule.PrefixOperator
			elem["route_advertisement_types"] = stringList2Interface(advRule.RouteAdvertisementTypes)
			elem["subnets"] = stringList2Interface(advRule.Subnets)

			advRuleList = append(advRuleList, elem)
		}
		sra["advertisement_rule"] = advRuleList
		sra["in_filter_prefix_list"] = stringList2Interface(obj.StaticRouteAdvertisement.InFilterPrefixList)

		sraList = []interface{}{sra}
	}
	d.Set("static_route_advertisement", sraList)
	d.Set("target_path", obj.TargetPath)

	return nil
}

func resourceNsxtPolicyTier0InterVRFRoutingUpdate(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := tier_0s.NewInterVrfRoutingClient(connector)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("Tier0 gateway path expected, got %s", gwPolicyPath)
	}

	id := d.Get("nsx_id").(string)

	obj := getPolicyInterVRFRoutingFromSchema(d)
	revision := int64(d.Get("revision").(int))
	obj.Revision = &revision

	err := client.Patch(gwID, id, obj)
	if err != nil {
		return handleUpdateError("InterVRFRouting", *obj.DisplayName, err)
	}

	return resourceNsxtPolicyTier0InterVRFRoutingRead(d, m)
}

func resourceNsxtPolicyTier0InterVRFRoutingDelete(d *schema.ResourceData, m interface{}) error {
	connector := getPolicyConnector(m)
	client := tier_0s.NewInterVrfRoutingClient(connector)

	gwPolicyPath := d.Get("gateway_path").(string)
	isT0, gwID := parseGatewayPolicyPath(gwPolicyPath)
	if gwID == "" {
		return fmt.Errorf("gateway_path is not valid")
	}
	if !isT0 {
		return fmt.Errorf("Tier0 gateway path expected, got %s", gwPolicyPath)
	}

	id := d.Get("nsx_id").(string)

	err := client.Delete(gwID, id)
	if err != nil {
		return handleDeleteError("InterVRFRouting", id, err)
	}

	return nil
}

func resourceNsxtPolicyTier0InterVRFRoutingImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	importID := d.Id()

	rd, err := nsxtPolicyPathResourceImporterHelper(d, m)
	if err == nil {
		d.Set("nsx_id", d.Id())
		gwPath, err := getParameterFromPolicyPath("", "/inter-vrf-routing/", importID)
		if err != nil {
			return nil, fmt.Errorf("incorrect policy path for inter vrf routing resource: %v", err)
		}
		d.Set("gateway_path", gwPath)
		return rd, nil
	}
	return nil, fmt.Errorf("not a policy path of an InterVRFRouting resource")
}
