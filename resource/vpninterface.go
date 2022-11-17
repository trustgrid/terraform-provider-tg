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

type vpnInterface struct {
}

type HCLVPNInterfaceInsideNAT struct {
	NetworkCIDR string `tf:"network_cidr"`
	LocalCIDR   string `tf:"local_cidr"`
	Description string `tf:"description"`
}

type HCLVPNInterfaceOutsideNAT struct {
	NetworkCIDR string `tf:"network_cidr"`
	LocalCIDR   string `tf:"local_cidr"`
	Description string `tf:"description"`
	ProxyARP    bool   `tf:"proxy_arp"`
}

type HCLVPNInterface struct {
	NodeID      string `tf:"node_id"`
	ClusterFQDN string `tf:"cluster_fqdn"`
	NetworkName string `tf:"network"`

	InterfaceName   string                      `tf:"interface_name"`
	Description     string                      `tf:"description"`
	InDefaultRoute  bool                        `tf:"in_default_route"`
	OutDefaultRoute bool                        `tf:"out_default_route"`
	InsideNATs      []HCLVPNInterfaceInsideNAT  `tf:"inside_nat"`
	OutsideNATs     []HCLVPNInterfaceOutsideNAT `tf:"outside_nat"`
}

func (h *HCLVPNInterface) url() string {
	if h.NodeID != "" {
		return fmt.Sprintf("/v2/node/%s/vpn/%s/interface", h.NodeID, h.NetworkName)
	}
	return fmt.Sprintf("/v2/cluster/%s/vpn/%s/interface", h.ClusterFQDN, h.NetworkName)
}

func (h *HCLVPNInterface) resourceURL() string {
	return h.url() + "/" + h.InterfaceName
}

func VPNInterface() *schema.Resource {
	r := vpnInterface{}

	return &schema.Resource{
		Description: "Manage a VPN interface.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"interface_name": {
				Description: "Interface name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"in_default_route": {
				Description: "Inbound traffic not matching a NAT should be allowed on this interface",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"out_default_route": {
				Description: "Outbound traffic not matching a NAT should be allowed on this interface",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"inside_nat": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Inside NATs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network CIDR",
							ValidateFunc: validation.IsCIDR,
						},
						"local_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Local CIDR",
							ValidateFunc: validation.IsCIDR,
						},
						"description": {
							Description: "Description",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
			"outside_nat": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Outside NATs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Network CIDR",
							ValidateFunc: validation.IsCIDR,
						},
						"local_cidr": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Local CIDR",
							ValidateFunc: validation.IsCIDR,
						},
						"description": {
							Description: "Description",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"proxy_arp": {
							Description: "Proxy ARP",
							Type:        schema.TypeBool,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}

func (h *HCLVPNInterface) toTGVPNInterface() tg.VPNInterface {
	iface := tg.VPNInterface{
		InterfaceName:   h.InterfaceName,
		Description:     h.Description,
		InDefaultRoute:  h.InDefaultRoute,
		OutDefaultRoute: h.OutDefaultRoute,
		InsideNATs:      make([]tg.VPNInterfaceNAT, 0),
		OutsideNATs:     make([]tg.VPNInterfaceNAT, 0),
	}

	for _, nat := range h.InsideNATs {
		iface.InsideNATs = append(iface.InsideNATs, tg.VPNInterfaceNAT{
			NetworkCIDR: nat.NetworkCIDR,
			LocalCIDR:   nat.LocalCIDR,
			Description: nat.Description,
		})
	}

	for _, nat := range h.OutsideNATs {
		iface.OutsideNATs = append(iface.OutsideNATs, tg.VPNInterfaceNAT{
			NetworkCIDR: nat.NetworkCIDR,
			LocalCIDR:   nat.LocalCIDR,
			Description: nat.Description,
			ProxyARP:    nat.ProxyARP,
		})
	}

	return iface
}

func (h *HCLVPNInterface) updateFromTGVPNInterface(vpn tg.VPNInterface) {
	h.InterfaceName = vpn.InterfaceName
	h.Description = vpn.Description
	h.InDefaultRoute = vpn.InDefaultRoute
	h.OutDefaultRoute = vpn.OutDefaultRoute
	h.InsideNATs = make([]HCLVPNInterfaceInsideNAT, 0)
	h.OutsideNATs = make([]HCLVPNInterfaceOutsideNAT, 0)

	for _, nat := range vpn.InsideNATs {
		h.InsideNATs = append(h.InsideNATs, HCLVPNInterfaceInsideNAT{
			NetworkCIDR: nat.NetworkCIDR,
			LocalCIDR:   nat.LocalCIDR,
			Description: nat.Description,
		})
	}

	for _, nat := range vpn.OutsideNATs {
		h.OutsideNATs = append(h.OutsideNATs, HCLVPNInterfaceOutsideNAT{
			NetworkCIDR: nat.NetworkCIDR,
			LocalCIDR:   nat.LocalCIDR,
			Description: nat.Description,
			ProxyARP:    nat.ProxyARP,
		})
	}
}

func (vn *vpnInterface) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLVPNInterface{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	vpn := tf.toTGVPNInterface()

	if err := tgc.Post(ctx, tf.url(), &vpn); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.InterfaceName)

	return nil
}

func (vn *vpnInterface) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLVPNInterface{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	vpn := tf.toTGVPNInterface()

	if err := tgc.Put(ctx, tf.resourceURL(), &vpn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vpnInterface) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLVPNInterface{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.resourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vpnInterface) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := HCLVPNInterface{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	ifaces := []tg.VPNInterface{}
	if err := tgc.Get(ctx, tf.url(), &ifaces); err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, iface := range ifaces {
		if iface.InterfaceName == tf.InterfaceName {
			found = true
			tf.updateFromTGVPNInterface(iface)
			break
		}
	}
	if !found {
		d.SetId("")
		return nil
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
