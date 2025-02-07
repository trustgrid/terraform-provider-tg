package datasource

import (
	"context"
	"errors"
	"fmt"

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

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
			},
			"uid": {
				Description: "Image ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"display_name": {
				Description: "Display Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"os": {
				Description: "OS",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"location": {
				Description: "Location",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (r *kvmImage) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMImage](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgimg := tg.KVMImage{}
	err = tgc.Get(ctx, tf.ResourceURL(tf.UID), &tgimg)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(fmt.Errorf("error with url %s: %w", tf.ResourceURL(tf.UID), err))
	}

	tf.UpdateFromTG(tgimg)

	d.SetId(tf.UID)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
