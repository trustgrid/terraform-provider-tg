package datasource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type userDS struct {
}

// User returns the TF schema for a user data source
func User() *schema.Resource {
	r := userDS{}

	return &schema.Resource{
		Description: "Fetch a user by ID or Email.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description:  "User ID",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"uid", "email"},
			},
			"email": {
				Description:  "User email address",
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"uid", "email"},
			},
			"first_name": {
				Description: "User's first name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"last_name": {
				Description: "User's last name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"phone": {
				Description: "User's phone number",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"admin": {
				Description: "Whether the user is an admin",
				Type:        schema.TypeBool,
				Computed:    true,
			},
			"active": {
				Description: "Whether the user is active",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

// Read will look up a user by the UID or email provided. Errors if not found.
func (r *userDS) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	uid, hasUID := d.Get("uid").(string)
	email, hasEmail := d.Get("email").(string)

	var tgUser tg.User
	var err error

	if hasUID && uid != "" {
		// Look up by UID
		err = tgc.Get(ctx, "/v2/user/"+uid, &tgUser)
		if err != nil {
			return diag.FromErr(err)
		}
	} else if hasEmail && email != "" {
		// Look up by email - need to list all users and filter
		users := make([]tg.User, 0)
		err = tgc.Get(ctx, "/v2/user", &users)
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
	} else {
		return diag.FromErr(errors.New("either uid or email must be provided"))
	}

	tf := hcl.User{}.UpdateFromTG(tgUser)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	// Use UID as the ID
	d.SetId(tgUser.UID)

	return nil
}
