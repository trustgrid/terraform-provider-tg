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

func clusterConnectorImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected import ID in form {cluster_fqdn}:{connector_id}, got %q", d.Id())
	}
	if err := d.Set("cluster_fqdn", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("connector_id", parts[1]); err != nil {
		return nil, err
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func clusterConnectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	if cluster.Config.Connectors == nil {
		d.SetId("")
		return nil
	}

	id := d.Id()
	var found *tg.Connector
	for i := range cluster.Config.Connectors.Connectors {
		if cluster.Config.Connectors.Connectors[i].ID == id {
			found = &cluster.Config.Connectors.Connectors[i]
			break
		}
	}
	if found == nil {
		d.SetId("")
		return nil
	}

	tf, err := hcl.DecodeResourceData[hcl.ClusterConnector](d)
	if err != nil {
		return diag.FromErr(err)
	}
	updated := tf.UpdateFromTG(*found)
	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func ClusterConnector() *schema.Resource {
	md := majordomo.NewResource(majordomo.ResourceArgs[tg.Connector, hcl.ClusterConnector]{
		CreateURL: func(r hcl.ClusterConnector) string {
			return fmt.Sprintf("/v2/cluster/%s/config/connectors", r.ClusterFQDN)
		},
		UpdateURL: func(r hcl.ClusterConnector) string {
			return fmt.Sprintf("/v2/cluster/%s/config/connectors/%s", r.ClusterFQDN, r.ConnectorID)
		},
		DeleteURL: func(r hcl.ClusterConnector) string {
			return fmt.Sprintf("/v2/cluster/%s/config/connectors/%s", r.ClusterFQDN, r.ConnectorID)
		},
		OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.Connector, hcl.ClusterConnector]) (string, error) {
			var conn tg.Connector
			if err := json.Unmarshal(args.Body, &conn); err != nil {
				return "", fmt.Errorf("parsing create response: %w", err)
			}
			if conn.ID == "" {
				return "", fmt.Errorf("cluster connector created but response had no id")
			}
			if err := args.TF.Set("connector_id", conn.ID); err != nil {
				return "", err
			}
			return conn.ID, nil
		},
		ID:       func(c hcl.ClusterConnector) string { return c.ConnectorID },
		RemoteID: func(c tg.Connector) string { return c.ID },
	})

	return &schema.Resource{
		Description:   "Manage a forwarding connector on a cluster (V2). Requires the cluster to have been upgraded to V2 connectors config — see `tg_cluster_connectors_v2_upgrade`.",
		CreateContext: md.Create,
		ReadContext:   clusterConnectorRead,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		Importer: &schema.ResourceImporter{
			StateContext: clusterConnectorImport,
		},
		Schema: map[string]*schema.Schema{
			"connector_id": {
				Description: "Connector unique ID. Computed after create.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"cluster_fqdn": {
				Description:  "FQDN of the cluster that owns this connector.",
				Type:         schema.TypeString,
				ValidateFunc: validators.IsHostname,
				Required:     true,
				ForceNew:     true,
			},
			"node": {
				Description: "Node identifier (`local` or a node UID).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"service": {
				Description: "Upstream service address (`host:port`).",
				Type:        schema.TypeString,
				Required:    true,
			},
			"port": {
				Description:  "Listening port on the gateway (1-65535).",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"protocol": {
				Description:  "Protocol. One of: udp, tcp, tftp, ftp, rdp, vnc, ssh.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"udp", "tcp", "tftp", "ftp", "rdp", "vnc", "ssh"}, false),
			},
			"description": {
				Description: "Description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"enabled": {
				Description: "Whether the connector is enabled.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"rate_limit": {
				Description: "Max throughput in Mbps (0 for unlimited).",
				Type:        schema.TypeInt,
				Optional:    true,
			},
			"nic": {
				Description: "NIC to listen on.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
