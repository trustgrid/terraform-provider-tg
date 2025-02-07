package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type virtualNetwork struct {
}

func VirtualNetwork() *schema.Resource {
	r := virtualNetwork{}

	return &schema.Resource{
		Description: "Fetch a domain virtual network.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network_cidr": {
				Description: "Network CIDR",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"no_nat": {
				Description: "Run the virtual network in NONAT mode",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func (vn *virtualNetwork) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	vnet, err := hcl.DecodeResourceData[tg.VirtualNetwork](d)
	if err != nil {
		return diag.FromErr(err)
	}

	vnets := make([]tg.VirtualNetwork, 0)

	err = tgc.Get(ctx, "/v2/domain/"+tgc.Domain+"/network", &vnets)
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
	} else {
		d.SetId(fmt.Sprintf("%d", vnet.ID))
	}

	return nil
}
