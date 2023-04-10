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
	"golang.org/x/sync/errgroup"
)

type container struct {
}

type HCLContainerImage struct {
	Repository string `tf:"repository"`
	Tag        string `tf:"tag"`
}

type HCLContainerHealthCheck struct {
	Command     string `tf:"command"`
	Interval    int    `tf:"interval"`
	Timeout     int    `tf:"timeout"`
	StartPeriod int    `tf:"start_period"`
	Retries     int    `tf:"retries"`
}

type HCLContainerULimit struct {
	Type string `tf:"type"`
	Hard int    `tf:"hard"`
	Soft int    `tf:"soft"`
}

type HCLContainerLimit struct {
	CPUMax  int `tf:"cpu_max"`
	IORBPS  int `tf:"io_rbps"`
	IOWBPS  int `tf:"io_wbps"`
	IORIOPS int `tf:"io_riops"`
	IOWIOPS int `tf:"io_wiops"`
	MemMax  int `tf:"mem_max"`
	MemHigh int `tf:"mem_high"`

	Limits []HCLContainerULimit `tf:"limits"`
}

type HCLContainerMount struct {
	UID    string `tf:"uid"`
	Type   string `tf:"type"`
	Source string `tf:"source"`
	Dest   string `tf:"dest"`
}

type HCLContainerPortMapping struct {
	UID           string `tf:"uid"`
	Protocol      string `tf:"protocol"`
	IFace         string `tf:"iface"`
	HostPort      int    `tf:"host_port"`
	ContainerPort int    `tf:"container_port"`
}

type HCLContainerVirtualNetwork struct {
	UID           string `tf:"uid"`
	Network       string `tf:"network"`
	IP            string `tf:"ip"`
	AllowOutbound bool   `tf:"allow_outbound"`
}

type HCLContainerInterface struct {
	UID  string `tf:"uid"`
	Name string `tf:"name"`
	Dest string `tf:"dest"`
}

type HCLContainer struct {
	NodeID              string              `tf:"node_id"`
	ClusterFQDN         string              `tf:"cluster_fqdn"`
	ID                  string              `tf:"id"`
	Command             string              `tf:"command"`
	Description         string              `tf:"description"`
	Enabled             bool                `tf:"enabled"`
	ExecType            string              `tf:"exec_type"`
	Hostname            string              `tf:"hostname"`
	Image               []HCLContainerImage `tf:"image"`
	Name                string              `tf:"name"`
	Privileged          bool                `tf:"privileged"`
	RequireConnectivity bool                `tf:"require_connectivity"`
	StopTime            int                 `tf:"stop_time"`
	UseInit             bool                `tf:"use_init"`
	User                string              `tf:"user"`
	VRF                 string              `tf:"vrf"`

	AddCaps      []string                  `tf:"add_caps"`
	DropCaps     []string                  `tf:"drop_caps"`
	Variables    map[string]string         `tf:"variables"`
	Healthchecks []HCLContainerHealthCheck `tf:"healthcheck"`

	LogMaxFileSize int `tf:"log_max_file_size"`
	LogMaxNumFiles int `tf:"log_max_num_files"`

	Limits          []HCLContainerLimit          `tf:"limits"`
	Mounts          []HCLContainerMount          `tf:"mount"`
	PortMappings    []HCLContainerPortMapping    `tf:"port_mapping"`
	VirtualNetworks []HCLContainerVirtualNetwork `tf:"virtual_network"`
	Interfaces      []HCLContainerInterface      `tf:"interface"`
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

func (cr *container) urlRoot(c tg.Container) string {
	if c.NodeID != "" {
		return "/v2/node/" + c.NodeID + "/exec/container"
	}
	return "/v2/cluster/" + c.ClusterFQDN + "/exec/container"
}

func (cr *container) containerURL(c tg.Container) string {
	return cr.urlRoot(c) + "/" + c.ID
}

func (cr *container) getContainer(ctx context.Context, tgc *tg.Client, c tg.Container) (tg.Container, error) {
	res := tg.Container{}
	err := tgc.Get(ctx, cr.containerURL(c), &res)
	if err != nil {
		return res, err
	}

	res.NodeID = c.NodeID
	res.ClusterFQDN = c.ClusterFQDN

	g := errgroup.Group{}

	cc := tg.ContainerConfig{}
	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/capability", &cc.Capabilities)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/variable", &cc.Variables)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/logging", &cc.Logging)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/mount", &cc.Mounts)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/port-mapping", &cc.PortMappings)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/virtual-network", &cc.VirtualNetworks)
		if err != nil {
			return err
		}

		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, cr.containerURL(c)+"/interface", &cc.Interfaces)
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
	return tgc.Put(ctx, cr.containerURL(c)+"/config", c.Config)
}

