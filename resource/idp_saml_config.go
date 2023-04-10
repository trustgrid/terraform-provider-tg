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

type idpSAMLConfig struct{}

// IDPSAMLConfig returns a Terraform resource for managing SAML IDP configuration
func IDPSAMLConfig() *schema.Resource {
	r := idpSAMLConfig{}

	return &schema.Resource{
		Description: "Manage SAML IDP configuration",

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
			"login_url": {
				Description:  "Login URL",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"issuer": {
				Description:  "Issuer",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsURLWithHTTPS,
			},
			"cert": {
				Description: "IDP Certificate",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// Create updates a SAML IDP configuration. Because this is tied to an IDP, no resource gets created on the TG side.
func (r *idpSAMLConfig) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDPSAMLConfig{}
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
func (r *idpSAMLConfig) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return r.Create(ctx, d, meta)
}

// Delete is a noop for SAML config - delete the IDP instead
func (r *idpSAMLConfig) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

// Read reads the SAML IDP configuration from the TG API and updates the local state.
func (r *idpSAMLConfig) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDPSAMLConfig{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgidp := tg.IDPSAMLConfig{}
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
