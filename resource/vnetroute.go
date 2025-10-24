package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type vnetRoute struct {
}

func VNetRoute() *schema.Resource {
	r := vnetRoute{}

	return &schema.Resource{
		Description: "Manage a virtual network route",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "Unique identifier of the route",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"network": {
				Description: "Virtual network name - use the tg_virtual_network resource's exported name to help Terraform build a consistent dependency graph",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"dest": {
				Description: "Destination Node or Cluster name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network_cidr": {
				Description:  "Network CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"metric": {
				Description: "Metric",
				Type:        schema.TypeInt,
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

func (vn *vnetRoute) findRoute(ctx context.Context, tgc *tg.Client, route tg.VNetRoute) (tg.VNetRoute, error) {
	routes := []tg.VNetRoute{}
	err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route", &routes)
	if err != nil {
		return tg.VNetRoute{}, err
	}

	for _, r := range routes {
		if r.UID == route.UID {
			return r, nil
		}
		if route.UID == "" &&
			r.Dest == route.Dest &&
			r.NetworkCIDR == route.NetworkCIDR &&
			r.Metric == route.Metric &&
			r.Description == route.Description {
			return r, nil
		}
	}

	return tg.VNetRoute{}, &tg.NotFoundError{URL: "route " + route.UID}
}

func (vn *vnetRoute) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route", &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	route, err = vn.findRoute(ctx, tgc, route)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(route.UID)
	if err := d.Set("uid", route.UID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if _, err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route/"+route.UID, &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	route, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgc.Lock.Lock()
	defer tgc.Lock.Unlock()

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+route.NetworkName+"/route/"+route.UID, &route); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, route.NetworkName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *vnetRoute) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.VNetRoute](d)
	if err != nil {
		return diag.FromErr(err)
	}

	route, err := vn.findRoute(ctx, tgc, tf)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	route.NetworkName = tf.NetworkName
	if err := hcl.EncodeResourceData(route, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
