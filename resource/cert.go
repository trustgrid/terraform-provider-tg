package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func Cert() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.Cert, hcl.Cert]{
			CreateURL: func(_ hcl.Cert) string { return "/v2/certificates" },
			UpdateURL: func(cert hcl.Cert) string { return "/v2/certificates/" + cert.FQDN },
			DeleteURL: func(cert hcl.Cert) string { return "/v2/certificates/" + cert.FQDN },
			IndexURL:  func() string { return "/v2/certificates" },
			ID: func(cert hcl.Cert) string {
				return cert.FQDN
			},
			RemoteID: func(cert tg.Cert) string {
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
