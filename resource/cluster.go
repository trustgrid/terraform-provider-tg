package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type cluster struct {
}

func Cluster() *schema.Resource {
	c := cluster{}

	return &schema.Resource{
		Description: "Manage a TG node cluster",

		ReadContext:   c.Read,
		DeleteContext: c.Delete,
		CreateContext: c.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Cluster Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"fqdn": {
				Description: "Cluster FQDN",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (cr *cluster) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	cluster := tg.Cluster{
		Name: d.Get("name").(string),
	}

	err := tgc.Post(ctx, "/cluster", &cluster)
	if err != nil {
		return diag.FromErr(err)
	}

	fqdn := cluster.Name + "." + tgc.Domain
	if err := d.Set("fqdn", fqdn); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cluster.Name + "." + tgc.Domain)
	return diag.Diagnostics{}
}

func (cr *cluster) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	var cluster tg.Cluster

	if err := tgc.Get(ctx, "/cluster/"+d.Id(), &cluster); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (cr *cluster) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	if err := tgc.Delete(ctx, "/cluster/"+d.Id(), nil); err != nil {
		return diag.FromErr(fmt.Errorf("error issuing delete to /cluster/%s: %w", d.Id(), err))
	}

	return diag.Diagnostics{}
}