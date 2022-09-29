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
			"udp_enabled": {
				Description: "Enable gateway UDP mode",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"cert": {
				Description: "Gateway TLS certificate",
				Type:        schema.TypeString,
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
			"udp_port": {
				Description:  "UDP Port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"maxmbps": {
				Description: "Max Gateway throughput",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description:  "Gateway Type (public, private, or hub)",
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
	if err := d.Set("node_id", gw.NodeID); err != nil {
		return diag.FromErr(err)
	}

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
