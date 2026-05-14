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
)

func nodeConnectorImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected import ID in form {node_id}:{connector_id}, got %q", d.Id())
	}
	if err := d.Set("node_id", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("connector_id", parts[1]); err != nil {
		return nil, err
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func nodeConnectorRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	nodeID, _ := d.Get("node_id").(string)
	if nodeID == "" {
		d.SetId("")
		return nil
	}

	var node tg.Node
	err := tgc.Get(ctx, fmt.Sprintf("/node/%s", nodeID), &node)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	id := d.Id()
	var found *tg.Connector
	for i := range node.Config.Connectors.Connectors {
		if node.Config.Connectors.Connectors[i].ID == id {
			found = &node.Config.Connectors.Connectors[i]
			break
		}
	}
	if found == nil {
		d.SetId("")
		return nil
	}

	tf, err := hcl.DecodeResourceData[hcl.NodeConnector](d)
	if err != nil {
		return diag.FromErr(err)
	}
	updated := tf.UpdateFromTG(*found)
	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func NodeConnector() *schema.Resource {
	md := majordomo.NewResource(majordomo.ResourceArgs[tg.Connector, hcl.NodeConnector]{
		CreateURL: func(r hcl.NodeConnector) string {
			return fmt.Sprintf("/v2/node/%s/config/connectors", r.NodeID)
		},
		UpdateURL: func(r hcl.NodeConnector) string {
			return fmt.Sprintf("/v2/node/%s/config/connectors/%s", r.NodeID, r.ConnectorID)
		},
		DeleteURL: func(r hcl.NodeConnector) string {
			return fmt.Sprintf("/v2/node/%s/config/connectors/%s", r.NodeID, r.ConnectorID)
		},
		OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.Connector, hcl.NodeConnector]) (string, error) {
			var conn tg.Connector
			if err := json.Unmarshal(args.Body, &conn); err != nil {
				return "", fmt.Errorf("parsing create response: %w", err)
			}
			if conn.ID == "" {
				return "", fmt.Errorf("node connector created but response had no id")
			}
			if err := args.TF.Set("connector_id", conn.ID); err != nil {
				return "", err
			}
			return conn.ID, nil
		},
		ID:       func(c hcl.NodeConnector) string { return c.ConnectorID },
		RemoteID: func(c tg.Connector) string { return c.ID },
	})

	return &schema.Resource{
		Description:   "Manage a forwarding connector on a node (V2). Requires the node to have been upgraded to V2 connectors config — see `tg_node_connectors_v2_upgrade`.",
		CreateContext: md.Create,
		ReadContext:   nodeConnectorRead,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		Importer: &schema.ResourceImporter{
			StateContext: nodeConnectorImport,
		},
		Schema: map[string]*schema.Schema{
			"connector_id": {
				Description: "Connector unique ID. Computed after create.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"node_id": {
				Description:  "Node UID that owns this connector.",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
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
				Description:  "Listening port (1-65535).",
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
