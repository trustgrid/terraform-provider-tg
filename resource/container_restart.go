package resource

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

const containerStopWaitDuration = 2 * time.Second

type containerRestart struct{}

type containerRestartHCL struct {
	NodeID      string
	ClusterFQDN string
	ContainerID string
	Triggers    map[string]string
}

// ContainerRestart triggers a container restart when triggers map changes.
func ContainerRestart() *schema.Resource {
	cr := containerRestart{}

	return &schema.Resource{
		Description: "Trigger a container restart when the triggers map changes.",

		ReadContext:   cr.Read,
		CreateContext: cr.Create,
		DeleteContext: cr.Delete,

		Schema: map[string]*schema.Schema{
			"node_id": {
				Description:  "Node ID",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"cluster_fqdn": {
				Description:  "Cluster FQDN",
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validators.IsHostname,
				ExactlyOneOf: []string{"node_id", "cluster_fqdn"},
			},
			"container_id": {
				Description:  "Container ID",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},
			"triggers": {
				Description: "Map of values that, when changed, trigger a container restart",
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func (cr *containerRestart) containerURL(nodeID, clusterFQDN, containerID string) string {
	if nodeID != "" {
		return "/v2/node/" + nodeID + "/exec/container/" + containerID
	}
	return "/v2/cluster/" + clusterFQDN + "/exec/container/" + containerID
}

func (cr *containerRestart) resourceID(nodeID, clusterFQDN, containerID string) string {
	if nodeID != "" {
		return nodeID + "/" + containerID
	}
	return clusterFQDN + "/" + containerID
}

func (cr *containerRestart) parseHCL(d *schema.ResourceData) (containerRestartHCL, error) {
	h := containerRestartHCL{}

	nodeID, ok := d.Get("node_id").(string)
	if !ok {
		return h, errors.New("node_id must be a string")
	}
	h.NodeID = nodeID

	clusterFQDN, ok := d.Get("cluster_fqdn").(string)
	if !ok {
		return h, errors.New("cluster_fqdn must be a string")
	}
	h.ClusterFQDN = clusterFQDN

	containerID, ok := d.Get("container_id").(string)
	if !ok {
		return h, errors.New("container_id must be a string")
	}
	h.ContainerID = containerID

	if raw, ok := d.GetOk("triggers"); ok {
		rawMap, ok := raw.(map[string]any)
		if !ok {
			return h, errors.New("triggers must be a map")
		}
		h.Triggers = make(map[string]string, len(rawMap))
		for k, v := range rawMap {
			s, ok := v.(string)
			if !ok {
				return h, fmt.Errorf("triggers value for key %q must be a string", k)
			}
			h.Triggers[k] = s
		}
	}

	return h, nil
}

func (cr *containerRestart) Create(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	h, err := cr.parseHCL(d)
	if err != nil {
		return diag.FromErr(err)
	}

	url := cr.containerURL(h.NodeID, h.ClusterFQDN, h.ContainerID)

	ct := tg.Container{}
	if err := tgc.Get(ctx, url, &ct); err != nil {
		return diag.FromErr(err)
	}

	originalEnabled := ct.Enabled

	// Stop the container
	ct.Enabled = false
	if _, err := tgc.Put(ctx, url, ct); err != nil {
		return diag.FromErr(err)
	}

	// Wait for container to stop
	select {
	case <-time.After(containerStopWaitDuration):
	case <-ctx.Done():
		// Best-effort restore original state
		ct.Enabled = originalEnabled
		_, _ = tgc.Put(ctx, url, ct)
		return diag.FromErr(ctx.Err())
	}

	// Start the container
	ct.Enabled = true
	if _, err := tgc.Put(ctx, url, ct); err != nil {
		// Attempt to restore original state
		ct.Enabled = originalEnabled
		if _, restoreErr := tgc.Put(ctx, url, ct); restoreErr != nil {
			return diag.FromErr(fmt.Errorf("re-enable failed: %w; also failed to restore original state: %w", err, restoreErr))
		}
		return diag.FromErr(fmt.Errorf("re-enable failed (restored original state): %w", err))
	}

	d.SetId(cr.resourceID(h.NodeID, h.ClusterFQDN, h.ContainerID))

	return nil
}

func (cr *containerRestart) Read(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	tgc := tg.GetClient(meta)

	h, err := cr.parseHCL(d)
	if err != nil {
		return diag.FromErr(err)
	}

	url := cr.containerURL(h.NodeID, h.ClusterFQDN, h.ContainerID)

	ct := tg.Container{}
	err = tgc.Get(ctx, url, &ct)

	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return nil
	case err != nil:
		return diag.FromErr(err)
	}

	// triggers are stored in state only — nothing to read from the API
	return nil
}

func (cr *containerRestart) Delete(_ context.Context, _ *schema.ResourceData, _ any) diag.Diagnostics {
	return nil
}
