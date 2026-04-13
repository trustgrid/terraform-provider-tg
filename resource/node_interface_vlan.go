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

type nodeInterfaceVLAN struct{}

// NodeInterfaceVLAN returns a Terraform resource for managing a VLAN
// sub-interface on a specific network interface. It uses a read-modify-write
// proxy against the full network config endpoint.
func NodeInterfaceVLAN() *schema.Resource {
	n := nodeInterfaceVLAN{}
	return &schema.Resource{
		Description:   "Manage a VLAN sub-interface on a node or cluster network interface.",
		CreateContext: n.Create,
		ReadContext:   n.Read,
		UpdateContext: n.Update,
		DeleteContext: n.Delete,
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
				Description: "Parent NIC name (e.g. ens192)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"vlan_id": {
				Description:  "VLAN ID (0-4095)",
				Type:         schema.TypeInt,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(0, 4095),
			},
			"ip": {
				Description:  "IP address in CIDR notation",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"vrf": {
				Description: "VRF name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"additional_ips": {
				Description: "Additional IP addresses in CIDR notation",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsCIDR,
				},
			},
			"route": {
				Description: "VLAN static routes",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"route": {
							Description:  "Destination CIDR",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"next": {
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
				},
			},
		},
	}
}

func (n *nodeInterfaceVLAN) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	nic := d.Get("nic").(string)     //nolint: errcheck // ForceNew string field
	vlanID := d.Get("vlan_id").(int) //nolint: errcheck // ForceNew int field
	sub := n.buildTGSub(d, vlanID)

	// Find or create the parent interface.
	found := false
	for i, iface := range nc.Interfaces {
		if iface.NIC != nic {
			continue
		}
		found = true
		replaced := false
		for j, existing := range iface.SubInterfaces {
			if existing.VLANID == vlanID {
				nc.Interfaces[i].SubInterfaces[j] = sub
				replaced = true
				break
			}
		}
		if !replaced {
			nc.Interfaces[i].SubInterfaces = append(nc.Interfaces[i].SubInterfaces, sub)
		}
		break
	}
	if !found {
		return diag.Errorf("interface %q not found in network config; create a tg_node_interface resource for it first", nic)
	}

	if err := putNetworkConfig(ctx, tgc, endpoint, isCluster, nc); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(encodeIfaceVLANID(endpoint, nic, vlanID))
	return nil
}

func (n *nodeInterfaceVLAN) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string)     //nolint: errcheck // ForceNew string field
	vlanID := d.Get("vlan_id").(int) //nolint: errcheck // ForceNew int field

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
		for _, sub := range iface.SubInterfaces {
			if sub.VLANID == vlanID {
				return diag.FromErr(n.setFromTG(d, sub))
			}
		}
		// Interface found, sub-interface gone.
		d.SetId("")
		return nil
	}

	// Interface not found — treat as deleted.
	d.SetId("")
	return nil
}

func (n *nodeInterfaceVLAN) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return n.Create(ctx, d, meta)
}

func (n *nodeInterfaceVLAN) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string)     //nolint: errcheck // ForceNew string field
	vlanID := d.Get("vlan_id").(int) //nolint: errcheck // ForceNew int field

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, iface := range nc.Interfaces {
		if iface.NIC != nic {
			continue
		}
		filtered := iface.SubInterfaces[:0]
		for _, sub := range iface.SubInterfaces {
			if sub.VLANID != vlanID {
				filtered = append(filtered, sub)
			}
		}
		nc.Interfaces[i].SubInterfaces = filtered
		break
	}

	return diag.FromErr(putNetworkConfig(ctx, tgc, endpoint, isCluster, nc))
}

func (n *nodeInterfaceVLAN) buildTGSub(d *schema.ResourceData, vlanID int) tg.SubInterface {
	sub := tg.SubInterface{
		VLANID:      vlanID,
		IP:          d.Get("ip").(string),          //nolint: errcheck // typed schema field
		VRF:         d.Get("vrf").(string),         //nolint: errcheck // typed schema field
		Description: d.Get("description").(string), //nolint: errcheck // typed schema field
	}

	if v, ok := d.GetOk("additional_ips"); ok {
		for _, s := range v.([]interface{}) { //nolint:errcheck // schema ensures TypeList of TypeString
			sub.AdditionalIPs = append(sub.AdditionalIPs, s.(string)) //nolint:errcheck // schema ensures TypeString elements
		}
	}

	if v, ok := d.GetOk("route"); ok {
		for _, raw := range v.([]interface{}) { //nolint:errcheck // schema ensures TypeList of schema.Resource
			m := raw.(map[string]interface{}) //nolint:errcheck // schema ensures map type
			sub.Routes = append(sub.Routes, tg.VLANRoute{
				Route:       m["route"].(string),       //nolint:errcheck // schema ensures TypeString
				Next:        m["next"].(string),        //nolint:errcheck // schema ensures TypeString
				Description: m["description"].(string), //nolint:errcheck // schema ensures TypeString
			})
		}
	}

	return sub
}

func (n *nodeInterfaceVLAN) setFromTG(d *schema.ResourceData, sub tg.SubInterface) error {
	if err := d.Set("ip", sub.IP); err != nil {
		return err
	}
	if err := d.Set("vrf", sub.VRF); err != nil {
		return err
	}
	if err := d.Set("description", sub.Description); err != nil {
		return err
	}
	if err := d.Set("additional_ips", sub.AdditionalIPs); err != nil {
		return err
	}

	routes := make([]map[string]interface{}, 0, len(sub.Routes))
	for _, r := range sub.Routes {
		routes = append(routes, map[string]interface{}{
			"route":       r.Route,
			"next":        r.Next,
			"description": r.Description,
		})
	}
	return d.Set("route", routes)
}
