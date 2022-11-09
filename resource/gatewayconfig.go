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

func (gc *gatewayConfig) Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	gw, err := gc.marshalResourceData(ctx, d)
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

func (gc *gatewayConfig) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	gw, err := gc.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	n := tg.Node{}
	err = tgc.Get(ctx, "/node/"+gw.NodeID, &n)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := gc.unmarshalResourceData(ctx, n.Config.Gateway, d); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(gw.NodeID)
	if err := d.Set("node_id", gw.NodeID); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (gc *gatewayConfig) Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return gc.Create(ctx, d, meta)
}

func (gc *gatewayConfig) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	gw, err := gc.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/gateway", gw.NodeID), map[string]any{"enabled": false})
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (gc *gatewayConfig) marshalResourceData(ctx context.Context, d *schema.ResourceData) (tg.GatewayConfig, error) {
	gw := tg.GatewayConfig{}
	err := hcl.MarshalResourceData(d, &gw)
	if err != nil {
		return gw, err
	}

	if clients, ok := d.Get("client").([]interface{}); ok {
		for _, c := range clients {
			client := c.(map[string]interface{})
			gw.Clients = append(gw.Clients, tg.GatewayClient{
				Name:    client["name"].(string),
				Enabled: client["enabled"].(bool),
			})
		}
	}

	return gw, nil
}

func (gc *gatewayConfig) unmarshalResourceData(ctx context.Context, config tg.GatewayConfig, d *schema.ResourceData) error {
	if err := hcl.UnmarshalResourceData(&config, d); err != nil {
		return err
	}

	clients := make([]interface{}, 0)
	for _, c := range config.Clients {
		client := make(map[string]interface{})
		client["name"] = c.Name
		client["enabled"] = c.Enabled
		clients = append(clients, client)
	}
	if err := d.Set("client", clients); err != nil {
		return fmt.Errorf("clients=%+v error setting client: %w", clients, err)
	}

	return nil
}
