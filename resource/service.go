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
				Description: "Node UID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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

func (r *service) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Service{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(uuid.NewString())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	config := node.Config.Services
	config.Services = append(config.Services, payload)

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/services", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *service) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Service{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(d.Id())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	config := node.Config.Services
	for i, svc := range config.Services {
		if svc.ID == d.Id() {
			config.Services[i] = payload
		}
	}

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/services", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *service) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Service{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	services := make([]tg.Service, 0)
	for _, svc := range node.Config.Services.Services {
		if svc.ID != d.Id() {
			services = append(services, svc)
		}
	}

	config := tg.ServicesConfig{Services: services}

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/services", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func (r *service) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Service{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, svc := range node.Config.Services.Services {
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
