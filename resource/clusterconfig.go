package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type clusterconfig struct{}

func ClusterConfig() *schema.Resource {
	c := clusterconfig{}

	return &schema.Resource{
		Description: "Node Cluster Gossip Config",

		CreateContext: c.Create,
		ReadContext:   c.Read,
		UpdateContext: c.Update,
		DeleteContext: c.Delete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Enable the cluster gossip plugin",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"host": {
				Description:  "Host IP",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"port": {
				Description:  "Host Port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"status_host": {
				Description:  "Load balancer status IP",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPv4Address,
				Optional:     true,
			},
			"status_port": {
				Description:  "Load balancer status port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"active": {
				Description: "Whether the node should be the active cluster member. *This is only set on create.*",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (cr *clusterconfig) writeConfig(ctx context.Context, tgc *tg.Client, cc tg.ClusterConfig) error {
	return tgc.Put(ctx, fmt.Sprintf("/node/%s/config/cluster", cc.NodeID), &cc)
}

func (cr *clusterconfig) getConfig(ctx context.Context, tgc *tg.Client, uid string) (*tg.ClusterConfig, error) {
	var n tg.Node
	err := tgc.Get(ctx, fmt.Sprintf("/node/%s", uid), &n)
	if err != nil {
		return nil, err
	}

	return &n.Config.Cluster, nil
}

func (cr *clusterconfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	cc := tg.ClusterConfig{}
	if err := hcl.DecodeResourceData(d, &cc); err != nil {
		return diag.FromErr(err)
	}

	if err := cr.writeConfig(ctx, tgc, cc); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(d.Get("node_id").(string))

	return diag.Diagnostics{}
}

func (cr *clusterconfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	cc, err := cr.getConfig(ctx, tgc, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	oldActive, activeSet := d.GetOk("active")

	if err := hcl.EncodeResourceData(cc, d); err != nil {
		return diag.FromErr(err)
	}

	if activeSet {
		if err := d.Set("active", oldActive.(bool)); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("node_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (cr *clusterconfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	existing, err := cr.getConfig(ctx, tgc, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cc := tg.ClusterConfig{}
	if err := hcl.DecodeResourceData(d, &cc); err != nil {
		return diag.FromErr(err)
	}

	cc.Active = existing.Active
	if err := cr.writeConfig(ctx, tgc, cc); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (cr *clusterconfig) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	cc, err := cr.getConfig(ctx, tgc, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	cc.Enabled = false

	if err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/cluster", d.Id()), &cc); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
