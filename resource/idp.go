package resource

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type idp struct{}

// IDP returns a Terraform resource for managing IDPs.
func IDP() *schema.Resource {
	r := idp{}

	return &schema.Resource{
		Description: "Manage an IDP.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"type": {
				Description:  "Type - one of `GSuite`, `OpenID`, `SAML`, or `AzureAD`",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"GSuite", "OpenID", "SAML", "AzureAD"}, false),
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"uid": {
				Description: "UID",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

// Create creates a new IDP.
func (r *idp) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDP{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}
	tf.UID = uuid.NewString()
	tgidp := tf.ToTG()

	_, err := tgc.Post(ctx, tf.URL(), &tgidp)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("uid", tf.UID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(tf.UID)

	return nil
}

// Update updates an existing IDP.
func (r *idp) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDP{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgidp := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(), &tgidp); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Delete deletes an existing IDP.
func (r *idp) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDP{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Read pulls an IDP from the TG API and updates the Terraform state.
func (r *idp) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.IDP{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgidp := tg.IDP{}
	err := tgc.Get(ctx, tf.ResourceURL(), &tgidp)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
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
