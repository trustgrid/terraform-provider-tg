package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetGroup struct {
}

func VNetGroup() *schema.Resource {
	r := vnetGroup{}

	return &schema.Resource{
		Description: "Manage a virtual network group. See [Network Groups](https://docs.trustgrid.io/docs/domain/virtual-networks/network-groups/) for more information.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Group name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func (vn *vnetGroup) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	group, err := hcl.DecodeResourceData[hcl.VNetGroup](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+group.NetworkName+"/network-group", group.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, group.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(group.Name)

	return nil
}

func (vn *vnetGroup) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	group, err := hcl.DecodeResourceData[hcl.VNetGroup](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+group.NetworkName+"/network-group/"+group.Name, group.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, group.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetGroup) findGroup(ctx context.Context, tgc *tg.Client, group hcl.VNetGroup) (tg.VNetGroup, error) {
	groups := []tg.VNetGroup{}
	err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+group.NetworkName+"/network-group", &groups)
	if err != nil {
		return tg.VNetGroup{}, err
	}

	for _, r := range groups {
		if r.Name == group.Name {
			return r, nil
		}
	}

	return tg.VNetGroup{}, &tg.NotFoundError{URL: "network-group " + group.Name}
}

func (vn *vnetGroup) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	group, err := hcl.DecodeResourceData[hcl.VNetGroup](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+group.NetworkName+"/network-group/"+group.Name, group.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, group.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetGroup) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VNetGroup](d)
	if err != nil {
		return diag.FromErr(err)
	}

	group, err := vn.findGroup(ctx, tgc, tf)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(tf.UpdateFromTG(group), d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
