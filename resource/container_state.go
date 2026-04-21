package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

func ContainerState() *schema.Resource {
	schemaMap := containerTargetSchema()
	schemaMap["enabled"] = &schema.Schema{
		Description: "Whether the container should be running",
		Type:        schema.TypeBool,
		Required:    true,
	}

	return &schema.Resource{
		Description: "Manage whether a node or cluster container is running",

		ReadContext:   containerStateRead,
		CreateContext: containerStateCreate,
		UpdateContext: containerStateUpdate,
		DeleteContext: containerActionNoop,

		Schema: schemaMap,
	}
}

func containerStateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return containerStateUpsert(ctx, d, meta)
}

func containerStateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return containerStateUpsert(ctx, d, meta)
}

func containerStateUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	target := containerTarget{
		NodeID:      d.Get("node_id").(string),      //nolint: errcheck // schema guarantees type
		ClusterFQDN: d.Get("cluster_fqdn").(string), //nolint: errcheck // schema guarantees type
		ContainerID: d.Get("container_id").(string), //nolint: errcheck // schema guarantees type
	}
	enabled := d.Get("enabled").(bool) //nolint: errcheck // schema guarantees type

	if err := updateContainerEnabled(ctx, tg.GetClient(meta), target, enabled); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(target.resourceID())
	return containerStateRead(ctx, d, meta)
}

func containerStateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	container, target, diags := readContainerActionTarget(ctx, d, meta)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return nil
	}

	state := hcl.ContainerState{
		NodeID:      target.NodeID,
		ClusterFQDN: target.ClusterFQDN,
		ContainerID: target.ContainerID,
		Enabled:     container.Enabled,
	}

	if err := hcl.EncodeResourceData(state, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(target.resourceID())
	return nil
}
