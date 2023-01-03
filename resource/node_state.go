package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type nodeState struct {
}

func NodeState() *schema.Resource {
	r := nodeState{}

	return &schema.Resource{
		Description: "Manage a Node state.",

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
			"enabled": {
				Description: "Enable the node",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func (r *nodeState) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Node{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgnode := tf.ToTG()

	err := tgc.Put(ctx, tf.ResourceURL(tf.UID), &tgnode)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)

	return nil
}

func (r *nodeState) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Node{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgnode := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgnode); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *nodeState) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}

func (r *nodeState) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.Node{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgnode := tg.NodeState{}
	err := tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgnode)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgnode)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
