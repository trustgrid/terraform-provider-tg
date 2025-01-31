package resource

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
		Description: "Manage a KVM volume.",

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
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"size": {
				Description: "Volume size, in bytes",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"path": {
				Description: "Path to the volume",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"encrypted": {
				Description: "Whether the volume is encrypted",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"provision_type": {
				Description:  "Provision type",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"thin", "lazy", "eager"}, false),
			},
			"device_type": {
				Description:  "Device type",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"disk", "cdrom"}, false),
			},
			"device_bus": {
				Description:  "Device bus",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ide", "scsi", "virtio", "sata"}, false),
			},
		},
	}
}

func (r *kvmVolume) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.KVMVolume{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgimg := tf.ToTG()

	_, err := tgc.Post(ctx, tf.URL(), &tgimg)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.Name)

	return nil
}

func (r *kvmVolume) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.KVMVolume{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgimg := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(), &tgimg); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *kvmVolume) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.KVMVolume{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *kvmVolume) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.KVMVolume{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgimg := tg.KVMVolume{}
	err := tgc.Get(ctx, tf.ResourceURL(), &tgimg)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(fmt.Errorf("error with url %s: %w", tf.ResourceURL(), err))
	}

	tf.UpdateFromTG(tgimg)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
