package resource

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
	"golang.org/x/sync/errgroup"
)

type container struct {
}

// Container manages a node or cluster container.
func Container() *schema.Resource {
	c := container{}

	return &schema.Resource{
		Description: "Manage a node or cluster container",

		ReadContext:   c.Read,
		UpdateContext: c.Update,
		DeleteContext: c.Delete,
		CreateContext: c.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"id": {
				Description: "Container ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Container Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"command": {
				Description: "Command",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"exec_type": {
				Description:  `Container execution type - one of "onDemand", "service", or "recurring"`,
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"onDemand", "service", "recurring"}, false),
			},
			"hostname": {
				Description: "Host name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"image": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"repository": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Image repository",
						},
						"tag": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Image tag",
						},
					},
				},
			},
			"privileged": {
				Description: "Grant extended privileges to the container",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"require_connectivity": {
				Description: "Ensures that a container that has encrypted volumes won't start unless the node has connectivity to the control plane",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"stop_time": {
				Description: "Time to wait, in seconds, for container to stop gracefully",
				Default:     30,
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"use_init": {
				Description: "Indicates that an init process should be used as PID 1 in the container. Ensures responsibilities of an init system are performed inside the container (i.e., handling exit signals)",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"user": {
				Description: "User",
				Type:        schema.TypeString,
				Optional:    true,
			},

			// extended config
			"add_caps": {
				Description: "Add Linux capabilities from the container to have fine grain control over kernel features and device access",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"AUDIT_CONTROL",
						"AUDIT_READ",
						"BLOCK_SUSPEND",
						"BPF",
						"CHECKPOINT_RESTORE",
						"DAC_READ_SEARCH",
						"IPC_LOCK",
						"IPC_OWNER",
						"LEASE",
						"LINUX_IMMUTABLE",
						"MAC_ADMIN",
						"MAC_OVERRIDE",
						"NET_ADMIN",
						"NET_BROADCAST",
						"PERFMON",
						"SYS_ADMIN",
						"SYS_BOOT",
						"SYS_MODULE",
						"SYS_NICE",
						"SYS_PACCT",
						"SYS_PTRACE",
						"SYS_RAWIO",
						"SYS_RESOURCE",
						"SYS_TIME",
						"SYS_TTY_CONFIG",
						"SYSLOG",
						"WAKE_ALARM",
					}, false),
				},
			},
			"drop_caps": {
				Description: "Drop Linux capabilities from the container to have fine grain control over kernel features and device access",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"AUDIT_WRITE",
						"CHOWN",
						"DAC_OVERRIDE",
						"FOWNER",
						"FSETID",
						"KILL",
						"MKNOD",
						"NET_BIND_SERVICE",
						"NET_RAW",
						"SETFCAP",
						"SETGID",
						"SETPCAP",
						"SETUID",
						"SYS_CHROOT",
					}, false),
				},
			},
			"variables": {
				Description: "Environment variables",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vrf": {
				Description: "Container VRF",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"log_max_file_size": {
				Description: "Maximum log file size (MB)",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"log_max_num_files": {
				Description: "Maximum log files to keep",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"healthcheck": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"command": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Command",
						},
						"interval": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Interval",
						},
						"timeout": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Timeout",
						},
						"start_period": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Grace period before health checks are monitored, in seconds",
						},
						"retries": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Number of health checks that must fail before a container is considered unhealthy",
						},
					},
				},
			},
			"mount": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Mount ID (for API use only)",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Mount type - `volume` or `bind`",
							ValidateFunc: validation.StringInSlice([]string{"volume", "bind"}, false),
						},
						"source": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "For `volume` mounts, the name of the volume. For `bind` mounts, the path to the file or directory on the host datastore",
						},
						"dest": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Destination path in the container filesystem",
						},
					},
				},
			},
			"limits": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_max": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "CPU max allocation %",
						},
						"io_rbps": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max allowed read throughput (bytes per second)",
						},
						"io_wbps": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max allowed write throughput (bytes per second)",
						},
						"io_riops": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max allowed read throughput (IOPS)",
						},
						"io_wiops": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Max allowed write throughput (IOPS)",
						},
						"mem_high": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Soft RAM allocation limit (MB)",
						},
						"mem_max": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Hard RAM allocation limit (MB)",
						},
						"limits": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Linux kernel limits",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Limit type",
										ValidateFunc: validation.StringInSlice([]string{
											"core",
											"cpu",
											"data",
											"fsize",
											"locks",
											"memlock",
											"msgqueue",
											"nice",
											"nofile",
											"nproc",
											"rss",
											"rtprio",
											"rttime",
											"sigpending",
											"stack",
										}, false),
									},
									"hard": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Hard limit",
									},
									"soft": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Soft limit",
									},
								},
							},
						},
					},
				},
			},
			"port_mapping": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Port Mapping ID (for API use only)",
						},
						"protocol": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Protocol - `tcp` or `udp`",
							ValidateFunc: validation.StringInSlice([]string{"tcp", "udp"}, false),
						},
						"iface": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Host interface to expose port",
						},
						"host_port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Host port",
						},
						"container_port": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Container port",
						},
					},
				},
			},
			"virtual_network": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VNet ID (for API use only)",
						},
						"network": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Virtual network name to attach",
						},
						"ip": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Virtual IP address",
						},
						"allow_outbound": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Allow outbound connections on this network",
						},
					},
				},
			},
			"interface": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Interface ID (for API use only)",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Virtual interface name",
						},
						"dest": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Internal interface destination IP",
							ValidateFunc: validation.IsIPv4Address,
						},
					},
				},
			},
		},
	}
}

