package datasource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type policyDS struct {
}

// Policy returns the TF schema for a policy data source
func Policy() *schema.Resource {
	r := policyDS{}

	return &schema.Resource{
		Description: "Fetch a policy by name.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Policy name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Policy description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"statement": {
				Description: "Policy statements",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"actions": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Actions",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"effect": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Permission effect",
						},
					},
				},
			},
			"conditions": {
				Description: "Policy conditions",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"all": {
							Description: "ALL conditions - all of these must match for the policy to apply",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"eq": {
										Description: "EQ conditions - the property at the provided key must be in the list of values to match",
										Type:        schema.TypeList,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Type:        schema.TypeString,
													Computed:    true,
													Description: "Condition key",
												},
												"values": {
													Type:        schema.TypeList,
													Computed:    true,
													Description: "Condition values",
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
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// Read will look up a policy by the name provided. Errors if not found.
func (r *policyDS) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	name, ok := d.Get("name").(string)
	if !ok {
		return diag.FromErr(errors.New("name must be a string"))
	}

	tgPolicy := tg.Policy{}
	err := tgc.Get(ctx, "/v2/policy/"+name, &tgPolicy)
	if err != nil {
		return diag.FromErr(err)
	}

	tf := hcl.Policy{}.UpdateFromTG(tgPolicy)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	// Use name as the ID since policies don't have a separate UID
	d.SetId(name)

	return nil
}
