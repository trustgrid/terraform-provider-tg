package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type portalAuth struct{}

// PortalAuth returns a Terraform resource for managing Portal authentication.
func PortalAuth() *schema.Resource {
	r := portalAuth{}

	return &schema.Resource{
		Description: "Manage Portal authentication.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		CreateContext: r.Create,
		DeleteContext: r.Delete,

		Schema: map[string]*schema.Schema{
			"idp_id": {
				Description: "Either your IDP uid or `trustgrid`",
				Type:        schema.TypeString,
				Required:    true,
			},
			"domain": {
				Description: "The domain name users should connect to when accessing the Portal, like mycompany.trustgrid.io",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

// Create configures Portal authentication - it's the same as update, since this can't be deleted.
func (r *portalAuth) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.PortalAuth{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tgauth := tf.ToTG()

	err := tgc.Put(ctx, "/org/auth", &tgauth)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tf.Domain)

	return nil
}

// Update configures Portal authentication
func (r *portalAuth) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.PortalAuth{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgauth := tf.ToTG()
	if err := tgc.Put(ctx, "/org/auth", &tgauth); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Delete sets the Portal authentication provider back to the default Trustgrid provider.
func (r *portalAuth) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.PortalAuth{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tf.IDPID = "trustgrid"

	tgauth := tf.ToTG()
	if err := tgc.Put(ctx, "/org/auth", &tgauth); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Read pulls the authentication provider configuration from the TG API and updates the Terraform state.
func (r *portalAuth) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.PortalAuth{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgauth := tg.PortalAuth{}
	err := tgc.Get(ctx, "/org/auth", &tgauth)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgauth)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
