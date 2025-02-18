package datasource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Cert() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches cert info from Trustgrid - will error if it doesn't exist",

		ReadContext: certRead,

		Schema: map[string]*schema.Schema{
			"fqdn": {
				Description: "Certificate FQDN",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func certRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	certs := make([]tg.Cert, 0)

	if err := tgc.Get(ctx, "/v2/certificates", &certs); err != nil {
		return diag.FromErr(err)
	}

	cert := tg.Cert{}

	fqdn, ok := d.Get("fqdn").(string)
	if !ok {
		return diag.FromErr(errors.New("fqdn must be a string"))
	}

	for _, c := range certs {
		if c.FQDN == fqdn {
			cert = c
			break
		}
	}

	if cert.FQDN == "" {
		d.SetId("")
		return nil
	}

	if err := hcl.EncodeResourceData(&cert, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
