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

type connector struct {
}

func Connector() *schema.Resource {
	r := connector{}

	return &schema.Resource{
		Description: "Node or Cluster Connector",

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
			"node": {
				Description: "Node name providing the service",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"service": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service name",
			},
			"protocol": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Protocol",
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp", "tftp", "ftp"}, false),
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

func (r *connector) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Connector{}
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

	config := node.Config.Connectors
	config.Connectors = append(config.Connectors, payload)

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/connectors", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *connector) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Connector{}
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

	config := node.Config.Connectors
	for i, conn := range config.Connectors {
		if conn.ID == d.Id() {
			config.Connectors[i] = payload
		}
	}

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/connectors", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *connector) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	connectors := make([]tg.Connector, 0)
	for _, conn := range node.Config.Connectors.Connectors {
		if conn.ID != d.Id() {
			connectors = append(connectors, conn)
		}
	}

	config := tg.ConnectorsConfig{Connectors: connectors}

	err := tgc.Put(ctx, fmt.Sprintf("/node/%s/config/connectors", tf.NodeID), config)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func (r *connector) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	node := tg.Node{}
	if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, conn := range node.Config.Connectors.Connectors {
		if conn.ID == d.Id() {
			tf.UpdateFromTG(conn)
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
