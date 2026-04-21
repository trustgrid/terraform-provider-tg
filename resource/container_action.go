package resource

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/trustgrid/terraform-provider-tg/tg"
	"github.com/trustgrid/terraform-provider-tg/validators"
)

const containerRestartDelay = time.Second

type containerTarget struct {
	NodeID      string `tf:"node_id"`
	ClusterFQDN string `tf:"cluster_fqdn"`
	ContainerID string `tf:"container_id"`
}

func containerTargetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
			Description: "Container ID",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
	}
}

func (t containerTarget) resourceID() string {
	if t.NodeID != "" {
		return t.NodeID + "/" + t.ContainerID
	}

	return t.ClusterFQDN + "/" + t.ContainerID
}

func (t containerTarget) containerURL() string {
	if t.NodeID != "" {
		return "/v2/node/" + t.NodeID + "/exec/container/" + t.ContainerID
	}

	return "/v2/cluster/" + t.ClusterFQDN + "/exec/container/" + t.ContainerID
}

func getContainerBase(ctx context.Context, tgc *tg.Client, target containerTarget) (tg.Container, error) {
	container := tg.Container{}
	err := tgc.Get(ctx, target.containerURL(), &container)
	if err != nil {
		return container, err
	}

	container.NodeID = target.NodeID
	container.ClusterFQDN = target.ClusterFQDN
	return container, nil
}

func updateContainerEnabled(ctx context.Context, tgc *tg.Client, target containerTarget, enabled bool) error {
	container, err := getContainerBase(ctx, tgc, target)
	if err != nil {
		return err
	}

	container.Enabled = enabled
	_, err = tgc.Put(ctx, target.containerURL(), container)
	return err
}

func restartContainer(ctx context.Context, tgc *tg.Client, target containerTarget) error {
	if err := updateContainerEnabled(ctx, tgc, target, false); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(containerRestartDelay):
	}

	return updateContainerEnabled(ctx, tgc, target, true)
}

func readContainerActionTarget(ctx context.Context, d *schema.ResourceData, meta any) (tg.Container, containerTarget, diag.Diagnostics) {
	target := containerTarget{
		NodeID:      d.Get("node_id").(string),      //nolint: errcheck // schema guarantees type
		ClusterFQDN: d.Get("cluster_fqdn").(string), //nolint: errcheck // schema guarantees type
		ContainerID: d.Get("container_id").(string), //nolint: errcheck // schema guarantees type
	}

	container, err := getContainerBase(ctx, tg.GetClient(meta), target)
	var nferr *tg.NotFoundError
	switch {
	case errors.As(err, &nferr):
		d.SetId("")
		return tg.Container{}, target, nil
	case err != nil:
		return tg.Container{}, target, diag.FromErr(err)
	default:
		return container, target, nil
	}
}

func triggerHash(triggers map[string]string) string {
	keys := make([]string, 0, len(triggers))
	for key := range triggers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+triggers[key])
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(strings.Join(parts, "\n"))))
}
