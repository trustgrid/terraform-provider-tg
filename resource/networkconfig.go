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
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
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

func (nr *network) unmarshalResourceData(ctx context.Context, c tg.NetworkConfig, d *schema.ResourceData) error {
	if err := hcl.UnmarshalResourceData(&c, d); err != nil {
		return err
	}

	tunnels := make([]any, 0)
	for _, t := range c.Tunnels {
		tunnel := make(map[string]any)
		tunnel["enabled"] = t.Enabled
		tunnel["name"] = t.Name
		tunnel["ike"] = t.IKE
		tunnel["ike_cipher"] = t.IKECipher
		tunnel["ike_group"] = t.IKEGroup
		tunnel["rekey_interval"] = t.RekeyInterval
		tunnel["ip"] = t.IP
		tunnel["destination"] = t.Destination
		tunnel["ipsec_cipher"] = t.IPSecCipher
		tunnel["psk"] = t.PSK
		tunnel["vrf"] = t.VRF
		tunnel["type"] = t.Type
		tunnel["mtu"] = t.MTU
		tunnel["network_id"] = t.NetworkID
		tunnel["local_id"] = t.LocalID
		tunnel["remote_id"] = t.RemoteID
		tunnel["dpd_retries"] = t.DPDRetries
		tunnel["dpd_interval"] = t.DPDInterval
		tunnel["iface"] = t.IFace
		tunnel["pfs"] = t.PFS
		tunnel["replay_window"] = t.ReplayWindow

		tunnels = append(tunnels, tunnel)
	}
	if err := d.Set("tunnel", tunnels); err != nil {
		return fmt.Errorf("error setting tunnels: %w", err)
	}

	ifaces := make([]any, 0)
	for _, i := range c.Interfaces {
		iface := make(map[string]any)
		iface["nic"] = i.NIC
		iface["cluster_ip"] = i.ClusterIP
		iface["routes"] = make([]string, 0)
		for _, r := range i.Routes {
			iface["routes"] = append(iface["routes"].([]string), r.Route)
		}
		iface["dhcp"] = i.DHCP
		iface["gateway"] = i.Gateway
		iface["ip"] = i.IP
		iface["mode"] = i.Mode
		iface["duplex"] = i.Duplex
		iface["speed"] = i.Speed
		iface["dns"] = make([]string, 0)
		for _, d := range i.DNS {
			iface["dns"] = append(iface["dns"].([]string), d)
		}

		ifaces = append(ifaces, iface)
	}
	if err := d.Set("interface", ifaces); err != nil {
		return fmt.Errorf("error setting interfaces: %w", err)
	}

	vrfs := make([]any, 0)
	for _, v := range c.VRFs {
		out := make(map[string]any)
		out["name"] = v.Name
		out["forwarding"] = v.Forwarding

		out["acl"] = make([]map[string]any, 0)
		for _, a := range v.ACLs {
			out["acl"] = append(out["acl"].([]map[string]any), map[string]any{
				"action":      a.Action,
				"description": a.Description,
				"dest":        a.Dest,
				"line":        a.Line,
				"protocol":    a.Protocol,
				"source":      a.Source,
			})
		}

		out["nat"] = make([]map[string]any, 0)
		for _, n := range v.NATs {
			out["nat"] = append(out["nat"].([]map[string]any), map[string]any{
				"dest":       n.Dest,
				"source":     n.Source,
				"to_source":  n.ToSource,
				"to_dest":    n.ToDest,
				"masquerade": n.Masquerade,
			})
		}

		out["rule"] = make([]map[string]any, 0)
		for _, r := range v.Rules {
			out["rule"] = append(out["rule"].([]map[string]any), map[string]any{
				"action":      r.Action,
				"description": r.Description,
				"dest":        r.Dest,
				"line":        r.Line,
				"protocol":    r.Protocol,
				"source":      r.Source,
				"vrf":         r.VRF,
			})
		}

		out["route"] = make([]map[string]any, 0)
		for _, r := range v.Routes {
			out["route"] = append(out["route"].([]map[string]any), map[string]any{
				"dest":        r.Dest,
				"dev":         r.Dev,
				"metric":      r.Metric,
				"description": r.Description,
			})
		}

		vrfs = append(vrfs, out)
	}
	if err := d.Set("vrf", vrfs); err != nil {
		return fmt.Errorf("error setting VRFs: %w", err)
	}

	return nil
}

