package resource

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type appAccessRule struct {
}

func ruleItemSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"everyone": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "If true, this rule always matches",
		},
		"emails": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of emails",
		},
		"ip_ranges": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsCIDR,
			},
			Description: "List of IP ranges",
		},
		"countries": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of countries",
		},
		"emails_ending_in": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of email suffixes",
		},
		"idp_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of IDP group IDs",
		},
		"access_groups": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "List of access group IDs",
		},
	}
}

func AppAccessRule() *schema.Resource {
	r := appAccessRule{}

	return &schema.Resource{
		Description: "Manage a ZTNA application access rule.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

		Schema: map[string]*schema.Schema{
			"app": {
				Description: "App ID",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Description: "App Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"action": {
				Description:  "Rule action",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "block"}, false),
			},
			"include": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				MinItems:    1,
				Description: "Includes",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
			"require": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Requires",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
			"exception": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Exceptions",
				Elem: &schema.Resource{
					Schema: ruleItemSchema(),
				},
			},
		},
	}
}

func (r *appAccessRule) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.AccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgrule := tf.ToTG()

	reply, err := tgc.Post(ctx, tf.URL(), &tgrule)
	if err != nil {
		return diag.FromErr(err)
	}
	var response struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(reply, &response); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(response.ID)

	return nil
}

func (r *appAccessRule) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.AccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgrule := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgrule); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appAccessRule) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.AccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *appAccessRule) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.AccessRule{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgrule := tg.AppAccessRule{}
	err := tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgrule)
	switch {
	case errors.Is(err, tg.ErrNotFound):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgrule)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
