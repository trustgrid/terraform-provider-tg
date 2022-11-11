package resource

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"golang.org/x/sync/errgroup"
)

type container struct {
}

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
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
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
							Description:  "Internal interface destination",
							ValidateFunc: validation.IsCIDR,
						},
					},
				},
			},
		},
	}
}

func (cr *container) urlRoot(tgc *tg.Client, c tg.Container) string {
	if c.NodeID != "" {
		return "/v2/node/" + c.NodeID + "/exec/container"
	}
	return "/v2/cluster/" + c.ClusterFQDN + "/exec/container"
}

func (cr *container) containerURL(tgc *tg.Client, c tg.Container) string {
	return cr.urlRoot(tgc, c) + "/" + c.ID
}

func (cr *container) getContainer(ctx context.Context, tgc *tg.Client, c tg.Container) (tg.Container, error) {
	res := tg.Container{}
	err := tgc.Get(ctx, cr.containerURL(tgc, c), &res)
	if err != nil {
		return res, err
	}

	res.NodeID = c.NodeID
	res.ClusterFQDN = c.ClusterFQDN

	g := errgroup.Group{}

	cc := tg.ContainerConfig{}
	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/capability", &cc.Capabilities)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/variable", &cc.Variables)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/logging", &cc.Logging)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/mount", &cc.Mounts)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/port-mapping", &cc.PortMappings)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/virtual-network", &cc.VirtualNetworks)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(tgc, c)+"/interface", &cc.Interfaces)
		if err != nil {
			return err
		}

		return nil
	})

	err = g.Wait()
	res.Config = cc
	return res, err
}

func (cr *container) writeExtendedConfig(ctx context.Context, tgc *tg.Client, c tg.Container) error {
	return tgc.Put(ctx, cr.containerURL(tgc, c)+"/config", c.Config)
}

func (cr *container) unmarshalResourceData(ctx context.Context, c tg.Container, d *schema.ResourceData) error {
	if err := hcl.UnmarshalResourceData(&c, d); err != nil {
		return err
	}

	mounts := make([]any, 0)
	for _, m := range c.Config.Mounts {
		mount := make(map[string]any)
		mount["uid"] = m.UID
		mount["type"] = m.Type
		mount["source"] = m.Source
		mount["dest"] = m.Dest
		mounts = append(mounts, mount)
	}
	if err := d.Set("mount", mounts); err != nil {
		return fmt.Errorf("error setting mounts: %w", err)
	}

	mappings := make([]any, 0)
	for _, m := range c.Config.PortMappings {
		mapping := make(map[string]any)
		mapping["uid"] = m.UID
		mapping["protocol"] = m.Protocol
		mapping["iface"] = m.IFace
		mapping["host_port"] = m.HostPort
		mapping["container_port"] = m.ContainerPort
		mappings = append(mappings, mapping)
	}
	if err := d.Set("port_mapping", mappings); err != nil {
		return fmt.Errorf("error setting port mappings: %w", err)
	}

	vnets := make([]any, 0)
	for _, m := range c.Config.VirtualNetworks {
		vnet := make(map[string]any)
		vnet["uid"] = m.UID
		vnet["network"] = m.Network
		vnet["ip"] = m.IP
		vnet["allow_outbound"] = m.AllowOutbound
		vnets = append(vnets, vnet)
	}
	if err := d.Set("virtual_network", vnets); err != nil {
		return fmt.Errorf("error setting virtual_network: %w", err)
	}

	ifaces := make([]any, 0)
	for _, m := range c.Config.Interfaces {
		iface := make(map[string]any)
		iface["uid"] = m.UID
		iface["name"] = m.Name
		iface["dest"] = m.Dest
		ifaces = append(ifaces, iface)
	}
	if err := d.Set("interface", ifaces); err != nil {
		return fmt.Errorf("error setting interface: %w", err)
	}
	return nil
}

