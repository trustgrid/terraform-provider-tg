package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Org() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches org info from Trustgrid",

		ReadContext: orgRead,

		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Domain",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"uid": {
				Description: "UID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func orgRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	org := tg.Org{}

	tflog.Debug(ctx, "orgRead: reading org")
	if err := tgc.Get(ctx, "/org/mine", &org); err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(&org, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(org.UID)
	tflog.Debug(ctx, "orgRead: returning", map[string]any{"domain": org.Domain})

	return diag.Diagnostics{}
}
