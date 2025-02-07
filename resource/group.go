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

type group struct {
}

func Group() *schema.Resource {
	r := group{}

	return &schema.Resource{
		Description: "Manages a user group.",

		ReadContext:   r.Read,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"idp_id": {
				Description: "IDP ID - will be blank for local groups",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

func (r *group) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Group{}

	id, ok := d.Get("uid").(string)
	if !ok {
		return diag.FromErr(errors.New("uid must be a string"))
	}

	tgapp := tg.Group{}
	err := tgc.Get(ctx, tf.ResourceURL(id), &tgapp)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgapp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}

func (r *group) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Group](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tggrp := tf.ToTG()

	_, err = tgc.Post(ctx, tf.URL(), &tggrp)
	if err != nil {
		return diag.FromErr(err)
	}

	groups := make([]tg.Group, 0)

	if err := tgc.Get(ctx, tf.URL(), &groups); err != nil {
		return diag.FromErr(err)
	}

	for _, g := range groups {
		if g.ReferenceID == "local-"+g.Name {
			tf.UpdateFromTG(g)
			if err := hcl.EncodeResourceData(tf, d); err != nil {
				return diag.FromErr(err)
			}
			d.SetId(g.UID)
			return nil
		}
	}

	d.SetId("")
	return diag.FromErr(fmt.Errorf("group %s not found after creation", tf.Name))
}

func (r *group) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.Group](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(fmt.Errorf("error issuing delete to group API: %w", err))
	}

	return nil
}
