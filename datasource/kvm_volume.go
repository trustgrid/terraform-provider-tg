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

type kvmVolume struct {
}

func KVMVolume() *schema.Resource {
	r := kvmVolume{}

	return &schema.Resource{
		Description: "Fetch a KVM volume.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
				Required:     true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"size": {
				Description: "Volume size, in bytes",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"path": {
				Description: "Path to the volume",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"encrypted": {
				Description: "Whether the volume is encrypted",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"provision_type": {
				Description: "Provision type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"device_type": {
				Description: "Device type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"device_bus": {
				Description: "Device bus",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (r *kvmVolume) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.KVMVolume](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgimg := tg.KVMVolume{}
	err = tgc.Get(ctx, tf.ResourceURL(), &tgimg)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(fmt.Errorf("error with url %s: %w", tf.ResourceURL(), err))
	}

	tf.UpdateFromTG(tgimg)
	d.SetId(tf.Name)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