func (cr *container) convertToTFConfig(c tg.Container, d *schema.ResourceData) error {
	tfc := HCLContainer{
		NodeID:      c.NodeID,
		ClusterFQDN: c.ClusterFQDN,
		ID:          c.ID,
		Command:     c.Command,
		Description: c.Description,
		Enabled:     c.Enabled,
		ExecType:    c.ExecType,
		Hostname:    c.Hostname,
		Image: []HCLContainerImage{
			{Repository: c.Image.Repository, Tag: c.Image.Tag},
		},
		Name:                c.Name,
		Privileged:          c.Privileged,
		RequireConnectivity: c.RequireConnectivity,
		StopTime:            c.StopTime,
		UseInit:             c.UseInit,
		User:                c.User,
		Variables:           make(map[string]string),
	}

	if c.Config.VRF != nil {
		tfc.VRF = c.Config.VRF.Name
	}

	tfc.AddCaps = append(tfc.AddCaps, c.Config.Capabilities.AddCaps...)
	tfc.DropCaps = append(tfc.DropCaps, c.Config.Capabilities.DropCaps...)

	for _, v := range c.Config.Variables {
		tfc.Variables[v.Name] = v.Value
	}

	if c.Config.HealthCheck != nil {
		hc := c.Config.HealthCheck
		tfc.Healthchecks = []HCLContainerHealthCheck{
			{
				Command:     hc.Command,
				Interval:    hc.Interval,
				Timeout:     hc.Timeout,
				StartPeriod: hc.StartPeriod,
				Retries:     hc.Retries,
			},
		}
	}

	if c.Config.Logging.MaxFileSize > 0 {
		tfc.LogMaxFileSize = c.Config.Logging.MaxFileSize
	}
	if c.Config.Logging.NumFiles > 0 {
		tfc.LogMaxNumFiles = c.Config.Logging.NumFiles
	}

	if c.Config.Limits != nil {
		limits := c.Config.Limits
		tlimit := HCLContainerLimit{
			CPUMax:  limits.CPUMax,
			IORBPS:  limits.IORBPS,
			IORIOPS: limits.IORIOPS,
			IOWBPS:  limits.IOWBPS,
			IOWIOPS: limits.IOWIOPS,
			MemMax:  limits.MemMax,
			MemHigh: limits.MemHigh,
		}

		for _, l := range limits.Limits {
			tlimit.Limits = append(tlimit.Limits, HCLContainerULimit{
				Type: l.Type,
				Hard: l.Hard,
				Soft: l.Soft,
			})
		}
		tfc.Limits = []HCLContainerLimit{tlimit}
	}

	for _, m := range c.Config.Mounts {
		tfc.Mounts = append(tfc.Mounts, HCLContainerMount{
			UID:    m.UID,
			Type:   m.Type,
			Source: m.Source,
			Dest:   m.Dest,
		})
	}

	for _, pm := range c.Config.PortMappings {
		tfc.PortMappings = append(tfc.PortMappings, HCLContainerPortMapping{
			UID:           pm.UID,
			Protocol:      pm.Protocol,
			IFace:         pm.IFace,
			HostPort:      pm.HostPort,
			ContainerPort: pm.ContainerPort,
		})
	}

	for _, vnet := range c.Config.VirtualNetworks {
		tfc.VirtualNetworks = append(tfc.VirtualNetworks, HCLContainerVirtualNetwork{
			UID:           vnet.UID,
			Network:       vnet.Network,
			IP:            vnet.IP,
			AllowOutbound: vnet.AllowOutbound,
		})
	}

	for _, i := range c.Config.Interfaces {
		tfc.Interfaces = append(tfc.Interfaces, HCLContainerInterface{
			UID:  i.UID,
			Name: i.Name,
			Dest: i.Dest,
		})
	}

	return hcl.EncodeResourceData(&tfc, d)
}

