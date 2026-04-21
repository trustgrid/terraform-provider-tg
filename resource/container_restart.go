package resource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
)

type containerRestart struct {
	NodeID      string            `tf:"node_id"`
	ClusterFQDN string            `tf:"cluster_fqdn"`
	ContainerID string            `tf:"container_id"`
	Triggers    map[string]string `tf:"triggers"`
}

func ContainerRestart() *schema.Resource {
	schemaMap := containerTargetSchema()
	schemaMap["triggers"] = &schema.Schema{
		Description: "Values that should trigger a container restart when they change",
		Type:        schema.TypeMap,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}

	return &schema.Resource{
		Description: "Restart a node or cluster container when trigger values change",

		ReadContext:   containerRestartRead,
		CreateContext: containerRestartCreate,
		DeleteContext: containerActionNoop,

		Schema: schemaMap,
	}
}

func containerRestartCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tf, err := hcl.DecodeResourceData[containerRestart](d)
	if err != nil {
		return diag.FromErr(err)
	}

	target := containerTarget{
		NodeID:      tf.NodeID,
		ClusterFQDN: tf.ClusterFQDN,
		ContainerID: tf.ContainerID,
	}

	if err := restartContainer(ctx, tg.GetClient(meta), target); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(target.resourceID() + "/" + triggerHash(tf.Triggers))
	return containerRestartRead(ctx, d, meta)
}

func containerRestartRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	_, target, diags := readContainerActionTarget(ctx, d, meta)
	if diags != nil {
		return diags
	}
	if d.Id() == "" {
		return nil
	}

	tf, err := hcl.DecodeResourceData[containerRestart](d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(target.resourceID() + "/" + triggerHash(tf.Triggers))
	return nil
}

func containerActionNoop(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
