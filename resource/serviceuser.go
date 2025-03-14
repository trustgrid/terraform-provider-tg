package resource

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func ServiceUser() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.ServiceUser, hcl.ServiceUser]{
			CreateURL: func(_ hcl.ServiceUser) string { return "/v2/service-user" },
			UpdateURL: func(user hcl.ServiceUser) string { return "/v2/service-user/" + user.Name },
			DeleteURL: func(user hcl.ServiceUser) string { return "/v2/service-user/" + user.Name },
			GetURL:    func(user hcl.ServiceUser) string { return "/v2/service-user/" + user.Name },
			AfterCreate: func(ctx context.Context, d *schema.ResourceData, meta any) error {
				tgc := tg.GetClient(meta)
				reply, err := tgc.Post(ctx, "/v2/service-user/"+d.Id()+"/token", nil)
				if err != nil {
					return err
				}
				var token tg.APIToken
				if err := json.Unmarshal(reply, &token); err != nil {
					return err
				}

				if err := d.Set("client_id", token.ClientID); err != nil {
					return err
				}
				if err := d.Set("secret", token.Secret); err != nil {
					return err
				}

				return nil
			},
			ID: func(user hcl.ServiceUser) string {
				return user.Name
			},
			RemoteID: func(user tg.ServiceUser) string {
				return user.Name
			},
		})

	return &schema.Resource{
		Description: "Manage a Trustgrid service user",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Service user name",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"status": {
				Description:  "Service user status",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
			},
			"policy_ids": {
				Description: "Attached policies",
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"client_id": {
				Description: "API client ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"secret": {
				Description: "API client secret",
				Type:        schema.TypeString,
				Sensitive:   true,
				Computed:    true,
			},
		},
	}
}
