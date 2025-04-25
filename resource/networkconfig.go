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
	"github.com/trustgrid/terraform-provider-tg/validators"
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
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
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
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "Cluster IP",
							ValidateFunc:  validation.IsIPv4Address,
							ConflictsWith: []string{"node_id"},
						},
						"route": {
							Description: "Interface routes",
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
									"next_hop": {
										Description:  "Next Hop",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.IsIPv4Address,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"subinterface": {
							Description: "VLAN interfaces",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vlan_id": {
										Description:  "VLAN ID",
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(0, 4095),
									},
									"ip": {
										Description:  "IP CIDR",
										Type:         schema.TypeString,
										ValidateFunc: validation.IsCIDR,
										Required:     true,
									},
									"vrf": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "VRF",
									},
									"additional_ips": {
										Description: "Additional IP CIDRs",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.IsCIDR,
										},
									},
									"route": {
										Description: "VLAN routes",
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
													Description:  "Next IP",
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsIPv4Address,
												},
												"description": {
													Description: "Description",
													Type:        schema.TypeString,
													Optional:    true,
												},
											},
										},
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"cloud_route": {
							Description: "Cluster interface routes",
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
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"cluster_route_tables": {
							Description: "Cluster route tables - should be a list of either AWS or Azure route table IDs",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
							Description:   "IP address",
							Type:          schema.TypeString,
							Optional:      true,
							ValidateFunc:  validation.IsCIDR,
							ConflictsWith: []string{"cluster_fqdn"},
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
						"remote_subnet": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Interesting traffic remote subnet",
							ValidateFunc: validation.IsCIDR,
						},
						"local_subnet": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Interesting traffic local subnet",
							ValidateFunc: validation.IsCIDR,
						},
						"replay_window": {
							Type:         schema.TypeInt,
							Optional:     true,
							Description:  "Replay Window",
							ValidateFunc: validation.IntInSlice([]int{32, 64, 128, 256, 512, 1024, 2048, 4096, 8192}),
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description",
						},
					},
				},
			},
		},
	}
}

func (nr *network) decodeTFConfig(_ context.Context, d *schema.ResourceData) (tg.NetworkConfig, error) {
	tf, err := hcl.DecodeResourceData[hcl.NetworkConfig](d)
	if err != nil {
		return tg.NetworkConfig{}, err
	}

	return tf.ToTG(), nil
}

func (nr *network) endpoint(d *schema.ResourceData) (string, bool) {
	nodeid, ok := d.GetOk("node_id")
	isCluster := false
	if !ok || nodeid == "" {
		id, ok := d.Get("cluster_fqdn").(string)
		if !ok {
			panic("network resource: no node_id and no cluster_fqdn")
		}
		isCluster = true
		return id, isCluster
	}
	id, ok := nodeid.(string)
	if !ok {
		panic("node_id must be a string")
	}
	return id, isCluster
}

func (nr *network) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)
	nc, err := nr.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	id, isCluster := nr.endpoint(d)

	if isCluster {
		_, err = tgc.Put(ctx, fmt.Sprintf("/cluster/%s/config/network", id), &nc)
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		_, err = tgc.Put(ctx, fmt.Sprintf("/node/%s/config/network", id), &nc)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(id)

	return nil
}

func (nr *network) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	id, isCluster := nr.endpoint(d)

	tf, err := hcl.DecodeResourceData[hcl.NetworkConfig](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if isCluster {
		n := tg.Cluster{}
		err := tgc.Get(ctx, "/cluster/"+id, &n)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return nil
		case err != nil:
			return diag.FromErr(fmt.Errorf("cannot lookup cluster id=%s isCluster=%t %w", id, isCluster, err))
		}

		if n.Config.Network != nil {
			tf.UpdateFromTG(*n.Config.Network)
		} else {
			tf.UpdateFromTG(tg.NetworkConfig{})
		}
	} else {
		n := tg.Node{}
		err := tgc.Get(ctx, "/node/"+id, &n)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return nil
		case err != nil:
			return diag.FromErr(err)
		}

		tf.UpdateFromTG(n.Config.Network)
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (nr *network) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nr.Create(ctx, d, meta)
}

func (nr *network) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	// Noop

	return nil
}
