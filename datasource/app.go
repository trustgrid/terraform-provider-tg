package datasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type app struct {
}

func App() *schema.Resource {
	r := app{}

	return &schema.Resource{
		Description: "Fetch a ZTNA application.",

		ReadContext: r.Read,

		Schema: map[string]*schema.Schema{
			"uid": {
				Description: "ID",
				Type:        schema.TypeString,
				Required:    true,
			},
			"type": {
				Description: "App Type",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"edge_node": {
				Description: "Edge node ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"gateway_node": {
				Description: "Gateway node ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"idp": {
				Description: "IDP ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"ip": {
				Description: "IP",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"port": {
				Description: "Port",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"protocol": {
				Description: "Protocol",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"hostname": {
				Description: "Hostname",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"session_duration": {
				Description: "Session Duration",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"tls_verification_mode": {
				Description: "TLS Verification Mode",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"trust_mode": {
				Description: "Trust Mode",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"vrf": {
				Description: "VRF",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"virtual_network": {
				Description: "Virtual network name",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"virtual_source_ip": {
				Description: "Virtual source IP, if using a virtual network",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"wireguard_template": {
				Description: "WireGuard Template",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"visibility_groups": {
				Description: "Visibility Groups",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func (r *app) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := meta.(*tg.Client)

	tf := hcl.App{}

	id := d.Get("uid").(string)

	tgapp := tg.App{}
	err := tgc.Get(ctx, tf.ResourceURL(id), &tgapp)
	if err != nil {
		return diag.FromErr(err)
	}

	tf.UpdateFromTG(tgapp)

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(id)

	return nil
}
