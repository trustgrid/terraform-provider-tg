package datasource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type serviceUserDS struct {
}

func ServiceUser() *schema.Resource {
	r := serviceUserDS{}

	return &schema.Resource{
		Description: "Fetch a service user by ID or Name.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Service user name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"status": {
				Description: "Service user status",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"policy_ids": {
				Description: "Attached policies",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func (r *serviceUserDS) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	name, ok := d.Get("name").(string)
	if !ok || name == "" {
		return diag.FromErr(errors.New("name must be provided"))
	}

	var tgServiceUser tg.ServiceUser
	err := tgc.Get(ctx, "/v2/service-user/"+name, &tgServiceUser)
	if err != nil {
		return diag.FromErr(err)
	}

	tf := hcl.ServiceUser{}.UpdateFromTG(tgServiceUser)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return nil
}
