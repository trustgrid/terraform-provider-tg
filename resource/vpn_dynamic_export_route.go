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

func VPNDynamicExportRoute() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.VPNRoute, hcl.VPNRoute]{
			CreateURL: func(r hcl.VPNRoute) string {
				if r.ClusterFQDN != "" {
					return fmt.Sprintf("/v2/cluster/%s/vpn/%s/dynamic/export-route", r.ClusterFQDN, r.NetworkName)
				}
				return fmt.Sprintf("/v2/node/%s/vpn/%s/dynamic/export-route", r.NodeID, r.NetworkName)
			},
			UpdateURL: func(r hcl.VPNRoute) string {
				if r.ClusterFQDN != "" {
					return fmt.Sprintf("/v2/cluster/%s/vpn/%s/dynamic/export-route/%s", r.ClusterFQDN, r.NetworkName, r.UID)
				}
				return fmt.Sprintf("/v2/node/%s/vpn/%s/dynamic/export-route/%s", r.NodeID, r.NetworkName, r.UID)
			},
			DeleteURL: func(r hcl.VPNRoute) string {
				if r.ClusterFQDN != "" {
					return fmt.Sprintf("/v2/cluster/%s/vpn/%s/dynamic/export-route/%s", r.ClusterFQDN, r.NetworkName, r.UID)
				}
				return fmt.Sprintf("/v2/node/%s/vpn/%s/dynamic/export-route/%s", r.NodeID, r.NetworkName, r.UID)
			},
			IndexURL: func(r hcl.VPNRoute) string {
				if r.ClusterFQDN != "" {
					return fmt.Sprintf("/v2/cluster/%s/vpn/%s/dynamic/export-route", r.ClusterFQDN, r.NetworkName)
				}
				return fmt.Sprintf("/v2/node/%s/vpn/%s/dynamic/export-route", r.NodeID, r.NetworkName)
			},
			OnCreateReply: func(ctx context.Context, args majordomo.CallbackArgs[tg.VPNRoute, hcl.VPNRoute]) (string, error) {
				tgc := tg.GetClient(args.Meta)
				routes := []tg.VPNRoute{}

				url := fmt.Sprintf("/v2/node/%s/vpn/%s/dynamic/export-route", args.Resource.NodeID, args.Resource.NetworkName)
				if args.Resource.ClusterFQDN != "" {
					url = fmt.Sprintf("/v2/cluster/%s/vpn/%s/dynamic/export-route", args.Resource.ClusterFQDN, args.Resource.NetworkName)
				}

				err := tgc.Get(ctx, url, &routes)
				if err != nil {
					return "", err
				}
				for _, route := range routes {
					if route.Node == args.Resource.Node &&
						route.Path == args.Resource.Path &&
						route.Metric == args.Resource.Metric &&
						route.NetworkCIDR == args.Resource.NetworkCIDR {
						return route.UID, nil
					}
				}

				return "", fmt.Errorf("dynamic export route not found")
			},
			AfterCreate: func(_ context.Context, args majordomo.CallbackArgs[tg.VPNRoute, hcl.VPNRoute]) (string, error) {
				return "", args.TF.Set("uid", args.TF.Id())
			},
			ID: func(route hcl.VPNRoute) string {
				return route.UID
			},
			RemoteID: func(route tg.VPNRoute) string {
				return route.UID
			},
		})

	return &schema.Resource{
		Description: "Manage a VPN dynamic export route on a node or cluster.",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Route unique ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"node_id": {
				Description:  "Node ID - required if cluster_fqdn is not specified",
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - required if node_id is not specified",
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"network_name": {
				Description: "Network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"metric": {
				Description:  "Metric",
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 200),
			},
			"network_cidr": {
				Description:  "Network CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"path": {
				Description:   "Path",
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"cluster_fqdn"},
			},
			"node": {
				Description: "Node ID",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}
