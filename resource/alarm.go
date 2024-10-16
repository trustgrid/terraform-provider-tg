package resource

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type alarm struct {
}

func Alarm() *schema.Resource {
	r := alarm{}

	return &schema.Resource{
		Description: "Manage an alarm.",

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
				Description: "Name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "When true, this alarm can generate alerts",
				Type:        schema.TypeBool,
				Optional:    true,
			},
			"channels": {
				Description: "Channel IDs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"nodes": {
				Description: "Node names",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"freetext": {
				Description: "Free text match",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"expr": {
				Description: "CEL expression",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"operator": {
				Description:  "Criteria operator",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"all", "any", "none"}, false),
			},
			"threshold": {
				Description:  "Severity threshold",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"INFO", "WARNING", "ERROR", "CRITICAL"}, false),
			},
			"types": {
				Description: "Event types",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"All Gateways Disconnected",
						"All Peers Disconnected",
						"Certificate Expiring",
						"Cluster Unhealthy",
						"Cluster Failover",
						"Cluster Healthy",
						"Configuration Update Failure",
						"Connection Flapping",
						"Connection Timeout",
						"Data Plane Disruption",
						"DNS Resolution",
						"Gateway Connectivity Health Check",
						"Gateway Ingress Limit Reached",
						"Gateway Latency",
						"Invalid Configuration",
						"Metric Threshold Violation",
						"Network Error",
						"Network Route Error",
						"Node Connect",
						"Node Delete",
						"Node Device Reboot",
						"Node Disconnect",
						"Node Restart",
						"Order Created",
						"Order Customer Update",
						"Order Status Change",
						"Order Comment",
						"Repo Connectivity",
						"Site Outage",
						"Site Recovery",
						"SSH Lockdown",
						"Threat Activity",
						"Unauthorized IP",
					}, false),
				},
			},
			"tag_operator": {
				Description:  "Tag match operator",
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"any", "all"}, false),
			},
			"tag": {
				Description: "Tag pairs",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "Tag name",
							Type:        schema.TypeString,
							Required:    true,
						},
						"value": {
							Description: "Tag value",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

func (r *alarm) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Alarm{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
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

func (r *alarm) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Alarm{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgch := tf.ToTG()
	if err := tgc.Put(ctx, tf.ResourceURL(d.Id()), &tgch); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *alarm) Delete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Alarm{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	if err := tgc.Delete(ctx, tf.ResourceURL(d.Id()), nil); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (r *alarm) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf := hcl.Alarm{}
	if err := hcl.DecodeResourceData(d, &tf); err != nil {
		return diag.FromErr(err)
	}

	tgch := tg.Alarm{}
	err := tgc.Get(ctx, tf.ResourceURL(d.Id()), &tgch)
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
