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

type gatewayConfig struct {
}

func GatewayConfig() *schema.Resource {
	c := gatewayConfig{}

	return &schema.Resource{
		Description: "Node Gateway Config",

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
				Description:  "Host port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"udp_port": {
				Description:  "UDP port",
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"maxmbps": {
				Description: "Max gateway throughput",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description:  "Gateway type (public, private, or hub)",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"public", "private", "hub"}, false),
			},
			"connect_to_public": {
				Description: "Allow connectivity to public gateways",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"client": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Private gateway clients",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Client node name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Client enabled",
						},
					},
				},
			},
		},
	}
}

func (gc *gatewayConfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	gw, err := gc.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/gateway", gw.NodeID), &gw)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(gw.NodeID)

	return nil
}

func (gc *gatewayConfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	gw, err := gc.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+gw.NodeID, &n)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if err := gc.convertToTFConfig(ctx, n.Config.Gateway, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gw.NodeID)
	if err := d.Set("node_id", gw.NodeID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (gc *gatewayConfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return gc.Create(ctx, d, meta)
}

func (gc *gatewayConfig) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	gw, err := gc.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/gateway", gw.NodeID), map[string]any{"enabled": false})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (gc *gatewayConfig) decodeTFConfig(_ context.Context, d *schema.ResourceData) (tg.GatewayConfig, error) {
	gw := tg.GatewayConfig{}
	err := hcl.DecodeResourceData(d, &gw)
	return gw, err
}

func (gc *gatewayConfig) convertToTFConfig(_ context.Context, config tg.GatewayConfig, d *schema.ResourceData) error {
	return hcl.EncodeResourceData(&config, d)
}
