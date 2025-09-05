package resource

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/golang-jwt/jwt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
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
				Description:  "Node Name - lowercase letters and numbers only",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validators.StringIsLowercaseAndNumbers,
			},
			"uid": {
				Description: "Node UID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"fqdn": {
				Description: "Node FQDN",
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

func licenseNoop(ctx context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	tflog.Debug(ctx, "hi from licensenoop")
	return nil
}

func licenseCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	parser := jwt.Parser{ValidMethods: []string{"RS512"}}

	tgc := tg.GetClient(meta)
	tf, err := hcl.DecodeResourceData[licenseData](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if tf.License == "" {
		reply, err := tgc.RawGet(ctx, "/node/license?name="+tf.Name)
		var verr *tg.ValidationError
		switch {
		case errors.As(err, &verr):
			return diag.FromErr(fmt.Errorf("invalid license - usually this means the name is already taken"))
		case err != nil:
			return diag.FromErr(err)
		}

		defer reply.Close()
		body, err := io.ReadAll(reply)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(tf.Name)
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

		if err := d.Set("fqdn", tf.Name+"."+tgc.Domain); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
