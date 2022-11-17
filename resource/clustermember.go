package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type clustermember struct {
}

func ClusterMember() *schema.Resource {
	c := clustermember{}

	return &schema.Resource{
		Description: "Manage a TG cluster member.",

		ReadContext:   c.Read,
		UpdateContext: c.Update,
		DeleteContext: c.Delete,
		CreateContext: c.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cluster_fqdn": {
				Description: "Node cluster FQDN",
				Type:        schema.TypeString,
				Required:    true,
			},
			"active": {
				Description: "Whether the node should be the active cluster member",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (nr *clustermember) updateCluster(ctx context.Context, tgc *tg.Client, uid string, cluster *string) error {
	payload := map[string]any{
		"cluster": cluster,
	}

	return tgc.Put(ctx, fmt.Sprintf("/node/%s", uid), &payload)
}

func (nr *clustermember) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	uid := d.Get("node_id").(string)
	fqdn := d.Get("cluster_fqdn").(string)

	if err := nr.updateCluster(ctx, tgc, uid, &fqdn); err != nil {
		return diag.FromErr(fmt.Errorf("error setting cluster fqdn %s: %w", fqdn, err))
	}

	d.SetId(uid)

	return nil
}

func (nr *clustermember) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	n := tg.Node{}
	err := tgc.Get(ctx, "/node/"+d.Id(), &n)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_fqdn", n.Cluster); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (nr *clustermember) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	fqdn := d.Get("cluster_fqdn").(string)

	if err := nr.updateCluster(ctx, tgc, d.Id(), &fqdn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (nr *clustermember) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	if err := nr.updateCluster(ctx, tgc, d.Id(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
