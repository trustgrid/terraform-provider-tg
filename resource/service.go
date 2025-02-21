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
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type service struct {
}

func Service() *schema.Resource {
	r := service{}

	return &schema.Resource{
		Description: "Node or Cluster Service",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID - required if cluster_fqdn not set",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - required if node_id not set",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Protocol",
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp", "tftp", "ftp", "rdp", "vnc", "ssh"}, false),
			},
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Host",
			},
			"port": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Port",
				ValidateFunc: validation.IsPortNumber,
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description",
			},
		},
	}
}

func (r *service) getConfig(ctx context.Context, tgc *tg.Client, tf hcl.Service) (tg.ServicesConfig, error) {
	if tf.NodeID != "" {
		node := tg.Node{}
		if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
			return tg.ServicesConfig{}, err
		}
		return node.Config.Services, nil
	}
	cluster := tg.Cluster{}
	if err := tgc.Get(ctx, fmt.Sprintf("/cluster/%s", tf.ClusterFQDN), &cluster); err != nil {
		return tg.ServicesConfig{}, err
	}
	if cluster.Config.Services != nil {
		return *cluster.Config.Services, nil
	}
	return tg.ServicesConfig{}, nil
}

func (r *service) writeConfig(ctx context.Context, tgc *tg.Client, tf hcl.Service, config tg.ServicesConfig) error {
	url := fmt.Sprintf("/node/%s/config/services", tf.NodeID)
	if tf.NodeID == "" {
		url = fmt.Sprintf("/cluster/%s/config/services", tf.ClusterFQDN)
	}

	_, err := tgc.Put(ctx, url, config)
	return err
}

func (r *service) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Service](d)
	if err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(uuid.NewString())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	config.Services = append(config.Services, payload)

	if err := r.writeConfig(ctx, tgc, tf, config); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *service) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Service](d)
	if err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(d.Id())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, svc := range config.Services {
		if svc.ID == d.Id() {
			config.Services[i] = payload
		}
	}

	if err := r.writeConfig(ctx, tgc, tf, config); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *service) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Service](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	services := make([]tg.Service, 0)
	for _, svc := range config.Services {
		if svc.ID != d.Id() {
			services = append(services, svc)
		}
	}

	updated := tg.ServicesConfig{Services: services}

	if err := r.writeConfig(ctx, tgc, tf, updated); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func (r *service) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Service](d)
	if err != nil {
		return diag.FromErr(err)
	}

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, svc := range config.Services {
		if svc.ID == d.Id() {
			tf.UpdateFromTG(svc)
			found = true
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
