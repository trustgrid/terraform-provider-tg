package resource

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type kvmImage struct {
}

func KVMImage() *schema.Resource {
	r := kvmImage{}

	return &schema.Resource{
		Description: "Manage a KVM image.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
				ForceNew:     true,
			},
			"uid": {
				Description: "Volume ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"display_name": {
				Description: "Display Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"os": {
				Description: "OS",
				Type:        schema.TypeString,
				Required:    true,
			},
			"location": {
				Description: "Location",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func (r *kvmImage) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMImage](d)
	if err != nil {
		return diag.FromErr(err)
	}

	id := uuid.NewString()
	tgimg := tf.ToTG()
	tgimg.ID = id

	_, err = tgc.Post(ctx, tf.URL(), &tgimg)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}

func (r *kvmImage) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMImage](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgimg := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgimg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *kvmImage) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMImage](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *kvmImage) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMImage](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgimg := tg.KVMImage{}
	err = tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgimg)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(fmt.Errorf("error with url %s: %w", tf.ResourceURL(d.Id()), err))
	}

	tf.UpdateFromTG(tgimg)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
