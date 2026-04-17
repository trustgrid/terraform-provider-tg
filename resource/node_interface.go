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

type nodeInterface struct{}

// NodeInterface returns a Terraform resource for managing a single network
// interface on a node or cluster. It uses a read-modify-write proxy pattern
// against the full network config endpoint.
func NodeInterface() *schema.Resource {
	n := nodeInterface{}
	return &schema.Resource{
		Description:   "Manage a network interface on a node or cluster.",
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
				Description: "NIC name (e.g. ens192)",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"ip": {
				Description:  "IP address in CIDR notation",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"dhcp": {
				Description: "Enable DHCP. Only applicable to WAN interfaces.",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"gateway": {
				Description:  "Gateway IP address",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"vrf": {
				Description: "VRF name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dns": {
				Description: "DNS server IP addresses",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
			"mode": {
				Description:  "Auto Negotiation mode. Valid values are \"auto\" and \"manual\". When set to \"manual\", speed and duplex must also be provided.",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
			},
			"duplex": {
				Description:  "Interface duplex. Required when mode is \"manual\".",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"full", "half"}, false),
			},
			"speed": {
				Description: "Interface speed in Mbps. Required when mode is \"manual\".",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"mtu": {
				Description: "Interface MTU",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"cluster_ip": {
				Description:   "Cluster IP (cluster only)",
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  validation.IsIPv4Address,
				ConflictsWith: []string{"node_id"},
			},
		},
	}
}

func (n *nodeInterface) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	nic := d.Get("nic").(string) //nolint: errcheck // ForceNew string field
	iface := n.buildTGIface(d, nic)

	// Replace existing or append.
	replaced := false
	for i, existing := range nc.Interfaces {
		if existing.NIC == nic {
			nc.Interfaces[i] = iface
			replaced = true
			break
		}
	}
	if !replaced {
		nc.Interfaces = append(nc.Interfaces, iface)
	}

	if err := putNetworkConfig(ctx, tgc, endpoint, isCluster, nc); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(encodeIfaceID(endpoint, nic))
	return nil
}

func (n *nodeInterface) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string) //nolint: errcheck // ForceNew string field

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
		if iface.NIC == nic {
			return diag.FromErr(n.setFromTG(d, iface))
		}
	}

	// Interface not found in network config — treat as deleted.
	d.SetId("")
	return nil
}

func (n *nodeInterface) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return n.Create(ctx, d, meta)
}

func (n *nodeInterface) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	endpoint, isCluster := ifaceEndpoint(d)
	nic := d.Get("nic").(string) //nolint: errcheck // ForceNew string field

	nc, err := getNetworkConfig(ctx, tgc, endpoint, isCluster)
	if err != nil {
		return diag.FromErr(err)
	}

	filtered := nc.Interfaces[:0]
	for _, iface := range nc.Interfaces {
		if iface.NIC != nic {
			filtered = append(filtered, iface)
		}
	}
	nc.Interfaces = filtered

	return diag.FromErr(putNetworkConfig(ctx, tgc, endpoint, isCluster, nc))
}

func (n *nodeInterface) buildTGIface(d *schema.ResourceData, nic string) tg.NetworkInterface {
	iface := tg.NetworkInterface{
		NIC:       nic,
		DHCP:      d.Get("dhcp").(bool),         //nolint:errcheck // schema ensures TypeBool
		Gateway:   d.Get("gateway").(string),    //nolint:errcheck // schema ensures TypeString
		VRF:       d.Get("vrf").(string),        //nolint:errcheck // schema ensures TypeString
		IP:        d.Get("ip").(string),         //nolint:errcheck // schema ensures TypeString
		Mode:      d.Get("mode").(string),       //nolint:errcheck // schema ensures TypeString
		Duplex:    d.Get("duplex").(string),     //nolint:errcheck // schema ensures TypeString
		Speed:     d.Get("speed").(int),         //nolint:errcheck // schema ensures TypeInt
		MTU:       d.Get("mtu").(int),           //nolint:errcheck // schema ensures TypeInt
		ClusterIP: d.Get("cluster_ip").(string), //nolint:errcheck // schema ensures TypeString
	}

	if v, ok := d.GetOk("dns"); ok {
		for _, s := range v.([]interface{}) { //nolint:errcheck // schema ensures TypeList of TypeString
			iface.DNS = append(iface.DNS, s.(string)) //nolint:errcheck // schema ensures TypeString elements
		}
	}

	return iface
}

func (n *nodeInterface) setFromTG(d *schema.ResourceData, iface tg.NetworkInterface) error {
	fields := map[string]interface{}{
		"nic":        iface.NIC,
		"dhcp":       iface.DHCP,
		"gateway":    iface.Gateway,
		"vrf":        iface.VRF,
		"ip":         iface.IP,
		"mode":       iface.Mode,
		"duplex":     iface.Duplex,
		"speed":      iface.Speed,
		"mtu":        iface.MTU,
		"cluster_ip": iface.ClusterIP,
		"dns":        iface.DNS,
	}
	for k, v := range fields {
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}
