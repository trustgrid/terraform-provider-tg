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
			"node": {
				Description: "Node or cluster name providing the service",
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
			"rate_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Rate limit in mbps",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description",
			},
		},
	}
}

func (r *connector) getConfig(ctx context.Context, tgc *tg.Client, tf hcl.Connector) (tg.ConnectorsConfig, error) {
	if tf.NodeID != "" {
		node := tg.Node{}
		if err := tgc.Get(ctx, fmt.Sprintf("/node/%s", tf.NodeID), &node); err != nil {
			return tg.ConnectorsConfig{}, err
		}
		return node.Config.Connectors, nil
	}
	cluster := tg.Cluster{}
	if err := tgc.Get(ctx, fmt.Sprintf("/cluster/%s", tf.ClusterFQDN), &cluster); err != nil {
		return tg.ConnectorsConfig{}, err
	}
	return cluster.Config.Connectors, nil
}

func (r *connector) writeConfig(ctx context.Context, tgc *tg.Client, tf hcl.Connector, config tg.ConnectorsConfig) error {
	url := fmt.Sprintf("/node/%s/config/connectors", tf.NodeID)
	if tf.NodeID == "" {
		url = fmt.Sprintf("/cluster/%s/config/connectors", tf.ClusterFQDN)
	}

	return tgc.Put(ctx, url, config)
}

func (r *connector) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(uuid.NewString())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	config.Connectors = append(config.Connectors, payload)

	if err := r.writeConfig(ctx, tgc, tf, config); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *connector) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	payload := tf.ToTG(d.Id())

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	for i, conn := range config.Connectors {
		if conn.ID == d.Id() {
			config.Connectors[i] = payload
		}
	}

	if err := r.writeConfig(ctx, tgc, tf, config); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(payload.ID)

	return nil
}

func (r *connector) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	connectors := make([]tg.Connector, 0)
	for _, conn := range config.Connectors {
		if conn.ID != d.Id() {
			connectors = append(connectors, conn)
		}
	}

	updated := tg.ConnectorsConfig{Connectors: connectors}

	if err := r.writeConfig(ctx, tgc, tf, updated); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func (r *connector) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Connector{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	config, err := r.getConfig(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, conn := range config.Connectors {
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
