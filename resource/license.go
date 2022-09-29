package resource

import (
	"context"
	"io"

	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type licenseData struct {
	Name    string `tf:"name" json:"-"`
	License string `tf:"license" json:"-"`
	UID     string `tf:"uid" json:"-"`
}

func License() *schema.Resource {
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
			"uid": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Computed:    true,
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
	tflog.Debug(ctx, "hi from licensenoop")
	return diag.Diagnostics{}
}

func licenseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	parser := jwt.Parser{ValidMethods: []string{"RS512"}}

	tg := meta.(*tg.Client)
	l := licenseData{}
	err := hcl.MarshalResourceData(d, &l)
	if err != nil {
		return diag.FromErr(err)
	}

	if l.License == "" {
		reply, err := tg.RawGet(ctx, "/node/license?name="+l.Name)
		if err != nil {
			return diag.FromErr(err)
		}

		defer reply.Close()
		body, err := io.ReadAll(reply)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(l.Name)
		if err := d.Set("license", string(body)); err != nil {
			return diag.FromErr(err)
		}

		var claims struct {
			jwt.StandardClaims
			Exp float64 `json:"exp"`
		}
		if _, _, err := parser.ParseUnverified(string(body), &claims); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("uid", claims.Id); err != nil {
			return diag.FromErr(err)
		}
	}

	return diag.Diagnostics{}
}
