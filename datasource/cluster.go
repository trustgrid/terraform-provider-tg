package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type cluster struct {
}

func Cluster() *schema.Resource {
	r := cluster{}

	return &schema.Resource{
		Description: "Fetch a cluster.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"fqdn": {
				Description: "FQDN",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (r *cluster) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Cluster](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgCluster := tg.Cluster{}
	err = tgc.Get(ctx, "/cluster/"+tf.FQDN, &tgCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	out := tf.UpdateFromTG(tgCluster)

	if err := hcl.EncodeResourceData(out, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.FQDN)

	return nil
}
