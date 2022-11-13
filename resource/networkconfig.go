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

type network struct{}

func NetworkConfig() *schema.Resource {
	n := network{}

	return &schema.Resource{
		Description: "Network Config",

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
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"dark_mode": {
				Description: "Dark mode",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"forwarding": {
				Description: "Forwarding",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"vrf": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "VRFs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "VRF name",
						},
						"forwarding": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Enable forwarding",
						},
						"acl": {
							Description: "ACLs",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Description:  "Action",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"allow", "drop", "reject"}, false),
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"protocol": {
										Description:  "Protocol",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"any", "icmp", "tcp", "udp"}, false),
									},
									"source": {
										Description:  "Source",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"dest": {
										Description:  "Destination",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"line": {
										Description:  "Line",
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 32768),
									},
								},
							},
						},
						"route": {
							Description: "Routes",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dev": {
										Description: "Dev",
										Type:        schema.TypeString,
										Required:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"dest": {
										Description:  "Destination",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"metric": {
										Description:  "Metric",
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 200),
									},
								},
							},
						},
						"nat": {
							Description: "NATs",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"source": {
										Description:  "Source",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"dest": {
										Description:  "Destination",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"to_source": {
										Description:  "To Source",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"to_dest": {
										Description:  "To Dest",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"masquerade": {
										Description: "Masquerade",
										Type:        schema.TypeBool,
										Optional:    true,
									},
								},
							},
						},
						"rule": {
							Description: "Rules",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"protocol": {
										Description:  "Protocol",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"any", "icmp", "tcp", "udp"}, false),
									},
									"line": {
										Description:  "Line",
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntBetween(1, 32768),
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"source": {
										Description:  "Source",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"dest": {
										Description:  "Destination",
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.IsCIDR,
									},
									"vrf": {
										Description: "VRF",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"action": {
										Description:  "To Dest",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"accept", "drop", "reject", "forward", "dnat"}, false),
									},
								},
							},
						},
					},
				},
			},
			"interface": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Network interfaces",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nic": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "NIC name",
						},
						"cluster_ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Cluster IP",
							ValidateFunc: validation.IsIPv4Address,
						},
						"routes": {
							Description: "Interface routes",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsCIDR,
							},
						},
						"dhcp": {
							Description: "Enable DHCP",
							Type:        schema.TypeBool,
							Optional:    true,
						},
						"gateway": {
							Description:  "Gateway IP address",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						"ip": {
							Description:  "IP address",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"mode": {
							Description:  "Interface mode",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"auto", "manual"}, false),
						},
						"duplex": {
							Description:  "Interface duplex",
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"full", "half"}, false),
						},
						"speed": {
							Description: "Interface speed",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"dns": {
							Description: "DNS servers",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPv4Address,
							},
						},
					},
				},
			},
			"tunnel": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Network tunnels",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Enable the tunnel",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Tunnel name",
						},
						"ike": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "IKE",
							ValidateFunc: validation.IntInSlice([]int{1, 2}),
						},
						"ike_cipher": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "IKE Cipher",
							ValidateFunc: validation.StringInSlice([]string{"aes128-sha1", "aes128-sha256", "aes256-sha1", "aes256-sha256"}, false),
						},
						"ike_group": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "IKE Group",
							ValidateFunc: validation.IntInSlice([]int{2, 5, 14, 15, 16}),
						},
						"rekey_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Rekey Interval",
						},
						"ip": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "IP",
							ValidateFunc: validation.IsCIDR,
						},
						"destination": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Destination",
							ValidateFunc: validation.IsIPv4Address,
						},
						"ipsec_cipher": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "IPSec Cipher",
							ValidateFunc: validation.StringInSlice([]string{"aes128-sha1", "aes128-sha256", "aes256-sha1", "aes256-sha256"}, false),
						},
						"psk": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "PSK",
							Sensitive:   true,
						},
						"vrf": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "VRF",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Tunnel type",
							ValidateFunc: validation.StringInSlice([]string{"ipsec", "vnet"}, false),
						},
						"mtu": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "MTU",
						},
						"network_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Network ID",
						},
						"local_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Local ID",
						},
						"remote_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Remote ID",
						},
						"dpd_retries": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "DPD Retries",
						},
						"dpd_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "DPD Interval",
						},
						"iface": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Interface",
						},
						"pfs": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "PFS",
							ValidateFunc: validation.IntInSlice([]int{0, 2, 5, 14, 15, 16}),
						},
						"replay_window": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Replay Window",
							ValidateFunc: validation.IntInSlice([]int{32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}),
						},
					},
				},
			},
		},
	}
}

