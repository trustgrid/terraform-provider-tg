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

func nodeServiceImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected import ID in form {node_id}:{service_id}, got %q", d.Id())
	}
	if err := d.Set("node_id", parts[0]); err != nil {
		return nil, err
	}
	if err := d.Set("service_id", parts[1]); err != nil {
		return nil, err
	}
	d.SetId(parts[1])
	return []*schema.ResourceData{d}, nil
}

func nodeServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
	var found *tg.Service
	for i := range node.Config.Services.Services {
		if node.Config.Services.Services[i].ID == id {
			found = &node.Config.Services.Services[i]
			break
		}
	}
	if found == nil {
		d.SetId("")
		return nil
	}

	tf, err := hcl.DecodeResourceData[hcl.NodeService](d)
	if err != nil {
		return diag.FromErr(err)
	}
	updated := tf.UpdateFromTG(*found)
	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func NodeService() *schema.Resource {
	md := majordomo.NewResource(majordomo.ResourceArgs[tg.Service, hcl.NodeService]{
		CreateURL: func(r hcl.NodeService) string {
			return fmt.Sprintf("/v2/node/%s/config/services", r.NodeID)
		},
		UpdateURL: func(r hcl.NodeService) string {
			return fmt.Sprintf("/v2/node/%s/config/services/%s", r.NodeID, r.ServiceID)
		},
		DeleteURL: func(r hcl.NodeService) string {
			return fmt.Sprintf("/v2/node/%s/config/services/%s", r.NodeID, r.ServiceID)
		},
		OnCreateReply: func(_ context.Context, args majordomo.CallbackArgs[tg.Service, hcl.NodeService]) (string, error) {
			var svc tg.Service
			if err := json.Unmarshal(args.Body, &svc); err != nil {
				return "", fmt.Errorf("parsing create response: %w", err)
			}
			if svc.ID == "" {
				return "", fmt.Errorf("node service created but response had no id")
			}
			if err := args.TF.Set("service_id", svc.ID); err != nil {
				return "", err
			}
			return svc.ID, nil
		},
		ID:       func(s hcl.NodeService) string { return s.ServiceID },
		RemoteID: func(s tg.Service) string { return s.ID },
	})

	return &schema.Resource{
		Description:   "Manage an L4 service on a node (V2). Requires the node to have been upgraded to V2 services config — see `tg_node_services_v2_upgrade`.",
		CreateContext: md.Create,
		ReadContext:   nodeServiceRead,
		UpdateContext: md.Update,
		DeleteContext: md.Delete,
		Importer: &schema.ResourceImporter{
			StateContext: nodeServiceImport,
		},
		Schema: map[string]*schema.Schema{
			"service_id": {
				Description: "Service unique ID. Computed after create.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"node_id": {
				Description:  "Node UID that owns this service.",
				Type:         schema.TypeString,
				ValidateFunc: validation.IsUUID,
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
				Description: "NIC used for the upstream connection. V2-only.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}
