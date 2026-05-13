package resource

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

// v2Upgrade returns a one-shot resource that POSTs to a V1→V2 upgrade
// endpoint on Create. The upgrade is one-way per the API, so Read/Update/Delete
// are intentionally no-ops — state presence represents "this target has been
// upgraded." Terraform destroy does NOT roll the target back; manage that
// expectation in user docs.
func v2Upgrade(description, targetField, urlPattern string, targetValidator schema.SchemaValidateFunc) *schema.Resource {
	return &schema.Resource{
		Description: description,
		CreateContext: func(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
			tgc := tg.GetClient(meta)
			target, _ := d.Get(targetField).(string)
			if target == "" {
				return diag.Errorf("%s is required", targetField)
			}
			url := fmt.Sprintf(urlPattern, target)
			if _, err := tgc.Post(ctx, url, struct{}{}); err != nil {
				return diag.FromErr(fmt.Errorf("posting upgrade %s: %w", url, err))
			}
			d.SetId(target)
			return nil
		},
		ReadContext: func(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
			// Upgrade has no read shape; presence in state is the only signal.
			return nil
		},
		// No UpdateContext — every field is ForceNew, so an update path is
		// unreachable and the SDK rejects defining one.
		DeleteContext: func(_ context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
			// Upgrade is one-way per the API. Drop from state without calling
			// anything — the target stays on V2.
			d.SetId("")
			return nil
		},
		Schema: map[string]*schema.Schema{
			targetField: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: targetValidator,
				Description:  fmt.Sprintf("%s to upgrade.", targetField),
			},
		},
	}
}

func ClusterServicesV2Upgrade() *schema.Resource {
	return v2Upgrade(
		"Trigger the V1→V2 services config upgrade on a cluster. One-shot — Terraform destroy is a no-op (the upgrade is irreversible per the API).",
		"cluster_fqdn",
		"/v2/cluster/%s/config/services/upgrade",
		validators.IsHostname,
	)
}

func ClusterConnectorsV2Upgrade() *schema.Resource {
	return v2Upgrade(
		"Trigger the V1→V2 connectors config upgrade on a cluster. One-shot — Terraform destroy is a no-op (the upgrade is irreversible per the API).",
		"cluster_fqdn",
		"/v2/cluster/%s/config/connectors/upgrade",
		validators.IsHostname,
	)
}

func NodeServicesV2Upgrade() *schema.Resource {
	return v2Upgrade(
		"Trigger the V1→V2 services config upgrade on a node. One-shot — Terraform destroy is a no-op (the upgrade is irreversible per the API).",
		"node_id",
		"/v2/node/%s/config/services/upgrade",
		validation.IsUUID,
	)
}

func NodeConnectorsV2Upgrade() *schema.Resource {
	return v2Upgrade(
		"Trigger the V1→V2 connectors config upgrade on a node. One-shot — Terraform destroy is a no-op (the upgrade is irreversible per the API).",
		"node_id",
		"/v2/node/%s/config/connectors/upgrade",
		validation.IsUUID,
	)
}