type HCLNetworkTunnel struct {
	Enabled       bool   `tf:"enabled"`
	Name          string `tf:"name"`
	IKE           int    `tf:"ike,omitempty"`
	IKECipher     string `tf:"ike_cipher,omitempty"`
	IKEGroup      int    `tf:"ike_group,omitempty"`
	RekeyInterval int    `tf:"rekey_interval,omitempty"`
	IP            string `tf:"ip,omitempty"`
	Destination   string `tf:"destination,omitempty"`
	IPSecCipher   string `tf:"ipsec_cipher,omitempty"`
	PSK           string `tf:"psk,omitempty"`
	VRF           string `tf:"vrf,omitempty"`
	Type          string `tf:"type"`
	MTU           int    `tf:"mtu"`
	NetworkID     int    `tf:"network_id"`
	LocalID       string `tf:"local_id,omitempty"`
	RemoteID      string `tf:"remote_id,omitempty"`
	DPDRetries    int    `tf:"dpd_retries,omitempty"`
	DPDInterval   int    `tf:"dpd_interval,omitempty"`
	IFace         string `tf:"iface,omitempty"`
	PFS           int    `tf:"pfs"` // TODO we should omit this when appropriate
	ReplayWindow  int    `tf:"replay_window,omitempty"`
}

type HCLNetworkInterface struct {
	NIC       string   `tf:"nic"`
	Routes    []string `tf:"routes,omitempty"`
	ClusterIP string   `tf:"cluster_ip,omitempty"`
	DHCP      bool     `tf:"dhcp"`
	Gateway   string   `tf:"gateway"`
	IP        string   `tf:"ip"`
	Mode      string   `tf:"mode,omitempty"`
	DNS       []string `tf:"dns,omitempty"`
	Duplex    string   `tf:"duplex,omitempty"`
	Speed     int      `tf:"speed,omitempty"`
}

type HCLVRFACL struct {
	Action      string `tf:"action"`
	Description string `tf:"description"`
	Protocol    string `tf:"protocol"`
	Source      string `tf:"source"`
	Dest        string `tf:"dest"`
	Line        int    `tf:"line"`
}

type HCLVRFRoute struct {
	Dest        string `tf:"dest"`
	Dev         string `tf:"dev"`
	Description string `tf:"description"`
	Metric      int    `tf:"metric"`
}

type HCLVRFNAT struct {
	Source     string `tf:"source,omitempty"`
	Dest       string `tf:"dest,omitempty"`
	Masquerade bool   `tf:"masquerade"`
	ToSource   string `tf:"to_source,omitempty"`
	ToDest     string `tf:"to_dest,omitempty"`
}

type HCLVRFRule struct {
	Protocol    string `tf:"protocol"`
	Line        int    `tf:"line"`
	Action      string `tf:"action"`
	Description string `tf:"description,omitempty"`
	Source      string `tf:"source,omitempty"`
	VRF         string `tf:"vrf,omitempty"`
	Dest        string `tf:"dest,omitempty"`
}

type HCLVRF struct {
	Name       string        `tf:"name"`
	Forwarding bool          `tf:"forwarding"`
	ACLs       []HCLVRFACL   `tf:"acls,omitempty"`
	Routes     []HCLVRFRoute `tf:"routes,omitempty"`
	NATs       []HCLVRFNAT   `tf:"nats,omitempty"`
	Rules      []HCLVRFRule  `tf:"rules,omitempty"`
}

type HCLNetworkConfigData struct {
	DarkMode   bool `tf:"dark_mode"`
	Forwarding bool `tf:"forwarding"`

	Tunnels []HCLNetworkTunnel `tf:"tunnel"`

	Interfaces []HCLNetworkInterface `tf:"interface"`

	VRFs []HCLVRF
}

