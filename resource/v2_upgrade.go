package resource

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

// postUpgradeIdempotent POSTs to the upgrade endpoint and treats a 422
// response as success (target is already V2). This makes the resource
// idempotent — running `terraform apply` against an already-upgraded
// target is a no-op rather than an error.
func postUpgradeIdempotent(ctx context.Context, tgc *tg.Client, url string) error {
	_, err := tgc.Post(ctx, url, struct{}{})
	if err == nil {
		return nil
	}
	// The portal returns 422 with body "Cannot modify ... for V1 config" or
	// similar when the target is already V2. Treat any 422 from the upgrade
	// endpoint as idempotent success.
	if strings.Contains(err.Error(), ": 422") {
		return nil
	}
	return err
}

func ClusterServicesV2Upgrade() *schema.Resource {
	return &schema.Resource{
		Description: "Trigger the V1→V2 services config upgrade on a cluster. Idempotent — applying against an already-V2 cluster is a no-op. Destroy is also a no-op (the upgrade is irreversible per the API).",
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			tgc := tg.GetClient(meta)
			fqdn, _ := d.Get("cluster_fqdn").(string)
			url := fmt.Sprintf("/v2/cluster/%s/config/services/upgrade", fqdn)
			if err := postUpgradeIdempotent(ctx, tgc, url); err != nil {
				return diag.FromErr(fmt.Errorf("posting upgrade %s: %w", url, err))
			}
			d.SetId(fqdn)
			return nil
		},
		ReadContext:   noopRead,
		DeleteContext: noopDelete,
		Schema: map[string]*schema.Schema{
			"cluster_fqdn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				Description:  "FQDN of the cluster to upgrade.",
			},
		},
	}
}

func ClusterConnectorsV2Upgrade() *schema.Resource {
	return &schema.Resource{
		Description: "Trigger the V1→V2 connectors config upgrade on a cluster. Idempotent — applying against an already-V2 cluster is a no-op. Destroy is also a no-op.",
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			tgc := tg.GetClient(meta)
			fqdn, _ := d.Get("cluster_fqdn").(string)
			url := fmt.Sprintf("/v2/cluster/%s/config/connectors/upgrade", fqdn)
			if err := postUpgradeIdempotent(ctx, tgc, url); err != nil {
				return diag.FromErr(fmt.Errorf("posting upgrade %s: %w", url, err))
			}
			d.SetId(fqdn)
			return nil
		},
		ReadContext:   noopRead,
		DeleteContext: noopDelete,
		Schema: map[string]*schema.Schema{
			"cluster_fqdn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				Description:  "FQDN of the cluster to upgrade.",
			},
		},
	}
}

func NodeServicesV2Upgrade() *schema.Resource {
	return &schema.Resource{
		Description: "Trigger the V1→V2 services config upgrade on a node. Idempotent — applying against an already-V2 node is a no-op. Destroy is also a no-op.",
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			tgc := tg.GetClient(meta)
			nodeID, _ := d.Get("node_id").(string)
			url := fmt.Sprintf("/v2/node/%s/config/services/upgrade", nodeID)
			if err := postUpgradeIdempotent(ctx, tgc, url); err != nil {
				return diag.FromErr(fmt.Errorf("posting upgrade %s: %w", url, err))
			}
			d.SetId(nodeID)
			return nil
		},
		ReadContext:   noopRead,
		DeleteContext: noopDelete,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "UUID of the node to upgrade.",
			},
		},
	}
}

func NodeConnectorsV2Upgrade() *schema.Resource {
	return &schema.Resource{
		Description: "Trigger the V1→V2 connectors config upgrade on a node. Idempotent — applying against an already-V2 node is a no-op. Destroy is also a no-op.",
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			tgc := tg.GetClient(meta)
			nodeID, _ := d.Get("node_id").(string)
			url := fmt.Sprintf("/v2/node/%s/config/connectors/upgrade", nodeID)
			if err := postUpgradeIdempotent(ctx, tgc, url); err != nil {
				return diag.FromErr(fmt.Errorf("posting upgrade %s: %w", url, err))
			}
			d.SetId(nodeID)
			return nil
		},
		ReadContext:   noopRead,
		DeleteContext: noopDelete,
		Schema: map[string]*schema.Schema{
			"node_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				Description:  "UUID of the node to upgrade.",
			},
		},
	}
}

func noopRead(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func noopDelete(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
	d.SetId("")
	return nil
}
