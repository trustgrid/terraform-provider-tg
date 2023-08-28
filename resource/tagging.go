package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type tagging struct {
}

func Tagging() *schema.Resource {
	r := tagging{}

	return &schema.Resource{
		Description: "Node or Cluster tags",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node UID - required if cluster_fqdn not set",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - required if node_id not set",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"tags": {
				Description: "Include Tag Filters",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func (r *tagging) writeTags(ctx context.Context, tgc *tg.Client, tf hcl.Tagging) (string, error) {
	url := "/node/" + tf.NodeID + "/tags"
	id := tf.NodeID
	if tf.NodeID == "" {
		url = "/cluster/" + tf.ClusterFQDN + "/tags"
		id = tf.ClusterFQDN
	}

	if tf.Tags == nil {
		tf.Tags = make(map[string]string)
	}

	if err := tgc.Put(ctx, url, map[string]any{"tags": tf.Tags}); err != nil {
		return "", err
	}

	return id, nil
}

func (r *tagging) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Tagging{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	id, err := r.writeTags(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)
	return nil
}

func (r *tagging) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Tagging{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	id, err := r.writeTags(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}

func (r *tagging) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Tagging{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tf.Tags = nil
	_, err := r.writeTags(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func (r *tagging) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Tagging{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if tf.NodeID != "" {
		url := "/node/" + tf.NodeID
		tgnode := tg.Node{}
		err := tgc.Get(ctx, url, &tgnode)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return diag.FromErr(fmt.Errorf("url %s not found", url))
		case err != nil:
			return diag.FromErr(err)
		}

		tf.Tags = tgnode.Tags
	} else {
		url := "/cluster/" + tf.ClusterFQDN
		tgcluster := tg.Cluster{}
		err := tgc.Get(ctx, url, &tgcluster)
		var nferr *tg.NotFoundError
		switch {
		case errors.As(err, &nferr):
			d.SetId("")
			return nil
		case err != nil:
			return diag.FromErr(err)
		}

		tf.Tags = tgcluster.Tags
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
