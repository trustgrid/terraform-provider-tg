package resource

import (
	"context"

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
			"image_repository": {
				Description: "Image repository",
				Type:        schema.TypeString,
				Required:    true,
			},
			"image_tag": {
				Description: "Image tag",
				Type:        schema.TypeString,
				Required:    true,
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
		},
	}
}

func (c *container) urlRoot(tgc *tg.Client, ct tg.Container) string {
	if ct.NodeID != "" {
		return "/v2/node/" + ct.NodeID + "/exec/container"
	}
	return "/v2/cluster/" + ct.ClusterFQDN + "/exec/container"
}

func (c *container) containerURL(tgc *tg.Client, ct tg.Container) string {
	return c.urlRoot(tgc, ct) + "/" + ct.ID
}

func (c *container) getContainer(ctx context.Context, tgc *tg.Client, ct tg.Container) (tg.Container, error) {
	res := tg.Container{}
	err := tgc.Get(ctx, c.containerURL(tgc, ct), &res)
	if err != nil {
		return res, err
	}

	res.ImageRepository = res.Image.Repository
	res.ImageTag = res.Image.Tag
	res.NodeID = ct.NodeID
	res.ClusterFQDN = ct.ClusterFQDN

	g := errgroup.Group{}

	cc := tg.ContainerConfig{}
	g.Go(func() error {
		err = tgc.Get(ctx, c.containerURL(tgc, ct)+"/capability", &cc.Capabilities)
		if err != nil {
			return err
		}

		res.AddCaps = make([]interface{}, 0)
		for _, c := range cc.Capabilities.AddCaps {
			res.AddCaps = append(res.AddCaps, c)
		}
		res.DropCaps = make([]interface{}, 0)
		for _, c := range cc.Capabilities.DropCaps {
			res.DropCaps = append(res.DropCaps, c)
		}
		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, c.containerURL(tgc, ct)+"/variable", &cc.Variables)
		if err != nil {
			return err
		}

		res.Variables = make(map[string]interface{})
		for _, v := range cc.Variables {
			res.Variables[v.Name] = v.Value
		}
		return nil
	})

	g.Go(func() error {
		err = tgc.Get(ctx, c.containerURL(tgc, ct)+"/logging", &cc.Logging)
		if err != nil {
			return err
		}

		res.LogMaxFileSize = cc.Logging.MaxFileSize
		res.LogMaxNumFiles = cc.Logging.NumFiles
		return nil
	})

	err = g.Wait()
	return res, err
}

func (c *container) writeExtendedConfig(ctx context.Context, tgc *tg.Client, ct tg.Container) error {
	return tgc.Put(ctx, c.containerURL(tgc, ct)+"/config", ct.Config())
}

func (c *container) Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct := tg.Container{}
	if err := hcl.MarshalResourceData(d, &ct); err != nil {
		return diag.FromErr(err)
	}

	ct.Image.Repository = ct.ImageRepository
	ct.Image.Tag = ct.ImageTag
	ct.ID = uuid.New().String()

	if err := tgc.Post(ctx, c.urlRoot(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	if err := c.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (c *container) Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct := tg.Container{}
	if err := hcl.MarshalResourceData(d, &ct); err != nil {
		return diag.FromErr(err)
	}

	ct.Image.Repository = ct.ImageRepository
	ct.Image.Tag = ct.ImageTag

	if err := tgc.Put(ctx, c.containerURL(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	if err := c.writeExtendedConfig(ctx, tgc, ct); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (c *container) Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct := tg.Container{}
	if err := hcl.MarshalResourceData(d, &ct); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, c.containerURL(tgc, ct), &ct); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func (c *container) Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	ct := tg.Container{}
	if err := hcl.MarshalResourceData(d, &ct); err != nil {
		return diag.FromErr(err)
	}

	ct, err := c.getContainer(ctx, tgc, ct)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.UnmarshalResourceData(&ct, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ct.ID)

	return diag.Diagnostics{}
}