func (cr *container) urlRoot(c hcl.Container) string {
	if c.NodeID != "" {
		return "/v2/node/" + c.NodeID + "/exec/container"
	}
	return "/v2/cluster/" + c.ClusterFQDN + "/exec/container"
}

func (cr *container) containerURL(c hcl.Container) string {
	return cr.urlRoot(c) + "/" + c.ID
}

func (cr *container) getContainer(ctx context.Context, tgc *tg.Client, c hcl.Container) (tg.Container, error) {
	res := tg.Container{}
	err := tgc.Get(ctx, cr.containerURL(c), &res)
	if err != nil {
		return res, err
	}

	res.NodeID = c.NodeID
	res.ClusterFQDN = c.ClusterFQDN

	g := errgroup.Group{}

	containerURL := cr.containerURL(c)

	cc := tg.ContainerConfig{}
	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/healthcheck", &cc.HealthCheck)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/limit", &cc.Limits)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/capability", &cc.Capabilities)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/variable", &cc.Variables)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/logging", &cc.Logging)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/mount", &cc.Mounts)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/port-mapping", &cc.PortMappings)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/virtual-network", &cc.VirtualNetworks)
	})

	g.Go(func() error {
		return tgc.Get(ctx, containerURL+"/interface", &cc.Interfaces)
	})

	err = g.Wait()
	res.Config = cc
	return res, err
}

func (cr *container) writeExtendedConfig(ctx context.Context, tgc *tg.Client, c hcl.Container) error {
	_, err := tgc.Put(ctx, cr.containerURL(c)+"/config", c.ToTG().Config)
	return err
}

func (cr *container) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := hcl.DecodeResourceData[hcl.Container](d)
	if err != nil {
		return diag.FromErr(err)
	}
	ct.SetUIDs()

	ct.ID = uuid.New().String()

	if _, err := tgc.Post(ctx, cr.urlRoot(ct), ct.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(ct, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func (cr *container) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := hcl.DecodeResourceData[hcl.Container](d)
	if err != nil {
		return diag.FromErr(err)
	}
	ct.SetUIDs()

	if _, err := tgc.Put(ctx, cr.containerURL(ct), ct.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(ct, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cr *container) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := hcl.DecodeResourceData[hcl.Container](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, cr.containerURL(ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cr *container) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Container](d)
	if err != nil {
		return diag.FromErr(err)
	}

	ct, err := cr.getContainer(ctx, tgc, tf)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(ct)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
