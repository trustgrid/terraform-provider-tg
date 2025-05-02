package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetObject struct {
}

func VNetObject() *schema.Resource {
	r := vnetObject{}

	return &schema.Resource{
		Description: "Manage a virtual network object. See [Network Objects](https://docs.trustgrid.io/docs/domain/virtual-networks/network-objects/) for more information.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Object name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"cidr": {
				Description: "Object CIDR",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func (vn *vnetObject) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	obj, err := hcl.DecodeResourceData[hcl.VNetObject](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+obj.NetworkName+"/network-object", obj.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, obj.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(obj.Name)

	return nil
}

func (vn *vnetObject) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	obj, err := hcl.DecodeResourceData[hcl.VNetObject](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+obj.NetworkName+"/network-object/"+obj.Name, obj.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, obj.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetObject) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	obj, err := hcl.DecodeResourceData[hcl.VNetObject](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+obj.NetworkName+"/network-object/"+obj.Name, obj.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, obj.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetObject) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VNetObject](d)
	if err != nil {
		return diag.FromErr(err)
	}

	var obj tg.VNetObject
	if err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+tf.NetworkName+"/network-object/"+tf.Name, &obj); err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(tf.UpdateFromTG(obj), d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
