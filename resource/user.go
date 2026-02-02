package resource

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func User() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.User, hcl.User]{
			CreateURL: func(_ hcl.User) string { return "/v2/user" },
			UpdateURL: func(u hcl.User) string { return "/v2/user/" + u.UID },
			DeleteURL: func(u hcl.User) string { return "/v2/user/" + u.UID },
			GetURL:    func(u hcl.User) string { return "/v2/user/" + u.UID },
			ID: func(user hcl.User) string {
				return user.UID
			},
			RemoteID: func(user tg.User) string {
				return user.UID
			},
			OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.User, hcl.User]) (string, error) {
				// Parse the response to get the UID
				var user tg.User
				if err := json.Unmarshal(args.Body, &user); err != nil {
					return "", err
				}
				return user.UID, nil
			},
		})

	return &schema.Resource{
		Description: "Manage a Trustgrid user",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "User ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"email": {
				Description: "User email address",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"first_name": {
				Description: "User's first name",
				Type:        schema.TypeString,
				Required:    true,
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
