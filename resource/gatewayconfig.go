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

func GatewayConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Node Gateway Config",

		CreateContext: gatewayConfigCreate,
		ReadContext:   gatewayConfigRead,
		UpdateContext: gatewayConfigUpdate,
		DeleteContext: gatewayConfigDelete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Gateway Enabled",
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
				Description:  "Gateway Port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"type": {
				Description:  "Gateway Type",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"public", "private", "hub"}, false),
			},
		},
	}
}

func gatewayConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw := tg.GatewayConfig{}
	err := hcl.MarshalResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/gateway", gw.NodeID), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(gw.NodeID)

	return diag.Diagnostics{}
}

func gatewayConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	gw := tg.GatewayConfig{}
	err := hcl.MarshalResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+gw.NodeID, &n)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.UnmarshalResourceData(&n.Config.Gateway, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gw.NodeID)
	d.Set("node_id", gw.NodeID)

	return diag.Diagnostics{}
}

func gatewayConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return gatewayConfigCreate(ctx, d, meta)
}

func gatewayConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	gw := tg.GatewayConfig{}
	err := hcl.MarshalResourceData(d, &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/gateway", gw.NodeID), map[string]any{"enabled": false})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}
