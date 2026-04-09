package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/hcl"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

const (
	// containerStopWaitDuration is the time to wait after disabling a container
	// before re-enabling it. This allows the container runtime to fully stop
	// the container before restarting. 2 seconds is typically sufficient for
	// most containers to receive and process SIGTERM.
	containerStopWaitDuration = 2 * time.Second
)

type containerRestart struct{}

// ContainerRestart triggers a container restart when triggers change.
// This is useful for forcing a container to pull new images when the image tag or other configuration changes.
func ContainerRestart() *schema.Resource {
	cr := containerRestart{}

	return &schema.Resource{
		Description: "Trigger a container restart when triggers change. This resource is useful for forcing a container to pull new images when the image tag or other configuration changes. The restart is performed by disabling the container, waiting briefly, and re-enabling it. If the container is already disabled, no restart is performed.",

		ReadContext:   cr.Read,
		DeleteContext: cr.Delete,
		CreateContext: cr.Create,

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
				Description:  "Container ID - the UUID of the container to restart",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"triggers": {
				Description: "A map of arbitrary strings that, when changed, will trigger a container restart. This can be used to restart a container when an image tag, configuration, or other external value changes.",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// ContainerRestart HCL type for reading Terraform state.
type containerRestartHCL struct {
	NodeID      string            `tf:"node_id"`
	ClusterFQDN string            `tf:"cluster_fqdn"`
	ContainerID string            `tf:"container_id"`
	Triggers    map[string]string `tf:"triggers"`
}

func (cr *containerRestart) containerURL(c containerRestartHCL) string {
	if c.NodeID != "" {
		return "/v2/node/" + c.NodeID + "/exec/container/" + c.ContainerID
	}
	return "/v2/cluster/" + c.ClusterFQDN + "/exec/container/" + c.ContainerID
}

func (cr *containerRestart) resourceID(c containerRestartHCL) string {
	if c.NodeID != "" {
		return c.NodeID + "/" + c.ContainerID
	}
	return c.ClusterFQDN + "/" + c.ContainerID
}

// performRestart disables the container (if it is enabled), waits briefly,
// then re-enables it. If the container is already disabled, it is left as-is
// and no restart is performed.
func (cr *containerRestart) performRestart(ctx context.Context, tgc *tg.Client, tf containerRestartHCL) error {
	url := cr.containerURL(tf)

	// Read current container state
	var current tg.Container
	if err := tgc.Get(ctx, url, &current); err != nil {
		return err
	}

	// Only restart if the container is currently enabled.
	// A disabled container should not be force-enabled by a restart.
	if !current.Enabled {
		return nil
	}

	// Disable the container
	current.Enabled = false
	if _, err := tgc.Put(ctx, url, current); err != nil {
		return err
	}

	// Wait briefly to allow the container to stop, but respect context cancellation
	select {
	case <-time.After(containerStopWaitDuration):
		// Continue with restart
	case <-ctx.Done():
		// Context was cancelled - attempt to restore original state
		current.Enabled = true
		_, _ = tgc.Put(context.Background(), url, current) // best-effort restore, ignore error
		return ctx.Err()
	}

	// Re-enable the container (it was enabled before we started)
	current.Enabled = true
	if _, err := tgc.Put(ctx, url, current); err != nil {
		// Attempt to restore original state on failure
		if _, restoreErr := tgc.Put(context.Background(), url, current); restoreErr != nil {
			return fmt.Errorf("restart failed and could not restore original state: %w (restore error: %s)", err, restoreErr.Error())
		}
		return fmt.Errorf("restart failed, restored to original state: %w", err)
	}

	return nil
}

func (cr *containerRestart) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[containerRestartHCL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Perform the restart
	if err := cr.performRestart(ctx, tgc, tf); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cr.resourceID(tf))

	return nil
}

// Delete is a no-op - we just remove the resource from Terraform state.
// The container will remain in whatever state it was in.
func (cr *containerRestart) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}

func (cr *containerRestart) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	tf, err := hcl.DecodeResourceData[containerRestartHCL](d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Verify the container still exists
	var current tg.Container
	err = tgc.Get(ctx, cr.containerURL(tf), &current)

	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	// The triggers are stored in state and don't come from the API.
	// We don't need to update them from the API response since they're
	// purely Terraform-side values used to trigger restarts.
	// Just keep the current state as-is.

	return nil
}

// importContainerResource parses an import ID of the form
// "<node-id-or-cluster-fqdn>/<container-id>" and sets the appropriate
// fields in state. If the first segment is a valid UUID it is treated as a
// node_id; otherwise it is treated as a cluster_fqdn.
func importContainerResource(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
	parts := strings.SplitN(d.Id(), "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("import ID must be in the format <node-id-or-cluster-fqdn>/<container-id>, got: %s", d.Id())
	}

	nodeOrCluster := parts[0]
	containerID := parts[1]

	if _, err := uuid.Parse(nodeOrCluster); err == nil {
		if err := d.Set("node_id", nodeOrCluster); err != nil {
			return nil, err
		}
	} else {
		if err := d.Set("cluster_fqdn", nodeOrCluster); err != nil {
			return nil, err
		}
	}

	if err := d.Set("container_id", containerID); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