func (cr *container) decodeTFConfig(_ context.Context, d *schema.ResourceData) (tg.Container, error) {
	tfc := HCLContainer{}
	c := tg.Container{}

	if err := hcl.DecodeResourceData(d, &tfc); err != nil {
		return c, err
	}

	c.NodeID = tfc.NodeID
	c.ClusterFQDN = tfc.ClusterFQDN
	c.ID = tfc.ID
	c.Command = tfc.Command
	c.Description = tfc.Description
	c.Enabled = tfc.Enabled
	c.ExecType = tfc.ExecType
	c.Hostname = tfc.Hostname
	c.Image.Repository = tfc.Image[0].Repository
	c.Image.Tag = tfc.Image[0].Tag
	c.Name = tfc.Name
	c.Privileged = tfc.Privileged
	c.RequireConnectivity = tfc.RequireConnectivity
	c.StopTime = tfc.StopTime
	c.UseInit = tfc.UseInit
	c.User = tfc.User

	cc := &c.Config

	cc.Capabilities.AddCaps = append(cc.Capabilities.AddCaps, tfc.AddCaps...)
	cc.Capabilities.DropCaps = append(cc.Capabilities.DropCaps, tfc.DropCaps...)

	for k, v := range tfc.Variables {
		cc.Variables = append(cc.Variables, tg.ContainerVar{Name: k, Value: v})
	}

	cc.Logging.MaxFileSize = tfc.LogMaxFileSize
	cc.Logging.NumFiles = tfc.LogMaxNumFiles

	if len(tfc.Healthchecks) > 0 {
		hc := tfc.Healthchecks[0]
		cc.HealthCheck = &tg.HealthCheck{
			Command:     hc.Command,
			Interval:    hc.Interval,
			Timeout:     hc.Timeout,
			StartPeriod: hc.StartPeriod,
			Retries:     hc.Retries,
		}
	}

	if len(tfc.Limits) > 0 {
		limit := tfc.Limits[0]
		cc.Limits = &tg.ContainerLimits{
			CPUMax:  limit.CPUMax,
			IORBPS:  limit.IORBPS,
			IOWBPS:  limit.IOWBPS,
			IORIOPS: limit.IORIOPS,
			IOWIOPS: limit.IOWIOPS,
			MemMax:  limit.MemMax,
			MemHigh: limit.MemHigh,
		}

		for _, l := range limit.Limits {
			cc.Limits.Limits = append(cc.Limits.Limits, tg.ULimit{
				Type: l.Type,
				Hard: l.Hard,
				Soft: l.Soft,
			})
		}
	}

	for _, m := range tfc.Mounts {
		mount := tg.Mount{
			UID:    m.UID,
			Type:   m.Type,
			Source: m.Source,
			Dest:   m.Dest,
		}
		if m.UID == "" {
			m.UID = uuid.NewString()
		}
		cc.Mounts = append(cc.Mounts, mount)
	}

	for _, m := range tfc.PortMappings {
		pm := tg.PortMapping{
			UID:           m.UID,
			Protocol:      m.Protocol,
			IFace:         m.IFace,
			HostPort:      m.HostPort,
			ContainerPort: m.ContainerPort,
		}
		if pm.UID == "" {
			pm.UID = uuid.NewString()
		}
		cc.PortMappings = append(cc.PortMappings, pm)
	}

	for _, vn := range tfc.VirtualNetworks {
		vnet := tg.ContainerVirtualNetwork{
			UID:           vn.UID,
			Network:       vn.Network,
			IP:            vn.IP,
			AllowOutbound: vn.AllowOutbound,
		}
		if vnet.UID == "" {
			vnet.UID = uuid.NewString()
		}
		cc.VirtualNetworks = append(cc.VirtualNetworks, vnet)
	}

	for _, i := range tfc.Interfaces {
		iface := tg.ContainerInterface{
			UID:  i.UID,
			Name: i.Name,
			Dest: i.Dest,
		}
		if iface.UID == "" {
			iface.UID = uuid.NewString()
		}
		cc.Interfaces = append(cc.Interfaces, iface)
	}

	if tfc.VRF != "" {
		cc.VRF = &tg.ContainerVRF{Name: tfc.VRF}
	}

	return c, nil
}

func (cr *container) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := cr.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ct.ID = uuid.New().String()

	if _, err := tgc.Post(ctx, cr.urlRoot(ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cr *container) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := cr.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, cr.containerURL(ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	if err := cr.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cr *container) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	ct, err := cr.decodeTFConfig(ctx, d)
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

	tf, err := cr.decodeTFConfig(ctx, d)
	if err != nil {
		return diag.FromErr(err)
	}

	ct, err := cr.getContainer(ctx, tgc, tf)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if err := cr.convertToTFConfig(ct, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
