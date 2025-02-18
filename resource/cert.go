package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
)

func Cert() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[hcl.Cert]{
			CreateURL: func(_ hcl.Cert) string { return "/v2/certificates" },
			UpdateURL: func(cert hcl.Cert) string { return "/v2/certificates/" + cert.FQDN },
			DeleteURL: func(cert hcl.Cert) string { return "/v2/certificates/" + cert.FQDN },
			IndexURL:  func(cert hcl.Cert) string { return "/v2/certificates" },
			ID: func(cert hcl.Cert) string {
				return cert.FQDN
			},
		})

	return &schema.Resource{
		Description: "Manage a certificate stored in Trustgrid.",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

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

/*
func certCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.Cert](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, err := tgc.Post(ctx, "/v2/certificates", &tf); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.FQDN)

	return nil
}
*/

/*
func certUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.Cert](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Put(ctx, "/v2/certificates/"+tf.FQDN, &tf); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
*/

/*
func certDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[tg.Cert](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, "/v2/certificates/"+tf.FQDN, &tf); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
*/

/*
func certRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

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
		d.SetId("")
	}

	return nil
}

*/
