package resource

import (
	"context"
	"encoding/json"
	"errors"

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
		Description: "Manage an alarm channel.",

		ReadContext:   r.Read,
		UpdateContext: r.Update,
		DeleteContext: r.Delete,
		CreateContext: r.Create,

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

func (r *alarmChannel) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AlarmChannel](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgch := tf.ToTG()

	reply, err := tgc.Post(ctx, tf.URL(), &tgch)
	if err != nil {
		return diag.FromErr(err)
	}
	var response struct {
		ID string `json:"uid"`
	}
	if err := json.Unmarshal(reply, &response); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("uid", response.ID); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(response.ID)

	return nil
}

func (r *alarmChannel) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AlarmChannel](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgch := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgch); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *alarmChannel) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AlarmChannel](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *alarmChannel) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.AlarmChannel](d)
	if err != nil {
		return diag.FromErr(err)
	}

	tgch := tg.AlarmChannel{}
	err = tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgch)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgch)
	tf.UID = d.Id()

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