func (nr *network) convertToTFConfig(ctx context.Context, c tg.NetworkConfig, d *schema.ResourceData) error {
	nc := HCLNetworkConfigData{
		DarkMode:   c.DarkMode,
		Forwarding: c.Forwarding,
	}

	for _, t := range c.Tunnels {
		nc.Tunnels = append(nc.Tunnels, HCLNetworkTunnel{
			Enabled:       t.Enabled,
			Name:          t.Name,
			IKE:           t.IKE,
			IKECipher:     t.IKECipher,
			IKEGroup:      t.IKEGroup,
			RekeyInterval: t.RekeyInterval,
			IP:            t.IP,
			Destination:   t.Destination,
			IPSecCipher:   t.IPSecCipher,
			PSK:           t.PSK,
			VRF:           t.VRF,
			Type:          t.Type,
			MTU:           t.MTU,
			NetworkID:     t.NetworkID,
			LocalID:       t.LocalID,
			RemoteID:      t.RemoteID,
			DPDRetries:    t.DPDRetries,
			DPDInterval:   t.DPDInterval,
			IFace:         t.IFace,
			PFS:           t.PFS,
			ReplayWindow:  t.ReplayWindow,
		})
	}

	for _, i := range c.Interfaces {
		iface := HCLNetworkInterface{
			NIC:       i.NIC,
			ClusterIP: i.ClusterIP,
			DHCP:      i.DHCP,
			Gateway:   i.Gateway,
			IP:        i.IP,
			Mode:      i.Mode,
			Duplex:    i.Duplex,
			Speed:     i.Speed,
			DNS:       i.DNS,
		}

		for _, r := range i.Routes {
			iface.Routes = append(iface.Routes, r.Route)
		}
		nc.Interfaces = append(nc.Interfaces, iface)
	}

	for _, v := range c.VRFs {
		vrf := HCLVRF{
			Name:       v.Name,
			Forwarding: v.Forwarding,
		}

		for _, a := range v.ACLs {
			vrf.ACLs = append(vrf.ACLs, HCLVRFACL{
				Action:      a.Action,
				Description: a.Description,
				Protocol:    a.Protocol,
				Source:      a.Source,
				Dest:        a.Dest,
				Line:        a.Line,
			})
		}

		for _, n := range v.NATs {
			vrf.NATs = append(vrf.NATs, HCLVRFNAT{
				Source:     n.Source,
				Dest:       n.Dest,
				Masquerade: n.Masquerade,
				ToSource:   n.ToSource,
				ToDest:     n.ToDest,
			})
		}

		for _, r := range v.Rules {
			vrf.Rules = append(vrf.Rules, HCLVRFRule{
				Protocol:    r.Protocol,
				Line:        r.Line,
				Action:      r.Action,
				Description: r.Description,
				Source:      r.Source,
				VRF:         r.VRF,
				Dest:        r.Dest,
			})
		}

		for _, r := range v.Routes {
			vrf.Routes = append(vrf.Routes, HCLVRFRoute{
				Dest:        r.Dest,
				Dev:         r.Dev,
				Description: r.Description,
				Metric:      r.Metric,
			})
		}
	}

	if err := hcl.EncodeResourceData(&nc, d); err != nil {
		return err
	}

	return nil
}

