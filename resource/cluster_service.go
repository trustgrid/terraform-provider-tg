package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/majordomo"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

// clusterServiceValidate rejects source_from_cluster_ip=true unless
// source_interface is also set. Enforced at plan time so the user sees the
// error before any API write.
func clusterServiceValidate(_ context.Context, d *schema.ResourceDiff, _ any) error {
	source, _ := d.Get("source_interface").(string)
	useClusterIP, _ := d.Get("source_from_cluster_ip").(bool)
	if useClusterIP && source == "" {
		return fmt.Errorf("source_from_cluster_ip = true requires source_interface to be set")
	}
	return nil
}

// clusterServiceImport parses an import ID of the form "{cluster_fqdn}:{service_id}".
func clusterServiceImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected import ID in form {cluster_fqdn}:{service_id}, got %q", d.Id())
	}
	if err := d.Set("cluster_fqdn", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("service_id", parts[1]); err != nil {
		return nil, err
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

// clusterServiceRead pulls the cluster, looks up the service by ID. Relies on
// tg.ServicesConfig.UnmarshalJSON to handle both V1 and V2 cluster shapes.
func clusterServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	fqdn, _ := d.Get("cluster_fqdn").(string)
	if fqdn == "" {
		d.SetId("")
		return nil
	}

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

	id := d.Id()
	var found *tg.Service
	for i := range cluster.Config.Services.Services {
		if cluster.Config.Services.Services[i].ID == id {
			found = &cluster.Config.Services.Services[i]
			break
		}
	}
	if found == nil {
		d.SetId("")
		return nil
	}

	tf, err := hcl.DecodeResourceData[hcl.ClusterService](d)
	if err != nil {
		return diag.FromErr(err)
	}
	updated := tf.UpdateFromTG(*found)
	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func ClusterService() *schema.Resource {
	md := majordomo.NewResource(majordomo.ResourceArgs[tg.Service, hcl.ClusterService]{
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
			if err := args.TF.Set("service_id", svc.ID); err != nil {
				return "", err
			}
			return svc.ID, nil
		},
		ID:       func(s hcl.ClusterService) string { return s.ServiceID },
		RemoteID: func(s tg.Service) string { return s.ID },
	})

	return &schema.Resource{
		Description:   "Manage an L4 service on a cluster (V2). Requires the cluster to have been upgraded to V2 services config — see `tg_cluster_services_v2_upgrade`.",
		CreateContext: md.Create,
		ReadContext:   clusterServiceRead,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		CustomizeDiff: clusterServiceValidate,
		Importer: &schema.ResourceImporter{
			StateContext: clusterServiceImport,
		},
		Schema: map[string]*schema.Schema{
			"service_id": {
				Description: "Service unique ID. Computed after create.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_fqdn": {
				Description:  "FQDN of the cluster that owns this service.",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Required:     true,
				ForceNew:     true,
			},
			"name": {
				Description: "Service name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"protocol": {
				Description:  "Protocol. One of: udp, tcp, tftp, ftp, rdp, vnc, ssh.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp", "tftp", "ftp", "rdp", "vnc", "ssh"}, false),
			},
			"host": {
				Description: "Destination host.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description:  "Destination port (1-65535).",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"description": {
				Description: "Description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Whether the service is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"source_interface": {
				Description: "NIC used for the upstream connection (e.g. `ens192`). V2-only. Note: setting this alone may produce VIP-sourcing depending on cluster IP topology (e.g. if the cluster VIP is a secondary IP on this NIC on the active node). For predictable VIP-sourcing behavior, set both `source_interface` and `source_from_cluster_ip = true`.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"source_from_cluster_ip": {
				Description: "When true, bind the outbound socket to the cluster VIP on `source_interface`. Requires `source_interface` to be set. V2-only.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
		},
	}
}