func (cr *container) marshalResourceData(ctx context.Context, d *schema.ResourceData) (tg.Container, error) {
	ct := tg.Container{}

	if err := hcl.MarshalResourceData(d, &ct); err != nil {
		return ct, err
	}

	images, ok := d.Get("image").([]any)
	if !ok {
		return ct, fmt.Errorf("error getting image")
	}
	if len(images) == 0 {
		return ct, fmt.Errorf("no image specified")
	}
	image := images[0].(map[string]any)

	ct.Image.Tag = image["tag"].(string)
	ct.Image.Repository = image["repository"].(string)
	cc := &ct.Config

	cc.Capabilities.AddCaps = make([]string, 0)
	if caps, ok := d.GetOk("add_caps"); ok {
		for _, c := range caps.([]any) {
			cc.Capabilities.AddCaps = append(cc.Capabilities.AddCaps, c.(string))
		}
	}

	cc.Capabilities.DropCaps = make([]string, 0)
	if caps, ok := d.GetOk("drop_caps"); ok {
		for _, c := range caps.([]any) {
			cc.Capabilities.DropCaps = append(cc.Capabilities.DropCaps, c.(string))
		}
	}

	cc.Variables = make([]tg.ContainerVar, 0)
	if vars, ok := d.GetOk("variables"); ok {
		for k, v := range vars.(map[string]any) {
			cc.Variables = append(cc.Variables, tg.ContainerVar{Name: k, Value: v.(string)})
		}
	}

	if lfs, ok := d.GetOk("log_max_file_size"); ok {
		cc.Logging.MaxFileSize = lfs.(int)
	}

	if lnf, ok := d.GetOk("log_max_num_files"); ok {
		cc.Logging.NumFiles = lnf.(int)
	}

	healthchecks, ok := d.Get("healthcheck").([]any)
	if ok && len(healthchecks) > 0 {
		healthcheck := healthchecks[0].(map[string]any)
		cc.HealthCheck = &tg.HealthCheck{
			Command:     healthcheck["command"].(string),
			Interval:    healthcheck["interval"].(int),
			Timeout:     healthcheck["timeout"].(int),
			StartPeriod: healthcheck["start_period"].(int),
			Retries:     healthcheck["retries"].(int),
		}
	}

	limits, ok := d.Get("limits").([]any)
	if ok && len(limits) > 0 {
		limit := limits[0].(map[string]any)
		cc.Limits = &tg.ContainerLimits{
			CPUMax:  limit["cpu_max"].(int),
			IORBPS:  limit["io_rbps"].(int),
			IOWBPS:  limit["io_wbps"].(int),
			IORIOPS: limit["io_riops"].(int),
			IOWIOPS: limit["io_wiops"].(int),
			MemMax:  limit["mem_max"].(int),
			MemHigh: limit["mem_high"].(int),
		}
		if limits, ok := limit["limits"].([]any); ok {
			for _, l := range limits {
				limit := l.(map[string]any)
				cc.Limits.Limits = append(cc.Limits.Limits, tg.ULimit{
					Type: limit["type"].(string),
					Hard: limit["hard"].(int),
					Soft: limit["soft"].(int),
				})
			}
		}
	}

	cc.Mounts = make([]tg.Mount, 0)
	if mounts, ok := d.Get("mount").([]any); ok {
		for _, m := range mounts {
			mount := m.(map[string]any)
			_, ok := mount["uid"]
			if !ok {
				mount["uid"] = uuid.NewString()
			}
			cc.Mounts = append(cc.Mounts, tg.Mount{
				UID:    mount["uid"].(string),
				Type:   mount["type"].(string),
				Source: mount["source"].(string),
				Dest:   mount["dest"].(string),
			})
		}
		if err := d.Set("mount", mounts); err != nil {
			return ct, fmt.Errorf("error updating mount: %w", err)
		}
	}

	cc.PortMappings = make([]tg.PortMapping, 0)
	if mappings, ok := d.Get("port_mapping").([]any); ok {
		for _, m := range mappings {
			mapping := m.(map[string]any)
			_, ok := mapping["uid"]
			if !ok {
				mapping["uid"] = uuid.NewString()
			}
			cc.PortMappings = append(cc.PortMappings, tg.PortMapping{
				UID:           mapping["uid"].(string),
				Protocol:      mapping["protocol"].(string),
				IFace:         mapping["iface"].(string),
				HostPort:      mapping["host_port"].(int),
				ContainerPort: mapping["container_port"].(int),
			})
		}
		if err := d.Set("port_mapping", mappings); err != nil {
			return ct, fmt.Errorf("error updating port_mapping: %w", err)
		}
	}

	cc.VirtualNetworks = make([]tg.ContainerVirtualNetwork, 0)
	if vnets, ok := d.Get("virtual_network").([]any); ok {
		for _, m := range vnets {
			vnet := m.(map[string]any)
			_, ok := vnet["uid"]
			if !ok {
				vnet["uid"] = uuid.NewString()
			}
			cc.VirtualNetworks = append(cc.VirtualNetworks, tg.ContainerVirtualNetwork{
				UID:           vnet["uid"].(string),
				Network:       vnet["network"].(string),
				IP:            vnet["ip"].(string),
				AllowOutbound: vnet["allow_outbound"].(bool),
			})
		}
		if err := d.Set("virtual_network", vnets); err != nil {
			return ct, fmt.Errorf("error updating virtual_network: %w", err)
		}
	}

	cc.Interfaces = make([]tg.ContainerInterface, 0)
	if ifaces, ok := d.Get("interface").([]any); ok {
		for _, i := range ifaces {
			iface := i.(map[string]any)
			_, ok := iface["uid"]
			if !ok {
				iface["uid"] = uuid.NewString()
			}
			cc.Interfaces = append(cc.Interfaces, tg.ContainerInterface{
				UID:  iface["uid"].(string),
				Name: iface["name"].(string),
				Dest: iface["dest"].(string),
			})
		}
		if err := d.Set("interface", ifaces); err != nil {
			return ct, fmt.Errorf("error updating interface: %w", err)
		}
	}

	if vrf, ok := d.GetOk("vrf"); ok {
		cc.VRF = &tg.ContainerVRF{Name: vrf.(string)}
	}

	return ct, nil
}

func (cr *container) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct, err := cr.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ct.ID = uuid.New().String()

	if err := tgc.Post(ctx, cr.urlRoot(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (cr *container) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct, err := cr.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, cr.containerURL(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cr *container) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct, err := cr.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, cr.containerURL(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (cr *container) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct, err := cr.marshalResourceData(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ct, err = cr.getContainer(ctx, tgc, ct)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := cr.unmarshalResourceData(ctx, ct, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	return diag.Diagnostics{}
}
