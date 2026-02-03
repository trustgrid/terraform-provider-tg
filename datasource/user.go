package datasource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type user struct {
}

func User() *schema.Resource {
	r := user{}

	return &schema.Resource{
		Description: "Fetch a user by email.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"email": {
				Description: "User email address",
				Type:        schema.TypeString,
				Required:    true,
			},
			"policy_ids": {
				Description: "List of policy IDs assigned to the user",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"status": {
				Description: "User status (active or inactive)",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func (r *user) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	email := d.Get("email").(string)

	var tgUser tg.User
	var err error

	users := make([]tg.User, 0)
	err = tgc.Get(ctx, "/user"+email, nil)
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, u := range users {
		if u.Email == email {
			tgUser = u
			found = true
			break
		}
	}

	if !found {
		return diag.FromErr(errors.New("user with email " + email + " not found"))
	}

	tf := hcl.User{}.UpdateFromTG(tgUser)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tgUser.Email)

	return nil
}
