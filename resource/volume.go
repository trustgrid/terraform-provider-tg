package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type volume struct {
}

func Volume() *schema.Resource {
	r := volume{}

	return &schema.Resource{
		Description: "Manage a volume.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

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
			"name": {
				Description: "Volume name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"encrypted": {
				Description: "Encrypt the volume",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func (vr *volume) urlRoot(v tg.Volume) string {
	if v.NodeID != "" {
		return "/v2/node/" + v.NodeID + "/exec/volume"
	}
	return "/v2/cluster/" + v.ClusterFQDN + "/exec/volume"
}

func (vr *volume) volumeURL(v tg.Volume) string {
	return vr.urlRoot(v) + "/" + v.Name
}

func (vr *volume) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	v := tg.Volume{}
	if err := hcl.DecodeResourceData(d, &v); err != nil {
		return diag.FromErr(err)
	}

	if _, err := tgc.Post(ctx, vr.urlRoot(v), &v); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(v.Name)

	return nil
}

func (vr *volume) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	v := tg.Volume{}
	if err := hcl.DecodeResourceData(d, &v); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, vr.volumeURL(v), &v); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vr *volume) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	v := tg.Volume{}
	if err := hcl.DecodeResourceData(d, &v); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, vr.volumeURL(v), &v); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vr *volume) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	v := tg.Volume{}
	if err := hcl.DecodeResourceData(d, &v); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Get(ctx, vr.volumeURL(v), &v); err != nil {
		var nferr *tg.NotFoundError
		if errors.As(err, &nferr) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	return nil
}
