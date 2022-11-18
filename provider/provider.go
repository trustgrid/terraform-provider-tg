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
					Description: "Trustgrid Portal API Key ID. Will use `TG_API_KEY_ID` environment variable if not set.",
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_ID", nil),
				},
				"api_key_secret": {
					Type:        schema.TypeString,
					Description: "Trustgrid Portal API Key secret. Will use `TG_API_KEY_SECRET` environment variable if not set.",
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_KEY_SECRET", nil),
				},
				"api_host": {
					Type:        schema.TypeString,
					Description: "Trustgrid Portal endpoint. Used for development.",
					Optional:    true,
					Sensitive:   false,
					DefaultFunc: schema.EnvDefaultFunc("TG_API_HOST", "api.trustgrid.io"),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"tg_app":             datasource.App(),
				"tg_cert":            datasource.Cert(),
				"tg_idp":             datasource.IDP(),
				"tg_node":            datasource.Node(),
				"tg_nodes":           datasource.Nodes(),
				"tg_org":             datasource.Org(),
				"tg_virtual_network": datasource.VirtualNetwork(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"tg_app":                         resource.App(),
				"tg_app_access_rule":             resource.AppAccessRule(),
				"tg_app_acl":                     resource.AppACL(),
				"tg_cert":                        resource.Cert(),
				"tg_cluster":                     resource.Cluster(),
				"tg_cluster_member":              resource.ClusterMember(),
				"tg_compute_limits":              resource.CPULimits(),
				"tg_container":                   resource.Container(),
				"tg_container_volume":            resource.Volume(),
				"tg_gateway_config":              resource.GatewayConfig(),
				"tg_idp":                         resource.IDP(),
				"tg_license":                     resource.License(),
				"tg_network_config":              resource.NetworkConfig(),
				"tg_node_cluster_config":         resource.ClusterConfig(),
				"tg_snmp":                        resource.SNMP(),
				"tg_virtual_network":             resource.VirtualNetwork(),
				"tg_virtual_network_route":       resource.VNetRoute(),
				"tg_virtual_network_access_rule": resource.VNetAccessRule(),
				"tg_virtual_network_attachment":  resource.VNetAttachment(),
				"tg_vpn_interface":               resource.VPNInterface(),
				"tg_ztna_gateway_config":         resource.ZTNAConfig(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		c, err := tg.NewClient(ctx, d.Get("api_key_id").(string), d.Get("api_key_secret").(string), d.Get("api_host").(string))
		if err != nil {
			return c, diag.FromErr(err)
		}
		return c, nil
	}
}
