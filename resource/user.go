package resource

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"email": {
				Description: "User email address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"policy_ids": {
				Description: "List of policy IDs assigned to the user",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"last_name": {
				Description: "User's last name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"phone": {
				Description: "User's phone number",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"admin": {
				Description: "Whether the user is an admin",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"active": {
				Description: "Whether the user is active",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		},
	}
}
