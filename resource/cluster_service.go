package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

func ClusterService() *schema.Resource {
	md := majordomo.NewResource(
		majordomo.ResourceArgs[tg.Service, hcl.ClusterService]{
			CreateURL: func(r hcl.ClusterService) string {
				return fmt.Sprintf("/v2/cluster/%s/config/services", r.ClusterFQDN)
			},
			UpdateURL: func(r hcl.ClusterService) string {
				return fmt.Sprintf("/v2/cluster/%s/config/services/%s", r.ClusterFQDN, r.ServiceID)
			},
			DeleteURL: func(r hcl.ClusterService) string {
				return fmt.Sprintf("/v2/cluster/%s/config/services/%s", r.ClusterFQDN, r.ServiceID)
			},
			IndexURL: func(r hcl.ClusterService) string {
				return fmt.Sprintf("/v2/cluster/%s/config/services", r.ClusterFQDN)
			},
			OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.Service, hcl.ClusterService]) (string, error) {
				var svc tg.Service
				if err := json.Unmarshal(args.Body, &svc); err != nil {
					return "", fmt.Errorf("parsing create response: %w", err)
				}
				if svc.ID == "" {
					return "", fmt.Errorf("cluster service created but response had no id")
				}
				return svc.ID, nil
			},
			AfterCreate: func(_ context.Context, args majordomo.CallbackArgs[tg.Service, hcl.ClusterService]) (string, error) {
				return "", args.TF.Set("service_id", args.TF.Id())
			},
			ID: func(s hcl.ClusterService) string {
				return s.ServiceID
			},
			RemoteID: func(s tg.Service) string {
				return s.ID
			},
		})

	return &schema.Resource{
		Description: "Manage an L4 service on a cluster.",

		ReadContext:   md.Read,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,

		Schema: map[string]*schema.Schema{
			"service_id": {
				Description: "Service unique ID",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Required:     true,
				ForceNew:     true,
			},
			"name": {
				Description: "Service name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"protocol": {
				Description:  "Protocol",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp", "tftp", "ftp", "rdp", "vnc", "ssh"}, false),
			},
			"host": {
				Description: "Destination host",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description:  "Destination port",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"description": {
				Description: "Description",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Whether the service is enabled",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"source_interface": {
				Description: "NIC used for the upstream connection (e.g. ens192)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"source_from_cluster_ip": {
				Description: "When true (and source_interface is set), bind the outbound socket to the cluster VIP on that NIC",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
