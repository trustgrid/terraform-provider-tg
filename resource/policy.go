package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Policy() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.Policy, hcl.Policy]{
			CreateURL: func(_ hcl.Policy) string { return "/v2/policy" },
			UpdateURL: func(p hcl.Policy) string { return "/v2/policy/" + p.Name },
			DeleteURL: func(p hcl.Policy) string { return "/v2/policy/" + p.Name },
			GetURL:    func(p hcl.Policy) string { return "/v2/policy/" + p.Name },
			ID: func(user hcl.Policy) string {
				return user.Name
			},
			RemoteID: func(user tg.Policy) string {
				return user.Name
			},
		})

	return &schema.Resource{
		Description: "Manage a Trustgrid permissions policy",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Policy name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Policy description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"statement": {
				Description: "Policy statements",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Actions",
							MinItems:    1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"effect": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Permission effect",
							ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
						},
					},
				},
			},
			"conditions": {
				Description: "Policy conditions",
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all": {
							Description: "ALL conditions - all of these must match for the policy to apply",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"ne": {
										Description: "NE conditions - the property at the provided key must NOT be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
								},
							},
						},
						"any": {
							Description: "ANY conditions - at least one of these must match for the policy to apply",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"ne": {
										Description: "NE conditions - the property at the provided key must NOT be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
								},
							},
						},
						"none": {
							Description: "NONE conditions - none of these can match for the policy to apply",
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
									"ne": {
										Description: "NE conditions - the property at the provided key must NOT be in the list of values to match",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Required:    true,
													Description: "Condition values",
													MinItems:    1,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"resources": {
				Description: "Resources",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