func (nr *network) marshalResourceData(ctx context.Context, d *schema.ResourceData) (tg.NetworkConfig, error) {
	nc := tg.NetworkConfig{}

	if err := hcl.MarshalResourceData(d, &nc); err != nil {
		return nc, err
	}

	nc.Tunnels = make([]tg.NetworkTunnel, 0)
	if tunnels, ok := d.GetOk("tunnel"); ok {
		for _, t := range tunnels.([]any) {
			tunnel := t.(map[string]any)

			nc.Tunnels = append(nc.Tunnels, tg.NetworkTunnel{
				Enabled:       tunnel["enabled"].(bool),
				Name:          tunnel["name"].(string),
				IKE:           tunnel["ike"].(int),
				IKECipher:     tunnel["ike_cipher"].(string),
				IKEGroup:      tunnel["ike_group"].(int),
				RekeyInterval: tunnel["rekey_interval"].(int),
				IP:            tunnel["ip"].(string),
				Destination:   tunnel["destination"].(string),
				IPSecCipher:   tunnel["ipsec_cipher"].(string),
				PSK:           tunnel["psk"].(string),
				VRF:           tunnel["vrf"].(string),
				Type:          tunnel["type"].(string),
				MTU:           tunnel["mtu"].(int),
				NetworkID:     tunnel["network_id"].(int),
				LocalID:       tunnel["local_id"].(string),
				RemoteID:      tunnel["remote_id"].(string),
				DPDRetries:    tunnel["dpd_retries"].(int),
				DPDInterval:   tunnel["dpd_interval"].(int),
				IFace:         tunnel["iface"].(string),
				PFS:           tunnel["pfs"].(int),
				ReplayWindow:  tunnel["replay_window"].(int),
			})
		}
	}

	nc.Interfaces = make([]tg.NetworkInterface, 0)
	if interfaces, ok := d.GetOk("interface"); ok {
		for _, i := range interfaces.([]any) {
			data := i.(map[string]any)

			iface := tg.NetworkInterface{
				NIC:       data["nic"].(string),
				ClusterIP: data["cluster_ip"].(string),
				DHCP:      data["dhcp"].(bool),
				Gateway:   data["gateway"].(string),
				IP:        data["ip"].(string),
				Mode:      data["mode"].(string),
				Duplex:    data["duplex"].(string),
				Speed:     data["speed"].(int),
				DNS:       make([]string, 0),
				Routes:    make([]tg.NetworkRoute, 0),
			}

			if routes, ok := data["routes"].([]any); ok {
				for _, r := range routes {
					iface.Routes = append(iface.Routes, tg.NetworkRoute{Route: r.(string)})
				}
			}

			if dns, ok := data["dns"].([]any); ok {
				for _, d := range dns {
					iface.DNS = append(iface.DNS, d.(string))
				}
			}

			nc.Interfaces = append(nc.Interfaces, iface)
		}
	}

	nc.VRFs = make([]tg.VRF, 0)
	if vrfs, ok := d.GetOk("vrf"); ok {
		for _, v := range vrfs.([]any) {
			data := v.(map[string]any)

			vrf := tg.VRF{
				Name:       data["name"].(string),
				Forwarding: data["forwarding"].(bool),
			}

			if acls, ok := data["acl"].([]any); ok {
				for _, a := range acls {
					data := a.(map[string]any)
					vrf.ACLs = append(vrf.ACLs, tg.VRFACL{
						Action:      data["action"].(string),
						Description: data["description"].(string),
						Protocol:    data["protocol"].(string),
						Source:      data["source"].(string),
						Dest:        data["dest"].(string),
						Line:        data["line"].(int),
					})
				}
			}

			if nats, ok := data["nat"].([]any); ok {
				for _, n := range nats {
					data := n.(map[string]any)
					vrf.NATs = append(vrf.NATs, tg.VRFNAT{
						Source:     data["source"].(string),
						Dest:       data["dest"].(string),
						Masquerade: data["masquerade"].(bool),
						ToSource:   data["to_source"].(string),
						ToDest:     data["to_dest"].(string),
					})
				}
			}

			if rules, ok := data["rule"].([]any); ok {
				for _, r := range rules {
					data := r.(map[string]any)
					vrf.Rules = append(vrf.Rules, tg.VRFRule{
						Action:      data["action"].(string),
						Description: data["description"].(string),
						Dest:        data["dest"].(string),
						Line:        data["line"].(int),
						Protocol:    data["protocol"].(string),
						Source:      data["source"].(string),
						VRF:         data["vrf"].(string),
					})
				}
			}

			if routes, ok := data["route"].([]any); ok {
				for _, r := range routes {
					data := r.(map[string]any)
					vrf.Routes = append(vrf.Routes, tg.VRFRoute{
						Description: data["description"].(string),
						Dest:        data["dest"].(string),
						Dev:         data["dev"].(string),
						Metric:      data["metric"].(int),
					})
				}
			}

			nc.VRFs = append(nc.VRFs, vrf)
		}
	}

	return nc, nil
}

func (nr *network) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)
	nc, err := nr.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/network", nc.NodeID), &nc)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(nc.NodeID)

	return diag.Diagnostics{}
}

func (nr *network) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	n := tg.Node{}
	if err := tgc.Get(ctx, "/node/"+d.Id(), &n); err != nil {
		return diag.FromErr(err)
	}

	if err := nr.unmarshalResourceData(ctx, n.Config.Network, d); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("node_id", d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (nr *network) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nr.Create(ctx, d, meta)
}

func (nr *network) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	// Noop

	return diag.Diagnostics{}
}
