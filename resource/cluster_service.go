package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

func clusterServiceValidate(_ context.Context, d *schema.ResourceDiff, _ any) error {
	tf, err := hcl.DecodeResourceDiff[hcl.ClusterService](d)
	if err != nil {
		return err
	}
	if tf.SourceFromClusterIP && tf.SourceInterface == "" {
		return fmt.Errorf("source_from_cluster_ip = true requires source_interface to be set")
	}
	return nil
}

func clusterServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	fqdn, _ := d.Get("cluster_fqdn").(string)
	id := d.Id()

	var cluster tg.Cluster
	err := tgc.Get(ctx, fmt.Sprintf("/cluster/%s", fqdn), &cluster)

	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	if cluster.Config.Services == nil {
		d.SetId("")
		return nil
	}

	svc, ok := cluster.Config.Services.Items[id]
	if !ok {
		d.SetId("")
		return nil
	}

	tf, decErr := hcl.DecodeResourceData[hcl.ClusterService](d)
	if decErr != nil {
		return diag.FromErr(decErr)
	}
	updated := tf.UpdateFromTG(svc)
	if encErr := hcl.EncodeResourceData(updated, d); encErr != nil {
		return diag.FromErr(encErr)
	}
	return nil
}

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

		ReadContext:   clusterServiceRead,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CreateContext: md.Create,
		CustomizeDiff: clusterServiceValidate,

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
				Description: "When true, bind the outbound socket to the cluster VIP on that NIC. Requires source_interface to be set.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
