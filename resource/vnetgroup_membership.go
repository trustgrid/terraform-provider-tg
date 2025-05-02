package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetGroupMembership struct {
}

func VNetGroupMembership() *schema.Resource {
	r := vnetGroupMembership{}

	return &schema.Resource{
		Description: "Manage a virtual network object group member. See [Network Groups](https://docs.trustgrid.io/docs/domain/virtual-networks/network-groups/) for more information.",

		ReadContext:   r.Read,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"object": {
				Description: "Object name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"group": {
				Description: "Group name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
		},
	}
}

func (vn *vnetGroupMembership) url(tgc *tg.Client, obj hcl.VNetGroupMembership) string {
	return "/v2/domain/" + tgc.Domain + "/network/" + obj.NetworkName + "/network-group/" + obj.Group + "/" + obj.Object
}

func (vn *vnetGroupMembership) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	obj, err := hcl.DecodeResourceData[hcl.VNetGroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, vn.url(tgc, obj), obj.ToTG()); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, obj.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(obj.Group + "-" + obj.Object)

	return nil
}

func (vn *vnetGroupMembership) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	obj, err := hcl.DecodeResourceData[hcl.VNetGroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, vn.url(tgc, obj), nil); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, obj.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetGroupMembership) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.VNetGroupMembership](d)
	if err != nil {
		return diag.FromErr(err)
	}

	var obj []tg.VNetGroupMembership
	if err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+tf.NetworkName+"/network-group/"+tf.Group, &obj); err != nil {
		return diag.FromErr(err)
	}

	var membership tg.VNetGroupMembership
	found := false
	for _, r := range obj {
		if r.Object == tf.Object {
			membership = r
			found = true
			break
		}
	}

	if !found {
		d.SetId("")
		return nil
	}

	if err := hcl.EncodeResourceData(tf.UpdateFromTG(membership), d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
