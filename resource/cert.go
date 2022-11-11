package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Cert() *schema.Resource {
	return &schema.Resource{
		Description: "Manage a certificate stored in Trustgrid.",

		ReadContext:   certRead,
		UpdateContext: certUpdate,
		DeleteContext: certDelete,
		CreateContext: certCreate,

		Schema: map[string]*schema.Schema{
			"fqdn": {
				Description: "Certificate FQDN",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"body": {
				Description: "PEM encoded certificate body",
				Type:        schema.TypeString,
				Required:    true,
			},
			"chain": {
				Description: "PEM encoded certificate chain",
				Type:        schema.TypeString,
				Required:    true,
			},
			"private_key": {
				Description: "PEM encoded private key",
				Type:        schema.TypeString,
				Sensitive:   true,
				Required:    true,
			},
		},
	}
}

func certCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	cert := tg.Cert{}
	if err := hcl.MarshalResourceData(d, &cert); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Post(ctx, "/v2/certificates", &cert); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cert.FQDN)

	return diag.Diagnostics{}
}

func certUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	cert := tg.Cert{}
	if err := hcl.MarshalResourceData(d, &cert); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, "/v2/certificates/"+cert.FQDN, &cert); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func certDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	cert := tg.Cert{}
	if err := hcl.MarshalResourceData(d, &cert); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, "/v2/certificates/"+cert.FQDN, &cert); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{}
}

func certRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	certs := make([]tg.Cert, 0)

	if err := tgc.Get(ctx, "/v2/certificates", &certs); err != nil {
		return diag.FromErr(err)
	}

	cert := tg.Cert{}

	for _, c := range certs {
		if c.FQDN == d.Get("fqdn").(string) {
			cert = c
			break
		}
	}

	if cert.FQDN == "" {
		return diag.Errorf("certificate not found")
	}

	d.SetId(cert.FQDN)

	return diag.Diagnostics{}
}
