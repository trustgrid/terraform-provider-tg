package datasource

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type usersDS struct{}

// Users returns the TF schema for listing users
func Users() *schema.Resource {
	return &schema.Resource{
		Description: "Fetches users from Trustgrid",

		ReadContext: usersRead,

		Schema: map[string]*schema.Schema{
			"email_filter": {
				Description: "Filter users by email (substring match)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"admin_filter": {
				Description: "Filter users by admin status (true, false, or omit for all)",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"active_filter": {
				Description: "Filter users by active status (true, false, or omit for all)",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"users": {
				Type:        schema.TypeList,
				Description: "List of users",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"uid": {
							Type:        schema.TypeString,
							Description: "User ID",
							Computed:    true,
						},
						"email": {
							Type:        schema.TypeString,
							Description: "User email",
							Computed:    true,
						},
						"first_name": {
							Type:        schema.TypeString,
							Description: "User's first name",
							Computed:    true,
						},
						"last_name": {
							Type:        schema.TypeString,
							Description: "User's last name",
							Computed:    true,
						},
						"phone": {
							Type:        schema.TypeString,
							Description: "User's phone number",
							Computed:    true,
						},
						"admin": {
							Type:        schema.TypeBool,
							Description: "Whether the user is an admin",
							Computed:    true,
						},
						"active": {
							Type:        schema.TypeBool,
							Description: "Whether the user is active",
							Computed:    true,
						},
					},
				},
			},
			"emails": {
				Type:        schema.TypeSet,
				Description: "List of matching user emails",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type userFilter struct {
	EmailFilter  string `tf:"email_filter"`
	AdminFilter  *bool  `tf:"admin_filter"`
	ActiveFilter *bool  `tf:"active_filter"`
}

func (f *userFilter) match(u tg.User) bool {
	if f.EmailFilter != "" {
		// Simple substring match
		if !strings.Contains(u.Email, f.EmailFilter) {
			return false
		}
	}

	if f.AdminFilter != nil && u.Admin != *f.AdminFilter {
		return false
	}

	if f.ActiveFilter != nil && u.Active != *f.ActiveFilter {
		return false
	}

	return true
}

func usersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	d.SetId(fmt.Sprintf("%d", time.Now().Unix()))

	tgc := tg.GetClient(meta)

	// Get filter from user input
	filter := userFilter{}
	if emailFilter, ok := d.Get("email_filter").(string); ok && emailFilter != "" {
		filter.EmailFilter = emailFilter
	}

	// Check if admin_filter was set
	if adminFilterRaw, ok := d.GetOk("admin_filter"); ok {
		if adminFilter, ok := adminFilterRaw.(bool); ok {
			filter.AdminFilter = &adminFilter
		}
	}

	// Check if active_filter was set
	if activeFilterRaw, ok := d.GetOk("active_filter"); ok {
		if activeFilter, ok := activeFilterRaw.(bool); ok {
			filter.ActiveFilter = &activeFilter
		}
	}

	users := make([]tg.User, 0)
	err := tgc.Get(ctx, "/v2/user", &users)
	if err != nil {
		return diag.FromErr(err)
	}

	emails := make([]string, 0)
	userList := make([]map[string]interface{}, 0)

	for _, user := range users {
		if filter.match(user) {
			emails = append(emails, user.Email)
			userList = append(userList, map[string]interface{}{
				"uid":        user.UID,
				"email":      user.Email,
				"first_name": user.FirstName,
				"last_name":  user.LastName,
				"phone":      user.Phone,
				"admin":      user.Admin,
				"active":     user.Active,
			})
		}
	}

	err = d.Set("emails", emails)
	if err != nil {
		return diag.FromErr(err)
	}

	err = d.Set("users", userList)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
