package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type ztnaConfig struct{}

func ZTNAConfig() *schema.Resource {
	r := ztnaConfig{}

	return &schema.Resource{
		Description: "Manage ZTNA Gateway config for a node or cluster.",

		CreateContext: r.Create,
		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID - required if cluster_fqdn is not set",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - required if node_id is not set",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"enabled": {
				Description: "Enable the gateway plugin",
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
			"wg_enabled": {
				Description: "Enable the wireguard gateway feature",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"wg_endpoint": {
				Description: "Wireguard endpoint",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"wg_port": {
				Description:  "Wireguard port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"cert": {
				Description: "Certificate",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (z *ztnaConfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, z.url(gw), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	if gw.NodeID != "" {
		d.SetId(gw.NodeID)
	} else {
		d.SetId(gw.ClusterFQDN)
	}

	return nil
}

func (z *ztnaConfig) url(c tg.ZTNAConfig) string {
	if c.NodeID != "" {
		return fmt.Sprintf("/node/%s/config/ztnagw", c.NodeID)
	}
	return fmt.Sprintf("/cluster/%s/config/ztnagw", c.ClusterFQDN)
}

func (z *ztnaConfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	ztna := tg.ZTNAConfig{}

	if gw.NodeID != "" {
		n := tg.Node{}
		err = tgc.Get(ctx, "/node/"+d.Id(), &n)
		ztna = n.Config.ZTNA
		ztna.NodeID = gw.NodeID
	} else {
		c := tg.Cluster{}
		err = tgc.Get(ctx, "/cluster/"+d.Id(), &c)
		ztna = c.Config.ZTNA
		ztna.ClusterFQDN = gw.ClusterFQDN
	}

	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(&ztna, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (z *ztnaConfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return z.Create(ctx, d, meta)
}

func (z *ztnaConfig) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, z.url(gw), map[string]any{"enabled": false, "wireguardEnabled": false}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
