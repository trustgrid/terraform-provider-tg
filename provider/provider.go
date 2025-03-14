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
				"api_jwt": {
					Type:        schema.TypeString,
					Description: "Trustgrid Portal JWT. Used for short-lived authentication. Will use the `TG_JWT` environment variable.",
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("TG_JWT", nil),
				},
				"org_id": {
					Type:        schema.TypeString,
					Description: "Trustgrid Org ID. If provided and the credentials aren't for that org, the provider will fail early.",
					Optional:    true,
					Sensitive:   false,
					DefaultFunc: schema.EnvDefaultFunc("TG_ORG_ID", nil),
				},
			},
			DataSourcesMap: map[string]*schema.Resource{
				"tg_alarm":           datasource.Alarm(),
				"tg_alarm_channel":   datasource.AlarmChannel(),
				"tg_app":             datasource.App(),
				"tg_cert":            datasource.Cert(),
				"tg_cluster":         datasource.Cluster(),
				"tg_device_info":     datasource.Device(),
				"tg_group":           datasource.Group(),
				"tg_idp":             datasource.IDP(),
				"tg_network_config":  datasource.NetworkConfig(),
				"tg_node":            datasource.Node(),
				"tg_nodes":           datasource.Nodes(),
				"tg_org":             datasource.Org(),
				"tg_kvm_image":       datasource.KVMImage(),
				"tg_kvm_volume":      datasource.KVMVolume(),
				"tg_shadow":          datasource.Shadow(),
				"tg_virtual_network": datasource.VirtualNetwork(),
			},
			ResourcesMap: map[string]*schema.Resource{
				"tg_alarm":                        resource.Alarm(),
				"tg_alarm_channel":                resource.AlarmChannel(),
				"tg_app":                          resource.App(),
				"tg_app_access_rule":              resource.AppAccessRule(),
				"tg_app_acl":                      resource.AppACL(),
				"tg_cert":                         resource.Cert(),
				"tg_cluster":                      resource.Cluster(),
				"tg_cluster_member":               resource.ClusterMember(),
				"tg_compute_limits":               resource.CPULimits(),
				"tg_connector":                    resource.Connector(),
				"tg_container":                    resource.Container(),
				"tg_container_volume":             resource.Volume(),
				"tg_gateway_config":               resource.GatewayConfig(),
				"tg_group":                        resource.Group(),
				"tg_group_member":                 resource.GroupMember(),
				"tg_idp":                          resource.IDP(),
				"tg_idp_openid_config":            resource.IDPOpenIDConfig(),
				"tg_idp_saml_config":              resource.IDPSAMLConfig(),
				"tg_kvm_image":                    resource.KVMImage(),
				"tg_kvm_volume":                   resource.KVMVolume(),
				"tg_license":                      resource.License(),
				"tg_network_config":               resource.NetworkConfig(),
				"tg_node_state":                   resource.NodeState(),
				"tg_node_cluster_config":          resource.ClusterConfig(),
				"tg_policy":                       resource.Policy(),
				"tg_portal_auth":                  resource.PortalAuth(),
				"tg_service":                      resource.Service(),
				"tg_serviceuser":                  resource.ServiceUser(),
				"tg_snmp":                         resource.SNMP(),
				"tg_tagging":                      resource.Tagging(),
				"tg_virtual_network":              resource.VirtualNetwork(),
				"tg_virtual_network_access_rule":  resource.VNetAccessRule(),
				"tg_virtual_network_attachment":   resource.VNetAttachment(),
				"tg_virtual_network_port_forward": resource.VNetPortForward(),
				"tg_virtual_network_route":        resource.VNetRoute(),
				"tg_vpn_interface":                resource.VPNInterface(),
				"tg_ztna_gateway_config":          resource.ZTNAConfig(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(_ string, _ *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		cp := tg.ClientParams{
			APIKey:    d.Get("api_key_id").(string),     //nolint: errcheck // just trusting TF validation here
			APISecret: d.Get("api_key_secret").(string), //nolint: errcheck // just trusting TF validation here
			APIHost:   d.Get("api_host").(string),       //nolint: errcheck // just trusting TF validation here
			JWT:       d.Get("api_jwt").(string),        //nolint: errcheck // just trusting TF validation here
		}
		if orgid, ok := d.Get("org_id").(string); ok {
			cp.OrgID = orgid
		}
		c, err := tg.NewClient(ctx, cp)

		if err != nil {
			return c, diag.FromErr(err)
		}
		return c, nil
	}
}
