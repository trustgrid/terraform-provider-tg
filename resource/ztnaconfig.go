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

func ZTNAConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Node ZTNA Gateway Config",

		CreateContext: ztnaConfigCreate,
		ReadContext:   ztnaConfigRead,
		UpdateContext: ztnaConfigUpdate,
		DeleteContext: ztnaConfigDelete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
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

func ztnaConfigCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/apigw", gw.NodeID), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(gw.NodeID)

	return diag.Diagnostics{}
}

func ztnaConfigRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	gw := tg.ZTNAConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+gw.NodeID, &n)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(&n.Config.ZTNA, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gw.NodeID)

	return diag.Diagnostics{}
}

func ztnaConfigUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return ztnaConfigCreate(ctx, d, meta)
}

func ztnaConfigDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	if err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/apigw", d.Id()), map[string]any{"enabled": false, "wireguardEnabled": false}); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
