package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/datasource"
	"github.com/trustgrid/terraform-provider-tg/resource"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key_id": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_ID", nil),
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_SECRET", nil),
				},
				"api_host": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_HOST", "api.trustgrid.io"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"tg_node": datasource.NodeDataSource(),
				"tg_org":  datasource.OrgDataSource(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"tg_compute_limits": resource.CPULimitsResource(),
				"tg_snmp":           resource.SNMPResource(),
				"tg_license":        resource.LicenseResource(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return &tg.Client{
			APIKey:    d.Get("api_key_id").(string),
			APISecret: d.Get("api_key_secret").(string),
			APIHost:   d.Get("api_host").(string),
		}, nil
	}
}
