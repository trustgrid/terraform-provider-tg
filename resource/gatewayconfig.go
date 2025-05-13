package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func GatewayConfig() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.GatewayConfig, hcl.GatewayConfig]{
			OnUpdateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.GatewayConfig, hcl.GatewayConfig]) (string, error) {
				nodeID := args.TF.Get("node_id")

				return nodeID.(string), nil //nolint: errcheck // terraform take the wheel
			},
			UpdateURL: func(a hcl.GatewayConfig) string { return fmt.Sprintf("/node/%s/config/gateway", a.NodeID) },
			DeleteURL: func(a hcl.GatewayConfig) string { return fmt.Sprintf("/node/%s/config/gateway", a.NodeID) },
			GetFromNode: func(n tg.Node) (tg.GatewayConfig, bool, error) {
				gw := n.Config.Gateway
				gw.NodeName = n.Name
				gw.Cluster = n.Cluster
				gw.Domain = n.Domain
				return gw, true, nil
			},
			GetURL: func(a hcl.GatewayConfig) string { return "/node/" + a.NodeID },
			ID: func(a hcl.GatewayConfig) string {
				return a.NodeID
			},
		})

	return &schema.Resource{
		Description: "Node Gateway Config",

		CreateContext: md.Create,
		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,

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
				Description: "Max gateway ingress throughput",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
			"type": {
				Description:  "Gateway type (public, private, or hub) - required with `enabled`",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"enabled"},
				ValidateFunc: validation.StringInSlice([]string{"public", "private", "hub"}, false),
			},
			"connect_to_public": {
				Description: "Allow connectivity to public gateways",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"monitor_hops": {
				Description: "Monitor hop latency from this node to its gateways",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"max_client_write_mbps": {
				Description:  "Maximum gateway egress throughput",
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"path": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Gateway paths",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Path ID",
						},
						"host": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Path host",
							ValidateFunc: validation.IsIPv4Address,
						},
						"port": {
							Type:         schema.TypeInt,
							Required:     true,
							Description:  "Path host",
							ValidateFunc: validation.IsPortNumber,
						},
						"node": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Path node",
						},
					},
				},
			},
			"route": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Gateway routes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"route": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Route",
						},
						"dest": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination",
						},
						"metric": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Metric",
						},
					},
				},
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
