package resource

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func AlarmChannel() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.AlarmChannel, hcl.AlarmChannel]{
			CreateURL: func(_ hcl.AlarmChannel) string { return "/v2/alarm-channel" },
			OnCreateReply: func(d *schema.ResourceData, reply []byte) (string, error) {
				var response struct {
					ID string `json:"uid"`
				}
				if err := json.Unmarshal(reply, &response); err != nil {
					return "", err
				}

				if err := d.Set("uid", response.ID); err != nil {
					return "", err
				}
				return response.ID, nil
			},
			UpdateURL: func(a hcl.AlarmChannel) string { return "/v2/alarm-channel/" + a.UID },
			DeleteURL: func(a hcl.AlarmChannel) string { return "/v2/alarm-channel/" + a.UID },
			GetURL:    func(a hcl.AlarmChannel) string { return "/v2/alarm-channel/" + a.UID },
			ID: func(a hcl.AlarmChannel) string {
				return a.UID
			},
		})
	return &schema.Resource{
		Description: "Manage an alarm channel.",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "ID",
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Description: "Channel name",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"emails": {
				Description: "Emails",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generic_webhook": {
				Description: "Generic webhook",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ms_teams": {
				Description: "Microsoft Teams webhook",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"ops_genie": {
				Description: "OpsGenie key",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"pagerduty": {
				Description: "Pagerduty key",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"slack": {
				Description: "Slack configuration",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Description: "Slack channel",
							Type:        schema.TypeString,
							Required:    true,
						},
						"webhook": {
							Description: "Slack webhook",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}
