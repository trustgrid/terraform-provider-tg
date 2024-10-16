package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type alarm struct {
}

func Alarm() *schema.Resource {
	r := alarm{}

	return &schema.Resource{
		Description: "Fetch an alarm.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"enabled": {
				Description: "When true, this alarm can generate alerts",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"channels": {
				Description: "Channel IDs",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"nodes": {
				Description: "Node names",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"freetext": {
				Description: "Free text match",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expr": {
				Description: "CEL expression",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"operator": {
				Description: "Criteria operator",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"threshold": {
				Description: "Severity threshold",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"types": {
				Description: "Event types",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tag_operator": {
				Description: "Tag match operator",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tag": {
				Description: "Tag pairs",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Tag name",
							Type:        schema.TypeString,
							Required:    true,
						},
						"value": {
							Description: "Tag value",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *alarm) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Alarm{}

	id, ok := d.Get("uid").(string)
	if !ok {
		return diag.FromErr(fmt.Errorf("uid must be a string"))
	}

	tgchannel := tg.Alarm{}
	err := tgc.Get(ctx, tf.ResourceURL(id), &tgchannel)
	if err != nil {
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgchannel)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
