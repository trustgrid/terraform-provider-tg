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

func Alarm() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.Alarm, hcl.Alarm]{
			CreateURL: func(_ hcl.Alarm) string { return "/v2/alarm" },
			OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.Alarm, hcl.Alarm]) (string, error) {
				var response struct {
					ID string `json:"uid"`
				}
				if err := json.Unmarshal(args.Body, &response); err != nil {
					return "", err
				}

				if err := args.TF.Set("uid", response.ID); err != nil {
					return "", err
				}
				return response.ID, nil
			},
			UpdateURL: func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
			DeleteURL: func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
			GetURL:    func(a hcl.Alarm) string { return "/v2/alarm/" + a.UID },
			ID: func(a hcl.Alarm) string {
				return a.UID
			},
		})

	return &schema.Resource{
		Description: "Manage an alarm.",

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
