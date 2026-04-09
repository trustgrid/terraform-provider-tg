package resource

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

type containerState struct{}

// ContainerState manages the enabled state of a container independently of its configuration.
func ContainerState() *schema.Resource {
	cs := containerState{}

	return &schema.Resource{
		Description: "Manage the enabled state of a container. This resource allows you to control whether a container is running or stopped, independent of the container configuration.",

		ReadContext:   cs.Read,
		UpdateContext: cs.Update,
		DeleteContext: cs.Delete,
		CreateContext: cs.Create,

		Importer: &schema.ResourceImporter{
			StateContext: importContainerResource,
		},

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID - the UUID of the node where the container runs",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN - the fully qualified domain name of the cluster where the container runs",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"container_id": {
				Description:  "Container ID - the UUID of the container to manage",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"enabled": {
				Description: "Whether the container should be enabled (running) or disabled (stopped)",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func (cs *containerState) containerURL(c hcl.ContainerState) string {
	if c.NodeID != "" {
		return "/v2/node/" + c.NodeID + "/exec/container/" + c.ContainerID
	}
	return "/v2/cluster/" + c.ClusterFQDN + "/exec/container/" + c.ContainerID
}

func (cs *containerState) resourceID(c hcl.ContainerState) string {
	if c.NodeID != "" {
		return c.NodeID + "/" + c.ContainerID
	}
	return c.ClusterFQDN + "/" + c.ContainerID
}

func (cs *containerState) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.ContainerState](d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Read current container state
	var current tg.Container
	if err := tgc.Get(ctx, cs.containerURL(tf), &current); err != nil {
		return diag.FromErr(err)
	}

	// Update the enabled field
	current.Enabled = tf.Enabled

	// Write back the updated container
	if _, err := tgc.Put(ctx, cs.containerURL(tf), current); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cs.resourceID(tf))

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (cs *containerState) Update(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.ContainerState](d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Read current container state
	var current tg.Container
	if err := tgc.Get(ctx, cs.containerURL(tf), &current); err != nil {
		return diag.FromErr(err)
	}

	// Update the enabled field
	current.Enabled = tf.Enabled

	// Write back the updated container
	if _, err := tgc.Put(ctx, cs.containerURL(tf), current); err != nil {
		return diag.FromErr(err)
	}

	if err := hcl.EncodeResourceData(tf, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// Delete is a no-op for container state - we just remove it from Terraform state.
// The container will remain in whatever state it was in.
func (cs *containerState) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func (cs *containerState) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[hcl.ContainerState](d)
	if err != nil {
		return diag.FromErr(err)
	}

	var current tg.Container
	err = tgc.Get(ctx, cs.containerURL(tf), &current)

	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	// Update the HCL state from the API response
	apiState := tg.ContainerState{
		NodeID:      tf.NodeID,
		ClusterFQDN: tf.ClusterFQDN,
		ContainerID: tf.ContainerID,
		Enabled:     current.Enabled,
	}

	updated := tf.UpdateFromTG(apiState)

	if err := hcl.EncodeResourceData(updated, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
