package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func IDP() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches an IDP by ID",

		ReadContext: idpRead,

		Schema: map[string]*schema.Schema{
			"type": {
				Description: "Type",
				Type:        schema.TypeString,
				Computed:    true,
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
			"uid": {
				Description: "UID",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func idpRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.IDP{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgidp := tg.IDP{}
	err := tgc.Get(ctx, tf.ResourceURL(), &tgidp)
	if err != nil {
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgidp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)

	return nil
}
