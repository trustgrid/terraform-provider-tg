package datasource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type alarmChannel struct {
}

func AlarmChannel() *schema.Resource {
	r := alarmChannel{}

	return &schema.Resource{
		Description: "Fetch an alarm channel.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"emails": {
				Description: "Emails",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generic_webhook": {
				Description: "Generic webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ms_teams": {
				Description: "Microsoft Teams webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Channel name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ops_genie": {
				Description: "OpsGenie key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"pagerduty": {
				Description: "Pagerduty key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"slack": {
				Description: "Slack configuration",
				Type:        schema.TypeList,
				Computed:    true,
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

func (r *alarmChannel) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.AlarmChannel{}

	id, ok := d.Get("uid").(string)
	if !ok {
		return diag.FromErr(fmt.Errorf("uid must be a string"))
	}

	tgchannel := tg.AlarmChannel{}
	err := tgc.Get(ctx, tf.ResourceURL(id), &tgchannel)
	if err != nil {
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgchannel)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
