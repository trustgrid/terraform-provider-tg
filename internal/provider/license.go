package provider

import (
	"context"
	"io"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type License struct {
	Name    string `tf:"name" json:"-"`
	License string `tf:"license" json:"-"`
}

func licenseResource() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a TG node license. The license will be stored in TF state.",

		ReadContext:   licenseNoop,
		DeleteContext: licenseNoop,
		CreateContext: licenseCreate,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Node Name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"license": {
				Type:        schema.TypeString,
				Description: "License JWT",
				Computed:    true,
			},
		},
	}
}

func licenseNoop(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Diagnostics{}
}

func licenseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tg := meta.(*tgClient)
	l := License{}
	err := marshalResourceData(d, &l)
	if err != nil {
		return diag.FromErr(err)
	}

	if l.License == "" {
		reply, err := tg.rawGet(ctx, "/node/license?name="+l.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		defer reply.Close()
		body, err := io.ReadAll(reply)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(l.Name)
		d.Set("license", string(body))
	}

	return diag.Diagnostics{}
}
