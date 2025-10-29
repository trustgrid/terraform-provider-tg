package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
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
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
				ForceNew:     true,
			},
			"cluster_fqdn": {
				Description:  "Node cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Required:     true,
			},
			"active": {
				Description: "Whether the node should be the active cluster member",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func (nr *clustermember) updateCluster(ctx context.Context, tgc *tg.Client, uid string, cluster *string) error {
	payload := map[string]any{
		"cluster": cluster,
	}

	_, err := tgc.Put(ctx, fmt.Sprintf("/node/%s", uid), &payload)
	return err
}

func (nr *clustermember) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	uid, ok := d.Get("node_id").(string)
	if !ok {
		return diag.FromErr(errors.New("node_id must be a string"))
	}
	fqdn, ok := d.Get("cluster_fqdn").(string)
	if !ok {
		return diag.FromErr(errors.New("cluster_fqdn must be a string"))
	}

	if err := nr.updateCluster(ctx, tgc, uid, &fqdn); err != nil {
		return diag.FromErr(fmt.Errorf("error setting cluster fqdn %s: %w", fqdn, err))
	}

	d.SetId(uid)

	return nil
}

func (nr *clustermember) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	n := tg.Node{}
	err := tgc.Get(ctx, "/node/"+d.Id(), &n)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
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
	tgc := tg.GetClient(meta)

	fqdn, ok := d.Get("cluster_fqdn").(string)
	if !ok {
		return diag.FromErr(errors.New("cluster_fqdn must be a string"))
	}

	if err := nr.updateCluster(ctx, tgc, d.Id(), &fqdn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (nr *clustermember) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	if err := nr.updateCluster(ctx, tgc, d.Id(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
