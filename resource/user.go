package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func User() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.User, hcl.User]{
			CreateURL: func(_ hcl.User) string { return "/user/add" },
			UpdateURL: func(u hcl.User) string { return "/user/" + u.Email },
			DeleteURL: func(u hcl.User) string { return "/user/" + u.Email },
			GetURL:    func(u hcl.User) string { return "/user/" + u.Email },
			ID: func(user hcl.User) string {
				return user.Email
			},
			RemoteID: func(user tg.User) string {
				return user.Email
			},
		})

	return &schema.Resource{
		Description: "Manage a Trustgrid user",

		ReadContext: func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			oldIDP, ok := d.Get("idp").(string)
			if !ok {
				oldIDP = ""
			}

			diags := md.Read(ctx, d, m)
			if diags.HasError() {
				return diags
			}

			newIDP, ok := d.Get("idp").(string)
			if !ok {
				newIDP = ""
			}

			if newIDP == "" && oldIDP != "" {
				if err := d.Set("idp", oldIDP); err != nil {
					return diag.FromErr(err)
				}
			}
			return diags
		},
		UpdateContext: func(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
			tgc := tg.GetClient(m)
			user, err := hcl.DecodeResourceData[hcl.User](d)
			if err != nil {
				return diag.FromErr(err)
			}

			tgUser := user.ToTG()
			tgUser.IDP = "" // Clear IDP for update as it causes 500 error

			// Call PUT
			_, err = tgc.Put(ctx, "/user/"+user.Email, &tgUser)
			if err != nil {
				return diag.FromErr(err)
			}

			// Read back
			return md.Read(ctx, d, m)
		},
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "User unique identifier (UUID)",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "User email address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"idp": {
				Description: "Identity provider ID",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
			"policy_ids": {
				Description: "List of policy IDs assigned to the user",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Description:  "User status (active or inactive)",
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
			},
		},
	}
}
