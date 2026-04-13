package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type nodeInterfaceRoute struct{}

// NodeInterfaceRoute returns a Terraform resource for managing a single static
// route on a specific network interface. It uses a read-modify-write proxy
// against the full network config endpoint.
func NodeInterfaceRoute() *schema.Resource {
	r := nodeInterfaceRoute{}
	return &schema.Resource{
		Description:   "Manage a static route on a node or cluster network interface.",
		CreateContext: r.Create,
		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
				ValidateFunc: validation.IsUUID,
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
				ValidateFunc: validators.IsHostname,
			},
			"nic": {
				Description: "NIC name (e.g. ens192)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"route": {
				Description:  "Destination CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"next_hop": {
				Description:  "Next hop IP address",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"description": {
				Description: "Route description",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func (r *nodeInterfaceRoute) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	nic := d.Get("nic").(string)    //nolint: errcheck // ForceNew string field
	dest := d.Get("route").(string) //nolint: errcheck // ForceNew string field
	route := tg.NetworkRoute{
		Route:       dest,
		Next:        d.Get("next_hop").(string),    //nolint: errcheck // typed schema field
		Description: d.Get("description").(string), //nolint: errcheck // typed schema field
	}

	// Find the interface and upsert the route.
	found := false
	for i, iface := range nc.Interfaces {
		if iface.NIC == nic {
			found = true
			replaced := false
			for j, existing := range iface.Routes {
				if existing.Route == dest {
					nc.Interfaces[i].Routes[j] = route
					replaced = true
					break
				}
			}
			if !replaced {
				nc.Interfaces[i].Routes = append(nc.Interfaces[i].Routes, route)
			}
			break
		}
	}
	if !found {
		return diag.Errorf("interface %q not found in network config; create a tg_node_interface resource for it first", nic)
	}

	if err := putNetworkConfig(ctx, tgc, endpoint, isCluster, nc); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(encodeIfaceRouteID(endpoint, nic, dest))
	return nil
}

func (r *nodeInterfaceRoute) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string)    //nolint: errcheck // ForceNew string field
	dest := d.Get("route").(string) //nolint: errcheck // ForceNew string field

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	for _, iface := range nc.Interfaces {
		if iface.NIC != nic {
			continue
		}
		for _, route := range iface.Routes {
			if route.Route == dest {
				if err := d.Set("next_hop", route.Next); err != nil {
					return diag.FromErr(err)
				}
				if err := d.Set("description", route.Description); err != nil {
					return diag.FromErr(err)
				}
				return nil
			}
		}
		// Interface found but route is gone.
		d.SetId("")
		return nil
	}

	// Interface not found — treat as deleted.
	d.SetId("")
	return nil
}

func (r *nodeInterfaceRoute) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return r.Create(ctx, d, meta)
}

func (r *nodeInterfaceRoute) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string)    //nolint: errcheck // ForceNew string field
	dest := d.Get("route").(string) //nolint: errcheck // ForceNew string field

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, iface := range nc.Interfaces {
		if iface.NIC != nic {
			continue
		}
		filtered := iface.Routes[:0]
		for _, route := range iface.Routes {
			if route.Route != dest {
				filtered = append(filtered, route)
			}
		}
		nc.Interfaces[i].Routes = filtered
		break
	}

	return diag.FromErr(putNetworkConfig(ctx, tgc, endpoint, isCluster, nc))
}
