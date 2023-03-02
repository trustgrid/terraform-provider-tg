package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type group struct {
}

// Group returns the TF schema for a group data source
func Group() *schema.Resource {
	r := group{}

	return &schema.Resource{
		Description: "Fetch a user group.",

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
			"idp_id": {
				Description: "IDP id - either local or the IDP uid",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// Read will look up a group by the UID provided. Errors if not found.
func (r *group) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Group{}

	id := d.Get("uid").(string)

	tgapp := tg.Group{}
	err := tgc.Get(ctx, tf.ResourceURL(id), &tgapp)
	if err != nil {
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgapp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