func (nr *network) decodeTFConfig(ctx context.Context, d *schema.ResourceData) (tg.NetworkConfig, error) {
	data := HCLNetworkConfigData{}

	if err := hcl.DecodeResourceData(d, &data); err != nil {
		return tg.NetworkConfig{}, err
	}

	nc := tg.NetworkConfig{
		DarkMode:   data.DarkMode,
		Forwarding: data.Forwarding,
	}

	for _, t := range data.Tunnels {
		nc.Tunnels = append(nc.Tunnels, tg.NetworkTunnel{
			Enabled:       t.Enabled,
			Name:          t.Name,
			IKE:           t.IKE,
			IKECipher:     t.IKECipher,
			IKEGroup:      t.IKEGroup,
			RekeyInterval: t.RekeyInterval,
			IP:            t.IP,
			Destination:   t.Destination,
			IPSecCipher:   t.IPSecCipher,
			PSK:           t.PSK,
			VRF:           t.VRF,
			Type:          t.Type,
			MTU:           t.MTU,
			NetworkID:     t.NetworkID,
			LocalID:       t.LocalID,
			RemoteID:      t.RemoteID,
			DPDRetries:    t.DPDRetries,
			DPDInterval:   t.DPDInterval,
			IFace:         t.IFace,
			PFS:           t.PFS,
			ReplayWindow:  t.ReplayWindow,
		})
	}

	for _, i := range data.Interfaces {
		iface := tg.NetworkInterface{
			NIC:       i.NIC,
			ClusterIP: i.ClusterIP,
			DHCP:      i.DHCP,
			Gateway:   i.Gateway,
			IP:        i.IP,
			Mode:      i.Mode,
			Duplex:    i.Duplex,
			Speed:     i.Speed,
			DNS:       i.DNS,
		}
		for _, r := range i.Routes {
			iface.Routes = append(iface.Routes, tg.NetworkRoute{Route: r})
		}

		nc.Interfaces = append(nc.Interfaces, iface)
	}

	for _, v := range data.VRFs {
		vrf := tg.VRF{
			Name:       v.Name,
			Forwarding: v.Forwarding,
		}

		for _, a := range v.ACLs {
			vrf.ACLs = append(vrf.ACLs, tg.VRFACL{
				Action:      a.Action,
				Description: a.Description,
				Protocol:    a.Protocol,
				Source:      a.Source,
				Dest:        a.Dest,
				Line:        a.Line,
			})
		}

		for _, n := range v.NATs {
			vrf.NATs = append(vrf.NATs, tg.VRFNAT{
				Source:     n.Source,
				Dest:       n.Dest,
				Masquerade: n.Masquerade,
				ToSource:   n.ToSource,
				ToDest:     n.ToDest,
			})
		}

		for _, r := range v.Rules {
			vrf.Rules = append(vrf.Rules, tg.VRFRule{
				Protocol:    r.Protocol,
				Line:        r.Line,
				Action:      r.Action,
				Description: r.Description,
				Source:      r.Source,
				VRF:         r.VRF,
				Dest:        r.Dest,
			})
		}

		for _, r := range v.Routes {
			vrf.Routes = append(vrf.Routes, tg.VRFRoute{
				Description: r.Description,
				Dest:        r.Dest,
				Dev:         r.Dev,
				Metric:      r.Metric,
			})
		}

		nc.VRFs = append(nc.VRFs, vrf)
	}

	return nc, nil
}

func (nr *network) endpoint(d *schema.ResourceData) (id string, isCluster bool) {
	nodeid, ok := d.GetOk("node_id")
	isCluster = false
	if !ok || nodeid == "" {
		id = d.Get("cluster_fqdn").(string)
		isCluster = true
		return
	}
	id = nodeid.(string)
	return
}

func (nr *network) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	nc, err := nr.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	id, isCluster := nr.endpoint(d)

	if isCluster {
		err = tgc.Put(ctx, fmt.Sprintf("/cluster/%s/config/network", id), &nc)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/network", id), &nc)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}

func (nr *network) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	id, isCluster := nr.endpoint(d)

	if isCluster {
		n := tg.Cluster{}
		if err := tgc.Get(ctx, "/cluster/"+id, &n); err != nil {
			return diag.FromErr(fmt.Errorf("cannot lookup cluster id=%s isCluster=%t %w", id, isCluster, err))
		}

		if err := nr.convertToTFConfig(ctx, n.Config.Network, d); err != nil {
			return diag.FromErr(err)
		}

		return diag.Diagnostics{}
	} else {
		n := tg.Node{}
		if err := tgc.Get(ctx, "/node/"+id, &n); err != nil {
			return diag.FromErr(err)
		}

		if err := nr.convertToTFConfig(ctx, n.Config.Network, d); err != nil {
			return diag.FromErr(err)
		}

		return diag.Diagnostics{}
	}
}

func (nr *network) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nr.Create(ctx, d, meta)
}

func (nr *network) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Noop

	return diag.Diagnostics{}
}
