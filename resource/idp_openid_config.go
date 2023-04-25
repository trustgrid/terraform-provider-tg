package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type idpOpenIDConfig struct{}

// IDPOpenIDConfig returns a schema.Resource for OpenID IDP configuration
func IDPOpenIDConfig() *schema.Resource {
	r := idpOpenIDConfig{}

	return &schema.Resource{
		Description: "Manage OpenID IDP configuration",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"idp_id": {
				Description:  "IDP ID",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"issuer": {
				Description:  "Issuer",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"client_id": {
				Description: "Client ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"secret": {
				Description: "Secret",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
			"auth_endpoint": {
				Description:  "Authorization endpoint URL",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"token_endpoint": {
				Description:  "Token endpoint URL",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"user_info_endpoint": {
				Description:  "User info endpoint URL",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
		},
	}
}

// Create updates a OpenID IDP configuration. Because this is tied to an IDP, no resource gets created on the TG side.
func (r *idpOpenIDConfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDPOpenIDConfig{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tgidp := tf.ToTG()

	if err := tgc.Put(ctx, tf.ResourceURL(tf.UID), &tgidp); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.UID)

	return nil
}

// Update updates a SAML IDP configuration.
func (r *idpOpenIDConfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return r.Create(ctx, d, meta)
}

// Delete is a noop for OpenID config - delete the IDP instead
func (r *idpOpenIDConfig) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

// Read reads the OpenID IDP configuration from the TG API and updates the local state.
func (r *idpOpenIDConfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDPOpenIDConfig{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgidp := tg.IDPOpenIDConfig{}
	err := tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgidp)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgidp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
