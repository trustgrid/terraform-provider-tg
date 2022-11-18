package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type virtualNetwork struct {
}

func VirtualNetwork() *schema.Resource {
	r := virtualNetwork{}

	return &schema.Resource{
		Description: "Manage a domain virtual network.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"network_cidr": {
				Description:  "Network CIDR",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"no_nat": {
				Description: "Run the virtual network in NONAT mode",
				Type:        schema.TypeBool,
				Optional:    true,
			},
		},
	}
}

func vnetCommit(ctx context.Context, tgc *tg.Client, network string) error {
	var reply struct {
		Digest string `json:"digest"`
	}

	if err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network/"+network+"/change/validate", &reply); err != nil {
		return fmt.Errorf("error validating network changes: %w", err)
	}

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network/"+network+"/change/commit", &reply); err != nil {
		return fmt.Errorf("error committing network changes: %w", err)
	}

	return nil
}

func (vn *virtualNetwork) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	vnet := tg.VirtualNetwork{}
	if err := hcl.DecodeResourceData(d, &vnet); err != nil {
		return diag.FromErr(err)
	}

	if _, err := tgc.Post(ctx, "/v2/domain/"+tgc.Domain+"/network", &vnet); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vnet.Name)

	return nil
}

func (vn *virtualNetwork) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	vnet := tg.VirtualNetwork{}
	if err := hcl.DecodeResourceData(d, &vnet); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, "/v2/domain/"+tgc.Domain+"/network/"+vnet.Name, &vnet); err != nil {
		return diag.FromErr(err)
	}

	if err := vnetCommit(ctx, tgc, vnet.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *virtualNetwork) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	vnet := tg.VirtualNetwork{}
	if err := hcl.DecodeResourceData(d, &vnet); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, "/v2/domain/"+tgc.Domain+"/network/"+vnet.Name, &vnet); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (vn *virtualNetwork) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	vnet := tg.VirtualNetwork{}
	if err := hcl.DecodeResourceData(d, &vnet); err != nil {
		return diag.FromErr(err)
	}

	vnets := make([]tg.VirtualNetwork, 0)

	err := tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network", &vnets)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, v := range vnets {
		if v.Name == vnet.Name {
			found = true
			vnet = v
			break
		}
	}

	if !found {
		d.SetId("")
	}

	return nil
}
