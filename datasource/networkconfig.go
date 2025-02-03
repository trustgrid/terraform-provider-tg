package datasource

import (
	"context"
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

		ReadContext: n.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"dark_mode": {
				Description: "Dark mode",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"forwarding": {
				Description: "Forwarding",
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
			},
			"vrf": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "VRFs",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VRF name",
						},
						"forwarding": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Enable forwarding",
						},
						"acl": {
							Description: "ACLs",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Description: "Action",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"protocol": {
										Description: "Protocol",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"source": {
										Description: "Source",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"dest": {
										Description: "Destination",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"line": {
										Description: "Line",
										Type:        schema.TypeInt,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"route": {
							Description: "Routes",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dev": {
										Description: "Dev",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"dest": {
										Description: "Destination",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"metric": {
										Description: "Metric",
										Type:        schema.TypeInt,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"nat": {
							Description: "NATs",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"source": {
										Description: "Source",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"dest": {
										Description: "Destination",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"to_source": {
										Description: "To Source",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"to_dest": {
										Description: "To Dest",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"masquerade": {
										Description: "Masquerade",
										Type:        schema.TypeBool,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"rule": {
							Description: "Rules",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"protocol": {
										Description: "Protocol",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"line": {
										Description: "Line",
										Type:        schema.TypeInt,
										Optional:    true,
										Computed:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"source": {
										Description: "Source",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"dest": {
										Description: "Destination",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"vrf": {
										Description: "VRF",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
									"action": {
										Description: "To Dest",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
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
				Computed:    true,
				Description: "Network interfaces",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"nic": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "NIC name",
						},
						"cluster_ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Cluster IP",
						},
						"subinterface": {
							Description: "VLAN interfaces",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"vlan_id": {
										Description: "VLAN ID",
										Type:        schema.TypeInt,
										Required:    true,
									},
									"ip": {
										Description: "IP CIDR",
										Type:        schema.TypeString,
										Required:    true,
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
											Type: schema.TypeString,
										},
									},
									"route": {
										Description: "VLAN routes",
										Type:        schema.TypeList,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"route": {
													Description: "Destination CIDR",
													Type:        schema.TypeString,
													Required:    true,
												},
												"next": {
													Description: "Next IP",
													Type:        schema.TypeString,
													Optional:    true,
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
						"route": {
							Description: "Interface routes",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"route": {
										Description: "Protocol",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"cloud_route": {
							Description: "Cluster interface routes",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"route": {
										Description: "Protocol",
										Type:        schema.TypeString,
										Computed:    true,
									},
									"description": {
										Description: "Description",
										Type:        schema.TypeString,
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"cluster_route_tables": {
							Description: "Cluster route tables",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"dhcp": {
							Description: "Enable DHCP",
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
						},
						"gateway": {
							Description: "Gateway IP address",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"ip": {
							Description: "IP address",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"mode": {
							Description: "Interface mode",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"duplex": {
							Description: "Interface duplex",
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
						},
						"speed": {
							Description: "Interface speed",
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
						},
						"dns": {
							Description: "DNS servers",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"tunnel": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Network tunnels",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Enable the tunnel",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tunnel name",
						},
						"ike": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "IKE",
						},
						"ike_cipher": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "IKE Cipher",
						},
						"ike_group": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "IKE Group",
						},
						"rekey_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Rekey Interval",
						},
						"ip": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "IP",
						},
						"destination": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Destination",
						},
						"ipsec_cipher": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "IPSec Cipher",
						},
						"psk": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "PSK",
							Sensitive:   true,
						},
						"vrf": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "VRF",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Tunnel type",
						},
						"mtu": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "MTU",
						},
						"network_id": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Network ID",
						},
						"local_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Local ID",
						},
						"remote_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Remote ID",
						},
						"dpd_retries": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "DPD Retries",
						},
						"dpd_interval": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "DPD Interval",
						},
						"iface": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Interface",
						},
						"pfs": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "PFS",
						},
						"remote_subnet": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Interesting traffic remote subnet",
						},
						"local_subnet": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Interesting traffic local subnet",
						},
						"replay_window": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Replay Window",
						},
						"description": {
							Description: "Description",
							Type:        schema.TypeString,
							Optional:    true,
						},
					},
				},
			},
		},
	}
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

func (nr *network) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	id, isCluster := nr.endpoint(d)

	var tf hcl.NetworkConfig
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if isCluster {
		n := tg.Cluster{}
		err := tgc.Get(ctx, "/cluster/"+id, &n)
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot lookup cluster id=%s isCluster=%t %w", id, isCluster, err))
		}
		tf.UpdateFromTG(n.Config.Network)
	} else {
		n := tg.Node{}
		err := tgc.Get(ctx, "/node/"+id, &n)
		if err != nil {
			return diag.FromErr(err)
		}
		tf.UpdateFromTG(n.Config.Network)
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
